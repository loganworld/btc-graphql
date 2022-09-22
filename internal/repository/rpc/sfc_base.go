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

//go:generate tools/abigen.sh --abi ./contracts/abi/sfc-1.1.abi --pkg contracts --type SfcV1Contract --out ./contracts/sfc-v1.go
//go:generate tools/abigen.sh --abi ./contracts/abi/sfc-2.0.4-rc.2.abi --pkg contracts --type SfcV2Contract --out ./contracts/sfc-v2.go
//go:generate tools/abigen.sh --abi ./contracts/abi/sfc-3.0-rc.1.abi --pkg contracts --type SfcContract --out ./contracts/sfc-v3.go
//go:generate tools/abigen.sh --abi ./contracts/abi/sfc-tokenizer.abi --pkg contracts --type SfcTokenizer --out ./contracts/sfc_tokenizer.go

import (
	"galaxychain-graphql/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// sfcFirstLockEpoch represents the first epoch with stake locking available.
const sfcFirstLockEpoch uint64 = 1600

// SfcVersion returns current version of the SFC contract as a single number.
func (glxy *GlxyBridge) SfcVersion() (hexutil.Uint64, error) {
	// get the version information from the contract
	var ver [3]byte
	var err error
	ver, err = glxy.SfcContract().Version(nil)
	if err != nil {
		glxy.log.Criticalf("failed to get the SFC version; %s", err.Error())
		return 0, err
	}
	return hexutil.Uint64((uint64(ver[0]) << 16) | (uint64(ver[1]) << 8) | uint64(ver[2])), nil
}

// CurrentEpoch extract the current epoch id from SFC smart contract.
func (glxy *GlxyBridge) CurrentEpoch() (hexutil.Uint64, error) {
	// get the value from the contract
	epoch, err := glxy.SfcContract().CurrentEpoch(glxy.DefaultCallOpts())
	if err != nil {
		glxy.log.Errorf("failed to get the current sealed epoch: %s", err.Error())
		return 0, err
	}
	return hexutil.Uint64(epoch.Uint64()), nil
}

// CurrentSealedEpoch extract the current sealed epoch id from SFC smart contract.
func (glxy *GlxyBridge) CurrentSealedEpoch() (hexutil.Uint64, error) {
	// get the value from the contract
	epoch, err := glxy.SfcContract().CurrentSealedEpoch(glxy.DefaultCallOpts())
	if err != nil {
		glxy.log.Errorf("failed to get the current sealed epoch: %s", err.Error())
		return 0, err
	}
	return hexutil.Uint64(epoch.Uint64()), nil
}

// Epoch extract information about an epoch from SFC smart contract.
func (glxy *GlxyBridge) Epoch(id hexutil.Uint64) (*types.Epoch, error) {
	// extract epoch snapshot
	epo, err := glxy.SfcContract().GetEpochSnapshot(nil, big.NewInt(int64(id)))
	if err != nil {
		glxy.log.Errorf("failed to extract epoch information: %s", err.Error())
		return nil, err
	}

	return &types.Epoch{
		Id:                    id,
		EndTime:               hexutil.Uint64(epo.EndTime.Uint64()),
		EpochFee:              (hexutil.Big)(*epo.EpochFee),
		TotalBaseRewardWeight: (hexutil.Big)(*epo.TotalBaseRewardWeight),
		TotalTxRewardWeight:   (hexutil.Big)(*epo.TotalTxRewardWeight),
		BaseRewardPerSecond:   (hexutil.Big)(*epo.BaseRewardPerSecond),
		StakeTotalAmount:      (hexutil.Big)(*epo.TotalStake),
		TotalSupply:           (hexutil.Big)(*epo.TotalSupply),
	}, nil
}

// RewardsAllowed returns if the rewards can be manipulated with.
func (glxy *GlxyBridge) RewardsAllowed() (bool, error) {
	glxy.log.Debug("rewards lock always open")
	return true, nil
}

// LockingAllowed indicates if the stake locking has been enabled in SFC.
func (glxy *GlxyBridge) LockingAllowed() (bool, error) {
	// get the current sealed epoch value from the contract
	epoch, err := glxy.SfcContract().CurrentSealedEpoch(nil)
	if err != nil {
		glxy.log.Errorf("failed to get the current sealed epoch: %s", err.Error())
		return false, err
	}

	return epoch.Uint64() >= sfcFirstLockEpoch, nil
}

// TotalStaked returns the total amount of staked tokens.
func (glxy *GlxyBridge) TotalStaked() (*big.Int, error) {
	return glxy.SfcContract().TotalStake(glxy.DefaultCallOpts())
}

// SfcMinValidatorStake extracts a value of minimal validator self stake.
func (glxy *GlxyBridge) SfcMinValidatorStake() (*big.Int, error) {
	return glxy.SfcContract().MinSelfStake(glxy.DefaultCallOpts())
}

// SfcMaxDelegatedRatio extracts a ratio between self delegation and received stake.
func (glxy *GlxyBridge) SfcMaxDelegatedRatio() (*big.Int, error) {
	return glxy.SfcContract().MaxDelegatedRatio(glxy.DefaultCallOpts())
}

// SfcMinLockupDuration extracts a minimal lockup duration.
func (glxy *GlxyBridge) SfcMinLockupDuration() (*big.Int, error) {
	return glxy.SfcContract().MinLockupDuration(glxy.DefaultCallOpts())
}

// SfcMaxLockupDuration extracts a maximal lockup duration.
func (glxy *GlxyBridge) SfcMaxLockupDuration() (*big.Int, error) {
	return glxy.SfcContract().MaxLockupDuration(glxy.DefaultCallOpts())
}

// SfcWithdrawalPeriodEpochs extracts a minimal number of epochs between un-delegate and withdraw.
func (glxy *GlxyBridge) SfcWithdrawalPeriodEpochs() (*big.Int, error) {
	return glxy.SfcContract().WithdrawalPeriodEpochs(glxy.DefaultCallOpts())
}

// SfcWithdrawalPeriodTime extracts a minimal number of seconds between un-delegate and withdraw.
func (glxy *GlxyBridge) SfcWithdrawalPeriodTime() (*big.Int, error) {
	return glxy.SfcContract().WithdrawalPeriodTime(glxy.DefaultCallOpts())
}
