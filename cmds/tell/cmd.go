package tell

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-juicedev/juicecli/internal/command"
	"github.com/go-juicedev/juicecli/internal/namespace"
	"github.com/spf13/cobra"
)

func do(targetType string) {
	cmp := &namespace.AutoComplete{TypeName: targetType}
	data, err := cmp.Autocomplete()
	if err != nil {
		fmt.Println(err)
		return
	}
	color.Green(data)
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
	cmd.Run = func(cmd *cobra.Command, args []string) {
		targetType, _ := cmd.Flags().GetString(targetType.Name)
		do(targetType)
	}
	return cmd
}
