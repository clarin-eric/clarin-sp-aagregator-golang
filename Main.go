package main

import (
	"log"
	"errors"
    	"fmt"
	"os"
	"strings"
	"strconv"
	"bytes"
	"time"
	"encoding/json"
	"net/url"
	"net/http"
	"net/http/cgi"
	"io/ioutil"
	"crypto/tls"
	"launchpad.net/xmlpath"
)

const key_shib_assertion_count = "Shib_Assertion_Count"
const key_shib_assertion = "Shib_Assertion_"
const key_config_submit_stats = "submit_sp_stats"
const key_config_log_path = "log_path"
const key_config_log_file = "log_file"
const key_config_aag_url = "aag_url"
const key_config_aag_path = "aag_path"

const default_log_path = "/var/log/sp-session-hook/"
const default_log_file = "session-hook-golang.log"
const default_aggregator_url = "https://clarin-aa.ms.mff.cuni.cz"
const default_aggregator_path = "/aaggreg/v1/got"
const default_submitSpStats = false

type apiResponse struct {
    ok bool `json:"ok"`
}

type attributeInfo struct {
	idp string
	sp string
	ts string
	attributes []string
	suspicious string
}

func sendErrorResponse(code int, msg string) {
    	fmt.Printf("Status:%d %s\r\n", code, msg)
    	fmt.Printf("Content-Type: text/plain\r\n")
    	fmt.Printf("\r\n")
    	fmt.Printf("%s\r\n", msg)
}

func sendRedirectResponse(location string) {
	fmt.Printf("Status: 302 Found\r\n")
    	fmt.Printf("Location: %s\r\n", location)
    	fmt.Printf("\r\n")
}

func getAttributeAssertions(url string) (*attributeInfo, error) {
	aInfo := new(attributeInfo)
	aInfo.sp = "unkown"
	aInfo.suspicious = ""
	t := time.Now()
	aInfo.ts = t.Format(time.RFC3339)

	//Get XML response from SP attribute endpoint
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout: time.Second * 10,
	}

	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	//Parse and process XML
	buf, _ := ioutil.ReadAll(response.Body)
	reader := bytes.NewReader(buf)
	root, err := xmlpath.Parse(reader)
	if err != nil {
		return nil, errors.New("Failed to parse Issuer xpath expression")
	}

	//Parse SPNameQualifier
	path := xmlpath.MustCompile("//NameID/@SPNameQualifier")
	if value, ok := path.String(root); ok {
		aInfo.sp = value
	}

	//Parse Issuer
	path = xmlpath.MustCompile("//Issuer")
	if value, ok := path.String(root); ok {
		aInfo.idp = value
	}

	//Parse all attributes
	path = xmlpath.MustCompile("//Attribute/@Name")
	aInfo.attributes = make([]string, 1)
	iter := path.Iter(root)
	for iter.Next() {
		if value, ok := path.String(iter.Node()); ok {
			aInfo.attributes = append(aInfo.attributes, value)
		}
	}

	return aInfo, nil
}

func getShibbolethAssertionUrl() (string, error) {
	_shibAssertionCount := os.Getenv(key_shib_assertion_count)
	if(len(_shibAssertionCount) == 0) {
		return "", errors.New(key_shib_assertion_count+" variable not found")
	}

	_shibAssertion := os.Getenv(key_shib_assertion+_shibAssertionCount)
	if(len(_shibAssertion) == 0) {
		return "", errors.New(key_shib_assertion+_shibAssertionCount+" variable not found")
	}

	return strings.Replace(_shibAssertion, "localhost", "127.0.0.1", 1), nil
}

func sendInfo(submit bool, aggregator_url string, aggregator_path string, aInfo *attributeInfo) error {
	aagUrl, err := url.Parse(aggregator_url)
    	if err != nil {
        	return err
    	}

	aagUrl.Path += aggregator_path
    	parameters := url.Values{}
    	parameters.Add("idp", aInfo.idp)
    	parameters.Add("sp", aInfo.sp)
    	parameters.Add("timestamp", aInfo.ts)
	parameters.Add("warn", aInfo.suspicious)
	for _, attr := range aInfo.attributes {
		if attr != "" {
			parameters.Add("attributes[]", attr)
		}
    	}
	aagUrl.RawQuery = parameters.Encode()

	if(submit) {
		logInfo(fmt.Sprintf("Submitting statistics: %s", aagUrl.String()))
		client := &http.Client{
			Timeout: time.Second * 10,
		}

		response, err := client.Get(aagUrl.String())
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		var s = new(apiResponse)
		err = json.Unmarshal(body, &s)
		if (err != nil) {
			return err
		}

		if (s.ok) {
			logInfo("Success")
		} else {
			logInfo(fmt.Sprintf("Failed: [%v]", body))
		}
	} else {
		logInfo(fmt.Sprintf("Submission of statistics is disabled: %s", aagUrl.String()))
	}

	return nil
}

func logInfo(msg string) {
	log.Printf("[INFO] %s", msg)
}

func logError(msg string) {
	log.Printf("[ERROR] %s", msg)
}

func logErrorWithResponse(msg string) {
	logError(msg)
	sendErrorResponse(500, msg)
}

func getStringConfigValue(key string, defaultValue string) (string) {
	value := os.Getenv(key)
	if len(value) > 0 {
		return value
	} else {
		return defaultValue
	}
}

func getBooleanConfigValue(key string, defaultValue bool) (bool) {
	value := os.Getenv(key)
	if len(value) > 0 {
		x, err := strconv.ParseBool(value)
		if err != nil {
			logError(fmt.Sprintf("error parsing %s value to bool: %v", key, err))
			return false
		}
		return x
	} else {
		return defaultValue
	}
}

func main() {
	//TODO: log errors to file and provide more generic error messages

	//Parse configuration options
	log_path := getStringConfigValue(key_config_log_path, default_log_path)
	log_file := getStringConfigValue(key_config_log_file, default_log_file)
	aggregator_url := getStringConfigValue(key_config_aag_url, default_aggregator_url)
	aggregator_path := getStringConfigValue(key_config_aag_path, default_aggregator_path)
	submitSpStats := getBooleanConfigValue(key_config_submit_stats, default_submitSpStats)

	//Switch to file based logging
	f, err := os.OpenFile(fmt.Sprintf("%s%s", log_path, log_file), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		logError(fmt.Sprintf("failed to open logfile: %v", err))
		return
	}
	defer f.Close()
	log.SetOutput(f)

	//Get url query parameters
    	req, err := cgi.Request()
    	if err != nil {
        	logErrorWithResponse("cannot get cgi request: " + err.Error())
       		return
    	}
	_return := req.URL.Query().Get("return")

	//Get assertion url
	_shibAssertionUrl, err := getShibbolethAssertionUrl()
	if err != nil {
		logErrorWithResponse("Failed to parse shibboleth variables: " + err.Error())
		return
	}

	//Get attribute assertions
	attrInfo, err := getAttributeAssertions(_shibAssertionUrl)
	if err != nil {
		logErrorWithResponse("Failed to parse shibboleth attribute assertions: " + err.Error())
		return
	}

	//Send info to aagregator
	errSendInfo := sendInfo(submitSpStats, aggregator_url, aggregator_path, attrInfo)
	if errSendInfo != nil {
		logErrorWithResponse("Failed to send attribute information to aagregator: " + errSendInfo.Error())
		return
	}

	//Redirect client
	sendRedirectResponse(_return)
}