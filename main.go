package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	Reset    = "\x1b[0m"
	Bold     = "\033[1m"
	Gray     = "\x1b[38;2;148;163;184m"
	Cyan     = "\x1b[38;2;80;190;220m"
	Red      = "\x1b[38;2;235;92;92m"
	Blue     = "\x1b[38;2;100;160;245m"
	Green    = "\x1b[38;2;80;190;120m"
	Yellow   = "\x1b[38;2;235;170;60m"
	BgHex    = "\x1b[48;2;23;78;99m"
	BgString = "\x1b[48;2;110;75;20m"
	BgByte   = "\x1b[41m"
)

func main() {
	cols := flag.Int("c", 16, "# of cols per line")
	gs := flag.Int("g", 1, "# of octets per group")
	bt := flag.Int("b", -1, "Byte offset to highlight") // -1 means no bytes selected
	de := flag.String("d", "", "Decode selected byte(s) defined with '-b'")
	xstart := flag.Int("x", 0, "Show bytes from this zero-based offset (requires an end offset)")
	tail := flag.Int("X", 0, "Show this many bytes from the end of the file")

	ffind := flag.String("F", "", "String to search for")
	xfind := flag.String("f", "", "Hex bytes to search for")
	flagArgs, args, byteEnd, truncateEnd, hasTruncate, hasTail, err := reorderArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := flag.CommandLine.Parse(flagArgs); err != nil {
		return
	}
	be := *bt
	if byteEnd != nil {
		be = *byteEnd
	}

	if len(args) != 1 {
		fmt.Printf("E: Expected filepath\n")
		return
	}
	file := args[0]
	if *de != "" {
		if *bt < 1 {
			fmt.Fprintf(os.Stderr, "-d bit requires a -b selection\n")
			return
		}
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		if be > len(data) {
			fmt.Fprintf(os.Stderr, "selected range contains EOF\n")
			return
		}
		if *de == "bit" {
			fmt.Printf("bits: ")
			for _, b := range data[*bt-1 : be] {
				fmt.Printf("%08b ", b)
			}
		} else if *de == "byte" {
			fmt.Printf("bytes: ")
			for _, b := range data[*bt-1 : be] {
				fmt.Printf("%02x ", b)
			}
		} else if *de == "dec" {
			fmt.Printf("decimals: ")
			for _, b := range data[*bt-1 : be] {
				fmt.Printf("%02d ", b)
			}
		} else if *de == "str" {
			fmt.Printf("strings: %s", string(data[*bt-1:be]))
		}
		fmt.Printf("\n")
	}
	//	if *de == "byte" {
	//		if *bt < 1 {
	//			fmt.Fprintf(os.Stderr, "-d byte requires a -b selection\n")
	//		}
	//		data, err := os.ReadFile(file)
	//		if err != nil {
	//			panic(err)
	//		}
	//		if be > len(data) {
	//			fmt.Fprintf(os.Stderr, "selected range contains EOF\n")
	//			return
	//		}
	//		fmt.Printf("Bytes: ")
	//		for _, b := range data[*bt-1 : be] {
	//			fmt.Printf("%d ", b)
	//		}
	//		fmt.Printf("\n")
	//	}
	if *cols <= 0 || *gs < 0 {
		fmt.Fprintln(os.Stderr, "columns must be positive and group size cannot be negative")
		return
	}

	//	pattern := []byte(*ffind)
	var pattern []byte
	if *xfind != "" {
		decoded, err := hex.DecodeString(*xfind)
		if err != nil {
			fmt.Fprintf(os.Stderr, ": invalid hex pattern: %v\n", err)
			return
		}
		pattern = decoded
	} else {
		pattern = []byte(*ffind)
	}
	//	hlcolor := Bgcyan
	//	if *xfind == "" && *ffind != "" {
	//		hlcolor = Bgmagenta
	//	} else {
	//		hlcolor = Bgcyan
	//	}
	var matches []int
	var stringmatches []int
	if len(pattern) > 0 {
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		matches = find(data, pattern)
		stringmatches = find(data, []byte(*ffind))
		//		for _, at := range matches {
		//			if len(*ffind) > 0 {
		//				fmt.Printf("Found string '%v' on index 0x%08x\n", *ffind, at)
		//			}
		//			if len(*xfind) > 0 {
		//				fmt.Printf("Found hex sequence '%v' on index 0x%08x\n", *xfind, at)
		//			}
		//
		//		}
	}

	buf := make([]byte, *cols)
	offset := 0
	//	buf := make([]byte, 16)
	//file, err := os.Open(os.Args[1])
	//	n, err := file.Read(buf)
	//	if err != nil {
	//		panic(err)
	//	}
	//	inf, err := file.Stat()
	//	if err != nil {
	//		panic(err)
	//	}
	//	size := inf.Size()
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	info, err := os.Stat(file)
	if err != nil {
		panic(err)
	}

	size := info.Size()
	start, end := int64(0), size
	if hasTruncate && hasTail {
		fmt.Fprintln(os.Stderr, "-x and -X cannot be used together")
		return
	}
	if hasTruncate {
		if *xstart < 0 || truncateEnd == nil || *truncateEnd < *xstart {
			fmt.Fprintln(os.Stderr, "-x requires a non-negative start and an end offset not less than the start")
			return
		}
		start, end = int64(*xstart), int64(*truncateEnd)
		if end > size {
			fmt.Fprintln(os.Stderr, "-x range extends past EOF")
			return
		}
	}
	if hasTail {
		if *tail < 0 {
			fmt.Fprintln(os.Stderr, "-X byte count cannot be negative")
			return
		}
		start = size - int64(*tail)
		if start < 0 {
			start = 0
		}
	}

	fmt.Printf("Read file %v; size: %d bytes\n", file, size)

	reader := io.NewSectionReader(f, start, end-start)
	offset = int(start)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			fmt.Printf("%v%08x: %v", Gray, offset, Reset)
			for i, b := range buf[:n] {
				if *gs > 0 && i > 0 && i%*gs == 0 {
					fmt.Printf(" ")
				}
				position := offset + i
				byteh := *bt >= 0 &&
					position+1 >= *bt &&
					position+1 <= be

				stringh := *ffind != "" && hl(position, stringmatches, len(*ffind))
				hexh := *xfind != "" && hl(position, matches, len(pattern))

				if byteh {
					fmt.Printf(BgByte)
				} else if stringh {
					fmt.Print(BgString)
				} else if hexh {
					fmt.Print(BgHex)
				}

				fmt.Printf("%v%s%02x", Bold, byteColor(b), b)

				if byteh || stringh || hexh {
					fmt.Print(Reset)
				}
			}
			for i := n; i < *cols; i++ {
				if *gs > 0 && i > 0 && i%*gs == 0 {
					fmt.Print(" ")
				}
				fmt.Print("  ")
			}
			fmt.Printf(" \x1b[0m   ")
			for _, b := range buf[:n] {
				if unicode.IsPrint(rune(b)) && b < 128 {
					fmt.Printf("%v%c", Bold, b)
				} else {
					fmt.Printf("%v.%v", Red, Reset)
				}
			}
			fmt.Println(Reset)
			//			fmt.Printf("%08x : % x\n", offset, buf[:n])
			offset += n

		}
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

	}
}

func reorderArgs(args []string) ([]string, []string, *int, *int, bool, bool, error) {
	valueCounts := map[string]int{"-b": 1, "-c": 1, "-d": 1, "-f": 1, "-F": 1, "-g": 1, "-x": 2, "-X": 1}
	var flagArgs, positional []string
	var byteEnd, truncateEnd *int
	hasTruncate, hasTail := false, false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		name, _, hasInlineValue := strings.Cut(arg, "=")
		count, isFlag := valueCounts[name]
		if !isFlag {
			if strings.HasPrefix(arg, "-") {
				flagArgs = append(flagArgs, arg)
			} else {
				positional = append(positional, arg)
			}
			continue
		}
		flagArgs = append(flagArgs, arg)
		valueIndexOffset := 0
		if hasInlineValue {
			count--
			if name == "-x" {
				valueIndexOffset = 1
			}
		}
		for valueIndex := 0; valueIndex < count; valueIndex++ {
			if i+1 >= len(args) {
				return nil, nil, nil, nil, false, false, fmt.Errorf("%s requires %d value(s)", name, valueCounts[name])
			}
			i++
			value := args[i]
			if name == "-x" && valueIndex+valueIndexOffset == 1 {
				end, err := strconv.Atoi(value)
				if err != nil {
					return nil, nil, nil, nil, false, false, fmt.Errorf("-x end offset must be an integer")
				}
				truncateEnd = &end
				hasTruncate = true
				continue
			}
			flagArgs = append(flagArgs, value)
		}
		if name == "-b" {
			byteEnd = nil
			if i+1 < len(args) {
				if end, err := strconv.Atoi(args[i+1]); err == nil {
					i++
					byteEnd = &end
				}
			}
		}
		if name == "-x" {
			hasTruncate = true
		}
		if name == "-X" {
			hasTail = true
		}
	}
	return flagArgs, positional, byteEnd, truncateEnd, hasTruncate, hasTail, nil
}

func byteColor(b byte) string {
	switch {
	case b == 0x00:
		return Blue
	case b == 0xFF:
		return Red
	case b >= 0x20 && b <= 0x7E:
		return Green
	case b >= 0x80:
		return Cyan
	default:
		return Gray
	}
}

func find(data, pattern []byte) []int {
	if len(pattern) == 0 {
		return nil
	}

	var offsets []int
	searchf := 0

	for searchf <= len(data)-len(pattern) {
		relative := bytes.Index(data[searchf:], pattern)
		if relative == -1 {
			break
		}

		matchat := searchf + relative
		offsets = append(offsets, matchat)
		searchf = matchat + 1
	}
	return offsets
}

func hl(position int, matches []int, patternl int) bool {
	for _, q := range matches {
		if position >= q && position < q+patternl {
			return true
		}
	}
	return false
}
