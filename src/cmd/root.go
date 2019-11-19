package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	// domainFilter      *[]string
	// charlesExecutable string
	// userLicense       string

	RootCmd = &cobra.Command{
		Use:   "kade6",
		Short: "Takes charles/.har files and converts them into k6 load tests",
		Long: `"kade6" is a cli tool which allows minimal effort in creating a k6 load test.
The typical flow here would be to simply perform the steps in the actual application,
capture charles or and HTTP Archive (.har file), and import it into this tool. 
Once done, you'll have a runnable k6 load test that provides ultimate flexibility!`,
	}

	// cmdPrintDefaultHookFile = &cobra.Command{
	// 	Use:   "show-default-hookfile",
	// 	Short: "Prints the default hookfile",
	// 	Long:  `Prints the default hookfile to be used for bootstrapping a new custom hook file`,
	// 	Args:  cobra.MaximumNArgs(0),
	// 	Run:   printDefaultHookFile,
	// }

// 	cmdCreate = &cobra.Command{
// 		Use:   "create [inputFile] [outputFile]",
// 		Short: "Takes a .har or .chls file, and creats a load test",
// 		Long: `Takes a HTTP Archive or Charles log, converts it into a .har int he case of it being
// a Charles Log, does some basic processing on the .har, and outputs a k6 load test that very closely
// mimics the actions performed in the logs. If it doesn't write the test 100% for you, it'll be 99.9999%`,
// 		Args: cobra.RangeArgs(1, 2),
// 		Run:  CreateK6LoadTest,
// 	}
)

// func printDefaultHookFile(*cobra.Command, []string) {
// 	fmt.Println("This is the default hookfile")
// }

// func CreateK6LoadTest(cmd *cobra.Command, args []string) {
// 	var hookFile string
// 	inputFile := args[0]
// 	if len(args) == 1 {
// 		hookFile = "defaultHookFile.js"
// 	} else {
// 		hookFile = args[1]
// 	}
// 	fmt.Println("Creating a load test with inputFile:", inputFile, "and using:", hookFile, "as the hook file...")
// 	fmt.Println(*domainFilter)
// 	fmt.Println("Charles Executable:", charlesExecutable)

// }

// Execute executes the root command.
func Execute() error {
	return RootCmd.Execute()
}
