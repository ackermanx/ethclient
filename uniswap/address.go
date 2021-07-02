package uniswap

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const (
	FactoryAddrV3 = "0x1F98431c8aD98523631AE4a59f267346ea31F984"
	FactoryAddrV2 = "0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f"
)

var (
	Address, _       = abi.NewType("address", "", nil)
	Uint24, _        = abi.NewType("uint24", "", nil)
	saltAbiArguments = abi.Arguments{
		abi.Argument{
			Name:    "token0",
			Type:    Address,
			Indexed: false,
		},
		abi.Argument{
			Name:    "token1",
			Type:    Address,
			Indexed: false,
		},
		abi.Argument{
			Name:    "fee",
			Type:    Uint24,
			Indexed: false,
		},
	}
	PoolInitCodeV3, _ = hex.DecodeString("e34f199b19b2b4f47f68442619d555527d244f78a3297ea89325f843f87b8b54")
	PoolInitCodeV2, _ = hex.DecodeString("96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f")
)

// CalculatePoolAddressV2 calculate uniswapV2 pool address offline from pool tokens
func CalculatePoolAddressV2(token0, token1 string) (pairAddress common.Address, err error) {
	factoryAddr := common.HexToAddress(FactoryAddrV2)
	tkn0, tkn1 := sortAddressess(common.HexToAddress(token0), common.HexToAddress(token1))

	msg := []byte{255}
	msg = append(msg, factoryAddr.Bytes()...)
	addrBytes := tkn0.Bytes()
	addrBytes = append(addrBytes, tkn1.Bytes()...)
	msg = append(msg, crypto.Keccak256(addrBytes)...)

	msg = append(msg, PoolInitCodeV2...)
	hash := crypto.Keccak256(msg)
	pairAddressBytes := big.NewInt(0).SetBytes(hash)
	pairAddressBytes = pairAddressBytes.Abs(pairAddressBytes)
	return common.BytesToAddress(pairAddressBytes.Bytes()), nil
}

func sortAddressess(tkn0, tkn1 common.Address) (common.Address, common.Address) {
	token0Rep := new(big.Int).SetBytes(tkn0.Bytes())
	token1Rep := new(big.Int).SetBytes(tkn1.Bytes())

	if token0Rep.Cmp(token1Rep) > 0 {
		tkn0, tkn1 = tkn1, tkn0
	}

	return tkn0, tkn1
}

// CalculatePoolAddressV3 calculate uniswapV3 pool address offline from pool tokens and fee
func CalculatePoolAddressV3(tokenA, tokenB string, fee *big.Int) (poolAddress common.Address, err error) {
	tkn0, tkn1 := sortAddressess(common.HexToAddress(tokenA), common.HexToAddress(tokenB))
	paramsPacked, err := saltAbiArguments.Pack(tkn0, tkn1, fee)
	if err != nil {
		err = errors.Wrap(err, "pack arguments")
		return
	}

	salt := crypto.Keccak256(paramsPacked)
	// "0xff"
	msg := []byte{255}
	msg = append(msg, common.HexToAddress(FactoryAddrV3).Bytes()...)
	msg = append(msg, salt...)
	msg = append(msg, PoolInitCodeV3...)

	hash := crypto.Keccak256(msg)
	return common.BytesToAddress(hash[12:]), nil
}
