/*
Package rpc implements bridge to Lachesis full node API interface.

We recommend using local IPC for fast and the most efficient inter-process communication between the API server
and an Opera/Lachesis node. Any remote RPC connection will work, but the performance may be significantly degraded
by extra networking overhead of remote RPC calls.

You should also consider security implications of opening Lachesis RPC interface for a remote access.
If you considering it as your deployment strategy, you should establish encrypted channel between the API server
and Lachesis RPC interface with connection limited to specified endpoints.

We strongly discourage opening Lachesis RPC interface for unrestricted Internet access.
*/
package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// AccountBalance reads balance of account from Lachesis node.
func (glxy *GlxyBridge) AccountBalance(addr *common.Address) (*hexutil.Big, error) {
	// use RPC to make the call
	var balance string
	err := glxy.rpc.Call(&balance, "glxy_getBalance", addr.Hex(), "latest")
	if err != nil {
		glxy.log.Errorf("can not get balance of account [%s]", addr.Hex())
		return nil, err
	}

	// decode the response from remote server
	val, err := hexutil.DecodeBig(balance)
	if err != nil {
		glxy.log.Errorf("can not decode balance of account [%s]", addr.Hex())
		return nil, err
	}

	return (*hexutil.Big)(val), nil
}

// AccountNonce returns the total number of transaction of account from Lachesis node.
func (glxy *GlxyBridge) AccountNonce(addr *common.Address) (uint64, error) {
	// use RPC to make the call
	var nonce string
	err := glxy.rpc.Call(&nonce, "glxy_getTransactionCount", addr.Hex(), "latest")
	if err != nil {
		glxy.log.Errorf("can not get number of transaction of account [%s]", addr.Hex())
		return 0, err
	}

	// decode the response from remote server
	val, err := hexutil.DecodeUint64(nonce)
	if err != nil {
		glxy.log.Errorf("can not decode number of transaction of account [%s]", addr.Hex())
		return 0, err
	}

	return val, nil
}
