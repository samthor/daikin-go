// Package api discovers and talks to Daikin devices which have a wireless interface.
package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	ErrStatusNotOk           = errors.New("status was not OK")
	ErrStatusInvalidResponse = errors.New("no status in response")
)

// Post makes a change to the Daikin.
func Post(address, cmd string, values url.Values) (url.Values, error) {
	t := targetFor(address, cmd)
	resp, err := http.PostForm(t, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return processBody(resp.Body)
}

// Get makes a request to the Daikin.
func Get(address, cmd string) (url.Values, error) {
	t := targetFor(address, cmd)
	resp, err := http.Get(t)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return processBody(resp.Body)
}

func processBody(r io.Reader) (url.Values, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	o := ParseValues(string(body))
	ret := o.Get("ret")
	if ret == "" {
		return nil, ErrStatusInvalidResponse
	} else if ret != "OK" {
		return nil, ErrStatusNotOk
	}

	return o, nil
}

func targetFor(address, cmd string) string {
	return fmt.Sprintf("http://%s/%s", address, cmd)
}
