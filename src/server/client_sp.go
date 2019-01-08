package server

import (
	"time"
	"crypto/tls"
	"io/ioutil"
	"bytes"
	"launchpad.net/xmlpath"
	"errors"
	"fmt"
	"net/http"
	"net"
	"context"
	"strings"
)

type attributeInfo struct {
	idp string
	sp string
	ts string
	attributes []string
	suspicious string
}

func newDefaultAttributeInfo() (*attributeInfo) {
	aInfo := attributeInfo{}
	aInfo.sp = "unkown"
	aInfo.suspicious = ""
	t := time.Now()
	aInfo.ts = t.Format(time.RFC3339)
	return &aInfo
}


func (h *Handler) getAttributeAssertions(shib_assertion_url string, entity_id string) (*attributeInfo, error) {
	h.logPtr.Debug("Assertion url: %s", shib_assertion_url)

	aInfo := newDefaultAttributeInfo()

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	hosts_map := map[string]string{}
	hosts_map["local.sp1.clarin.eu"] = "proxy"
	hosts_map["local.sp2.clarin.eu"] = "proxy"
	hosts_map["local.compreg.clarin.eu"] = "proxy"
	hosts_map["local.vcr.clarin.eu"] = "proxy"


	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			org_addr := addr

			//TODO: use regex here
			hostname := addr
			port := "443"
			if strings.Contains(addr, ":") {
				p := strings.Split(addr, ":")
				hostname = p[0]
				port = p[1]
			}

			if _, exists := hosts_map[hostname]; exists {
				addr = fmt.Sprintf("%s:%s", hosts_map[hostname], port)
				h.logPtr.Info("Address substitution: %s -> %s", org_addr, addr)
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout: time.Second * 10,
	}

	//Get XML response from SP attribute endpoint
	response, err := client.Get(shib_assertion_url)
	if err != nil {
		return nil, err
	}

	//Parse and process XML
	buf, _ := ioutil.ReadAll(response.Body)
	reader := bytes.NewReader(buf)
	root, err := xmlpath.Parse(reader)
	if err != nil {
		h.logPtr.Info("Response:\n%s\n", string(buf))
		return nil, errors.New("Failed to parse xml: "+ err.Error())
	}

	//Set SP entity id
	aInfo.sp = entity_id

	//Set IDP entity id (Issuer)
	path := xmlpath.MustCompile("//Issuer")
	if value, ok := path.String(root); ok {
		aInfo.idp = value
	}

	//Set attributes
	path = xmlpath.MustCompile("//Attribute/@Name")
	aInfo.attributes = make([]string, 1)
	iter := path.Iter(root)
	for iter.Next() {
		node := iter.Node()
		aInfo.attributes = append(aInfo.attributes, fmt.Sprintf("%v", node))
	}

	return aInfo, nil
}