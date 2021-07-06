package client

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestAbi(t *testing.T) {
	addr := common.HexToAddress("0x431beE0E54b49105964E11b9035A198A1D4735AD")
	m := AddrAbiMap{}
	parsedAbi, err := abi.JSON(strings.NewReader(ERC20Abi))
	if err != nil {
		t.Fatal(err)
	}

	m.Store(addr, parsedAbi)

	v, ok := m.Load(addr)
	if !ok {
		t.Log("value not exsit")
		t.FailNow()
	}
	assert.Equal(t, parsedAbi, v)
}
