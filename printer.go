package main

import (
	"fmt"
	"io"
	"net/textproto"
	"sort"
)

func PrintMessage(w io.Writer, msg *Message) {
	PrintHeader(w, msg.Header)

	fmt.Fprintln(w)
	if msg.Body != nil {
		PrintBody(w, msg.Body)
	}
	if msg.Parts != nil {
		printParts(w, msg.Parts, msg.Boundary)
	}
}

func PrintHeader(w io.Writer, header textproto.MIMEHeader) {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		values := header[name]
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", name, value)
		}
	}
}

func PrintBody(w io.Writer, body io.Reader) error {
	_, err := io.Copy(w, body)
	return err
}

func printParts(w io.Writer, parts []*Message, boundary string) {
	for _, part := range parts {
		fmt.Fprintf(w, "\n--%s\n", boundary)
		PrintMessage(w, part)
	}
	fmt.Fprintf(w, "\n--%s--\n", boundary)
}
