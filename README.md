## ethclient

ethclient is extend [go-ethereum](https://github.com/ethereum/go-ethereum) client. 

- Add `BalanceOf` for query token balance

- Add `CalculatePoolAddressV2` `CalculatePoolAddressV3` for calculate uniswap pool address offline

- Refactor `Call` for call smart contract method
## install

```
go get github.com/ackermanx/ethereum
```

## usage
Below is an example which shows some common use cases for ethclient.  Check [ethclient_test.go](https://github.com/ackermanx/ethereum/blob/main/ethereum/ethclient_test.go) for more usage.

### get balance

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/ackermanx/ethereum/client"
)

func main() {
	var binanceMainnet = `https://bsc-dataseed.binance.org`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	c, err := client.DialContext(ctx, binanceMainnet)
	cancel()
	if err != nil {
		panic(err)
	}

	// get latest height
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	blockNumber, err := c.BlockNumber(ctx)
	cancel()
	if err != nil {
		panic(err)
	}
	log.Println("latest block number: ", blockNumber)

	// get busd balance
	busdContractAddress := "0xe9e7cea3dedca5984780bafc599bd69add087d56"
	address := "0x0D022fA46e3124634c42219DF9587A91972c3930"
	balance, err := c.BalanceOf(address, busdContractAddress)
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

	"github.com/ackermanx/ethereum/uniswap"
)

func main() {
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	dai := "0x6b175474e89094c44da98b954eedeac495271d0f"
	pair, err := uniswap.CalculatePoolAddressV2(weth, dai)
	if err != nil {
		panic(err)
	}

	fmt.Printf("weth-dai pair address in uniswap v2: %s\n", pair.String())

	fee := big.NewInt(3000)
	poolAddress, err := uniswap.CalculatePoolAddressV3(weth, dai, fee)
	if err != nil {
		panic(err)
	}
	fmt.Printf("weth-dai pool address in uniswap v3: %s\n", poolAddress.String())
}

```

## JetBrains OS licenses

`ethclient` had been being developed with GoLand under the **free JetBrains Open Source license(s)** granted by JetBrains s.r.o., hence I would like to express my thanks here.

<a href="https://www.jetbrains.com/?from=ethclient" target="_blank"><img src="https://raw.githubusercontent.com/ackermanx/ethclient/main/docs/images/jetbrains/jetbrains-variant-3.svg" width="216" align="middle" align="middle"/></a>
