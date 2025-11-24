package main

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type HttpCheck struct {
	client *http.Client
	path   string
	port   int
}

func NewHttpCheck(path string, port int) HttpCheck {
	return HttpCheck{
		client: http.DefaultClient,
		path:   path,
		port:   port,
	}
}

func (hc HttpCheck) Check() Result {
	url := fmt.Sprintf("http://127.0.0.1:%d/%s", hc.port, hc.path)
	resp, err := hc.client.Get(url)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("error while trying to query HTTP endpoint")
		return Result{
			healthy: false,
			err:     err.Error(),
			output:  "",
		}
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	bodyBytes, err := io.ReadAll(resp.Body)
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
	return Result{
		healthy: healthy,
		err:     "",
		output:  body,
	}
}
