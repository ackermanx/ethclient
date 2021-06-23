package main

import (
	"fmt"
	"math/big"

	"github.com/ackermanx/ethclient/swap"
)

func main() {
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	dai := "0x6b175474e89094c44da98b954eedeac495271d0f"
	pair, err := swap.CalculatePoolAddressV2(weth, dai)
	if err != nil {
		panic(err)
	}

	fmt.Printf("weth-dai pair address in uniswap v2: %s\n", pair.String())

	fee := big.NewInt(3000)
	poolAddress, err := swap.CalculatePoolAddressV3(weth, dai, fee)
	if err != nil {
		panic(err)
	}
	fmt.Printf("weth-dai pool address in uniswap v3: %s\n", poolAddress.String())
}
