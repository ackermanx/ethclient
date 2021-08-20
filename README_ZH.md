🇨🇳中文 | [English](README.md)

## 简介

ethclient是对go-ethereum/client的扩展，添加uniswap v3相关工具、分层确定性钱包以及合约交互相关功能。

## 功能：

- 查询erc20代币余额

- uniswap v2/v3流动池地址离线计算

- token地址排序

- uniswap v3 path编码

- uniswap v3流动池x96格式价格转换

- 智能合约数据查询

- 智能合约/主币交易构建

- 分层确定性钱包

## 安装

```
go get -u github.com/ackermanx/ethclient
```

## 使用

下面是一些常用例子，更多使用方式可以查看[ethclient_test.go](https://github.com/ackermanx/ethereum/blob/main/ethclient_test.go)以及 examples下面的示例。

### 获取余额/代币余额/token转账/主币转账

```go
package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ackermanx/ethclient"
	"github.com/ackermanx/ethclient/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	testKey = "33f46d353f191f8067dc7d256e9d9ee7a2a3300649ff7c70fe1cd7e5d5237da5"
)

func main() {
	var binanceMainnet = `https://data-seed-prebsc-1-s1.binance.org:8545`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	c, err := ethclient.DialContext(ctx, binanceMainnet)
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
	busdContractAddress := common.HexToAddress("0xed24fc36d5ee211ea25a80239fb8c4cfd80f12ee")
	address := "0xe96e6b50db659935878f6f3b0491B7F192cf5F59"
	bnbBalance, err := c.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		panic(err)
	}
	log.Println("bnbBalance: ", bnbBalance.String())
	balance, err := c.BalanceOf(address, busdContractAddress.String())
	if err != nil {
		panic(err)
	}
	log.Printf("address busd balance: %s\n", balance.String())

	// build contract transfer
	tx, err := c.BuildContractTx(
		testKey, "transfer",
		abi.ERC20Abi,
		&busdContractAddress, &bind.TransactOpts{From: common.HexToAddress(address)},
		common.HexToAddress("0x38F32C2473a314d447d681D30e1C0f5D07194371"),
		big.NewInt(100000000000000000),
	)
	if err != nil {
		panic(err)
	}
	err = c.SendTransaction(context.Background(), tx)
	if err != nil {
		panic(err)
	}
	log.Println(tx.Hash().String())

	// send bnb
	tx, err = c.BuildTransferTx(testKey, "0x38F32C2473a314d447d681D30e1C0f5D07194371", &bind.TransactOpts{
		From:     common.HexToAddress("0xe96e6b50db659935878f6f3b0491b7f192cf5f59"),
		Value:    big.NewInt(20000000000000000),
		GasLimit: 21000,
	})
	if err != nil {
		panic(err)
	}
	log.Println(tx.Hash().String())
	err = c.SendTransaction(context.Background(), tx)
	if err != nil {
		panic(err)
	}
}
```

### 离线生成uniswap v3流动池地址

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/ackermanx/ethclient/uniswap"
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

## JetBrains 开源证书支持

`ethereum` 项目是在 JetBrains 公司旗下的 GoLand 集成开发环境中进行开发，基于 free JetBrains Open Source license(s) 正版免费授权，在此表达我的谢意。

<a href="https://www.jetbrains.com/?from=ethereum" target="_blank"><img src="https://raw.githubusercontent.com/ackermanx/ethereum/main/docs/images/jetbrains/jetbrains-variant-3.svg" width="216" align="middle" align="middle"/></a>