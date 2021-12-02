package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"os"
)

func RSAGenKey(customBlock string, bits int) ([]byte, []byte, error) {
	var (
		privateBytes []byte
		publicBytes  []byte
	)
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return privateBytes, publicBytes, errors.WithStack(err)
	}
	// 2. X509 ASN.1 DER str
	privateStream := x509.MarshalPKCS1PrivateKey(privateKey)
	// 3. set pem.Block
	privateBlock := pem.Block{
		Type:  customBlock + " PRIVATE KEY",
		Bytes: privateStream,
	}
	privateBytes = pem.EncodeToMemory(&privateBlock)

	// 4. gen public key
	publicKey := privateKey.PublicKey
	publicStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	publicBlock := pem.Block{
		Type:  customBlock + " PUBLIC KEY",
		Bytes: publicStream,
	}
	publicBytes = pem.EncodeToMemory(&publicBlock)
	return privateBytes, publicBytes, nil
}

func RSAReadKeyFromFile(filename string) []byte {
	f, err := os.Open(filename)
	var b []byte

	if err != nil {
		return b
	}
	defer f.Close()
	fileInfo, _ := f.Stat()
	b = make([]byte, fileInfo.Size())
	f.Read(b)
	return b
}

func RSAEncrypt(data, publicBytes []byte) ([]byte, error) {
	var res []byte
	block, _ := pem.Decode(publicBytes)

	if block == nil {
		return res, errors.Errorf("pem decode failed, may be public bytes is wrong")
	}

	keyInit, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return res, errors.Wrap(err, "x509 parse failed")
	}
	pubKey := keyInit.(*rsa.PublicKey)
	res, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return res, errors.Wrap(err, "rsa encrypt failed")
	}
	return []byte(EncodeStr2Base64(string(res))), nil
}

func RSADecrypt(base64Data, privateBytes []byte) ([]byte, error) {
	var res []byte
	data := []byte(DecodeStrFromBase64(string(base64Data)))
	block, _ := pem.Decode(privateBytes)
	if block == nil {
		return res, errors.Errorf("pem decode failed, may be public bytes is wrong")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return res, errors.Wrap(err, "x509 parse failed")
	}
	res, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return res, errors.Wrap(err, "rsa encrypt failed")
	}
	return res, nil
}
