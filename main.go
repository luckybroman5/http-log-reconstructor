package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/luckybroman5/http-log-reconstructor/k6/converter/har"
	harlibOpts "github.com/luckybroman5/http-log-reconstructor/k6/lib"
)

type harFilePostRequestDataParams struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type harFileEntryRequestPostData struct {
	MimeType string                         `json:"mimeType"`
	Params   []harFilePostRequestDataParams `json:"params"`
	Text     string                         `json:"text"`
}

type harFileHeader struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type harFileResponseBodyContent struct {
	Size        int    `json:"size"`
	Compression int    `json:"compression"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text"`
	Comment     string `json:"comment"`
}

type harFileCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type harFileQueryString struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type harFileEntryRequest struct {
	Method      string                      `json:"method"`
	HTTPVersion string                      `json:"httpVersion"`
	URL         string                      `json:"url"`
	Cookies     []harFileCookie             `json:"cookies"`
	Headers     []harFileHeader             `json:"headers"`
	QueryString []harFileQueryString        `json:"queryString"`
	PostData    harFileEntryRequestPostData `json:"postData"`
	HeaderSize  int                         `json:"headerSize"`
	BodySize    int                         `json:"bodySize"`
}

type harFileEntryResponse struct {
	Status      int                        `json:"status"`
	StatusText  string                     `json:"statusText"`
	HTTPVersion string                     `json:"httpVersion"`
	Cookies     []harFileCookie            `json:"cookies"`
	Headers     []harFileHeader            `json:"headers"`
	Content     harFileResponseBodyContent `json:"content"`
	RedirectURL string                     `json:"redirectUrl"`
	HeaderSize  int                        `json:"headerSize"`
	BodySize    int                        `json:"bodySize"`
}

type harFileEntry struct {
	StartedDateTime string               `json:"startedDateTime"`
	Time            int                  `json:"time"`
	Request         harFileEntryRequest  `json:"request"`
	Response        harFileEntryResponse `json:"response"`
}

type harFileLogCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harFileLog struct {
	Version string `json:"version"`
	Creator harFileLogCreator
	Entries []harFileEntry `json:"entries"`
	Pages   []string       `json:"pages"`
}

type harFile struct {
	Log harFileLog `json:"log"`
}

type domainConfig struct {
}

// @TODO change this to
var CharlesExecutable = "charles"

func convertLog(fileName string) {
	outputNameArr := strings.Split(fileName, "/")
	outputName := outputNameArr[len(outputNameArr)-1]
	// https://golang.org/pkg/io/ioutil/#TempDir use that later
	cmd := exec.Command(CharlesExecutable, "convert", fileName, "output/"+outputName+".har")
	log.Printf("Running command and waiting for it to finish...")
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
}

func filterHar(fileName string, domainWhiteList string) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	var parsedHarFile harFile
	json.Unmarshal(content, &parsedHarFile)

	var newHarFile harFile

	newHarFile.Log.Pages = make([]string, 0)
	newHarFile.Log.Version = "1.2"
	newHarFile.Log.Creator.Name = "RoboTodd"
	newHarFile.Log.Creator.Version = "0.1"

	for _, v := range parsedHarFile.Log.Entries {
		if strings.Index(v.Request.URL, domainWhiteList) != -1 {
			newHarFile.Log.Entries = append(newHarFile.Log.Entries, v)
		}
	}

	marshalled, err := json.Marshal(newHarFile)
	if err != nil {
		log.Print(err)
	}

	ioutil.WriteFile("output/test.har", marshalled, 0644)
	return marshalled
}

func createLoadTestFromHar(hookFile []byte, harString []byte) {
	var opts harlibOpts.Options
	only := make([]string, 1)
	only = append(only, "api3.fox.com")
	skip := make([]string, 0)
	// harBytes := []byte(harString)
	bytesReader := bytes.NewReader(harString)
	harObj, _ := har.Decode(bytesReader)
	log.Print(har.Convert(string(hookFile), harObj, opts, uint(0), uint(10), true, false, uint(10), false, false, only, skip))
}

func main() {
	// convertLog("test-data/Android.chls")
	harString := filterHar("output/Android.chls.har", "api3.fox.com")
	hookFile, err := ioutil.ReadFile("defaultHookFile.js")
	if err != nil {
		log.Fatal(err)
	}
	createLoadTestFromHar(hookFile, harString)
}
