package server

import (
	"net/url"
	"fmt"
	"time"
	"io/ioutil"
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	ok bool `json:"ok"`
}

func (h *Handler) sendInfo(aInfo *attributeInfo) error {
	aagUrl, err := url.Parse(h.aggregator_url)
	if err != nil {
		return err
	}

	aagUrl.Path += h.aggregator_path
	parameters, err := url.ParseQuery(aagUrl.RawQuery)
	if err != nil {
		return err
	}
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

	if(!h.simulate) {
		h.logPtr.Info("Submitting statistics: %s\n", aagUrl.String())
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
			h.logPtr.Info("Success\n")
		} else {
			h.logPtr.Info(fmt.Sprintf("Failed: [%v]\n", body))
		}
	} else {
		h.logPtr.Info("Simulating submission of statistics: %s\n", aagUrl.String())
	}

	return nil
}