package main

import (
	"io"
	"os"
	"time"
)

type tailReader struct { // https://stackoverflow.com/a/31122253
	io.ReadCloser
}

func (t tailReader) Read(b []byte) (int, error) {
	for {
		n, err := t.ReadCloser.Read(b)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func newTailReader(fileName string) (tailReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return tailReader{}, err
	}

	if _, err := f.Seek(0, 2); err != nil {
		return tailReader{}, err
	}
	return tailReader{f}, nil
}
