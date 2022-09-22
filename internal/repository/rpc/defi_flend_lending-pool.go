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
	"galaxychain-graphql/internal/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate tools/abigen.sh --abi ./contracts/abi/defi-flend-ilending-pool.abi --pkg contracts --type iLendingPool --out ./contracts/defi-flend-ilending-pool.go

// fLendConfig represents the configuration for DeFi fLend module.
type fLendConfig struct {
	// bridge represents the reference to the instantiated RPC bridge
	bridge *GlxyBridge

	// lendigPoolAddress represents the address
	// of the fLend LendingPool contract
	lendigPoolAddress common.Address
}

// FLendGetLendingPool resolves Lending pool contract instance
func (glxy *GlxyBridge) FLendGetLendingPool() (*contracts.ILendingPool, error) {
	// get the lending pool contract
	lp, err := contracts.NewILendingPool(glxy.fLendCfg.lendigPoolAddress, glxy.eth)
	if err != nil {
		glxy.log.Errorf("Can not get lending pool contract on address %s; %s", glxy.fLendCfg.lendigPoolAddress.String(), err.Error())
		return nil, err
	}
	return lp, nil
}

// FLendGetLendingPoolReserveData resolves reserve data
func (glxy *GlxyBridge) FLendGetLendingPoolReserveData(assetAddress *common.Address) (*types.ReserveData, error) {

	// get the lending pool contract
	lp, err := glxy.FLendGetLendingPool()
	if err != nil {
		glxy.log.Errorf("Can not access lending pool %s", err.Error())
		return nil, err
	}

	rd, err := lp.GetReserveData(&bind.CallOpts{}, *assetAddress)
	if err != nil {
		glxy.log.Errorf("Cannot get reserve data for asset %s: %s", assetAddress.String(), err.Error())
		return nil, err
	}

	rdt := &types.ReserveData{
		AssetAddress:                *assetAddress,
		ID:                          int32(rd.Id),
		Configuration:               hexutil.Big(*rd.Configuration.Data),
		LiquidityIndex:              hexutil.Big(*rd.LiquidityIndex),
		VariableBorrowIndex:         hexutil.Big(*rd.VariableBorrowIndex),
		CurrentLiquidityRate:        hexutil.Big(*rd.CurrentLiquidityRate),
		CurrentVariableBorrowRate:   hexutil.Big(*rd.CurrentVariableBorrowRate),
		CurrentStableBorrowRate:     hexutil.Big(*rd.CurrentStableBorrowRate),
		LastUpdateTimestamp:         hexutil.Big(*rd.LastUpdateTimestamp),
		ATokenAddress:               rd.ATokenAddress,
		StableDebtTokenAddress:      rd.StableDebtTokenAddress,
		VariableDebtTokenAddress:    rd.VariableDebtTokenAddress,
		InterestRateStrategyAddress: rd.InterestRateStrategyAddress,
	}

	return rdt, nil
}

// FLendGetReserveList resolves list of reserve addresses
func (glxy *GlxyBridge) FLendGetReserveList() ([]common.Address, error) {

	// get the lending pool contract
	lp, err := glxy.FLendGetLendingPool()
	if err != nil {
		glxy.log.Errorf("Can not access lending pool %s", err.Error())
		return nil, err
	}

	rl, err := lp.GetReservesList(&bind.CallOpts{})
	if err != nil {
		glxy.log.Errorf("Cannot get reserves list: %s", err.Error())
		return nil, err
	}
	return rl, nil
}

// FLendGetUserAccountData resolves user account data for fLend
func (glxy *GlxyBridge) FLendGetUserAccountData(userAddress *common.Address) (*types.FLendUserAccountData, error) {

	// get the lending pool contract
	lp, err := glxy.FLendGetLendingPool()
	if err != nil {
		glxy.log.Errorf("Can not access lending pool %s", err.Error())
		return nil, err
	}

	ua, err := lp.GetUserAccountData(&bind.CallOpts{}, *userAddress)
	if err != nil {
		glxy.log.Errorf("Cannot get user account data for address %s: %s", userAddress.String(), err.Error())
		return nil, err
	}

	uc, err := lp.GetUserConfiguration(&bind.CallOpts{}, *userAddress)
	if err != nil {
		glxy.log.Errorf("Cannot get user account configuration data for address %s: %s", userAddress.String(), err.Error())
		return nil, err
	}

	uad := &types.FLendUserAccountData{
		TotalCollateralFUSD:         hexutil.Big(*ua.TotalCollateralETH),
		TotalDebtFUSD:               hexutil.Big(*ua.TotalDebtETH),
		AvailableBorrowsFUSD:        hexutil.Big(*ua.AvailableBorrowsETH),
		CurrentLiquidationThreshold: hexutil.Big(*ua.CurrentLiquidationThreshold),
		Ltv:                         hexutil.Big(*ua.Ltv),
		HealthFactor:                hexutil.Big(*ua.HealthFactor),
		ConfigurationData:           hexutil.Big(*uc.Data),
	}
	return uad, nil
}

// FLendGetUserDepositHistory resolves deposit event history data for specified user and asset address
func (glxy *GlxyBridge) FLendGetUserDepositHistory(userAddress *common.Address, assetAddress *common.Address) ([]*types.FLendDeposit, error) {
	// create user filter
	userFilter := make([]common.Address, 0)
	if userAddress != nil {
		userFilter = append(userFilter, *userAddress)
	}

	// create asset filter
	assetFilter := make([]common.Address, 0)
	if assetAddress != nil {
		assetFilter = append(assetFilter, *assetAddress)
	}

	// get the lending pool contract
	lp, err := glxy.FLendGetLendingPool()
	if err != nil {
		glxy.log.Errorf("Can not access lending pool %s", err.Error())
		return nil, err
	}

	// filter logs
	fdi, err := lp.FilterDeposit(&bind.FilterOpts{}, assetFilter, userFilter, []uint16{0})
	if err != nil {
		glxy.log.Errorf("can not filter lending pool deposit logs: %s", err.Error())
		return nil, err
	}

	// results array
	depositArray := make([]*types.FLendDeposit, 0)

	// iterate thru filtered logs
	for fdi.Next() {
		// get block for timestamp information
		blkHash := fdi.Event.Raw.BlockHash.String()
		blk, err := glxy.BlockByHash(&blkHash)
		if err != nil {
			glxy.log.Errorf("fLend block with hash %s was not found: %s", blkHash, err.Error())
			continue
		}

		// add deposit event data to results
		depositArray = append(depositArray, &types.FLendDeposit{
			AssetAddress:      fdi.Event.Reserve,
			UserAddress:       fdi.Event.User,
			OnBehalfOfAddress: fdi.Event.OnBehalfOf,
			Amount:            hexutil.Big(*fdi.Event.Amount),
			ReferralCode:      int32(byte(fdi.Event.Referral)),
			Timestamp:         blk.TimeStamp,
		})
	}
	return depositArray, nil
}
