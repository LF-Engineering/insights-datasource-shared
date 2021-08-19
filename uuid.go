package ds

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/LF-Engineering/dev-analytics-libraries/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// uuidsNonEmptyCache caches UUIDNonEmpty calls
	uuidsNonEmptyCache    = map[string]string{}
	uuidsNonEmptyCacheMtx *sync.RWMutex
	// uuidsAffsCache caches UUIDAffs calls
	uuidsAffsCache    = map[string]string{}
	uuidsAffsCacheMtx *sync.RWMutex
)

// ResetUUIDCache - resets cache
func ResetUUIDCache() {
	uuidsNonEmptyCache = map[string]string{}
	uuidsAffsCache = map[string]string{}
}

// UUIDNonEmpty - generate UUID of string args (all must be non-empty)
// uses internal cache
// used to generate document UUID's
func UUIDNonEmpty(ctx *Ctx, args ...string) (h string) {
	k := strings.Join(args, ":")
	if MT {
		uuidsNonEmptyCacheMtx.RLock()
	}
	h, ok := uuidsNonEmptyCache[k]
	if MT {
		uuidsNonEmptyCacheMtx.RUnlock()
	}
	if ok {
		return
	}
	if ctx.Debug > 1 {
		defer func() {
			Printf("UUIDNonEmpty(%v) --> %s\n", args, h)
		}()
	}
	defer func() {
		if MT {
			uuidsNonEmptyCacheMtx.Lock()
		}
		uuidsNonEmptyCache[k] = h
		if MT {
			uuidsNonEmptyCacheMtx.Unlock()
		}
	}()
	var err error
	h, err = uuid.Generate(args...)
	if err != nil {
		Printf("UUIDNonEmpty error for: %+v\n", args)
		h = ""
	}
	return
}

// UUIDAffs - generate UUID of string args
// uses internal cache
// downcases arguments, all but first can be empty
func UUIDAffs(ctx *Ctx, args ...string) (h string) {
	k := strings.Join(args, ":")
	if MT {
		uuidsAffsCacheMtx.RLock()
	}
	h, ok := uuidsAffsCache[k]
	if MT {
		uuidsAffsCacheMtx.RUnlock()
	}
	if ok {
		return
	}
	if ctx.Debug > 1 {
		defer func() {
			Printf("UUIDAffs(%v) --> %s\n", args, h)
		}()
	}
	defer func() {
		if MT {
			uuidsAffsCacheMtx.Lock()
		}
		uuidsAffsCache[k] = h
		if MT {
			uuidsAffsCacheMtx.Unlock()
		}
	}()
	var err error
	if len(args) != 4 {
		err = fmt.Errorf("GenerateIdentity requires exactly 4 asrguments, got %+v", args)
	} else {
		h, err = uuid.GenerateIdentity(&args[0], &args[1], &args[2], &args[3])
	}
	if err != nil {
		Printf("UUIDAffs error for: %+v\n", args)
		h = ""
	}
	return
}

// ToUnicode converts string to unicode
func ToUnicode(s string) (string, error) {
	dst := make([]byte, len(s)+100)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	nDst, _, err := t.Transform(dst, []byte(s), true)
	if err != nil {
		return "", err
	}
	return string(dst[:nDst]), nil
}

// Generate uuid using sha1
func Generate(args ...string) (string, error) {
	for i := range args {
		// strip spaces
		args[i] = strings.TrimSpace(args[i])

		// check empty args
		if args[i] == "" {
			return "", errors.New("args cannot be empty")
		}
	}

	data := strings.Join(args, ":")

	hash := sha1.New()
	_, err := hash.Write([]byte(data))
	if err != nil {
		return "", err
	}
	hashed := fmt.Sprintf("%x", hash.Sum(nil))

	return hashed, nil

}

// GenerateIdentity generates uuid related to user ex. userUUID
func GenerateIdentity(source, email, name, username *string) (string, error) {

	if source == nil || *source == "" {
		return "", errors.New("source cannot be an empty string")
	}

	if (email == nil || *email == "") && (name == nil || *name == "") && (username == nil || *username == "") {
		return "", errors.New("identity data cannot be None or empty")
	}

	args := make([]string, 4)
	args[0] = *source

	if email == nil || *email == "" {
		args[1] = "none"
	} else {
		args[1] = *email
	}

	if name == nil || *name == "" {
		args[2] = "none"
	} else {
		args[2] = *name
	}

	if username == nil || *username == "" {
		args[3] = "none"
	} else {
		args[3] = *username
	}

	for i := range args {

		output := ""
		ss := args[i]
		for len(ss) > 0 {
			r, size := utf8.DecodeRuneInString(ss)
			if unicode.IsSymbol(r) {
				output += string(rune(ss[0]))
			} else {
				output += string(r)
			}
			ss = ss[size:]
		}
		args[i] = output

		// strip spaces
		args[i] = strings.TrimSpace(args[i])
	}

	data := strings.Join(args, ":")

	// to unicode
	output, err := ToUnicode(data)
	if err != nil {
		return "", err
	}
	data = strings.ToLower(output)

	hash := sha1.New()
	_, err = hash.Write([]byte(data))
	if err != nil {
		return "", err
	}
	hashed := fmt.Sprintf("%x", hash.Sum(nil))

	return hashed, nil

}
