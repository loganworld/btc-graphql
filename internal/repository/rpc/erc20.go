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
	"galaxychain-graphql/internal/repository/rpc/contracts"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate tools/abigen.sh --abi ./contracts/abi/erc20.abi --pkg contracts --type ERCTwenty --out ./contracts/erc20_token.go
//go:generate tools/abigen.sh --abi ./contracts/abi/WGLXY.abi --pkg contracts --type ErcWrappedFtm --out ./contracts/erc20WGLXY_token.go

// Erc20Name provides information about the name of the ERC20 token.
func (glxy *GlxyBridge) Erc20Name(token *common.Address) (string, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return "", err
	}

	// get the token name
	name, err := contract.Name(nil)
	if err != nil {
		glxy.log.Errorf("ERC20 token %s name not available; %s", token.String(), err.Error())
		return "", err
	}

	return name, nil
}

// Erc20Symbol provides information about the symbol of the ERC20 token.
func (glxy *GlxyBridge) Erc20Symbol(token *common.Address) (string, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return "", err
	}

	// get the token name
	symbol, err := contract.Symbol(nil)
	if err != nil {
		glxy.log.Errorf("ERC20 token %s symbol not available; %s", token.String(), err.Error())
		return "", err
	}

	return symbol, nil
}

// Erc20Decimals provides information about the decimals of the ERC20 token.
func (glxy *GlxyBridge) Erc20Decimals(token *common.Address) (int32, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return 0, err
	}

	// get the token name
	deci, err := contract.Decimals(nil)
	if err != nil {
		glxy.log.Errorf("ERC20 token %s decimals not available; %s", token.String(), err.Error())
		return 0, nil
	}

	return int32(deci), nil
}

// Erc20BalanceOf loads the current available balance of and ERC20 token identified by the token
// contract address for an identified owner address.
func (glxy *GlxyBridge) Erc20BalanceOf(token *common.Address, owner *common.Address) (hexutil.Big, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return hexutil.Big{}, err
	}

	// get the balance
	val, err := contract.BalanceOf(nil, *owner)
	if err != nil {
		glxy.log.Errorf("can not ERC20 %s balance for %s; %s", token.String(), owner.String(), err.Error())
		return hexutil.Big{}, err
	}

	// make sur we always have a value; at least zero
	// this should always be the case since the contract should
	// return zero even for unknown owners, but let's be sure here
	if val == nil {
		glxy.log.Errorf("no balance available for ERC20 %s, owner %s", token.String(), owner.String())
		val = new(big.Int)
	}

	// return the account balance
	return hexutil.Big(*val), nil
}

// Erc20Allowance loads the current amount of ERC20 tokens unlocked for DeFi
// contract by the token owner.
func (glxy *GlxyBridge) Erc20Allowance(token *common.Address, owner *common.Address, spender *common.Address) (hexutil.Big, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return hexutil.Big{}, err
	}

	// no spender? use fMint address by default
	if nil == spender {
		addr := glxy.fMintCfg.mustContractAddress(fMintAddressMinter)
		spender = &addr
	}

	// get the amount of tokens allowed for DeFi
	val, err := contract.Allowance(nil, *owner, *spender)
	if err != nil {
		glxy.log.Errorf("can not get defi ERC20 %s allowance for %s; %s", token.String(), owner.String(), err.Error())
		return hexutil.Big{}, err
	}

	// make sure we always have a value; at least zero
	// this should always be the case since the contract should
	// return zero even for unknown owners, but let's be sure here
	if val == nil {
		glxy.log.Errorf("no allowance available for ERC20 %s, owner %s", token.String(), owner.String())
		val = new(big.Int)
	}

	// return the account balance
	return hexutil.Big(*val), nil
}

// Erc20TotalSupply provides information about all available tokens
func (glxy *GlxyBridge) Erc20TotalSupply(token *common.Address) (hexutil.Big, error) {
	// connect the contract
	contract, err := contracts.NewERCTwenty(*token, glxy.eth)
	if err != nil {
		glxy.log.Errorf("can not contact ERC20 contract; %s", err.Error())
		return hexutil.Big{}, err
	}

	// get the amount of tokens allowed for DeFi
	val, err := contract.TotalSupply(nil)
	if err != nil {
		glxy.log.Errorf("can not get ERC20 %s total supply; %s", token.String(), err.Error())
		return hexutil.Big{}, err
	}

	// make sure we always have a value; at least zero
	if val == nil {
		glxy.log.Errorf("no supply available for ERC20 %s", token.String())
		val = new(big.Int)
	}

	// return the account balance
	return hexutil.Big(*val), nil
}
