package ds

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Base64EncodeHeaders - encode headers to base64 stream of bytes
func Base64EncodeHeaders(headers map[string][]string) (enc []byte) {
	var err error
	enc, err = jsoniter.Marshal(headers)
	if err != nil {
		return
	}
	// Printf("Base64EncodeHeaders.1(%+v) --> %s\n", headers, string(enc))
	enc = []byte(base64.StdEncoding.EncodeToString(enc))
	// Printf("Base64EncodeHeaders.2(%+v) --> %s\n", headers, string(enc))
	return
}

// Base64DecodeHeaders - decode headers stored as stream of bytes to map of string arrays
func Base64DecodeHeaders(enc []byte) (headers map[string][]string, err error) {
	var bts []byte
	bts, err = base64.StdEncoding.DecodeString(string(enc))
	if err != nil {
		return
	}
	// Printf("Base64DecodeHeaders.1(%s) --> %+v\n", string(enc), string(bts))
	var result map[string]interface{}
	err = jsoniter.Unmarshal(bts, &result)
	// Printf("Base64DecodeHeaders.2(%s) --> %+v,%v\n", string(bts), result, err)
	headers = make(map[string][]string)
	for k, v := range result {
		ary, ok := v.([]interface{})
		if !ok {
			continue
		}
		sAry := []string{}
		for _, v := range ary {
			vs, ok := v.(string)
			if ok {
				sAry = append(sAry, vs)
			}
		}
		headers[k] = sAry
	}
	// Printf("Base64DecodeHeaders.3 --> %+v\n", headers)
	return
}

// Base64EncodeCookies - encode cookies array (strings) to base64 stream of bytes
func Base64EncodeCookies(cookies []string) (enc []byte) {
	last := len(cookies) - 1
	for i, cookie := range cookies {
		b := []byte(base64.StdEncoding.EncodeToString([]byte(cookie)))
		enc = append(enc, b...)
		if i != last {
			enc = append(enc, []byte("#")...)
		}
	}
	// Printf("Base64EncodeCookies(%d,%+v) --> %s\n", len(cookies), cookies, string(enc))
	return
}

// Base64DecodeCookies - decode cookies stored as stream of bytes to array of strings
func Base64DecodeCookies(enc []byte) (cookies []string, err error) {
	ary := bytes.Split(enc, []byte("#"))
	for _, item := range ary {
		var s []byte
		s, err = base64.StdEncoding.DecodeString(string(item))
		if err != nil {
			return
		}
		if len(s) > 0 {
			cookies = append(cookies, string(s))
		}
	}
	// Printf("Base64DecodeCookies(%s) --> %d,%+v\n", string(enc), len(cookies), cookies)
	return
}

// CookieToString - convert cookie to string
func CookieToString(c *http.Cookie) (s string) {
	// Other properties (skipped because login works without them)
	/*
	   Path       string
	   Domain     string
	   Expires    time.Time
	   RawExpires string
	   MaxAge   int
	   Secure   bool
	   HttpOnly bool
	   Raw      string
	   Unparsed []stringo
	*/
	if c.Name == "" && c.Value == "" {
		return
	}
	s = c.Name + "===" + c.Value
	// Printf("cookie %+v ----> %s\n", c, s)
	return
}

// StringToCookie - convert string to cookie
func StringToCookie(s string) (c *http.Cookie) {
	ary := strings.Split(s, "===")
	if len(ary) < 2 {
		return
	}
	c = &http.Cookie{Name: ary[0], Value: ary[1]}
	// Printf("cookie string %s ----> %+v\n", s, c)
	return
}

// RequestNoRetry - wrapper to do any HTTP request
// jsonStatuses - set of status code ranges to be parsed as JSONs
// errorStatuses - specify status value ranges for which we should return error
// okStatuses - specify status value ranges for which we should return error (only taken into account if not empty)
func RequestNoRetry(
	ctx *Ctx,
	url, method string,
	headers map[string]string,
	payload []byte,
	cookies []string,
	jsonStatuses, errorStatuses, okStatuses, cacheStatuses map[[2]int]struct{},
) (result interface{}, status int, isJSON bool, outCookies []string, outHeaders map[string][]string, cache bool, err error) {
	var (
		payloadBody *bytes.Reader
		req         *http.Request
	)
	if len(payload) > 0 {
		payloadBody = bytes.NewReader(payload)
		req, err = http.NewRequest(method, url, payloadBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
		err = fmt.Errorf("new request error:%+v for method:%s url:%s payload:%s", err, method, url, sPayload)
		return
	}
	for _, cookieStr := range cookies {
		cookie := StringToCookie(cookieStr)
		req.AddCookie(cookie)
	}
	for header, value := range headers {
		req.Header.Set(header, value)
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
		err = fmt.Errorf("do request error:%+v for method:%s url:%s headers:%v payload:%s", err, method, url, headers, sPayload)
		if strings.Contains(err.Error(), "socket: too many open files") {
			Printf("too many open socets detected, sleeping for 3 seconds\n")
			time.Sleep(time.Duration(3) * time.Second)
		}
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
		err = fmt.Errorf("read request body error:%+v for method:%s url:%s headers:%v payload:%s", err, method, url, headers, sPayload)
		return
	}
	_ = resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		outCookies = append(outCookies, CookieToString(cookie))
	}
	outHeaders = resp.Header
	status = resp.StatusCode
	hit := false
	for r := range jsonStatuses {
		if status >= r[0] && status <= r[1] {
			hit = true
			break
		}
	}
	if hit {
		err = jsoniter.Unmarshal(body, &result)
		if err != nil {
			sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
			sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
			err = fmt.Errorf("unmarshall request error:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s", err, method, url, headers, status, sPayload, sBody)
			return
		}
		isJSON = true
	} else {
		result = body
	}
	hit = false
	for r := range errorStatuses {
		if status >= r[0] && status <= r[1] {
			hit = true
			break
		}
	}
	if hit {
		sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		var sResult string
		bResult, bOK := result.([]byte)
		if bOK {
			sResult = BytesToStringTrunc(bResult, MaxPayloadPrintfLen, true)
		} else {
			sResult = InterfaceToStringTrunc(result, MaxPayloadPrintfLen, true)
		}
		err = fmt.Errorf("status error:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s result:%+v", err, method, url, headers, status, sPayload, sBody, sResult)
	}
	if len(okStatuses) > 0 {
		hit = false
		for r := range okStatuses {
			if status >= r[0] && status <= r[1] {
				hit = true
				break
			}
		}
		if !hit {
			sPayload := BytesToStringTrunc(payload, MaxPayloadPrintfLen, true)
			sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
			var sResult string
			bResult, bOK := result.([]byte)
			if bOK {
				sResult = BytesToStringTrunc(bResult, MaxPayloadPrintfLen, true)
			} else {
				sResult = InterfaceToStringTrunc(result, MaxPayloadPrintfLen, true)
			}
			err = fmt.Errorf("status not success:%+v for method:%s url:%s headers:%v status:%d payload:%s body:%s result:%+v", err, method, url, headers, status, sPayload, sBody, sResult)
		}
	}
	if err == nil {
		for r := range cacheStatuses {
			if status >= r[0] && status <= r[1] {
				cache = true
				break
			}
		}
	}
	return
}

// Request - wrapper around RequestNoRetry supporting retries
func Request(
	ctx *Ctx,
	url, method string,
	headers map[string]string,
	payload []byte,
	cookies []string,
	jsonStatuses, errorStatuses, okStatuses, cacheStatuses map[[2]int]struct{},
	retryRequest bool,
	cacheFor *time.Duration,
	skipInDryRun bool,
) (result interface{}, status int, outCookies []string, outHeaders map[string][]string, err error) {
	if skipInDryRun && ctx.DryRun {
		if ctx.Debug > 0 {
			Printf("dry-run: %s.%s(#h=%d,pl=%d,cks=%d) skipped in dry-run mode\n", method, url, len(headers), len(payload), len(cookies))
		}
		return
	}
	var (
		isJSON bool
		cache  bool
	)
	// fmt.Printf("url=%s method=%s headers=%+v payload=%+v cookies=%+v\n", url, method, headers, payload, cookies)
	if cacheFor != nil && !ctx.NoCache {
		// cacheKey is hash(method,url,headers,payload,cookies
		b := []byte(method + url + fmt.Sprintf("%+v", headers))
		b = append(b, payload...)
		b = append(b, []byte(strings.Join(cookies, "==="))...)
		hash := sha1.New()
		_, e := hash.Write(b)
		if e == nil {
			hsh := hex.EncodeToString(hash.Sum(nil))
			cached, ok := GetL2Cache(ctx, hsh)
			if ok {
				// cache entry is 'status:isJson:b64cookies:headers:data
				ary := bytes.Split(cached, []byte(":"))
				if len(ary) >= 5 {
					var e error
					status, e = strconv.Atoi(string(ary[0]))
					if e == nil {
						var iJSON int
						iJSON, e = strconv.Atoi(string(ary[1]))
						if e == nil {
							outCookies, e = Base64DecodeCookies(ary[2])
							if e == nil {
								outHeaders, e = Base64DecodeHeaders(ary[3])
								if e == nil {
									resData := bytes.Join(ary[4:], []byte(":"))
									if iJSON == 0 {
										result = resData
										return
									}
									var r interface{}
									e = jsoniter.Unmarshal(resData, &r)
									if e == nil {
										result = r
										return
									}
								}
							}
						}
					}
				}
			}
			cacheDuration := *cacheFor
			defer func() {
				if err != nil || !cache {
					return
				}
				// cache entry is 'status:isJson:b64cookies:headers:data
				b64cookies := Base64EncodeCookies(outCookies)
				b64headers := Base64EncodeHeaders(outHeaders)
				data := []byte(fmt.Sprintf("%d:", status))
				if isJSON {
					bts, e := jsoniter.Marshal(result)
					if e != nil {
						return
					}
					data = append(data, []byte("1:")...)
					data = append(data, b64cookies...)
					data = append(data, []byte(":")...)
					data = append(data, b64headers...)
					data = append(data, []byte(":")...)
					data = append(data, bts...)
					tag := FilterRedacted(fmt.Sprintf("%s.%s(#h=%d,pl=%d,cks=%d) -> sts=%d,js=1,resp=%d,cks=%d,hdrs=%d", method, url, len(headers), len(payload), len(cookies), status, len(bts), len(outCookies), len(outHeaders)))
					SetL2Cache(ctx, hsh, tag, data, cacheDuration)
					return
				}
				data = append(data, []byte("0:")...)
				data = append(data, b64cookies...)
				data = append(data, []byte(":")...)
				data = append(data, b64headers...)
				data = append(data, []byte(":")...)
				data = append(data, result.([]byte)...)
				tag := FilterRedacted(fmt.Sprintf("%s.%s(#h=%d,pl=%d,cks=%d) -> sts=%d,js=0,resp=%d,cks=%d,hdrs=%d", method, url, len(headers), len(payload), len(cookies), status, len(result.([]byte)), len(outCookies), len(outHeaders)))
				SetL2Cache(ctx, hsh, tag, data, cacheDuration)
				return
			}()
		}
	}
	if !retryRequest {
		result, status, isJSON, outCookies, outHeaders, cache, err = RequestNoRetry(ctx, url, method, headers, payload, cookies, jsonStatuses, errorStatuses, okStatuses, cacheStatuses)
		return
	}
	retry := 0
	for {
		result, status, isJSON, outCookies, outHeaders, cache, err = RequestNoRetry(ctx, url, method, headers, payload, cookies, jsonStatuses, errorStatuses, okStatuses, cacheStatuses)
		info := func() (inf string) {
			inf = fmt.Sprintf("%s.%s:%s=%d", method, url, BytesToStringTrunc(payload, MaxPayloadPrintfLen, true), status)
			if ctx.Debug > 1 {
				inf += fmt.Sprintf(" error: %+v", err)
			} else if err != nil {
				inf += fmt.Sprintf(" error: %+v", StringTrunc(err.Error(), MaxPayloadPrintfLen, true))
			}
			return
		}
		if err != nil {
			retry++
			if retry > ctx.Retry {
				Printf("%s failed after %d retries\n", info(), retry)
				return
			}
			seconds := (retry + 1) * (retry + 1)
			Printf("will do #%d retry of %s after %d seconds\n", retry, info(), seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
			Printf("retrying #%d retry of %s after %d seconds\n", retry, info(), seconds)
			continue
		}
		if retry > 0 {
			Printf("#%d retry of %s succeeded\n", retry, info())
		}
		return
	}
}
