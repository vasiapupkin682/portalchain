package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
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
		CmdListTasks(),
		CmdGetTask(),
		CmdAgentTasks(),
	)
	return cmd
}

func CmdListTasks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tasks",
		Short: "List all tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ListTasks(cmd.Context(), &types.QueryListTasksRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintString(res.TasksJson + "\n")
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdGetTask() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-task [task-id]",
		Short: "Get task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetTask(cmd.Context(), &types.QueryGetTaskRequest{TaskId: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintString(res.TaskJson + "\n")
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdAgentTasks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-tasks [agent-address]",
		Short: "List pending tasks assigned to agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AgentTasks(cmd.Context(), &types.QueryAgentTasksRequest{Agent: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintString(res.TasksJson + "\n")
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
