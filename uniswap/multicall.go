package uniswap

import (
	"math/big"
	"strings"

	"github.com/ackermanx/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	MultiCallAddr    = "0x5ba1e12693dc8f9c48aad8770482f4739beed696"
	MultiFragmentAbi = `[{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall2.Call[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall2.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall2.Call[]","name":"calls","type":"tuple[]"}],"name":"tryBlockAndAggregate","outputs":[{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bytes32","name":"blockHash","type":"bytes32"},{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"internalType":"struct Multicall2.Result[]","name":"returnData","type":"tuple[]"}],"stateMutability":"nonpayable","type":"function"}]`
)

type Multicall2Call struct {
	Target   common.Address
	CallData []byte
}

type Multicall2Result struct {
	Success    bool
	ReturnData *big.Int
}

func MultiCall(client *ethclient.Client, methodName string, opts *bind.CallOpts, multiCallParam []Multicall2Call) (out []interface{}, err error) {

	out = make([]interface{}, 0)
	parsedAbi, err := abi.JSON(strings.NewReader(MultiFragmentAbi))
	if err != nil {
		err = errors.Wrap(err, "parsed multi call abi")
		return
	}
	boundedContract := bind.NewBoundContract(common.HexToAddress(MultiCallAddr), parsedAbi, client, client, client)

	err = boundedContract.Call(opts, &out, methodName, multiCallParam)
	if err != nil {
		err = errors.Wrap(err, "call multi call")
	}
	return
}
