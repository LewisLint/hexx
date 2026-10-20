package main

import (
	"fmt"
	"io"
	"os"
)

const (
	Reset  = "\x1b[0m"
	Gray   = "\x1b[38;2;110;106;134m" // Subtle
	Cyan   = "\x1b[38;2;156;207;216m" // Foam
	Red    = "\x1b[38;2;235;111;146m" // Love
	Purple = "\x1b[38;2;196;167;231m" // Iris
	Green  = "\x1b[38;2;49;116;143m"  // Pine
	Yellow = "\x1b[38;2;246;193;119m" // Gold
)

func main() {

	buf := make([]byte, 16)
	offset := 0
	//	buf := make([]byte, 16)
	file, err := os.Open(os.Args[1])
	//	n, err := file.Read(buf)
	if err != nil {
		panic(err)
	}
	inf, err := file.Stat()
	if err != nil {
		panic(err)
	}
	size := inf.Size()
	fmt.Printf("Read file %v; size: %d bytes\n", os.Args[1], size)

	for {
		n, err := file.Read(buf)
		if n > 0 {
			fmt.Printf("%08x: ", offset)
			for _, b := range buf[:n] {
				fmt.Printf("%s %02x\x1b[0m", byteColor(b), b)
			}
			fmt.Println()
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
