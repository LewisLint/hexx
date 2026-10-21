package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode"
)

const (
	Reset     = "\x1b[0m"
	Bold      = "\033[1m"
	Gray      = "\x1b[38;2;110;106;134m" // Subtle
	Cyan      = "\x1b[38;2;156;207;216m" // Foam
	Bgcyan    = "\x1b[46m"               // Background cyan
	Red       = "\x1b[38;2;235;111;146m" // Love
	Purple    = "\x1b[38;2;196;167;231m" // Iris
	Green     = "\x1b[38;2;49;116;143m"  // Pine
	Yellow    = "\x1b[38;2;246;193;119m" // Gold
	Bgmagenta = "\x1b[45m"
)

func main() {
	cols := flag.Int("c", 16, "# of cols per line")
	gs := flag.Int("g", 1, "# of octets per group")
	ffind := flag.String("F", "", "String to search for")
	xfind := flag.String("f", "", "Hex bytes to search for")
	flag.Parse()
	args := flag.Args()
	if len(args) > 1 {
		fmt.Printf("E: Expected filepath\n")
		return
	}
	file := args[0]

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

	fmt.Printf("Read file %v; size: %d bytes\n", file, size)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Printf("%08x: ", offset)
			for i, b := range buf[:n] {
				if i > 0 && i%*gs == 0 { // some weird modulo shit. checks if i is divisible by *gs
					fmt.Printf(" ")
				}
				position := offset + i
				stringh := *ffind != "" && hl(position, stringmatches, len(*ffind))
				hexh := *xfind != "" && hl(position, matches, len(pattern))

				if stringh {
					fmt.Print(Bgmagenta)
				} else if hexh {
					fmt.Print(Bgcyan)
				}

				fmt.Printf("%v%s%02x", Bold, byteColor(b), b)

				if stringh || hexh {
					fmt.Print(Reset)
				}
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

func byteColor(b byte) string {
	switch {
	case b == 0x00:
		return Purple
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
