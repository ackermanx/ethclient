package swap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	AddrSize = 20
	FeeSize  = 3
	Offset   = AddrSize + FeeSize
	DataSize = Offset + AddrSize
)

// EncodePath encode path to bytes
func EncodePath(path []common.Address, fees []int) (encoded []byte, err error) {
	if len(path) != len(fees)+1 {
		err = errors.New("path/fee lengths do not match")
		return
	}
	encoded = make([]byte, 0, len(fees)*Offset+AddrSize)
	for i := 0; i < len(fees); i++ {
		encoded = append(encoded, path[i].Bytes()...)
		feeBytes := big.NewInt(int64(fees[i])).Bytes()
		feeBytes = common.LeftPadBytes(feeBytes, 3)
		encoded = append(encoded, feeBytes...)
	}
	encoded = append(encoded, path[len(path)-1].Bytes()...)
	return
}
