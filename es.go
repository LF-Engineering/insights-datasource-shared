package ds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	esCacheMtx *sync.RWMutex
)

// ESCacheEntry - single cache entry
type ESCacheEntry struct {
	K string    `json:"k"` // cache key
	G string    `json:"g"` // cache tag
	B []byte    `json:"b"` // cache data
	T time.Time `json:"t"` // when cached
	E time.Time `json:"e"` // when expires
}

// ESCacheGet - get value from cache
func ESCacheGet(ctx *Ctx, key string) (entry *ESCacheEntry, ok bool) {
	if ctx.ESURL == "" {
		return
	}
	data := `{"query":{"term":{"k.keyword":{"value": "` + JSONEscape(key) + `"}}}}`
	payloadBytes := []byte(data)
	payloadBody := bytes.NewReader(payloadBytes)
	method := "POST"
	url := fmt.Sprintf("%s/dads_cache/_search", ctx.ESURL)
	req, err := http.NewRequest(method, url, payloadBody)
	if err != nil {
		Printf("New request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Printf("do request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		Printf("ReadAll non-ok request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		Printf("Method:%s url:%s data: %s status:%d\n%s\n", method, url, data, resp.StatusCode, sBody)
		return
	}
	type R struct {
		H struct {
			H []struct {
				S ESCacheEntry `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	var r R
	err = jsoniter.Unmarshal(body, &r)
	if err != nil {
		Printf("Unmarshal error: %+v\n", err)
		return
	}
	if len(r.H.H) == 0 {
		return
	}
	entry = &(r.H.H[0].S)
	ok = true
	return
}

// ESCacheSet - set cache value
func ESCacheSet(ctx *Ctx, key string, entry *ESCacheEntry) {
	if ctx.ESURL == "" {
		return
	}
	entry.K = key
	payloadBytes, err := jsoniter.Marshal(entry)
	if err != nil {
		sEntry := "none"
		if entry != nil {
			sEntry = InterfaceToStringTrunc(*entry, MaxPayloadPrintfLen, true)
		}
		Printf("json %s marshal error: %+v\n", sEntry, err)
		return
	}
	payloadBody := bytes.NewReader(payloadBytes)
	method := "POST"
	url := fmt.Sprintf("%s/dads_cache/_doc?refresh=true", ctx.ESURL)
	req, err := http.NewRequest(method, url, payloadBody)
	if err != nil {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		Printf("New request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		Printf("do request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 201 {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Printf("ReadAll non-ok request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
			return
		}
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		Printf("Method:%s url:%s data: %s status:%d\n%s\n", method, url, sData, resp.StatusCode, sBody)
		return
	}
	return
}

// ESCacheDelete - delete cache key
func ESCacheDelete(ctx *Ctx, key string) {
	if ctx.ESURL == "" {
		return
	}
	data := `{"query":{"term":{"k.keyword":{"value": "` + JSONEscape(key) + `"}}}}`
	payloadBytes := []byte(data)
	payloadBody := bytes.NewReader(payloadBytes)
	method := "POST"
	url := fmt.Sprintf("%s/dads_cache/_delete_by_query?conflicts=proceed&refresh=true", ctx.ESURL)
	req, err := http.NewRequest(method, url, payloadBody)
	if err != nil {
		Printf("New request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Printf("do request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Printf("ReadAll non-ok request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
			return
		}
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		Printf("Method:%s url:%s data: %s status:%d\n%s\n", method, url, data, resp.StatusCode, sBody)
		return
	}
}

// ESCacheDeleteExpired - delete expired cache entries
func ESCacheDeleteExpired(ctx *Ctx) {
	if ctx.Debug > 1 {
		Printf("running ESCacheDeleteExpired\n")
	}
	if ctx.ESURL == "" {
		return
	}
	data := `{"query":{"range":{"e":{"lte": "now"}}}}`
	payloadBytes := []byte(data)
	payloadBody := bytes.NewReader(payloadBytes)
	method := "POST"
	url := fmt.Sprintf("%s/dads_cache/_delete_by_query?conflicts=proceed&refresh=true", ctx.ESURL)
	req, err := http.NewRequest(method, url, payloadBody)
	if err != nil {
		Printf("New request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		Printf("do request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Printf("ReadAll non-ok request error: %+v for %s url: %s, data: %s\n", err, method, url, data)
			return
		}
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		Printf("Method:%s url:%s data: %s status:%d\n%s\n", method, url, data, resp.StatusCode, sBody)
		return
	}
}

// GetESCache - get value from cache - thread safe and support expiration
func GetESCache(ctx *Ctx, k string) (b []byte, tg string, expires time.Time, ok bool) {
	defer MaybeESCacheCleanup(ctx)
	if MT {
		esCacheMtx.RLock()
	}
	entry, ok := ESCacheGet(ctx, k)
	if MT {
		esCacheMtx.RUnlock()
	}
	if !ok {
		if ctx.Debug > 1 {
			Printf("GetESCache(%s): miss\n", k)
		}
		return
	}
	if time.Now().After(entry.E) {
		ok = false
		if MT {
			esCacheMtx.Lock()
		}
		ESCacheDelete(ctx, k)
		if MT {
			esCacheMtx.Unlock()
		}
		if ctx.Debug > 1 {
			Printf("GetESCache(%s,%s): expired %v\n", k, entry.G, entry.E)
		}
		return
	}
	b = entry.B
	tg = entry.G
	expires = entry.E
	if ctx.Debug > 1 {
		Printf("GetESCache(%s,%s): hit (%v)\n", k, tg, expires)
	}
	return
}

// GetL2Cache - get value from cache - thread safe and support expiration
func GetL2Cache(ctx *Ctx, k string) (b []byte, ok bool) {
	defer MaybeMemCacheCleanup(ctx)
	if MT {
		memCacheMtx.RLock()
	}
	entry, ok := memCache[k]
	if MT {
		memCacheMtx.RUnlock()
	}
	if !ok {
		if ctx.Debug > 1 {
			Printf("GetL2Cache(%s): miss\n", k)
		}
		var (
			g string
			e time.Time
		)
		b, g, e, ok = GetESCache(ctx, k)
		if ok {
			t := time.Now()
			if MT {
				memCacheMtx.Lock()
			}
			memCache[k] = &MemCacheEntry{G: g, B: b, T: t, E: e}
			if MT {
				memCacheMtx.Unlock()
			}
			if ctx.Debug > 1 {
				Printf("GetL2Cache(%s,%s): L2 hit (%v)\n", k, g, e)
			}
		}
		return
	}
	if time.Now().After(entry.E) {
		ok = false
		if MT {
			memCacheMtx.Lock()
		}
		delete(memCache, k)
		if MT {
			memCacheMtx.Unlock()
		}
		if ctx.Debug > 1 {
			Printf("GetL2Cache(%s,%s): expired %v\n", k, entry.G, entry.E)
		}
		var (
			g string
			e time.Time
		)
		b, g, e, ok = GetESCache(ctx, k)
		if ok {
			t := time.Now()
			if MT {
				memCacheMtx.Lock()
			}
			memCache[k] = &MemCacheEntry{G: g, B: b, T: t, E: e}
			if MT {
				memCacheMtx.Unlock()
			}
			if ctx.Debug > 1 {
				Printf("GetL2Cache(%s,%s): L2 hit (%v)\n", k, g, e)
			}
		}
		return
	}
	b = entry.B
	if ctx.Debug > 1 {
		Printf("GetL2Cache(%s,%s): hit (%v)\n", k, entry.G, entry.E)
	}
	return
}

// SetESCache - set cache value, expiration date and handles multithreading etc
func SetESCache(ctx *Ctx, k, tg string, b []byte, expires time.Duration) {
	defer MaybeESCacheCleanup(ctx)
	t := time.Now()
	e := t.Add(expires)
	if MT {
		esCacheMtx.RLock()
	}
	_, ok := ESCacheGet(ctx, k)
	if MT {
		esCacheMtx.RUnlock()
	}
	if ok {
		if MT {
			esCacheMtx.Lock()
		}
		ESCacheDelete(ctx, k)
		ESCacheSet(ctx, k, &ESCacheEntry{B: b, T: t, E: e, G: tg})
		if MT {
			esCacheMtx.Unlock()
		}
		if ctx.Debug > 1 {
			Printf("SetESCache(%s,%s): replaced (%v)\n", k, tg, e)
		}
	} else {
		if MT {
			esCacheMtx.Lock()
		}
		ESCacheSet(ctx, k, &ESCacheEntry{B: b, T: t, E: e, G: tg})
		if MT {
			esCacheMtx.Unlock()
		}
		if ctx.Debug > 1 {
			Printf("SetESCache(%s,%s): added (%v)\n", k, tg, e)
		}
	}
}

// SetL2Cache - set cache value, expiration date and handles multithreading etc
func SetL2Cache(ctx *Ctx, k, tg string, b []byte, expires time.Duration) {
	defer MaybeMemCacheCleanup(ctx)
	SetESCache(ctx, k, tg, b, expires)
	t := time.Now()
	e := t.Add(expires)
	if MT {
		memCacheMtx.Lock()
	}
	_, ok := memCache[k]
	memCache[k] = &MemCacheEntry{G: tg, B: b, T: t, E: e}
	if MT {
		memCacheMtx.Unlock()
	}
	if ok {
		if ctx.Debug > 1 {
			Printf("SetL2Cache(%s,%s): replaced (%v)\n", k, tg, e)
		}
		return
	}
	if ctx.Debug > 1 {
		Printf("SetL2Cache(%s,%s): added (%v)\n", k, tg, e)
	}
}

// MaybeESCacheCleanup - chance of cleaning expired cache entries
func MaybeESCacheCleanup(ctx *Ctx) {
	// chance for cache cleanup
	if rand.Intn(100) < CacheCleanupProb {
		go func() {
			if MT {
				esCacheMtx.Lock()
			}
			ESCacheDeleteExpired(ctx)
			if MT {
				esCacheMtx.Unlock()
			}
			if ctx.Debug > 2 {
				Printf("ContributorsCache: deleted expired items\n")
			}
		}()
	}
}

// CreateESCache - creates dads_cache index needed for caching
func CreateESCache(ctx *Ctx) {
	if ctx.ESURL == "" {
		return
	}
	// Create index, ignore if exists (see status 400 is not in error statuses)
	_, _, _, _, err := Request(ctx, ctx.ESURL+"/dads_cache", "PUT", nil, []byte{}, []string{}, nil, map[[2]int]struct{}{{401, 599}: {}}, nil, nil, false, nil, false)
	FatalOnError(err)
}

// GetLastUpdate - get last update date from ElasticSearch
func GetLastUpdate(ctx *Ctx, key string) (lastUpdate *time.Time) {
	if ctx.ESURL == "" {
		return
	}
	// curl -s -XPOST -H 'Content-type: application/json' '${URL}/last-update-cache/_search?size=0' -d '{"query":{"bool":{"filter":{"term":{"key.keyword":"ds:endpoint"}}}},"aggs":{"m":{"max":{"field":"last_update"}}}}' | jq -r '.aggregations.m.value_as_string'
	escapedKey := JSONEscape(ctx.DS + ":" + key)
	payloadBytes := []byte(`{"query":{"bool":{"filter":{"term":{"key.keyword":"` + escapedKey + `"}}}},"aggs":{"m":{"max":{"field":"last_update"}}}}`)
	url := ctx.ESURL + "/last-update-cache/_search?size=0"
	if ctx.Debug > 0 {
		Printf("resume from date query key=%s: %s\n", escapedKey, string(payloadBytes))
	}
	method := "POST"
	resp, status, _, _, err := Request(
		ctx,
		url,
		method,
		map[string]string{"Content-Type": "application/json"}, // headers
		payloadBytes, // payload
		[]string{},   // cookies
		nil,          // JSON statuses
		nil,          // Error statuses
		map[[2]int]struct{}{
			{200, 200}: {},
			{404, 404}: {},
		}, // OK statuses
		nil,   // Cache statuses
		false, // retry
		nil,   // cache for
		false, // skip in dry-run mode
	)
	if status == 404 {
		return
	}
	FatalOnError(err)
	type resultStruct struct {
		Aggs struct {
			M struct {
				Str string `json:"value_as_string"`
			} `json:"m"`
		} `json:"aggregations"`
	}
	var res resultStruct
	err = jsoniter.Unmarshal(resp.([]byte), &res)
	if err != nil {
		Printf("resume from date JSON decode error: %+v for %s url: %s, query: %s\n", err, method, url, string(payloadBytes))
		return
	}
	if res.Aggs.M.Str != "" {
		var tm time.Time
		tm, err = TimeParseAny(res.Aggs.M.Str)
		if err != nil {
			Printf("resume from date decode aggregations error: %+v for %s url: %s, query: %s\n", err, method, url, string(payloadBytes))
			return
		}
		lastUpdate = &tm
	}
	return
}

// SetLastUpdate - set last update date for a given data source
func SetLastUpdate(ctx *Ctx, key string, when time.Time) {
	if ctx.ESURL == "" {
		return
	}
	escapedKey := JSONEscape(ctx.DS + ":" + key)
	type docType struct {
		Key        string    `json:"key"`
		LastUpdate time.Time `json:"last_update"`
		SavedAt    time.Time `json:"saved_at"`
	}
	doc := docType{Key: escapedKey, LastUpdate: when, SavedAt: time.Now()}
	payloadBytes, err := jsoniter.Marshal(doc)
	if err != nil {
		Printf("json %s marshal error: %+v\n", doc, err)
		return
	}
	payloadBody := bytes.NewReader(payloadBytes)
	method := "POST"
	url := fmt.Sprintf("%s/last-update-cache/_doc?refresh=true", ctx.ESURL)
	req, err := http.NewRequest(method, url, payloadBody)
	if err != nil {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		Printf("New request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		Printf("do request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 201 {
		sData := BytesToStringTrunc(payloadBytes, MaxPayloadPrintfLen, true)
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Printf("ReadAll non-ok request error: %+v for %s url: %s, data: %s\n", err, method, url, sData)
			return
		}
		sBody := BytesToStringTrunc(body, MaxPayloadPrintfLen, true)
		Printf("Method:%s url:%s data: %s status:%d\n%s\n", method, url, sData, resp.StatusCode, sBody)
		return
	}
	return
}
