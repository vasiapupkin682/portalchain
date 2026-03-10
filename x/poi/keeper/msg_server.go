package keeper

import (
	"context"
	"crypto/sha256"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.FullMsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.FullMsgServer = &msgServer{}

func (k msgServer) SubmitEpochReport(goCtx context.Context, msg *types.MsgSubmitEpochReport) (*types.MsgSubmitEpochReportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	report := types.EpochReport{
		Epoch:            msg.Epoch,
		Validator:        msg.Validator,
		TasksProcessed:   msg.TasksProcessed,
		WeightedTaskSum:  msg.WeightedTaskSum,
		AvgLatency:       msg.AvgLatency,
		Reliability:      msg.Reliability,
		SamplingFailures: msg.SamplingFailures,
		Timestamp:        msg.Timestamp,
	}

	k.SetEpochReport(ctx, report)
	k.UpdateReputation(ctx, report)

	epochStr := strconv.FormatInt(msg.Epoch, 10)
	k.Logger(ctx).Info("SAMPLING CHECK START", "epoch", msg.Epoch)

	samplingResult := k.ShouldSample(ctx, []byte(epochStr))

	k.Logger(ctx).Info("SAMPLING CHECK RESULT",
		"epoch", msg.Epoch,
		"result", samplingResult,
	)

	if samplingResult {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("sampling_selected",
				sdk.NewAttribute("epoch", epochStr),
				sdk.NewAttribute("validator", msg.Validator),
			),
		)
		k.Logger(ctx).Info("SAMPLING TRIGGERED",
			"epoch", msg.Epoch,
			"validator", msg.Validator,
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_epoch_report",
			sdk.NewAttribute("epoch", epochStr),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("tasks_processed", strconv.FormatInt(msg.TasksProcessed, 10)),
			sdk.NewAttribute("reliability", msg.Reliability.String()),
		),
	)

	return &types.MsgSubmitEpochReportResponse{}, nil
}

// RemoveAgent handles governance-initiated agent removal with mandatory
// agent consent (S1 sacred principle). Only the gov module account can
// submit this message, and the agent must have signed a consent payload.
func (k msgServer) RemoveAgent(goCtx context.Context, msg *types.MsgRemoveAgent) (*types.MsgRemoveAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. Verify authority is the gov module account.
	govAddr := k.accountKeeper.GetModuleAddress("gov")
	if msg.Authority != govAddr.String() {
		return nil, types.ErrGovAuthority.Wrapf(
			"expected %s, got %s", govAddr.String(), msg.Authority,
		)
	}

	// 2. Check the agent record exists.
	if !k.HasReputation(ctx, msg.AgentAddress) {
		return nil, types.ErrAgentNotFound.Wrapf(
			"no reputation record for %s", msg.AgentAddress,
		)
	}

	// 3. Verify agent consent signature.
	agentAddr, err := sdk.AccAddressFromBech32(msg.AgentAddress)
	if err != nil {
		return nil, err
	}

	account := k.accountKeeper.GetAccount(ctx, agentAddr)
	if account == nil {
		return nil, types.ErrAgentNotFound.Wrapf(
			"account not found for %s", msg.AgentAddress,
		)
	}

	pubKey := account.GetPubKey()
	if pubKey == nil {
		return nil, types.ErrAgentConsentInvalid.Wrap(
			"agent account has no public key on-chain; agent must send at least one transaction first",
		)
	}

	// Check consent expiry to prevent replay attacks.
	consentExpiry := time.Unix(msg.ConsentExpiry, 0)
	if ctx.BlockTime().After(consentExpiry) {
		return nil, types.ErrAgentConsentInvalid.Wrapf(
			"agent consent expired at %s, current block time is %s",
			consentExpiry.UTC().String(), ctx.BlockTime().UTC().String(),
		)
	}

	// Consent payload includes expiry to bind the signature to a time window.
	// payload = SHA256(AgentAddress + Authority + Reason + ChainID + ConsentExpiry)
	expiryStr := strconv.FormatInt(msg.ConsentExpiry, 10)
	payload := sha256.Sum256([]byte(msg.AgentAddress + msg.Authority + msg.Reason + ctx.ChainID() + expiryStr))

	if !pubKey.VerifySignature(payload[:], msg.AgentConsent) {
		return nil, types.ErrAgentConsentInvalid.Wrap(
			"agent consent signature does not match",
		)
	}

	// 4. Delete the reputation record.
	k.DeleteReputation(ctx, msg.AgentAddress)

	// 5. Emit event.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_removed",
			sdk.NewAttribute("agent", msg.AgentAddress),
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	k.Logger(ctx).Info("agent removed via governance with consent",
		"agent", msg.AgentAddress,
		"reason", msg.Reason,
	)

	return &types.MsgRemoveAgentResponse{}, nil
}
