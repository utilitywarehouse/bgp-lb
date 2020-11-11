package main

import (
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Status struct {
	name    string
	desc    string
	healthy bool
	err     string
	output  string
}

type HttpCheck struct {
	name   string
	desc   string
	client *http.Client
	url    string
}

type HttpCheckInterface interface {
	Check() *Status
}

func NewHttpCheck(name, desc, url string) *HttpCheck {
	return &HttpCheck{
		name:   name,
		desc:   desc,
		client: http.DefaultClient,
		url:    url,
	}
}

func (hc *HttpCheck) Check() *Status {
	resp, err := hc.client.Get(hc.url)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("error while trying to query HTTP endpoint")
		return &Status{
			name:    hc.name,
			desc:    hc.desc,
			healthy: false,
			err:     err.Error(),
			output:  "",
		}
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body := string(bodyBytes)
	healthy := true
	// Non-2XX
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.WithFields(log.Fields{"code": resp.StatusCode}).Warn("invalid response from endpoint")
		healthy = false
	}
	return &Status{
		name:    hc.name,
		desc:    hc.desc,
		healthy: healthy,
		err:     "",
		output:  body,
	}
}
