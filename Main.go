package main

import (
	"log"
	"errors"
    	"net/http/cgi"
    	"net/http"
    	"fmt"
	"os"
	"strings"
	"bytes"
	"time"
	"io/ioutil"
	"crypto/tls"
	"launchpad.net/xmlpath"
)

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
	aInfo.ts = t.Format(time.RFC3339)//"2012-11-01T22:08:41+00:00Z") //TODO: set timestamp

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
	_shibAssertionCount := os.Getenv("Shib_Assertion_Count")
	if(len(_shibAssertionCount) == 0) {
		return "", errors.New("Shib-Assertion-Count variable not found")
	}

	_shibAssertion := os.Getenv("Shib_Assertion_"+_shibAssertionCount)
	if(len(_shibAssertion) == 0) {
		return "", errors.New("Shib_Assertion_"+_shibAssertionCount+" variable not found")
	}

	return strings.Replace(_shibAssertion, "localhost", "127.0.0.1", 1), nil
}

func sendInfo(aggregator_url string, aInfo *attributeInfo) error {
	url := fmt.Sprintf("%s?idp=%s&sp=%s&timestamp=%s&warn=%s",
		aggregator_url, aInfo.idp, aInfo.sp, aInfo.ts, aInfo.suspicious)

	for _, attr := range aInfo.attributes {
		if attr != "" {
			url = fmt.Sprintf("%s&attributes[]=%s", url, attr)
		}
    	}

	log.Printf("[info] %s", url)
	return nil
}

func main() {
	log_file := "/var/log/sp-session-hook/session-hook-golang.log"
	aggregator_url := "'https://clarin-aa.ms.mff.cuni.cz/aaggreg/v1/got"

	//Initialize logging
	f, err := os.OpenFile(log_file, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		sendErrorResponse(500, fmt.Sprintf("error opening file: %v", err))
		return
	}
	defer f.Close()
	log.SetOutput(f)

	//Get url query parameters
    	req, err := cgi.Request()
    	if err != nil {
        	sendErrorResponse(500, "cannot get cgi request: " + err.Error())
       		return
    	}
	_return := req.URL.Query().Get("return")

	//Get assertion url
	_shibAssertionUrl, err := getShibbolethAssertionUrl()
	if err != nil {
		sendErrorResponse(500, "Failed to parse shibboleth variables: " + err.Error())
		return
	}

	//Get attribute assertions
	attrInfo, err := getAttributeAssertions(_shibAssertionUrl)
	if err != nil {
		sendErrorResponse(500, "Failed to parse shibboleth attribute assertions: " + err.Error())
		return
	}

	//Send info to aagregator
	errSendInfo := sendInfo(aggregator_url, attrInfo)
	if errSendInfo != nil {
		sendErrorResponse(500, "Failed to send attribute information to aagregator: " + errSendInfo.Error())
		return
	}

	//Redirect client
	sendRedirectResponse(_return)
}