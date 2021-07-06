package main

import (
	"context"
	"log"
	"time"

	"github.com/ackermanx/ethclient/ethclient"
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
	busdContractAddress := "0xe9e7cea3dedca5984780bafc599bd69add087d56"
	address := "0x0D022fA46e3124634c42219DF9587A91972c3930"
	balance, err := client.BalanceOf(address, busdContractAddress)
	if err != nil {
		panic(err)
	}

	log.Printf("address busd balance: %s\n", balance.String())
}
