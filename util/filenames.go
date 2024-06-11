/*
\
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   __       _______        #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) /  \     / ___   )       #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  | \/) )    \/   )  |       #
# | |      | || || || (___) | Assistant | (____         /   )   | |        /   )       #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    | |      _/   /        #
# | | \_  )| |   | || (   ) |                 ) )    /   _/     | |     /   _/         #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ __) (_ _ (   (__/\       #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \____/(_)\_______/       #
#                                                                                      #
########################################################################################
*/

package util

import (
	"os"
	"strconv"
	"time"

	"github.com/lestrrat-go/strftime"
)

//
// FancyFileName expands tokens found in the path string to allow the user
// to specify dynamically-named files at runtime. If there's a problem with
// the formatting, an error is returned along with the original path.
//
// The tokens which may appear in the path include the following
// (note that all of these are modified as appropriate to the locale's
// national conventions and language):
//    %A   full weekday name
//    %a   abbreviated weekday name
//    %B   full month name
//    %b   abbreviated month name
//    %C   zero-padded two-digit year 00-99
//    %c   time and date
//    %d   day of month as number 01-31 (zero padded)
//    %e   day of month as number  1-31 (space padded)
//    %F   == %Y-%m-%d
//    %H   hour as number 00-23 (zero padded)
//    %h   abbreviated month name (same as %b)
//    %I   hour as number 01-12 (zero padded)
//    %j   day of year as number 001-366
//    %k   hour as number  0-23 (space padded)
//    %L   milliseconds as number 000-999
//    %l   hour as number  1-12 (space padded)
//    %M   minute as number 00-59
//    %m   month as number 01-12
//    %P   process ID
//    %p   AM or PM
//    %R   == %H:%M
//    %r   == %I:%M:%S %p
//    %S   second as number 00-60
//    %s   Unix timestamp as a number
//    %T   == %H:%M:%S
//    %U   week of the year as number 00-53 (Sunday as first day of week)
//    %u   weekday as number (1=Monday .. 7=Sunday)
//    %V   week of the year as number 00-53 (Monday as first day of week)
//    %v   == %e-%b-%Y
//    %W   week of the year as number 00-53 (Monday as first day of week)
//    %w   weekday as number (0=Sunday .. 6=Saturday)
//    %X   time
//    %x   date
//    %Y   full year
//    %y   two-digit year (00-99)
//    %Z   time zone name
//    %z   time zone offset from UTC
//    %µ   microseconds as number 000-999
//    %%   literal % character
//
// The extras parameter maps token names to static string values, so more
// tokens specific to the task at hand can be added, such as these that
// the mapper client adds:
//    %G   "GM" if logged in as the GM, otherwise ""
//    %N   username
//    %n   module name
//
func FancyFileName(path string, extras map[byte]string) (string, error) {
	ss := strftime.NewSpecificationSet()

	if err := ss.Delete('n'); err != nil {
		return path, err
	}
	if err := ss.Delete('t'); err != nil {
		return path, err
	}
	if err := ss.Delete('D'); err != nil {
		return path, err
	}
	if err := ss.Set('P', strftime.Verbatim(strconv.Itoa(os.Getpid()))); err != nil {
		return path, err
	}
	if extras != nil {
		for tokenName, tokenValue := range extras {
			if err := ss.Set(tokenName, strftime.Verbatim(tokenValue)); err != nil {
				return path, err
			}
		}
	}

	return strftime.Format(path, time.Now(),
		strftime.WithSpecificationSet(ss),
		strftime.WithUnixSeconds('s'),
		strftime.WithMilliseconds('L'),
		strftime.WithMicroseconds('µ'),
	)

}
