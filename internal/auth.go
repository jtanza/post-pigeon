package internal

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
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

// TODO
func Fingerprint(pubKey string) (string, error) {
	parts := strings.Split(pubKey, "\n")
	base64Parts := parts[1 : len(parts)-1]

	key, err := base64.StdEncoding.DecodeString(strings.Join(base64Parts, ""))
	if err != nil {
		return "", fmt.Errorf("can not parse key %s", err)
	}

	s := sha256.New()
	s.Write(key)
	return base64.RawURLEncoding.EncodeToString(s.Sum(nil)), nil
}
