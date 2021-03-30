package pkg

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	decodedInitCode []byte
)

func GetCreate2Address(factory, token0, token1 string) (pairAddress common.Address, err error) {
	factoryAddr := common.HexToAddress(factory)
	tkn0, tkn1 := sortAddressess(common.HexToAddress(token0), common.HexToAddress(token1))

	msg := []byte{255}
	msg = append(msg, factoryAddr.Bytes()...)
	addrBytes := tkn0.Bytes()
	addrBytes = append(addrBytes, tkn1.Bytes()...)
	msg = append(msg, crypto.Keccak256(addrBytes)...)

	// initCodeHash
	if len(decodedInitCode) == 0 {
		b, err1 := hex.DecodeString("96e8ac4277198ff8b6f785478aa9a39f403cb768dd02cbee326c3e7da348845f")
		if err1 != nil {
			err = errors.Wrap(err, "decode init code hash")
			return
		}
		decodedInitCode = b
	}

	msg = append(msg, decodedInitCode...)
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
