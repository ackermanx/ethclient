package wallet

import (
	"encoding/hex"
	"testing"
)

func TestAes(t *testing.T) {
	oriData := "hello world"
	key := "123443211234432112341234123412"
	encryptData, err := AesCBCEncrypt([]byte(oriData), []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	encryptDataHex := hex.EncodeToString(encryptData)
	t.Logf("aes crypto content: %s", encryptDataHex)

	encryptData, err = hex.DecodeString(encryptDataHex)
	if err != nil {
		t.Fatal(err)
	}

	decodeData, err := AesCBCDecrypt(encryptData, []byte(key))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(decodeData))
}
