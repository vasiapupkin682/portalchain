package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	modelregistrytypes "portalchain/x/model-registry/types"
)

// modelRegistryKey returns the store key for a model record.
func modelRegistryKey(operator string) []byte {
	return []byte(modelregistrytypes.ModelRegistryPrefix + operator)
}

// GetModelRecord returns the model record for an operator, if it exists.
func (k Keeper) GetModelRecord(ctx sdk.Context, operator string) (modelregistrytypes.ModelRecord, bool) {
	store := ctx.KVStore(k.modelRegistryStoreKey)
	bz := store.Get(modelRegistryKey(operator))
	if bz == nil {
		return modelregistrytypes.ModelRecord{}, false
	}
	var record modelregistrytypes.ModelRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return modelregistrytypes.ModelRecord{}, false
	}
	return record, true
}

// SetModelRecord writes a model record to the model registry store.
func (k Keeper) SetModelRecord(ctx sdk.Context, record modelregistrytypes.ModelRecord) {
	store := ctx.KVStore(k.modelRegistryStoreKey)
	bz, err := json.Marshal(record)
	if err != nil {
		k.Logger(ctx).Error("SetModelRecord: failed to marshal", "operator", record.Operator, "err", err)
		return
	}
	store.Set(modelRegistryKey(record.Operator), bz)
}

// DeleteModelRecord removes a model record from the model registry store.
func (k Keeper) DeleteModelRecord(ctx sdk.Context, operator string) {
	store := ctx.KVStore(k.modelRegistryStoreKey)
	store.Delete(modelRegistryKey(operator))
}

// UpdateModelCategoryRep updates category-based reputation in the model registry.
// Uses EMA: new = 0.95 * old + 0.05 * normalizedScore.
// If no ModelRecord exists for the operator, the update is skipped.
func (k Keeper) UpdateModelCategoryRep(
	ctx sdk.Context,
	operator string,
	taskType string,
	normalizedScore sdk.Dec,
) {
	if taskType == "" {
		taskType = "general"
	}

	store := ctx.KVStore(k.modelRegistryStoreKey)
	bz := store.Get(modelRegistryKey(operator))
	if bz == nil {
		return
	}

	var record modelregistrytypes.ModelRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		k.Logger(ctx).Error("UpdateModelCategoryRep: failed to unmarshal model record", "operator", operator, "err", err)
		return
	}

	alpha := sdk.NewDecWithPrec(5, 2)   // 0.05
	oneMinusAlpha := sdk.OneDec().Sub(alpha)

	getOldRep := func(s string) sdk.Dec {
		if s == "" {
			s = "0.0"
		}
		d, err := sdk.NewDecFromStr(s)
		if err != nil {
			return sdk.ZeroDec()
		}
		return d
	}

	setRep := func(s *string, d sdk.Dec) {
		*s = d.String()
	}

	var oldRep sdk.Dec
	switch taskType {
	case "text":
		oldRep = getOldRep(record.RepText)
		newRep := oneMinusAlpha.Mul(oldRep).Add(alpha.Mul(normalizedScore))
		setRep(&record.RepText, newRep)
	case "code":
		oldRep = getOldRep(record.RepCode)
		newRep := oneMinusAlpha.Mul(oldRep).Add(alpha.Mul(normalizedScore))
		setRep(&record.RepCode, newRep)
	case "analysis":
		oldRep = getOldRep(record.RepAnalysis)
		newRep := oneMinusAlpha.Mul(oldRep).Add(alpha.Mul(normalizedScore))
		setRep(&record.RepAnalysis, newRep)
	default:
		oldRep = getOldRep(record.RepGeneral)
		newRep := oneMinusAlpha.Mul(oldRep).Add(alpha.Mul(normalizedScore))
		setRep(&record.RepGeneral, newRep)
	}

	bz, err := json.Marshal(record)
	if err != nil {
		k.Logger(ctx).Error("UpdateModelCategoryRep: failed to marshal model record", "operator", operator, "err", err)
		return
	}
	store.Set(modelRegistryKey(operator), bz)

	k.Logger(ctx).Info("model category reputation updated",
		"operator", operator,
		"task_type", taskType,
	)
}
