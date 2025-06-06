package impl

import (
	"fmt"
	"github.com/go-juicedev/juicecli/cmds/impl/internal"
	"github.com/go-juicedev/juicecli/internal/command"
	"github.com/spf13/cobra"
	"io"
)

func do(targetType, namespace, output, cfg string) error {
	parser := internal.NewParser(targetType).WithNamespace(namespace).WithOutput(output).WithConfig(cfg)
	config, err := parser.Config()
	if err != nil {
		return err
	}
	iface, file, err := parser.TypeInterface()
	if err != nil {
		return err
	}
	namespace, err = parser.Namespace()
	if err != nil {
		return err
	}
	implement := internal.NewImplement(file, iface, targetType, targetType+"Impl")
	generator := internal.NewGenerator(namespace, config, implement)
	reader, err := generator.Generate()
	if err != nil {
		return err
	}
	writer, err := parser.Output()
	if err != nil {
		return err
	}
	defer func() {
		if closer, ok := writer.(io.Closer); ok {
			_ = closer.Close()
		}
	}()
	_, err = io.Copy(writer, reader)
	return err
}

func NewCommand() *cobra.Command {
	typeArg := command.Arg{
		Name:      "type",
		ShortHand: "t",
		Required:  true,
		Usage:     "The interface type name to generate implementation for (e.g. UserRepository)",
	}
	namespaceArg := command.Arg{
		Name:      "namespace",
		ShortHand: "n",
		Usage:     "The package name for the generated implementation (e.g. repository). If not specified, it will be auto-generated",
	}
	outputArg := command.Arg{
		Name:      "output",
		ShortHand: "o",
		Usage:     "The output file path for the generated implementation. If not specified, output will be written to stdout",
	}
	configArg := command.Arg{
		Name:      "config",
		ShortHand: "c",
		Usage:     "The configuration file path. If not specified, it will search for juice.xml, config/juice.xml, config.xml, or config/config.xml",
	}
	args := []command.Arg{
		typeArg,
		namespaceArg,
		outputArg,
		configArg,
	}
	cmd := command.NewCommand("impl", args...)
	cmd.Short = "Generate implementation for an interface"
	cmd.Long = "Generate implementation for an interface based on configuration. It supports customizing the implementation through XML configuration files."
	cmd.Example = "  juicecli impl --type UserRepository\n" +
		"  juicecli impl --type UserRepository --namespace repository --output user_repository.go\n" +
		"  juicecli impl --type UserRepository --config custom.xml"
	cmd.Run = func(cmd *cobra.Command, args []string) {
		targetType, _ := cmd.Flags().GetString(typeArg.Name)
		namespace, _ := cmd.Flags().GetString(namespaceArg.Name)
		output, _ := cmd.Flags().GetString(outputArg.Name)
		config, _ := cmd.Flags().GetString(configArg.Name)
		if err := do(targetType, namespace, output, config); err != nil {
			fmt.Println(err)
		}
	}
	return cmd
}
