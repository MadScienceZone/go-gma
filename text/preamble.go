/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______      __     __        __              #
# (  ____ \(       )(  ___  ) Game      (  ____ \    /  \   /  \      /  \             #
# | (    \/| () () || (   ) | Master's  | (    \/    \/) )  \/) )     \/) )            #
# | |      | || || || (___) | Assistant | (____        | |    | |       | |            #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \       | |    | |       | |            #
# | | \_  )| |   | || (   ) |                 ) )      | |    | |       | |            #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _  __) (_ __) (_ _  __) (_           #
# (_______)|/     \||/     \| Client    \______/ (_) \____/ \____/(_) \____/           #
#                                                                                      #
########################################################################################
*/

package text

// <<<<<<<<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// DO NOT EDIT the following code. It is automatically generated
// and any changes you make here will be overwritten.
// <<<<<<<<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

//
// The common preamble we use for a lot of the PostScript
// things we write. This provides some common form generation code
// among other things.
//
// @@:go:form-preamble:begin:commonPostScriptPreamble@@
const commonPostScriptPreamble = `%!PS
%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%  _______  _______  _______              ______      _____                            %
% (  ____ \(       )(  ___  )            / ____ \    / ___ \                           %
% | (    \/| () () || (   ) |           ( (    \/   ( (   ) )                          %
% | |      | || || || (___) |           | (____     ( (___) |                          %
% | | ____ | |(_)| ||  ___  |           |  ___ \     \____  |                          %
% | | \_  )| |   | || (   ) | Game      | (   ) )         ) |                          %
% | (___) || )   ( || )   ( | Master's  ( (___) ) _ /\____) )                          %
% (_______)|/     \||/     \| Assistant  \_____/ (_)\______/                           %
%                                                                                      %
%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%
% @[00]@| GMA Core 6.9
% @[01]@|
% @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
% @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
% @[12]@| Aloha, Oregon, USA. All Rights Reserved.
% @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
% @[14]@| License as described in the accompanying LICENSE file distributed
% @[15]@| with GMA.
% @[16]@|
% @[20]@| Redistribution and use in source and binary forms, with or without
% @[21]@| modification, are permitted provided that the following conditions
% @[22]@| are met:
% @[23]@| 1. Redistributions of source code must retain the above copyright
% @[24]@|    notice, this list of conditions and the following disclaimer.
% @[25]@| 2. Redistributions in binary form must reproduce the above copy-
% @[26]@|    right notice, this list of conditions and the following dis-
% @[27]@|    claimer in the documentation and/or other materials provided
% @[28]@|    with the distribution.
% @[29]@| 3. Neither the name of the copyright holder nor the names of its
% @[30]@|    contributors may be used to endorse or promote products derived
% @[31]@|    from this software without specific prior written permission.
% @[32]@|
% @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
% @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
% @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
% @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
% @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
% @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
% @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
% @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
% @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
% @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
% @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
% @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
% @[45]@| SUCH DAMAGE.
% @[46]@|
% @[50]@| This software is not intended for any use or application in which
% @[51]@| the safety of lives or property would be at risk due to failure or
% @[52]@| defect of the software.
%
%
% This is my standard set of form-building elements for PostScript forms
% Copyright (c) 2003, 2004, 2009, 2010, 2013, 2015 Steve Willoughby, Aloha, Oregon, 
% USA.  All Rights Reserved.
%
% Version 1.4
% (This file used to be part of one monolithic file
% implemented in the "ps.tcl" module of my "gma" software; broken
% into separate files as of version 1.3 to make the form-building
% code easier to use in other places).
%------------------------------------------------------------------------
% Shortcuts for general PostScript operations
% Note that applications may override some of these.
%
% np cp mv ln rln rmv ... ok     basic line drawing
% {wide|med|thin}LineWidth       standard line sizes
% eject                          ship out page border and eject page

/np  { newpath   } def 
/mv  { moveto    } def
/ln  { lineto    } def
/rln { rlineto   } def
/rmv { rmoveto   } def
/cp  { closepath } def
/ok  { stroke    } def

/SetLine_wide  { 3.0 setlinewidth } def
/SetLine_med   { 1.0 setlinewidth } def
/SetLine_thin  { 0.5 setlinewidth } def
%
% w R -
%   move X right w points
%
/R {
    X add /X exch def
} def
%
% h D -
%   move Y down h points
%
/D {
    Y exch sub /Y exch def
} def
%
% Reserve some space for what's coming next, and
% start a new page if we can't do that.
%
/RequiredVerticalSpace {
    Y exch sub PageBottomMargin lt {
        eject
    } if
} def
%
% x y w h BoxFrame -
%   Draw empty, unfilled box with (x,y) at top left
% w h RelativeBoxFrame -
%   Same, but relative to current point
%
/BoxPath {
    4 2 roll        % w h x y
    np mv           % w h     @(x,y)
    RelativeBoxPath
} def
/RelativeBoxPath {      % w h
    dup neg             % w h -h
    0 exch              % w h 0 -h
    rln                 % w h     down h
    exch 0              % h w 0
    rln                 % h       right w
    0 exch              % 0 h
    rln                 %         up h
    cp                  % 
} def
/BoxFrame { BoxPath stroke} def
/BoxFill  { BoxPath fill  } def
/RelativeBoxFrame { RelativeBoxPath stroke} def
/RelativeBoxFill  { RelativeBoxPath fill  } def

/appEject {
    % Applications should override this
    % (called at end of each page)
} def
/appStart {
    % Applications should override this
    % (called at start of each page)
    % WARNING: an extra appStart may be 
    % called at end of job, so
    % don't emit anything here!
} def
/eject {
    appEject
    showpage
    appStart
} def
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
% THEME MANAGEMENT
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
% The form theme should redefine the following commands which
% set colors of elements:
/BlankHue       { 1.00 setgray } def    % blank cell background
/GreyHue    { 0.80 setgray } def    % greyed out elements
/AltGreyHue { 0.60 setgray } def    % for bonus
/DarkHue    { 0.40 setgray } def    % disabled (heavy grey)
/FormHue    { 0.20 setgray } def    % borders, etc.
/LabelTagHue { 0 setgray } def 
/SetColor_form  { /FormColor { FormHue   } def FormColor } def
/SetColor_data  { /FormColor { 0 setgray } def FormColor } def

% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
% FONT MANAGEMENT
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
/FontSize_footnote 6 def
/FontSize_data_ttl 6 def
/FontSize_body 8 def
/FontLead_body 9 def

/FontFace_footnote /Helvetica def
/FontFace_data_ttl /Helvetica def
/FontFace_body /Palatino-Roman def

/SelectFootnoteFont  { FontFace_footnote findfont FontSize_footnote scalefont setfont } def
/SelectDataTitleFont { FontFace_data_ttl findfont FontSize_data_ttl scalefont setfont } def
/SelectBodyFont      { FontFace_body     findfont FontSize_body     scalefont setfont } def

%
% Formatted output with line wrapping
%
% [ chunk0 chunk1 ... chunkN ] width showproc nlproc WrapAndFormat -
%
% prints each chunk in order, calling showproc for each in turn.  Calls
% nlproc as needed to start a new line to avoid going over width.
%
% Each chunk is an array of 3 elements: [ eproc [s0 s1 ... sN] sproc ]
% sproc is called to setup the chunk (font selection, etc), then each
% string s0...sN are displayed as will fit on the line.  Unlike the
% BreakIntoLines procedure, we don't do the breaking ourselves here,
% the generating application does that by assembling the arrays used.
% eproc is called to clean up after the chunk is complete.  
%
% It is legal for sproc or eproc to call showproc and/or nlproc themselves.

/WrapAndFormat {
    /WaF__nlproc exch def
    /WaF__showproc exch def
    /WaF__width exch def
    /WaF__curwidth 0 def
    {
        aload pop exec  % run chunk's sproc
        {
            dup stringwidth pop dup WaF__curwidth add % strN len len+cw 
            WaF__width ge {               % strN len 
                WaF__nlproc
                /WaF__curwidth exch def
            } {
                WaF__curwidth add /WaF__curwidth exch def
            } ifelse
            WaF__showproc
        } forall
        exec    % run chunk's eproc
    } forall
} def

% Generic nlproc and showproc (and friends) you can use with WrapAndFormat
% Call WaF_init first to set starting location for rendering
% x y lead WaF_init -
/WaF_init {
    /WaF__Y exch def
    /WaF__X exch def
    /WaF__lead exch def
    WaF__X WaF__Y moveto
} def
/WaF_nl {
    /WaF__Y WaF__Y WaF__lead sub def
    WaF__X WaF__Y moveto
} def
/WaF_show {
    show
} def

%
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
% Adapted from the Blue Book -- line breaks in running text
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
/wordbreak ( ) def
% string width proc BreakIntoLines -
% Calls:    line proc    for every line broken out
/BreakIntoLines {
    /proc exch def
    /linewidth exch def
    /textstring exch def

    /breakwidth wordbreak stringwidth pop def
    /curwidth 0 def
    /lastwordbreak 0 def
    /startchar 0 def
    /restoftext textstring def

    {
        restoftext wordbreak search {
            /nextword exch def pop
            /restoftext exch def
            /wordwidth nextword stringwidth pop def
            curwidth wordwidth add linewidth gt {
                textstring startchar
                lastwordbreak startchar sub
                getinterval proc
                /startchar lastwordbreak def
                /curwidth wordwidth breakwidth add def
            } {
                /curwidth curwidth wordwidth add
                breakwidth add def
            } ifelse
            /lastwordbreak lastwordbreak nextword length add
            1 add def
        } {
            pop exit
        } ifelse
    } loop
    /lastchar textstring length def
    textstring startchar lastchar startchar sub
    getinterval proc
} def
%
% text x y RenderText w h
% Draw text in current font at (x,y) baseline
% Return text dimensions
%
/RenderText {
    mv dup stringwidth  % text w h
    3 -1 roll       % w h text
    show            % w h
} def
%
% text font size w x y FitText -
% Draw text in specified font at (x,y) baseline, scaled down
% if necessary to fit in maximum width w
%
/FitTextCommon {
    /FormFT_y   exch def 
    /FormFT_x   exch def
    RelativeFitTextCommon
} def
/RelativeFitTextCommon {
    /FormFT_w   exch def
    /FormFT_sz  exch def
    /FormFT_fnt exch def
    /FormFT_txt exch def

    {
        FormFT_fnt findfont FormFT_sz scalefont setfont
        FormFT_txt stringwidth pop FormFT_w le {exit} if
        /FormFT_sz FormFT_sz 0.1 sub def
        FormFT_sz 3 le {exit} if
    } loop
} def

/RelativeFitText {
    RelativeFitTextCommon
    FormFT_txt show
} def
/FitText {
    FitTextCommon
    FormFT_x FormFT_y mv FormFT_txt show
} def
/FitTextCtr {
    FitTextCommon
    FormFT_x FormFT_w 2 div add
    FormFT_txt stringwidth pop 2 div sub
    FormFT_y mv
    FormFT_txt show
} def
/FitTextR {
    FitTextCommon
    FormFT_x FormFT_w add
    FormFT_txt stringwidth pop sub
    FormFT_y mv
    FormFT_txt show
} def

%
% Row of check-boxes, shaded in blocks 
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? CheckBoxMatrix -
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? minboxx maxboxx nl CheckBoxMatrixBounded -
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? title TitledCheckBoxMatrix -
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? title minboxx maxboxx nl TitledCheckBoxMatrixBounded -
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? [labels] minboxx maxboxx nl LabelledCheckBoxMatrixBounded -
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval shadeChecked shadeHalf differential? [labels] LabelledCheckBoxMatrix -
%

/LabelledCheckBoxMatrixBounded {
    () 4 1 roll _coreCBMrender
} def

/TitledCheckBoxMatrixBounded {
    [] 5 1 roll _coreCBMrender
} def

/TitledCheckBoxMatrix {
    [] exch 0 0 {} _coreCBMrender 
} def

/CheckBoxMatrix {
    [] () 0 0 {} _coreCBMrender
} def

/CheckBoxMatrixBounded {
    [] () 5 2 roll _coreCBMrender
} def

/LabelledCheckBoxMatrix {
    () 0 0 {} _coreCBMrender
} def

/LabelledTitledCheckBoxMatrixBounded {_coreCBMrender} def

% ... [labels] title minboxx|0 maxboxx|0 nlfunc _coreCBMrender -
/_coreCBMrender {
    /CBM_nl_proc exch def
    /CBM_maxboxx exch def
    /CBM_minboxx exch def
    /CBM_title exch def
    /CBM_labels exch def
    /CBM_label_i 0 def
    /CBM_label_n CBM_labels length def

    /CBM_differential exch def      % should we use shadeChecked or shadeHalf? else we just keep checking boxes regardless of shade
    exch                            % ... half checked
    dup                             % ... half checked checked
    /CBM_shadedCh exch def          % checked, but for shaded boxes only
    add
    /CBM_shadedHC exch def          % half-checked, but for shaded boxes only
    /CBM_shadeinter exch def
    /CBM_gap_y exch def
    /CBM_gap_x exch def
    /CBM_bx_h exch def
    /CBM_bx_w exch def
    dup /CBM_bx_y exch def /CBM__y exch def
    dup /CBM_bx_x exch def /CBM__x exch def
    /CBM_bpr  exch def      % q u c h
    3 1 roll            % q h u c
    exch                % q h c u
    /CBM_qty_used exch def      % q h c
    dup             % q h c c
    /CBM_qty_chkd exch def      % q h c
    add             % q h+c
    /CBM_qty_half exch def      % q
    /CBM_qty_total exch def     % this adjusts _half and _chkd to be the position for each:
    %  X X X X X / / [] [] [] _ _ _
    % |<-chkd-->|
    % |<---half---->|
    % |<-------used--------->|
    % |<------------qty----------->|
    % 
    % (__x,__y) give the position for the next box to draw
    CBM_title () ne {
        CBM_title stringwidth pop CBM_bx_x add CBM_maxboxx gt {
            /CBM_bx_x CBM_minboxx def
            /CBM__x CBM_minboxx def
            CBM_nl_proc
        } if

	% TODO "7" here should be calculated based on font height
	CBM__x CBM__y 7 sub moveto CBM_title show
        /CBM_bx_x CBM_bx_x CBM_title stringwidth pop add CBM_gap_x add def
	/CBM__x CBM_bx_x def
    } if

    /CBM_shademode false def
    1 1 CBM_qty_total {
        CBM__boxpath CBM_shademode { AltGreyHue } { BlankHue } ifelse fill
        SetColor_form
        CBM__boxpath                % i         (path set)
        dup CBM_qty_used le {           % i i<=u?
            stroke              % i         used block: solid box
            SetColor_data
            CBM_qty_half 0 gt {         % half-checked-qty > 0?
                /CBM_qty_half CBM_qty_half 1 sub def
                np          % i     within chkd/half zone, make 1/2 check
                    CBM__x CBM__y CBM_bx_h sub mv
                    CBM_bx_w CBM_bx_h rln
                stroke
            } if
            CBM_qty_chkd 0 gt {         % checked qty > 0 ?
                /CBM_qty_chkd CBM_qty_chkd 1 sub def
                np          %       within chks zone, make other 1/2 check
                    CBM__x CBM__y mv
                    CBM_bx_w CBM_bx_h neg rln
                stroke
            } if
        } {
                            % outside used zone; draw dashed outline of box
            SetLine_thin
            [1 1] .5 setdash
            stroke
        } ifelse
        % i still on stack here, regardless

        %
        % if we have labels left to place, set the next one on top of this box
        % in the current font
        %
        CBM_label_i CBM_label_n lt {
            CBM__x 1 add CBM__y 1 add CBM_bx_h sub mv 
            gsave /Helvetica findfont CBM_bx_h scalefont setfont LabelTagHue
            CBM_labels CBM_label_i get show
            grestore
            /CBM_label_i CBM_label_i 1 add def
        } if

        dup CBM_shadeinter mod 0 eq {       % i i%int?
            /CBM_shademode CBM_shademode not def    % flip shade mode
            CBM_differential {
                /CBM_qty_chkd CBM_shadedCh /CBM_shadedCh CBM_qty_chkd def def
                /CBM_qty_half CBM_shadedHC /CBM_shadedHC CBM_qty_half def def
            } if
        } if
        /CBM__x CBM__x CBM_gap_x CBM_bx_w add add def   % advance to next horizontal position

        CBM_maxboxx 0 gt CBM__x CBM_maxboxx ge and {
            /CBM__x CBM_minboxx def
            /CBM__y CBM__y CBM_gap_y CBM_bx_h add sub def
            CBM_nl_proc
        } if

        CBM_bpr mod 0 eq {          % box number at end of row?
            /CBM__x CBM_bx_x def
            /CBM__y CBM__y CBM_gap_y CBM_bx_h add sub def
        } if
    } for
    [] 0 setdash
} def

/CBM__boxpath {
    np 
        CBM__x CBM__y mv
        CBM_bx_w 0 rln
        0 CBM_bx_h neg rln
        CBM_bx_w neg 0 rln
    cp
} def


%
% Row of data blocks:
%
% textfont textsize boxheight BeginDataBlock 
% [set-hue-command] data title w DataBlock
% [set-hue-command] data w TitleBlockCtr      centered reversed colors, no title
%   .
%   .
%   .
% EndDataBlock
%
/BeginDataBlock {
    dup RequiredVerticalSpace
    /FormDBboxht exch def
    ChangeDataBlockFont
    /FormDB__X__ X def
} def
/ChangeDataBlockFont {
    /FormDBtsize exch def
    /FormDBtfont exch def
} def
/DataBlockCommon {
    /FormDBbox_w exch def
    SetLine_med
    X Y FormDBbox_w FormDBboxht BoxFill
    FormHue  X Y FormDBbox_w FormDBboxht BoxFrame
    SelectDataTitleFont
    X 1 add Y FontSize_data_ttl sub RenderText pop pop
    FormColor 
    FormDBtfont FormDBtsize FormDBbox_w 6 sub X 3 add Y FormDBboxht .80 mul sub 
    FormDBbox_w R
} def
/DataBlock {
    DataBlockCommon
    FitText
} def
/DataBlockR {
    DataBlockCommon
    FitTextR
} def
/DataBlockCtr {
    DataBlockCommon
    FitTextCtr
} def
/DataBlockCR {  % take 2 data args; first on stack is centered in white, then second right-justified over it
    DataBlockCommon
    1 setgray
    /DBcr__y exch def
    /DBcr__x exch def
    /DBcr__w exch def
    /DBcr__s exch def
    /DBcr__f exch def
    DBcr__f DBcr__s DBcr__w DBcr__x DBcr__y FitTextCtr
    FormColor
    DBcr__f DBcr__s DBcr__w DBcr__x DBcr__y FitTextR
} def

/DataBlockLR {  % take 2 data args; first on stack is centered in white, then second right-justified over it
    DataBlockCommon
    1 setgray
    /DBcr__y exch def
    /DBcr__x exch def
    /DBcr__w exch def
    /DBcr__s exch def
    /DBcr__f exch def
    DBcr__f DBcr__s DBcr__w DBcr__x DBcr__y FitText
    FormColor
    DBcr__f DBcr__s DBcr__w DBcr__x DBcr__y FitTextR
} def

/TitleBlockCtr {
    /FormDBbox_w exch def
    FormHue X Y FormDBbox_w FormDBboxht BoxFill
    SetLine_med X Y FormDBbox_w FormDBboxht BoxFrame
    1 setgray FormDBtfont FormDBtsize FormDBbox_w X Y FormDBboxht .80 mul sub FitTextCtr
    FormDBbox_w R
} def

/EndDataBlock {
    /X FormDB__X__ def
    FormDBboxht D
} def
%
% (End of form-preamble.ps)
%
`
// @@:go:form-preamble:end:@@
// @@:go:gma-preamble:begin:gmaPostScriptPreamble@@
const gmaPostScriptPreamble = `%!PS
%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%  _______  _______  _______              ______      _____                            %
% (  ____ \(       )(  ___  )            / ____ \    / ___ \                           %
% | (    \/| () () || (   ) |           ( (    \/   ( (   ) )                          %
% | |      | || || || (___) |           | (____     ( (___) |                          %
% | | ____ | |(_)| ||  ___  |           |  ___ \     \____  |                          %
% | | \_  )| |   | || (   ) | Game      | (   ) )         ) |                          %
% | (___) || )   ( || )   ( | Master's  ( (___) ) _ /\____) )                          %
% (_______)|/     \||/     \| Assistant  \_____/ (_)\______/                           %
%                                                                                      %
%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%
% @[00]@| GMA Core 6.9
% @[01]@|
% @[10]@| Copyright © 1992–2023 by Steven L. Willoughby (AKA MadScienceZone)
% @[11]@| steve@madscience.zone (previously AKA Software Alchemy),
% @[12]@| Aloha, Oregon, USA. All Rights Reserved.
% @[13]@| Distributed under the terms and conditions of the BSD-3-Clause
% @[14]@| License as described in the accompanying LICENSE file distributed
% @[15]@| with GMA.
% @[16]@|
% @[20]@| Redistribution and use in source and binary forms, with or without
% @[21]@| modification, are permitted provided that the following conditions
% @[22]@| are met:
% @[23]@| 1. Redistributions of source code must retain the above copyright
% @[24]@|    notice, this list of conditions and the following disclaimer.
% @[25]@| 2. Redistributions in binary form must reproduce the above copy-
% @[26]@|    right notice, this list of conditions and the following dis-
% @[27]@|    claimer in the documentation and/or other materials provided
% @[28]@|    with the distribution.
% @[29]@| 3. Neither the name of the copyright holder nor the names of its
% @[30]@|    contributors may be used to endorse or promote products derived
% @[31]@|    from this software without specific prior written permission.
% @[32]@|
% @[33]@| THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
% @[34]@| CONTRIBUTORS “AS IS” AND ANY EXPRESS OR IMPLIED WARRANTIES,
% @[35]@| INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
% @[36]@| MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
% @[37]@| DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS
% @[38]@| BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
% @[39]@| OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
% @[40]@| PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
% @[41]@| PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
% @[42]@| THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR
% @[43]@| TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
% @[44]@| THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
% @[45]@| SUCH DAMAGE.
% @[46]@|
% @[50]@| This software is not intended for any use or application in which
% @[51]@| the safety of lives or property would be at risk due to failure or
% @[52]@| defect of the software.
%
%
% GMA d20/Pathfinder Form Definitions
% Copyright (c) 2003, 2004, 2009, 2010, 2011, 2013, 2015, 2023
% Steve Willoughby, Aloha, Oregon, USA.  All Rights Reserved.
%
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
% COLOR SCHEMES
% %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
statusdict begin
    true setduplexmode
end

%/SetUpFontAttributes {
%    findfont dup length dict
%    begin
%        {
%            1 index /FID ne {
%                def
%            } {
%                pop pop
%            } ifelse
%        } forall
%        /Encoding ISOLatin1Encoding def
%        currentdict
%    end
%    definefont
%    pop
%} def

/_my_encoding StandardEncoding dup length array copy def
_my_encoding 8#200 /Adieresis     put
_my_encoding 8#220 /adieresis     put
_my_encoding 8#345 /copyright     put
_my_encoding 8#346 /registered    put
_my_encoding 8#347 /degree        put
_my_encoding 8#354 /plusminus     put
_my_encoding 8#355 /onequarter    put
_my_encoding 8#356 /onehalf       put
_my_encoding 8#357 /threequarters put
_my_encoding 8#360 /multiply      put
_my_encoding 8#362 /endash        put
_my_encoding 8#330 /zerosuperior  put
_my_encoding 8#331 /onesuperior   put
_my_encoding 8#332 /twosuperior   put
_my_encoding 8#333 /threesuperior put
_my_encoding 8#334 /foursuperior  put
_my_encoding 8#335 /fivesuperior  put
_my_encoding 8#336 /sixsuperior   put
_my_encoding 8#337 /sevensuperior put
_my_encoding 8#340 /eightsuperior put
_my_encoding 8#342 /ninesuperior  put
_my_encoding 8#344 /divide        put

/SetUpFontAttributes {
    findfont dup length dict
    begin
        {
            1 index /FID ne {
                def
            } {
                pop pop 
            } ifelse 
        } forall
        /Encoding _my_encoding def
        currentdict
    end
    definefont
    pop
} def

/HISO /Helvetica      SetUpFontAttributes
/RISO /Palatino-Roman SetUpFontAttributes
/BISO /Palatino-Bold  SetUpFontAttributes
/IBSO /Palatino-BoldItalic SetUpFontAttributes
/CISO /Courier-Bold   SetUpFontAttributes
/ZISO /ZapfChancery-MediumItalic SetUpFontAttributes
/IISO /Palatino-Italic SetUpFontAttributes

/FontFace_footnote /HISO def
/FontFace_data_ttl /HISO def
/FontFace_body     /RISO def
% 
% These define the colors used for standard form features.
% Our character record sheets and GM reference materials are
% different colors in order to easily distinguish them.
%
/SetTheme_d20_Red { 
    /BlankHue       { 1.00 1.00 1.00 setrgbcolor } def
    /GreyHue    { 1.00 0.80 0.80 setrgbcolor } def
    /AltGreyHue { 1.00 0.60 0.60 setrgbcolor } def
    /DarkHue    { 0.80 0.40 0.40 setrgbcolor } def
    /FormHue    { 0.60 0.20 0.20 setrgbcolor } def
    /FaintHue   { 1.00 0.70 0.70 setrgbcolor } def
    /LabelTagHue    { 0.40 0.80 0.40 setrgbcolor } def
    /HighlightHue   { 1.00 1.00 0.50 setrgbcolor } def
} def

/SetTheme_d20_Blue {
    /BlankHue       { 1.00 1.00 1.00 setrgbcolor } def
    /GreyHue        { 0.80 0.80 1.00 setrgbcolor } def
    /AltGreyHue     { 0.60 0.60 1.00 setrgbcolor } def
    /DarkHue        { 0.40 0.40 0.80 setrgbcolor } def
    /FormHue        { 0.20 0.20 0.60 setrgbcolor } def
    /FaintHue   { 0.70 0.70 1.00 setrgbcolor } def
    /LabelTagHue    { 0.80 0.40 0.40 setrgbcolor } def
    /HighlightHue   { 1.00 1.00 0.50 setrgbcolor } def
} def

/SetTheme_d20_Green {
    /BlankHue       { 1.00 1.00 1.00 setrgbcolor } def
    /GreyHue        { 0.80 1.00 0.80 setrgbcolor } def
    /AltGreyHue     { 0.60 1.00 0.60 setrgbcolor } def
    /DarkHue        { 0.40 0.80 0.40 setrgbcolor } def
    /FormHue        { 0.20 0.60 0.20 setrgbcolor } def
    /FaintHue   { 0.70 1.00 0.70 setrgbcolor } def
    /LabelTagHue    { 0.80 0.40 0.40 setrgbcolor } def
    /HighlightHue   { 1.00 1.00 0.50 setrgbcolor } def
} def
/SetTheme_d20_Violet {
    /BlankHue       { 1.00 1.00 1.00 setrgbcolor } def
    /GreyHue        { 1.00 0.80 1.00 setrgbcolor } def
    /AltGreyHue     { 1.00 0.60 1.00 setrgbcolor } def
    /DarkHue        { 0.80 0.40 0.80 setrgbcolor } def
    /FormHue        { 0.60 0.20 0.60 setrgbcolor } def
    /FaintHue   { 1.00 0.70 1.00 setrgbcolor } def
    /LabelTagHue    { 0.80 0.40 0.40 setrgbcolor } def
    /HighlightHue   { 1.00 1.00 0.50 setrgbcolor } def
} def
/SetTheme_d20_Monochrome {
    /BlankHue       { 1.00 setgray } def    % blank cell background
    /GreyHue    { 0.80 setgray } def    % greyed out elements
    /AltGreyHue { 0.60 setgray } def    % for bonus
    /DarkHue    { 0.40 setgray } def    % disabled (heavy grey)
    /FormHue    { 0.20 setgray } def    % borders, etc.
    /FaintHue   { 0.90 setgray } def
    /LabelTagHue    { 0.80 0.40 0.40 setrgbcolor } def
    /HighlightHue  { 1.00 1.00 0.50 setrgbcolor } def
} def

/SetTheme_d20_Character { 
    SetTheme_d20_Red 
    /PageTitleText (Pathfinder Character Record Sheet) def
    /CopyrightText1 (PATHFINDER CHARACTER RECORD SHEET / ) def
    /CopyrightText2 (\2512010, 2015, 2023 Steve Willoughby / REV 4 / 20-NOV-2015) def
} def
/SetTheme_d20_Monsters  { 
    SetTheme_d20_Blue
    /PageTitleText (Pathfinder Encounter Run Sheet) def
    /CopyrightText1 (PATHFINDER ENCOUNTER RUN SHEET / ) def
    /CopyrightText2 (\2512013, 2015, 2023 Steve Willoughby / REV 4 / 07-JUN-2023) def
} def
/SetTheme_d20_GM_Matrix { 
    SetTheme_d20_Green 
    /PageTitleText (Pathfinder GM Player Matrix) def
    /CopyrightText1 (PATHFINDER GM PLAYER MATRIX / ) def
    /CopyrightText2 (\2512013, 2015, 2023 Steve Willoughby / REV 4 / 07-JUN-2023) def
} def
%
% Character Record PostScript Form
% Steve Willoughby
%
% Setup (without emitting anything) at beginning of page
/PF_Continuation_Proc {} def
/PF_Page_Proc {} def
/appStart {
    %
    % page dimensions
    %
    /PageLength         11. 72 mul def
    /PageWidth          8.5 72 mul def
    SetMargins
    PF_Page_Proc
    PF_Continuation_Proc
} def
/SetMargins {
    %
    % margins from origin point
    %
    /PageTopMargin      PageLength 36 sub def
    /PageBottomMargin   36 def
    /PageLeftMargin     36 def
    /PageRightMargin    PageWidth 36 sub def
    %
    /PageTextWidth      PageRightMargin PageLeftMargin sub def
    %
    % home (X,Y)
    %
    /X PageLeftMargin def
    /Y PageTopMargin def
} def
/appEject {
    %
    % Frame page
    %
    SetLine_wide SetColor_form 
    PageLeftMargin 1 sub PageTopMargin 1 add
      PageRightMargin PageLeftMargin sub 2 add
      PageTopMargin PageBottomMargin sub 2 add
      BoxFrame
    %
    % Page title
    %
    /RISO findfont 30 scalefont setfont
    FaintHue PageTitleText 
    PageLeftMargin PageTopMargin 4 add RenderText pop pop

    /Y PageBottomMargin FontSize_footnote sub 2 sub def
    /X PageLeftMargin def
    SelectFootnoteFont 

    FaintHue 
    CopyrightText1 X Y RenderText pop R
    PageTitle X Y RenderText pop pop
    FontSize_footnote D
    CopyrightText2 PageLeftMargin Y RenderText pop pop
} def



%
% ID Char Player CharacterTitleBlock -
%
/DataFont /RISO def
/DataFontBold /BISO def
/DataFontBoldItalic /IBSO def
/DataFontItalic /IISO def
/HeadingFont /HISO def
/HeadingFontSize 8 def

/InterBlockGap { 5 D } def
/CharacterTitleBlock {
    SetColor_data
    30 20 20 20 add add add RequiredVerticalSpace
    DataFont 20 30 BeginDataBlock
    (PLAYER NAME)                  200     BlankHue DataBlock
    (CHARACTER NAME) PageTextWidth 300 sub BlankHue DataBlock
    /CISO 18 ChangeDataBlockFont SetColor_form
    (ID)                           100     GreyHue  DataBlock
    EndDataBlock
    SetColor_data
    DataFont 10 20 BeginDataBlock
    (CLASSES)     PageTextWidth 60 60 60 60 30 20 30 20 30 30 add add add add add add add add add sub BlankHue DataBlock
    (LEVELS)                        60     BlankHue DataBlock
    (HIT DIE TYPE)                  60     BlankHue DataBlock
    (FAVORED)                       60     BlankHue DataBlock
    (RACE)                          60     BlankHue DataBlock
    (ALIGN)                         30     BlankHue DataBlock
    (SEX)                           20     BlankHue DataBlock
    (AGE)                           30     BlankHue DataBlock
    (SIZE)                          20     BlankHue DataBlock
    (HEIGHT)                        30     BlankHue DataBlock
    (WEIGHT)                        30     BlankHue DataBlock
    EndDataBlock
    DataFont 10 20 BeginDataBlock
    (DEITY)                                                 100     BlankHue DataBlock
    (GENERAL APPEARANCE) PageTextWidth 100 30 30 30 add add add sub BlankHue DataBlock
    (EYES)                                                   30     BlankHue DataBlock
    (HAIR)                                                   30     BlankHue DataBlock
    (SKIN)                                                   30     BlankHue DataBlock
    EndDataBlock
    DataFont 10 20 BeginDataBlock
    (CHARACTER BACKGROUND)          PageTextWidth BlankHue DataBlock
    EndDataBlock
} def
% save Y position for next columns in block; vertical list of abilities
/BeginAbilityScores {
    /NextX X 60 add 30 add 40 add 40 add 40 add def
    /NextY Y def
} def
/AbilityScore {
    SetColor_data
    DataFontBold 18 20 BeginDataBlock
        60 TitleBlockCtr
        () 30 BlankHue DataBlockR
        DataFont 12 ChangeDataBlockFont
        40 BlankHue DataBlockR
        (TMP) 40 GreyHue DataBlockR
        (TMP MOD) 40 GreyHue DataBlockR
    EndDataBlock
} def
/EndAbilityScores {
    /BottomY Y def
    /X NextX def
    /Y NextY def
} def
        
/BeginSavingThrowBlock {
    /NextX X 240 add def
    /NextY Y def
    /OurX X def
} def
/StrengthBlock {
    /X OurX def
    SetColor_data
    DataFont 12 20 BeginDataBlock
        (CARRYING WEIGHT) 80 BlankHue DataBlockR
        (LIFTING WT)      40 BlankHue DataBlockR
        (PUSH/DRAG)       50 BlankHue DataBlockR
        (SPEED)           30 BlankHue DataBlockR
        (ARM/LOAD)        40 BlankHue DataBlockR
    EndDataBlock
} def
/SavingThrow {
    /X OurX def
    SetColor_data
    DataFontBold 18 20 BeginDataBlock
        60 TitleBlockCtr
        () 30 BlankHue DataBlockR
        DataFont 12 ChangeDataBlockFont
        (= BASE) 30 GreyHue DataBlockR
        30 GreyHue DataBlockR
        (+ MAG) 30 GreyHue DataBlockR
        (+ RACE) 30 GreyHue DataBlockR
        (+ MISC) 30 GreyHue DataBlockR
    EndDataBlock
} def
/EndSavingThrowBlock {
    2 {
        /X OurX def
        DataFontBold 14 20 BeginDataBlock
            (Mods) 60 TitleBlockCtr
            () PageRightMargin X sub DataBlock
        EndDataBlock
    } repeat

    Y BottomY gt {
        /BottomY Y def
    } if
    /X NextX def
    /Y NextY def
} def
        
/BeginXPBlock {
    /OurX X 5 add def
    /TopY Y def
} def
/XPBlock {
    /X OurX def
    SetColor_data
    DataFont 18 20 BeginDataBlock
        (XP) PageRightMargin X sub BlankHue DataBlockR
    EndDataBlock
    DataFont 12 20 BeginDataBlock
        (NEEDED FOR NEXT LEVEL) PageRightMargin X sub GreyHue DataBlockR
    EndDataBlock
    DataFont 18 20 BeginDataBlock
        (HERO POINTS) PageRightMargin X sub BlankHue DataBlockR
    EndDataBlock
    DataFont 12 20 BeginDataBlock
        (GOLD) PageRightMargin X sub BlankHue DataBlockR
    EndDataBlock
    FormHue SetLine_wide OurX 1 sub TopY 1 add PageRightMargin X sub 2 add 80 BoxFrame
} def
/EndXPBlock {
    /X PageLeftMargin def
    /Y BottomY def
} def
/BeginAttackBlock {
    /AB_atk_x PageLeftMargin def
    /AB_ini_w 80 def
    /AB_dam_w 60 def
    /AB_atk_w PageTextWidth AB_ini_w sub AB_dam_w sub 10 sub def
    /AB_dam_x AB_atk_x AB_atk_w add def
    /AB_ini_x AB_dam_x AB_dam_w add 5 add def

    60 HeadingFontSize 4 add add RequiredVerticalSpace
    /PF_Continuation_Proc {
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        (ATTACKS) AB_atk_w TitleBlockCtr
        (DAMAGE)  AB_dam_w 10 add TitleBlockCtr
        (INITIATIVE) AB_ini_w TitleBlockCtr
        EndDataBlock
        /AB_start_y Y def
        /AB_end_y Y def
    } def
    PF_Continuation_Proc
} def
/AttackBlock {
    SetColor_data
    DataFontBold 18 20 BeginDataBlock
        75 TitleBlockCtr
        () AB_atk_w 75 sub 80 sub 30 4 mul sub BlankHue DataBlockR
        DataFont 12 ChangeDataBlockFont
        (= BASE) 80 GreyHue DataBlockR
        30 GreyHue DataBlockR
        30 GreyHue DataBlockR
        (+ MISC) 30 GreyHue DataBlockR
        (+ TEMP) 30 GreyHue DataBlockR
    EndDataBlock
    /AB_end_y Y def
} def
/DamageBlock {
    SetColor_data
    /X AB_dam_x 5 add def
    /Y AB_start_y def
    DataFont 12 20 BeginDataBlock
        (NORM)30 BlankHue DataBlockR
        (OFF-H)30 BlankHue DataBlockR
    EndDataBlock
    /X AB_dam_x 5 add def
    DataFont 12 20 BeginDataBlock
        (TWO-HANDED)60 BlankHue DataBlockR
    EndDataBlock
    /X AB_dam_x 5 add def
    DataFont 12 20 BeginDataBlock
        (BOW/SLING)60 BlankHue DataBlockR
    EndDataBlock
} def
/InitiativeBlock {
    SetColor_data
    /X AB_ini_x 5 add def
    /Y AB_start_y def
    DataFontBold 18 20 BeginDataBlock
        (Init) 40 TitleBlockCtr
        () 40 BlankHue DataBlockR
    EndDataBlock
    2 {
        /X AB_ini_x 5 add def
        DataFont 12 20 BeginDataBlock
            40 GreyHue DataBlockR
            40 GreyHue DataBlockR
        EndDataBlock
    } repeat
} def
/EndAttackBlock {
    /X PageLeftMargin def
    /Y AB_end_y def
    /PF_Continuation_Proc {} def
} def
/BeginWeaponBlock {
    SetColor_data
    28 HeadingFontSize 4 add add RequiredVerticalSpace
    /PF_Continuation_Proc {
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    /WB_desc_w PageTextWidth 50 sub 50 sub 85 sub 35 sub 20 sub 50 sub 30 sub 20 sub 20 sub 35 sub def
    (WEAPON DESCRIPTION) WB_desc_w TitleBlockCtr
    (ATTACK)                  50 TitleBlockCtr
    (DAMAGE)                  85 TitleBlockCtr
    (THREAT)                  35 TitleBlockCtr
    (CRIT)                    20 TitleBlockCtr
    (RANGE)                   50 TitleBlockCtr
    (REACH)                   30 TitleBlockCtr
    (WT)                      20 TitleBlockCtr
    (SZ)                      20 TitleBlockCtr
    (TYPE)                    35 TitleBlockCtr
    (ID)                      50 TitleBlockCtr
    EndDataBlock
    } def
    PF_Continuation_Proc
} def
/Weapon {
    SetColor_data
    DataFont 12 14 BeginDataBlock
        () WB_desc_w BlankHue DataBlock
        () 50 BlankHue DataBlockR
        () 85 BlankHue DataBlockR
        () 35 BlankHue DataBlockR
        () 20 BlankHue DataBlockR
        () 50 BlankHue DataBlockR
        () 30 BlankHue DataBlockR
        () 20 BlankHue DataBlockR
        () 20 BlankHue DataBlock
        () 35 BlankHue DataBlock
        /Courier-Bold 10 ChangeDataBlockFont SetColor_form
        () 50 GreyHue  DataBlock
    EndDataBlock
} def
/EndWeaponBlock {
    /PF_Continuation_Proc {} def
} def

/BeginMonsterDefenseBlock {
    /DB_ac_x PageLeftMargin def
    /DB_ac_w 60 30 add 15 add 25 8 mul add def
    /DB_misc_x DB_ac_x DB_ac_w add def
    /DB_misc_w Mon_Skill_x DB_misc_x sub def

    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (DEFENSES) DB_ac_w DB_misc_w add TitleBlockCtr
    EndDataBlock
    /DB_start_y Y def
    /DB_ac_y Y def
} def
    
/BeginDefenseBlock {
    /DB_ac_x PageLeftMargin def
    /DB_ac_w 60 30 add 15 add 25 8 mul add def
    /DB_hp_x DB_ac_x DB_ac_w add def
    /DB_hp_w PageTextWidth DB_ac_w sub 5 sub def

    80 HeadingFontSize 4 add add RequiredVerticalSpace
    /PF_Continuation_Proc {
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        (DEFENSES) DB_ac_w TitleBlockCtr
        (HEALTH) PageTextWidth DB_ac_w sub TitleBlockCtr
        EndDataBlock
        /DB_start_y Y def
        /DB_ac_y Y def
    } def
    PF_Continuation_Proc
} def
/ACBlock {
    SetColor_data
    DataFontBold 18 20 BeginDataBlock
        60 TitleBlockCtr
        () 30 BlankHue DataBlockR
        DataFont 12 ChangeDataBlockFont
        (10)(= 10) 15 GreyHue DataBlockR
        8 {
            0 eq {
                25 GreyHue DataBlockR
            } {
                25 DarkHue DataBlockR
            } ifelse
        } repeat
    EndDataBlock
    /DB_ac_y Y def
} def
/HP {
    SetColor_data
    /X DB_hp_x 5 add def
    /Y DB_start_y def
    DataFontBold 18 20 BeginDataBlock
        (HP) 40 TitleBlockCtr
        () DB_hp_w 70 sub 40 sub 40 sub 20 sub BlankHue DataBlockR
        DataFont 12 ChangeDataBlockFont
        (HEAL RATE) 40 BlankHue DataBlockR
        (DAMAGE REDUCTION) 70 BlankHue DataBlockR
        (SR) 20 BlankHue DataBlockR
    EndDataBlock

    DataFont 18 60 BeginDataBlock
        ()(CURRENT HIT POINT TRACKING) DB_hp_w BlankHue DataBlock
    /X DB_hp_x 5 add def

    % qty used checked 1/2 bpr x y w h gapx y interval Sc Sh diff? CheckBoxMatrix
    % makes a 30x6 matrix of little boxes (up to 180 HP)
    % 180 15 0 0 
    % Auto-adjust size based on max number of boxes
    4 -1 roll dup   % TTL used ch 1/2  -> used ch 1/2 TTL TTL
    45 le {
        % 0-45 boxes: 10x10 boxes, 15/row
        4 1 roll 15 X 10 add Y 8 sub 10 10 4 4 5 0 0 false CheckBoxMatrix
    } {
        dup 80 le {
            % 46-80 boxes: use 8x8, 20/row
            4 1 roll 20 X 5 add Y 8 sub 8 8 3 3 5 0 0 false CheckBoxMatrix
        } {
            % 81-180 boxes: use 5x5, 30/row
            4 1 roll 30 X 10 add Y 8 sub 5 5 2 2 5 0 0 false CheckBoxMatrix
        } ifelse
    } ifelse
        
    SetLine_med
    4 3 1 1 5 X 10 add DB_start_y 72 sub 5 5 2 2 5 0 0 false CheckBoxMatrix
    SelectDataTitleFont
    (= Lethal Wound, Non-Lethal Wound, Available HP, Future/Temp HP) X 40 add DB_start_y 77 sub RenderText pop pop
    EndDataBlock
    /Y DB_ac_y def
} def
/BeginArmourBlock {
    SetColor_data
    28 HeadingFontSize 4 add add RequiredVerticalSpace
    /PF_Continuation_Proc {
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /ArB_desc_w PageTextWidth 10 sub 30 sub 50 sub 30 5 mul sub 50 sub def
        (ON)                        10 TitleBlockCtr
        (ARMOR DESCRIPTION) ArB_desc_w TitleBlockCtr
        (BONUS)                     30 TitleBlockCtr
        (TYPE)                      50 TitleBlockCtr
        (MDEX)              30 TitleBlockCtr
        (CHK)               30 TitleBlockCtr
        (SPL)               30 TitleBlockCtr
        (SPD)               30 TitleBlockCtr
        (WT)                30 TitleBlockCtr
        (ID)                        50 TitleBlockCtr
        EndDataBlock
    } def
    PF_Continuation_Proc
} def
/EndArmourBlock {
    /PF_Continuation_Proc {} def
} def
/Armour {
    SetColor_data
    DataFont 12 14 BeginDataBlock
        () () 10 BlankHue DataBlock
        1 1 3 -1 roll 0 1 X 10 sub 2.5 add Y 4.5 sub 5 5 2 2 5 0 0 false CheckBoxMatrix
        () ArB_desc_w BlankHue DataBlock
        () 30 BlankHue DataBlockR
        () 50 BlankHue DataBlock
        () 30 BlankHue DataBlockR
        () 30 BlankHue DataBlockR
        () 30 BlankHue DataBlockR
        () 30 BlankHue DataBlockR
        () 30 BlankHue DataBlockR
        /Courier-Bold 10 ChangeDataBlockFont SetColor_form
        () 50 GreyHue  DataBlock
    EndDataBlock
} def
/BeginFeatBlock {
    SetColor_data
    28 HeadingFontSize 4 add add RequiredVerticalSpace
    /FB_desc_w PageTextWidth 2 div 40 sub def
    /PF_Continuation_Proc {
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        2 {
            (FEAT/SPECIAL ABILITY DESCRIPTION) FB_desc_w TitleBlockCtr
            (REF) 40 TitleBlockCtr
        } repeat
        EndDataBlock
        /FB_col 0 def
    } def
    PF_Continuation_Proc
} def

/EndFeatBlockNoFooter {
    FB_col 0 gt {
        ()()Feat
    } if
    /PF_Continuation_Proc {} def
} def

/EndFeatBlock {
    EndFeatBlockNoFooter
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        ()(\256=Racial Ability; \251=Class Ability; \253Modifier already calculated into other stats\273) 
        PageTextWidth BlankHue DataBlock
    EndDataBlock
} def
/Feat {
    SetColor_data
    FB_col 0 eq {
        DataFont 12 14 BeginDataBlock
        /FB_col 1 def
    } {
            /FB_col 0 def
    } ifelse
    () FB_desc_w BlankHue DataBlock
    () 40 BlankHue DataBlockR
    FB_col 0 eq {
        EndDataBlock
    } if
} def

/BeginSkillBlock {
    14 35 mul HeadingFontSize 4 add add RequiredVerticalSpace
    DataFontBold 18 20 BeginDataBlock
    /SkB_header_w PageTextWidth 60 sub 60 sub 30 sub 30 sub 60 sub 60 sub 60 sub def
    SetColor_form
        (SKILLS) () SkB_header_w BlankHue DataBlock
        DataFont 12 ChangeDataBlockFont SetColor_data
        (TOTAL POINTS) 60 BlankHue DataBlockR
        (POINTS/LEVEL) 60 BlankHue DataBlockR
        (= BASE \(CLASS\)) 60 GreyHue DataBlockR
        (+ INT) 30 GreyHue DataBlockR
        (+ RACE) 30 GreyHue DataBlockR
        (MAX RANKS \(= HD\)) 60 GreyHue DataBlockR
        (ADDED @ LEVEL) 60 GreyHue DataBlockR
    EndDataBlock

    /PF_hdr {
        SetColor_data
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /SkB_desc_w PageTextWidth 10 3 mul sub 125 sub 30 8 mul sub def
        (C) 10 TitleBlockCtr
        (T) 10 TitleBlockCtr
        (A) 10 TitleBlockCtr
        (SKILL NAME) 125 TitleBlockCtr
        (MOD) 30 TitleBlockCtr
        (=ABIL) 30 TitleBlockCtr
        (+RANK) 30 TitleBlockCtr
        (+CLS) 30 TitleBlockCtr
        (+SYN) 30 TitleBlockCtr
        (+RACE) 30 TitleBlockCtr
        (+ITEM) 30 TitleBlockCtr
        (+MISC) 30 TitleBlockCtr
        (SKILL DESCRIPTION) SkB_desc_w TitleBlockCtr
        EndDataBlock
    } def
    PF_hdr
    /PF_Continuation_Proc {
        DataFontBold 18 20 BeginDataBlock
        /SkB_header_w PageTextWidth def
        SetColor_form
            (SKILLS \(Continued\)) () SkB_header_w BlankHue DataBlock
        EndDataBlock
        PF_hdr
    } def
} def
/EndSkillBlock {
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        ()(C=Class skill; T=Training Required; A=Armor Check Penalty \(Swim=2\327\)) PageTextWidth BlankHue DataBlock
    EndDataBlock
    /PF_Continuation_Proc {} def
} def
/Skill {
    SetColor_data
    DataFont 12 14 BeginDataBlock
        3 {
            () () 10 BlankHue DataBlock
            1 1 3 -1 roll 0 1 X 10 sub 2.5 add Y 4.5 sub 5 5 2 2 5 0 0 false CheckBoxMatrix
        } repeat
        () 125 BlankHue DataBlock
        () 30 BlankHue DataBlockR
        () 30 GreyHue DataBlockLR
        6 {
            () 30 GreyHue DataBlockR
        } repeat
        () SkB_desc_w BlankHue DataBlock
    EndDataBlock
} def
    
% title BeginAnimalBlock -
/BeginAnimalBlock {
    /ABtitle exch def
    InterBlockGap
    14 10 mul HeadingFontSize 4 add add RequiredVerticalSpace
    DataFontBold 18 20 BeginDataBlock
        SetColor_form
        ABtitle () PageTextWidth BlankHue DataBlock
    EndDataBlock
    /PF_Continuation_Proc {
        DataFontBold 18 20 BeginDataBlock
        SetColor_form
            ABtitle () PageTextWidth BlankHue DataBlock
        EndDataBlock
    } def
} def
/EndAnimalBlock {
    /PF_Continuation_Proc {} def
} def

/BeginRelationshipBlock {
   InterBlockGap
   14 5 mul HeadingFontSize 4 add RequiredVerticalSpace
   DataFontBold 18 20 BeginDataBlock
        SetColor_form
        (Relationships)()PageTextWidth BlankHue DataBlock
   EndDataBlock
    /PF_hdr {
        SetColor_data
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /ItB_desc_w PageTextWidth 40 sub 30 sub 50 sub def
        (NPC)     100 TitleBlockCtr
        (SCORE)   30 TitleBlockCtr
        (LEVEL)   80 TitleBlockCtr
        (EFFECT)  200 TitleBlockCtr
        EndDataBlock
    } def
    PF_hdr
   /PF_Continuation_Proc {
   DataFontBold 18 20 BeginDataBlock
        SetColor_form
        (Relationships \(contd\))()PageTextWidth BlankHue DataBlock
   EndDataBlock
   PF_hdr
   } def
} def

% effect level score name RelationshipScore -
/RelationshipScore {
    SetColor_data
    DataFont 12 14 BeginDataBlock
        () 100 BlankHue DataBlock
        ()  30 BlankHue DataBlockR
        ()  80 BlankHue DataBlock
        () 200 BlankHue DataBlock
    EndDataBlock
} def

/EndRelationshipBlock {
    /PF_Continuation_Proc {} def
} def
        
/BeginItemBlock {
    14 10 mul HeadingFontSize 4 add add RequiredVerticalSpace
    DataFontBold 18 20 BeginDataBlock
    /ItB_header_w PageTextWidth 60 sub def
    SetColor_form
        (POSSESSIONS) () ItB_header_w BlankHue DataBlock
        DataFont 12 ChangeDataBlockFont SetColor_data
        (TOTAL WEIGHT) 60 BlankHue DataBlockR
    EndDataBlock

    /PF_hdr {
        SetColor_data
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /ItB_desc_w PageTextWidth 40 sub 30 sub 50 sub def
        (DESCRIPTION) ItB_desc_w TitleBlockCtr
        (LOC)            40 TitleBlockCtr
        (WT)              30 TitleBlockCtr
        (ID)                  50 TitleBlockCtr
        EndDataBlock
    } def
    PF_hdr
    /PF_Continuation_Proc {
        DataFontBold 18 20 BeginDataBlock
        SetColor_form
            (POSSESSIONS \(Continued\)) () PageTextWidth BlankHue DataBlock
        EndDataBlock
        PF_hdr
    } def
} def
/EndItemBlock {
    /PF_Continuation_Proc {} def
} def
% id wt loc desc total used checked half Item
/Item {             % I W L D t u c h
    4 -1 roll dup       % I W L D u c h t t
    54 add 55 idiv dup  % I W L D u c h t rows rows
    10 mul 14 add dup   % I W L D u c h t rows vsize vsize
    RequiredVerticalSpace   % I W L D u c h t rows vsize
    SetColor_data
    DataFont 12         % I W L D u c h t rows vsize f 12
    3 -1 roll       % I W L D u c h t rows f 12 vsize
    BeginDataBlock      % I W L D u c h t rows
        9 5 roll    % u c h t rows I W L D
        () ItB_desc_w BlankHue DataBlock
        () 40 BlankHue DataBlock
        () 30 BlankHue DataBlockR
        /Courier-Bold 10 ChangeDataBlockFont SetColor_form
        () 50 GreyHue DataBlock
        0 gt {          % u c h t rows>0?
            4 1 roll    % t u c h
            55 PageLeftMargin 5 add Y 4 sub 5 5 2 2 5 0 0 false CheckBoxMatrix
        } {
            pop pop pop pop
        } ifelse
    EndDataBlock
} def
% ttl avail cast prep norm DCAST DPREP known_9..known_0 title SpellSlotBlock -
% \________9..0______________________/
/SpellSlotBlock {
    14 10 mul HeadingFontSize 4 add add RequiredVerticalSpace
    DataFontBold 18 20 BeginDataBlock
    /SpB_header_w PageTextWidth 30 sub 20 9 mul sub def
    SetColor_form
        () SpB_header_w BlankHue DataBlock
        DataFont 12 ChangeDataBlockFont SetColor_data
        (KNOW 0) 30 BlankHue DataBlockR
        (I) 20 BlankHue DataBlockR
        (II) 20 BlankHue DataBlockR
        (III) 20 BlankHue DataBlockR
        (IV) 20 BlankHue DataBlockR
        (V) 20 BlankHue DataBlockR
        (VI) 20 BlankHue DataBlockR
        (VII) 20 BlankHue DataBlockR
        (VIII) 20 BlankHue DataBlockR
        (IX) 20 BlankHue DataBlockR
    EndDataBlock
    SetColor_form
    20 D
    /SpB_col_w PageTextWidth 3 div def
    /SpB_row_0 Y def
    /SpB_font /BISO def
    /SpB_sz   12 def
    /X PageLeftMargin def
    /Y SpB_row_0 def
    (III)(II)(I)(0) 4 {
        SpB_font SpB_sz 40 X Y FitTextR
        14 D
    } repeat
    /Y SpB_row_0 14 sub def
    /X X SpB_col_w add def
    (VI)(V)(IV) 3 {
        SpB_font SpB_sz 40 X Y FitTextR
        14 D
    } repeat
    /Y SpB_row_0 14 sub def
    /X X SpB_col_w add def
    (IX)(VIII)(VII) 3 {
        SpB_font SpB_sz 40 X Y FitTextR
        14 D
    } repeat
    /X PageLeftMargin def
    /Y SpB_row_0 def
    /BxSz 10 def
    4 {                 % ttl used chk half unshaded
        /SSB_diff exch def
        /SSB_dp   exch def
        /SSB_dc   exch def
        3 index 0 lt {
            % number of boxes < 0 means "unlimited use"
            pop pop pop pop pop
            ((unlimited)) SpB_font SpB_sz 200 X 50 add Y FitText
        } {
            20 X 50 add Y 9 add BxSz BxSz 2 2   % t u c h u 20 x+40 y 5 5 2 2
            8 -1 roll           % t u c h 20 x+40 y 5 5 2 2 u
            SSB_dc SSB_dp SSB_diff CheckBoxMatrix
        } ifelse
        14 D
    } repeat

    /X X SpB_col_w add def
    /Y SpB_row_0 14 sub def
    3 {             % ttl used chk half unshaded
        /SSB_diff exch def
        /SSB_dp   exch def
        /SSB_dc   exch def
        20 X 50 add Y 9 add BxSz BxSz 2 2   % t u c h u 20 x+40 y 5 5 2 2
        8 -1 roll       % t u c h 20 x+40 y 5 5 2 2 u
        SSB_dc SSB_dp SSB_diff CheckBoxMatrix
        14 D
    } repeat

    /X X SpB_col_w add def
    /Y SpB_row_0 14 sub def
    3 {             % ttl used chk half unshaded
        /SSB_diff exch def
        /SSB_dp   exch def
        /SSB_dc   exch def
        20 X 50 add Y 9 add BxSz BxSz 2 2   % t u c h u 20 x+40 y 5 5 2 2
        8 -1 roll       % t u c h 20 x+40 y 5 5 2 2 u
        SSB_dc SSB_dp SSB_diff CheckBoxMatrix
        14 D
    } repeat

    4 4 1 1 4 PageRightMargin SpB_col_w sub 30 sub SpB_row_0 15 add 5 5 2 2 3 0 0 false CheckBoxMatrix
    SetColor_form
    (= Cast, Prepared, Spare, Domain Spell Slots) HeadingFont HeadingFontSize SpB_col_w
        PageRightMargin SpB_col_w sub SpB_row_0 10 add FitText


    /X PageLeftMargin def
% qty used checked half-checked boxesperrow box-x y w h gap-x y shadeinterval sc sh diff? CheckBoxMatrix -
} def

% mod ability DC9 .. DC0 concentration title BeginSpellBlock
/BeginSpellBlock {
    /SpB_title exch def
    14 10 mul HeadingFontSize 4 add add RequiredVerticalSpace
    /SpB_header_w PageTextWidth 20 10 mul sub 20 sub 20 sub 20 sub 30 sub def
    DataFontBold 18 20 BeginDataBlock
    SetColor_form
        SpB_title () SpB_header_w BlankHue DataBlock
        DataFont 12 ChangeDataBlockFont SetColor_data
        (CONCEN) 30 BlankHue DataBlockR
        (DC 0) 20 BlankHue DataBlockR
        (I) 20 BlankHue DataBlockR
        (II) 20 BlankHue DataBlockR
        (III) 20 BlankHue DataBlockR
        (IV) 20 BlankHue DataBlockR
        (V) 20 BlankHue DataBlockR
        (VI) 20 BlankHue DataBlockR
        (VII) 20 BlankHue DataBlockR
        (VIII) 20 BlankHue DataBlockR
        (IX) 20 BlankHue DataBlockR
        (10)(=10) 20 GreyHue DataBlockR
        20 GreyHue DataBlockR
        (+SL)(+LVL) 20 GreyHue DataBlockR
    EndDataBlock

    /PF_hdr {
        SetColor_data
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /SpB_desc_w PageTextWidth 60 sub 25 sub 35 sub 50 sub 40 sub 40 sub 30 sub 20 sub 20 sub 20 sub 20 sub def
        (PREP) 20 TitleBlockCtr
        (LVL) 20 TitleBlockCtr
        (TITLE) SpB_desc_w TitleBlockCtr
        (SCHOOL)       60 TitleBlockCtr
        (CAST)         25 TitleBlockCtr
        (COMP)         35 TitleBlockCtr
        (RANGE)        50 TitleBlockCtr
        (TARGET)       40 TitleBlockCtr
        (DURATION)     40 TitleBlockCtr
        (SAVE)         30 TitleBlockCtr
        (SR)           20 TitleBlockCtr
        (REF)          20 TitleBlockCtr
        EndDataBlock
    } def
    PF_hdr
    /PF_Continuation_Proc {
        SetColor_form
        DataFontBold 12 20 BeginDataBlock
        SpB_title (SPELL LIST CONTINUED FROM PREVIOUS PAGE) PageTextWidth BlankHue DataBlock
        EndDataBlock
        PF_hdr
    } def
} def
/EndSpellBlock {
    /PF_Continuation_Proc {} def
} def
/SpellInstanceExtension {} def
/Spell {            % ref sr sv dur trg rng comp ct sch name lvl #prep #cast [flags] txt
    %
    % start by counting the number of lines of text this will require
    %
    SelectBodyFont
    /Sp_text_block exch def
    /Sp_lines 0 def
    %/Sp_text exch def
    %-- Sp_text PageTextWidth 10 sub { pop /Sp_lines Sp_lines 1 add def } BreakIntoLines
    /Sp_tb_w PageTextWidth 10 sub def
    /Sp_sFlags exch def
    /Sp_nCast exch def
    /Sp_nPrep exch def
    
    /Sp_inline_qty 2 def
    /Sp_xCast 0 def
    /Sp_xPrep 0 def
    Sp_nCast Sp_inline_qty gt {
        /Sp_xCast Sp_nCast Sp_inline_qty sub def
        /Sp_nCast Sp_inline_qty def
        /Sp_inline_qty 0 def
    } {
        /Sp_inline_qty Sp_inline_qty Sp_nCast sub def
    } ifelse
    Sp_nPrep Sp_inline_qty gt {
        /Sp_xPrep Sp_nPrep Sp_inline_qty sub def
        /Sp_nPrep Sp_inline_qty def
        /Sp_inline_qty 0 def
    } {
        /Sp_inline_qty Sp_inline_qty Sp_nPrep sub def
    } ifelse
    %
    % At this point,
    %   Sp_nCast, Sp_nPrep are the number of spells cast and prepared shown on the main line
    %   Sp_inline_qty is the number of slots on the main line still unused
    %   Sp_xCast, Sp_xPrep are the number of spells cast and prepared that need to be added on an additional line
    %

        

    X Y FontLead_body FontSize_body PsFF_init
    Sp_text_block Sp_tb_w PsFF_vertspace 12 add RequiredVerticalSpace
    %12 FontLead_body Sp_lines mul add RequiredVerticalSpace
    SetColor_data
    DataFont 9 10 BeginDataBlock
      ()() 20 GreyHue DataBlock
      2 2 Sp_nCast Sp_nPrep 2 X 18.0 sub Y 1.5 sub 7 7 2 2 100 0 0 false Sp_sFlags LabelledCheckBoxMatrix
        () 20 GreyHue DataBlockCtr
        () SpB_desc_w GreyHue DataBlock
        () 60 GreyHue DataBlock
        () 25 GreyHue DataBlock
        () 35 GreyHue DataBlockCtr
        () 50 GreyHue DataBlock
        () 40 GreyHue DataBlock
        () 40 GreyHue DataBlockR
        () 30 GreyHue DataBlock
        () 20 GreyHue DataBlockCtr
        () 20 GreyHue DataBlockR
    EndDataBlock
    Sp_sFlags length 2 gt {
        /Sp_sFlags Sp_sFlags 2 Sp_sFlags length 2 sub getinterval def
    } {
        /Sp_sFlags [] def
    } ifelse

    Sp_xCast Sp_xPrep add 0 gt {
        DataFont 9 10 BeginDataBlock
         ()() PageTextWidth GreyHue DataBlock
         Sp_xCast Sp_xPrep add dup Sp_xCast Sp_xPrep 100 PageLeftMargin 1.5 add Y 1.5 sub 7 7 2 2 100 0 0 false 
          Sp_sFlags LabelledCheckBoxMatrix
        EndDataBlock
    } if

    SpellInstanceExtension
    /SpellInstanceExtension {} def

    Sp_text_block length 0 gt {
        SelectBodyFont
        /X PageLeftMargin def
        %%% Sp_text PageTextWidth 10 sub { FontLead_body D X 5 add Y 2 add mv show } BreakIntoLines
        X 5 add Y FontLead_body 2 sub sub FontLead_body FontSize_body PsFF_init
        Sp_text_block Sp_tb_w PsFF_WaF
        /X PageLeftMargin def
        2 D
    } if
} def

/BeginExploitBlock {
    /SpB_title exch def
    14 10 mul HeadingFontSize 3 add add RequiredVerticalSpace
    DataFontBold 16 20 BeginDataBlock
    /ItB_header_w PageTextWidth 80 sub def
    SetColor_form
        (ARCANIST EXPLOITS) () ItB_header_w BlankHue DataBlock
        DataFont 12 ChangeDataBlockFont SetColor_data
        (DAILY POOL) 40 BlankHue DataBlockCtr
        (MAX POOL) 40 BlankHue DataBlockCtr
    EndDataBlock

    /PF_hdr {
        SetColor_data
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        /SpB_desc_w PageTextWidth 25 sub 20 sub 20 sub 25 sub 50 sub 40 sub 50 sub 30 sub def
        (COST)         20 TitleBlockCtr
        (TYPE)         25 TitleBlockCtr
        (NAME) SpB_desc_w TitleBlockCtr
        (DC)           20 TitleBlockCtr
        (CAST)         25 TitleBlockCtr
        (RANGE)        50 TitleBlockCtr
        (TARGET)       40 TitleBlockCtr
        (DURATION)     50 TitleBlockCtr
        (SAVE)         30 TitleBlockCtr
        EndDataBlock
    } def
    PF_hdr
    /PF_Continuation_Proc {
        SetColor_form
        DataFontBold 12 20 BeginDataBlock
        SpB_title (EXPLOIT LIST CONTINUED FROM PREVIOUS PAGE) PageTextWidth BlankHue DataBlock
        EndDataBlock
        PF_hdr
    } def
} def

/EndExploitBlock {
    /PF_Continuation_Proc {} def
} def

/Exploit {            % sv dur trg rng ct sch name lvl #prep #cast [flags] txt
    SelectBodyFont
    /Exp_text_block exch def
    /Exp_tb_w PageTextWidth 10 sub def

    X Y FontLead_body FontSize_body PsFF_init
    Exp_text_block Exp_tb_w PsFF_vertspace 12 add RequiredVerticalSpace
    SetColor_data
    /ExpB_desc_w PageTextWidth 25 sub 20 sub 20 sub 25 sub 50 sub 40 sub 50 sub 30 sub def
    DataFont 9 10 BeginDataBlock
        () 20 GreyHue DataBlockCtr
        () 25 GreyHue DataBlockCtr
        () ExpB_desc_w GreyHue DataBlock
        () 20 GreyHue DataBlockCtr
        () 25 GreyHue DataBlockCtr
        () 50 GreyHue DataBlockCtr
        () 40 GreyHue DataBlock
        () 50 GreyHue DataBlockCtr
        () 30 GreyHue DataBlockCtr
    EndDataBlock

    Exp_text_block length 0 gt {
        SelectBodyFont
        /X PageLeftMargin def
        X 5 add Y FontLead_body 2 sub sub FontLead_body FontSize_body PsFF_init
        Exp_text_block Exp_tb_w PsFF_WaF
        /X PageLeftMargin def
        2 D
    } if
} def

/EndDefenseBlock {
    /X PageLeftMargin def
    /PF_Continuation_Proc {} def
} def

/EndMonsterDefenseBlock {
    EndDefenseBlock
} def
%
% Formatted text block support
%
% Call this first before starting into a text block
%   baselineX baselineY leading fontsize PsFF_init -
%
/PsFF_init {
    /PsFF_sz exch def
    /PsFF_ld exch def
    /PsFF__y exch def
    /PsFF__x exch def
    /PsFF__i 0 def
    /PsFF_TcolPad 2 def  % padding inside table cells on each side
    /PsFF_TcolSpn PsFF_TcolPad 2 mul def  % space reclaimed each span
    /PsFF_rm { DataFont findfont PsFF_sz scalefont setfont } def
    /PsFF_bf { DataFontBold findfont PsFF_sz scalefont setfont } def
    /PsFF_it { DataFontItalic findfont PsFF_sz scalefont setfont } def
    /PsFF_bi { DataFontBoldItalic findfont PsFF_sz scalefont setfont } def
    /PsFF_section { DataFontBold findfont PsFF_sz 1.2 mul scalefont setfont } def
    /PsFF_tbl_caption { PsFF_section } def
    /PsFF_subsection { DataFontBold findfont PsFF_sz 1.1 mul scalefont setfont } def
    /PsFF_tbl_footer { DataFont findfont PsFF_sz 0.8 mul scalefont setfont } def
    /PsFF_tbl_footer_it { DataFontItalic findfont PsFF_sz 0.8 mul scalefont setfont } def
    /PsFF_tbl_footer_bf { DataFontBold findfont PsFF_sz 0.8 mul scalefont setfont } def
    /PsFF_tbl_footer_bi { DataFontBoldItalic findfont PsFF_sz 0.8 mul scalefont setfont } def

    /X PsFF__x def
    /Y PsFF__y def
    X Y moveto
} def

/PsFF_p {
	PsFF_allowOutput {PsFF_show} {pop} ifelse
} def

/PsFF_indsz 10 def  % points of indent for each level

% level PsFF_ind -
%   set indentation level; moves cursor to (indented) start position

/PsFF_ind {
    PsFF_indsz mul /PsFF__i exch def
    /X PsFF__x PsFF__i add def
    X Y moveto
} def

%
% PsFF_nl -
%   Skip to next line
%

/PsFF_nl {
    /Y Y PsFF_ld sub def
    /X PsFF__x PsFF__i add def
    X Y moveto
    /WaF__curwidth 0 def
} def

/PsFF_show { show } def
%
% [ chunk0 chunk1 ... chunkN ] w PsFF_WaF -
%   Wrap the text in the array of chunks into a column w points wide
%   each chunk is a 3-element array:
%       [ eproc [s0 s1 s2 ... sN] sproc ]
%       where sproc is executed, then the strings s0...sN
%       are printed (calling PsFF_nl as needed to wrap lines between them)
%       then eproc is called after the strings are printed.
%       a new page is ejected if the whole set won't fit on the current page
%
% [ chunk0 chunk1 ... chunkN ] w PsFF_vertspace y
%       Calculates the amount of vertical space (y) required by the block.

/PsFF_vertspace {
    /PsFF__c exch def   % column width to wrap within
    /PsFF__L 0 def
    /PsFF_allowOutput false def
    /PsFF_nl  { /PsFF__L PsFF__L 1 add def /WaF__curwidth 0 def } def
    /PsFF_par { /PsFF__L PsFF__L 1.5 add def /WaF__curwidth 0 def } def
    PsFF__c { pop } { PsFF_nl } WrapAndFormat
    PsFF__L PsFF_ld mul
    /PsFF_allowOutput true def
} def

/PsFF_WaF {
    /PsFF__c exch def
    %dup
    %PsFF__c PsFF_vertspace RequiredVerticalSpace
    /PsFF__i 0 def
    /X PsFF__x PsFF__i add def
    /Y PsFF__y def
    X Y moveto
    /PsFF_nl { /Y Y PsFF_ld sub def /X PsFF__x PsFF__i add def X Y moveto /WaF__curwidth 0 def } def
    /PsFF_par { /Y Y PsFF_ld 1.5 mul sub def /X PsFF__x PsFF__i add def X Y moveto /WaF__curwidth 0 def } def
    PsFF__c { show } { PsFF_nl } WrapAndFormat
} def

%
% PsFF_Lc -
%   Fake "newline" command which just increments the line counter

/PsFF_Lc {
    /PsFF__L PsFF__L 1 add def
} def

%[chunklist] width showproc nlproc WrapAndFormat -
% height x y WaF_init
% WaF_nl
% (str) WaF_show

% PsFF_Cw<n>  width in points for table column <n> (0-origin)
% (text) width PsFF_t<t><a>
%   set table cell of type <t>={h (header), d (data)}, align <a>={C, L, R}
% PsFF_nl assumed to start next row of table properly
%
% str width PsFF_t[hd][CLR]
%   typeset a table cell, either (h)eader or (d)ata style,
%   (C)entered, (L)eft, or (R)ight justified
%
/PsFF_allowOutput true def

/PsFF_tBoxPath {
    newpath
        X Y 2 sub moveto
        PsFF__cl_w__ 4 add 0 rlineto
        0 PsFF_ld rlineto
        PsFF__cl_w__ 4 add neg 0 rlineto
        closepath
} def
%
% Table cells are set at (X,Y) with a rule drawn around them;
% after the cell is typeset, (X,Y) is moved to where the next
% cell in the row would start:
%
%        _____________________________
%       |                             |
%       |<-->|                    |<->|
%       | pad                      pad|
% ______@____TEXT_GOES_HERE___________@'__________________________
%  pad ^|                             |                  BASELINE 
%      v|_____________________________|
%             colwidth + spn
%
% @=(X,Y) starting point
% @'=(X,Y) ending point
% pad=padding (horiz and vert) around text inside box
% spn=distance reclaimed in box for each span (=2*pad)
%
% NEW table formatting support which alows for multi-line table cells with a mixture of fonts within
% the cell body.
%
% datawidth header? centered? right? colwidth PsFF_cell[TMBF_] --
% 	start cell with rules T=top, M=middle, B=bottom, F=full, _=none
% 	if centered? then move in enough distance for the full text to be centered
% 	if right? then move in enough distance for the full text to be right-aligned
% str PsFF_cellfragment --
% 	typeset (next) part of the cell's contents
%
/PsFF_ntcommon_ { 
	/PsFF__cl_w__ exch def 
	/PsFF__cl_R__ exch def
	/PsFF__cl_C__ exch def
	/PsFF__cl_hdr__ exch def
	/PsFF__cl_dw__ exch def
} def
/PsFF_ntalign_ {
	PsFF__cl_C__ {
		PsFF__cl_dw__ 2 div neg PsFF__cl_w__ 2 div add X add PsFF_TcolPad add Y moveto
	} {
		PsFF__cl_R__ {
			PsFF__cl_dw__ neg PsFF__cl_w__ add X add PsFF_TcolPad add Y moveto
		} {
			X PsFF_TcolPad add Y moveto
		} ifelse
	} ifelse
} def

/PsFF_cellfragment {
	PsFF_allowOutput {
		PsFF__cl_hdr__ {
			1 setgray show 0 setgray
		} {
			show
		} ifelse
	} { 
		pop 
	} ifelse
} def

/PsFF_cellT { 
	PsFF_ntcommon_ 
	PsFF_allowOutput {
		SetLine_thin
		PsFF_pathT ok
		PsFF_hdrfill
		PsFF_ntalign_
	} if
} def 

/PsFF_hdrfill {
	PsFF__cl_hdr__ {
		PsFF_pathF fill
	} if
} def

/PsFF_pathT {
	np X Y 2 sub mv
	0 PsFF_ld rln 
	PsFF__cl_w__ 4 add 0 rln 
	0 PsFF_ld neg rln 
} def

/PsFF_cellM {
	PsFF_ntcommon_ 
	PsFF_allowOutput {
		SetLine_thin
		PsFF_pathMl ok
		PsFF_pathMr ok
		PsFF_hdrfill
		PsFF_ntalign_
	} if
} def 

/PsFF_pathMl {
	np X Y 2 sub mv
	0 PsFF_ld rln 
} def

/PsFF_pathMr {
	np X PsFF__cl_w__ add 4 add Y 2 sub mv
	0 PsFF_ld rln
} def

/PsFF_cellB {
	PsFF_ntcommon_ 
	PsFF_allowOutput {
		SetLine_thin
		PsFF_pathB ok
		PsFF_hdrfill
		PsFF_ntalign_
	} if
} def

/PsFF_pathB {
	np X Y 2 sub PsFF_ld add mv
	0 PsFF_ld neg rln 
	PsFF__cl_w__ 4 add 0 rln
	0 PsFF_ld rln
} def

/PsFF_cellF {
	PsFF_ntcommon_
	PsFF_allowOutput {
		SetLine_thin
		PsFF_pathF ok
		PsFF_hdrfill
		PsFF_ntalign_
	} if
} def

/PsFF_pathF {
	np X Y 2 sub mv
	PsFF__cl_w__ 4 add 0 rln
	0 PsFF_ld rln
	PsFF__cl_w__ 4 add neg 0 rln
	cp
} def

/PsFF_cell_ {
	PsFF_ntcommon_
	PsFF_allowOutput {
		PsFF_hdrfill
		PsFF_ntalign_
	} if
} def

%
% Older table formatting commands
%
/PsFF_tcommon {
    % X Y is at baseline of next character.  In our case, we descend 2pt below for rules
    % and move in 2pt on each side.  We leave X Y at the baseline for the next cell after
    % us.
    
    /PsFF__cl_w__ exch def
    SetLine_thin PsFF_tBoxPath 
} def
/PsFF_tEnd {
    /X X PsFF_TcolSpn add PsFF__cl_w__ add def
    X Y moveto
} def

/PsFF_tdL {
    PsFF_allowOutput {
        PsFF_tcommon stroke
        X PsFF_TcolPad add Y moveto show
        PsFF_tEnd
    } { pop pop } ifelse
} def
/PsFF_tdR {
    PsFF_allowOutput {
        PsFF_tcommon stroke
        dup stringwidth pop neg PsFF__cl_w__ add X add PsFF_TcolPad add Y moveto show
        PsFF_tEnd
    } { pop pop } ifelse
} def
/PsFF_tdC {
    PsFF_allowOutput {
        PsFF_tcommon stroke
        dup stringwidth pop 2 div neg PsFF__cl_w__ 2 div add X add PsFF_TcolPad add Y moveto show
        PsFF_tEnd
    } { pop pop } ifelse
} def 

/PsFF_thL { 
    PsFF_allowOutput {
        PsFF_tcommon fill 1 setgray
        X PsFF_TcolPad add Y moveto show 0 setgray
        PsFF_tEnd
    } { pop pop } ifelse
} def

/PsFF_thC {
    PsFF_allowOutput {
        PsFF_tcommon fill 1 setgray
        dup stringwidth pop 2 div neg PsFF__cl_w__ 2 div add X add PsFF_TcolPad add Y moveto show
        0 setgray
        PsFF_tEnd
    } { pop pop } ifelse
} def
/PsFF_thR {
    PsFF_allowOutput {
        PsFF_tcommon stroke 1 setgray
        dup stringwidth pop neg PsFF__cl_w__ add X add PsFF_TcolPad add Y moveto show
        0 setgray
        PsFF_tEnd 
    } { pop pop } ifelse
} def

/ThePageCounter 1 def
/ThePage (**) def
/__PFGMA_buf 20 string def
/BeginModule {
    SetColor_data
    /AreaID () def
    /MonsterID () def
    /PF_Page_Proc {
        /ThePageCounter ThePageCounter 1 add def
        /ThePage ThePageCounter __PFGMA_buf cvs def
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        SetColor_form
        ModuleID  ()  20 GreyHue DataBlockCtr
        AreaID    ()  25 GreyHue DataBlockCtr
        MonsterID () PageTextWidth 20 sub 25 sub 20 sub BlankHue DataBlock
        ThePage   ()  20 BlankHue DataBlockR
        EndDataBlock
    } def

    % RequiredVerticalSpace
    DataFontBold 28 30 BeginDataBlock
        () PageTextWidth 50 sub BlankHue DataBlock
        dup /ModuleID exch def
        DataFontBold 20 ChangeDataBlockFont
        (MODULE ID) 50 GreyHue DataBlockCtr
    EndDataBlock
} def
/EndModule {
    eject
    /ModuleID () def
} def
/BeginArea {
    25                                          % title block
    20 add                                      % first monster
    20 add                                      % stats block
        FontLead_body 2 add add                 % description (1 line) 
        HeadingFontSize 4 add 20 3 mul add add  % defenses (3 lines)
        HeadingFontSize 4 add 20 3 mul add add  % attacks (3 lines)
        RequiredVerticalSpace
    dup /AreaID exch def
    SetColor_data
    DataFontBold 18 20 BeginDataBlock
        DataFontBold 12 ChangeDataBlockFont
        (AREA)        50     GreyHue DataBlockCtr
        DataFontBold  18     ChangeDataBlockFont
        PageTextWidth 50 sub TitleBlockCtr
    EndDataBlock
    InterBlockGap
} def
/EndArea {
    InterBlockGap
    /AreaID () def
} def
% 
% init sz spc reach(with?) AC ff touch HD saves DR SR Spd baseatk cmb(..) cmd(...)
% senses
% align 
%
% S D C I W C 
% m m m m m m 
%
%

/ReserveMonsterSpace {
    20                                          % title block
    20 add                                      % name block
    20 add                                      % stats block
        FontLead_body 2 add add                 % description (1 line) 
        HeadingFontSize 4 add 20 3 mul add add  % defenses (3 lines)
        HeadingFontSize 4 add 20 3 mul add add  % attacks (3 lines)
        RequiredVerticalSpace
} def
/ReserveMinimalMonsterSpace {
    20                                          % title block
    20 add                                      % name block
        RequiredVerticalSpace
} def
/BeginMonster {
    % We need at least the stat, description, 3x defenses, 3x attacks worth of space here.
    % xp cr ref align tag type name generic
    /GenName exch def

    ReserveMonsterSpace
    SetColor_data
    dup /MonsterID exch def 
    DataFontBold 18 20 BeginDataBlock
        GenName () eq {
            () PageTextWidth 190 sub 60 sub 20 sub 20 sub 20 sub 40 sub BlankHue DataBlock
            DataFont 12 ChangeDataBlockFont SetColor_data
            (TYPE) 190 BlankHue DataBlock
        } {
            () PageTextWidth 160 sub 60 sub 60 sub 20 sub 20 sub 20 sub 40 sub BlankHue DataBlock
            DataFont 12 ChangeDataBlockFont SetColor_data
            (TYPE) 160 BlankHue DataBlock
            GenName (GENERIC) 60 BlankHue DataBlock
        } ifelse
        (IMAGE TAG) 60 HighlightHue DataBlock
        (ALIGN) 20 BlankHue DataBlockCtr
        (REF)   20 GreyHue DataBlockR
        (CR)    20 GreyHue DataBlockR
        (XP)    40 GreyHue DataBlockR
    EndDataBlock
    /DB_skill_y Y def
} def
/MonsterDescription {
    /Mon_Skill_x PageRightMargin 100 sub def
    Mon_Skill_x PageLeftMargin sub LimitedFreeFormTextBlock
} def
/MonsterStats {
    SetColor_data
    DataFont 12 20 BeginDataBlock
        /Mon_Skill_x PageRightMargin 100 sub def
        /Mon_Skill_w 100 def 
        /stretch PageTextWidth Mon_Skill_w sub 20 sub 15 sub 20 sub 25 sub 20 sub 25 sub 40 sub
            25 sub 30 sub 45 sub 20 sub 45 sub 20 sub 20 sub def
        (INIT)     20 BlankHue DataBlockR
        (DEX)      20 HighlightHue DataBlockR
        (MOD)      20 HighlightHue DataBlockR
        (SIZE)     15 HighlightHue DataBlockCtr
        (SPC)      20 BlankHue DataBlockCtr
        ()         25 GreyHue  DataBlock
        (RCH)      20 BlankHue DataBlockCtr
        ()         25 GreyHue  DataBlock
        (HIT DICE) stretch BlankHue DataBlock
        (SPEED)    40 BlankHue DataBlockR
        (BAB)      25 GreyHue  DataBlockR
        (CMB)      30 BlankHue DataBlockR
        ()         45 GreyHue  DataBlock
        (CMD)      20 BlankHue DataBlockR
        ()         45 GreyHue  DataBlock
    EndDataBlock
} def
/MonsterMiscDefenses {
    /X DB_misc_x def /Y DB_start_y def
    DataFont 12 20 BeginDataBlock
        (DR)             DB_misc_w 20 sub BlankHue DataBlock
        (SR)             20 BlankHue DataBlockR
    EndDataBlock
    /X DB_misc_x def
    DataFont 12 20 BeginDataBlock
        (RESIST/*IMMUNE) DB_misc_w BlankHue DataBlock
    EndDataBlock
    /X DB_misc_x def
    DataFont 12 20 BeginDataBlock
        (VULNERABLE)     DB_misc_w 35 sub BlankHue DataBlock
        (REGEN)          35 BlankHue DataBlockR
    EndDataBlock
    /DB_mmisc_y Y def
} def
/MonsterAbilityScores {
    /Y DB_skill_y def
    /X Mon_Skill_x def
    /as_w Mon_Skill_w 6 div def
    DataFont 8 10 BeginDataBlock
    (S) as_w GreyHue DataBlockR
    (D) as_w GreyHue DataBlockR
    (C) as_w HighlightHue DataBlockR
    (I) as_w GreyHue DataBlockR
    (W) as_w GreyHue DataBlockR
    (C) as_w GreyHue DataBlockR
    EndDataBlock
    /X Mon_Skill_x def
    DataFont 8 10 BeginDataBlock
    () as_w GreyHue DataBlockR
    () as_w GreyHue DataBlockR
    () as_w GreyHue DataBlockR
    () as_w GreyHue DataBlockR
    () as_w GreyHue DataBlockR
    () as_w GreyHue DataBlockR
    EndDataBlock
    /DB_skill_y Y def
} def
/MonsterSavingThrows {
    /Y DB_skill_y def
    /X Mon_Skill_x def
    /as_w Mon_Skill_w 3 div def
    DataFont 10 12 BeginDataBlock
    (FORT) as_w BlankHue DataBlockR
    (REFL) as_w BlankHue DataBlockR
    (WILL) as_w BlankHue DataBlockR
    EndDataBlock
    /DB_skill_y Y def
} def
/BeginMonsterSkills {
    /Y DB_skill_y def
    /X Mon_Skill_x def
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (SKILLS) Mon_Skill_w TitleBlockCtr
    EndDataBlock
    /DB_skill_y Y def
} def
/MonsterSkill {
    /X Mon_Skill_x def
    DataFont 8 10 BeginDataBlock
    () Mon_Skill_w 20 sub BlankHue DataBlock
    () 20 BlankHue DataBlockR
    EndDataBlock
    /DB_skill_y Y def
} def
/EndMonsterSkills {
    /Y DB_ac_y def
    Y DB_mmisc_y gt {
        /Y DB_mmisc_y def
    } if
    Y DB_skill_y gt {
        /Y DB_skill_y def
    } if
    Y AB_end_y gt {
        /Y AB_end_y def
    } if
    /X PageLeftMargin def
} def
/BeginMonsterWeaponBlock {
    /AB_atk_x PageLeftMargin def
    /AB_atk_w Mon_Skill_x AB_atk_x sub def
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (ATTACK DESCRIPTION) AB_atk_w 50 sub 85 sub 35 sub 20 sub 50 sub 30 sub TitleBlockCtr
    (ATTACK)             50 TitleBlockCtr
    (DAMAGE)             85 TitleBlockCtr
    (THREAT)             35 TitleBlockCtr
    (CRIT)               20 TitleBlockCtr
    (RANGE)              50 TitleBlockCtr
    (REACH)              30 TitleBlockCtr
    EndDataBlock
    /AB_end_y Y def
} def
/MonsterWeapon {
    SetColor_data
    { 
        /MWhue {AltGreyHue} def 
        pop 
    } {
        { 
            /MWhue {BlankHue} def 
        } { 
            /MWhue {GreyHue} def 
        } ifelse
    } ifelse
    DataFontBold 12 14 BeginDataBlock
    () AB_atk_w 50 sub 85 sub 35 sub 20 sub 50 sub 30 sub MWhue DataBlock
    () 50 MWhue DataBlockR
    () 85 MWhue DataBlockR
    () 35 MWhue DataBlockR
    () 20 MWhue DataBlockR
    () 50 MWhue DataBlockR
    () 30 MWhue DataBlock
    EndDataBlock
    /AB_end_y Y def
} def
/EndMonsterWeaponBlock {
} def

/FreeFormTextBlock {
    PageTextWidth LimitedFreeFormTextBlock
} def

/LimitedFreeFormTextBlock {
    /fftb_width exch def
    /Note_text_block exch def
    /Notes_tb_w fftb_width 10 sub def
    X Y FontLead_body FontSize_body PsFF_init
    Note_text_block Notes_tb_w PsFF_vertspace 12 add RequiredVerticalSpace
    SelectBodyFont
    /X PageLeftMargin def
    X 5 add Y FontLead_body 2 sub sub FontLead_body FontSize_body PsFF_init
    Note_text_block Notes_tb_w PsFF_WaF
    /X PageLeftMargin def
    2 D
} def

/DividerBlock {
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    PageTextWidth TitleBlockCtr
    EndDataBlock
} def

/NotesBlock {
    X Y FontLead_body FontSize_body PsFF_init
    dup PageTextWidth 10 sub PsFF_vertspace 12 add RequiredVerticalSpace
    SetColor_form
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (NOTES) PageTextWidth TitleBlockCtr
    EndDataBlock
    SetColor_data
    FreeFormTextBlock
} def

/ConditionsBlock {
	(CONDITIONS) DividerBlock
	SetColor_data
	DataFont 12 20 BeginDataBlock
	(STARTING CONDITION LIST) BlankHue PageTextWidth DataBlock
	EndDataBlock
} def

/TemplateDeclaration {
    SetColor_form
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    PageTextWidth TitleBlockCtr
    EndDataBlock
    SetColor_data
    /DB_skill_y Y def
} def
    
/BeginExtraMonsterBlock {
	(EXTRA CREATURES NOT INVOLVED IN COMBAT) DividerBlock
} def

/EndExtraMonsterBlock {
	InterBlockGap
} def

% spawn sizes color tag ExtraMonster -
/ExtraMonster {
	SetColor_data
	DataFont 12 20 BeginDataBlock
	(IMAGE TAG) BlankHue PageTextWidth 175 sub 75 sub 50 sub DataBlock
	(COLOR) 50 BlankHue DataBlock
	(SIZE\(S\)) 175 BlankHue DataBlock
	(SPAWN POINT) 75 BlankHue DataBlock
	EndDataBlock
} def
/BeginMonsterInstanceBlock {
} def
/EndMonsterInstanceBlock {
} def
/MonsterInstance {
    /nonlethalHP exch def
    /lethalHP exch def
    /totalHP exch def
    /hpspace PageTextWidth def
    totalHP 35 le {
        /bxw 10 def /bxg 4 def
    } {
        /bxw 5 def /bxg 2 def
    } ifelse 

    /bpr hpspace cvi bxw bxg add idiv dup 5 mod sub def
    /rows totalHP bpr div ceiling def
    rows bxw bxg add mul HeadingFontSize 4 add add RequiredVerticalSpace
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    PageTextWidth TitleBlockCtr
    EndDataBlock
    totalHP totalHP lethalHP nonlethalHP bpr X 2 add Y 2 sub bxw bxw bxg bxg 5 0 0 false CheckBoxMatrix
    rows bxw bxg add mul 2 add D
} def

/EndMonster{
    /MonsterID () def
    InterBlockGap
} def

/StartLandscape {
    gsave
    0 0 moveto
    90 rotate
    0 PageWidth neg translate
    /PageLength PageWidth /PageWidth PageLength def def
    SetMargins
} def
/EndLandscape {
    grestore
    /PageLength PageWidth /PageWidth PageLength def def
    SetMargins
} def

/BeginStatMatrix {
    SetColor_data
    /PF_Page_Proc {
        /ThePageCounter ThePageCounter 1 add def
        /ThePage ThePageCounter __PFGMA_buf cvs def
        HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
        SetColor_form
        ThePage () 20 BlankHue DataBlockR
        EndDataBlock
    } def
} def
/BeginSkillMatrix {
    SetColor_data
} def
/EndStatMatrix {
    InterBlockGap
} def
/BeginSkillMatrixTitle {
    SetColor_data
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (NAME) SMRnameW TitleBlockCtr
} def
/SMSkillName {
    SMSkillW TitleBlockCtr
} def
/EndSkillMatrixTitle {
    EndDataBlock
} def
/EndSkillMatrix {
    
} def
/MXsBeginBasic {
    SetColor_data
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (NAME) SMRnameW TitleBlockCtr
    (CLASS) 50 TitleBlockCtr
    (LEVEL) 50 TitleBlockCtr
    (R) 20 TitleBlockCtr
    (AL) 20 TitleBlockCtr
    (FORT) 25 TitleBlockCtr
    (REFL) 25 TitleBlockCtr
    (WILL) 25 TitleBlockCtr
    (--STR--)  40 TitleBlockCtr
    (--DEX--)  40 TitleBlockCtr
    (--CON--)  40 TitleBlockCtr
    (--INT--)  40 TitleBlockCtr
    (--WIS--)  40 TitleBlockCtr
    (--CHA--)  40 TitleBlockCtr
    (INITIATIVE) 45 TitleBlockCtr
    (AC)   20 TitleBlockCtr
    (TCH)20 TitleBlockCtr
    (FF) 20 TitleBlockCtr
    (CMD)  20 TitleBlockCtr
    (DC)   20 TitleBlockCtr
    EndDataBlock
} def
/SMRnameW {
    PageTextWidth
    50 50 add 20 add 20 add 25 3 mul add 15 25 add 6 mul add 
    20 add 25 add 20 4 mul add 20 add
    sub
} def
/BeginPlayerSkillBlock {
    DataFontBold 12 14 BeginDataBlock
    RowHue () SMRnameW DataBlock
} def
/SMSkillW 25 def
/SMSkillValue {
    RowHue () SMSkillW DataBlockR
} def
/SMSkillNull {
    DarkHue (//////) () SMSkillW DataBlockCtr
} def
/EndPlayerSkillBlock {
    EndDataBlock
} def
/SM_EvenRow {
    /RowHue {GreyHue} def
} def
/SM_OddRow {
    /RowHue {BlankHue} def
} def
/StatMatrixRow {
    DataFontBold 12 14 BeginDataBlock
     RowHue () SMRnameW DataBlock   % name
     RowHue () 50 DataBlock   % class
     RowHue () 50 DataBlock   % level
     RowHue () 20 DataBlock   % race
     RowHue () 20 DataBlock   % alignment
     RowHue () 25 DataBlockR  % fort
     RowHue () 25 DataBlockR  % refl
     RowHue () 25 DataBlockR  % will
     AltGreyHue  () 15 DataBlockR  % str
     RowHue () 25 DataBlockR  % str mod
     AltGreyHue  () 15 DataBlockR  % dex
     RowHue () 25 DataBlockR  % dex mod
     AltGreyHue  () 15 DataBlockR  % con
     RowHue () 25 DataBlockR  % con mod
     AltGreyHue  () 15 DataBlockR  % int
     RowHue () 25 DataBlockR  % int mod
     AltGreyHue  () 15 DataBlockR  % wis
     RowHue () 25 DataBlockR  % wis mod
     AltGreyHue  () 15 DataBlockR  % cha
     RowHue () 25 DataBlockR  % cha mod
     AltGreyHue  () 20 DataBlockR  % init mod other than dex
     RowHue () 25 DataBlockR  % init
     RowHue () 20 DataBlockR  % AC
     RowHue () 20 DataBlockR  % touch
     RowHue () 20 DataBlockR  % flat
     AltGreyHue  () 20 DataBlockR  % CMD
     RowHue () 20 DataBlockR  % spell DC
    EndDataBlock
} def
/MRnameW {
    50
} def
/BeginMatrix {
    %MHt mul RequiredVerticalSpace 
    pop
    SetColor_data
    HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
    (SKILL) MRnameW TitleBlockCtr
} def
/MatrixCNbegin {} def
/MatrixCharacterName {
%                          +---+
%                          |   |
%                     ^    |   |MCNt
%x      MCNw          x    |   |
%-->*-----------+   <-y    *---+
%   |           |MCNh
%   +-----------+
%   ^
%   |
%   y
%    /MCNh HeadingFontSize 4 add def
%    /MCNw 40 def
%    /MCNt 50 def
%    SetColor_form X Y MCNh MCNt BoxFill
%    SetColor_data 
%    gsave
%    X Y MCNh sub moveto
%    90 rotate
%    DataFontBold 14 MCNt RelativeFitText
%    grestore
%    MCNw R

%    DataFontBold 12 40 BeginDataBlock
%    GreyHue () 50 DataBlock
%    EndDataBlock
     
    grestore
} def
/MatrixCNend {} def
/MatrixRow {} def
/MatrixCell {} def
/MatrixNullCell {} def
/MatrixRowEnd {} def
/EndMatrix{} def
/BeginSpellCardBlock {
    /cardwidth PageTextWidth 2 div def
    /cardheight PageTopMargin PageBottomMargin sub 3 div def
	/SPCtype exch def
	/SPCdc0  exch def
	/SPCdcI  exch def
	/SPCdcII exch def
	/SPCdcIII exch def
	/SPCdcIV exch def
	/SPCdcV  exch def
	/SPCdcVI exch def
	/SPCdcVII exch def
	/SPCdcVIII exch def
	/SPCdcIX exch def
	/SPCdcability exch def
	/SPCdcmod exch def
} def
/EndSpellCardBlock {
} def
/cats {
     exch dup length    
     2 index length add string    
     dup dup 4 2 roll copy length
     4 -1 roll putinterval
} def 
% ref sr sv dur trg rng comp ct sch name lvl #prep #cast flags txt Spell -
/SpellCard {
	/SPCx X def
	/SPCy Y def

	SelectBodyFont
	/SPC_text_block exch def
    /SPC_flags exch def
    /SPC_nCast exch def
    /SPC_nPrep exch def
    /SPC_level exch def
	/SPC_lines 0 def
	/SPC_tb_w cardwidth 10 sub def


	SetColor_data
	/SPC_name_w cardwidth 25 sub 25 sub def
	DataFont 14 16 BeginDataBlock
	 SPC_level () 25 GreyHue DataBlockCtr
	 () SPC_name_w BlankHue DataBlock
     (SPCdc) SPC_level cats token pop exec exch pop
     /FormDBtsize 10 def
     (DC) 25 GreyHue DataBlockR
	EndDataBlock

	/SPC_range_w cardwidth 60 sub 35 sub 45 sub def
	HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
	(SCHOOL)            60 TitleBlockCtr
	(CAST)              35 TitleBlockCtr
	(COMP)              45 TitleBlockCtr
	(RANGE)    SPC_range_w TitleBlockCtr
	EndDataBlock
	DataFont 9 10 BeginDataBlock
	 () 60 GreyHue DataBlock
	 () 35 GreyHue DataBlock
	 () 45 GreyHue DataBlock
	 () SPC_range_w GreyHue DataBlock
	EndDataBlock
	/SPC_target_w cardwidth 40 sub 40 sub 20 sub 20 sub def
	HeadingFont HeadingFontSize HeadingFontSize 4 add BeginDataBlock
	(TARGET)   SPC_target_w TitleBlockCtr
	(DURATION)           40 TitleBlockCtr
	(SAVE)               40 TitleBlockCtr
	(SR)                 20 TitleBlockCtr
	(REF)                20 TitleBlockCtr
	EndDataBlock
	DataFont 9 10 BeginDataBlock
	 () SPC_target_w GreyHue DataBlock
	 () 40 GreyHue DataBlock
	 () 40 GreyHue DataBlock
	 () 20 GreyHue DataBlock
	 () 20 GreyHue DataBlock
	EndDataBlock

	SPC_text_block length 0 gt {
		SelectBodyFont
		/X SPCx def
		X 5 add Y FontLead_body 2 sub sub FontLead_body FontSize_body PsFF_init
		SPC_text_block SPC_tb_w PsFF_WaF
		/X SPCx def
		2 D
	} if

	/X SPCx def
	/Y SPCy def

	np
		X Y mv cardwidth 0 rln 0 cardheight neg rln
		cardwidth neg 0 rln cp stroke
	/X X cardwidth add def
	X PageRightMargin ge {
		/X PageLeftMargin def
		/Y Y cardheight sub def
		Y PageBottomMargin le {
			eject
		} if
	} if
} def
%
% End GMA Preamble
%
`
// @@:go:gma-preamble:end:@@

// @[00]@| Go-GMA 5.11.1
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
