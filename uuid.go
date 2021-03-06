package ds

import (
	"fmt"
	"strings"
	"sync"

	"github.com/LF-Engineering/insights-datasource-shared/uuid"
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
