package se

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dumacp/smartcard"
	"github.com/dumacp/smartcard/nxp/mifare"
	"github.com/dumacp/smartcard/nxp/mifare/samav2"
)

const (
	AES_CBC_128 = "AES-CBC-128"
)

const (
	CEKeyNum = byte(0)
	CEKeyVer = byte(0)
)

const (
	limitKeyNumber = 128
)

type samAv2 struct {
	dev        samav2.SamAv2
	reader     smartcard.IReader
	enableKeys map[int]int
	keyCrypto  int
}

// type SamAv2 interface {
// 	Connect() error
// 	Serial() []byte
// 	Disconnect() error
// 	GenerateKey(slot int, alg string) error
// 	ImportKey(key []byte, slot int, alg string) error
// 	Decrypt(data, iv, divInput []byte, slot int) ([]byte, error)
// 	Encrypt(data, iv, divInput []byte, slot int) ([]byte, error)
// 	DumpSecretKey(slot int) ([]byte, error)
// 	// CreateEntryKey(alg string, slot, keyChangeID int, keys []byte) error
// 	// GetEntryKey(slot int) ([][]byte, []byte, error)
// }

func NewSamAV2(reader smartcard.IReader) (SE, error) {
	return &samAv2{reader: reader, enableKeys: make(map[int]int)}, nil
}

func (se *samAv2) Serial() []byte {
	uid, err := se.dev.UID()
	if err != nil {
		return nil
	}
	return uid
}

func (se *samAv2) EnableKeys() ([]int, error) {
	if len(se.enableKeys) <= 0 {
		return nil, errors.New("enabled keys is empty")
	}

	result := make([]int, 0)
	for k := range se.enableKeys {
		result = append(result, k)
	}

	return result, nil
}

func enableKeys(se samav2.SamAv2) (map[int]int, error) {
	keys := make(map[int]int)
	for keyNumber := range make([]int, 128) {
		resp, err := se.SAMGetKeyEntry(keyNumber)
		if err != nil {
			return keys, err
		}
		keyData := samav2.NewEntryKeyData(resp, samav2.AES_128)
		// log.Printf("keyData %d: %+v, [%X]", keyNumber, keyData, resp)
		//TODO: why is key to enable?
		if int(keyData.Vc) == int(2) {
			keys[keyNumber] = int(keyData.Vc)
		}
	}

	return keys, nil
}

func (se *samAv2) Connect() error {
	dev, err := se.reader.ConnectSamCard()
	if err != nil {
		return err
	}
	card := samav2.SamAV2(dev)
	// uid, _ := card.UID()
	// logs.LogBuild.Printf("sam UID: [% X]", uid)
	// atr, _ := card.ATR()
	// logs.LogBuild.Printf("sam ATR: %X", atr)
	// logs.LogBuild.Printf("sam ATR (ascii): %s", atr)

	keyMaster := make([]byte, 16)
	if _, err := card.AuthHostAV2(keyMaster, 0, 0, 0); err != nil {
		return err
	}

	se.dev = card
	se.enableKeys, err = enableKeys(card)
	if err != nil {
		return err
	}
	log.Printf("keys enable: %+v", se.enableKeys)
	return nil
}

func (se *samAv2) Disconnect() error {
	if err := se.dev.DisconnectCard(); err != nil {
		return err
	}
	return nil
}

func (se *samAv2) GenerateKey(slot int, alg string) error {
	switch {
	case strings.Contains(alg, AES_CBC_128):
		key := make([]byte, 3*16)
		if _, err := rand.Read(key); err != nil {
			return err
		}
		setConfig := samav2.SETConfigurationSettings(
			false, false, samav2.AES_128,
			false, false, false, false,
			false, false, false, false)
		extSetConfig := samav2.ExtSETConfigurationSettings(
			// samav2.OfflineCrypto_KEY|samav2.PICC_KEY|samav2.OfflineChange_KEY,
			samav2.OfflineCrypto_KEY,
			true, false)
		// logs.LogBuild.Printf("extSet: [% X], %X", extSetConfig, samav2.OfflineCrypto_KEY|samav2.PICC_KEY|samav2.OfflineChange_KEY)
		resp, err := se.dev.ChangeKeyEntry(slot, 0xFF,
			key[0:16], key[16:32], key[32:48],
			0x00,
			CEKeyNum, CEKeyVer,
			0xFF,
			0x00, 0x01, 0x02,
			extSetConfig,
			[]byte{0x00, 0x00, 0x00},
			setConfig)
		if err != nil {
			return err
		}

		if err := mifare.VerifyResponseIso7816(resp); err != nil {
			return err
		}
	default:
		return errors.New("invalid Alg")
	}
	return nil
}

func (se *samAv2) ImportKey(key []byte, slot int, alg string) error {

	if _, ok := se.enableKeys[slot]; ok {
		return fmt.Errorf("key is invalid, key already enable, slot: %d", slot)
	}

	if len(key) <= 0 {
		return fmt.Errorf("key len is invalid, key: %X", key)
	}
	if len(key)%16 != 0 {
		return fmt.Errorf("key len is invalid, key: %X", key)
	}
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	if len(keyCopy) < 3*16 {
		keyCopy = append(keyCopy, make([]byte, 3*16-len(key))...)
	}

	switch {
	case strings.Contains(alg, AES_CBC_128):

		if len(keyCopy) != 3*16 {
			return errors.New("key len is invalid")
		}
		setConfig := samav2.SETConfigurationSettings(
			false, false, samav2.AES_128,
			false, false, false, false,
			false, false, false, false)
		extSetConfig := samav2.ExtSETConfigurationSettings(
			// samav2.OfflineCrypto_KEY|samav2.PICC_KEY|samav2.OfflineChange_KEY,
			samav2.OfflineCrypto_KEY,
			true, false)
		resp, err := se.dev.ChangeKeyEntry(slot, 0xFF,
			keyCopy[0:16], keyCopy[16:32], keyCopy[32:48],
			0x00,
			CEKeyNum, CEKeyVer,
			0xFF,
			0x00, 0x01, 0x02,
			extSetConfig,
			[]byte{0x00, 0x00, 0x00},
			setConfig)
		if err != nil {
			return err
		}

		if err := mifare.VerifyResponseIso7816(resp); err != nil {
			return err
		}
	}
	return nil
}

func (se *samAv2) Decrypt(data, iv, divInput []byte, slot int) ([]byte, error) {
	if _, ok := se.enableKeys[slot]; !ok {
		return nil, ErrKeyNotEnable
	}
	if slot != se.keyCrypto {
		if _, err := se.dev.ActivateOfflineKey(slot, 0, divInput); err != nil {
			return nil, err
		}
		se.keyCrypto = slot
	}

	if _, err := se.dev.SAMLoadInitVector(samav2.AES_ALG, iv); err != nil {
		return nil, err
	}

	resp, err := se.dev.SAMDecipherOfflineData(samav2.AES_ALG, data)
	if err != nil {
		// logs.LogBuild.Printf("decipher err: %s, %X", err, data)
		return nil, err
	}
	// if err := mifare.VerifyResponseIso7816(resp); err != nil {
	// 	return nil, err
	// }
	return resp[:], nil
}

func (se *samAv2) Encrypt(data, iv, divInput []byte, slot int) ([]byte, error) {
	if _, ok := se.enableKeys[slot]; !ok {
		return nil, ErrKeyNotEnable
	}
	if slot != se.keyCrypto {
		if _, err := se.dev.ActivateOfflineKey(slot, 0, divInput); err != nil {
			return nil, err
		}
		se.keyCrypto = slot
	}

	if _, err := se.dev.SAMLoadInitVector(samav2.AES_ALG, iv); err != nil {
		return nil, err
	}

	resp, err := se.dev.SAMEncipherOfflineData(samav2.AES_ALG, data)
	if err != nil {
		return nil, err
	}

	// if err := mifare.VerifyResponseIso7816(resp); err != nil {
	// 	return nil, err
	// }
	return resp[:], nil
}

func (se *samAv2) DumpSecretKey(slot int) ([]byte, error) {
	if _, ok := se.enableKeys[slot]; !ok {
		return nil, ErrKeyNotEnable
	}
	key, err := se.dev.DumpSecretKey(slot, 0, nil)
	if err != nil {
		return nil, err
	}
	return key[:len(key)-2], nil
}

func (se *samAv2) CreateEntryKey(alg string, slot, keyChangeID int, keys []byte) error {

	switch {
	case strings.Contains(alg, AES_CBC_128):

		keys = append(keys, make([]byte, len(keys)%48)...)

		setConfig := samav2.SETConfigurationSettings(
			true, true, samav2.AES_128,
			false, false, false, false,
			false, false, false, false)
		extSetConfig := samav2.ExtSETConfigurationSettings(
			samav2.OfflineCrypto_KEY|samav2.PICC_KEY|samav2.OfflineChange_KEY,
			true, false)
		resp, err := se.dev.ChangeKeyEntry(slot, 0x00,
			keys[0:16], keys[16:32], keys[32:48],
			0x00,
			byte(keyChangeID), CEKeyVer,
			0xFF,
			0x00, 0x01, 0x02,
			extSetConfig,
			[]byte{0x00, 0x00, 0x00},
			setConfig)
		if err != nil {
			return err
		}

		if err := mifare.VerifyResponseIso7816(resp); err != nil {
			return err
		}
	}
	return nil
}

// func (se *samAv2) GetEntryKey(slot int) ([][]byte, []byte, error) {
// 	return nil, nil, nil
// }
