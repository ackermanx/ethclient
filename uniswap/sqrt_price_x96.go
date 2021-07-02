package uniswap

import (
	"math/big"

	"github.com/shopspring/decimal"
)

var (
	X96 = decimal.NewFromInt(2).Pow(decimal.NewFromInt(96))
)

// SqrtPriceX96ToPrice convert uniswap v3 sqrt price in x96 format to decimal.Decimal
// zeroForOne true: price = token0/token1 false: price = token1/token0
func SqrtPriceX96ToPrice(sqrtPriceX96 *big.Int, zeroForOne bool) (price decimal.Decimal) {
	d := decimal.NewFromBigInt(sqrtPriceX96, 0).Div(X96)
	p := d.Mul(d)

	if !zeroForOne {
		price = decimal.NewFromInt(1).Div(p)
		return
	}
	price = p
	return
}
