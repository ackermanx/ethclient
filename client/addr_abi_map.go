package client

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type AddrAbiMap struct {
	m sync.Map
}

func (addrAbiMap *AddrAbiMap) Delete(key common.Address) {
	addrAbiMap.m.Delete(key)
}

func (addrAbiMap *AddrAbiMap) Load(key common.Address) (value abi.ABI, ok bool) {
	v, ok := addrAbiMap.m.Load(key)
	if v != nil {
		value = v.(abi.ABI)
	}
	return
}

func (addrAbiMap *AddrAbiMap) LoadOrStore(key common.Address, value abi.ABI) (actual abi.ABI, loaded bool) {
	a, loaded := addrAbiMap.m.LoadOrStore(key, value)
	actual = a.(abi.ABI)
	return
}

func (addrAbiMap *AddrAbiMap) Store(key common.Address, value abi.ABI) {
	addrAbiMap.m.Store(key, value)
}

func (addrAbiMap *AddrAbiMap) Range(f func(key common.Address, value abi.ABI) bool) {
	f1 := func(key, value interface{}) bool {
		return f(key.(common.Address), value.(abi.ABI))
	}
	addrAbiMap.m.Range(f1)
}
