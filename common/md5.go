package common

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func GetBytesMd5(b []byte) []byte {
	r := bytes.NewReader(b)
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		fmt.Println(err)
	}

	return h.Sum(nil)
}

func GetFileMd5(fp string) []byte {
	f, err := os.Open(fp)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Println(err)
	}

	return h.Sum(nil)
}
