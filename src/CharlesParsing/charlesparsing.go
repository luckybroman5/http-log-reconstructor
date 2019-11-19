package CharlesParsing

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func ConvertLog(fileName string, CharlesExecutable string) []byte {
	// Charles has to create the output file for whatever reason.
	// Using Go's Reliable TMP File Naming
	tmpOutfile, _ := ioutil.TempFile("", "kade6tmp*.har")
	outFileName := tmpOutfile.Name()
	os.Remove(tmpOutfile.Name()) // clean up

	cmd := exec.Command(CharlesExecutable, "convert", fileName, outFileName)

	log.Printf("Converting Charles File to .har ...")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := ioutil.ReadAll(stderr)
	fmt.Printf("%s\n", slurp)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	outBytes, _ := ioutil.ReadFile(outFileName)

	return outBytes
}
