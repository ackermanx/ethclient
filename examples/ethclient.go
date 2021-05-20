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
	address := common.HexToAddress("0x06b31f4E60Cc3Ed7992Fc0F60A2A1AC1060E7824")
	balance, err := client.BalanceOf(address, busdContractAddress)
	if err != nil {
		panic(err)
	}

	log.Printf("address busd balance: %s\n", balance.String())
}
