package pkg

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestGetCreate2Address(t *testing.T) {
	factory := "0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"
	token0 := "0xc7ad46e0b8a400bb3c915120d284aafba8fc4735"
	token1 := "0xd1822505796C4eba9379D5a8B4141573444042c6"
	pair, err := GetCreate2Address(factory, token0, token1)

	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}
	assert.Equal(t, common.HexToAddress("0x0E47Bf6489C12470f73681d6B3f20789295968D9"), pair)
}
