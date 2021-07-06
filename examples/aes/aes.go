package main

import (
	"encoding/hex"
	"log"

	"github.com/ackermanx/ethereum/wallet"
)

func main() {
	oriData := "hello world"
	key := "123443211234432112341234123412"
	encryptData, err := wallet.AesCBCEncrypt([]byte(oriData), []byte(key))
	if err != nil {
		log.Fatal(err)
	}
	encryptDataHex := hex.EncodeToString(encryptData)
	log.Printf("aes crypto content: %s", encryptDataHex)

	encryptData, err = hex.DecodeString(encryptDataHex)
	if err != nil {
		log.Fatal(err)
	}

	decodeData, err := wallet.AesCBCDecrypt(encryptData, []byte(key))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(decodeData))
}
