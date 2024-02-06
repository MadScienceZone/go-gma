/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __     ______     _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \   / ____ \   (  __   )       #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) ) ( (    \/   | (  )  |       #
# | |      | || || || (___) | Assistant | (____        | | | (____     | | /   |       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | | |  ___ \    | (/ /) |       #
# | | \_  )| |   | || (   ) |                 ) )      | | | (   ) )   |   / | |       #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_( (___) ) _ |  (__) |       #
# (_______)|/     \||/     \| Client    \______/ (_) \____/ \_____/ (_)(_______)       #
#                                                                                      #
########################################################################################
*/

//
// Package gma is the main port of the GMA Core API into the Go language.
// Parts of the API which don't necessarily belong in their own individual packages (e.g., dice, mapper, etc.) will go here.
// Currently, the game calendar is implemented here.
//
// # Game Calendar
//
// Call NewCalendar(calSystem) to create a new game calendar which is set up for the particular calendaring system applicable for your world.
// Then you can set the time and date, advance the time as needed, etc. by calling the various methods described below.
//
package gma

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/MadScienceZone/go-gma/v5/dice"
)

//
// MonthInfo describes a month in a given calendar system.
//
type MonthInfo struct {
	// Month's full and abbreviated names
	FullName, Abbrev string

	// How many days are in the month, in regular and leap years
	Days, LYDays int
}

//
// Calendar describes the calendar system in play and the
// current date and time.
//
// Since a given game calendar will be in play for the entire
// GMA system at any given point in time, in terms of type design
// it's more important that Calendar be the data type rather than
// the calendar type in play. Thus, we specify the calendar system
// when constructing a new Calendar with NewCalendar, which sets up
// all the parameters for managing a calendar of that type.
//
type Calendar struct {
	// The name of the calendering system in use.
	System string

	// Calendar-specific names and other details about the months,
	// seasons and weekday names.
	Months         []MonthInfo
	Seasons        []string
	Days           []string
	MoonPhaseNames []string

	// The current time, broken out into parts
	Year, Month, Date, Hour, Minute, Second, Tick int
	FHour, FMinute, FSecond                       float64

	// The day of the week (0-origin)
	Dow int

	// The phase of the moon index (0-origin)
	Pom int

	// The season index (0-origin)
	Season int

	// Time units in terms of ticks
	SecondUnits, RoundUnits, MinuteUnits, HourUnits int
	DayUnits, WeekUnits, PomUnits, SeasonUnits      int

	// Unit ranges
	TickMod, SecondMod, MinuteMod, HourMod int
	DowMod, PomMod, SeasonMod              int

	// Unit definitions
	UnitMultiplier map[string]int64

	// Epoch basis
	EpochYear, EpochDow, EpochPom, EpochSeason int

	// Current time
	Now int64

	// Internal functions
	j   func(int, int, int) int64
	ymd func(int64) (int, int, int)
	r   func(*Calendar)
	ly  func(int) bool
}

type CalendarOption func(*Calendar)

//
// WithTime specifies an option to the NewCalendar constructor
// to set the current time (i.e. the value of the Now attribute)
// for the newly-created Calendar object.
//
func WithTime(timeValue int64) CalendarOption {
	return func(c *Calendar) {
		c.Now = timeValue
	}
}

//
// SetTimeValue sets the underlying time value (the Now attribute)
// of the Calendar receiver.
//
func (c *Calendar) SetTimeValue(n int64) {
	c.Now = n
	c.recalc()
}

//
// Advance adds delta to the current time tracked by the Calendar
// receiver. If unitName is empty, the default is to interpret delta
// in units of ticks.
//
// Returns the interval in ticks that the clock advanced.
//
func (c *Calendar) Advance(delta int64, unitName string) (int64, error) {
	previous := c.Now
	if unitName == "" {
		c.Now += delta
	} else {
		m, ok := c.UnitMultiplier[unitName]
		if !ok {
			return 0, fmt.Errorf("invalid time unit \"%s\" specified", unitName)
		}
		c.Now += delta * m
	}
	c.recalc()
	return c.Now - previous, nil

}

//
// Delta returns the difference between the receiver and the given Calendar
// value, in ticks.
//
func (c Calendar) Delta(other Calendar) int64 {
	return c.Now - other.Now
}

//
// AdvanceToNext increments the clock to the next even interval unit.
// Defaults to ticks if unitName is empty.
//
func (c *Calendar) AdvanceToNext(unitName string) error {
	t, err := c.TicksToInterval(unitName)
	if err != nil {
		return err
	}
	c.Advance(t, "")
	return nil
}

//
// Image generates and returns an image of the current calendar
// year, suitable for printing out in traditional calendar layout.
// The output is a slice of months, each of which is a slice of weeks,
// with zeroes or the date in each element of the week
// sub-slices.
//
func (c Calendar) Image() ([][][]int, error) {
	basis, err := NewCalendar(c.System)
	if err != nil {
		return nil, err
	}
	if err := basis.SetTime(c.Year, 1, 1, 0, 0, 0, 0); err != nil {
		return nil, err
	}
	startDow := basis.Dow
	lastDow := 0
	var cal [][][]int
	for _, m := range basis.Months {
		currentDate := 1
		var month [][]int
		daysLeft := m.Days
		if basis.IsLeapYear() {
			daysLeft = m.LYDays
		}
		for daysLeft > 0 {
			var week []int
			for dow, _ := range basis.Days {
				if startDow > dow || daysLeft <= 0 {
					week = append(week, 0)
				} else {
					startDow = 0
					week = append(week, currentDate)
					currentDate++
					daysLeft--
					lastDow = dow
				}
			}
			month = append(month, week)
		}
		cal = append(cal, month)
		startDow = (lastDow + 1) % len(basis.Days)
	}
	return cal, nil
}

//
// DeltaString returns a formatted string showing a time delta.
// If strict is true, then the output is strictly conforming to
// what ScanInterval would accept; otherwise, a more general
// format is used which is more human-friendly.
//
func (c Calendar) DeltaString(delta int64, strict bool) string {
	var sign = "+"
	var mult = 1

	if delta < 0 {
		sign = "-"
		mult = -1
		delta = -delta
	}

	var res strings.Builder
	fmt.Fprint(&res, sign)
	if strict {
		if delta == 0 {
			return "nil"
		}
		for _, data := range []struct {
			u      int64
			lim    int64
			uname  string
			plural string
		}{
			{int64(c.DayUnits), 0, "day", "days"},
			{int64(c.HourUnits), int64(c.DayUnits), "hour", "hours"},
			{int64(c.MinuteUnits), int64(c.HourUnits), "minute", "minutes"},
			{int64(c.RoundUnits), 10 * int64(c.MinuteUnits), "round", "rounds"},
			{int64(c.SecondUnits), int64(c.MinuteUnits), "second", "seconds"},
			{1, 10, "", ""},
		} {
			q := pdiv(delta, data.u)
			if imod64(delta, data.u) == 0 && (data.lim == 0 || delta < data.lim) {
				if q == 1 && sign == "+" {
					return data.uname
				}
				if q == 1 {
					return fmt.Sprintf("%v %s", int64(mult)*q, data.uname)
				} else {
					return fmt.Sprintf("%v %s", int64(mult)*q, data.plural)
				}
			}
		}

		if delta >= int64(c.DayUnits) {
			fmt.Fprintf(&res, "%v:", pdiv(delta, int64(c.DayUnits)))
		}
		if delta >= int64(c.HourUnits) {
			fmt.Fprintf(&res, "%02d:", imod64(int64(pdiv(delta, int64(c.HourUnits))), int64(c.HourMod)))
		}
		fmt.Fprintf(&res, "%02d:%02d",
			imod64(int64(pdiv(delta, int64(c.MinuteUnits))), int64(c.MinuteMod)),
			imod64(int64(pdiv(delta, int64(c.SecondUnits))), int64(c.SecondMod)),
		)

		if imod64(delta, int64(c.TickMod)) != 0 {
			fmt.Fprintf(&res, ".%d", imod64(delta, int64(c.TickMod)))
		}
		return res.String()
	}

	dy := pdiv(delta, int64(c.DayUnits))
	hr := imod64(int64(pdiv(delta, int64(c.HourUnits))), int64(c.HourMod))

	if dy != 0 {
		fmt.Fprintf(&res, "%dd", dy)
	}
	if dy != 0 || hr != 0 {
		fmt.Fprintf(&res, "%02d:", hr)
	}
	fmt.Fprintf(&res, "%02d:%02d.%d %4.1fr",
		imod64(int64(pdiv(delta, int64(c.MinuteUnits))), int64(c.MinuteMod)),
		imod64(int64(pdiv(delta, int64(c.SecondUnits))), int64(c.SecondMod)),
		imod64(delta, int64(c.TickMod)),
		float64(delta)/float64(c.RoundUnits),
	)
	return res.String()
}

//
// ScanInterval reads the provided relative time in one of the
// following formats:
//   nil
//   <n> [<unit>]
//   <unit>
//   [<d>:[<hh>:]]<mm>:<ss>[.<t>]
// Returns the number of ticks represented by that interval of time.
// Square brackets above indicate optional components.
//
// If literal square brackets are input (e.g., "[2d6+1] rounds"), then
// the text inside the brackets is processed as a random die roll, the
// result of which will be used for the time interval.
//
func (c *Calendar) ScanInterval(intSpec string) (int64, error) {
	if strings.TrimSpace(intSpec) == "nil" {
		return 0, nil
	}
	//
	// Replace [...] with results of dice rolls
	//
	roll := regexp.MustCompile(`\[(.*?)]`)
	intSpec = roll.ReplaceAllStringFunc(intSpec, func(dspec string) string {
		d, err := dice.New(dice.ByDescription(dspec))
		if err != nil {
			return fmt.Sprintf("\u00ab%s: %v\u00bb", dspec, err)
		}
		v, err := d.Roll()
		if err != nil {
			return fmt.Sprintf("\u00ab%s: %v\u00bb", dspec, err)
		}
		return fmt.Sprintf("%v", v)
	})
	//
	// match against our main pattern         .___2___________.
	//                                        |.__3___.       |            .___8___.
	//                              __1___    ||__4__ | __5__ | __6__ __7__|  __9__|
	spec := regexp.MustCompile(`^\s*([+-])?\s*(((\d+):)?(\d+):)?(\d+):(\d+)(\.(\d+))?\s*$`)
	//                              sign        days    hours   mins   sec     tick
	bits := spec.FindStringSubmatch(intSpec)
	if bits != nil {
		sign := 1
		if bits[1] == "-" {
			sign = -1
		}
		d, err := numberOrDef(bits[4], 0, 0)
		if err != nil {
			return 0, err
		}
		h, err := numberOrDef(bits[5], 0, 0)
		if err != nil {
			return 0, err
		}
		m, err := numberOrDef(bits[6], 0, 0)
		if err != nil {
			return 0, err
		}
		s, err := numberOrDef(bits[7], 0, 0)
		if err != nil {
			return 0, err
		}
		t, err := numberOrDef(bits[9], 0, 0)
		if err != nil {
			return 0, err
		}
		return int64(sign)*int64(d)*int64(c.DayUnits) +
			int64(h)*int64(c.HourUnits) +
			int64(m)*int64(c.MinuteUnits) +
			int64(s)*int64(c.SecondUnits) +
			int64(t), nil
	}
	//
	// match number of units
	//                                  .______1_________.
	//                                  |        __2____ |   __3__
	tickSpec := regexp.MustCompile(`^\s*([+-]?\d+(\.\d+)?)\s*(\w+)\s*$`)
	//                                        value           units
	var delta float64
	var units string
	var err error
	bits = tickSpec.FindStringSubmatch(intSpec)
	if bits != nil {
		delta, err = strconv.ParseFloat(bits[1], 64)
		if err != nil {
			return 0, err
		}
		units = bits[3]
	} else {
		//
		// No? then try just a unit-less number
		//
		units = ""
		delta, err = strconv.ParseFloat(intSpec, 64)
		if err != nil {
			//
			// Still not? maybe it's just a magnitude-less unit
			//
			units = intSpec
		}
	}

	if units == "" {
		return int64(delta), nil
	}

	mult, ok := c.UnitMultiplier[units]
	if !ok {
		return 0, fmt.Errorf("invalid unit name \"%s\"", units)
	}
	return int64(delta * float64(mult)), nil
}

//
// Scan reads the provided string of the format
//   [[[yyyy-]mmm-]dd] hh:mm[:ss[.t]]
// where mmm may be numeric or the month's name.
//
// Optional values are defaulted from the current date and time
// in the receiver.
//
func (c *Calendar) Scan(timeString string) (int64, error) {
	return c.ScanRelative(timeString, *c)
}

//
// ScanRelative is like Scan but rather than pulling defaults
// from the receiver it uses the provided Calendar value.
//
func (c *Calendar) ScanRelative(timeString string, relativeTo Calendar) (int64, error) {
	//                           ._______________1__________.
	//                           |._______2_______.         |            ._______9_______.
	//                           ||.__3___.       |         |            |      .__11__. |
	//                           |||__4__ | __5__ | __6__   | __7__ __8__| _10_ |  _12_| |
	ts := regexp.MustCompile(`\s*((((\d+)-)?(\w+)-)?(\d+)\s+)?(\d+):(\d+)(:(\d+)(\.(\d))?)?\s*$`)
	bits := ts.FindStringSubmatch(timeString)
	if bits == nil {
		return 0, fmt.Errorf("invalid time expression \"%s\". Should be in format \"[[[yyyy-]mmm-]dd] HH:MM[:SS[.T]]\"", timeString)
	}

	monthname := strings.ToUpper(bits[5])
	mon, err := numberOrDef(monthname, relativeTo.Month, len(c.Months))
	if err != nil {
		// not a number, so let's see if it's the name of a month
		mon = 0
		for m, mi := range c.Months {
			if strings.ToUpper(mi.FullName) == monthname || strings.ToUpper(mi.Abbrev) == monthname {
				mon = m + 1
				break
			}
		}
		if mon == 0 {
			return 0, err
		}
	}

	year, err := numberOrDef(bits[4], relativeTo.Year, 0)
	if err != nil {
		return 0, err
	}

	var date int
	if c.ly(year) {
		date, err = numberOrDef(bits[6], relativeTo.Date, c.Months[mon-1].LYDays)
	} else {
		date, err = numberOrDef(bits[6], relativeTo.Date, c.Months[mon-1].Days)
	}
	if err != nil {
		return 0, err
	}

	hour, err := numberOrDef(bits[7], relativeTo.Hour, -23)
	if err != nil {
		return 0, err
	}

	minute, err := numberOrDef(bits[8], relativeTo.Minute, -59)
	if err != nil {
		return 0, err
	}

	second, err := numberOrDef(bits[10], relativeTo.Second, -59)
	if err != nil {
		return 0, err
	}

	tick, err := numberOrDef(bits[12], relativeTo.Tick, -9)
	if err != nil {
		return 0, err
	}

	return relativeTo.j(year, mon, date)*int64(relativeTo.DayUnits) +
		int64(hour)*int64(relativeTo.HourUnits) +
		int64(minute)*int64(relativeTo.MinuteUnits) +
		int64(second)*int64(relativeTo.SecondUnits) + int64(tick), nil
}

func numberOrDef(value string, defVal int, maxVal int) (int, error) {
	if value == "" {
		return defVal, nil
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if maxVal > 0 && (v < 1 || v > maxVal) {
		return 0, fmt.Errorf("value out of range 1-%d", maxVal)
	}
	if maxVal < 0 && (v < 0 || v > -maxVal) {
		return 0, fmt.Errorf("value out of range 0-%d", -maxVal)
	}
	return v, nil
}

//
// SetTime sets the current time and date by discrete values.
//
func (c *Calendar) SetTime(year, month, date, hour, min, sec, ticks int) error {
	mi := month - 1
	if month < 1 || month > len(c.Months) {
		return fmt.Errorf("month number %d out of range 1-%d", month, len(c.Months))
	}
	if c.ly(year) {
		if date < 1 || date > c.Months[mi].LYDays {
			return fmt.Errorf("date %d out of range 1-%d", date, c.Months[mi].LYDays)
		}
	} else if date < 1 || date > c.Months[mi].Days {
		return fmt.Errorf("date %d out of range 1-%d", date, c.Months[mi].Days)
	}

	c.Now = c.j(year, month, date)*int64(c.DayUnits) + int64(hour)*int64(c.HourUnits) + int64(min)*int64(c.MinuteUnits) + int64(sec)*int64(c.SecondUnits) + int64(ticks)
	c.recalc()
	return nil
}

//
// SetTimeNamed is like SetTime but uses month names instead of numeric values.
//
func (c *Calendar) SetTimeNamed(year int, month string, date, hour, min, sec, ticks int) error {
	for m, mi := range c.Months {
		if mi.FullName == month || mi.Abbrev == month {
			return c.SetTime(year, m+1, date, hour, min, sec, ticks)
		}
	}
	return fmt.Errorf("no month named \"%s\"", month)
}

//
// TicksToInterval returns the number of ticks between the current time
// in the Calendar receiver and the start of the next specified interval.
//
func (c *Calendar) TicksToInterval(unitName string) (int64, error) {
	if unitName == "" {
		return 1, nil
	}
	base := c.Now
	m, ok := c.UnitMultiplier[unitName]
	if !ok {
		return 0, fmt.Errorf("invalid time unit \"%s\" specified", unitName)
	}
	nextMod := imod64(base, m)
	if nextMod == 0 {
		return m, nil
	}
	return m - nextMod, nil
}

//
// NewCalendar creates a new Calendar object which is set up
// for the specified calendaring system.
//
// Currently supports the following calendar systems:
//   "donuttus"   The game world created by the author and his friends.
//   "golarion"   The game world for Paizo's Pathfinder system.
//   "gregorian"  The calendar used in most of Earth.
//
func NewCalendar(calSystem string, options ...CalendarOption) (Calendar, error) {
	// The code in here needs some refactoring (for one thing, it fails the open/close principle).
	const (
		s  = 10
		r  = 6 * s
		m  = 60 * s
		h  = 60 * m
		d  = 24 * h
		w  = 7 * d
		se = 91 * d
	)

	newCal := Calendar{
		System:         calSystem,
		SecondUnits:    s,
		RoundUnits:     r,
		MinuteUnits:    m,
		HourUnits:      h,
		DayUnits:       d,
		WeekUnits:      w,
		PomUnits:       7,
		SeasonUnits:    se,
		TickMod:        10,
		SecondMod:      60,
		MinuteMod:      60,
		HourMod:        24,
		DowMod:         7,
		PomMod:         4,
		SeasonMod:      4,
		MoonPhaseNames: []string{"NM", "1Q", "FM", "3Q"},
		UnitMultiplier: make(map[string]int64),
	}

	//
	// system-specific settings
	//
	switch calSystem {
	//   ____                           _
	//  / ___|_ __ ___  __ _  ___  _ __(_) __ _ _ __
	// | |  _| '__/ _ \/ _` |/ _ \| '__| |/ _` | '_ \
	// | |_| | | |  __/ (_| | (_) | |  | | (_| | | | |
	//  \____|_|  \___|\__, |\___/|_|  |_|\__,_|_| |_|
	//                 |___/
	case "gregorian":
		newCal.Months = []MonthInfo{
			{"January", "JAN", 31, 31},
			{"February", "FEB", 28, 29},
			{"March", "MAR", 31, 31},
			{"April", "APR", 30, 30},
			{"May", "MAY", 31, 31},
			{"June", "JUN", 30, 30},
			{"July", "JUL", 31, 31},
			{"August", "AUG", 31, 31},
			{"September", "SEP", 30, 30},
			{"October", "OCT", 31, 31},
			{"November", "NOV", 30, 30},
			{"December", "DEC", 31, 31},
		}
		newCal.Seasons = []string{"Winter", "Spring", "Summer", "Fall"}
		newCal.Days = []string{"Sunday", "Monday", "Tuesday", "Wednesday",
			"Thursday", "Friday", "Saturday"}
		newCal.EpochYear = 0
		newCal.EpochDow = 3
		newCal.EpochPom = 0
		newCal.EpochSeason = 0
		newCal.j = func(y, m, d int) int64 {
			m = imod((m + 9), 12)
			y -= m / 10
			return 365*int64(y) + pdiv(int64(y), 4) - pdiv(int64(y), 100) + pdiv(int64(y), 400) + pdiv(306*int64(m)+5, 10) + (int64(d) - 1)
		}
		newCal.ymd = func(j int64) (int, int, int) {
			y := pdiv(10000*j+14780, 3652425)
			d := int(j - (365*y + pdiv(y, 4) - pdiv(y, 100) + pdiv(y, 400)))
			if d < 0 {
				y -= 1
				d = int(j - (365*y + pdiv(y, 4) - pdiv(y, 100) + pdiv(y, 400)))
			}
			m := int(pdiv(100*int64(d)+52, 3060))

			return int(y) + (m+2)/12, imod((m+2), 12) + 1, d - (306*m+5)/10 + 1
		}
		newCal.r = func(c *Calendar) {
			// Not /quite/ accurate, but close enough for our purposes.
			// We'll just assume the solstices and equinoxes occur on the 21st
			// every year.
			c.Season = (c.Month - 1) / 3
			if (imod(c.Month, 4) == 3) && c.Date >= 21 {
				c.Season = imod(c.Season+1, 4)
			}
		}
		newCal.ly = func(y int) bool {
			return imod(y, 4) == 0 && (imod(y, 100) != 0 || imod(y, 400) == 0)
		}

	//  ____                    _   _
	// |  _ \  ___  _ __  _   _| |_| |_ _   _ ___
	// | | | |/ _ \| '_ \| | | | __| __| | | / __|
	// | |_| | (_) | | | | |_| | |_| |_| |_| \__ \
	// |____/ \___/|_| |_|\__,_|\__|\__|\__,_|___/
	//
	case "donuttus":
		newCal.Months = []MonthInfo{
			{"Lith", "LIT", 28, 28},
			{"Estuary", "EST", 28, 28},
			{"Elwinder", "ELW", 28, 28},
			{"Umbar", "UMB", 28, 28},
			{"Ektar", "EKT", 28, 28},
			{"Balthmaar", "BAL", 28, 28},
			{"Frobuary", "FRO", 28, 28},
			{"Trovuary", "TRO", 28, 28},
			{"Solmark", "SOL", 28, 28},
			{"Wedmark", "WED", 28, 28},
			{"Blotmath", "BLO", 28, 28},
			{"Aberlith", "ABE", 28, 28},
		}
		newCal.Seasons = []string{"Spring", "Summer", "Fall", "Winter"}
		newCal.Days = []string{"Sunday", "1stday", "2ndday", "3rdday",
			"4thday", "5thday", "6thday"}
		newCal.MoonPhaseNames = []string{"R", "Y", "W", "G", "C", "B", "K", "M"}
		newCal.EpochYear = 2195
		newCal.EpochDow = 0
		newCal.EpochPom = 0
		newCal.EpochSeason = 0
		newCal.PomMod = 8
		newCal.j = func(y, m, d int) int64 {
			return (int64(y)-int64(newCal.EpochYear))*364 + (int64(m)-1)*28 + (int64(d) - 1)
		}
		newCal.ymd = func(j int64) (int, int, int) {
			return int(pdiv(j, 364)) + newCal.EpochYear,
				int(pdiv(imod64(j, 364), 28)) + 1,
				int(imod64(j, 28) + 1)
		}
		newCal.r = func(c *Calendar) {}
		newCal.ly = func(y int) bool {
			return false
		}

	//   ____       _            _
	//  / ___| ___ | | __ _ _ __(_) ___  _ __
	// | |  _ / _ \| |/ _` | '__| |/ _ \| '_ \
	// | |_| | (_) | | (_| | |  | | (_) | | | |
	//  \____|\___/|_|\__,_|_|  |_|\___/|_| |_|
	//
	case "golarion":
		newCal.Months = []MonthInfo{
			{"Abadius", "ABA", 31, 31},
			{"Calistril", "CAL", 28, 29},
			{"Pharast", "PHA", 31, 31},
			{"Gozran", "GOZ", 30, 30},
			{"Desnus", "DES", 31, 31},
			{"Sarenith", "SAR", 30, 30},
			{"Erastus", "ERA", 31, 31},
			{"Arodus", "ARO", 31, 31},
			{"Rova", "ROV", 30, 30},
			{"Lamashan", "LAM", 31, 31},
			{"Neth", "NET", 30, 30},
			{"Kuthona", "KUT", 31, 31},
		}
		newCal.Seasons = []string{"Winter", "Spring", "Summer", "Fall"}
		newCal.Days = []string{"Moonday", "Toilday", "Wealday", "Oathday",
			"Fireday", "Starday", "Sunday"}
		newCal.EpochYear = 0
		newCal.EpochDow = 2
		newCal.EpochPom = 0
		newCal.EpochSeason = 0
		newCal.j = func(y, m, d int) int64 {
			m = imod((m + 9), 12)
			y -= m / 10
			return 365*int64(y) + pdiv(int64(y), 8) + pdiv(306*int64(m)+5, 10) + (int64(d) - 1)
		}
		newCal.ymd = func(j int64) (int, int, int) {
			y := pdiv(10000*j+14780, 3651250)
			d := int(j - (365*y + pdiv(y, 8)))
			if d < 0 {
				y -= 1
				d = int(j - (365*y + pdiv(y, 8)))
			}
			m := int(pdiv(100*int64(d)+52, 3060))

			return int(y) + (m+2)/12, imod((m+2), 12) + 1, d - (306*m+5)/10 + 1
		}
		newCal.r = func(c *Calendar) {
			// Not /quite/ accurate, but close enough for our purposes.
			// We'll just assume the solstices and equinoxes occur on the 21st
			// every year.
			c.Season = (c.Month - 1) / 3
			if (imod(c.Month, 4) == 3) && c.Date >= 21 {
				c.Season = imod(c.Season+1, 4)
			}
		}
		newCal.ly = func(y int) bool {
			return imod(y, 8) == 0
		}

	default:
		return newCal, fmt.Errorf("\"%s\" is not a supported calendaring system", calSystem)
	}

	for _, o := range options {
		o(&newCal)
	}

	for _, u := range []struct {
		names []string
		mult  int64
	}{
		{[]string{"s", "sec", "secs", "second", "seconds"}, int64(newCal.SecondUnits)},
		{[]string{"r", "rnd", "rnds", "round", "rounds"}, int64(newCal.RoundUnits)},
		{[]string{"m", "min", "mins", "minute", "minutes"}, int64(newCal.MinuteUnits)},
		{[]string{"h", "hr", "hrs", "hour", "hours"}, int64(newCal.HourUnits)},
		{[]string{"w", "wk", "wks", "week", "weeks"}, int64(newCal.WeekUnits)},
		{[]string{"d", "dy", "dys", "day", "days"}, int64(newCal.DayUnits)},
	} {
		for _, n := range u.names {
			newCal.UnitMultiplier[n] = u.mult
		}
	}
	newCal.recalc()
	return newCal, nil
}

//
// String renders the Calendar value as a string, in the format
//   <day> <date> <month> <year> <hh>:<mm>:<ss>.<tick> <season> <phase>
//
// This is equivalent to calling the ToString(1) method.
//
func (c Calendar) String() string {
	return c.ToString(1)
}

//
// ToString renders the Calendar value as a string, in the specified
// format. Supported formats include:
//   0: <date> <month> <year> <hh>:<mm>:<ss>.<tick>
//   1: <day> <date> <month> <year> <hh>:<mm>:<ss>.<tick> <season> <phase>
//   2: <date>-<month>-<year> <hh>:<mm>:<ss>.<tick>
//   3: <year>-<month>-<date> <hh>:<mm>:<ss>.<tick>
//   4: <hh>:<mm>:<ss>.<tick>
//
func (c Calendar) ToString(style int) string {
	switch style {
	case 0:
		return fmt.Sprintf("%02d %s %d %02d:%02d:%02d.%d",
			c.Date, c.Months[c.Month-1].FullName, c.Year,
			c.Hour, c.Minute, c.Second, c.Tick)
	case 2:
		return fmt.Sprintf("%02d-%s-%d %02d:%02d:%02d.%d",
			c.Date, c.Months[c.Month-1].Abbrev, c.Year,
			c.Hour, c.Minute, c.Second, c.Tick)
	case 3:
		return fmt.Sprintf("%d-%s-%02d %02d:%02d:%02d.%d",
			c.Year, c.Months[c.Month-1].Abbrev, c.Date,
			c.Hour, c.Minute, c.Second, c.Tick)
	case 4:
		return fmt.Sprintf("%02d:%02d:%02d.%d",
			c.Hour, c.Minute, c.Second, c.Tick)
	default:
		return fmt.Sprintf("%s %2d %s %d %02d:%02d:%02d.%d %s %s",
			c.Days[c.Dow], c.Date, c.Months[c.Month-1].FullName,
			c.Year, c.Hour, c.Minute, c.Second, c.Tick,
			c.Seasons[c.Season], c.MoonPhaseNames[c.Pom])
	}
}

//
// IsLeapYear returns true if the year tracked by the
// receiver is a leap year.
//
func (c *Calendar) IsLeapYear() bool {
	return c.ly(c.Year)
}

func pdiv(x, y int64) int64 {
	return int64(math.Floor(float64(x) / float64(y)))
}

func imod(x, y int) int {
	m := x % y
	if m < 0 {
		m += y
	}
	return m
}

func imod64(x, y int64) int64 {
	m := x % y
	if m < 0 {
		m += y
	}
	return m
}

func (c *Calendar) recalc() {
	j := c.Now / int64(c.DayUnits)
	t := imod64(c.Now, int64(c.DayUnits))
	c.Year, c.Month, c.Date = c.ymd(j)
	c.Tick = int(imod64(t, int64(c.TickMod)))
	c.Second = int(imod64(t/int64(c.SecondUnits), int64(c.SecondMod)))
	c.Minute = int(imod64(t/int64(c.MinuteUnits), int64(c.MinuteMod)))
	c.Hour = int(imod64(t/int64(c.HourUnits), int64(c.HourMod)))
	c.Dow = int(imod64(j+int64(c.EpochDow), int64(c.DowMod)))
	c.Pom = int(imod64((j+int64(c.EpochPom))/int64(c.PomUnits), int64(c.PomMod)))
	c.Season = int(imod64((j+int64(c.EpochSeason))/int64(c.SeasonUnits), int64(c.SeasonMod)))
	c.FSecond = math.Mod(float64(t)/float64(c.SecondUnits), float64(c.SecondMod))
	c.FMinute = math.Mod(float64(t)/float64(c.MinuteUnits), float64(c.MinuteMod))
	c.FHour = math.Mod(float64(t)/float64(c.HourUnits), float64(c.HourMod))
	c.r(c)
}

/*
#
# @[00]@| Go-GMA 5.16.0
# @[01]@|
# @[10]@| Overall GMA package Copyright © 1992–2024 by Steven L. Willoughby (AKA MadScienceZone)
# @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
# @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
# @[13]@| points along that historical time line.
# @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
# @[15]@| License as described in the accompanying LICENSE file distributed
# @[16]@| with GMA.
# @[17]@|
# @[20]@| Redistribution and use in source and binary forms, with or without
# @[21]@| modification, are permitted provided that the following conditions
# @[22]@| are met:
# @[23]@| 1. Redistributions of source code must retain the above copyright
# @[24]@|    notice, this list of conditions and the following disclaimer.
# @[25]@| 2. Redistributions in binary form must reproduce the above copy-
# @[26]@|    right notice, this list of conditions and the following dis-
# @[27]@|    claimer in the documentation and/or other materials provided
# @[28]@|    with the distribution.
# @[29]@| 3. Neither the name of the copyright holder nor the names of its
# @[30]@|    contributors may be used to endorse or promote products derived
# @[31]@|    from this software without specific prior written permission.
# @[32]@|
# @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
# @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
# @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
# @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
# @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
# @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
# @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
# @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
# @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
# @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
# @[45]@| SUCH DAMAGE.
# @[46]@|
# @[50]@| This software is not intended for any use or application in which
# @[51]@| the safety of lives or property would be at risk due to failure or
# @[52]@| defect of the software.
#
*/
