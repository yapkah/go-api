package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/smartblock/gta-api/pkg/e"
)

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey() (*rsa.PrivateKey, error) {
	var err error

	res, err := ioutil.ReadFile("storage/encrypt_private.pem")
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_PRIVATE_KEY_MISSING, Data: map[string]interface{}{"err": err.Error()}}
	}

	block, _ := pem.Decode(res)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.DECREPT_PEM_PRIVATE_ERROR, Data: map[string]interface{}{"err": err.Error()}}
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_RSA_PRIVATE_KEY, Data: map[string]interface{}{"err": err.Error()}}
	}
	return key, nil
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey() (*rsa.PublicKey, error) {
	var err error
	res, err := ioutil.ReadFile("storage/encrypt_public.pem")
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_PRIVATE_KEY_MISSING, Data: map[string]interface{}{"err": err.Error()}}
	}

	block, _ := pem.Decode(res)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.DECREPT_PEM_PUBLIC_ERROR, Data: map[string]interface{}{"err": err.Error()}}
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_RSA_PUBLIC_KEY, Data: map[string]interface{}{"err": err.Error()}}
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_RSA_PUBLIC_KEY, Data: map[string]interface{}{"err": err.Error()}}
	}
	return key, nil
}

// EncryptWithPublicKey encrypts data with public key [only work with EncryptOAEP]
func EncryptWithPublicKey(msg []byte) (string, error) {
	pub, err := BytesToPublicKey()
	if err != nil {
		return "", err
	}

	hash := sha512.New()
	ct, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_ENCRYPT_ERROR, Data: map[string]interface{}{"err": err.Error()}}
	}
	ciphertext := base64.StdEncoding.EncodeToString(ct)
	return ciphertext, nil
}

// DecryptWithPrivateKey decrypts data with private key [only work with DecryptOAEP]
func DecryptWithPrivateKey(ciphertext string) (string, error) {
	priv, err := BytesToPrivateKey()
	if err != nil {
		return "", err
	}

	ctext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_INVALID_TEXT, Data: map[string]interface{}{"err": err.Error()}}
	}

	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ctext, nil)
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_DECRYPT_ERROR, Data: map[string]interface{}{"err": err.Error()}}
	}
	return string(plaintext), nil
}

// func RsaEncryptPKCS1v15 will work with EncryptPKCS1v15
// can refer more detail in https://gist.github.com/hothero/93c69bbd57001ce0a1997f5dd1ba89f6
func RsaEncryptPKCS1v15(plainText string) (string, error) {
	ciphertext := []byte(plainText)

	pubKey, err := ioutil.ReadFile("storage/encrypt_public.pem")
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_PRIVATE_KEY_MISSING, Data: map[string]interface{}{"err": err.Error()}}
	}
	block, _ := pem.Decode(pubKey)
	if block == nil {
		return "", errors.New("public key error!")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), ciphertext)
	if err != nil {
		return "", err
	}
	encryptedText := base64.StdEncoding.EncodeToString(encrypted)
	return encryptedText, err
}

// func RsaDecryptPKCS1v15 will work with DecryptPKCS1v15
// can refer more detail in https://gist.github.com/hothero/93c69bbd57001ce0a1997f5dd1ba89f6
func RsaDecryptPKCS1v15(encryptedText string) (string, error) {
	res, err := ioutil.ReadFile("storage/encrypt_private.pem")
	if err != nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.RSA_PRIVATE_KEY_MISSING, Data: map[string]interface{}{"err": err.Error()}}
	}
	block, _ := pem.Decode(res)
	if block == nil {
		return "", &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Data: map[string]interface{}{"err": "private key error!"}}
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	encryptedData, err := base64.StdEncoding.DecodeString(encryptedText)
	ciphertext := []byte(encryptedData)

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
