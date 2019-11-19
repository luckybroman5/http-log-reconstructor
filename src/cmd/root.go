package cmd

import (
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "kade6",
		Short: "Takes charles/.har files and converts them into k6 load tests",
		Long: `"kade6" is a cli tool which allows minimal effort in creating a k6 load test.
The typical flow here would be to simply perform the steps in the actual application,
capture charles or and HTTP Archive (.har file), and import it into this tool. 
Once done, you'll have a runnable k6 load test that provides ultimate flexibility!`,
	}
)

// Execute executes the root command.
func Execute() error {
	return RootCmd.Execute()
}
