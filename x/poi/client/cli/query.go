package cli

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"portalchain/x/poi/types"
)

func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		Long:                       "Query commands for the Proof of Intelligence (poi) module — inspect validator reputations and epoch reports.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryReputation(),
		CmdQueryReport(),
		CmdQueryReports(),
		CmdQuerySamplingRecord(),
		CmdQueryPendingSamplings(),
		CmdQueryParams(),
	)

	return cmd
}

func CmdQueryReputation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reputation [validator-addr]",
		Short: "Query the reputation score for a validator",
		Long: `Query the current Proof-of-Intelligence reputation value for a given
validator address. The reputation is a decimal between 0 and 1 calculated
from the validator's cumulative epoch reports using an exponential moving
average of reliability, task throughput, and failure rate.`,
		Example: `  portalchaind q poi reputation portal1abc...
  portalchaind q poi reputation portal1abc... --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ValidatorReputation(cmd.Context(), &types.QueryValidatorReputationRequest{
				Validator: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to query reputation for %s: %w", args[0], err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report [epoch] [validator-addr]",
		Short: "Query a specific epoch report",
		Long: `Query a single epoch report submitted by a validator for a particular
epoch. Returns the full report including tasks processed, weighted task sum,
average latency, reliability score, and sampling failures.`,
		Example: `  portalchaind q poi report 42 portal1abc...
  portalchaind q poi report 1 portal1abc... --output json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch %q: must be a 64-bit integer", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EpochReport(cmd.Context(), &types.QueryEpochReportRequest{
				Epoch:     epoch,
				Validator: args[1],
			})
			if err != nil {
				return fmt.Errorf("failed to query report for epoch %d, validator %s: %w", epoch, args[1], err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryReports() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports [validator-addr]",
		Short: "List all epoch reports for a validator",
		Long: `List every epoch report submitted by a validator, ordered by epoch.
Supports standard Cosmos SDK pagination via --page, --limit, --count-total,
and --offset flags.`,
		Example: `  portalchaind q poi reports portal1abc...
  portalchaind q poi reports portal1abc... --limit 10 --offset 5
  portalchaind q poi reports portal1abc... --page 2 --limit 20 --count-total`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return fmt.Errorf("invalid pagination flags: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ReportsByValidator(cmd.Context(), &types.QueryReportsByValidatorRequest{
				Validator:  args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return fmt.Errorf("failed to query reports for %s: %w", args[0], err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "reports")

	return cmd
}

// samplingStoreKey mirrors keeper.samplingStoreKey so the CLI can
// construct the exact binary key used in the KVStore.
func samplingStoreKey(epoch int64, validator string) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(epoch))
	return append(append([]byte(types.SamplingPrefix), buf...), []byte(":"+validator)...)
}

func CmdQuerySamplingRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sampling-record [epoch] [validator-addr]",
		Short: "Query the sampling record for a specific epoch and validator",
		Long: `Query the verification sampling record created when an epoch report is
randomly selected for review. Shows the current status (pending, verified,
or failed), the block-height deadline, and the verifier address.`,
		Example: `  portalchaind q poi sampling-record 6004 portal1abc...
  portalchaind q poi sampling-record 6004 portal1abc... --output json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch %q: must be a 64-bit integer", args[0])
			}

			key := samplingStoreKey(epoch, args[1])

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/key", types.StoreKey),
				Data: key,
			})
			if err != nil {
				return fmt.Errorf("failed to query sampling record: %w", err)
			}

			if len(resp.Value) == 0 {
				return fmt.Errorf("no sampling record found for epoch %d validator %s", epoch, args[1])
			}

			var record types.SamplingRecord
			if err := json.Unmarshal(resp.Value, &record); err != nil {
				return fmt.Errorf("failed to unmarshal sampling record: %w", err)
			}

			if clientCtx.OutputFormat == "json" {
				bz, err := json.MarshalIndent(record, "", "  ")
				if err != nil {
					return err
				}
				return clientCtx.PrintBytes(bz)
			}

			verifier := record.VerifiedBy
			if verifier == "" {
				verifier = "(none)"
			}
			return clientCtx.PrintString(fmt.Sprintf(
				"Sampling Record:\n  Epoch:       %d\n  Validator:   %s\n  Status:      %s\n  Deadline:    %d\n  Verified By: %s\n",
				record.Epoch, record.Validator, record.Status, record.Deadline, verifier,
			))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query PoI module parameters (reward interval, percent, min reputation)",
		Long: `Query the current Proof-of-Intelligence tokenomics parameters:
  - Reward Interval: blocks between reward distributions
  - Reward Percent: fraction of community pool (DAAI) distributed per interval
  - Min Reputation: minimum reputation score to receive rewards`,
		Example: `  portalchaind q poi params
  portalchaind q poi params --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/key", types.StoreKey),
				Data: []byte(types.ParamsKey),
			})
			if err != nil {
				return fmt.Errorf("failed to query params: %w", err)
			}

			params := types.DefaultParams()
			if len(resp.Value) > 0 {
				if err := json.Unmarshal(resp.Value, &params); err != nil {
					return fmt.Errorf("failed to unmarshal params: %w", err)
				}
			}

			if clientCtx.OutputFormat == "json" {
				bz, err := json.MarshalIndent(params, "", "  ")
				if err != nil {
					return err
				}
				return clientCtx.PrintBytes(bz)
			}

			// Reward percent as percentage: 0.001 -> 0.100%
			pct := params.RewardPercent.MulInt64(100)
			return clientCtx.PrintString(fmt.Sprintf(
				"PoI Parameters:\n  Reward Interval:  %d blocks\n  Reward Percent:   %s%%\n  Min Reputation:   %s\n",
				params.RewardInterval,
				pct.String(),
				params.MinReputationForReward.String(),
			))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryPendingSamplings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sampling-records-pending",
		Short: "List all pending sampling records awaiting verification",
		Long: `List every sampling record that is still in "pending" status and
awaiting verification from another validator. These are epoch reports that
were randomly selected for review but have not yet been verified or expired.`,
		Example: `  portalchaind q poi sampling-records-pending
  portalchaind q poi sampling-records-pending --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/subspace", types.StoreKey),
				Data: []byte(types.SamplingPrefix),
			})
			if err != nil {
				return fmt.Errorf("failed to query sampling records: %w", err)
			}

			if len(resp.Value) == 0 {
				return clientCtx.PrintString("No pending sampling records found.\n")
			}

			var pairs kv.Pairs
			if err := pairs.Unmarshal(resp.Value); err != nil {
				return fmt.Errorf("failed to decode store response: %w", err)
			}

			var pending []types.SamplingRecord
			for _, pair := range pairs.Pairs {
				var record types.SamplingRecord
				if err := json.Unmarshal(pair.Value, &record); err != nil {
					continue
				}
				if record.Status == types.SamplingStatusPending {
					pending = append(pending, record)
				}
			}

			if len(pending) == 0 {
				return clientCtx.PrintString("No pending sampling records found.\n")
			}

			if clientCtx.OutputFormat == "json" {
				bz, err := json.MarshalIndent(pending, "", "  ")
				if err != nil {
					return err
				}
				return clientCtx.PrintBytes(bz)
			}

			out := fmt.Sprintf("Pending sampling records: %d\n\n", len(pending))
			for i, r := range pending {
				out += fmt.Sprintf(
					"  [%d] Epoch: %d  Validator: %s  Deadline: %d\n",
					i+1, r.Epoch, r.Validator, r.Deadline,
				)
			}
			return clientCtx.PrintString(out)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
