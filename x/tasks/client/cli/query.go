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
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdListTasks(),
	)

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
