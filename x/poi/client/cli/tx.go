package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"portalchain/x/poi/types"
)

const (
	FlagEpoch            = "epoch"
	FlagTasksProcessed   = "tasks-processed"
	FlagWeightedTaskSum  = "weighted-task-sum"
	FlagAvgLatency       = "avg-latency"
	FlagReliability      = "reliability"
	FlagSamplingFailures = "sampling-failures"
	FlagTimestamp        = "timestamp"
	FlagTaskType         = "task-type"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		Long:                       "Transaction commands for the Proof of Intelligence (poi) module.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSubmitReport())

	return cmd
}

func CmdSubmitReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-report",
		Short: "Submit an epoch performance report for the signing validator",
		Long: `Submit an epoch performance report that records the validator's work
during a given epoch. The report captures task throughput, weighted task
complexity, average latency, a reliability score (decimal between 0 and 1),
and a sampling failure count.

The --from account is used as the validator address. A timestamp defaults
to the current unix time when omitted.`,
		Example: `  portalchaind tx poi submit-report \
    --epoch 1 \
    --tasks-processed 100 \
    --weighted-task-sum 5000 \
    --avg-latency 150 \
    --reliability 0.95 \
    --sampling-failures 0 \
    --from alice --chain-id portalchain --yes

  portalchaind tx poi submit-report \
    --epoch 42 \
    --tasks-processed 1500 \
    --weighted-task-sum 75000 \
    --avg-latency 120 \
    --reliability 0.99 \
    --sampling-failures 2 \
    --timestamp 1700000000 \
    --from bob`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epoch, err := cmd.Flags().GetInt64(FlagEpoch)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagEpoch, err)
			}
			if epoch < 0 {
				return fmt.Errorf("--%s must be non-negative, got %d", FlagEpoch, epoch)
			}

			tasksProcessed, err := cmd.Flags().GetInt64(FlagTasksProcessed)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagTasksProcessed, err)
			}

			weightedTaskSum, err := cmd.Flags().GetInt64(FlagWeightedTaskSum)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagWeightedTaskSum, err)
			}

			avgLatency, err := cmd.Flags().GetInt64(FlagAvgLatency)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagAvgLatency, err)
			}

			reliabilityStr, err := cmd.Flags().GetString(FlagReliability)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagReliability, err)
			}
			reliability, err := sdk.NewDecFromStr(reliabilityStr)
			if err != nil {
				return fmt.Errorf("--%s must be a valid decimal (e.g. \"0.95\"): %w", FlagReliability, err)
			}

			samplingFailures, err := cmd.Flags().GetInt64(FlagSamplingFailures)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagSamplingFailures, err)
			}

			ts, err := cmd.Flags().GetInt64(FlagTimestamp)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", FlagTimestamp, err)
			}
			if ts == 0 {
				ts = time.Now().Unix()
			}

			taskType, _ := cmd.Flags().GetString(FlagTaskType)
			if taskType == "" {
				taskType = "general"
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
				ts,
				taskType,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int64(FlagEpoch, 0, "Epoch number for this report (required)")
	cmd.Flags().Int64(FlagTasksProcessed, 0, "Number of tasks processed during the epoch")
	cmd.Flags().Int64(FlagWeightedTaskSum, 0, "Weighted sum of task complexities")
	cmd.Flags().Int64(FlagAvgLatency, 0, "Average task latency in milliseconds")
	cmd.Flags().String(FlagReliability, "1.0", "Reliability score between 0 and 1 (decimal)")
	cmd.Flags().Int64(FlagSamplingFailures, 0, "Number of sampling failures observed")
	cmd.Flags().Int64(FlagTimestamp, 0, "Unix timestamp; defaults to current time when 0")
	cmd.Flags().String(FlagTaskType, "general", "Task category: text|code|analysis|general")

	_ = cmd.MarkFlagRequired(FlagEpoch)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
