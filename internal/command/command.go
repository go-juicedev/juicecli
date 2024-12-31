package command

import "github.com/spf13/cobra"

func NewCommand(name string, args ...Arg) *cobra.Command {
	var cmd = &cobra.Command{Use: name}
	for _, arg := range args {
		cmd.Flags().StringP(arg.Name, arg.ShortHand, arg.Value, arg.Usage)
		if arg.Required {
			_ = cmd.MarkFlagRequired(arg.Name)
		}
	}
	return cmd
}
