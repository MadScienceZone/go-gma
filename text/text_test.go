/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      _____       __                   #
# (  ____ \(       )(  ___  ) Game      (  ____ \    / ___ \     /  \                  #
# | (    \/| () () || (   ) | Master's  | (    \/   ( (   ) )    \/) )                 #
# | |      | || || || (___) | Assistant | (____     ( (___) |      | |                 #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \     \____  |      | |                 #
# | | \_  )| |   | || (   ) |                 ) )         ) |      | |                 #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ /\____) ) _  __) (_                #
# (_______)|/     \||/     \| Client    \______/ (_)\______/ (_) \____/                #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the text package
//

package text

import (
	"testing"
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func TestToRoman(t *testing.T) {
	type testcase struct {
		in  int
		out string
		err bool
	}

	for i, test := range []testcase{
		{-1, "", true},
		{0, "0", false},
		{1, "I", false},
		{2, "II", false},
		{3, "III", false},
		{4, "IV", false},
		{5, "V", false},
		{6, "VI", false},
		{7, "VII", false},
		{8, "VIII", false},
		{9, "IX", false},
		{10, "X", false},
		{11, "XI", false},
		{14, "XIV", false},
		{15, "XV", false},
		{16, "XVI", false},
		{20, "XX", false},
		{42, "XLII", false},
		{44, "XLIV", false},
		{45, "XLV", false},
		{49, "XLIX", false},
		{50, "L", false},
		{51, "LI", false},
		{5555, "MMMMMDLV", false},
		{6666, "MMMMMMDCLXVI", false},
		{4444, "MMMMCDXLIV", false},
		{9876, "MMMMMMMMMDCCCLXXVI", false},
	} {
		r, err := ToRoman(test.in)
		if err != nil && !test.err {
			t.Errorf("Case %d: unexpected error: %v", i, err)
		} else if err == nil && test.err {
			t.Errorf("Case %d: error expected but not found", i)
		} else if test.out != r {
			t.Errorf("Case %d: %d -> %s but expected %s", i, test.in, r, test.out)
		}
	}
}

func TestFromRoman(t *testing.T) {
	type testcase struct {
		out int
		in  string
		err bool
	}

	for i, test := range []testcase{
		{0, "-1", true},
		{0, "0", false},
		{1, "I", false},
		{2, "II", false},
		{3, "  III", false},
		{4, "IV\t", false},
		{5, "V", false},
		{6, "VI", false},
		{7, "VII", false},
		{8, "VIII", false},
		{9, "IX", false},
		{10, "X", false},
		{0, "X X", true},
		{0, "X-X", true},
		{11, "XI", false},
		{14, "XIV", false},
		{15, "XV", false},
		{16, "XVI", false},
		{20, "XX", false},
		{42, "XLII", false},
		{44, "xliv", false},
		{45, "XLV", false},
		{49, "XLIX", false},
		{50, "L", false},
		{51, "LI", false},
		{5555, "MMMMMDLV", false},
		{6666, "MMMMMMDCLXVI", false},
		{4444, "MMMMCDXLIV", false},
		{9876, "MMMMMMMMMDCCCLXXVI", false},
	} {
		v, err := FromRoman(test.in)
		if err != nil && !test.err {
			t.Errorf("Case %d: unexpected error: %v", i, err)
		} else if err == nil && test.err {
			t.Errorf("Case %d: error expected but not found", i)
		} else if test.out != v {
			t.Errorf("Case %d: %s -> %d but expected %d", i, test.in, v, test.out)
		}
	}
}

func TestMarkupTextNull(t *testing.T) {
	type testcase struct {
		in   string
		out  string
		err  bool
		opts []func(*renderOptSet)
	}

	for i, test := range []testcase{
		{"foo", "foo", false, nil},
		{"", "", false, nil},
		{"foo\nbar", "foo bar", false, nil},
		{"foo\n\nbar", "foo\n\nbar", false, nil},
		{"\n\nfoo", "foo", false, nil},
		{"foo\\\\bar", "foo\nbar", false, nil},
		{"aa//bb//cc", "aabbcc", false, nil},
		{"aa**bb**cc", "aabbcc", false, nil},
		{"aa**bb", "aabb", false, nil},
		{"a//b**c**d//e", "abcde", false, nil},
		{"a //b **c //d **e", "a b c d e", false, nil},
		{"a[[b]]d", "abd", false, nil},
		{"a[[b|c]]d", "acd", false, nil},
		{"a //it b\nc\\\\de//f", "a it b c\ndef", false, nil},
		{"a //it b\nc\n\nde//f", "a it b c\n\ndef", false, nil},
		{`This is a bullet list:
*Item One
*Item //Tw//o
*Item Three
  * this is not a\\bullet list
*But this is
**and a sub-list
*** and sub-sub-list

And this should start a new list:
*Not that you can tell with bullets.
`, `This is a bullet list:
•  Item One
•  Item Two
•  Item Three * this is not a
   bullet list
•  But this is
   •  and a sub-list
      •  and sub-sub-list

And this should start a new list:
•  Not that you can tell with bullets.`, false, nil},
		{`This is a bullet list:
*Item One
*Item //Tw//o
*Item Three
  * this is not a\\bullet list
*But this is
**and a sub-list
*** and sub-sub-list
****But this is
*****and a sub-list

And this should start a new list:
*Not that you can tell with bullets.
`, `This is a bullet list:
•  Item One
•  Item Two
•  Item Three * this is not a
   bullet list
•  But this is
   ‣  and a sub-list
      ◦  and sub-sub-list
         •  But this is
            ‣  and a sub-list

And this should start a new list:
•  Not that you can tell with bullets.`, false, []func(*renderOptSet){WithBullets('•', '‣', '◦')}},
		{`This is a numbered list:
#Item One
#Item Two
#Item Three
 # this is not a\\bullet list
#But this is
##and a sub-list
### and sub-sub-list

And this should start a new list:
#and this should be re-sequenced.
`, `This is a numbered list:
1. Item One
2. Item Two
3. Item Three # this is not a
   bullet list
4. But this is
   a. and a sub-list
      i. and sub-sub-list

And this should start a new list:
1. and this should be re-sequenced.`, false, nil},
		{`Table test:
|=Column A|=Column B|
|left     |    right|
|  center |filled|
|  aaa    |   bbb
And this is after the table.`, `Table test:
+----------+----------+
| COLUMN A | COLUMN B |
+----------+----------+
| left     |    right |
|  center  | filled   |
|   aaa    |      bbb |
+----------+----------+
And this is after the table.`, false, nil},
		{`Table test:
|=Column A|=Column B|=Column C|
|left     |    right| some other |
|  center |filled and more stuff too |-
|  aaa and |-  |   bbb
And this is after the table.`, `Table test:
+----------+------------+--------------+
| COLUMN A |  COLUMN B  |   COLUMN C   |
+----------+------------+--------------+
| left     |      right |  some other  |
|  center  | filled and more stuff too |
|        aaa and        |          bbb |
+----------+------------+--------------+
And this is after the table.`, false, nil},
	} {
		test.opts = append(test.opts, AsPlainText)
		v, err := Render(test.in, test.opts...)
		if err != nil && !test.err {
			t.Errorf("Case %d: unexpected error: %v", i, err)
		} else if err == nil && test.err {
			t.Errorf("Case %d: error expected but not found", i)
		} else if test.out != v {
			t.Errorf("Case %d: %v -> %v but expected %v", i, test.in, v, test.out)
		}
	}
}

func TestMarkupTextHTML(t *testing.T) {
	type testcase struct {
		in   string
		out  string
		err  bool
		opts []func(*renderOptSet)
	}

	for i, test := range []testcase{
		{"foo", "<P>foo</P>", false, nil},
		{"", "<P></P>", false, nil},
		{"foo\nbar", "<P>foo bar</P>", false, nil},
		{"foo\n\nbar", "<P>foo</P><P>bar</P>", false, nil},
		{"\n\nfoo", "<P>foo</P>", false, nil},
		{`foo\\bar`, "<P>foo<BR/>bar</P>", false, nil},
		{"aa//bb//cc", "<P>aa<I>bb</I>cc</P>", false, nil},
		{"aa**bb**cc", "<P>aa<B>bb</B>cc</P>", false, nil},
		{"aa**bb", "<P>aa<B>bb</B></P>", false, nil},
		{"a//b**c**d//e", "<P>a<I>b<B>c</B>d</I>e</P>", false, nil},
		{"a //b **c //d **e", "<P>a <I>b <B>c </I>d </B>e</P>", false, nil},
		{"a[[b]]d", "<P>a<A HREF=\"B\">b</A>d</P>", false, nil},
		{"a[[b|c]]d", "<P>a<A HREF=\"B\">c</A>d</P>", false, nil},
		{"a //it b\nc\\\\de//f", "<P>a <I>it b c<BR/>de</I>f</P>", false, nil},
		{"a //it b\nc\n\nde//f", "<P>a <I>it b c</I></P><P>de<I>f</I></P>", false, nil},
		{`This is a bullet list:
*Item One
*Item //Tw//o
*Item Three
 * this is not a\\bullet list
*But this is
**and a sub-list
*** and sub-sub-list

And this should start a new list:
*Not that you can tell with bullets.
`, "<P>This is a bullet list:<UL><LI>Item One<LI>Item <I>Tw</I>o<LI>Item Three * this is not a<BR/>bullet list<LI>But this is<UL><LI>and a sub-list<UL><LI>and sub-sub-list</UL></UL></UL></P><P>And this should start a new list:<UL><LI>Not that you can tell with bullets.</UL></P>", false, nil},
		{`This is a bullet list:
*Item One
*Item //Tw//o
*Item Three
 * this is not a\\bullet list
*But this is
**and a sub-list
*** and sub-sub-list
**** four

And this should start a new list:
*Not that you can tell with bullets.
`, "<P>This is a bullet list:<UL style='list-style-type:\"disc\";'><LI>Item One<LI>Item <I>Tw</I>o<LI>Item Three * this is not a<BR/>bullet list<LI>But this is<UL style='list-style-type:\"\\2023\";'><LI>and a sub-list<UL style='list-style-type:\"\\0025e6\";'><LI>and sub-sub-list<UL style='list-style-type:\"disc\";'><LI>four</UL></UL></UL></UL></P><P>And this should start a new list:<UL style='list-style-type:\"disc\";'><LI>Not that you can tell with bullets.</UL></P>", false, []func(*renderOptSet){WithBullets('•', '‣', '◦')}},
		{`This is a numbered list:
#Item One
#Item Two
#Item Three
 # this is not a\\bullet list
#But this is
##and a sub-list
### and sub-sub-list

And this should start a new list:
#and this should be re-sequenced.
`, "<P>This is a numbered list:<OL style=\"list-style-type: decimal;\"><LI>Item One<LI>Item Two<LI>Item Three # this is not a<BR/>bullet list<LI>But this is<OL style=\"list-style-type: lower-alpha;\"><LI>and a sub-list<OL style=\"list-style-type: lower-roman;\"><LI>and sub-sub-list</OL></OL></OL></P><P>And this should start a new list:<OL style=\"list-style-type: decimal;\"><LI>and this should be re-sequenced.</OL></P>", false, nil},
		{`Table test:
|=Column A|=Column B|
|left     |    right|
|  center |filled|
|  aaa    |   bbb
And this is after the table.`, "<P>Table test:<TABLE BORDER=1><TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH></TR><TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD></TR><TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT>filled</TD></TR><TR><TD ALIGN=CENTER>aaa</TD><TD ALIGN=RIGHT>bbb</TD></TR></TABLE>And this is after the table.</P>", false, nil},
		{`Table test:
|=Column A|=Column B|=Column C|
|left     |    right| some stuff |
|  center |filled and extended to the other |-
|  aaa and |-   |   bbb
And this is after the table.`, "<P>Table test:<TABLE BORDER=1><TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH><TH ALIGN=LEFT>Column C</TH></TR><TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD><TD ALIGN=CENTER>some stuff</TD></TR><TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT COLSPAN=2>filled and extended to the other</TD></TR><TR><TD ALIGN=CENTER COLSPAN=2>aaa and</TD><TD ALIGN=RIGHT>bbb</TD></TR></TABLE>And this is after the table.</P>", false, nil},
	} {
		test.opts = append(test.opts, AsHTML)
		v, err := Render(test.in, test.opts...)
		if err != nil && !test.err {
			t.Errorf("Case %d: unexpected error: %v", i, err)
		} else if err == nil && test.err {
			t.Errorf("Case %d: error expected but not found", i)
		} else if test.out != v {
			t.Errorf("Case %d: %v -> %v but expected %v", i, test.in, v, test.out)
		}
	}
}

func TestMarkupTextPostScript(t *testing.T) {
	type testcase struct {
		in   string
		out  string
		err  bool
		opts []func(*renderOptSet)
	}

	for i, test := range []testcase{
		{"foo", " [  [ {} [ (foo) ] {PsFF_rm} ]  ] ", false, nil},
		{"", " [  ] ", false, nil},
		{"foo\nbar", " [  [ {} [ (foo )(bar) ] {PsFF_rm} ]  ] ", false, nil},
		{"foo\n\nbar", " [  [ {PsFF_par} [ (foo) ] {PsFF_rm} ]  [ {} [ (bar) ] {} ]  ] ", false, nil},
		{"\n\nfoo", " [  [ {} [ (foo) ] {PsFF_rm} ]  ] ", false, nil},
		{`foo\\bar`, " [  [ {PsFF_nl} [ (foo) ] {PsFF_rm} ]  [ {} [ (bar) ] {} ]  ] ", false, nil},
		{"aa//bb//cc", " [  [ {} [ (aa) ] {PsFF_rm} ]  [ {} [ (bb) ] {PsFF_it} ]  [ {} [ (cc) ] {PsFF_rm} ]  ] ", false, nil},
		{"aa**bb", " [  [ {} [ (aa) ] {PsFF_rm} ]  [ {} [ (bb) ] {PsFF_bf} ]  ] ", false, nil},
		{"a//b**c**d//e", " [  [ {} [ (a) ] {PsFF_rm} ]  [ {} [ (b) ] {PsFF_it} ]  [ {} [ (c) ] {PsFF_bi} ]  [ {} [ (d) ] {PsFF_it} ]  [ {} [ (e) ] {PsFF_rm} ]  ] ", false, nil},
		{"a //b **c //d **e", " [  [ {} [ (a ) ] {PsFF_rm} ]  [ {} [ (b ) ] {PsFF_it} ]  [ {} [ (c ) ] {PsFF_bi} ]  [ {} [ (d ) ] {PsFF_bf} ]  [ {} [ (e) ] {PsFF_rm} ]  ] ", false, nil},
		{"a[[b]]d", " [  [ {} [ (a) ] {PsFF_rm} ]  [ {} [ (b) ] {PsFF_it} ]  [ {} [ (d) ] {PsFF_rm} ]  ] ", false, nil},
		{"a[[b|c]]d", " [  [ {} [ (a) ] {PsFF_rm} ]  [ {} [ (c) ] {PsFF_it} ]  [ {} [ (d) ] {PsFF_rm} ]  ] ", false, nil},
		{`a //it b
c\\de//f`, " [  [ {} [ (a ) ] {PsFF_rm} ]  [ {PsFF_nl} [ (it )(b )(c) ] {PsFF_it} ]  [ {} [ (de) ] {} ]  [ {} [ (f) ] {PsFF_rm} ]  ] ",
			false, nil},
		{`a //it b
c

de//f`, " [  [ {} [ (a ) ] {PsFF_rm} ]  [ {} [ (it )(b )(c) ] {PsFF_it} ]  [ {PsFF_par} [  ] {PsFF_rm} ]  [ {} [ (de) ] {} ]  [ {} [ (f) ] {PsFF_it} ]  ] ", false, nil},
		{`This is a bullet list:
*Item One
*Item //Tw//o
*Item Three
 * this is not a\\bullet list
*But this is
**and a sub-list
*** and sub-sub-list

And this should start a new list:
*Not that you can tell with bullets.
`, " [  [ {PsFF_nl} [ (This )(is )(a )(bullet )(list:) ] {PsFF_rm} ] " +
			" [ { 1 PsFF_ind } [ (\\267) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (Item )(One) ] {} ] " +
			" [ { 1 PsFF_ind } [ (\\267) ] { 0 PsFF_ind } ] " +
			" [ {} [ (Item ) ] {} ] " +
			" [ {} [ (Tw) ] {PsFF_it} ] " +
			" [ {PsFF_nl} [ (o) ] {PsFF_rm} ] " +
			" [ { 1 PsFF_ind } [ (\\267) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (Item )(Three )(* )(this )(is )(not )(a) ] {} ] " +
			" [ {PsFF_nl} [ (bullet )(list) ] {} ] " +
			" [ { 1 PsFF_ind } [ (\\267) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (But )(this )(is) ] {} ] " +
			" [ { 2 PsFF_ind } [ (\\267) ] { 1 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (and )(a )(sub\\255list) ] {} ] " +
			" [ { 3 PsFF_ind } [ (\\267) ] { 2 PsFF_ind } ] " +
			" [ {PsFF_par 0 PsFF_ind} [ (and )(sub\\255sub\\255list) ] {} ] " +
			" [ {PsFF_nl} [ (And )(this )(should )(start )(a )(new )(list:) ] {} ] " +
			" [ { 1 PsFF_ind } [ (\\267) ] { 0 PsFF_ind } ] " +
			" [ {} [ (Not )(that )(you )(can )(tell )(with )(bullets.) ] {} ]  ] ", false, nil},
		{`This is a numbered list:
#Item One
#Item //Tw//o
#Item Three
 # this is not a\\bullet list
#But this is
##and a sub-list
### and sub-sub-list

And this should start a new list:
#and this should be re-sequenced.
`, " [  [ {PsFF_nl} [ (This )(is )(a )(numbered )(list:) ] {PsFF_rm} ] " +
			" [ { 1 PsFF_ind } [ (1.) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (Item )(One) ] {} ] " +
			" [ { 1 PsFF_ind } [ (2.) ] { 0 PsFF_ind } ] " +
			" [ {} [ (Item ) ] {} ] " +
			" [ {} [ (Tw) ] {PsFF_it} ] " +
			" [ {PsFF_nl} [ (o) ] {PsFF_rm} ] " +
			" [ { 1 PsFF_ind } [ (3.) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (Item )(Three )(# )(this )(is )(not )(a) ] {} ] " +
			" [ {PsFF_nl} [ (bullet )(list) ] {} ] " +
			" [ { 1 PsFF_ind } [ (4.) ] { 0 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (But )(this )(is) ] {} ] " +
			" [ { 2 PsFF_ind } [ (a.) ] { 1 PsFF_ind } ] " +
			" [ {PsFF_nl} [ (and )(a )(sub\\255list) ] {} ] " +
			" [ { 3 PsFF_ind } [ (i.) ] { 2 PsFF_ind } ] " +
			" [ {PsFF_par 0 PsFF_ind} [ (and )(sub\\255sub\\255list) ] {} ] " +
			" [ {PsFF_nl} [ (And )(this )(should )(start )(a )(new )(list:) ] {} ] " +
			" [ { 1 PsFF_ind } [ (1.) ] { 0 PsFF_ind } ] " +
			" [ {} [ (and )(this )(should )(be )(re\\255sequenced.) ] {} ]  ] ", false, nil},
		{`Table test:
|=Column A|=Column B|
|left     |    right|
|  center |filled|
|  aaa    |   bbb
And this is after the table.`, ` [  [ {PsFF_nl} [ (Table )(test:) ] {PsFF_rm} ] ` +
			` [ {PsFF_rm
%
% Data Table: calculate column widths
%
/PsFF_Cw0 0 def
[(Column A) (left) (center) (aaa) ] {
	stringwidth pop dup PsFF_Cw0 gt {
		/PsFF_Cw0 exch def
	} {
		pop
	} ifelse
} forall
/PsFF_Cw1 0 def
[(Column B) (right) (filled) (bbb) ] {
	stringwidth pop dup PsFF_Cw1 gt {
		/PsFF_Cw1 exch def
	} {
		pop
	} ifelse
} forall
    (Column A) PsFF_Cw0 PsFF_thL
    (Column B) PsFF_Cw1 PsFF_thL
PsFF_nl
    (left) PsFF_Cw0 PsFF_tdL
    (right) PsFF_Cw1 PsFF_tdR
PsFF_nl
    (center) PsFF_Cw0 PsFF_tdC
    (filled) PsFF_Cw1 PsFF_tdL
PsFF_nl
    (aaa) PsFF_Cw0 PsFF_tdC
    (bbb) PsFF_Cw1 PsFF_tdR
PsFF_nl
} [  ] {} ] ` +
			` [ {} [ (And )(this )(is )(after )(the )(table.) ] {} ]  ] `, false, nil},
	} {
		test.opts = append(test.opts, AsPostScript)
		v, err := Render(test.in, test.opts...)
		if err != nil && !test.err {
			t.Errorf("Case %d: unexpected error: %v", i, err)
		} else if err == nil && test.err {
			t.Errorf("Case %d: error expected but not found", i)
		} else if test.out != v {
			t.Errorf("Case %d: %v -> %v but expected %v", i, test.in, v, test.out)
			rActual := []rune(v)
			rExpected := []rune(test.out)
			for i := 0; i < len(rActual) && i < len(rExpected); i++ {
				l := min(len(rActual), i+10)
				e := min(len(rExpected), i+10)

				if rActual[i] != rExpected[i] {
					t.Errorf("  at position %d/%d (actual \"%s\" vs. expected \"%s\")", i, len(v), string(rActual[i:l]), string(rExpected[i:e]))
					break
				}
			}
		}
	}
}

// @[00]@| Go-GMA 5.9.1
// @[01]@|
// @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved.
// @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[14]@| License as described in the accompanying LICENSE file distributed
// @[15]@| with GMA.
// @[16]@|
// @[20]@| Redistribution and use in source and binary forms, with or without
// @[21]@| modification, are permitted provided that the following conditions
// @[22]@| are met:
// @[23]@| 1. Redistributions of source code must retain the above copyright
// @[24]@|    notice, this list of conditions and the following disclaimer.
// @[25]@| 2. Redistributions in binary form must reproduce the above copy-
// @[26]@|    right notice, this list of conditions and the following dis-
// @[27]@|    claimer in the documentation and/or other materials provided
// @[28]@|    with the distribution.
// @[29]@| 3. Neither the name of the copyright holder nor the names of its
// @[30]@|    contributors may be used to endorse or promote products derived
// @[31]@|    from this software without specific prior written permission.
// @[32]@|
// @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
// @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
// @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
// @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
// @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
// @[45]@| SUCH DAMAGE.
// @[46]@|
// @[50]@| This software is not intended for any use or application in which
// @[51]@| the safety of lives or property would be at risk due to failure or
// @[52]@| defect of the software.
//
