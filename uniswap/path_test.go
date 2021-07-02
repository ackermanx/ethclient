package uniswap

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestEncodePath(t *testing.T) {
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	dai := "0x6b175474e89094c44da98b954eedeac495271d0f"
	path := []common.Address{common.HexToAddress(weth), common.HexToAddress(dai)}
	encoded, err := EncodePath(path, []int{3000})
	if err != nil {
		t.Fatal(err)
	}
	encodedPathHex := hex.EncodeToString(encoded)
	assert.Equal(t, "c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000bb86b175474e89094c44da98b954eedeac495271d0f", encodedPathHex)
}
