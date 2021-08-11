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
	// KeywordMaxlength - max description length
	KeywordMaxlength = 1000
	// MaxBodyLength - max length of body to store
	MaxBodyLength = 0x40000
	// MissingName - common constant string
	MissingName = "-MISSING-NAME"
	// RedactedEmail - common constant string
	RedactedEmail = "-REDACTED-EMAIL"
)

var (
	memCacheMtx *sync.RWMutex
	memCache    = map[string]*MemCacheEntry{}
	// RawFields - standard raw fields
	RawFields = []string{"metadata__updated_on", "metadata__timestamp", "origin", "tags", "uuid", "offset"}
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

// Dig interface for array of keys
func Dig(iface interface{}, keys []string, fatal, silent bool) (v interface{}, ok bool) {
	miss := false
	defer func() {
		if !ok && fatal {
			Fatalf("cannot dig %+v in %s", keys, DumpKeys(iface))
		}
	}()
	item, o := iface.(map[string]interface{})
	if !o {
		if !silent {
			Printf("Interface cannot be parsed: %+v\n", iface)
		}
		return
	}
	last := len(keys) - 1
	for i, key := range keys {
		var o bool
		if i < last {
			item, o = item[key].(map[string]interface{})
		} else {
			v, o = item[key]
		}
		if !o {
			if !silent {
				Printf("dig %+v, current: %s, %d/%d failed\n", keys, key, i+1, last+1)
			}
			miss = true
			break
		}
	}
	ok = !miss
	return
}

// DumpKeys - dump interface structure, but only keys, no values
func DumpKeys(i interface{}) string {
	return strings.Replace(fmt.Sprintf("%v", KeysOnly(i)), "map[]", "", -1)
}

// PreviewOnly - return a corresponding interface with preview values
func PreviewOnly(i interface{}, l int) (o interface{}) {
	if i == nil {
		return
	}
	is, ok := i.(map[string]interface{})
	if !ok {
		str := InterfaceToStringTrunc(i, l, false)
		str = strings.Replace(str, "\n", " ", -1)
		o = str
		return
	}
	iface := make(map[string]interface{})
	for k, v := range is {
		iface[k] = PreviewOnly(v, l)
	}
	o = iface
	return
}

// KeysOnly - return a corresponding interface contining only keys
func KeysOnly(i interface{}) (o map[string]interface{}) {
	if i == nil {
		return
	}
	is, ok := i.(map[string]interface{})
	if !ok {
		return
	}
	o = make(map[string]interface{})
	for k, v := range is {
		o[k] = KeysOnly(v)
	}
	return
}

// DeepSet - set deep property of non-type decoded interface
func DeepSet(m interface{}, ks []string, v interface{}, create bool) (err error) {
	c, ok := m.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("cannot access %v as a string map", m)
		return
	}
	last := len(ks) - 1
	for i, k := range ks {
		if i < last {
			obj, ok := c[k]
			if !ok {
				if create {
					c[k] = make(map[string]interface{})
					obj = c[k]
				} else {
					err = fmt.Errorf("cannot access #%d key %s from %v, all keys %v", i+1, k, DumpKeys(c), ks)
					return
				}
			}
			c, ok = obj.(map[string]interface{})
			if !ok {
				err = fmt.Errorf("cannot access %v as a string map, #%d key %s, all keys %v", c, i+1, k, ks)
				return
			}
			continue
		}
		c[k] = v
	}
	return
}

// RedactEmail - possibly redact email from "in"
// If in contains @, replace part after last "@" with suff
// If in doesn't contain "@" then return it or (if forceSuff is set) return in + suff
func RedactEmail(in, suff string, forceSuff bool) string {
	ary := strings.Split(strings.Trim(strings.TrimSpace(in), "@"), "@")
	n := len(ary)
	if n <= 1 {
		if forceSuff {
			return in + suff
		}
		return in
	}
	return strings.TrimSpace(strings.Join(ary[:n-1], "@")) + suff
}

// PostprocessNameUsername - check name field, if it is empty then copy from email (if not empty) or username (if not empty)
// Then check name and username - it cannot contain email addess, if it does - replace a@domain with a-MISSING-NAME
func PostprocessNameUsername(name, username, email string) (outName, outUsername string) {
	defer func() {
		outName = name
		outUsername = username
	}()
	copiedName := false
	if name == "" || name == "none" {
		if email != "" && email != "none" {
			name = RedactEmail(email, MissingName, true)
			copiedName = true
		}
	}
	if name == "" || name == "none" {
		if username != "" && username != "none" {
			name = RedactEmail(username, MissingName, true)
			copiedName = true
		}
	}
	if !copiedName && name != "" && name != "none" {
		name = RedactEmail(name, RedactedEmail, false)
	}
	if username != "" && username != "none" {
		username = RedactEmail(username, RedactedEmail, false)
	}
	return
}
