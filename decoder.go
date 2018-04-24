package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
)

type Message struct {
	Header   textproto.MIMEHeader
	Body     io.Reader
	Parts    []*Message
	Boundary string
}

func (m *Message) ContentType() string {
	return m.Header.Get("Content-Type")
}

func (m *Message) MediaType() (string, map[string]string, error) {
	return mime.ParseMediaType(m.ContentType())
}

func DecodeMail(input io.Reader) (*Message, error) {
	msg, err := mail.ReadMessage(input)
	if err != nil {
		return nil, err
	}

	return decodeMail(textproto.MIMEHeader(msg.Header), msg.Body)
}

func decodeMail(header textproto.MIMEHeader, body io.Reader) (*Message, error) {
	decoded := &Message{}
	decHeader, err := decodeHeader(header)
	if err != nil {
		return nil, err
	}
	decoded.Header = decHeader
	decoded.Body = body
	wrapTransferEncoding(decoded)

	mediaType, params, err := decoded.MediaType()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(mediaType, "text/") {
		if err := wrapContentEncoding(decoded); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(mediaType, "multipart") {
		rawBody, err := ioutil.ReadAll(decoded.Body)
		if err != nil {
			return nil, err
		}
		body, err := readUntilBoundary(bytes.NewReader(rawBody), params["boundary"])
		if err != nil {
			return nil, err
		}
		parts, err := decodeMultipart(bytes.NewReader(rawBody), mediaType, params)
		if err != nil {
			return nil, err
		}
		if body != nil {
			decoded.Body = bytes.NewReader(body)
		}
		decoded.Parts = parts
		decoded.Boundary = params["boundary"]
	}

	return decoded, nil
}

func decodeHeader(header textproto.MIMEHeader) (textproto.MIMEHeader, error) {
	mdec := &mime.WordDecoder{
		CharsetReader: EncodingCache.GetDecodeReader,
	}

	decoded := make(textproto.MIMEHeader)
	for name, values := range header {
		for _, value := range values {
			decv, err := mdec.DecodeHeader(value)
			if err != nil {
				return nil, err
			}
			decoded.Add(name, decv)
		}
	}

	return decoded, nil
}

func wrapTransferEncoding(msg *Message) {
	encoding := msg.Header.Get("Content-Transfer-Encoding")
	if encoding == "base64" {
		msg.Header.Del("Content-Transfer-Encoding")
		msg.Body = base64.NewDecoder(base64.StdEncoding, msg.Body)
	} else if encoding == "quoted-printable" {
		msg.Header.Del("Content-Transfer-Encoding")
		msg.Body = quotedprintable.NewReader(msg.Body)
	}
}

func wrapContentEncoding(msg *Message) error {
	mediaType, params, err := msg.MediaType()
	if err != nil {
		return err
	} else if !strings.HasPrefix(mediaType, "text/") {
		return nil
	}

	if charset, ok := params["charset"]; ok && charset != "" {
		decoder, err := EncodingCache.GetDecodeReader(charset, msg.Body)
		if err != nil {
			return err
		}
		params["charset"] = "utf-8"
		msg.Header.Set("Content-Type", mime.FormatMediaType(mediaType, params))
		msg.Body = decoder
	}

	return nil
}

func readUntilBoundary(r io.Reader, boundary string) ([]byte, error) {
	nlBoundary := []byte("\r\n--" + boundary)

	scanBoundaries := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, nlBoundary); i >= 0 {
			return i + 1, data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(scanBoundaries)

	if scanner.Scan() {
		return scanner.Bytes(), nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func decodeMultipart(r io.Reader, mediaType string, params map[string]string) ([]*Message, error) {
	if params["boundary"] == "" {
		return nil, errors.New("boundary not specified")
	}

	var parts []*Message
	mr := multipart.NewReader(r, params["boundary"])
	for {
		p, err := mr.NextPart()
		if p != nil {
			var body bytes.Buffer
			if _, err := body.ReadFrom(p); err != nil {
				return nil, err
			}
			part, err := decodeMail(p.Header, &body)
			if err != nil {
				return nil, err
			}
			parts = append(parts, part)
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
	}

	return parts, nil
}
