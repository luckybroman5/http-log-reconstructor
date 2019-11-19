package K6Wrapper

import (
	"bytes"
	"log"

	"github.com/luckybroman5/http-log-reconstructor/k6/converter/har"
	harlibOpts "github.com/luckybroman5/http-log-reconstructor/k6/lib"
)

func CreateLoadTestFromHar(hookFile []byte, harString []byte) {
	var opts harlibOpts.Options
	only := make([]string, 1)
	only = append(only, "api3.fox.com")
	skip := make([]string, 0)
	// harBytes := []byte(harString)
	bytesReader := bytes.NewReader(harString)
	harObj, _ := har.Decode(bytesReader)
	log.Print(har.Convert(string(hookFile), harObj, opts, uint(0), uint(10), true, false, uint(10), false, false, only, skip))
}
