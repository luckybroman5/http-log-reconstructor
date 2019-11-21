package K6Wrapper

import (
	"bytes"
	"log"

	"github.com/luckybroman5/http-log-reconstructor/k6/converter/har"
	harlibOpts "github.com/luckybroman5/http-log-reconstructor/k6/lib"
)

func CreateLoadTestFromHar(hookFile []byte, harString []byte, havocMode bool) string {
	var opts harlibOpts.Options
	only := make([]string, 0)
	var minSleep uint
	var maxSleep uint
	if havocMode == true {
		minSleep = 0
		maxSleep = 0
	} else {
		minSleep = 0
		maxSleep = 10
	}

	skip := make([]string, 0)
	bytesReader := bytes.NewReader(harString)
	harObj, err := har.Decode(bytesReader)

	if err != nil {
		log.Println("Error Decoding Har!", err)
	}

	loadTest, err := har.Convert(string(hookFile), harObj, opts, minSleep, maxSleep, true, false, uint(10), false, false, only, skip, havocMode)

	if err != nil {
		log.Println("Error Creating Load test from Har!", err)
	}

	return loadTest
}
