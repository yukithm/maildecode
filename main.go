package main

import (
	"os"

	"golang.org/x/text/transform"
)

func main() {
	msg, err := DecodeMail(os.Stdin)
	if err != nil {
		panic(err)
	}

	normalizer := &NewlineNormalizer{
		Newline: []byte("\r\n"),
	}
	w := transform.NewWriter(os.Stdout, normalizer)
	defer w.Close()
	PrintMessage(w, msg)
}
