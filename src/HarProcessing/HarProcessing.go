package HarProcessing

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
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

func FilterHar(harBytes []byte, domainWhiteList string) []byte {
	var parsedHarFile harFile
	json.Unmarshal(harBytes, &parsedHarFile)

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
