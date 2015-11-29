package server

import (
	// go standard library
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	// within this project
	"util"

	// third party
	"github.com/ddliu/go-httpclient"
)

func Start(port int, preBuffer int, postBuffer int, stock bool) error {
	options := []string{"-p", strconv.Itoa(port)}
	options = append(options, []string{"-b", strconv.Itoa(preBuffer)}...)
	options = append(options, []string{"-B", strconv.Itoa(postBuffer)}...)
	if stock {
		options = append(options, "-s")
	}

	cmd := exec.Command("alda-server", options...)
	return cmd.Start()
}

func Stop(host string, port int) error {
	statusCode, body, err := Get(host, port, "/stop", nil)

	if err != nil {
		return util.SanitizeError(err)
	}

	if statusCode != 200 {
		return errors.New(fmt.Sprintf("(%d) %s", statusCode, body))
	}

	return nil
}

func Play(host string, port int, codeOrFilename string, argType string, replace bool) error {
	var httpFn func(string, int, string, string) (int, string, error)
	switch {
	case argType == "code" && replace:
		httpFn = PutString
	case argType == "code":
		httpFn = PostString
	case argType == "file" && replace:
		httpFn = PutFile
	case argType == "file":
		httpFn = PostFile
	}

	statusCode, body, err := httpFn(host, port, "/play", codeOrFilename)

	if err != nil {
		return util.SanitizeError(err)
	}

	if statusCode != 200 {
		return errors.New(fmt.Sprintf("(%d) %s", statusCode, body))
	}

	return nil
}

func CheckForConnection(host string, port int) error {
	statusCode, body, err := Get(host, port, "/", nil)

	if err != nil {
		return err
	}

	if statusCode != 200 {
		return errors.New(fmt.Sprintf("(%d) %s", statusCode, body))
	}

	return nil
}

func WaitForConnection(host string, port int) error {
	timeout := 30 * time.Second
	waiting := true
	result := make(chan error, 1)

	go func() {
		for waiting {
			statusCode, body, err := Get(host, port, "/", nil)

			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					continue // server not up yet -- keep trying
				}

				// if it's some other error, return it
				result <- err
				return
			}

			if statusCode != 200 {
				result <- errors.New(fmt.Sprintf("(%d) %s", statusCode, body))
				return
			}

			result <- nil
			return
		}
	}()

	select {
	case err := <-result:
		waiting = false
		return err
	case <-time.After(timeout):
		waiting = false
		return errors.New("Timed out trying to reach server.")
	}
}

func WaitForLackOfConnection(host string, port int) error {
	timeout := 30 * time.Second
	waiting := true
	result := make(chan error, 1)

	go func() {
		for waiting {
			_, _, err := Get(host, port, "/", nil)

			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					result <- nil
				}
			}
		}
	}()

	select {
	case serverDown := <-result:
		waiting = false
		return serverDown
	case <-time.After(timeout):
		waiting = false
		return errors.New("Timed out trying to stop server.")
	}
}

func formatUrl(host string, port int, endpoint string) string {
	return fmt.Sprintf("%s:%d%s", host, port, endpoint)
}

func handleResponse(res *httpclient.Response, err error) (int, string, error) {
	if err != nil {
		return 0, "", err
	}

	body, err := res.ToString()
	if err != nil {
		return 0, "", err
	}

	return res.StatusCode, body, nil
}

func Get(host string, port int, endpoint string, params map[string]string) (int, string, error) {
	url := formatUrl(host, port, endpoint)
	res, err := httpclient.Get(url, nil)
	return handleResponse(res, err)
}

func PostString(host string, port int, endpoint string, body string) (int, string, error) {
	url := formatUrl(host, port, endpoint)
	headers := make(map[string]string)
	bodyReader := strings.NewReader(body)

	res, err := httpclient.Do("POST", url, headers, bodyReader)
	return handleResponse(res, err)
}

func PutString(host string, port int, endpoint string, body string) (int, string, error) {
	url := formatUrl(host, port, endpoint)
	headers := make(map[string]string)
	bodyReader := strings.NewReader(body)

	res, err := httpclient.Do("PUT", url, headers, bodyReader)
	return handleResponse(res, err)
}

func PostFile(host string, port int, endpoint string, filename string) (int, string, error) {
	url := formatUrl(host, port, endpoint)
	res, err := httpclient.Post(url, map[string]string{"@file": filename})
	return handleResponse(res, err)
}

func PutFile(host string, port int, endpoint string, filename string) (int, string, error) {
	url := formatUrl(host, port, endpoint)
	res, err := httpclient.Put(url, map[string]string{"@file": filename})
	return handleResponse(res, err)
}

func Delete(host string, port int, endpoint string, params map[string]string) (int, string, error) {
	url := fmt.Sprintf("%s:%d%s", host, port, endpoint)
	res, err := httpclient.Delete(url, nil)
	return handleResponse(res, err)
}
