package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"portalchain/x/tasks/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdGetTask(),
		CmdListTasks(),
		CmdMyTasks(),
	)

	return cmd
}

func CmdGetTask() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-task [task-id]",
		Short: "Get task by ID (placeholder)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return clientCtx.PrintString("get-task query is not implemented yet; use portalchaind query tx <txhash> to see task details\n")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdListTasks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tasks",
		Short: "List tasks (placeholder)",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return clientCtx.PrintString("list-tasks query is not implemented yet\n")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMyTasks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "my-tasks [creator]",
		Short: "List tasks by creator (placeholder)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return clientCtx.PrintString("my-tasks query is not implemented yet; use portalchaind query tx <txhash> to see task details\n")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
