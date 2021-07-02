package uniswap

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqrtPX96ToPrice(t *testing.T) {
	// weth/dai
	wethPrice, _ := new(big.Int).SetString("1575977176149746651316824132", 10)

	price := SqrtPriceX96ToPrice(wethPrice, true)
	assert.Equal(t, "0.00039567688472264843148441298129", price.String())
}
