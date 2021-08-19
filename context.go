package ds

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultPackSize - default pack size for events pack produced by data sources
	DefaultPackSize = 1000
)

// Ctx - environment context packed in structure
// It gets configuration (named, say: xyz abc) from command line (--dsname-xyz-abc) or from env (DSNAME_XYZ_ABC), env value has higher priority than commandline flag
type Ctx struct {
	DS            string              // original data source name
	DSEnv         string              // prefix for env variables: "abc xyz" -> "ABC_XYZ_"
	DSFlag        string              // prefix for commanding flags: "abc xyz" -> "--abc-xyz"
	Debug         int                 // debug level: 0-no, 1-info, 2-verbose
	Retry         int                 // how many times retry failed operatins, default 5
	ST            bool                // use single threaded version, false: use multi threaded version, default false
	NCPUs         int                 // set to override number of CPUs to run, this overwrites --st, default 0 (which means do not use it, use all CPU reported by go library)
	NCPUsScale    float64             // scale number of CPUs, for example 2.0 will report number of cpus 2.0 the number of actually available CPUs
	Tags          []string            // tags 'tag1,tag2,...,tagN'
	DryRun        bool                // only output data to console
	Project       string              // set project can be for example "ONAP"
	ProjectFilter bool                // set project filter (normally you only specify project, if you add project-filter flag, DS will try to filter by this project on an actual data source level)
	PackSize      int                 // data sources are outputting events in packs - here you can specify pack size, default is 1000
	ESURL         string              // set ES cluster URL (optional but rather recommended)
	NoCache       bool                // do not cache *any* HTTP requests
	Categories    map[string]struct{} // some data sources allow specifying categories, you can pass them with --dsname-categories 'category1,category2,...' flag, it will keep unique set of them.
	DateFrom      *time.Time          // date from (for resuming)
	DateTo        *time.Time          // date to (for limiting)
}

// Env - get env value using current DS prefix
// Used for extracting data from environment, Ctx.Env must be set first
func (ctx *Ctx) Env(k string) string {
	return os.Getenv(ctx.DSEnv + k)
}

// BoolEnv - parses env variable as bool
// returns false for anything that was parsed as false, zero, empty etc:
// f, F, false, False, fALSe, 0, "", 0.00
// else returns true
func (ctx *Ctx) BoolEnv(k string) bool {
	v := os.Getenv(ctx.DSEnv + k)
	return StringToBool(v)
}

// BoolEnvSet - like BoolEnv but also returns information if variable was set or not
func (ctx *Ctx) BoolEnvSet(k string) (bool, bool) {
	v, present := os.LookupEnv(ctx.DSEnv + k)
	if !present {
		return false, false
	}
	return StringToBool(v), true
}

// EnvSet - is a given environment variable set?
func (ctx *Ctx) EnvSet(k string) bool {
	_, present := os.LookupEnv(ctx.DSEnv + k)
	return present
}

// InitEnv - initialize environment variables parser
func (ctx *Ctx) InitEnv(dsName string) {
	ctx.DS = dsName
	dsName = strings.Replace(dsName, ".", "", -1)
	ctx.DSEnv = strings.Replace(strings.ToUpper(dsName), " ", "_", -1) + "_"
	ctx.DSFlag = strings.Replace(strings.ToLower(dsName), " ", "-", -1) + "-"
}

// Init - get context from environment variables
// Configuration can be specified by both cmd line flags and by ENV variables
func (ctx *Ctx) Init() {
	// Flags
	flagDebug := flag.Int(ctx.DSFlag+"debug", 0, "debug level: 0-no, 1-info, 2-verbose")
	flagRetry := flag.Int(ctx.DSFlag+"retry", 5, "how many times retry failed operatins, default 5")
	flagST := flag.Bool(ctx.DSFlag+"st", false, "use single threaded version")
	flagNCPUs := flag.Int(ctx.DSFlag+"ncpus", 0, "set to override number of CPUs to run, this overwrites --st, default 0 (which means do not use it, use all CPU reported by go library)")
	flagNCPUsScale := flag.Float64(ctx.DSFlag+"ncpus-scale", 1.0, "scale number of CPUs, for example 2.0 will report number of cpus 2.0 the number of actually available CPUs")
	flagTags := flag.String(ctx.DSFlag+"tags", "", "'tag1,tag2,...,tagN'")
	flagDryRun := flag.Bool(ctx.DSFlag+"dry-run", false, "only output data to console")
	flagProject := flag.String(ctx.DSFlag+"project", "", "set project can be for example \"ONAP\"")
	flagProjectFilter := flag.Bool(ctx.DSFlag+"project-filter", false, "set project filter (normally you only specify project, if you add project-filter flag, DS will try to filter by this project on an actual data source level)")
	flagPackSize := flag.Int(ctx.DSFlag+"pack-size", 1000, "data sources are outputting events in packs - here you can specify pack size, default is 1000")
	flagESURL := flag.String(ctx.DSFlag+"es-url", "", "ElasticSearch URL (optional but recommended)")
	flagNoCache := flag.Bool(ctx.DSFlag+"no-cache", false, "do *NOT* cache any HTTP requests")
	flagDateFrom := flag.String(ctx.DSFlag+"date-from", "", "date-from (for resuming)")
	flagDateTo := flag.String(ctx.DSFlag+"date-to", "", "date-to (for limiting)")
	flagCategories := flag.String(ctx.DSFlag+"categories", "", "some data sources allow specifying categories, you can pass them with --dsname-categories 'category1,category2,...' flag, it will keep unique set of them.")
	flag.Parse()

	// Debug
	if FlagPassed(ctx, "debug") && *flagDebug != 0 {
		ctx.Debug = *flagDebug
	}
	if ctx.EnvSet("DEBUG") {
		debug, err := strconv.Atoi(ctx.Env("DEBUG"))
		FatalOnError(err)
		if debug != 0 {
			ctx.Debug = debug
		}
	}

	// Retry
	if FlagPassed(ctx, "retry") && *flagRetry >= 0 {
		ctx.Retry = *flagRetry
	}
	if !FlagPassed(ctx, "retry") {
		if !ctx.EnvSet("RETRY") {
			ctx.Retry = 5
		} else {
			retry, err := strconv.Atoi(ctx.Env("RETRY"))
			FatalOnError(err)
			ctx.Retry = retry
		}
	} else {
		retry, err := strconv.Atoi(ctx.Env("RETRY"))
		if err == nil {
			ctx.Retry = retry
		}
	}

	// Threading
	if FlagPassed(ctx, "st") {
		ctx.ST = *flagST
	}
	st, present := ctx.BoolEnvSet("ST")
	if present {
		ctx.ST = st
	}
	// NCPUs
	if FlagPassed(ctx, "ncpus") && *flagNCPUs >= 0 {
		ctx.NCPUs = *flagNCPUs
	}
	if ctx.EnvSet("NCPUS") {
		nCPUs, err := strconv.Atoi(ctx.Env("NCPUS"))
		FatalOnError(err)
		if nCPUs > 0 {
			ctx.NCPUs = nCPUs
			if ctx.NCPUs == 1 {
				ctx.ST = true
			}
		}
	}
	if ctx.NCPUs == 1 {
		ctx.ST = true
	}

	// NCPUs scale
	ctx.NCPUsScale = 1.0
	if FlagPassed(ctx, "ncpus-scale") && *flagNCPUsScale > 0.0 {
		ctx.NCPUsScale = *flagNCPUsScale
	}
	if ctx.EnvSet("NCPUS_SCALE") {
		nCPUsScale, err := strconv.ParseFloat(ctx.Env("NCPUS_SCALE"), 64)
		FatalOnError(err)
		if nCPUsScale > 0.0 {
			ctx.NCPUsScale = nCPUsScale
		}
	}

	// Tags
	tags := map[string]interface{}{}
	if FlagPassed(ctx, "tags") {
		ary := strings.Split(*flagTags, ",")
		for _, tag := range ary {
			tag := strings.TrimSpace(tag)
			if tag != "" {
				tags[tag] = struct{}{}
			}
		}
	}
	ary := strings.Split(ctx.Env("TAGS"), ",")
	for _, tag := range ary {
		tag := strings.TrimSpace(tag)
		if tag != "" {
			tags[tag] = struct{}{}
		}
	}
	for tag := range tags {
		ctx.Tags = append(ctx.Tags, tag)
	}

	// Dry run
	if FlagPassed(ctx, "dry-run") {
		ctx.DryRun = *flagDryRun
	}
	dryRun, present := ctx.BoolEnvSet("DRY_RUN")
	if present {
		ctx.DryRun = dryRun
	}

	// Project
	if FlagPassed(ctx, "project") && *flagProject != "" {
		ctx.Project = *flagProject
	}
	if ctx.EnvSet("PROJECT") {
		ctx.Project = ctx.Env("PROJECT")
	}

	// ProjectFilter
	if FlagPassed(ctx, "project-filter") {
		ctx.ProjectFilter = *flagProjectFilter
	}
	projectFilter, present := ctx.BoolEnvSet("PROJECT_FILTER")
	if present {
		ctx.ProjectFilter = projectFilter
	}

	// Categories
	cats := ""
	if FlagPassed(ctx, "categories") && *flagCategories != "" {
		cats = *flagCategories
	}
	if ctx.EnvSet("CATEGORIES") {
		cats = ctx.Env("CATEGORIES")
	}
	catsAry := []string{}
	ary = strings.Split(cats, ",")
	for _, cat := range ary {
		cat := strings.TrimSpace(cat)
		if cat != "" {
			catsAry = append(catsAry, cat)
		}
	}
	ctx.Categories = make(map[string]struct{})
	for _, cat := range catsAry {
		ctx.Categories[cat] = struct{}{}
	}

	// ES URL
	if FlagPassed(ctx, "es-url") && *flagESURL != "" {
		ctx.ESURL = *flagESURL
	}
	if ctx.EnvSet("ES_URL") {
		ctx.ESURL = ctx.Env("ES_URL")
	}
	if ctx.ESURL != "" {
		AddRedacted(ctx.ESURL, false)
	}

	// No cache
	if FlagPassed(ctx, "no-cache") {
		ctx.NoCache = *flagNoCache
	}
	noCache, present := ctx.BoolEnvSet("NO_CACHE")
	if present {
		ctx.NoCache = noCache
	}

	// Events pack size
	ctx.PackSize = DefaultPackSize
	if FlagPassed(ctx, "pack-size") && *flagPackSize > 0 {
		ctx.PackSize = *flagPackSize
	}
	if ctx.EnvSet("PACK_SIZE") {
		packSize, err := strconv.Atoi(ctx.Env("PACK_SIZE"))
		FatalOnError(err)
		if packSize > 0 {
			ctx.PackSize = packSize
		}
	}

	// Date from/to (optional)
	if FlagPassed(ctx, "date-from") {
		t, err := TimeParseAny(*flagDateFrom)
		FatalOnError(err)
		ctx.DateFrom = &t
	}
	if FlagPassed(ctx, "date-to") {
		t, err := TimeParseAny(*flagDateTo)
		FatalOnError(err)
		ctx.DateTo = &t
	}
	if ctx.EnvSet("DATE_FROM") {
		t, err := TimeParseAny(ctx.Env("DATE_FROM"))
		FatalOnError(err)
		ctx.DateFrom = &t
	}
	if ctx.EnvSet("DATE_TO") {
		t, err := TimeParseAny(ctx.Env("DATE_TO"))
		FatalOnError(err)
		ctx.DateTo = &t
	}
}

// Print context contents
func (ctx *Ctx) Print() {
	fmt.Printf("Environment Context Dump\n%+v\n", ctx)
}

// Info - return context in human readable form
func (ctx Ctx) Info() string {
	return fmt.Sprintf("%+v", ctx)
}
