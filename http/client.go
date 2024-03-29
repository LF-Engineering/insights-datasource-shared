package http

import (
	"bytes"
	"encoding/csv"
	"log"
	"net/http"
	"time"
)

// ClientProvider ...
type ClientProvider struct {
	httpclient *http.Client
}

// NewClientProvider initiate a new client object
//
//	timeout - http request time
//	allowRedirect - allow request redirects on requested url or not
func NewClientProvider(timeout time.Duration, allowRedirect bool) *ClientProvider {
	CancelRedirect := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	httpClientProvider := &ClientProvider{
		httpclient: &http.Client{
			Timeout: timeout,
		},
	}
	if !allowRedirect {
		httpClientProvider.httpclient.CheckRedirect = CancelRedirect
	}
	return httpClientProvider
}

// Response returned from http request
type Response struct {
	StatusCode int
	Body       []byte
}

// Request http
func (h *ClientProvider) Request(url string, method string, header map[string]string, body []byte, params map[string]string) (statusCode int, resBody []byte, err error) {
	resp, resBody, err := h.RequestWithResponse(url, method, header, body, params)
	if err != nil || resBody == nil {
		return 0, nil, err
	}
	return resp.StatusCode, resBody, err
}

// RequestWithResponse http
func (h *ClientProvider) RequestWithResponse(url string, method string, header map[string]string, body []byte, params map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}

	if cType, ok := header["Content-Type"]; !ok || cType == "application/json" {
		req.Header.Add("Content-Type", "application/json")
		delete(header, "Content-Type")
	}

	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Do request
	res, err := h.httpclient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return res, buf.Bytes(), nil
}

// RequestWithRequest http
func (h *ClientProvider) RequestWithRequest(req *http.Request, header map[string]string, params map[string]string) (*http.Response, []byte, error) {
	if cType, ok := header["Content-Type"]; !ok || cType == "application/json" {
		req.Header.Add("Content-Type", "application/json")
		delete(header, "Content-Type")
	}

	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Do request
	res, err := h.httpclient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return res, buf.Bytes(), nil
}

// RequestCSV requests http API that returns csv result
func (h *ClientProvider) RequestCSV(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Err: %s", err.Error())
		}
	}()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}
