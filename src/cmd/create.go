package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/luckybroman5/http-log-reconstructor/src/CharlesParsing"
	"github.com/luckybroman5/http-log-reconstructor/src/HarProcessing"
	"github.com/luckybroman5/http-log-reconstructor/src/K6Wrapper"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	domainFilter      *[]string
	charlesExecutable string
	hookTemplate      string

	cmdCreate = &cobra.Command{
		Use:   "create [inputFile]",
		Short: "Takes a .har or .chls file, and creats a load test",
		Long: `Takes a HTTP Archive or Charles log, converts it into a .har int he case of it being
a Charles Log, does some basic processing on the .har, and outputs a k6 load test that very closely
mimics the actions performed in the logs. If it doesn't write the test 100% for you, it'll be 99.9999%`,
		Args: cobra.ExactArgs(1),
		Run:  CreateK6LoadTest,
	}
)

func readFromStdIn() []byte {
	// info, err := os.Stdin.Stat()
	// if err != nil {
	// 	panic(err)
	// }

	// if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
	// 	fmt.Println("The command is intended to work with pipes.")
	// 	fmt.Println("Usage: fortune | gocowsay")
	// 	return
	// }

	reader := bufio.NewReader(os.Stdin)
	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	return []byte(string(output))
}

func CreateK6LoadTest(cmd *cobra.Command, args []string) {
	var hookFile string
	inputFile := args[0]
	if hookTemplate == "" {
		hookFile = "defaultHookFile.js"
	} else {
		hookFile = hookTemplate
	}
	fmt.Println("Creating a load test with inputFile:", inputFile, "and using:", hookFile, "as the hook file...")
	fmt.Println(*domainFilter)
	fmt.Println("Charles Executable:", charlesExecutable)

	charlesLogHarBytes := CharlesParsing.ConvertLog(inputFile, charlesExecutable)
	charlesLogHarBytes = HarProcessing.FilterHar(charlesLogHarBytes, (*domainFilter)[0])

	hookFileBytes, _ := ioutil.ReadFile(hookFile)
	K6Wrapper.CreateLoadTestFromHar(hookFileBytes, charlesLogHarBytes)
}

// Execute executes the root command.
func init() {
	//@TODO: Make it possible to just read the charles file from stdin!
	RootCmd.AddCommand(cmdCreate)
	domainFilter = cmdCreate.Flags().StringArrayP("domainFilter", "f", []string{"example.com"}, "filter the load test to certain domains")
	cmdCreate.Flags().StringVarP(&charlesExecutable, "charles-executable", "c", "charles", "path the charles executable")
	cmdCreate.Flags().StringVarP(&hookTemplate, "hookTemplate", "t", "", "the hooktemplate to be placed into the load test")
}
