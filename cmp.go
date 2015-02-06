package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var lOpt = flag.Bool("l", false, "List differing bytes")
var sOpt = flag.Bool("s", false, "Silent")

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Exactly two arguments are required\n")
	}

	var exitStatus int

	if "-" == args[0] || "-" == args[1] {
		// To deal with the obscure corner case:
		//   cmp - /dev/null <&-
		// We stat stdin before using it.
		_, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
	}

	var in []*os.File
	for _, f := range args {
		fd, err := open(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}
		in = append(in, fd)
	}

	var byteCount uint64
	var lineCount uint64 = 1

	for {
		a := make([]byte, 1, 1)
		b := make([]byte, 1, 1)
		_, errA := in[0].Read(a)
		_, errB := in[1].Read(b)

		byteCount += 1

		if errA == io.EOF && errB == io.EOF {
			// Files are same length.
			os.Exit(exitStatus)
		}

		err := errA
		if err == nil {
			err = errB
		}
		if err == io.EOF {
			// One file is shorter.
			shorter := args[0]
			if errB == io.EOF {
				shorter = args[1]
			}
			if !*sOpt {
				fmt.Fprintf(os.Stderr, "cmp: EOF on %s\n", shorter)
			}
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		}

		if a[0] != b[0] {
			exitStatus = 1
			if !*sOpt && !*lOpt {
				fmt.Fprintf(os.Stdout, "%s %s differ: char %d, line %d\n", args[0], args[1], byteCount, lineCount)
			}
			if !*lOpt {
				os.Exit(exitStatus)
			}
			fmt.Fprintf(os.Stdout, "%d %o %o\n", byteCount, a[0], b[0])
		}
		if a[0] == '\n' {
			lineCount += 1
		}
	}
}

func open(file string) (*os.File, error) {
	if file == "-" {
		return os.Stdin, nil
	}
	return os.Open(file)
}
