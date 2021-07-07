package tests

import (
	"crypto/sha1"
	"strings"
	"testing"

	"github.com/ackermanx/ethereum/client"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// go test -bench=.  -cpu=4 -benchmem=true -run=none -memprofile mem.out -cpuprofile=cpu.prof
// 查看mem.out pprof -http=:8080  mem.out
func BenchmarkAddressToString(b *testing.B) {
	m := make(map[common.Address]abi.ABI)
	r := strings.NewReader(client.ERC20Abi)

	a := common.HexToAddress("0x431beE0E54b49105964E11b9035A198A1D4735AD")
	for i := 0; i < b.N; i++ {
		parsedAbi, ok := m[a]
		if !ok {
			p, err := abi.JSON(r)
			if err != nil {
				b.Fatal(err)
			}
			m[a] = p
			parsedAbi = p
		}
		_ = parsedAbi
	}
}

func BenchmarkSha1(b *testing.B) {
	m := make(map[string]abi.ABI)

	for i := 0; i < b.N; i++ {
		s := sha1.New()
		s.Write([]byte(client.ERC20Abi))
		d := s.Sum(nil)
		key := string(d)
		parsedAbi, ok := m[key]
		if !ok {
			p, err := abi.JSON(strings.NewReader(client.ERC20Abi))
			if err != nil {
				b.Fatal(err)
			}
			m[string(key)] = p
			parsedAbi = p
		}
		_ = parsedAbi
	}
}

func BenchmarkAbiJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parsedAbi, err := abi.JSON(strings.NewReader(client.ERC20Abi))
		if err != nil {
			b.Fatal(err)
		}
		_ = parsedAbi
	}
}
