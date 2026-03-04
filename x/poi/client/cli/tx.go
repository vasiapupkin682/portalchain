package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSubmitReport())

	return cmd
}

func CmdSubmitReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-report [epoch] [tasks-processed] [weighted-task-sum] [avg-latency] [reliability] [sampling-failures] [timestamp]",
		Short: "Submit an epoch performance report",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch: %w", err)
			}

			tasksProcessed, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid tasks-processed: %w", err)
			}

			weightedTaskSum, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid weighted-task-sum: %w", err)
			}

			avgLatency, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid avg-latency: %w", err)
			}

			reliability, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid reliability: %w", err)
			}

			samplingFailures, err := strconv.ParseInt(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid sampling-failures: %w", err)
			}

			timestamp, err := strconv.ParseInt(args[6], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timestamp: %w", err)
			}

			validator := clientCtx.GetFromAddress().String()

			msg := types.NewMsgSubmitEpochReport(
				epoch,
				validator,
				tasksProcessed,
				weightedTaskSum,
				avgLatency,
				reliability,
				samplingFailures,
				timestamp,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
