package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"portalchain/x/constitution/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		Long:                       "Query commands for the constitution module — inspect governance parameters and proposal classifications.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryProposalClass(),
	)

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current constitution parameters",
		Long: `Query the active constitution governance parameters including voting
power limits, quorum thresholds, timelock duration, and sacred principles hash.`,
		Example: `  portalchaind q constitution params
  portalchaind q constitution params --output json`,
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
				return fmt.Errorf("failed to query constitution params: %w", err)
			}

			if len(resp.Value) == 0 {
				return printParamsText(clientCtx, types.DefaultParams())
			}

			var params types.ConstitutionParams
			if err := json.Unmarshal(resp.Value, &params); err != nil {
				return fmt.Errorf("failed to unmarshal constitution params: %w", err)
			}

			return printParamsText(clientCtx, params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func printParamsText(clientCtx client.Context, params types.ConstitutionParams) error {
	if clientCtx.OutputFormat == "json" {
		bz, err := json.MarshalIndent(params, "", "  ")
		if err != nil {
			return err
		}
		return clientCtx.PrintBytes(bz)
	}

	out := fmt.Sprintf(`Constitution Parameters:
  Max Voting Power Percent:      %s
  Constitutional Quorum:         %s
  Network Param Quorum:          %s
  Constitutional Timelock Days:  %d
  Sacred Principles Hash:        %s`,
		params.MaxVotingPowerPercent.String(),
		params.ConstitutionalQuorum.String(),
		params.NetworkParamQuorum.String(),
		params.ConstitutionalTimelockDays,
		params.SacredPrinciplesHash,
	)

	return clientCtx.PrintString(out + "\n")
}

func CmdQueryProposalClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal-class [proposal-id]",
		Short: "Query the constitutional classification for a governance proposal",
		Long: `Query the stored classification for a governance proposal. Returns one of:
  SACRED_VIOLATION  — proposal violates an immutable sacred principle (rejected)
  CONSTITUTIONAL    — requires 66% supermajority + timelock
  NETWORK_PARAM     — standard governance flow (50% quorum)

The classification is assigned when the proposal is submitted and stored
permanently so that the EndBlocker can enforce the correct thresholds.`,
		Example: `  portalchaind q constitution proposal-class 1
  portalchaind q constitution proposal-class 42 --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id %q: must be a positive integer", args[0])
			}

			key := fmt.Sprintf("%s%d", types.ProposalClassKey, proposalID)

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/key", types.StoreKey),
				Data: []byte(key),
			})
			if err != nil {
				return fmt.Errorf("failed to query proposal class: %w", err)
			}

			if len(resp.Value) == 0 {
				return fmt.Errorf("no classification found for proposal %d", proposalID)
			}

			class := types.ProposalClass(resp.Value[0])

			if clientCtx.OutputFormat == "json" {
				bz, err := json.MarshalIndent(map[string]interface{}{
					"proposal_id": proposalID,
					"class":       class.String(),
					"class_value": int(class),
				}, "", "  ")
				if err != nil {
					return err
				}
				return clientCtx.PrintBytes(bz)
			}

			return clientCtx.PrintString(fmt.Sprintf(
				"Proposal %d: %s\n", proposalID, class.String(),
			))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
