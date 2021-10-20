package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io/ioutil"
)

// compress string by zlib
func CompressStrByZlib(s string) (string, error) {
	var b bytes.Buffer
	gz := zlib.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	res := base64.StdEncoding.EncodeToString(b.Bytes())
	return res, nil
}

// decompression string by zlib
func DeCompressStrByZlib(s string) string {
	data, _ := base64.StdEncoding.DecodeString(s)
	rData := bytes.NewReader(data)
	r, _ := zlib.NewReader(rData)
	b, _ := ioutil.ReadAll(r)
	return string(b)
}
