package utils

import (
	"errors"
	"fmt"

	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
)


func CheckCfxAddress(chain string, addr string) (*cfxaddress.Address, error) {
	chainType, chainId, err := ChainInfoByName(chain)
	if err != nil {
		return nil, err
	}
	if chainType != CHAIN_TYPE_CFX {
		return nil, errors.New("not cfx chain")
	}
	addrItem, err := cfxaddress.NewFromBase32(addr)
	if err != nil {
		return nil, err
	}
	if addrItem.GetNetworkID() != uint32(chainId) {
		return nil, fmt.Errorf("invalid conflux network address, want %v, got %v", addrItem.GetNetworkID(), uint32(chainId))
	}
	return &addrItem, nil
}

