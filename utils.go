package ds

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// MaxPayloadPrintfLen - truncate messages longer than this
	MaxPayloadPrintfLen = 0x2000
	// CacheCleanupProb - 2% chance of cleaning the cache
	CacheCleanupProb = 2
)

var (
	memCacheMtx *sync.RWMutex
	memCache    = map[string]*MemCacheEntry{}
)

// MemCacheEntry - single cache entry
type MemCacheEntry struct {
	G string    `json:"g"` // cache tag
	B []byte    `json:"b"` // cache data
	T time.Time `json:"t"` // when cached
	E time.Time `json:"e"` // when expires
}

// MemCacheDeleteExpired - delete expired cache entries
func MemCacheDeleteExpired(ctx *Ctx) {
	t := time.Now()
	ks := []string{}
	for k, v := range memCache {
		if t.After(v.E) {
			ks = append(ks, k)
		}
	}
	if ctx.Debug > 1 {
		Printf("running MemCacheDeleteExpired - deleting %d entries\n", len(ks))
	}
	for _, k := range ks {
		delete(memCache, k)
	}
}

// MaybeMemCacheCleanup - chance of cleaning expired cache entries
func MaybeMemCacheCleanup(ctx *Ctx) {
	// chance for cache cleanup
	if rand.Intn(100) < CacheCleanupProb {
		go func() {
			if MT {
				memCacheMtx.Lock()
			}
			MemCacheDeleteExpired(ctx)
			if MT {
				memCacheMtx.Unlock()
			}
		}()
	}
}

// StringToBool - convert string value to boolean value
// returns false for anything that was parsed as false, zero, empty etc:
// f, F, false, False, fALSe, 0, "", 0.00
// else returns true
func StringToBool(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err == nil {
		return b
	}
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return f != 0.0
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err == nil {
		return i != 0
	}
	if v == "no" || v == "n" {
		return false
	}
	return true
}

// NoSSLVerify - turn off SSL validation
func NoSSLVerify() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

// FlagPassed - was that flag actually passed (returns true) or the default value was used? (returns false)
func FlagPassed(ctx *Ctx, name string) bool {
	name = ctx.DSFlag + name
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// StringTrunc - truncate string to no more than maxLen
func StringTrunc(data string, maxLen int, addLenInfo bool) (str string) {
	lenInfo := ""
	if addLenInfo {
		lenInfo = "(" + strconv.Itoa(len(data)) + "): "
	}
	if len(data) <= maxLen {
		return lenInfo + data
	}
	half := maxLen >> 1
	str = lenInfo + data[:half] + "(...)" + data[len(data)-half:]
	return
}

// BytesToStringTrunc - truncate bytes stream to no more than maxLen
func BytesToStringTrunc(data []byte, maxLen int, addLenInfo bool) (str string) {
	lenInfo := ""
	if addLenInfo {
		lenInfo = "(" + strconv.Itoa(len(data)) + "): "
	}
	if len(data) <= maxLen {
		return lenInfo + string(data)
	}
	half := maxLen >> 1
	str = lenInfo + string(data[:half]) + "(...)" + string(data[len(data)-half:])
	return
}

// InterfaceToStringTrunc - truncate interface representation
func InterfaceToStringTrunc(iface interface{}, maxLen int, addLenInfo bool) (str string) {
	data := fmt.Sprintf("%+v", iface)
	lenInfo := ""
	if addLenInfo {
		lenInfo = "(" + strconv.Itoa(len(data)) + "): "
	}
	if len(data) <= maxLen {
		return lenInfo + data
	}
	half := maxLen >> 1
	str = "(" + strconv.Itoa(len(data)) + "): " + data[:half] + "(...)" + data[len(data)-half:]
	return
}
