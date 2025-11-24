package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type HttpCheck struct {
	client *http.Client
	path   string
	port   int
	scheme string
}

func NewHttpCheck(path, scheme string, port int, insecureSkipVerify bool) HttpCheck {
	client := http.DefaultClient
	if insecureSkipVerify {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 3 * time.Second,
		}
	}
	return HttpCheck{
		client: client,
		path:   path,
		port:   port,
		scheme: scheme,
	}
}

func (hc HttpCheck) Check() Result {
	scheme := hc.scheme
	if scheme == "" {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://127.0.0.1:%d/%s", scheme, hc.port, hc.path)
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
