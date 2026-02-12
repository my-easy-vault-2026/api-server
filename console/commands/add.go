package commands

import (
	"github.com/my-easy-vault-2026/api-server/lib"

	"github.com/spf13/cobra"
)

type AddCommand struct {
	num1 int
	num2 int
}

func (r *AddCommand) Short() string {
	return "add two numbers"
}

func (r *AddCommand) Setup(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&r.num1, "num1", "a", 0, "first number to add")
	cmd.Flags().IntVarP(&r.num2, "num2", "b", 0, "second number to add")
}

func (r *AddCommand) Run() lib.CommandRunner {
	return func(l lib.Logger) {
		l.Info("running add command")
		l.Info("sum of numbers is", r.num1+r.num2)
	}
}

func NewAddCommand() *AddCommand {
	return &AddCommand{}
}
