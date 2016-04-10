package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func ConvToUTF8(b []byte, e encoding.Encoding) (result []byte, err error) {
	reader := transform.NewReader(bytes.NewReader(b), unicode.BOMOverride(e.NewDecoder()))
	return ioutil.ReadAll(reader)
}

func ConvTo(b []byte, e encoding.Encoding) (result []byte, err error) {
	w := new(bytes.Buffer)
	writer := transform.NewWriter(w, e.NewEncoder())
	defer writer.Close()

	if _, err = writer.Write(b); err != nil {
		return
	}
	return w.Bytes(), nil
}

// From https://groups.google.com/forum/#!topic/golang-nuts/eex1wLCvK58
var boms = map[string][]byte{
	"utf-16be": []byte{0xfe, 0xff},
	"utf-16le": []byte{0xff, 0xfe},
	"utf-8":    []byte{0xef, 0xbb, 0xbf},
}

func TrimBOM(b []byte, encoding string) []byte {
	bom := boms[encoding]
	if bom != nil {
		b = bytes.TrimPrefix(b, bom)
	}
	return b
}

/*
// naive implemntation
func NewUTF8Reader(label string, r io.Reader) (io.Reader, error) {
	e, name := charset.Lookup(label)
	if e == nil {
		return nil, fmt.Errorf("unsupported charset: %q", label)
	}
	// TODO: implement custom mulitreader to use a freelist?
	preview := make([]byte, 512)
	n, err := io.ReadFull(r, preview)
	switch {
	case err == io.ErrUnexpectedEOF:
		preview = TrimBOM(preview[:n], name)
		r = bytes.NewReader(preview)
	case err != nil:
		return nil, err
	default:
		preview = TrimBOM(preview, name)
		r = io.MultiReader(bytes.NewReader(preview), r)
	}
	return transform.NewReader(r, e.NewDecoder()), nil
}
*/

func NewUTF8Reader(label string, r io.Reader) (io.Reader, error) {
	e, _ := charset.Lookup(label)
	if e == nil {
		return nil, fmt.Errorf("unsupported charset: %q", label)
	}
	return transform.NewReader(r, unicode.BOMOverride(e.NewDecoder())), nil
}

func DumpReader(r io.Reader, n int) (reader []io.Reader, done <-chan struct{}) {
	var writer []io.Writer
	for i := 0; i < n; i++ {
		r, w := io.Pipe()
		reader = append(reader, r)
		writer = append(writer, w)
	}
	ch := make(chan struct{}, 1)
	go func() {
		io.Copy(io.MultiWriter(writer...), r)
		ch <- struct{}{}
		for i := 0; i < n; i++ {
			writer[i].(*io.PipeWriter).Close()
		}
	}()
	return reader, ch
}