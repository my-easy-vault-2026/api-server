package bootstrap

import (
	"api-server/console"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "easy_vault_api",
	Short: "easy vault api command runner",
	Long: `
		easy vault api command runner, you can run some console command with this, for example:
		easy_vault_api serve
		easy_vault_api random`,
	TraverseChildren: true,
}

// App root of the application
type App struct {
	*cobra.Command
}

// NewApp creates new root command
func NewApp() App {
	cmd := App{
		Command: rootCmd,
	}
	cmd.AddCommand(console.GetSubCommands(CommonModules)...)
	return cmd
}

var RootApp = NewApp()
