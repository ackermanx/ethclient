package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ackermanx/ethereum/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	testKey = "33f46d353f191f8067dc7d256e9d9ee7a2a3300649ff7c70fe1cd7e5d5237da5"
)

func main() {
	var binanceMainnet = `https://data-seed-prebsc-1-s1.binance.org:8545`

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
		client.ERC20Abi,
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
}
