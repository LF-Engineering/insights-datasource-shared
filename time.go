package ds

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DateCacheEntry - parse date cache entry
type DateCacheEntry struct {
	Dt     time.Time
	DtInTz time.Time
	TzOff  float64
	Valid  bool
}

var (
	parseDateCache    = map[string]DateCacheEntry{}
	parseDateCacheMtx *sync.RWMutex
	// DefaultDateFrom - default date from
	DefaultDateFrom = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	// DefaultDateTo - default date to
	DefaultDateTo = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ToYMDHMSDate - return time formatted as YYYY-MM-DD HH:MI:SS
func ToYMDHMSDate(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
}

// TimeParseAny - attempts to parse time from string YYYY-MM-DD HH:MI:SS
// Skipping parts from right until only YYYY id left
func TimeParseAny(dtStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.000000",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15",
		"2006-01-02",
		"2006-01",
		"2006",
	}
	for _, format := range formats {
		t, e := time.Parse(format, dtStr)
		if e == nil {
			return t, e
		}
	}
	e := fmt.Errorf("Error:\nCannot parse date: '%v'", dtStr)
	return time.Now(), e
}

// ParseDateWithTz - try to parse mbox date
func ParseDateWithTz(indt string) (dt, dtInTz time.Time, off float64, valid bool) {
	k := strings.TrimSpace(indt)
	if MT {
		parseDateCacheMtx.RLock()
	}
	entry, ok := parseDateCache[k]
	if MT {
		parseDateCacheMtx.RUnlock()
	}
	if ok {
		dt = entry.Dt
		dtInTz = entry.DtInTz
		off = entry.TzOff
		valid = entry.Valid
		return
	}
	defer func() {
		defer func() {
			entry := DateCacheEntry{Dt: dt, DtInTz: dtInTz, TzOff: off, Valid: valid}
			if MT {
				parseDateCacheMtx.Lock()
			}
			parseDateCache[k] = entry
			if MT {
				parseDateCacheMtx.Unlock()
			}
		}()
		if !valid {
			return
		}
		dtInTz = dt
		ary := strings.Split(indt, "+0")
		if len(ary) > 1 {
			last := ary[len(ary)-1]
			if TZOffsetRE.MatchString(last) {
				digs := TZOffsetRE.ReplaceAllString(last, `$1`)
				offH, _ := strconv.Atoi(digs[:1])
				offM, _ := strconv.Atoi(digs[1:])
				off = float64(offH) + float64(offM)/60.0
				dt = dt.Add(time.Minute * time.Duration(off*-60))
				return
			}
		}
		ary = strings.Split(indt, "+1")
		if len(ary) > 1 {
			last := ary[len(ary)-1]
			if TZOffsetRE.MatchString(last) {
				digs := TZOffsetRE.ReplaceAllString(last, `$1`)
				offH, _ := strconv.Atoi(digs[:1])
				offM, _ := strconv.Atoi(digs[1:])
				off = float64(10+offH) + float64(offM)/60.0
				dt = dt.Add(time.Minute * time.Duration(off*-60))
				return
			}
		}
		ary = strings.Split(indt, "-0")
		if len(ary) > 1 {
			last := ary[len(ary)-1]
			if TZOffsetRE.MatchString(last) {
				digs := TZOffsetRE.ReplaceAllString(last, `$1`)
				offH, _ := strconv.Atoi(digs[:1])
				offM, _ := strconv.Atoi(digs[1:])
				off = -(float64(offH) + float64(offM)/60.0)
				dt = dt.Add(time.Minute * time.Duration(off*-60))
				return
			}
		}
		ary = strings.Split(indt, "-1")
		if len(ary) > 1 {
			last := ary[len(ary)-1]
			if TZOffsetRE.MatchString(last) {
				digs := TZOffsetRE.ReplaceAllString(last, `$1`)
				offH, _ := strconv.Atoi(digs[:1])
				offM, _ := strconv.Atoi(digs[1:])
				off = -(float64(10+offH) + float64(offM)/60.0)
				dt = dt.Add(time.Minute * time.Duration(off*-60))
				return
			}
		}
	}()
	sdt := indt
	// https://www.broobles.com/eml2mbox/mbox.html
	// but the real world is not that simple
	for _, r := range []string{">", ",", ")", "("} {
		sdt = strings.Replace(sdt, r, "", -1)
	}
	for _, split := range []string{"+0", "+1", "."} {
		ary := strings.Split(sdt, split)
		sdt = ary[0]
	}
	for _, split := range []string{"-0", "-1"} {
		ary := strings.Split(sdt, split)
		lAry := len(ary)
		if lAry > 1 {
			_, err := strconv.Atoi(ary[lAry-1])
			if err == nil {
				sdt = strings.Join(ary[:lAry-1], split)
			}
		}
	}
	sdt = SpacesRE.ReplaceAllString(sdt, " ")
	sdt = strings.ToLower(strings.TrimSpace(sdt))
	ary := strings.Split(sdt, " ")
	day := ary[0]
	if len(day) > 3 {
		day = day[:3]
	}
	_, ok = LowerDayNames[day]
	if ok {
		sdt = strings.Join(ary[1:], " ")
	}
	sdt = strings.TrimSpace(sdt)
	for lm, m := range LowerFullMonthNames {
		sdt = strings.Replace(sdt, lm, m, -1)
	}
	for lm, m := range LowerMonthNames {
		sdt = strings.Replace(sdt, lm, m, -1)
	}
	ary = strings.Split(sdt, " ")
	if len(ary) > 4 {
		sdt = strings.Join(ary[:4], " ")
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02t15:04:05",
		"2006-01-02 15:04:05z",
		"2006-01-02t15:04:05z",
		"2 Jan 2006 15:04:05",
		"02 Jan 2006 15:04:05",
		"2 Jan 06 15:04:05",
		"02 Jan 06 15:04:05",
		"2 Jan 2006 15:04",
		"02 Jan 2006 15:04",
		"2 Jan 06 15:04",
		"02 Jan 06 15:04",
		"Jan 2 15:04:05 2006",
		"Jan 02 15:04:05 2006",
		"Jan 2 15:04:05 06",
		"Jan 02 15:04:05 06",
		"Jan 2 15:04 2006",
		"Jan 02 15:04 2006",
		"Jan 2 15:04 06",
		"Jan 02 15:04 06",
	}
	var (
		err  error
		errs []error
	)
	for _, format := range formats {
		dt, err = time.Parse(format, sdt)
		if err == nil {
			// Printf("Parsed %v\n", dt)
			valid = true
			return
		}
		errs = append(errs, err)
	}
	Printf("ParseDateWithTz: errors: %+v\n", errs)
	Printf("ParseDateWithTz: '%s' -> '%s', day: %s\n", indt, sdt, day)
	return
}

// PeriodParse - tries to parse period
func PeriodParse(perStr string) (dur time.Duration, ok bool) {
	idx := strings.Index(perStr, "[rate reset in ")
	if idx == -1 {
		return
	}
	rateStr := ""
	_, err := fmt.Sscanf(perStr[idx:], "[rate reset in %s", &rateStr)
	if err != nil || len(rateStr) < 2 {
		return
	}
	rateStr = rateStr[0 : len(rateStr)-1]
	if rateStr == "" {
		return
	}
	d, err := time.ParseDuration(rateStr)
	if err != nil {
		return
	}
	dur = d
	ok = true
	return
}

// ToYMDHMDate - return time formatted as YYYY-MM-DD HH:MI
func ToYMDHMDate(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute())
}

// ToESDate - return time formatted as YYYY-MM-DDTHH:MI:SS.uuuuuu+00:00
func ToESDate(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%06.0f+00:00", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), float64(dt.Nanosecond())/1.0e3)
}

// TimeParseES - parse datetime in ElasticSearch output format
func TimeParseES(dtStr string) (time.Time, error) {
	dtStr = strings.TrimSpace(strings.Replace(dtStr, "Z", "", -1))
	ary := strings.Split(dtStr, "+")
	ary2 := strings.Split(ary[0], ".")
	var s string
	if len(ary2) == 1 {
		s = ary2[0] + ".000"
	} else {
		if len(ary2[1]) > 3 {
			ary2[1] = ary2[1][:3]
		}
		s = strings.Join(ary2, ".")
	}
	return time.Parse("2006-01-02T15:04:05.000", s)
}

// TimeParseInterfaceString - parse interface{} -> string -> time.Time
func TimeParseInterfaceString(date interface{}) (dt time.Time, err error) {
	sDate, ok := date.(string)
	if !ok {
		err = fmt.Errorf("%+v %T is not a string", date, date)
		return
	}
	dt, err = TimeParseES(sDate)
	return
}

// ToYMDTHMSZDate - return time formatted as YYYY-MM-DDTHH:MI:SSZ
func ToYMDTHMSZDate(dt time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
}

// ConvertTimeToFloat ...
func ConvertTimeToFloat(t time.Time) float64 {
	return math.Round(float64(t.UnixNano())/float64(time.Second)*1e6) / 1e6
}

// GetDaysBetweenDates calculate days between two dates
func GetDaysBetweenDates(t1 time.Time, t2 time.Time) float64 {
	res := t1.Sub(t2).Hours() / 24
	return res
}

// GetOldestDate get the older date between two nullable dates
func GetOldestDate(t1 *time.Time, t2 *time.Time) *time.Time {
	from, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	if err != nil {
		return nil
	}

	isT1Empty := t1 == nil || t1.IsZero()
	isT2Empty := t2 == nil || t2.IsZero()

	if isT1Empty && !isT2Empty {
		from = *t2
	} else if !isT1Empty && isT2Empty {
		from = *t1
	} else if !isT1Empty && !isT2Empty {
		from = *t2
		if t1.Before(*t2) {
			from = *t1
		}
	}

	return &from
}
