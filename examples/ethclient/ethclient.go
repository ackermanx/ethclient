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
	testKey = "e923f0f5beacbc2b7f43a85016421a2a5260f76f3c6a63c5740fdd7d1a17755f"
)

func main() {
	var binanceMainnet = `https://rinkeby.infura.io/v3/19833d6f76a04d0ca7714eea34e0d670`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	c, err := ethclient.DialContext(ctx, binanceMainnet)
	cancel()
	if err != nil {
		panic(err)
	}

	// // get latest height
	// ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	// blockNumber, err := c.BlockNumber(ctx)
	// cancel()
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println("latest block number: ", blockNumber)

	// // get busd balance
	// busdContractAddress := common.HexToAddress("0xed24fc36d5ee211ea25a80239fb8c4cfd80f12ee")
	// address := "0xe96e6b50db659935878f6f3b0491B7F192cf5F59"
	// bnbBalance, err := c.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println("bnbBalance: ", bnbBalance.String())
	// balance, err := c.BalanceOf(address, busdContractAddress.String())
	// if err != nil {
	// 	panic(err)
	// }
	// log.Printf("address busd balance: %s\n", balance.String())

	// build contract transfer
	gasFeeCap := big.NewInt(2000000000)
	gasTipCap := big.NewInt(1000000000)
	usdt := common.HexToAddress("0xe2ccf22450855c7eae0f1d15421d851ff6a95656")
	tx, err := c.BuildContractTx(
		testKey, "transfer",
		abi.ERC20Abi,
		&usdt, &bind.TransactOpts{From: common.HexToAddress("0x431beE0E54b49105964E11b9035A198A1D4735AD"), GasFeeCap: gasFeeCap, GasTipCap: gasTipCap},
		common.HexToAddress("0x550f2A264299d7958D495023E1810064289A64C8"),
		big.NewInt(3000000),
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
	// tx, err = c.BuildTransferTx(testKey, "0x38F32C2473a314d447d681D30e1C0f5D07194371", &bind.TransactOpts{
	// 	From:     common.HexToAddress("0xe96e6b50db659935878f6f3b0491b7f192cf5f59"),
	// 	Value:    big.NewInt(20000000000000000),
	// 	GasLimit: 21000,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println(tx.Hash().String())
	// err = c.SendTransaction(context.Background(), tx)
	// if err != nil {
	// 	panic(err)
	// }
}
