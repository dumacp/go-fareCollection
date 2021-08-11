package se

import "errors"

type SE interface {
	Connect() error
	Disconnect() error
	Serial() []byte
	GenerateKey(slot int, alg string) error
	ImportKey(key []byte, slot int, alg string) error
	Decrypt(data, iv, divInput []byte, slot int) ([]byte, error)
	Encrypt(data, iv, divInput []byte, slot int) ([]byte, error)
	DumpSecretKey(slot int) ([]byte, error)
	EnableKeys() ([]int, error)
}

var ErrKeyNotEnable = errors.New("key not enable")
