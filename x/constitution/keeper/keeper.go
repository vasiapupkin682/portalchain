package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/constitution/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	govKeeper     types.GovKeeper
	stakingKeeper types.StakingKeeper
	poiKeeper     types.PoiKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	govKeeper types.GovKeeper,
	stakingKeeper types.StakingKeeper,
	poiKeeper types.PoiKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		govKeeper:     govKeeper,
		stakingKeeper: stakingKeeper,
		poiKeeper:     poiKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// --- Params ---

func (k Keeper) GetParams(ctx sdk.Context) types.ConstitutionParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.ParamsKey))
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.ConstitutionParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.ConstitutionParams) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		panic(fmt.Errorf("failed to marshal constitution params: %w", err))
	}
	store.Set([]byte(types.ParamsKey), bz)
}

// --- Proposal Class ---

func (k Keeper) SetProposalClass(ctx sdk.Context, proposalID uint64, class types.ProposalClass) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s%d", types.ProposalClassKey, proposalID))
	store.Set(key, []byte{byte(class)})

	k.Logger(ctx).Info("proposal classified",
		"proposal_id", proposalID,
		"class", class.String(),
	)
}

func (k Keeper) GetProposalClass(ctx sdk.Context, proposalID uint64) (types.ProposalClass, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s%d", types.ProposalClassKey, proposalID))
	bz := store.Get(key)
	if bz == nil {
		return types.ClassNetworkParam, false
	}
	return types.ProposalClass(bz[0]), true
}

// --- Proposal Snapshot ---
// Snapshots of constitutional params are taken at proposal classification
// time to prevent retroactive weakening of timelock/quorum via fast proposals.

type proposalSnapshot struct {
	TimelockDays int64  `json:"timelock_days"`
	Quorum       string `json:"quorum"`
}

func (k Keeper) SetProposalSnapshot(ctx sdk.Context, proposalID uint64, timelockDays int64, quorum sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s%d", types.ProposalSnapshotKey, proposalID))
	snap := proposalSnapshot{TimelockDays: timelockDays, Quorum: quorum.String()}
	bz, err := json.Marshal(snap)
	if err != nil {
		panic(fmt.Errorf("failed to marshal proposal snapshot: %w", err))
	}
	store.Set(key, bz)
}

func (k Keeper) GetProposalSnapshot(ctx sdk.Context, proposalID uint64) (timelockDays int64, quorum sdk.Dec, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s%d", types.ProposalSnapshotKey, proposalID))
	bz := store.Get(key)
	if bz == nil {
		return 0, sdk.ZeroDec(), false
	}
	var snap proposalSnapshot
	if err := json.Unmarshal(bz, &snap); err != nil {
		return 0, sdk.ZeroDec(), false
	}
	q, err := sdk.NewDecFromStr(snap.Quorum)
	if err != nil {
		return 0, sdk.ZeroDec(), false
	}
	return snap.TimelockDays, q, true
}

// --- Voting Power Check (S3) ---
//
// SECURITY NOTE — Sybil limitation:
// CheckVotingPowerLimit uses staking's GetDelegatorBonded, which returns the
// total bonded tokens for a SINGLE address. This prevents one address from
// accumulating >15% voting power, but cannot prevent a whale from splitting
// their stake across N addresses (each under 15%).
//
// This is a fundamental constraint of permissionless blockchains: on-chain
// code cannot link addresses to the same real-world entity. Possible off-chain
// mitigations include social governance norms and identity attestations.

// CheckVotingPowerLimit enforces S3: a single address cannot hold more than
// MaxVotingPowerPercent of total bonded tokens. Applied to all governance
// messages (proposals, votes).
func (k Keeper) CheckVotingPowerLimit(ctx sdk.Context, addr sdk.AccAddress) error {
	params := k.GetParams(ctx)
	totalBonded := k.stakingKeeper.TotalBondedTokens(ctx)
	if totalBonded.IsZero() {
		return nil
	}

	delegatorBonded := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
	votingPower := sdk.NewDecFromInt(delegatorBonded).Quo(sdk.NewDecFromInt(totalBonded))

	if votingPower.GT(params.MaxVotingPowerPercent) {
		k.Logger(ctx).Error("voting power exceeded",
			"address", addr.String(),
			"voting_power", votingPower.String(),
			"max_allowed", params.MaxVotingPowerPercent.String(),
		)
		return types.ErrVotingPowerExceeded.Wrapf(
			"address %s has %.2f%% voting power, max is %.2f%%",
			addr.String(),
			votingPower.MulInt64(100).MustFloat64(),
			params.MaxVotingPowerPercent.MulInt64(100).MustFloat64(),
		)
	}

	return nil
}

// CheckValidatorConcentration rejects governance actions when any bonded
// validator holds more than MaxVotingPowerPercent of total bonded tokens.
// This is only applied to Sacred/Constitutional proposals — not NetworkParam
// — to avoid blocking all governance on small testnets where 3-5 validators
// each naturally hold >15%.
func (k Keeper) CheckValidatorConcentration(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	totalBonded := k.stakingKeeper.TotalBondedTokens(ctx)
	if totalBonded.IsZero() {
		return nil
	}

	validators := k.stakingKeeper.GetAllValidators(ctx)
	for _, val := range validators {
		if !val.IsBonded() {
			continue
		}
		valPower := sdk.NewDecFromInt(val.GetBondedTokens()).Quo(sdk.NewDecFromInt(totalBonded))
		if valPower.GT(params.MaxVotingPowerPercent) {
			k.Logger(ctx).Error("validator concentration exceeds limit",
				"validator", val.GetOperator().String(),
				"power_percent", valPower.String(),
				"max_allowed", params.MaxVotingPowerPercent.String(),
			)
			return types.ErrVotingPowerExceeded.Wrapf(
				"validator %s holds %.2f%% of total bonded tokens, exceeding %.2f%% limit; constitutional governance actions are suspended until validator set is more distributed",
				val.GetOperator().String(),
				valPower.MulInt64(100).MustFloat64(),
				params.MaxVotingPowerPercent.MulInt64(100).MustFloat64(),
			)
		}
	}
	return nil
}
