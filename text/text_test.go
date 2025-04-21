/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  ______      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )/ ___  \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |\/   )  )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )    /  /    | | /   | ___ #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /    /  /     | (/ /) |(___ #
# | | \_  )| |   | || (   ) |                 ) )    /   _/    /  /      |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\ /  /     _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \_/     (_)(_______)     #
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
		{"foo\\\\bar", "foo\nbar", false, nil},
		{"foo\\\\\nbar", "foo\nbar", false, nil},
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
@Item One
@Item //Tw//o
@Item Three
  * this is not a\\bullet list
@But this is
@@and a sub-list
@@@ and sub-sub-list

And this should start a new list:
@Not that you can tell with bullets.
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
@Item One
@Item //Tw//o
@Item Three
  * this is not a\\bullet list
@But this is
@@and a sub-list
@@@ and sub-sub-list
@@@@But this is
@@@@@and a sub-list

And this should start a new list:
@Not that you can tell with bullets.
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
|: The table title |
|:: The table footnote |
|:: Another footnote\\with a line break |
And this is after the table.`, `Table test:
 THE TABLE TITLE 
┌──────────┬──────────┐
│ COLUMN A │ COLUMN B │
├──────────┼──────────┤
│ left     │    right │
│  center  │ filled   │
│   aaa    │      bbb │
└──────────┴──────────┘
 The table footnote 
 Another footnote
with a line break 
And this is after the table.`, false, nil},
		{`Table test:
|=Column A|=Column B|=Column C|
|left     |    right| some other |
|  center |filled and more stuff too |-
|  aaa and |-  |   bbb
And this is after the table.`, `Table test:
┌──────────┬────────────┬──────────────┐
│ COLUMN A │  COLUMN B  │   COLUMN C   │
├──────────┼────────────┼──────────────┤
│ left     │      right │  some other  │
│  center  │ filled and more stuff too │
│        aaa and        │          bbb │
└──────────┴────────────┴──────────────┘
And this is after the table.`, false, nil},
		{`Titles:
==[Main Title]==
Some text
==(Subtitle)==
Some more text`, `Titles: 
══╣ Main Title ╠════════════════════════════════════════════════════════════════
Some text 
──┤ Subtitle ├──────────────────────────────────────────────────────────────────
Some more text`,
			false, nil},
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
@Item One
@Item //Tw//o
@Item Three
 * this is not a\\bullet list
@But this is
@@and a sub-list
@@@ and sub-sub-list

And this should start a new list:
@Not that you can tell with bullets.
`, "<P>This is a bullet list:<UL><LI>Item One<LI>Item <I>Tw</I>o<LI>Item Three * this is not a<BR/>bullet list<LI>But this is<UL><LI>and a sub-list<UL><LI>and sub-sub-list</UL></UL></UL></P><P>And this should start a new list:<UL><LI>Not that you can tell with bullets.</UL></P>", false, nil},
		{`This is a bullet list:
@Item One
@Item //Tw//o
@Item Three
 * this is not a\\bullet list
@But this is
@@and a sub-list
@@@ and sub-sub-list
@@@@ four

And this should start a new list:
@Not that you can tell with bullets.
`, "<P>This is a bullet list:<UL style='list-style-type:disc;'><LI>Item One<LI>Item <I>Tw</I>o<LI>Item Three * this is not a<BR/>bullet list<LI>But this is<UL style='list-style-type:\"\\2023\";'><LI>and a sub-list<UL style='list-style-type:\"\\0025e6\";'><LI>and sub-sub-list<UL style='list-style-type:disc;'><LI>four</UL></UL></UL></UL></P><P>And this should start a new list:<UL style='list-style-type:disc;'><LI>Not that you can tell with bullets.</UL></P>", false, []func(*renderOptSet){WithBullets('•', '‣', '◦')}},
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
|: The table title |
|:: The table footnote |
|:: Another table footnote\\with a line break. |
And this is after the table.`, `<P>Table test:<TABLE BORDER=1><CAPTION> The table title </CAPTION>
<THEAD>
<TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH></TR></THEAD><TBODY>
<TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD></TR>
<TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT>filled</TD></TR>
<TR><TD ALIGN=CENTER>aaa</TD><TD ALIGN=RIGHT>bbb</TD></TR></TBODY><TFOOT><TR><TD COLSPAN=2> The table footnote </TD></TR><TR><TD COLSPAN=2> Another table footnote<BR/>with a line break. </TD></TR></TFOOT></TABLE>And this is after the table.</P>`, false, nil},
		{`Table test:
|=Column A|=Column B|=Column C|
|left     |    right| some stuff |
|  center |filled and extended to the other |-
|  aaa and |-   |   bbb
And this is after the table.`, `<P>Table test:<TABLE BORDER=1><THEAD>
<TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH><TH ALIGN=LEFT>Column C</TH></TR></THEAD><TBODY>
<TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD><TD ALIGN=CENTER>some stuff</TD></TR>
<TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT COLSPAN=2>filled and extended to the other</TD></TR>
<TR><TD ALIGN=CENTER COLSPAN=2>aaa and</TD><TD ALIGN=RIGHT>bbb</TD></TR></TBODY><TFOOT></TFOOT></TABLE>And this is after the table.</P>`, false, nil},
		{`Titles:
==[Main Title]==
Some text
==(Subtitle)==
Some more text`, `<P>Titles: <H1>Main Title</H1>Some text <H2>Subtitle</H2>Some more text</P>`, false, nil},
		{`\e`, `<P>\</P>`, false, nil},
		{`\v`, `<P>|</P>`, false, nil},
		{`\e\e`, `<P>\\</P>`, false, nil},
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
		{"foo", " [  [ {} [ (foo) ] {PsFF_rm } ]  ] ", false, nil},
		{"\\e", " [  [ {} [ (\\\\) ] {PsFF_rm } ]  ] ", false, nil},
		{"\\v", " [  [ {} [ (|) ] {PsFF_rm } ]  ] ", false, nil},
		{"\\e\\e", " [  [ {} [ (\\\\)(\\\\) ] {PsFF_rm } ]  ] ", false, nil},
		{"", " [  ] ", false, nil},
		{"foo\nbar", " [  [ {} [ (foo )(bar) ] {PsFF_rm } ]  ] ", false, nil},
		{"foo\n\nbar", " [  [ {PsFF_par} [ (foo) ] {PsFF_rm } ]  [ {} [ (bar) ] {} ]  ] ", false, nil},
		{"\n\nfoo", " [  [ {} [ (foo) ] {PsFF_rm } ]  ] ", false, nil},
		{`foo\\bar`, " [  [ {PsFF_nl} [ (foo) ] {PsFF_rm } ]  [ {} [ (bar) ] {} ]  ] ", false, nil},
		{"aa//bb//cc", " [  [ {} [ (aa) ] {PsFF_rm } ]  [ {} [ (bb) ] {PsFF_it } ]  [ {} [ (cc) ] {PsFF_rm } ]  ] ", false, nil},
		{"aa**bb", " [  [ {} [ (aa) ] {PsFF_rm } ]  [ {} [ (bb) ] {PsFF_bf } ]  ] ", false, nil},
		{"a//b**c**d//e", " [  [ {} [ (a) ] {PsFF_rm } ]  [ {} [ (b) ] {PsFF_it } ]  [ {} [ (c) ] {PsFF_bi } ]  [ {} [ (d) ] {PsFF_it } ]  [ {} [ (e) ] {PsFF_rm } ]  ] ", false, nil},
		{"a //b **c //d **e", " [  [ {} [ (a ) ] {PsFF_rm } ]  [ {} [ (b ) ] {PsFF_it } ]  [ {} [ (c ) ] {PsFF_bi } ]  [ {} [ (d ) ] {PsFF_bf } ]  [ {} [ (e) ] {PsFF_rm } ]  ] ", false, nil},
		{"a[[b]]d", " [  [ {} [ (a) ] {PsFF_rm } ]  [ {} [ (b) ] {PsFF_it } ]  [ {} [ (d) ] {PsFF_rm } ]  ] ", false, nil},
		{"a[[b|c]]d", " [  [ {} [ (a) ] {PsFF_rm } ]  [ {} [ (c) ] {PsFF_it } ]  [ {} [ (d) ] {PsFF_rm } ]  ] ", false, nil},
		{`a //it b
c\\de//f`, " [  [ {} [ (a ) ] {PsFF_rm } ]  [ {PsFF_nl} [ (it )(b )(c) ] {PsFF_it } ]  [ {} [ (de) ] {} ]  [ {} [ (f) ] {PsFF_rm } ]  ] ",
			false, nil},
		{`a //it b
c

de//f`, " [  [ {} [ (a ) ] {PsFF_rm } ]  [ {} [ (it )(b )(c) ] {PsFF_it } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {} [ (de) ] {} ]  [ {} [ (f) ] {PsFF_it } ]  ] ", false, nil},
		{`This is a bullet list:
@Item One
@Item //Tw//o
@Item Three
 * this is not a\\bullet list
@But this is
@@and a sub-list
@@@ and sub-sub-list

And this should start a new list:
@Not that you can tell with bullets.
`, ` [  [ {PsFF_nl} [ (This )(is )(a )(bullet )(list:) ] {PsFF_rm } ]  [ { 1 PsFF_ind } [ (\267) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (Item )(One) ] {} ]  [ { 1 PsFF_ind } [ (\267) ] { 0 PsFF_ind } ]  [ {} [ (Item ) ] {} ]  [ {} [ (Tw) ] {PsFF_it } ]  [ {PsFF_nl} [ (o) ] {PsFF_rm } ]  [ { 1 PsFF_ind } [ (\267) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (Item )(Three )(* )(this )(is )(not )(a) ] {} ]  [ {PsFF_nl} [ (bullet )(list) ] {} ]  [ { 1 PsFF_ind } [ (\267) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (But )(this )(is) ] {} ]  [ { 2 PsFF_ind } [ (\267) ] { 1 PsFF_ind } ]  [ {PsFF_nl} [ (and )(a )(sub-list) ] {} ]  [ { 3 PsFF_ind } [ (\267) ] { 2 PsFF_ind } ]  [ {PsFF_par 0 PsFF_ind} [ (and )(sub-sub-list) ] {} ]  [ {PsFF_nl} [ (And )(this )(should )(start )(a )(new )(list:) ] {} ]  [ { 1 PsFF_ind } [ (\267) ] { 0 PsFF_ind } ]  [ {} [ (Not )(that )(you )(can )(tell )(with )(bullets.) ] {} ]  ] `,
			false, nil},
		{`
This is a numbered list:
#Item One
#Item //Tw//o
#Item Three
 # this is not a\\bullet list
#But this is
##and a sub-list
### and sub-sub-list

And this should start a new list:
#and this should be re-sequenced.
`, ` [  [ {PsFF_nl} [ (This )(is )(a )(numbered )(list:) ] {PsFF_rm } ]  [ { 1 PsFF_ind } [ (1.) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (Item )(One) ] {} ]  [ { 1 PsFF_ind } [ (2.) ] { 0 PsFF_ind } ]  [ {} [ (Item ) ] {} ]  [ {} [ (Tw) ] {PsFF_it } ]  [ {PsFF_nl} [ (o) ] {PsFF_rm } ]  [ { 1 PsFF_ind } [ (3.) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (Item )(Three )(# )(this )(is )(not )(a) ] {} ]  [ {PsFF_nl} [ (bullet )(list) ] {} ]  [ { 1 PsFF_ind } [ (4.) ] { 0 PsFF_ind } ]  [ {PsFF_nl} [ (But )(this )(is) ] {} ]  [ { 2 PsFF_ind } [ (a.) ] { 1 PsFF_ind } ]  [ {PsFF_nl} [ (and )(a )(sub-list) ] {} ]  [ { 3 PsFF_ind } [ (i.) ] { 2 PsFF_ind } ]  [ {PsFF_par 0 PsFF_ind} [ (and )(sub-sub-list) ] {} ]  [ {PsFF_nl} [ (And )(this )(should )(start )(a )(new )(list:) ] {} ]  [ { 1 PsFF_ind } [ (1.) ] { 0 PsFF_ind } ]  [ {} [ (and )(this )(should )(be )(re-sequenced.) ] {} ]  ] `,
			false, nil},
		{`Table test:
|=Column A|=Column B|
|left     |    right|
|  center |filled|
|  aaa    |   bbb
|: The table caption |
|:: The table footer |
|:: Another table footer\\\\with a line break.|
And this is after the table.`, ` [  [ {PsFF_nl} [ (Table )(test:) ] {PsFF_rm } ]  [ {PsFF_par} [  ] {} ]  [ {} [ ( )(The )(table )(caption ) ] {} ]  [ {PsFF_nl PsFF_rm} [  ] {PsFF_bi PsFF_tbl_caption} ]  [ {PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
/PsFF_CLw0c0l0 0 def
/PsFF_CLw0c1l0 0 def
/PsFF_CLw1c0l0 0 def
/PsFF_CLw1c1l0 0 def
/PsFF_CLw2c0l0 0 def
/PsFF_CLw2c1l0 0 def
/PsFF_CLw3c0l0 0 def
/PsFF_CLw3c1l0 0 def

% Column #0 of 2
/PsFF_Cw0 0 def
PsFF_rm /PsFF_Cwi 0 def [(Column A)] {stringwidth pop dup dup (PsFF_CLw0c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(left)] {stringwidth pop dup dup (PsFF_CLw1c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(center)] {stringwidth pop dup dup (PsFF_CLw2c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(aaa)] {stringwidth pop dup dup (PsFF_CLw3c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall

% Column #1 of 2
/PsFF_Cw1 0 def
PsFF_rm /PsFF_Cwi 0 def [(Column B)] {stringwidth pop dup dup (PsFF_CLw0c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(right)] {stringwidth pop dup dup (PsFF_CLw1c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(filled)] {stringwidth pop dup dup (PsFF_CLw2c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(bbb)] {stringwidth pop dup dup (PsFF_CLw3c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
%
% Now adjust column widths for the spans
%
%
% Table contents
%
% Row 0 (1 line)
% Row 0, Col 0, Line 0
PsFF_CLw0c0l0 true false false PsFF_Cw0  PsFF_cellF PsFF_rm (Column A) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 1, Line 0
PsFF_CLw0c1l0 true false false PsFF_Cw1  PsFF_cellF PsFF_rm (Column B) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 1 (1 line)
% Row 1, Col 0, Line 0
PsFF_CLw1c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm (left) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 1, Line 0
PsFF_CLw1c1l0 false false true PsFF_Cw1  PsFF_cellF PsFF_rm (right) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 2 (1 line)
% Row 2, Col 0, Line 0
PsFF_CLw2c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (center) PsFF_cellfragment  PsFF_tEnd
% Row 2, Col 1, Line 0
PsFF_CLw2c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (filled) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 3 (1 line)
% Row 3, Col 0, Line 0
PsFF_CLw3c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (aaa) PsFF_cellfragment  PsFF_tEnd
% Row 3, Col 1, Line 0
PsFF_CLw3c1l0 false false true PsFF_Cw1  PsFF_cellF PsFF_rm (bbb) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
/X PsFF_Xsave def} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ ( )(The )(table )(footer ) ] {PsFF_tbl_footer} ]  [ {PsFF_nl} [ ( )(Another )(table )(footer) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (with )(a )(line )(break.) ] {PsFF_tbl_footer} ]  [ {} [ (And )(this )(is )(after )(the )(table.) ] {} ]  ] `, false, nil},
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
					/*
						for ii := 0; ii <= i; ii++ {
							fmt.Printf("Actual[%d]=%s, expected[%d]=%s\n", ii, string(rActual[ii]), ii, string(rExpected[ii]))
						}
					*/
					t.Errorf("  at position %d/%d (actual \"%s\" vs. expected \"%s\")", i, len(v), string(rActual[i:l]), string(rExpected[i:e]))
					break
				}
			}
		}
	}
}

func TestCompatibility(t *testing.T) {
	type testcase struct {
		in   string
		out  string
		err  bool
		opts []func(*renderOptSet)
	}

	for i, test := range []testcase{
		// testcase 0: plain text
		{MarkupSyntax, ` [  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {PsFF_section} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (GMA )(Text )(Markup )(Syntax) ] {} ]  [ {PsFF_par} [  ] {} ]  [ {PsFF_par} [ (The )(markup )(syntax )(is )(simple. )(Lines )(are )(collected )(together )(into )(a )(single )(logical )(line )(which )(is )(then )(wrapped )(as )(appropriate )(to )(the )(output )(format )(\(which )(may )(rely )(on )(whatever )(is )(printing )(the )(output )(to )(break )(lines )(as )(it )(prefers\).) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_subsection} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (Basic )(Markup) ] {} ]  [ {PsFF_par} [ (A )(blank )(line )(marks )(a )(paragraph )(break.) ] {} ]  [ {} [ (\\)(\\) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(marks )(a )(line )(break.) ] {PsFF_rm } ]  [ {} [ (A )(literal ) ] {} ]  [ {} [ (\\) ] {PsFF_bf } ]  [ {} [ ( )(or ) ] {PsFF_rm } ]  [ {} [ (|) ] {PsFF_bf } ]  [ {} [ ( )(character )(may )(be )(entered )(without )(being )(interpreted )(as )(part )(of )(markup )(syntax )(using )(the )(codes ) ] {PsFF_rm } ]  [ {} [ (\\)(e) ] {PsFF_bf } ]  [ {} [ ( )(and ) ] {PsFF_rm } ]  [ {} [ (\\)(v) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(respectively.) ] {PsFF_rm } ]  [ {} [ (/)(/text/)(/) ] {PsFF_bf } ]  [ {PsFF_nl} [ ( )(sets )("text" )(in )(Italics*\262) ] {PsFF_rm } ]  [ {} [ (*)(*text)(*) ] {PsFF_bf } ]  [ {PsFF_par} [ (* )(sets )("text" )(in )(boldface*\262) ] {PsFF_rm } ]  [ {} [ (@) ] {PsFF_bf } ]  [ {} [ (blah) ] {PsFF_it } ]  [ {PsFF_nl} [ (... )(Starts )(bulleted )(list )(item\263) ] {PsFF_rm } ]  [ {} [ (@@) ] {PsFF_bf } ]  [ {} [ (blah) ] {PsFF_it } ]  [ {PsFF_nl} [ (... )(Starts )(level-2 )(bulleted )(list )(item\263) ] {PsFF_rm } ]  [ {} [ (@@@) ] {PsFF_bf } ]  [ {} [ (blah) ] {PsFF_it } ]  [ {PsFF_par} [ (... )(Starts )(level-3 )(bulleted )(list )(item )(\(and )(so )(forth\)\263) ] {PsFF_rm } ]  [ {} [ (#) ] {PsFF_bf } ]  [ {} [ (blah) ] {PsFF_it } ]  [ {PsFF_nl} [ (... )(Starts )(enumerated )(list )(item\263) ] {PsFF_rm } ]  [ {} [ (##) ] {PsFF_bf } ]  [ {} [ (blah) ] {PsFF_it } ]  [ {PsFF_par} [ (... )(...and )(so )(forth\263) ] {PsFF_rm } ]  [ {} [ ([)([) ] {PsFF_bf } ]  [ {} [ (name) ] {PsFF_it } ]  [ {} [ (])(]) ] {PsFF_bf } ]  [ {} [ ( )(Creates )(a )(hyperlink )(to )(") ] {PsFF_rm } ]  [ {} [ (name) ] {PsFF_it } ]  [ {PsFF_nl} [ (" )(where )(this )(name )(itself )(adequately )(identifies )(the )(linked-to )(element )(in )(GMA )(\(e.g., )(the )(name )(of )(a )(spell\).) ] {PsFF_rm } ]  [ {} [ ([)([) ] {PsFF_bf } ]  [ {} [ (link) ] {PsFF_it } ]  [ {} [ (|) ] {PsFF_bf } ]  [ {} [ (name) ] {PsFF_it } ]  [ {} [ (])(]) ] {PsFF_bf } ]  [ {} [ ( )(Creates )(a )(hyperlink )(called )(") ] {PsFF_rm } ]  [ {} [ (name) ] {PsFF_it } ]  [ {} [ (" )(which )(links )(to )(GMA )(element )(") ] {PsFF_rm } ]  [ {} [ (link) ] {PsFF_it } ]  [ {PsFF_par} [ (".) ] {PsFF_rm } ]  [ {} [ (\\)(.) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(does )(nothing )(but )(serves )(to )(disambiguate )(things )(or )(prevent )(otherwise )(special )(symbols )(from )(being )(interpreted )(as )(markup )(syntax.) ] {PsFF_rm } ]  [ {} [ (There )(is )(also )(a )(special )(page-break )(marker ) ] {} ]  [ {} [ (<)(<-)(->)(>) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(which )(is )(not )(actually )(processed )(by )(this )(package, )(but )(some )(output )(subsystems )(recognize )(it )(when )(they )(see )(it )(in )(the )(output )(\(e.g., )(PostScript )(formatted )(text )(blocks\).) ] {PsFF_rm } ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_subsection} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (Special )(Characters) ] {} ]  [ {PsFF_par} [ (Many )(common )(Unicode )(characters )(are )(recognized )(on )(input, )(but )(support )(for )(them )(in )(the )(target )(output )(format )(is )(not )(guaranteed.) ] {} ]  [ {PsFF_nl} [ (The )(following )(markup )(symbols )(may )(also )(be )(used )(to )(represent )(special )(characters:) ] {} ]  [ {PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
/PsFF_CLw0c0l0 0 def
/PsFF_CLw0c1l0 0 def
/PsFF_CLw0c2l0 0 def
/PsFF_CLw0c3l0 0 def
/PsFF_CLw1c0l0 0 def
/PsFF_CLw1c1l0 0 def
/PsFF_CLw1c2l0 0 def
/PsFF_CLw1c3l0 0 def
/PsFF_CLw2c0l0 0 def
/PsFF_CLw2c1l0 0 def
/PsFF_CLw2c2l0 0 def
/PsFF_CLw2c3l0 0 def
/PsFF_CLw3c0l0 0 def
/PsFF_CLw3c1l0 0 def
/PsFF_CLw3c2l0 0 def
/PsFF_CLw3c3l0 0 def
/PsFF_CLw4c0l0 0 def
/PsFF_CLw4c1l0 0 def
/PsFF_CLw4c2l0 0 def
/PsFF_CLw4c3l0 0 def
/PsFF_CLw5c0l0 0 def
/PsFF_CLw5c1l0 0 def
/PsFF_CLw5c2l0 0 def
/PsFF_CLw5c3l0 0 def
/PsFF_CLw6c0l0 0 def
/PsFF_CLw6c1l0 0 def
/PsFF_CLw6c2l0 0 def
/PsFF_CLw6c3l0 0 def
/PsFF_CLw7c0l0 0 def
/PsFF_CLw7c1l0 0 def
/PsFF_CLw7c2l0 0 def
/PsFF_CLw7c3l0 0 def
/PsFF_CLw8c0l0 0 def
/PsFF_CLw8c1l0 0 def
/PsFF_CLw8c2l0 0 def
/PsFF_CLw8c3l0 0 def
/PsFF_CLw9c0l0 0 def
/PsFF_CLw9c1l0 0 def
/PsFF_CLw9c2l0 0 def
/PsFF_CLw9c3l0 0 def
/PsFF_CLw10c0l0 0 def
/PsFF_CLw10c1l0 0 def
/PsFF_CLw10c2l0 0 def
/PsFF_CLw10c3l0 0 def
/PsFF_CLw11c0l0 0 def
/PsFF_CLw11c1l0 0 def
/PsFF_CLw11c2l0 0 def
/PsFF_CLw11c3l0 0 def
/PsFF_CLw12c0l0 0 def
/PsFF_CLw12c1l0 0 def
/PsFF_CLw12c2l0 0 def
/PsFF_CLw12c3l0 0 def
/PsFF_CLw13c0l0 0 def
/PsFF_CLw13c1l0 0 def
/PsFF_CLw13c2l0 0 def
/PsFF_CLw13c3l0 0 def
/PsFF_CLw14c0l0 0 def
/PsFF_CLw14c1l0 0 def
/PsFF_CLw14c2l0 0 def
/PsFF_CLw14c3l0 0 def
/PsFF_CLw15c0l0 0 def
/PsFF_CLw15c1l0 0 def
/PsFF_CLw15c2l0 0 def
/PsFF_CLw15c3l0 0 def
/PsFF_CLw16c0l0 0 def
/PsFF_CLw16c1l0 0 def
/PsFF_CLw16c2l0 0 def
/PsFF_CLw16c3l0 0 def

% Column #0 of 4
/PsFF_Cw0 0 def
PsFF_bf /PsFF_Cwi 0 def [([S])] {stringwidth pop dup dup (PsFF_CLw0c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([c])] {stringwidth pop dup dup (PsFF_CLw1c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([<<])] {stringwidth pop dup dup (PsFF_CLw2c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([R])] {stringwidth pop dup dup (PsFF_CLw3c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(^o)] {stringwidth pop dup dup (PsFF_CLw4c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(+-)] {stringwidth pop dup dup (PsFF_CLw5c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(^.)] {stringwidth pop dup dup (PsFF_CLw6c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([0])] {stringwidth pop dup dup (PsFF_CLw7c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([1])] {stringwidth pop dup dup (PsFF_CLw8c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([2])] {stringwidth pop dup dup (PsFF_CLw9c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([3])] {stringwidth pop dup dup (PsFF_CLw10c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([4])] {stringwidth pop dup dup (PsFF_CLw11c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([5])] {stringwidth pop dup dup (PsFF_CLw12c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([6])] {stringwidth pop dup dup (PsFF_CLw13c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([7])] {stringwidth pop dup dup (PsFF_CLw14c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([8])] {stringwidth pop dup dup (PsFF_CLw15c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([9])] {stringwidth pop dup dup (PsFF_CLw16c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall

% Column #1 of 4
/PsFF_Cw1 0 def
PsFF_rm /PsFF_Cwi 0 def [(\247 section)] {stringwidth pop dup dup (PsFF_CLw0c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\345 copyright)] {stringwidth pop dup dup (PsFF_CLw1c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\253 << quotes)] {stringwidth pop dup dup (PsFF_CLw2c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\346 registered)] {stringwidth pop dup dup (PsFF_CLw3c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\347 degrees)] {stringwidth pop dup dup (PsFF_CLw4c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\354 plusminus)] {stringwidth pop dup dup (PsFF_CLw5c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\267 bullet)] {stringwidth pop dup dup (PsFF_CLw6c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\330 superscript 0)] {stringwidth pop dup dup (PsFF_CLw7c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\331 superscript 1)] {stringwidth pop dup dup (PsFF_CLw8c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\332 superscript 2)] {stringwidth pop dup dup (PsFF_CLw9c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\333 superscript 3)] {stringwidth pop dup dup (PsFF_CLw10c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\334 superscript 4)] {stringwidth pop dup dup (PsFF_CLw11c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\335 superscript 5)] {stringwidth pop dup dup (PsFF_CLw12c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\336 superscript 6)] {stringwidth pop dup dup (PsFF_CLw13c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\337 superscript 7)] {stringwidth pop dup dup (PsFF_CLw14c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\340 superscript 8)] {stringwidth pop dup dup (PsFF_CLw15c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\342 superscript 9)] {stringwidth pop dup dup (PsFF_CLw16c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall

% Column #2 of 4
/PsFF_Cw2 0 def
PsFF_bf /PsFF_Cwi 0 def [(AE)] {stringwidth pop dup dup (PsFF_CLw0c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(ae)] {stringwidth pop dup dup (PsFF_CLw1c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([>>])] {stringwidth pop dup dup (PsFF_CLw2c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(1/4)] {stringwidth pop dup dup (PsFF_CLw3c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(1/2)] {stringwidth pop dup dup (PsFF_CLw4c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(3/4)] {stringwidth pop dup dup (PsFF_CLw5c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([x])] {stringwidth pop dup dup (PsFF_CLw6c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([/])] {stringwidth pop dup dup (PsFF_CLw7c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(-)] {stringwidth pop dup dup (PsFF_CLw8c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(--)] {stringwidth pop dup dup (PsFF_CLw9c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(---)] {stringwidth pop dup dup (PsFF_CLw10c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(')] {stringwidth pop dup dup (PsFF_CLw11c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(` + "`" + `)] {stringwidth pop dup dup (PsFF_CLw12c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [('')] {stringwidth pop dup dup (PsFF_CLw13c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [(` + "``" + `)] {stringwidth pop dup dup (PsFF_CLw14c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([+])] {stringwidth pop dup dup (PsFF_CLw15c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_bf /PsFF_Cwi 0 def [([++])] {stringwidth pop dup dup (PsFF_CLw16c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall

% Column #3 of 4
/PsFF_Cw3 0 def
PsFF_rm /PsFF_Cwi 0 def [(\341 ligature)] {stringwidth pop dup dup (PsFF_CLw0c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\361 ligature)] {stringwidth pop dup dup (PsFF_CLw1c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\273 >> quotes)] {stringwidth pop dup dup (PsFF_CLw2c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\355)] {stringwidth pop dup dup (PsFF_CLw3c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\356)] {stringwidth pop dup dup (PsFF_CLw4c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\357)] {stringwidth pop dup dup (PsFF_CLw5c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\360 multiplication)] {stringwidth pop dup dup (PsFF_CLw6c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\344 division)] {stringwidth pop dup dup (PsFF_CLw7c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(- hyphen or minus sign)] {stringwidth pop dup dup (PsFF_CLw8c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\261 en dash)] {stringwidth pop dup dup (PsFF_CLw9c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\320 em dash)] {stringwidth pop dup dup (PsFF_CLw10c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\140 open single quote)] {stringwidth pop dup dup (PsFF_CLw11c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(' close single quote)] {stringwidth pop dup dup (PsFF_CLw12c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\252 open double quote)] {stringwidth pop dup dup (PsFF_CLw13c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [('' close double quote)] {stringwidth pop dup dup (PsFF_CLw14c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\262 dagger)] {stringwidth pop dup dup (PsFF_CLw15c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(\263 double dagger)] {stringwidth pop dup dup (PsFF_CLw16c3l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw3 gt {/PsFF_Cw3 exch def} {pop} ifelse} forall
%
% Now adjust column widths for the spans
%
%
% Table contents
%
% Row 0 (1 line)
% Row 0, Col 0, Line 0
PsFF_CLw0c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (S]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 0, Col 1, Line 0
PsFF_CLw0c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\247 section) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 2, Line 0
PsFF_CLw0c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (A) PsFF_cellfragment (E) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 0, Col 3, Line 0
PsFF_CLw0c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\341 ligature) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 1 (1 line)
% Row 1, Col 0, Line 0
PsFF_CLw1c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (c]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 1, Col 1, Line 0
PsFF_CLw1c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\345 copyright) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 2, Line 0
PsFF_CLw1c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (a) PsFF_cellfragment (e) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 1, Col 3, Line 0
PsFF_CLw1c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\361 ligature) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 2 (1 line)
% Row 2, Col 0, Line 0
PsFF_CLw2c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (<<]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 2, Col 1, Line 0
PsFF_CLw2c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\253 << quotes) PsFF_cellfragment  PsFF_tEnd
% Row 2, Col 2, Line 0
PsFF_CLw2c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (>>]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 2, Col 3, Line 0
PsFF_CLw2c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\273 >> quotes) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 3 (1 line)
% Row 3, Col 0, Line 0
PsFF_CLw3c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (R]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 3, Col 1, Line 0
PsFF_CLw3c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\346 registered) PsFF_cellfragment  PsFF_tEnd
% Row 3, Col 2, Line 0
PsFF_CLw3c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (1) PsFF_cellfragment (/4) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 3, Col 3, Line 0
PsFF_CLw3c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\355) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 4 (1 line)
% Row 4, Col 0, Line 0
PsFF_CLw4c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf (^) PsFF_cellfragment (o) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 4, Col 1, Line 0
PsFF_CLw4c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\347 degrees) PsFF_cellfragment  PsFF_tEnd
% Row 4, Col 2, Line 0
PsFF_CLw4c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (1) PsFF_cellfragment (/2) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 4, Col 3, Line 0
PsFF_CLw4c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\356) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 5 (1 line)
% Row 5, Col 0, Line 0
PsFF_CLw5c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf (+) PsFF_cellfragment (-) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 5, Col 1, Line 0
PsFF_CLw5c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\354 plusminus) PsFF_cellfragment  PsFF_tEnd
% Row 5, Col 2, Line 0
PsFF_CLw5c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (3) PsFF_cellfragment (/4) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 5, Col 3, Line 0
PsFF_CLw5c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\357) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 6 (1 line)
% Row 6, Col 0, Line 0
PsFF_CLw6c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf (^) PsFF_cellfragment (.) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 6, Col 1, Line 0
PsFF_CLw6c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\267 bullet) PsFF_cellfragment  PsFF_tEnd
% Row 6, Col 2, Line 0
PsFF_CLw6c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (x]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 6, Col 3, Line 0
PsFF_CLw6c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\360 multiplication) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 7 (1 line)
% Row 7, Col 0, Line 0
PsFF_CLw7c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (0]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 7, Col 1, Line 0
PsFF_CLw7c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\330 superscript 0) PsFF_cellfragment  PsFF_tEnd
% Row 7, Col 2, Line 0
PsFF_CLw7c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (/]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 7, Col 3, Line 0
PsFF_CLw7c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\344 division) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 8 (1 line)
% Row 8, Col 0, Line 0
PsFF_CLw8c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (1]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 8, Col 1, Line 0
PsFF_CLw8c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\331 superscript 1) PsFF_cellfragment  PsFF_tEnd
% Row 8, Col 2, Line 0
PsFF_CLw8c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (-) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 8, Col 3, Line 0
PsFF_CLw8c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (- hyphen or minus sign) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 9 (1 line)
% Row 9, Col 0, Line 0
PsFF_CLw9c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (2]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 9, Col 1, Line 0
PsFF_CLw9c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\332 superscript 2) PsFF_cellfragment  PsFF_tEnd
% Row 9, Col 2, Line 0
PsFF_CLw9c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (-) PsFF_cellfragment (-) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 9, Col 3, Line 0
PsFF_CLw9c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\261 en dash) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 10 (1 line)
% Row 10, Col 0, Line 0
PsFF_CLw10c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (3]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 10, Col 1, Line 0
PsFF_CLw10c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\333 superscript 3) PsFF_cellfragment  PsFF_tEnd
% Row 10, Col 2, Line 0
PsFF_CLw10c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (-) PsFF_cellfragment (-) PsFF_cellfragment (-) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 10, Col 3, Line 0
PsFF_CLw10c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\320 em dash) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 11 (1 line)
% Row 11, Col 0, Line 0
PsFF_CLw11c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (4]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 11, Col 1, Line 0
PsFF_CLw11c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\334 superscript 4) PsFF_cellfragment  PsFF_tEnd
% Row 11, Col 2, Line 0
PsFF_CLw11c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (') PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 11, Col 3, Line 0
PsFF_CLw11c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\140 open single quote) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 12 (1 line)
% Row 12, Col 0, Line 0
PsFF_CLw12c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (5]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 12, Col 1, Line 0
PsFF_CLw12c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\335 superscript 5) PsFF_cellfragment  PsFF_tEnd
% Row 12, Col 2, Line 0
PsFF_CLw12c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (` + "`" + `) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 12, Col 3, Line 0
PsFF_CLw12c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (' close single quote) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 13 (1 line)
% Row 13, Col 0, Line 0
PsFF_CLw13c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (6]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 13, Col 1, Line 0
PsFF_CLw13c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\336 superscript 6) PsFF_cellfragment  PsFF_tEnd
% Row 13, Col 2, Line 0
PsFF_CLw13c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ('') PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 13, Col 3, Line 0
PsFF_CLw13c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\252 open double quote) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 14 (1 line)
% Row 14, Col 0, Line 0
PsFF_CLw14c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (7]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 14, Col 1, Line 0
PsFF_CLw14c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\337 superscript 7) PsFF_cellfragment  PsFF_tEnd
% Row 14, Col 2, Line 0
PsFF_CLw14c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf (` + "``" + `) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 14, Col 3, Line 0
PsFF_CLw14c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm ('' close double quote) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 15 (1 line)
% Row 15, Col 0, Line 0
PsFF_CLw15c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (8]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 15, Col 1, Line 0
PsFF_CLw15c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\340 superscript 8) PsFF_cellfragment  PsFF_tEnd
% Row 15, Col 2, Line 0
PsFF_CLw15c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (+]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 15, Col 3, Line 0
PsFF_CLw15c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\262 dagger) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 16 (1 line)
% Row 16, Col 0, Line 0
PsFF_CLw16c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (9]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 16, Col 1, Line 0
PsFF_CLw16c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (\342 superscript 9) PsFF_cellfragment  PsFF_tEnd
% Row 16, Col 2, Line 0
PsFF_CLw16c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm PsFF_bf ([) PsFF_cellfragment (++]) PsFF_cellfragment PsFF_rm  PsFF_tEnd
% Row 16, Col 3, Line 0
PsFF_CLw16c3l0 false false false PsFF_Cw3  PsFF_cellF PsFF_rm (\263 double dagger) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
/X PsFF_Xsave def} [  ] {} ]  [ {PsFF_par} [  ] {} ]  [ {} [ (The )(letter ) ] {} ]  [ {} [ (x) ] {PsFF_bf } ]  [ {PsFF_nl} [ ( )(immediately )(next )(to )(a )(digit )(causes )(it )(to )(be )(printed )(as )(a )(multiplication )(sign )(\(e.g., )(\3602 )(or )(3\360\).) ] {PsFF_rm } ]  [ {} [ (A )(hyphen ) ] {} ]  [ {} [ (-) ] {PsFF_bf } ]  [ {PsFF_nl} [ ( )(immediately )(before )(a )(digit )(causes )(it )(to )(be )(printed )(as )(a )(minus )(sign )(instead )(of )(a )(hyphen )(\(e.g., )(\3621\).) ] {PsFF_rm } ]  [ {} [ (Separate )(numbers )(from )(fractions )(with )(an )(underscore )(\(e.g., ) ] {} ]  [ {} [ (12)(_)(1)(/2) ] {PsFF_bf } ]  [ {} [ ( )(prints )(as ) ] {PsFF_rm } ]  [ {} [ (12\356) ] {PsFF_bf } ]  [ {PsFF_par} [ (\).) ] {PsFF_rm } ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_subsection} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (Titles) ] {} ]  [ {PsFF_nl} [ (=)(=[Main )(\(top-level\) )(Heading)(]==) ] {} ]  [ {PsFF_par} [ (=)(=\(Subtitle )(\(2nd-level\))(\)==) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_subsection} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (Tables) ] {} ]  [ {} [ (Tables )(are )(specified )(by )(a )(set )(of )(lines )(beginning )(with )(a ) ] {} ]  [ {} [ (|) ] {PsFF_bf } ]  [ {} [ ( )(character.\263 )(Each )(column )(in )(the )(table )(is )(separated )(from )(the )(others )(with ) ] {PsFF_rm } ]  [ {} [ (|) ] {PsFF_bf } ]  [ {} [ ( )(characters )(as )(well. )(A ) ] {PsFF_rm } ]  [ {} [ (|) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(at )(the )(very )(end )(of )(the )(row )(is )(optional.) ] {PsFF_rm } ]  [ {} [ (|=Size )(Code|=Area|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(S )(| )(5|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(M )(| )(5|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(L )(| )(10|) ] {PsFF_bf } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {PsFF_par} [ (This )(produces )(a )(table )(like) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
/PsFF_CLw0c0l0 0 def
/PsFF_CLw0c1l0 0 def
/PsFF_CLw1c0l0 0 def
/PsFF_CLw1c1l0 0 def
/PsFF_CLw2c0l0 0 def
/PsFF_CLw2c1l0 0 def
/PsFF_CLw3c0l0 0 def
/PsFF_CLw3c1l0 0 def

% Column #0 of 2
/PsFF_Cw0 0 def
PsFF_rm /PsFF_Cwi 0 def [(Size Code)] {stringwidth pop dup dup (PsFF_CLw0c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(S)] {stringwidth pop dup dup (PsFF_CLw1c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(M)] {stringwidth pop dup dup (PsFF_CLw2c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(L)] {stringwidth pop dup dup (PsFF_CLw3c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall

% Column #1 of 2
/PsFF_Cw1 0 def
PsFF_rm /PsFF_Cwi 0 def [(Area)] {stringwidth pop dup dup (PsFF_CLw0c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(5)] {stringwidth pop dup dup (PsFF_CLw1c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(5)] {stringwidth pop dup dup (PsFF_CLw2c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(10)] {stringwidth pop dup dup (PsFF_CLw3c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
%
% Now adjust column widths for the spans
%
%
% Table contents
%
% Row 0 (1 line)
% Row 0, Col 0, Line 0
PsFF_CLw0c0l0 true false false PsFF_Cw0  PsFF_cellF PsFF_rm (Size Code) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 1, Line 0
PsFF_CLw0c1l0 true false false PsFF_Cw1  PsFF_cellF PsFF_rm (Area) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 1 (1 line)
% Row 1, Col 0, Line 0
PsFF_CLw1c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (S) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 1, Line 0
PsFF_CLw1c1l0 false false true PsFF_Cw1  PsFF_cellF PsFF_rm (5) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 2 (1 line)
% Row 2, Col 0, Line 0
PsFF_CLw2c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (M) PsFF_cellfragment  PsFF_tEnd
% Row 2, Col 1, Line 0
PsFF_CLw2c1l0 false false true PsFF_Cw1  PsFF_cellF PsFF_rm (5) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 3 (1 line)
% Row 3, Col 0, Line 0
PsFF_CLw3c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (L) PsFF_cellfragment  PsFF_tEnd
% Row 3, Col 1, Line 0
PsFF_CLw3c1l0 false false true PsFF_Cw1  PsFF_cellF PsFF_rm (10) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
/X PsFF_Xsave def} [  ] {} ]  [ {PsFF_par} [  ] {} ]  [ {} [ (Table )(cells )(beginning )(with ) ] {} ]  [ {} [ (=) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(are )(headers )(\(usually )(placed )(in )(the )(first )(row\)) ] {PsFF_rm } ]  [ {} [ (Cells )(are )(left- )(or )(right-justified )(if )(there )(is )(leading )(or )(trailing )(space )(between )(the ) ] {} ]  [ {} [ (|) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(separators )(for )(that )(cell, )(respectively. )(If )(there )(is )(space )(before )(and )(after )(the )(text, )(it )(is )(centered. )(In )(the )(example )(above, )(the )(size )(codes )(will )(be )(centered )(in )(their )(column )(and )(the )(area )(numbers )(are )(right-justified )(in )(theirs.) ] {PsFF_rm } ]  [ {} [ (Cells )(which )(begin )(with )(a )(hyphen )(\() ] {} ]  [ {} [ (-) ] {PsFF_bf } ]  [ {PsFF_par} [ (\) )(indicate )(that )(the )(cell )(to )(their )(left )(spans )(into )(them. )(For )(example:) ] {PsFF_rm } ]  [ {} [ (|=Column )(A|=Column )(B|=Column )(C) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (|stuff )(|more )(stuff|and )(more) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (|a )(really )(wide )(column|- )(|hello) ] {PsFF_bf } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {PsFF_par} [ (produces:) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
/PsFF_CLw0c0l0 0 def
/PsFF_CLw0c1l0 0 def
/PsFF_CLw0c2l0 0 def
/PsFF_CLw1c0l0 0 def
/PsFF_CLw1c1l0 0 def
/PsFF_CLw1c2l0 0 def
/PsFF_CLw2c0l0 0 def
/PsFF_CLw2c1l0 0 def
/PsFF_CLw2c2l0 0 def

% Column #0 of 3
/PsFF_Cw0 0 def
PsFF_rm /PsFF_Cwi 0 def [(Column A)] {stringwidth pop dup dup (PsFF_CLw0c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(stuff)] {stringwidth pop dup dup (PsFF_CLw1c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall

% Column #1 of 3
/PsFF_Cw1 0 def
PsFF_rm /PsFF_Cwi 0 def [(Column B)] {stringwidth pop dup dup (PsFF_CLw0c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(more stuff)] {stringwidth pop dup dup (PsFF_CLw1c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall

% Column #2 of 3
/PsFF_Cw2 0 def
PsFF_rm /PsFF_Cwi 0 def [(Column C)] {stringwidth pop dup dup (PsFF_CLw0c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(and more)] {stringwidth pop dup dup (PsFF_CLw1c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(hello)] {stringwidth pop dup dup (PsFF_CLw2c2l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw2 gt {/PsFF_Cw2 exch def} {pop} ifelse} forall
%
% Now adjust column widths for the spans
%
% span row 2, columns 0-1:
/PsFF__t__have PsFF_TcolSpn 1 mul PsFF_Cw0 add PsFF_Cw1 add def
/PsFF__t__need 0 def
PsFF_rm /PsFF_Cwi 0 def [(a really wide column)] {stringwidth pop dup dup (PsFF_CLw2c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF__t__need gt {/PsFF__t__need exch def} {pop} ifelse} forall

PsFF__t__need PsFF__t__have gt {
   /PsFF__t__add PsFF__t__need PsFF__t__have sub def
   /PsFF__t__each PsFF__t__add 1 1 add div def
   /PsFF_Cw0 PsFF_Cw0 PsFF__t__each add def
   /PsFF__t__add PsFF__t__add PsFF__t__each sub def
   /PsFF_Cw1 PsFF_Cw1 PsFF__t__each add def
   /PsFF__t__add PsFF__t__add PsFF__t__each sub def
   PsFF__t__add 0 gt {
      /PsFF_Cw0 PsFF_Cw0 PsFF__t__add add def
   } if
} if
%
% Table contents
%
% Row 0 (1 line)
% Row 0, Col 0, Line 0
PsFF_CLw0c0l0 true false false PsFF_Cw0  PsFF_cellF PsFF_rm (Column A) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 1, Line 0
PsFF_CLw0c1l0 true false false PsFF_Cw1  PsFF_cellF PsFF_rm (Column B) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 2, Line 0
PsFF_CLw0c2l0 true false false PsFF_Cw2  PsFF_cellF PsFF_rm (Column C) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 1 (1 line)
% Row 1, Col 0, Line 0
PsFF_CLw1c0l0 false false false PsFF_Cw0  PsFF_cellF PsFF_rm (stuff) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 1, Line 0
PsFF_CLw1c1l0 false false false PsFF_Cw1  PsFF_cellF PsFF_rm (more stuff) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 2, Line 0
PsFF_CLw1c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm (and more) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 2 (1 line)
% Row 2, Col 0, Line 0
PsFF_CLw2c0l0 false false false PsFF_Cw0 PsFF_Cw1 add PsFF_TcolSpn add PsFF_cellF PsFF_rm (a really wide column) PsFF_cellfragment  PsFF_tEnd
% Row 2, Col 1, Line 0
% (spanned through)
% Row 2, Col 2, Line 0
PsFF_CLw2c2l0 false false false PsFF_Cw2  PsFF_cellF PsFF_rm (hello) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
/X PsFF_Xsave def} [  ] {} ]  [ {PsFF_par} [  ] {} ]  [ {PsFF_par} [ (A )(row )(in )(the )(form) ] {} ]  [ {} [ (|: )(Table )(Caption )(|) ] {PsFF_bf } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {PsFF_par} [ (places )(a )(caption )(on )(the )(table )(\(usually )(above )(the )(table\), )(while )(each )(row )(in )(the )(form) ] {} ]  [ {} [ (|:: )(Footnote )(|) ] {PsFF_bf } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {} [ (adds )(a )(footnote )(at )(the )(bottom )(of )(the )(table. )(Footnotes )(may )(contain ) ] {} ]  [ {} [ (\\)(\\) ] {PsFF_bf } ]  [ {PsFF_par} [ ( )(to )(make )(explicit )(line )(breaks )(but )(captions )(cannot. )(Captions )(and )(footnotes )(should )(be )(a )(single )(cell )(per )(line )(regardless )(of )(the )(number )(of )(columns )(the )(table )(has.) ] {PsFF_rm } ]  [ {PsFF_par} [ (So:) ] {} ]  [ {} [ (|= )(Die )(Roll )(|= )(Color )(|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(1 )(| )(blue )(|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(2\2614 )(| )(green )(|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (| )(5+ )(| )(plaid[)(1] )(|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (|:Random )(Colors )(\(d8\)|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (|::[)(1]Or )(reroll.|) ] {PsFF_bf } ]  [ {PsFF_nl} [  ] {PsFF_rm } ]  [ {} [ (|::\(Subject )(to )(GM )(discretion.\)|) ] {PsFF_bf } ]  [ {PsFF_par} [  ] {PsFF_rm } ]  [ {PsFF_par} [ (produces:) ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_par} [  ] {} ]  [ {} [ (Random )(Colors )(\(d8\)) ] {} ]  [ {PsFF_nl PsFF_rm} [  ] {PsFF_bi PsFF_tbl_caption} ]  [ {PsFF_rm
/PsFF_Xsave X def
%
% Start of Data Table: calculate column widths
%
/PsFF_CLw0c0l0 0 def
/PsFF_CLw0c1l0 0 def
/PsFF_CLw1c0l0 0 def
/PsFF_CLw1c1l0 0 def
/PsFF_CLw2c0l0 0 def
/PsFF_CLw2c1l0 0 def
/PsFF_CLw3c0l0 0 def
/PsFF_CLw3c1l0 0 def

% Column #0 of 2
/PsFF_Cw0 0 def
PsFF_rm /PsFF_Cwi 0 def [(Die Roll)] {stringwidth pop dup dup (PsFF_CLw0c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(1)] {stringwidth pop dup dup (PsFF_CLw1c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(2\2614)] {stringwidth pop dup dup (PsFF_CLw2c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(5+)] {stringwidth pop dup dup (PsFF_CLw3c0l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw0 gt {/PsFF_Cw0 exch def} {pop} ifelse} forall

% Column #1 of 2
/PsFF_Cw1 0 def
PsFF_rm /PsFF_Cwi 0 def [(Color)] {stringwidth pop dup dup (PsFF_CLw0c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(blue)] {stringwidth pop dup dup (PsFF_CLw1c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(green)] {stringwidth pop dup dup (PsFF_CLw2c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
PsFF_rm /PsFF_Cwi 0 def [(plaid\331)] {stringwidth pop dup dup (PsFF_CLw3c1l) PsFF_Cwi mkvari exch def /PsFF_Cwi PsFF_Cwi 1 add def PsFF_Cw1 gt {/PsFF_Cw1 exch def} {pop} ifelse} forall
%
% Now adjust column widths for the spans
%
%
% Table contents
%
% Row 0 (1 line)
% Row 0, Col 0, Line 0
PsFF_CLw0c0l0 true true false PsFF_Cw0  PsFF_cellF PsFF_rm (Die Roll) PsFF_cellfragment  PsFF_tEnd
% Row 0, Col 1, Line 0
PsFF_CLw0c1l0 true true false PsFF_Cw1  PsFF_cellF PsFF_rm (Color) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 1 (1 line)
% Row 1, Col 0, Line 0
PsFF_CLw1c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (1) PsFF_cellfragment  PsFF_tEnd
% Row 1, Col 1, Line 0
PsFF_CLw1c1l0 false true false PsFF_Cw1  PsFF_cellF PsFF_rm (blue) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 2 (1 line)
% Row 2, Col 0, Line 0
PsFF_CLw2c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (2\2614) PsFF_cellfragment  PsFF_tEnd
% Row 2, Col 1, Line 0
PsFF_CLw2c1l0 false true false PsFF_Cw1  PsFF_cellF PsFF_rm (green) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
% Row 3 (1 line)
% Row 3, Col 0, Line 0
PsFF_CLw3c0l0 false true false PsFF_Cw0  PsFF_cellF PsFF_rm (5+) PsFF_cellfragment  PsFF_tEnd
% Row 3, Col 1, Line 0
PsFF_CLw3c1l0 false true false PsFF_Cw1  PsFF_cellF PsFF_rm (plaid\331) PsFF_cellfragment  PsFF_tEnd
PsFF_nl
/X PsFF_Xsave def} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (\331Or )(reroll.) ] {PsFF_tbl_footer} ]  [ {PsFF_nl PsFF_rm} [ (\(Subject )(to )(GM )(discretion.\)) ] {PsFF_tbl_footer} ]  [ {PsFF_par} [  ] {} ]  [ {PsFF_nl} [  ] {} ]  [ {PsFF_subsection} [  ] {} ]  [ {PsFF_nl PsFF_rm} [ (Notes:) ] {} ]  [ {PsFF_nl} [ (*May )(cross )(line )(boundaries )(but )(not )(paragraphs.) ] {} ]  [ {} [ (\262May )(nest )(as )(in ) ] {} ]  [ {} [ (/)(/Italic )(*)(*and*)(* )(bold/)(/.) ] {} ]  [ {PsFF_nl} [  ] {PsFF_it } ]  [ {} [ (\263Must )(appear )(at )(the )(very )(beginning )(of )(a )(line.) ] {} ]  ] `, false, nil,
		}} {
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

// @[00]@| Go-GMA 5.27.0-alpha.2
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2025 by Steven L. Willoughby (AKA MadScienceZone)
// @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
// @[12]@| Aloha, Oregon, USA. All Rights Reserved. Some components were introduced at different
// @[13]@| points along that historical time line.
// @[14]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[15]@| License as described in the accompanying LICENSE file distributed
// @[16]@| with GMA.
// @[17]@|
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
