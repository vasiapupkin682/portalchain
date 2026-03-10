package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

func (k Keeper) SetReputation(ctx sdk.Context, rep types.Reputation) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ReputationPrefix + rep.Validator)
	bz := k.cdc.MustMarshal(&rep)
	store.Set(key, bz)
}

func (k Keeper) GetReputation(ctx sdk.Context, validator string) (types.Reputation, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ReputationPrefix + validator)
	bz := store.Get(key)
	if bz == nil {
		return types.Reputation{
			Validator: validator,
			Value:     sdk.ZeroDec(),
		}, false
	}
	var rep types.Reputation
	k.cdc.MustUnmarshal(bz, &rep)
	return rep, true
}

func (k Keeper) DeleteReputation(ctx sdk.Context, validator string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ReputationPrefix + validator)
	store.Delete(key)
}

func (k Keeper) HasReputation(ctx sdk.Context, validator string) bool {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ReputationPrefix + validator)
	return store.Has(key)
}

func (k Keeper) GetAllReputations(ctx sdk.Context) []types.Reputation {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ReputationPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var reputations []types.Reputation
	for ; iterator.Valid(); iterator.Next() {
		var rep types.Reputation
		k.cdc.MustUnmarshal(iterator.Value(), &rep)
		reputations = append(reputations, rep)
	}
	return reputations
}
