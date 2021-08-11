package ds

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Ctx - environment context packed in structure
type Ctx struct {
	DS            string
	Debug         int        // --debug debug level: 0-no, 1-info, 2-verbose
	Retry         int        // --retry: how many times retry failed operatins, default 5
	ST            bool       // --st true: use single threaded version, false: use multi threaded version, default false
	NCPUs         int        // --ncpus, set to override number of CPUs to run, this overwrites --st, default 0 (which means do not use it, use all CPU reported by go library)
	NCPUsScale    float64    // --ncpus-scale, scale number of CPUs, for example 2.0 will report number of cpus 2.0 the number of actually available CPUs
	Tags          []string   // --tags - tags 'tag1,tag2,...,tagN'
	DryRun        bool       // --dry-run - only output data to console
	Project       string     // --project - set project can be for example "ONAP"
	ProjectFilter bool       // --project-filter - set project filter (normally you only specify project, if you add project-filter flag, DS will try to filter by this project on an actual data source level)
	DateFrom      *time.Time // --date-from (for resuming)
	DateTo        *time.Time // --date-to
}

// Env - get env value using current DS prefix
// Used for extracting data from environment, Ctx.Env must be set first
func (ctx *Ctx) Env(k string) string {
	return os.Getenv(ctx.DS + k)
}

// BoolEnv - parses env variable as bool
// returns false for anything that was parsed as false, zero, empty etc:
// f, F, false, False, fALSe, 0, "", 0.00
// else returns true
func (ctx *Ctx) BoolEnv(k string) bool {
	v := os.Getenv(ctx.DS + k)
	return StringToBool(v)
}

// BoolEnvSet - like BoolEnv but also returns information if variable was set or not
func (ctx *Ctx) BoolEnvSet(k string) (bool, bool) {
	v, present := os.LookupEnv(ctx.DS + k)
	if !present {
		return false, false
	}
	return StringToBool(v), true
}

// EnvSet - is a given environment variable set?
func (ctx *Ctx) EnvSet(k string) bool {
	_, present := os.LookupEnv(ctx.DS + k)
	return present
}

// InitEnv - initialize environment variables parser
func (ctx *Ctx) InitEnv(dsName string) {
	ctx.DS = strings.ToUpper(dsName) + "_"
}

// Init - get context from environment variables
// Configuration can be specified by both cmd line flags and by ENV variables
func (ctx *Ctx) Init() {
	// Flags
	flagDebug := flag.Int("debug", 0, "debug level: 0-no, 1-info, 2-verbose")
	flagRetry := flag.Int("retry", 5, "how many times retry failed operatins, default 5")
	flagST := flag.Bool("st", false, "use single threaded version")
	flagNCPUs := flag.Int("ncpus", 0, "set to override number of CPUs to run, this overwrites --st, default 0 (which means do not use it, use all CPU reported by go library)")
	flagNCPUsScale := flag.Float64("ncpus-scale", 1.0, "scale number of CPUs, for example 2.0 will report number of cpus 2.0 the number of actually available CPUs")
	flagTags := flag.String("tags", "", "'tag1,tag2,...,tagN'")
	flagDryRun := flag.Bool("dry-run", false, "only output data to console")
	flagProject := flag.String("project", "", "set project can be for example \"ONAP\"")
	flagProjectFilter := flag.Bool("project-filter", false, "set project filter (normally you only specify project, if you add project-filter flag, DS will try to filter by this project on an actual data source level)")
	flagDateFrom := flag.String("date-from", "", "date-from (for resuming)")
	flagDateTo := flag.String("date-to", "", "date-to (for limiting)")
	flag.Parse()

	// Debug
	if FlagPassed("debug") && *flagDebug != 0 {
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
	if FlagPassed("retry") && *flagRetry >= 0 {
		ctx.Retry = *flagRetry
	}
	if !FlagPassed("retry") {
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
	if FlagPassed("st") {
		ctx.ST = *flagST
	}
	st, present := ctx.BoolEnvSet("ST")
	if present {
		ctx.ST = st
	}
	// NCPUs
	if FlagPassed("ncpus") && *flagNCPUs >= 0 {
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
	if FlagPassed("ncpus-scale") && *flagNCPUsScale > 0.0 {
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
	if FlagPassed("tags") {
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
	if FlagPassed("dry-run") {
		ctx.DryRun = *flagDryRun
	}
	dryRun, present := ctx.BoolEnvSet("DRY_RUN")
	if present {
		ctx.DryRun = dryRun
	}

	// Project
	if FlagPassed("project") && *flagProject != "" {
		ctx.Project = *flagProject
	}
	if ctx.EnvSet("PROJECT") {
		ctx.Project = ctx.Env("PROJECT")
	}

	// ProjectFilter
	if FlagPassed("project-filter") {
		ctx.ProjectFilter = *flagProjectFilter
	}
	projectFilter, present := ctx.BoolEnvSet("PROJECT_FILTER")
	if present {
		ctx.ProjectFilter = projectFilter
	}

	// Date from/to (optional)
	if FlagPassed("date-from") {
		t, err := TimeParseAny(*flagDateFrom)
		FatalOnError(err)
		ctx.DateFrom = &t
	}
	if FlagPassed("date-to") {
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
