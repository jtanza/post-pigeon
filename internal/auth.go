package internal

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

func ValidateSignature(rawPubKey string, base64EncodedSignature string, message string) error {
	block, _ := pem.Decode([]byte(rawPubKey))
	if block == nil {
		return errors.New("invalid PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	pubKey := key.(*ecdsa.PublicKey)

	signature, err := base64.StdEncoding.DecodeString(base64EncodedSignature)
	if err != nil {
		return err
	}

	hash := sha1.Sum([]byte(message))
	if !ecdsa.VerifyASN1(pubKey, hash[:], signature) {
		return errors.New("invalid signature")
	}

	return nil
}

func Fingerprint(rawPubKey string) (string, error) {
	block, _ := pem.Decode([]byte(rawPubKey))
	if block == nil {
		return "", errors.New("invalid PEM block")
	}

	s := sha256.New()
	s.Write(block.Bytes)

	return base64.URLEncoding.EncodeToString(s.Sum(nil)), nil
}
