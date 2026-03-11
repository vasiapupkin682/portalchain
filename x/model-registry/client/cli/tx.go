package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"portalchain/x/model-registry/types"
)

const (
	FlagModelName    = "model-name"
	FlagEndpoint     = "endpoint"
	FlagCapabilities = "capabilities"
	FlagPricePerTask = "price-per-task"
	FlagActive       = "active"
	FlagStake        = "stake" // optional, MVP: always use min stake from params
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		Long:                       "Transaction commands for the model registry module — register, update, or deregister AI models.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegister(),
		CmdUpdate(),
		CmdDeregister(),
	)

	return cmd
}

func CmdRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register an AI model for task routing",
		Long: `Register an AI model so users and the Telegram bot can discover and
route tasks to it. The --from account is used as the operator address.`,
		Example: `  portalchaind tx model-registry register \
    --model-name "llama3.2" \
    --endpoint "http://1.2.3.4:8000" \
    --capabilities "text,code,analysis" \
    --price-per-task "10uportal" \
    --from alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			modelName, err := cmd.Flags().GetString(FlagModelName)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagModelName, err)
			}
			if modelName == "" {
				return fmt.Errorf("--%s is required", FlagModelName)
			}

			endpoint, err := cmd.Flags().GetString(FlagEndpoint)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagEndpoint, err)
			}
			if endpoint == "" {
				return fmt.Errorf("--%s is required", FlagEndpoint)
			}

			capStr, err := cmd.Flags().GetString(FlagCapabilities)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagCapabilities, err)
			}
			if capStr == "" {
				return fmt.Errorf("--%s is required", FlagCapabilities)
			}
			capabilities := strings.Split(capStr, ",")
			for i := range capabilities {
				capabilities[i] = strings.TrimSpace(capabilities[i])
				if capabilities[i] == "" {
					return fmt.Errorf("--%s cannot contain empty entries", FlagCapabilities)
				}
			}

			pricePerTask, err := cmd.Flags().GetString(FlagPricePerTask)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagPricePerTask, err)
			}
			if pricePerTask == "" {
				return fmt.Errorf("--%s is required", FlagPricePerTask)
			}

			operator := clientCtx.GetFromAddress().String()

			msg := &types.MsgRegisterModel{
				Operator:     operator,
				ModelName:    modelName,
				Endpoint:     endpoint,
				Capabilities: capabilities,
				PricePerTask: pricePerTask,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagModelName, "", "Model name (e.g. llama3.2, mistral, gemma)")
	cmd.Flags().String(FlagEndpoint, "", "HTTP(S) endpoint URL (e.g. http://1.2.3.4:8000)")
	cmd.Flags().String(FlagCapabilities, "", "Comma-separated capabilities (e.g. text,code,analysis)")
	cmd.Flags().String(FlagPricePerTask, "", "Minimum price per task (e.g. 10uportal)")
	cmd.Flags().String(FlagStake, "", "Amount to stake (default: min stake from params)")
	_ = cmd.MarkFlagRequired(FlagModelName)
	_ = cmd.MarkFlagRequired(FlagEndpoint)
	_ = cmd.MarkFlagRequired(FlagCapabilities)
	_ = cmd.MarkFlagRequired(FlagPricePerTask)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a registered model",
		Long: `Update the endpoint, capabilities, price, or active status of your
registered model. Only non-empty fields are updated. Use --active to set
whether the model accepts tasks.`,
		Example: `  portalchaind tx model-registry update \
    --endpoint "http://5.6.7.8:8000" \
    --active true \
    --from alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			endpoint, _ := cmd.Flags().GetString(FlagEndpoint)
			capStr, _ := cmd.Flags().GetString(FlagCapabilities)
			pricePerTask, _ := cmd.Flags().GetString(FlagPricePerTask)
			active, err := cmd.Flags().GetBool(FlagActive)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagActive, err)
			}

			var capabilities []string
			if capStr != "" {
				capabilities = strings.Split(capStr, ",")
				for i := range capabilities {
					capabilities[i] = strings.TrimSpace(capabilities[i])
					if capabilities[i] == "" {
						return fmt.Errorf("--%s cannot contain empty entries", FlagCapabilities)
					}
				}
			}

			operator := clientCtx.GetFromAddress().String()

			msg := &types.MsgUpdateModel{
				Operator:     operator,
				Endpoint:     endpoint,
				Capabilities: capabilities,
				PricePerTask: pricePerTask,
				Active:       active,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagEndpoint, "", "New HTTP(S) endpoint URL (optional)")
	cmd.Flags().String(FlagCapabilities, "", "Comma-separated capabilities (optional)")
	cmd.Flags().String(FlagPricePerTask, "", "New price per task (optional)")
	cmd.Flags().Bool(FlagActive, true, "Whether the model accepts tasks")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeregister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deregister",
		Short: "Deregister your model",
		Long: `Remove your model from the registry. The --from account must be the
operator who registered the model.`,
		Example: `  portalchaind tx model-registry deregister --from alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			operator := clientCtx.GetFromAddress().String()

			msg := &types.MsgDeregisterModel{
				Operator:  operator,
				ModelName: "",
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
