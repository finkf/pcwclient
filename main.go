package main // import "github.com/finkf/pcwclient"
import (
	"github.com/spf13/cobra"
)

// various command line flags
var mainArgs = struct {
	debug, skipVerify     bool
	authToken, pocowebURL string
}{}

var mainCommand = &cobra.Command{
	Use:   "pcwclient",
	Short: "Command line client for pocoweb",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// if mainArgs.debug {
		// 	log.SetLevel(log.DebugLevel)
		// }
	},
	Long: `
Command line client for pocoweb. You can use it to automate or test
the pocoweb post-correction.

In order to use the command line client, you should use the
POCOWEB_URL and POCOWEB_AUTH environment varibales to set the url and
the authentification token respectively or set the appropriate --url
and --auth parameters accordingly.`,
}

func init() {
	mainCommand.AddCommand(&listCommand)
	mainCommand.AddCommand(&newCommand)
	mainCommand.AddCommand(&loginCommand)
	mainCommand.AddCommand(&logoutCommand)
	mainCommand.AddCommand(&printCommand)
	mainCommand.AddCommand(&versionCommand)
	mainCommand.AddCommand(&searchCommand)
	mainCommand.AddCommand(&correctCommand)
	mainCommand.AddCommand(&downloadCommand)
	mainCommand.AddCommand(&pkgCommand)
	downloadCommand.AddCommand(&downloadBookCommand)
	downloadCommand.AddCommand(&downloadPoolCommand)
	pkgCommand.AddCommand(&pkgAssignCommand)
	pkgCommand.AddCommand(&pkgReassignCommand)
	pkgCommand.AddCommand(&pkgSplitCommand)
	mainCommand.AddCommand(&deleteCommand)
	mainCommand.AddCommand(&startCommand)
	listCommand.AddCommand(&listBooksCommand)
	listCommand.AddCommand(&listUsersCommand)
	listCommand.AddCommand(&listPatternsCommand)
	listCommand.AddCommand(&listSuggestionsCommand)
	listCommand.AddCommand(&listSuspiciousCommand)
	listCommand.AddCommand(&listAdaptiveCommand)
	listCommand.AddCommand(&listELCommand)
	listCommand.AddCommand(&listRRDMCommand)
	listCommand.AddCommand(&listCharsCommand)
	newCommand.AddCommand(&newUserCommand)
	newCommand.AddCommand(&newBookCommand)
	startCommand.AddCommand(&startProfileCommand)
	startCommand.AddCommand(&startELCommand)
	startCommand.AddCommand(&startRRDMCommand)
	deleteCommand.AddCommand(&deleteBooksCommand)
	deleteCommand.AddCommand(&deleteUsersCommand)

	mainCommand.SilenceUsage = true
	mainCommand.SilenceErrors = true
	mainCommand.PersistentFlags().BoolVarP(&formatArgs.json, "json", "J", false,
		"output raw json")
	mainCommand.PersistentFlags().BoolVarP(&mainArgs.skipVerify,
		"skip-verify", "S", false, "ignore invalid ssl certificates")
	mainCommand.PersistentFlags().BoolVarP(&mainArgs.debug, "debug", "D", false,
		"enable debug output")
	mainCommand.PersistentFlags().StringVarP(&mainArgs.pocowebURL, "url", "U",
		getURL(), "set pocoweb url")
	mainCommand.PersistentFlags().StringVarP(&formatArgs.template, "format", "F",
		"", "set output format")
	mainCommand.PersistentFlags().StringVarP(&mainArgs.authToken, "auth", "A",
		getAuth(), "set auth token")
}

func main() {
	chk(mainCommand.Execute())
}
