package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// TODO:
// - Move flush logic to a separate function,
// - Call the flush logic based on 2MB batch limit / time-based / manual request
// - Add a queue and 10 different connections
// - Extract severity <-- this can be performance heavy, consult Zap folks about providing Entry data without parsing content

const CORALOGIX_SINK_SCHEME = "coralogix"
const DEFAULT_CORALOGIX_PROD_URL = "https://api.coralogix.com/api/v1/logs"

type CoralogixRequest struct {
	PrivateKey   string              `json:"privateKey"`
	AppName      string              `json:"applicationName"`
	SubSysName   string              `json:"subsystemName"`
	ComputerName *string             `json:"computerName"`
	LogEntries   []CoralogixLogEntry `json:"logEntries"`
}

type CoralogixLogEntry struct {
	Timestamp  int64   `json:"timestamp"`
	Severity   uint8   `json:"severity"`
	Text       string  `json:"text"`
	Category   *string `json:"category"`
	ClassName  *string `json:"className"`
	MethodName *string `json:"methodName"`
	ThreadId   *string `json:"threadId"`
}

type coralogixSink struct {
	privateKey string
	appName    string
	subSysName string
	client     *http.Client
	msgBuffer  [][]byte
}

func NewCoralogixZapSinkFactory(privateKey string, appName string, subSysName string) func(u *url.URL) (zap.Sink, error) {
	return func(u *url.URL) (zap.Sink, error) {
		return &coralogixSink{
			privateKey: privateKey,
			appName:    appName,
			subSysName: subSysName,
			client:     &http.Client{},
		}, nil
	}
}

// Write implements zap.Sink Write function
func (s coralogixSink) Write(b []byte) (int, error) {

	entry := CoralogixLogEntry{
		Timestamp: time.Now().UnixMilli(),
		Severity:  1,
		Text:      string(b),
	}
	cReq, _ := s.NewCoralogixRequest([]CoralogixLogEntry{entry})

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(cReq)
	req, _ := http.NewRequest("POST", DEFAULT_CORALOGIX_PROD_URL, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return len(b), nil
	}
	fmt.Printf("Received bad status code: %d\n", resp.StatusCode)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0, nil
	}

	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	// var returnErr error
	return len(b), nil
}

func (s coralogixSink) NewCoralogixRequest(entries []CoralogixLogEntry) (*CoralogixRequest, error) {
	return &CoralogixRequest{
		PrivateKey: s.privateKey,
		AppName:    s.appName,
		SubSysName: s.subSysName,
		LogEntries: entries,
	}, nil
}

func (s coralogixSink) Sync() error {

	return nil
}
func (s coralogixSink) Close() error {
	return nil
}
