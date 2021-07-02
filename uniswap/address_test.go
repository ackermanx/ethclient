package uniswap

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePoolAddressV2(t *testing.T) {
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	dai := "0x6b175474e89094c44da98b954eedeac495271d0f"
	pair, err := CalculatePoolAddressV2(weth, dai)

	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}
	assert.Equal(t, common.HexToAddress("0xA478c2975Ab1Ea89e8196811F51A7B7Ade33eB11"), pair)
}

func TestCalculatePoolAddressV3(t *testing.T) {
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	dai := "0x6b175474e89094c44da98b954eedeac495271d0f"
	fee := big.NewInt(3000)
	pair, err := CalculatePoolAddressV3(weth, dai, fee)

	if !assert.Equal(t, nil, err) {
		t.FailNow()
	}
	assert.Equal(t, common.HexToAddress("0xC2e9F25Be6257c210d7Adf0D4Cd6E3E881ba25f8"), pair)
}
