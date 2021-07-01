/*
########################################################################################
#  _______  _______  _______                ___       ______      _______              #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___   )             #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   )  |             #
# | |      | || || || (___) |            / (_) (_       ___) /       /   )             #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (      _/   /              #
# | | \_  )| |   | || (   ) | Game           ) (           ) \    /   _/               #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _ (   (__/\             #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_)\_______/             #
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
		opts []renderOpts
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
*  Item One
*  Item Two
*  Item Three * this is not a
   bullet list
*  But this is
   *  and a sub-list
      *  and sub-sub-list

And this should start a new list:
*  Not that you can tell with bullets.`, false, nil},
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
•  Not that you can tell with bullets.`, false, []renderOpts{WithBullets('•', '‣', '◦')}},
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
		/*
			            self.maxDiff= None
			            self.assertMultiLineEqual(MarkupText(source).render(), expected)

			    def test_html(self):
			        for source, expected in (
			            ('foo',        '<P>foo</P>'),
			            ('',           '<P></P>'),
			            ('foo\nbar',   '<P>foo bar</P>'),
			            ('foo\n\nbar', '<P>foo</P><P>bar</P>'),
			            ('\n\nfoo',    '<P>foo</P>'),
			            (r'foo\\bar',  '<P>foo<BR/>bar</P>'),
			            ('aa//bb//cc', '<P>aa<I>bb</I>cc</P>'),
			            ('aa**bb**cc', '<P>aa<B>bb</B>cc</P>'),
			            ('aa**bb',     '<P>aa<B>bb</B></P>'),
			            ('a//b**c**d//e', '<P>a<I>b<B>c</B>d</I>e</P>'),
			            ('a //b **c //d **e', '<P>a <I>b <B>c </I>d </B>e</P>'),
			            ('a[[b]]d',    '<P>a<A HREF="B">b</A>d</P>'),
			            ('a[[b|c]]d',  '<P>a<A HREF="B">c</A>d</P>'),
			            ('a //it b\nc\\\\de//f', '<P>a <I>it b c<BR/>de</I>f</P>'),
			            ('a //it b\nc\n\nde//f', '<P>a <I>it b c</I></P><P>de<I>f</I></P>'),
			            ('''This is a bullet list:
			*Item One
			*Item //Tw//o
			*Item Three
			 * this is not a\\\\bullet list
			*But this is
			**and a sub-list
			*** and sub-sub-list

			And this should start a new list:
			*Not that you can tell with bullets.
			''', '''<P>This is a bullet list:<UL><LI>Item One<LI>Item <I>Tw</I>o<LI>Item Three * this is not a<BR/>bullet list<LI>But this is<UL><LI>and a sub-list<UL><LI>and sub-sub-list</UL></UL></UL></P><P>And this should start a new list:<UL><LI>Not that you can tell with bullets.</UL></P>'''),
			            ('''This is a numbered list:
			#Item One
			#Item Two
			#Item Three
			 # this is not a\\\\bullet list
			#But this is
			##and a sub-list
			### and sub-sub-list

			And this should start a new list:
			#and this should be re-sequenced.
			''', '''<P>This is a numbered list:<OL><LI>Item One<LI>Item Two<LI>Item Three # this is not a<BR/>bullet list<LI>But this is<OL><LI>and a sub-list<OL><LI>and sub-sub-list</OL></OL></OL></P><P>And this should start a new list:<OL><LI>and this should be re-sequenced.</OL></P>'''),
			            ('''Table test:
			|=Column A|=Column B|
			|left     |    right|
			|  center |filled|
			|  aaa    |   bbb
			And this is after the table.''', '''<P>Table test:<TABLE BORDER=1><TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH></TR><TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD></TR><TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT>filled</TD></TR><TR><TD ALIGN=CENTER>aaa</TD><TD ALIGN=RIGHT>bbb</TD></TR></TABLE>And this is after the table.</P>'''),
			            ('''Table test:
			|=Column A|=Column B|=Column C|
			|left     |    right| some stuff |
			|  center |filled and extended to the other |-
			|  aaa and |-   |   bbb
			And this is after the table.''', '''<P>Table test:<TABLE BORDER=1><TR><TH ALIGN=LEFT>Column A</TH><TH ALIGN=LEFT>Column B</TH><TH ALIGN=LEFT>Column C</TH></TR><TR><TD ALIGN=LEFT>left</TD><TD ALIGN=RIGHT>right</TD><TD ALIGN=CENTER>some stuff</TD></TR><TR><TD ALIGN=CENTER>center</TD><TD ALIGN=LEFT COLSPAN=2>filled and extended to the other</TD></TR><TR><TD ALIGN=CENTER COLSPAN=2>aaa and</TD><TD ALIGN=RIGHT>bbb</TD></TR></TABLE>And this is after the table.</P>'''),
			        ):
			            self.maxDiff = None
			            self.assertMultiLineEqual(MarkupText(source).render(HTMLFormatter()), expected)

			    # output is list of [pre, [str, ...], post] chunks

			    # sb {PsFF_rm} !b !i

			    # pc add chk+sp,...chk from \s+ -> buf
			    # ## [{n-1 PsFF_ind}, n., {n PsFF_ind}]
			    # ** [{n-1 PsFF_ind}, ^., {n PsFF_ind}]
			    # \n sb {PsFF_nl [0 PsFF_ind]}
			    # pp sb {PsFF_nl PsFF_nl [0 PsFF_ind]}
			    # ref==it
			    # sb [fc, chk, {}/?]
			    # fc {PsFF_bi/bf/it/rm}
			    # tbl:
			    # % Data Table: calculate column widths
			    # /PsFF_Cw<n> 0 def
			    # [(c0) (c1) ...] {
			    #     stringwidth pop dup PsFF_Cw<n> gt {
			    #         PsFF_Cw<n> exch def
			    #     } {
			    #         pop
			    #     } ifelse
			    # } forall
			    # x cols
			    # PsFF_nl
			    # ???
			    # x rows

			    def test_ps(self):
			        for source, expected in (
			            ('foo',        [['{PsFF_rm}', ['foo'], '{}']]),
			            ('',           []),
			            ('foo\nbar',   [['{PsFF_rm}', ['foo ', 'bar'], '{}']],),
			            ('foo\n\nbar', [['{PsFF_rm}', ['foo'], '{PsFF_par }'],
			                            ['{}', ['bar'], '{}']]),
			            ('\n\nfoo',    [['{PsFF_rm}', ['foo'], '{}']]),
			            (r'foo\\bar',  [['{PsFF_rm}', ['foo'], '{PsFF_nl}'],
			                            ['{}', ['bar'], '{}']]),
			            ('aa//bb//cc', [['{PsFF_rm}', ['aa'], '{}'],
			                            ['{PsFF_it}', ['bb'], '{}'],
			                            ['{PsFF_rm}', ['cc'], '{}']]),
			            ('aa**bb**cc', [['{PsFF_rm}', ['aa'], '{}'],
			                            ['{PsFF_bf}', ['bb'], '{}'],
			                            ['{PsFF_rm}', ['cc'], '{}']]),
			            ('aa**bb',     [['{PsFF_rm}', ['aa'], '{}'],
			                            ['{PsFF_bf}', ['bb'], '{}']]),
			            ('a//b**c**d//e', [['{PsFF_rm}', ['a'], '{}'],
			                               ['{PsFF_it}', ['b'], '{}'],
			                               ['{PsFF_bi}', ['c'], '{}'],
			                               ['{PsFF_it}', ['d'], '{}'],
			                               ['{PsFF_rm}', ['e'], '{}']]),
			            ('a //b **c //d **e', [['{PsFF_rm}', ['a '], '{}'],
			                                   ['{PsFF_it}', ['b '], '{}'],
			                                   ['{PsFF_bi}', ['c '], '{}'],
			                                   ['{PsFF_bf}', ['d '], '{}'],
			                                   ['{PsFF_rm}', ['e'], '{}']]),
			            ('a[[b]]d',    [['{PsFF_rm}', ['a'], '{}'],
			                            ['{PsFF_it}', ['b'], '{}'],
			                            ['{PsFF_rm}', ['d'], '{}']]),
			            ('a[[b|c]]d',  [['{PsFF_rm}', ['a'], '{}'],
			                            ['{PsFF_it}', ['c'], '{}'],
			                            ['{PsFF_rm}', ['d'], '{}']]),
			            ('a //it b\nc\\\\de//f', [['{PsFF_rm}', ['a '], '{}'],
			                                      ['{PsFF_it}', ['it ', 'b ', 'c'], '{PsFF_nl}'],
			                                      ['{}', ['de'], '{}'],
			                                      ['{PsFF_rm}', ['f'], '{}']]),
			            ('a //it b\nc\n\nde//f', [['{PsFF_rm}', ['a '], '{}'],
			                                      ['{PsFF_it}', ['it ', 'b ', 'c'], '{}'],
			                                      ['{PsFF_rm}', [], '{PsFF_par }'],
			                                      ['{}', ['de'], '{}'],
			                                      ['{PsFF_it}', ['f'], '{}']]),
			            ('''This is a bullet list:
			*Item One
			*Item //Tw//o
			*Item Three
			 * this is not a\\\\bullet list
			*But this is
			**and a sub-list
			*** and sub-sub-list

			And this should start a new list:
			*Not that you can tell with bullets.
			''', [['{PsFF_rm}', ['This ','is ','a ','bullet ','list:'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['^.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item ','One'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['^.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item '], '{}'],
			      ['{PsFF_it}', ['Tw'], '{}'],
			      ['{PsFF_rm}', ['o'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['^.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item ','Three ','* ','this ','is ','not ','a'], '{PsFF_nl}'],
			      ['{}', ['bullet ','list'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['^.'], '{ 1 PsFF_ind }'],
			      ['{}', ['But ','this ','is'], '{PsFF_nl}'],
			      ['{ 1 PsFF_ind }', ['^.'], '{ 2 PsFF_ind }'],
			      ['{}', ['and ','a ','sub-list'], '{PsFF_nl}'],
			      ['{ 2 PsFF_ind }', ['^.'], '{ 3 PsFF_ind }'],
			      ['{}', ['and ','sub-sub-list'], '{PsFF_par  0 PsFF_ind}'],
			      ['{}', ['And ','this ','should ','start ','a ','new ','list:'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['^.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Not ','that ','you ','can ','tell ','with ','bullets.'], '{}']]),
			            ('''This is a numbered list:
			#Item One
			#Item Two
			#Item Three
			 # this is not a\\\\bullet list
			#But this is
			##and a sub-list
			### and sub-sub-list

			And this should start a new list:
			#and this should be re-sequenced.
			''', [['{PsFF_rm}', ['This ','is ','a ','numbered ','list:'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['1.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item ','One'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['2.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item ','Two'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['3.'], '{ 1 PsFF_ind }'],
			      ['{}', ['Item ','Three ','# ','this ','is ','not ','a'], '{PsFF_nl}'],
			      ['{}', ['bullet ','list'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['4.'], '{ 1 PsFF_ind }'],
			      ['{}', ['But ','this ','is'], '{PsFF_nl}'],
			      ['{ 1 PsFF_ind }', ['a.'], '{ 2 PsFF_ind }'],
			      ['{}', ['and ','a ','sub-list'], '{PsFF_nl}'],
			      ['{ 2 PsFF_ind }', ['i.'], '{ 3 PsFF_ind }'],
			      ['{}', ['and ','sub-sub-list'], '{PsFF_par  0 PsFF_ind}'],
			      ['{}', ['And ','this ','should ','start ','a ','new ','list:'], '{PsFF_nl}'],
			      ['{ 0 PsFF_ind }', ['1.'], '{ 1 PsFF_ind }'],
			      ['{}', ['and ','this ','should ','be ','re-sequenced.'], '{}']]),
			            ('''Table test:
			|=Column A|=Column B|
			|left     |    right|
			|  center |filled|
			|  aaa    |   bbb
			And this is after the table.''', [
			    ['{PsFF_rm}', ['Table ', 'test:'], '{PsFF_nl}'],
			    ['{}', [], '{PsFF_rm\n%\n% Data Table: calculate column widths\n%\n/PsFF_Cw0 0 def\n[(Column A) (left) (center) (aaa)] {\n    stringwidth pop dup PsFF_Cw0 gt {\n        /PsFF_Cw0 exch def\n    } {\n        pop\n    } ifelse\n} forall\n\n/PsFF_Cw1 0 def\n[(Column B) (right) (filled) (bbb)] {\n    stringwidth pop dup PsFF_Cw1 gt {\n        /PsFF_Cw1 exch def\n    } {\n        pop\n    } ifelse\n} forall\n    (Column A) PsFF_Cw0 PsFF_thL\n    (Column B) PsFF_Cw1 PsFF_thL\nPsFF_nl\n    (left) PsFF_Cw0 PsFF_tdL\n    (right) PsFF_Cw1 PsFF_tdR\nPsFF_nl\n    (center) PsFF_Cw0 PsFF_tdC\n    (filled) PsFF_Cw1 PsFF_tdL\nPsFF_nl\n    (aaa) PsFF_Cw0 PsFF_tdC\n    (bbb) PsFF_Cw1 PsFF_tdR\nPsFF_nl\n}'],
			    ['{}', ['And ', 'this ', 'is ', 'after ', 'the ', 'table.'], '{}']
			    ]),
			        ):
			            #print "Trying", source
			            #print "Rendered", MarkupText(source).render(PsFormFormatter())

			            self.assertListEqual(MarkupText(source).render(PsFormFormatter()), expected,
			                msg="markup {0} -> {1}, expected {2}".format(
			                    source, MarkupText(source).render(PsFormFormatter()), expected))
		*/

	}
}

// @[00]@| GMA 4.3.2
// @[01]@|
// @[10]@| Copyright © 1992–2021 by Steven L. Willoughby
// @[11]@| (AKA Software Alchemy), Aloha, Oregon, USA. All Rights Reserved.
// @[12]@| Distributed under the terms and conditions of the BSD-3-Clause
// @[13]@| License as described in the accompanying LICENSE file distributed
// @[14]@| with GMA.
// @[15]@|
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
