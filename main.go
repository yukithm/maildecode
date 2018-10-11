package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/transform"
)

type Options struct {
	Version bool
}

var cmdOptions Options

func init() {
	flag.BoolVar(&cmdOptions.Version, "version", false, "Print version")
}

func main() {
	flag.Parse()

	if cmdOptions.Version {
		fmt.Printf("maildecode %s\n", Version)
		return
	}

	normalizer := &NewlineNormalizer{
		Newline: []byte("\r\n"),
	}
	w := transform.NewWriter(os.Stdout, normalizer)
	defer w.Close()

	if len(os.Args) > 1 {
		for _, file := range os.Args[1:] {
			printFile(w, file)
		}
	} else {
		printMail(w, os.Stdin)
	}

}

func printFile(w io.Writer, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return printMail(w, f)
}

func printMail(w io.Writer, f io.Reader) error {
	msg, err := DecodeMail(f)
	if err != nil {
		return err
	}
	PrintMessage(w, msg)
	return nil
}
