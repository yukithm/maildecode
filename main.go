package main

import (
	"io"
	"os"

	"golang.org/x/text/transform"
)

func main() {
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
