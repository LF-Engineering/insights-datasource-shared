package ds

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LF-Engineering/lfx-event-schema/service/insights"

	jsoniter "github.com/json-iterator/go"
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
	// DefaultRateLimitHeader - default value for rate limit header
	DefaultRateLimitHeader = "X-RateLimit-Remaining"
	// DefaultRateLimitResetHeader - default value for rate limit reset header
	DefaultRateLimitResetHeader = "X-RateLimit-Reset"
)

var (
	memCacheMtx *sync.RWMutex
	memCache    = map[string]*MemCacheEntry{}
	// RawFields - standard raw fields
	RawFields = []string{"metadata__updated_on", "metadata__timestamp", "origin", "tags", "uuid", "offset"}
	// postprocCache validation cache
	postprocCache = map[[3]string][2]string{}
	// postprocCacheMtx - emails validation cache mutex
	postprocCacheMtx *sync.RWMutex
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
		if MT {
			go func() {
				memCacheMtx.Lock()
				MemCacheDeleteExpired(ctx)
				memCacheMtx.Unlock()
			}()
			return
		}
		MemCacheDeleteExpired(ctx)
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

// DumpPreview - dump interface structure, keys and truncated values preview
func DumpPreview(i interface{}, l int) string {
	return strings.Replace(fmt.Sprintf("%v", PreviewOnly(i, l)), "map[]", "", -1)
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
	if MT {
		postprocCacheMtx.RLock()
	}
	data, ok := postprocCache[[3]string{name, username, email}]
	if MT {
		postprocCacheMtx.RUnlock()
	}
	if ok {
		outName = data[0]
		outUsername = data[1]
		return
	}
	inName, inUsername, inEmail := name, username, email
	defer func() {
		outName = name
		outUsername = username
		if MT {
			postprocCacheMtx.Lock()
		}
		postprocCache[[3]string{inName, inUsername, inEmail}] = [2]string{outName, outUsername}
		if MT {
			postprocCacheMtx.Unlock()
		}
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

// EnsurePath - craete archive directory (and all necessary parents as well)
// if noLastDir is set, then skip creating the last directory in the path
func EnsurePath(path string, noLastDir bool) (string, error) {
	ary := strings.Split(path, "/")
	nonEmpty := []string{}
	for i, dir := range ary {
		if i > 0 && dir == "" {
			continue
		}
		nonEmpty = append(nonEmpty, dir)
	}
	path = strings.Join(nonEmpty, "/")
	var createPath string
	if noLastDir {
		createPath = strings.Join(nonEmpty[:len(nonEmpty)-1], "/")
	} else {
		createPath = path
	}
	return path, os.MkdirAll(createPath, 0755)
}

// MatchGroups - return regular expression matching groups as a map
func MatchGroups(re *regexp.Regexp, arg string) (result map[string]string) {
	match := re.FindStringSubmatch(arg)
	result = make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[name] = match[i]
		}
	}
	return
}

// MatchGroupsArray - return regular expression matching groups as a map
func MatchGroupsArray(re *regexp.Regexp, arg string) (result map[string][]string) {
	match := re.FindAllStringSubmatch(arg, -1)
	//Printf("match(%d,%d): %+v\n", len(match), len(re.SubexpNames()), match)
	result = make(map[string][]string)
	names := re.SubexpNames()
	names = names[1:]
	for idx, m := range match {
		if idx == 0 {
			for i, name := range names {
				result[name] = []string{m[i+1]}
			}
			continue
		}
		for i, name := range names {
			ary := result[name]
			result[name] = append(ary, m[i+1])
		}
	}
	return
}

// UniqueStringArray - make array unique
func UniqueStringArray(ary []interface{}) []interface{} {
	m := map[string]struct{}{}
	for _, i := range ary {
		m[i.(string)] = struct{}{}
	}
	ret := []interface{}{}
	for i := range m {
		ret = append(ret, i)
	}
	return ret
}

// IndexAt - index of substring starting at a given position
func IndexAt(s, sep string, n int) int {
	idx := strings.Index(s[n:], sep)
	if idx > -1 {
		idx += n
	}
	return idx
}

// PartitionString - partition a string to [pre-sep, sep, post-sep]
func PartitionString(s string, sep string) [3]string {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return [3]string{parts[0], "", ""}
	}
	return [3]string{parts[0], sep, parts[1]}
}

// PrettyPrint - print any data as json
func PrettyPrint(data interface{}) string {
	j := jsoniter.Config{SortMapKeys: true, EscapeHTML: true}.Froze()
	pretty, err := j.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%T: %+v", data, data)
	}
	return string(pretty)
}

// AsJSON - print any data as json
func AsJSON(data interface{}) string {
	j := jsoniter.Config{SortMapKeys: true, EscapeHTML: false}.Froze()
	pretty, err := j.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%T: %+v", data, data)
	}
	return string(pretty)
}

// DedupContributors - there can be no multiple contributors having the same ID & role
func DedupContributors(inContributors []insights.Contributor) (outContributors []insights.Contributor) {
	m := make(map[string]struct{})
	for _, contributor := range inContributors {
		key := string(contributor.Role) + ":" + contributor.Identity.ID
		_, found := m[key]
		if !found {
			outContributors = append(outContributors, contributor)
			m[key] = struct{}{}
		}
	}
	return
}

// StripURL - return only host + path from URL, example: 'https://user:password@github.com/cncf/devstats?foo=bar&foo=baz#readme' -> 'github.com/cncf/devstats'
func StripURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		Printf("StripURL: '%s' is not a correct URL, returning unchanged", urlStr)
		return urlStr
	}
	return u.Host + u.Path
}

// IsBotIdentity check if username is for a bot identity
func IsBotIdentity(name, username, email, datasource string) bool {
	if datasource == "git" || datasource == "github" || datasource == "gerrit" {
		if name != "" {
			nameR := regexp.MustCompile(botNameR)
			if nameR.MatchString(name) {
				return true
			}
		}
		if username != "" {
			usernameR := regexp.MustCompile(botUsernameR)
			if usernameR.MatchString(username) {
				return true
			}
		} else if email != "" {
			emailR := regexp.MustCompile(botEmailR)
			if emailR.MatchString(email) {
				return true
			}
		}
	}
	return false
}

var (
	botNameR     = `^(facebook-github-whois-bot-0|fossabot|claassistant|containersshbuilder|knative-automation|covbot|cdk8s-automation|github-action-benchmark|wasmcloud-automation|goreleaserbot|imgbotapp|backstage-service|openssl-machine|sizebot|dependabot|cncf-ci|poiana|svcbot-qecnsdp|nsmbot|ti-srebot|cf-buildpacks-eng|bosh-ci-push-pull|gprasath|zephyr-github|zephyrbot|strimzi-ci|athenabot|k8s-reviewable|codecov-io|grpc-testing|k8s-teamcity-mesosphere|angular-builds|devstats-sync|googlebot|hibernate-ci|coveralls|rktbot|coreosbot|web-flow|prometheus-roobot|cncf-bot|kernelprbot|istio-testing|spinnakerbot|pikbot|spinnaker-release|golangcibot|opencontrail-ci-admin|titanium-octobot|asfgit|appveyorbot|cadvisorjenkinsbot|gitcoinbot|katacontainersbot|prombot|prowbot|zowe-robot|cf-gitbot|pfs-ci-gitbot|ElectronBot|electron-bot)$|((-bot|-robot|-jenkins|-testing|cibot|-ci|-gerrit)$|^(k8s-|bot-|robot-|jenkins-|codecov-)|\[bot\]|\[robot\]|clabot|cla-bot|-bot-|Dead Code Bot|envoy-filter-example)|^travis.*bot$|-ci.*bot$`
	botUsernameR = `^(facebook-github-whois-bot-0|fossabot|claassistant|containersshbuilder|knative-automation|covbot|cdk8s-automation|github-action-benchmark|wasmcloud-automation|goreleaserbot|imgbotapp|backstage-service|openssl-machine|sizebot|dependabot|cncf-ci|poiana|svcbot-qecnsdp|nsmbot|ti-srebot|cf-buildpacks-eng|bosh-ci-push-pull|gprasath|zephyr-github|zephyrbot|strimzi-ci|athenabot|k8s-reviewable|codecov-io|grpc-testing|k8s-teamcity-mesosphere|angular-builds|devstats-sync|googlebot|hibernate-ci|coveralls|rktbot|coreosbot|web-flow|prometheus-roobot|cncf-bot|kernelprbot|istio-testing|spinnakerbot|pikbot|spinnaker-release|golangcibot|opencontrail-ci-admin|titanium-octobot|asfgit|appveyorbot|cadvisorjenkinsbot|gitcoinbot|katacontainersbot|prombot|prowbot|zowe-robot|cf-gitbot|pfs-ci-gitbot|electron-bot)$|((-bot|-robot|-jenkins|-testing|cibot|-ci|-gerrit)$|^(k8s-|bot-|robot-|jenkins-|codecov-)|\[bot\]|\[robot\]|clabot|cla-bot|-bot-|envoy-filter-example)|^travis.*bot$|-ci.*bot$`
	botEmailR    = `^(zowe\.robot\@gmail\.com|github\+dockerlibrarybot\@infosiftr\.com|zziming\-ghbot\@vmware\.com|ci\@argoproj\.com)$|^(ci\@|bot\@|jenkins\-|jenkins\@)|(\-robot\@|\[bot\]\@|\-bot\@|\.ci\.robot\@|\.bot\@|releasebot\@|\-bot\-|\-robot\-|nsmbot\@|kubevirtbot\@|\-automation\@|cibot\@|jenkins\-releng\@)`
)
