package main

import "golang.org/x/tour/reader"

type MyReader struct{}

func (mr MyReader) Read([]byte) (int, error) {
	return mr.Read([]byte{'A'})
}

func main() {
	reader.Validate(MyReader{})
}
