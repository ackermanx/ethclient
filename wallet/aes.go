package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

func AesCBCEncrypt(data, key []byte) (encrypted []byte, err error) {

	if len(key) == 0 || len(key) > 32 {
		err = errors.New("key is empty or length is bigger than 32, key length must in (0,32]")
		return
	}

	if len(key) <= 16 {
		key = paddingLeft(key, '0', 16)
	} else if len(key) <= 24 {
		key = paddingLeft(key, '0', 24)
	} else {
		key = paddingLeft(key, '0', 32)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		// key length must be 16/24/32
		err = errors.Wrapf(err, "new cipher error, key length: %d", len(key))
		return
	}

	// get block length
	blockSize := block.BlockSize()
	// padding data
	data = PKCS7Padding(data, blockSize)
	// crypt mode
	iv := key[:blockSize]
	blockMode := cipher.NewCBCEncrypter(block, iv)
	// create slice
	encrypted = make([]byte, len(data))
	blockMode.CryptBlocks(encrypted, data)
	return
}

func AesCBCDecrypt(encryptedData, key []byte) (data []byte, err error) {

	if len(key) == 0 || len(key) > 32 {
		err = errors.New("key is empty or length is bigger than 32, key length must in (0,32]")
		return
	}

	if len(key) <= 16 {
		key = paddingLeft(key, '0', 16)
	} else if len(key) <= 24 {
		key = paddingLeft(key, '0', 24)
	} else {
		key = paddingLeft(key, '0', 32)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		// key length must be 16/24/32
		err = errors.Wrapf(err, "new cipher error, key length: %d", len(key))
		return
	}

	iv := key[:block.BlockSize()]
	// crypt mode
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// create slice
	data = make([]byte, len(encryptedData))
	blockMode.CryptBlocks(data, encryptedData)
	data = PKCS7UnPadding(data)

	return
}

func paddingLeft(ori []byte, pad byte, length int) []byte {
	if len(ori) >= length {
		return ori[:length]
	}
	pads := bytes.Repeat([]byte{pad}, length-len(ori))
	return append(pads, ori...)
}

func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unPadding := int(plantText[length-1])
	return plantText[:(length - unPadding)]
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
