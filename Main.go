package main

import (
	"fmt"
	"log"
	"strings"
	"bytes"
	"time"
//	"encoding/xml"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"github.com/gorilla/mux"
	"launchpad.net/xmlpath"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/hello", index).Methods("GET")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	//References:
	//	https://wiki.shibboleth.net/confluence/display/SHIB2/NativeSPAssertionExport
	//	https://wiki.shibboleth.net/confluence/display/SHIB2/NativeSPSessions

	log.Println("Responding to /hello request")

	//Get request query parameters
	_return := r.URL.Query().Get("return")
	_target := r.URL.Query().Get("target")

	if(len(_return) == 0) {
		http.Error(w, "return parameter is required", http.StatusBadRequest)
	} else if(len(_target) == 0) {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
	} else {
		//log.Println(fmt.Sprintf("return=%v, target=%v", _return, _target))

		//Get Shib-Assertion-Count parameter
		_shibAssertionCount := r.Header.Get("Shib-Assertion-Count")
		if(len(_shibAssertionCount) == 0) {
			http.Error(w, "invalid Shib-Assertion-Count header", http.StatusBadRequest)
		}

		//Get assertion query url via the Shib-Assertion-NN parameter
		_shibAssertion := r.Header.Get("Shib-Assertion-"+_shibAssertionCount)
		if(len(_shibAssertion) == 0) {
			http.Error(w, "invalid Shib-Assertion-"+_shibAssertionCount+" header", http.StatusBadRequest)
		}
		_shibAssertionUrl := strings.Replace(_shibAssertion, "localhost", "127.0.0.1", 1)

		//Get XML response from SP attribute endpoint
		tr := &http.Transport{
        		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr,
			Timeout: time.Second * 10,
		}
		response, err := client.Get(_shibAssertionUrl)

		//Process XML response
		if (err != nil) {
			fmt.Println(err)
			http.Error(w, "invalid Shib-Assertion", http.StatusBadRequest)
		} else {
			buf, _ := ioutil.ReadAll(response.Body)
			//log.Println(fmt.Sprintf("%s", buf))

			reader := bytes.NewReader(buf)

			//Parse issuer
			//path := xmlpath.MustCompile("/Assertion/Issuer")
			path := xmlpath.MustCompile("//Issuer")
			root, err := xmlpath.Parse(reader)
			if err != nil {
				log.Fatal(err)
			}
			if value, ok := path.String(root); ok {
				log.Println(fmt.Sprintf("Issuer: %s", value))
			}

			//Parse all attributes
			//path = xmlpath.MustCompile("/Assertion/AttributeStatement/Attribute/@Name")
			path2 := xmlpath.MustCompile("//Attribute/@Name")
			root2, err2 := xmlpath.Parse(reader)
			if err2 != nil {
				log.Fatal(err2)
			}
			iter := path2.Iter(root2)
			for iter.Next() {
				if value, ok := path.String(iter.Node()); ok {
					log.Println(fmt.Sprintf("Attribute: %s", value))
				}
			}

			http.Redirect(w, r, _return, http.StatusSeeOther)
		}
	}
}
