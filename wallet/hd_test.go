package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestDeriveAddressFromPathAndSeed(t *testing.T) {
	seed := bip39.NewSeed("foo", "")

	encryptSeed, err := AesCBCEncrypt([]byte(seed), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	encryptSeedInHex := hex.EncodeToString(encryptSeed)

	assert.Equal(
		t,
		"7704d7c96a90c8978994013fc79c7fc37c889719679a84976a5b7a5b46ec30201c3ee04288d697010fc9cf81e327d8734c8f30dbc1a638fc9e25f29d42653c84d764bd790bbed510e9270d7b3cfa205d",
		encryptSeedInHex)

	path := "m/44'/60'/0'/0/1"
	wallet, err := NewFromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	derivaPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		t.Fatal(err)
	}
	addr, err := wallet.deriveAddress(derivaPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "0x38F32C2473a314d447d681D30e1C0f5D07194371", addr.String())
}
