package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

var (
	cmdPrintDefaultHookFile = &cobra.Command{
		Use:   "default-template",
		Short: "Takes a .har or .chls file, and creats a load test",
		Long:  `Prints the default hookfile to be used for bootstrapping a new custom hook file`,
		Args:  cobra.ExactArgs(0),
		Run:   printDefaultHookFile,
	}
)

func printDefaultHookFile(*cobra.Command, []string) {
	bytes, _ := ioutil.ReadFile("defaultHookFile.js")

	fmt.Println(string(bytes))
}

func init() {
	//@TODO: Make it possible to just read the charles file from stdin!
	RootCmd.AddCommand(cmdPrintDefaultHookFile)
}
