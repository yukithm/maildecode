package main

import (
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"
)

type encodingCache map[string]encoding.Encoding

func (ec encodingCache) Get(name string) (encoding.Encoding, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if enc, ok := ec[name]; ok {
		return enc, nil
	}

	enc, err := ianaindex.MIME.Encoding(name)
	if err != nil || enc == nil {
		enc, err = htmlindex.Get(name)
		if err != nil || enc == nil {
			return nil, err
		}
	}

	ec[name] = enc

	return enc, nil
}

func (ec encodingCache) GetDecoder(charset string) (*encoding.Decoder, error) {
	enc, err := ec.Get(charset)
	if err != nil {
		return nil, err
	}
	return enc.NewDecoder(), nil
}

func (ec encodingCache) GetDecodeReader(charset string, input io.Reader) (io.Reader, error) {
	decoder, err := ec.GetDecoder(charset)
	if err != nil {
		return nil, err
	}

	return decoder.Reader(input), nil
}

var EncodingCache = make(encodingCache, 0)
