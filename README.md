## ethclient

ethclient is extend [go-ethereum](https://github.com/ethereum/go-ethereum) client. 

- Add `BalanceOf` for query token balance

- Add `CalculatePoolAddressV2` `CalculatePoolAddressV3` for calculate uniswap pool address offline

## install

```
go get github.com/ackermanx/ethclient
```

## usage
Below is an example which shows some common use cases for ethclient.  Check [ethclient_test.go](https://github.com/ackermanx/ethclient/blob/main/ethclient/ethclient_test.go) for more usage.

### get balance

```go
package main

import (
	"context"
	"log"
	"time"
	"github.com/ackermanx/ethclient/ethclient"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	var binanceMainnet = `https://bsc-dataseed.binance.org`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	client, err := ethclient.DialContext(ctx, binanceMainnet)
	cancel()
	if err != nil {
		panic(err)
	}

	// get latest height
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	blockNumber, err := client.BlockNumber(ctx)
	cancel()
	if err != nil {
		panic(err)
	}
	log.Println("latest block number: ", blockNumber)

	// get busd balance
	busdContractAddress := common.HexToAddress("0xe9e7cea3dedca5984780bafc599bd69add087d56")
	address := common.HexToAddress("0x0D022fA46e3124634c42219DF9587A91972c3930")
	balance, err := client.BalanceOf(address, busdContractAddress)
	if err != nil {
		panic(err)
	}
	
	log.Printf("address busd balance: %s\n", balance.String())
}
```

### generate pool address offline

```go
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

```