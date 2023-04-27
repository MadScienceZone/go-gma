/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     ______      _______               #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___  \    (  __   )              #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   \  \   | (  )  |              #
# | |      | || || || (___) | Assistant | (____        ___) /   | | /   |              #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      (___ (    | (/ /) |              #
# | | \_  )| |   | || (   ) |                 ) )         ) \   |   / | |              #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\___/  / _ |  (__) |              #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ (_)(_______)              #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the mapper calendar code
//

package gma

import (
	"reflect"
	"testing"
)

func TestGregorianJ(t *testing.T) {
	gc, err := NewCalendar("gregorian")
	if err != nil {
		t.Fatalf("%v", err)
	}

	for i, tc := range []struct {
		j       int64
		y, m, d int
	}{
		{-60, 0, 1, 1},
		{-2, 0, 2, 28},
		{-1, 0, 2, 29},
		{0, 0, 3, 1},
		{1, 0, 3, 2},
		{2, 0, 3, 3},
		{244, 0, 10, 31},
		{245, 0, 11, 1},
		{246, 0, 11, 2},
		{275, 0, 12, 1},
		{305, 0, 12, 31},
		{305, 0, 12, 31},
		{719298, 1969, 7, 15},
		{719468, 1970, 1, 1},
		{730424, 1999, 12, 31},
		{730425, 2000, 1, 1},
		{734262, 2010, 7, 4},
		{734566, 2011, 5, 4},
	} {
		cj := gc.j(tc.y, tc.m, tc.d)
		if tc.j != cj {
			t.Errorf("testcase %d: j=%v, expected %v", i, cj, tc.j)
		}
		cy, cm, cd := gc.ymd(tc.j)
		if cy != tc.y || cm != tc.m || cd != tc.d {
			t.Errorf("testcase %d: ymd=%v,%v,%v; expected %v,%v,%v",
				i, cy, cm, cd, tc.y, tc.m, tc.d)
		}
	}
}

func TestDonuttusJ(t *testing.T) {
	gc, err := NewCalendar("donuttus")
	if err != nil {
		t.Fatalf("%v", err)
	}

	for i, tc := range []struct {
		j       int64
		y, m, d int
	}{
		{0, 2195, 1, 1},
		{1, 2195, 1, 2},
		{2, 2195, 1, 3},
		{27, 2195, 1, 28},
		{28, 2195, 2, 1},
		{364, 2196, 1, 1},
		{-1, 2194, 13, 28},
	} {
		cj := gc.j(tc.y, tc.m, tc.d)
		if tc.j != cj {
			t.Errorf("testcase %d: j=%v, expected %v", i, cj, tc.j)
		}
		cy, cm, cd := gc.ymd(tc.j)
		if cy != tc.y || cm != tc.m || cd != tc.d {
			t.Errorf("testcase %d: ymd=%v,%v,%v; expected %v,%v,%v",
				i, cy, cm, cd, tc.y, tc.m, tc.d)
		}
	}
}

func TestGolarionJ(t *testing.T) {
	gc, err := NewCalendar("golarion")
	if err != nil {
		t.Fatalf("%v", err)
	}

	for i, tc := range []struct {
		j       int64
		y, m, d int
	}{
		{-60, 0, 1, 1},
		{-2, 0, 2, 28},
		{-1, 0, 2, 29},
		{0, 0, 3, 1},
		{1, 0, 3, 2},
		{2, 0, 3, 3},
		{244, 0, 10, 31},
		{245, 0, 11, 1},
		{246, 0, 11, 2},
		{275, 0, 12, 1},
		{305, 0, 12, 31},
		{305, 0, 12, 31},
		{306, 1, 1, 1},
		{364, 1, 2, 28},
		{365, 1, 3, 1},
		{719067, 1969, 7, 15},
		{719237, 1970, 1, 1},
		{730189, 1999, 12, 31},
		{730190, 2000, 1, 1},
		{734026, 2010, 7, 4},
		{734330, 2011, 5, 4},
		{1718867, 4707, 10, 11},
		{1719174, 4708, 8, 14},
	} {
		cj := gc.j(tc.y, tc.m, tc.d)
		if tc.j != cj {
			t.Errorf("testcase %d: j=%v, expected %v", i, cj, tc.j)
		}
		cy, cm, cd := gc.ymd(tc.j)
		if cy != tc.y || cm != tc.m || cd != tc.d {
			t.Errorf("testcase %d: ymd=%v,%v,%v; expected %v,%v,%v",
				i, cy, cm, cd, tc.y, tc.m, tc.d)
		}
	}
}

func TestDow(t *testing.T) {
	for i, tc := range []struct {
		world string
		n     int64
		dow   int
		msg   string
	}{
		{"donuttus", 0, 0, ""},
		{"gregorian", 719298 * 864000, 2, " 15 jul 1969"},
		{"gregorian", 719468 * 864000, 4, "  1 jan 1970"},
		{"gregorian", 730424 * 864000, 5, " 31 dec 1999"},
		{"gregorian", 734262 * 864000, 0, "  4 jul 2010"},
		{"gregorian", 734566 * 864000, 3, "  4 may 2011"},
		{"golarion", 1719174 * 864000, 4, " 14 aro 4708"},
		{"golarion", 1485101088000, 5, " 11 lam 4707"},
	} {
		gc, err := NewCalendar(tc.world, WithTime(tc.n))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if gc.Dow != tc.dow {
			t.Errorf("Testcase %d, dow for %s (%s)=%v, expected %v",
				i, tc.world, tc.msg, gc.Dow, tc.dow)
		}
	}
}

func TestEpoch(t *testing.T) {
	for i, tc := range []struct {
		world string
		dstr  string
	}{
		{"donuttus", "Sunday  1 Lith 2195 00:00:00.0 Spring R"},
	} {
		gc, err := NewCalendar(tc.world)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if gc.String() != tc.dstr {
			t.Errorf("Testcase %d, %q (expected %q)",
				i, gc.String(), tc.dstr)
		}
	}
}

func TestCurrent(t *testing.T) {
	for i, tc := range []struct {
		world string
		n     int64
		dstr  string
	}{
		{"gregorian", 621473472567, "Tuesday 15 July 1969 00:00:56.7 Summer NM"},
		{"golarion", 719237*864000 + 567, "Oathday  1 Abadius 1970 00:00:56.7 Winter NM"},
		{"golarion", 1719174*864000 + 567, "Fireday 14 Arodus 4708 00:00:56.7 Summer NM"},
		{"donuttus", 1, "Sunday  1 Lith 2195 00:00:00.1 Spring R"},
	} {
		gc, err := NewCalendar(tc.world, WithTime(tc.n))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if gc.String() != tc.dstr {
			t.Errorf("Testcase %d, %q (expected %q)",
				i, gc.String(), tc.dstr)
		}
	}
}

func TestDateBreakout(t *testing.T) {
	for i, tc := range []struct {
		world                        string
		n                            int64
		y, m, d, hr, mn, sc, dow, tk int
	}{
		{"donuttus", 0, 2195, 1, 1, 0, 0, 0, 0, 0},
		{"donuttus", 1, 2195, 1, 1, 0, 0, 0, 0, 1},
		{"donuttus", 1000, 2195, 1, 1, 0, 1, 40, 0, 0},
		{"donuttus", 1001, 2195, 1, 1, 0, 1, 40, 0, 1},
		{"donuttus", 100172, 2195, 1, 1, 2, 46, 57, 0, 2},
		{"gregorian", 719298 * 864000, 1969, 7, 15, 0, 0, 0, 2, 0},
		{"gregorian", 719468 * 864000, 1970, 1, 1, 0, 0, 0, 4, 0},
		{"gregorian", 730424 * 864000, 1999, 12, 31, 0, 0, 0, 5, 0},
		{"gregorian", 734262 * 864000, 2010, 7, 4, 0, 0, 0, 0, 0},
		{"gregorian", 734566 * 864000, 2011, 5, 4, 0, 0, 0, 3, 0},
		{"gregorian", 734566*864000 + 999, 2011, 5, 4, 0, 1, 39, 3, 9},
		{"golarion", 1719174 * 864000, 4708, 8, 14, 0, 0, 0, 4, 0},
	} {
		gc, err := NewCalendar(tc.world, WithTime(tc.n))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if gc.Year != tc.y {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Year, tc.y)
		}
		if gc.Month != tc.m {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Month, tc.m)
		}
		if gc.Date != tc.d {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Date, tc.d)
		}
		if gc.Hour != tc.hr {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Hour, tc.hr)
		}
		if gc.Minute != tc.mn {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Minute, tc.mn)
		}
		if gc.Second != tc.sc {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Second, tc.sc)
		}
		if gc.Dow != tc.dow {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Dow, tc.dow)
		}
		if gc.Tick != tc.tk {
			t.Errorf("testcase %d, %s year %d, expected %d", i, tc.world, gc.Tick, tc.tk)
		}
	}
}

func TestAdvanceByUnit(t *testing.T) {
	gc, err := NewCalendar("gregorian")
	if err != nil {
		t.Fatalf("%v", err)
	}

	for i, tc := range []struct {
		start  int64
		fld    string
		unames []string
	}{
		{1, "Date", []string{"day", "days", "dy", "dys"}},
		{0, "Hour", []string{"hour", "hours", "hr", "hrs"}},
		{0, "Minute", []string{"min", "minute", "mins", "minutes"}},
		{0, "Second", []string{"sec", "second", "seconds", "secs", "s"}},
	} {
		for _, uname := range tc.unames {
			gc.SetTimeValue(0)
			if v := reflect.ValueOf(gc).FieldByName(tc.fld).Int(); v != tc.start {
				t.Errorf("testcase %d.%s: error resetting clock %s=%v, expected %v", i, uname, tc.fld, v, tc.start)
			}
			_, err := gc.Advance(+1, uname)
			if err != nil {
				t.Fatalf("testcase %d.%s: %v", i, uname, err)
			}
			if v := reflect.ValueOf(gc).FieldByName(tc.fld).Int(); v != tc.start+1 {
				t.Errorf("testcase %d.%s: error advancing %s by %d %s=%v, expected %v", i, uname, uname, +1, tc.fld, v, tc.start+1)
			}
			_, err = gc.Advance(+1, uname)
			if err != nil {
				t.Fatalf("testcase %d.%s: %v", i, uname, err)
			}
			if v := reflect.ValueOf(gc).FieldByName(tc.fld).Int(); v != tc.start+2 {
				t.Errorf("testcase %d.%s: error advancing %s by %d %s=%v, expected %v", i, uname, uname, +2, tc.fld, v, tc.start+2)
			}
			_, err = gc.Advance(+5, uname)
			if err != nil {
				t.Fatalf("testcase %d.%s: %v", i, uname, err)
			}
			if v := reflect.ValueOf(gc).FieldByName(tc.fld).Int(); v != tc.start+7 {
				t.Errorf("testcase %d.%s: error advancing %s by %d %s=%v, expected %v", i, uname, uname, +7, tc.fld, v, tc.start+7)
			}
			_, err = gc.Advance(-3, uname)
			if err != nil {
				t.Fatalf("testcase %d.%s: %v", i, uname, err)
			}
			if v := reflect.ValueOf(gc).FieldByName(tc.fld).Int(); v != tc.start+4 {
				t.Errorf("testcase %d.%s: error advancing %s by %d %s=%v, expected %v", i, uname, uname, +4, tc.fld, v, tc.start+4)
			}
		}
	}
}

func TestAdvanceIntervals(t *testing.T) {
	gc, err := NewCalendar("gregorian")
	if err != nil {
		t.Fatalf("%v", err)
	}
	for i, tc := range []struct {
		intname string
		result  string
	}{
		{"second", "01-MAR-0 00:00:01.0"},
		{"second", "01-MAR-0 00:00:02.0"},
		{"minute", "01-MAR-0 00:01:00.0"},
		{"second", "01-MAR-0 00:01:01.0"},
		{"", "01-MAR-0 00:01:01.1"},
		{"", "01-MAR-0 00:01:01.2"},
		{"", "01-MAR-0 00:01:01.3"},
		{"round", "01-MAR-0 00:01:06.0"},
		{"hour", "01-MAR-0 01:00:00.0"},
		{"minute", "01-MAR-0 01:01:00.0"},
		{"hour", "01-MAR-0 02:00:00.0"},
	} {
		gc.AdvanceToNext(tc.intname)
		if gc.ToString(2) != tc.result {
			t.Errorf("testcase %d, advanced %q -> %q, expected %q",
				i, tc.intname, gc.ToString(2), tc.result)
		}
	}
}

func TestDelta(t *testing.T) {
	gc1, err := NewCalendar("gregorian", WithTime(15200))
	if err != nil {
		t.Fatalf("%v", err)
	}
	gc2, err := NewCalendar("gregorian", WithTime(32998))
	if err != nil {
		t.Fatalf("%v", err)
	}

	if gc1.ToString(2) != "01-MAR-0 00:25:20.0" {
		t.Errorf("gc1 %q", gc1.ToString(2))
	}
	if gc2.ToString(2) != "01-MAR-0 00:54:59.8" {
		t.Errorf("gc2 %q", gc2.ToString(2))
	}
	delta := gc2.Delta(gc1)
	if delta != 32998-15200 {
		t.Errorf("delta %v", delta)
	}
}

func TestImage(t *testing.T) {
	gc, err := NewCalendar("golarion")
	if err != nil {
		t.Fatalf("%v", err)
	}
	gc.SetTime(4718, 1, 1, 0, 0, 0, 0)
	image, err := gc.Image()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !reflect.DeepEqual(image, [][][]int{
		{{1, 2, 3, 4, 5, 6, 7},
			{8, 9, 10, 11, 12, 13, 14},
			{15, 16, 17, 18, 19, 20, 21},
			{22, 23, 24, 25, 26, 27, 28},
			{29, 30, 31, 0, 0, 0, 0}},
		{{0, 0, 0, 1, 2, 3, 4},
			{5, 6, 7, 8, 9, 10, 11},
			{12, 13, 14, 15, 16, 17, 18},
			{19, 20, 21, 22, 23, 24, 25},
			{26, 27, 28, 0, 0, 0, 0}},
		{{0, 0, 0, 1, 2, 3, 4},
			{5, 6, 7, 8, 9, 10, 11},
			{12, 13, 14, 15, 16, 17, 18},
			{19, 20, 21, 22, 23, 24, 25},
			{26, 27, 28, 29, 30, 31, 0}},
		{{0, 0, 0, 0, 0, 0, 1},
			{2, 3, 4, 5, 6, 7, 8},
			{9, 10, 11, 12, 13, 14, 15},
			{16, 17, 18, 19, 20, 21, 22},
			{23, 24, 25, 26, 27, 28, 29},
			{30, 0, 0, 0, 0, 0, 0}},
		{{0, 1, 2, 3, 4, 5, 6},
			{7, 8, 9, 10, 11, 12, 13},
			{14, 15, 16, 17, 18, 19, 20},
			{21, 22, 23, 24, 25, 26, 27},
			{28, 29, 30, 31, 0, 0, 0}},
		{{0, 0, 0, 0, 1, 2, 3},
			{4, 5, 6, 7, 8, 9, 10},
			{11, 12, 13, 14, 15, 16, 17},
			{18, 19, 20, 21, 22, 23, 24},
			{25, 26, 27, 28, 29, 30, 0}},
		{{0, 0, 0, 0, 0, 0, 1},
			{2, 3, 4, 5, 6, 7, 8},
			{9, 10, 11, 12, 13, 14, 15},
			{16, 17, 18, 19, 20, 21, 22},
			{23, 24, 25, 26, 27, 28, 29},
			{30, 31, 0, 0, 0, 0, 0}},
		{{0, 0, 1, 2, 3, 4, 5},
			{6, 7, 8, 9, 10, 11, 12},
			{13, 14, 15, 16, 17, 18, 19},
			{20, 21, 22, 23, 24, 25, 26},
			{27, 28, 29, 30, 31, 0, 0}},
		{{0, 0, 0, 0, 0, 1, 2},
			{3, 4, 5, 6, 7, 8, 9},
			{10, 11, 12, 13, 14, 15, 16},
			{17, 18, 19, 20, 21, 22, 23},
			{24, 25, 26, 27, 28, 29, 30}},
		{{1, 2, 3, 4, 5, 6, 7},
			{8, 9, 10, 11, 12, 13, 14},
			{15, 16, 17, 18, 19, 20, 21},
			{22, 23, 24, 25, 26, 27, 28},
			{29, 30, 31, 0, 0, 0, 0}},
		{{0, 0, 0, 1, 2, 3, 4},
			{5, 6, 7, 8, 9, 10, 11},
			{12, 13, 14, 15, 16, 17, 18},
			{19, 20, 21, 22, 23, 24, 25},
			{26, 27, 28, 29, 30, 0, 0}},
		{{0, 0, 0, 0, 0, 1, 2},
			{3, 4, 5, 6, 7, 8, 9},
			{10, 11, 12, 13, 14, 15, 16},
			{17, 18, 19, 20, 21, 22, 23},
			{24, 25, 26, 27, 28, 29, 30},
			{31, 0, 0, 0, 0, 0, 0}}}) {
		t.Errorf("cal image %v", image)
	}

	gc.SetTime(4719, 5, 5, 5, 5, 5, 5)
	image, err = gc.Image()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !reflect.DeepEqual(image, [][][]int{
		{{0, 1, 2, 3, 4, 5, 6}, {7, 8, 9, 10, 11, 12, 13}, {14, 15, 16, 17, 18, 19, 20}, {21, 22, 23, 24, 25, 26, 27}, {28, 29, 30, 31, 0, 0, 0}}, {{0, 0, 0, 0, 1, 2, 3}, {4, 5, 6, 7, 8, 9, 10}, {11, 12, 13, 14, 15, 16, 17}, {18, 19, 20, 21, 22, 23, 24}, {25, 26, 27, 28, 0, 0, 0}}, {{0, 0, 0, 0, 1, 2, 3}, {4, 5, 6, 7, 8, 9, 10}, {11, 12, 13, 14, 15, 16, 17}, {18, 19, 20, 21, 22, 23, 24}, {25, 26, 27, 28, 29, 30, 31}}, {{1, 2, 3, 4, 5, 6, 7}, {8, 9, 10, 11, 12, 13, 14}, {15, 16, 17, 18, 19, 20, 21}, {22, 23, 24, 25, 26, 27, 28}, {29, 30, 0, 0, 0, 0, 0}}, {{0, 0, 1, 2, 3, 4, 5}, {6, 7, 8, 9, 10, 11, 12}, {13, 14, 15, 16, 17, 18, 19}, {20, 21, 22, 23, 24, 25, 26}, {27, 28, 29, 30, 31, 0, 0}}, {{0, 0, 0, 0, 0, 1, 2}, {3, 4, 5, 6, 7, 8, 9}, {10, 11, 12, 13, 14, 15, 16}, {17, 18, 19, 20, 21, 22, 23}, {24, 25, 26, 27, 28, 29, 30}}, {{1, 2, 3, 4, 5, 6, 7}, {8, 9, 10, 11, 12, 13, 14}, {15, 16, 17, 18, 19, 20, 21}, {22, 23, 24, 25, 26, 27, 28}, {29, 30, 31, 0, 0, 0, 0}}, {{0, 0, 0, 1, 2, 3, 4}, {5, 6, 7, 8, 9, 10, 11}, {12, 13, 14, 15, 16, 17, 18}, {19, 20, 21, 22, 23, 24, 25}, {26, 27, 28, 29, 30, 31, 0}}, {{0, 0, 0, 0, 0, 0, 1}, {2, 3, 4, 5, 6, 7, 8}, {9, 10, 11, 12, 13, 14, 15}, {16, 17, 18, 19, 20, 21, 22}, {23, 24, 25, 26, 27, 28, 29}, {30, 0, 0, 0, 0, 0, 0}}, {{0, 1, 2, 3, 4, 5, 6}, {7, 8, 9, 10, 11, 12, 13}, {14, 15, 16, 17, 18, 19, 20}, {21, 22, 23, 24, 25, 26, 27}, {28, 29, 30, 31, 0, 0, 0}}, {{0, 0, 0, 0, 1, 2, 3}, {4, 5, 6, 7, 8, 9, 10}, {11, 12, 13, 14, 15, 16, 17}, {18, 19, 20, 21, 22, 23, 24}, {25, 26, 27, 28, 29, 30, 0}}, {{0, 0, 0, 0, 0, 0, 1}, {2, 3, 4, 5, 6, 7, 8}, {9, 10, 11, 12, 13, 14, 15}, {16, 17, 18, 19, 20, 21, 22}, {23, 24, 25, 26, 27, 28, 29}, {30, 31, 0, 0, 0, 0, 0}}}) {
		t.Errorf("cal image %v", image)
	}
}

/*
#
# @[00]@| Go-GMA 5.3.0
# @[01]@|
# @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
# @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
# @[12]@| Aloha, Oregon, USA. All Rights Reserved.
# @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
# @[14]@| License as described in the accompanying LICENSE file distributed
# @[15]@| with GMA.
# @[16]@|
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
# Adapted for the Pathfinder RPG, which is what we're playing now
# (and this software is primarily for our own use in our play group,
# anyway, but could be generalized later as a stand-alone product).
#
########################################################################
*/
