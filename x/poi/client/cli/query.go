package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
