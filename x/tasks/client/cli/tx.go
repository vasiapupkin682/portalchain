package cli

import (
	"crypto/sha256"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"portalchain/x/tasks/types"
)

const (
	FlagQuery    = "query"
	FlagTaskType = "task-type"
	FlagTaskID   = "task-id"
	FlagResult   = "result"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateTask(),
		CmdSubmitResult(),
	)

	return cmd
}

func CmdCreateTask() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-task",
		Short: "Create a new task",
		Example: `portalchaind tx tasks create-task \
  --query "your question here" \
  --task-type text \
  --from mykey \
  --chain-id portalchain \
  --keyring-backend test`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			query, err := cmd.Flags().GetString(FlagQuery)
			if err != nil {
				return err
			}
			if query == "" {
				return fmt.Errorf("--%s is required", FlagQuery)
			}

			taskType, err := cmd.Flags().GetString(FlagTaskType)
			if err != nil {
				return err
			}
			if taskType == "" {
				taskType = "text"
			}

			hash := sha256.Sum256([]byte(query))
			queryHash := fmt.Sprintf("%x", hash)
			queryURL := "local://" + queryHash

			msg := &types.MsgCreateTask{
				Creator:   clientCtx.GetFromAddress().String(),
				QueryHash: queryHash,
				QueryUrl:  queryURL,
				TaskType:  taskType,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagQuery, "", "Task query text")
	cmd.Flags().String(FlagTaskType, "text", "Task type: text/code/analysis")
	_ = cmd.MarkFlagRequired(FlagQuery)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitResult() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-result",
		Short: "Submit task result",
		Example: `portalchaind tx tasks submit-result \
  --task-id task-1 \
  --result "answer here" \
  --from mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			taskID, err := cmd.Flags().GetString(FlagTaskID)
			if err != nil {
				return err
			}
			if taskID == "" {
				return fmt.Errorf("--%s is required", FlagTaskID)
			}

			result, err := cmd.Flags().GetString(FlagResult)
			if err != nil {
				return err
			}
			if result == "" {
				return fmt.Errorf("--%s is required", FlagResult)
			}

			hash := sha256.Sum256([]byte(result))
			resultHash := fmt.Sprintf("%x", hash)
			resultURL := "local://" + resultHash

			msg := &types.MsgSubmitResult{
				Agent:      clientCtx.GetFromAddress().String(),
				TaskId:     taskID,
				ResultHash: resultHash,
				ResultUrl:  resultURL,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagTaskID, "", "Task ID")
	cmd.Flags().String(FlagResult, "", "Result text")
	_ = cmd.MarkFlagRequired(FlagTaskID)
	_ = cmd.MarkFlagRequired(FlagResult)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
