package md5

import (
	"crypto/md5"
	"fmt"
	"io"
)

func Md5(str string) (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, str)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
