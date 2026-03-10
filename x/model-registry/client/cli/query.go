package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"portalchain/x/model-registry/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		Long:                       "Query commands for the model registry — discover registered AI models and their endpoints.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryModel(),
		CmdQueryListActive(),
	)

	return cmd
}

func modelKey(operator string) []byte {
	return []byte(types.ModelRegistryPrefix + operator)
}

func CmdQueryModel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model [operator_address]",
		Short: "Query a registered model by operator address",
		Long: `Query the model record for a given operator address. Returns the model
name, endpoint, capabilities, price, and active status.`,
		Example: `  portalchaind q model-registry model portal1abc...
  portalchaind q model-registry model portal1abc... --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			operator := args[0]

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/key", types.StoreKey),
				Data: modelKey(operator),
			})
			if err != nil {
				return fmt.Errorf("failed to query model: %w", err)
			}

			if len(resp.Value) == 0 {
				return fmt.Errorf("no model found for operator %s", operator)
			}

			var record types.ModelRecord
			if err := json.Unmarshal(resp.Value, &record); err != nil {
				return fmt.Errorf("failed to unmarshal model record: %w", err)
			}

			return printModelRecord(clientCtx, record)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryListActive() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-active",
		Short: "List all active models accepting tasks",
		Long: `List every model that is currently active and accepting tasks.
Useful for the Telegram bot and users to discover available models.`,
		Example: `  portalchaind q model-registry list-active
  portalchaind q model-registry list-active --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			resp, err := clientCtx.QueryABCI(abcitypes.RequestQuery{
				Path: fmt.Sprintf("store/%s/subspace", types.StoreKey),
				Data: []byte(types.ModelRegistryPrefix),
			})
			if err != nil {
				return fmt.Errorf("failed to query models: %w", err)
			}

			if len(resp.Value) == 0 {
				if clientCtx.OutputFormat == "json" {
					return clientCtx.PrintBytes([]byte("[]"))
				}
				return clientCtx.PrintString("No active models found.\n")
			}

			var pairs kv.Pairs
			if err := pairs.Unmarshal(resp.Value); err != nil {
				return fmt.Errorf("failed to decode store response: %w", err)
			}

			var active []types.ModelRecord
			for _, pair := range pairs.Pairs {
				var record types.ModelRecord
				if err := json.Unmarshal(pair.Value, &record); err != nil {
					continue
				}
				if record.Active {
					active = append(active, record)
				}
			}

			if len(active) == 0 {
				if clientCtx.OutputFormat == "json" {
					return clientCtx.PrintBytes([]byte("[]"))
				}
				return clientCtx.PrintString("No active models found.\n")
			}

			if clientCtx.OutputFormat == "json" {
				bz, err := json.MarshalIndent(active, "", "  ")
				if err != nil {
					return err
				}
				return clientCtx.PrintBytes(bz)
			}

			out := fmt.Sprintf("Active models: %d\n\n", len(active))
			for i, r := range active {
				out += fmt.Sprintf("  [%d] %s (%s)\n    Operator:   %s\n    Endpoint:   %s\n    Price:      %s\n    Capabilities: %v\n\n",
					i+1, r.ModelName, r.Operator, r.Operator, r.Endpoint, r.PricePerTask, r.Capabilities)
			}
			return clientCtx.PrintString(out)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func printModelRecord(clientCtx client.Context, record types.ModelRecord) error {
	if clientCtx.OutputFormat == "json" {
		bz, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return err
		}
		return clientCtx.PrintBytes(bz)
	}

	activeStr := "no"
	if record.Active {
		activeStr = "yes"
	}
	out := fmt.Sprintf(`Model Record:
  Operator:     %s
  Model Name:   %s
  Endpoint:     %s
  Capabilities: %v
  Price/Task:   %s
  Active:       %s
  Registered:   block %d
  Updated:      block %d
`,
		record.Operator, record.ModelName, record.Endpoint, record.Capabilities,
		record.PricePerTask, activeStr, record.RegisteredAt, record.UpdatedAt,
	)
	return clientCtx.PrintString(out)
}
