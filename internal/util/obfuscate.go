package util

import (
	"encoding/base64"
	"io"
	"net/http"
)

const ObfuscationKey = "linx-evasion-key"

func XorTransform(data []byte, offset int) []byte {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ ObfuscationKey[(i+offset)%len(ObfuscationKey)]
	}
	return result
}

func Obfuscate(data []byte) string {
	transformed := XorTransform(data, 0)
	return base64.StdEncoding.EncodeToString(transformed)
}

func Deobfuscate(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return XorTransform(data, 0), nil
}

type XorResponseWriter struct {
	http.ResponseWriter
	Offset int
}

func (w *XorResponseWriter) Write(b []byte) (int, error) {
	transformed := XorTransform(b, w.Offset)
	n, err := w.ResponseWriter.Write(transformed)
	if n > 0 {
		w.Offset += n
	}
	return n, err
}

type XorReadSeekCloser struct {
	io.ReadSeekCloser
}

func (r *XorReadSeekCloser) Read(p []byte) (int, error) {
	offset, err := r.ReadSeekCloser.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	n, err := r.ReadSeekCloser.Read(p)
	if n > 0 {
		transformed := XorTransform(p[:n], int(offset))
		copy(p[:n], transformed)
	}
	return n, err
}
