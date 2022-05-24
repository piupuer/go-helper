package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
)

func RSAGenKey(customBlock string, bits int) (privateBytes []byte, publicBytes []byte, err error) {
	// generate private key
	var privateKey *rsa.PrivateKey
	privateKey, err = rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return
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
	var publicStream []byte
	publicStream, err = x509.MarshalPKIXPublicKey(&publicKey)
	publicBlock := pem.Block{
		Type:  customBlock + " PUBLIC KEY",
		Bytes: publicStream,
	}
	publicBytes = pem.EncodeToMemory(&publicBlock)
	return
}

func RSAEncrypt(data, publicBytes []byte) (base64Data []byte, err error) {
	pubKey := loadRsaPubKey(publicBytes)
	var encodeData []byte
	encodeData, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		err = errors.Wrap(err, "rsa encrypt failed")
		return
	}
	base64Data = []byte(EncodeStr2Base64(string(encodeData)))
	return
}

func RSADecrypt(base64Data, privateBytes []byte) (res []byte, err error) {
	data := []byte(DecodeStrFromBase64(string(base64Data)))
	priKey := loadRsaPriKey(privateBytes)
	res, err = rsa.DecryptPKCS1v15(rand.Reader, priKey, data)
	if err != nil {
		err = errors.Wrap(err, "rsa decrypt failed")
		return
	}
	return
}

func RSASign(data, privateBytes []byte) (base64Signature []byte, err error) {
	sum := loadSha256Sum(data)
	priKey := loadRsaPriKey(privateBytes)
	var signature []byte
	signature, err = rsa.SignPSS(rand.Reader, priKey, crypto.SHA256, sum, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	})
	if err != nil {
		err = errors.Wrap(err, "rsa sign failed")
		return
	}
	base64Signature = []byte(EncodeStr2Base64(string(signature)))
	return
}

func RSAVerify(data, base64Signature, publicBytes []byte) (flag bool) {
	signature := []byte(DecodeStrFromBase64(string(base64Signature)))
	sum := loadSha256Sum(data)
	pubKey := loadRsaPubKey(publicBytes)
	err := rsa.VerifyPSS(pubKey, crypto.SHA256, sum, signature, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	})
	flag = err == nil
	return
}

func loadRsaPriKey(privateBytes []byte) (privateKey *rsa.PrivateKey) {
	privateKey = &rsa.PrivateKey{}
	block, _ := pem.Decode(privateBytes)
	if block == nil {
		return
	}
	var err error
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}
	return
}

func loadRsaPubKey(publicBytes []byte) (publicKey *rsa.PublicKey) {
	publicKey = &rsa.PublicKey{}
	block, _ := pem.Decode(publicBytes)
	if block == nil {
		return
	}

	var keyInit interface{}
	var err error
	keyInit, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}
	publicKey = keyInit.(*rsa.PublicKey)
	return
}

func loadSha256Sum(data []byte) (sum []byte) {
	msgHash := crypto.SHA256.New()
	_, err := msgHash.Write(data)
	if err == nil {
		sum = msgHash.Sum(nil)
	}
	return
}
