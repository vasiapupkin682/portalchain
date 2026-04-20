package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	modelregistrytypes "portalchain/x/model-registry/types"
)

const (
	// Blocks without tasks before decay starts (~35 days at 6s/block)
	DecayStartBlocks = int64(5000)
	// Decay applied every N blocks of inactivity
	DecayInterval = int64(1000)
	// Minimum reputation before deregister
	MinReputationThreshold = "0.0001"
	// Grace period after registration before decay applies (~70 days)
	NewAgentGracePeriod = int64(1000)
	// Assumed block time for Timestamp → block conversions (seconds)
	decayBlockTimeSeconds = int64(6)
)

// Agent describes a registered operator for decay logic (derived from model registry + reports).
type Agent struct {
	Address       string
	RegisteredAt  int64
	LastTaskBlock int64
}

// GetAllAgents returns active model-registry operators with registration and last-activity hints.
// LastTaskBlock is an estimated chain height of last activity from the latest epoch report Timestamp
// (Unix seconds); if missing, RegisteredAt is used so new agents are not treated as infinitely idle.
func (k Keeper) GetAllAgents(ctx sdk.Context) []Agent {
	store := ctx.KVStore(k.modelRegistryStoreKey)
	prefix := []byte(modelregistrytypes.ModelRegistryPrefix)
	iter := store.Iterator(prefix, sdk.PrefixEndBytes(prefix))
	defer iter.Close()

	var agents []Agent
	for ; iter.Valid(); iter.Next() {
		var record modelregistrytypes.ModelRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if !record.Active {
			continue
		}
		lastBlk := record.RegisteredAt
		if report, found := k.GetLatestReport(ctx, record.Operator); found && report.Timestamp > 0 {
			sec := ctx.BlockTime().Unix() - report.Timestamp
			if sec < 0 {
				sec = 0
			}
			lastBlk = ctx.BlockHeight() - sec/decayBlockTimeSeconds
		}
		agents = append(agents, Agent{
			Address:       record.Operator,
			RegisteredAt:  record.RegisteredAt,
			LastTaskBlock: lastBlk,
		})
	}
	k.Logger(ctx).Info("decay: agents found", "count", len(agents))
	return agents
}

// ReturnStake is a no-op here: DeregisterAgent already returns remaining stake to the operator.
// Kept for API compatibility with callers that split deregister and stake return.
func (k Keeper) ReturnStake(_ sdk.Context, _ string) {}

// ApplyReputationDecay checks all registered agents and applies reputation
// decay to those who have been inactive (no tasks) for too long.
// Decay is only applied if tasks existed in the network during the period
// to avoid punishing agents when there is simply no work available.
func (k Keeper) ApplyReputationDecay(ctx sdk.Context) {
	minRep, _ := sdk.NewDecFromStr(MinReputationThreshold)
	decayFactor := sdk.NewDecWithPrec(95, 2) // -5% per decay interval

	// Check if tasks existed in network during last DecayInterval blocks
	totalTasksInNetwork := k.GetTotalTasksInPeriod(ctx, DecayInterval)

	agents := k.GetAllAgents(ctx)
	for _, agent := range agents {
		// Grace period — don't touch newly registered agents
		blocksSinceRegister := ctx.BlockHeight() - agent.RegisteredAt
		if blocksSinceRegister < NewAgentGracePeriod {
			continue
		}

		// Check last task block
		blocksSinceTask := ctx.BlockHeight() - agent.LastTaskBlock
		if blocksSinceTask < DecayStartBlocks {
			continue // agent is active
		}

		// Only decay if tasks existed in network — not agent's fault if no work
		if totalTasksInNetwork == 0 {
			continue
		}

		// Apply decay only every DecayInterval blocks
		if blocksSinceTask%DecayInterval != 0 {
			continue
		}

		rep, found := k.GetReputation(ctx, agent.Address)
		if !found {
			continue
		}

		oldValue := rep.Value
		rep.Value = rep.Value.Mul(decayFactor)
		k.SetReputation(ctx, rep)

		k.Logger(ctx).Info("reputation decay applied",
			"agent", agent.Address,
			"old_reputation", oldValue,
			"new_reputation", rep.Value,
			"blocks_since_task", blocksSinceTask,
		)

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"reputation_decay",
			sdk.NewAttribute("agent", agent.Address),
			sdk.NewAttribute("old_reputation", oldValue.String()),
			sdk.NewAttribute("new_reputation", rep.Value.String()),
		))

		// Deregister if below threshold
		if rep.Value.LT(minRep) {
			k.Logger(ctx).Info("deregistering agent due to low reputation",
				"agent", agent.Address,
				"reputation", rep.Value,
			)
			k.DeregisterAgent(ctx, agent.Address)
			k.ReturnStake(ctx, agent.Address)
		}
	}
}

// GetTotalTasksInPeriod returns total tasks processed across all epoch reports
// whose Timestamp falls within the last `blocks` worth of wall time (using decayBlockTimeSeconds).
func (k Keeper) GetTotalTasksInPeriod(ctx sdk.Context, blocks int64) int64 {
	reports := k.GetAllEpochReports(ctx)
	cutoff := ctx.BlockTime().Unix() - blocks*decayBlockTimeSeconds
	total := int64(0)
	for _, r := range reports {
		if r.Timestamp >= cutoff && r.Timestamp > 0 {
			total += r.TasksProcessed
		}
	}
	return total
}
