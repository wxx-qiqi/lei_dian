package tools

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"strings"
)

// GBKToUTF8 GBK 转 UTF-8
func GBKToUTF8(input []byte) ([]byte, error) {
	reader := transform.NewReader(strings.NewReader(string(input)), simplifiedchinese.GBK.NewDecoder())
	return ioutil.ReadAll(reader)
}
