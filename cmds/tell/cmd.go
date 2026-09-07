package tell

import (
	"github.com/fatih/color"
	"github.com/go-juicedev/juicecli/internal/command"
	"github.com/go-juicedev/juicecli/internal/namespace"
	"github.com/spf13/cobra"
)

func do(targetType string) error {
	cmp := &namespace.AutoComplete{TypeName: targetType}
	data, err := cmp.Autocomplete()
	if err != nil {
		return err
	}
	color.Green(data)
	return nil
}

func NewCommand() *cobra.Command {
	targetType := command.Arg{
		Name:      "type",
		ShortHand: "t",
		Required:  true,
		Usage:     "The interface type name to generate implementation for (e.g. UserRepository)",
	}
	cmd := command.NewCommand("tell", targetType)
	cmd.Short = "Auto-generate namespace for an interface type"
	cmd.Long = "Analyze the interface type and suggest an appropriate namespace based on its name and structure"
	cmd.Example = "  juicecli tell --type UserRepository\n" +
		"  juicecli tell -t UserRepository"
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		targetType, err := cmd.Flags().GetString(targetType.Name)
		if err != nil {
			return err
		}
		return do(targetType)
	}
	return cmd
}
