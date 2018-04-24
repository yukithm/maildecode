package main

import (
	"golang.org/x/text/transform"
)

// NewlineNormalizer transforms newline codes(CR,LF,CRLF) to t.Newline.
type NewlineNormalizer struct {
	transform.NopResetter
	Newline []byte
}

// Transform implements the transform.Transformer interface.
func (t *NewlineNormalizer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			return
		}

		switch src[nSrc] {
		case '\r':
			if nSrc+1 >= len(src) {
				if atEOF {
					if nDst+len(t.Newline) > len(dst) {
						err = transform.ErrShortDst
						return
					}
					nDst += copy(dst[nDst:], t.Newline)
				} else {
					err = transform.ErrShortSrc
					return
				}
			} else {
				if src[nSrc+1] != '\n' {
					if nDst+len(t.Newline) > len(dst) {
						err = transform.ErrShortDst
						return
					}
					nDst += copy(dst[nDst:], t.Newline)
				}
			}

		case '\n':
			if nDst+len(t.Newline) > len(dst) {
				err = transform.ErrShortDst
				return
			}
			nDst += copy(dst[nDst:], t.Newline)

		default:
			dst[nDst] = src[nSrc]
			nDst++
		}

		nSrc++
	}

	return
}
