/*
########################################################################################
#  _______  _______  _______                ___       ______       __    _______       #
# (  ____ \(       )(  ___  )              /   )     / ___  \     /  \  (  __   )      #
# | (    \/| () () || (   ) |             / /) |     \/   \  \    \/) ) | (  )  |      #
# | |      | || || || (___) |            / (_) (_       ___) /      | | | | /   |      #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (       | | | (/ /) |      #
# | | \_  )| |   | || (   ) | Game           ) (           ) \      | | |   / | |      #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _  __) (_|  (__) |      #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_) \____/(_______)      #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the mapper object load/save code
//

package mapper

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestObjLoadNil(t *testing.T) {
	objs, _, err := LoadMapFile(nil)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(objs) > 0 {
		t.Errorf("objs not nil: %d %v", len(objs), objs)
	}
}

func TestLegacyObjLoadImages(t *testing.T) {
	objs, meta, err := LoadMapFile(strings.NewReader(`__MAPPER__:17 {test {0 nil}}
I #SimonKazar 1.0 #SimonKazar@50.gif
I #Firefly 2.0 @OoSmGY0XERJRrA8ZiK_igg_Firefly@100
I #SimonKazar 0.25 @OoSmGY0XERJRrA8ZiK_igg_Firefly@12
`))
	if err != nil {
		t.Errorf("error %v", err)
	}
	if !reflect.DeepEqual(objs, []interface{}{
		ImageDefinition{
			Name: "#SimonKazar",
			Sizes: []ImageInstance{
				{Zoom: 1.0, File: "#SimonKazar@50.gif", IsLocalFile: true},
				{Zoom: 0.25, File: "OoSmGY0XERJRrA8ZiK_igg_Firefly@12", IsLocalFile: false},
			},
		},
		ImageDefinition{
			Name: "#Firefly",
			Sizes: []ImageInstance{
				{Zoom: 2.0, File: "OoSmGY0XERJRrA8ZiK_igg_Firefly@100", IsLocalFile: false},
			},
		},
	}) {
		t.Errorf("imgs %d %v", len(objs), objs)
	}
	if meta.Comment != "test" || meta.Location != "" || meta.Timestamp != 0 || meta.DateTime != "nil" {
		t.Errorf("meta %q", meta)
	}
}

func TestObjLoadImages(t *testing.T) {
	objs, meta, err := LoadMapFile(strings.NewReader(`__MAPPER__:20
__META__ {
    "Comment": "test",
    "Location": "somewhere",
    "Timestamp": 0,
    "DateTime": ""
}
__IMG__ {
    "Name": "#SimonKazar",
    "Sizes": [
        {
            "Zoom": 1.0,
            "File": "#SimonKazar@50.gif",
             "IsLocalFile": true
        },
        {
            "Name": "#SimonKazar",
            "Zoom": 0.25,
            "File": "OoSmGY0XERJRrA8ZiK_igg_Firefly@12"
        }
    ]
}
__IMG__ {
    "Name": "#Firefly",
    "Sizes": [
        {
            "Zoom": 2.0,
            "File": "OoSmGY0XERJRrA8ZiK_igg_Firefly@100"
        }
    ]
}
__EOF__
`))
	if err != nil {
		t.Errorf("error %v", err)
	}
	if meta.Comment != "test" || meta.Location != "somewhere" || meta.Timestamp != 0 || meta.DateTime != "" {
		t.Errorf("meta %q", meta)
	}
	if !reflect.DeepEqual(objs, []interface{}{
		ImageDefinition{
			Name: "#SimonKazar",
			Sizes: []ImageInstance{
				{Zoom: 1.0, File: "#SimonKazar@50.gif", IsLocalFile: true},
				{Zoom: 0.25, File: "OoSmGY0XERJRrA8ZiK_igg_Firefly@12", IsLocalFile: false},
			},
		},
		ImageDefinition{
			Name: "#Firefly",
			Sizes: []ImageInstance{
				{Zoom: 2.0, File: "OoSmGY0XERJRrA8ZiK_igg_Firefly@100", IsLocalFile: false},
			},
		},
	}) {
		t.Errorf("imgs %d %v", len(objs), objs)
	}
}

func TestLegacyObjLoadOnePlayer(t *testing.T) {
	expected := `__MAPPER__:20
__META__ {
    "Timestamp": 1,
    "DateTime": "n/a",
    "Comment": "no",
    "Location": "here"
}
__CREATURE__ {
    "ID": "PC73",
    "Name": "Jigu",
    "Health": {
        "MaxHP": 28,
        "LethalDamage": 6,
        "Con": 16,
        "HPBlur": 10
    },
    "Gx": 31,
    "Gy": 14,
    "Color": "blue",
    "Size": "M",
    "Area": "M",
    "Dim": true,
    "CreatureType": 2
}
__EOF__
`
	objs, meta, err := LoadMapFile(strings.NewReader(`__MAPPER__:17 {test {0 nil}}
P HEALTH:PC73 {28 6 0 16 0 0 {} 10}
P NAME:PC73 Jigu
P COLOR:PC73 blue
P GY:PC73 14
P DIM:PC73 1
P SKIN:PC73 0
P NOTE:PC73 {}
P SIZE:PC73 M
P GX:PC73 31
P AREA:PC73 M
P ELEV:PC73 0
P REACH:PC73 0
P MOVEMODE:PC73 {}
P TYPE:PC73 player
P KILLED:PC73 0
P SPAM:PC73 eggs
`))

	check := func() {
		if err != nil {
			t.Errorf("error %v", err)
		}
		if len(objs) != 1 {
			t.Fatalf("objs count: %d %v", len(objs), objs)
		}
		if reflect.TypeOf(objs[0]).Name() != "CreatureToken" {
			t.Errorf("obj type %v", reflect.TypeOf(objs[0]))
		}
		p := objs[0].(CreatureToken)
		sEq(p.ID, "PC73", "ID", t)
		sEq(p.ObjID(), "PC73", "ObjID", t)
		iEq(p.Health.MaxHP, 28, "MaxHP", t)
		if !reflect.DeepEqual(p, CreatureToken{
			CreatureType: CreatureTypePlayer,
			BaseMapObject: BaseMapObject{
				ID: "PC73",
			},
			Health: &CreatureHealth{
				MaxHP:           28,
				LethalDamage:    6,
				NonLethalDamage: 0,
				Con:             16,
				IsFlatFooted:    false,
				IsStable:        false,
				Condition:       "",
				HPBlur:          10,
			},
			Name:       "Jigu",
			Gx:         31.0,
			Gy:         14.0,
			Skin:       0,
			SkinSize:   nil,
			Elev:       0,
			Color:      "blue",
			Note:       "",
			Size:       "M",
			StatusList: nil,
			AoE:        nil,
			Area:       "M",
			MoveMode:   MoveModeLand,
			Reach:      false,
			Killed:     false,
			Dim:        true,
		}) {
			t.Errorf("jigu %q", objs)
		}

		var f strings.Builder
		err := SaveMapFile(&f, objs, MapMetaData{
			Timestamp:   1,
			DateTime:    "n/a",
			Comment:     "no",
			Location:    "here",
			FileVersion: 999,
		})
		if err != nil {
			t.Errorf("save err %v", err)
		}
		sEq(f.String(), expected, "jigu save data", t)
	}

	check()
	uEq(meta.FileVersion, 17, "file version", t)

	objs, meta, err = LoadMapFile(strings.NewReader(expected))
	check()
	uEq(meta.FileVersion, 20, "file version", t)
}

func TestObjLoadSmallMap(t *testing.T) {
	objs, _, err := LoadMapFile(strings.NewReader(`__MAPPER__:14 {{testing here} {1589430476 {Wed May 13 21:27:56 PDT 2020}}}
Z:0006bc4a7063427b8fb1f8990a24b980 1965
X:0006bc4a7063427b8fb1f8990a24b980 2100.0
Y:0006bc4a7063427b8fb1f8990a24b980 7350.0
IMAGE:0006bc4a7063427b8fb1f8990a24b980 parquet22
TYPE:0006bc4a7063427b8fb1f8990a24b980 tile
BBHEIGHT:0006bc4a7063427b8fb1f8990a24b980 57
BBWIDTH:0006bc4a7063427b8fb1f8990a24b980 100.5
LAYER:0006bc4a7063427b8fb1f8990a24b980 walls
POINTS:0006bc4a7063427b8fb1f8990a24b980 {}
FILL:0268e7eeb78e41ff82fddc4f5f0e2c1d #7e7f12
LINE:0268e7eeb78e41ff82fddc4f5f0e2c1d #7e7f12
SPLINE:0268e7eeb78e41ff82fddc4f5f0e2c1d 0
X:0268e7eeb78e41ff82fddc4f5f0e2c1d 10888.0
Y:0268e7eeb78e41ff82fddc4f5f0e2c1d 12642.0
Z:0268e7eeb78e41ff82fddc4f5f0e2c1d 271
WIDTH:0268e7eeb78e41ff82fddc4f5f0e2c1d 5
JOIN:0268e7eeb78e41ff82fddc4f5f0e2c1d bevel
TYPE:0268e7eeb78e41ff82fddc4f5f0e2c1d poly
LAYER:0268e7eeb78e41ff82fddc4f5f0e2c1d walls
POINTS:0268e7eeb78e41ff82fddc4f5f0e2c1d {10882.0 12698.0 10866.0 12738.0 10832.0 12776.0 10806.0 12816.0 10762.0 12866.0 10682.0 12928.0 10582.0 12970.0 10500.0 13024.0 10452.0 13074.0 10358.0 13100.0 10262.0 13098.0 10144.0 13060.0 10052.0 13044.0 9964.0 13044.0 9902.0 12992.0 9828.0 13010.0 9778.0 13048.0 9848.0 12282.0}
FILL:09426d492f784ad25684536c35e0d8d5 #000000
LINE:09426d492f784ad25684536c35e0d8d5 #000000
SPLINE:09426d492f784ad25684536c35e0d8d5 0
Y:09426d492f784ad25684536c35e0d8d5 13222.0
WIDTH:09426d492f784ad25684536c35e0d8d5 2
Z:09426d492f784ad25684536c35e0d8d5 25
X:09426d492f784ad25684536c35e0d8d5 9591.0
JOIN:09426d492f784ad25684536c35e0d8d5 bevel
TYPE:09426d492f784ad25684536c35e0d8d5 poly
LAYER:09426d492f784ad25684536c35e0d8d5 walls
POINTS:09426d492f784ad25684536c35e0d8d5 {9600.0 13233.0 9609.0 13249.0 9626.0 13290.0 9633.0 13310.0 9639.0 13327.0 9647.0 13356.0 9651.0 13373.0 9655.0 13403.0 9658.0 13418.0 9663.0 13445.0 9669.0 13477.0 9690.0 13517.0 9691.0 13535.0 9698.0 13553.0 9708.0 13579.0 9715.0 13605.0 9719.0 13618.0 9725.0 13641.0 9730.0 13664.0 9751.0 13690.0 9761.0 13707.0 9767.0 13721.0 9774.0 13733.0 9795.0 13752.0 9767.0 13754.0 9757.0 13743.0 9752.0 13732.0 9745.0 13720.0 9721.0 13684.0 9715.0 13672.0 9711.0 13665.0 9691.0 13637.0 9682.0 13621.0 9664.0 13596.0 9650.0 13581.0 9637.0 13564.0 9627.0 13548.0 9607.0 13524.0 9586.0 13504.0 9572.0 13492.0 9568.0 13489.0}
FILL:0c36b6a97a074bd174cda800f07206f4 #000000
LINE:0c36b6a97a074bd174cda800f07206f4 #000000
SPLINE:0c36b6a97a074bd174cda800f07206f4 0
X:0c36b6a97a074bd174cda800f07206f4 12598.0
WIDTH:0c36b6a97a074bd174cda800f07206f4 5
Y:0c36b6a97a074bd174cda800f07206f4 5697.0
Z:0c36b6a97a074bd174cda800f07206f4 57
JOIN:0c36b6a97a074bd174cda800f07206f4 bevel
TYPE:0c36b6a97a074bd174cda800f07206f4 poly
LAYER:0c36b6a97a074bd174cda800f07206f4 walls
POINTS:0c36b6a97a074bd174cda800f07206f4 {12630.0 5689.0 12646.0 5684.0 12668.0 5677.0 12694.0 5670.0 12738.0 5665.0 12793.0 5659.0 12845.0 5657.0 12900.0 5652.0 12945.0 5763.0 12901.0 5848.0 12865.0 5848.0 12814.0 5849.0 12752.0 5842.0 12716.0 5845.0 12675.0 5850.0 12647.0 5848.0 12612.0 5845.0 12581.0 5851.0}
FILL:13a2dd4a64a94e178509744e1a0a4481 #ff2600
START:13a2dd4a64a94e178509744e1a0a4481 20
EXTENT:13a2dd4a64a94e178509744e1a0a4481 225
X:13a2dd4a64a94e178509744e1a0a4481 473.0
WIDTH:13a2dd4a64a94e178509744e1a0a4481 5
Y:13a2dd4a64a94e178509744e1a0a4481 523.0
Z:13a2dd4a64a94e178509744e1a0a4481 8
ARCMODE:13a2dd4a64a94e178509744e1a0a4481 arc
TYPE:13a2dd4a64a94e178509744e1a0a4481 arc
LINE:13a2dd4a64a94e178509744e1a0a4481 black
LAYER:13a2dd4a64a94e178509744e1a0a4481 walls
POINTS:13a2dd4a64a94e178509744e1a0a4481 {321.0 669.0}
DASH:13a2dd4a64a94e178509744e1a0a4481 {}
FILL:2a1751827a954d8fad688da8e431502a #ff2600
DASH:2a1751827a954d8fad688da8e431502a -
Z:2a1751827a954d8fad688da8e431502a 1
WIDTH:2a1751827a954d8fad688da8e431502a 5
Y:2a1751827a954d8fad688da8e431502a 584.0
X:2a1751827a954d8fad688da8e431502a 675.0
LINE:2a1751827a954d8fad688da8e431502a black
ARROW:2a1751827a954d8fad688da8e431502a first
TYPE:2a1751827a954d8fad688da8e431502a line
LAYER:2a1751827a954d8fad688da8e431502a walls
POINTS:2a1751827a954d8fad688da8e431502a {779.0 291.0}
FILL:2c4a8ae53c5c4cbdb902d581402230e7 #ff2600
X:2c4a8ae53c5c4cbdb902d581402230e7 565.0
Y:2c4a8ae53c5c4cbdb902d581402230e7 707.0
Z:2c4a8ae53c5c4cbdb902d581402230e7 9
ANCHOR:2c4a8ae53c5c4cbdb902d581402230e7 center
TYPE:2c4a8ae53c5c4cbdb902d581402230e7 text
LAYER:2c4a8ae53c5c4cbdb902d581402230e7 walls
TEXT:2c4a8ae53c5c4cbdb902d581402230e7 {hello world}
FONT:2c4a8ae53c5c4cbdb902d581402230e7 {{Arial 24 normal roman}}
POINTS:2c4a8ae53c5c4cbdb902d581402230e7 {}
M ELEV:2df3e0a104614c5cb76f31836bc3f84d 0
M REACH:2df3e0a104614c5cb76f31836bc3f84d 0
M SKIN:2df3e0a104614c5cb76f31836bc3f84d 0
M DIM:2df3e0a104614c5cb76f31836bc3f84d 1
M KILLED:2df3e0a104614c5cb76f31836bc3f84d 1
M GY:2df3e0a104614c5cb76f31836bc3f84d 22
M GX:2df3e0a104614c5cb76f31836bc3f84d 27
M AREA:2df3e0a104614c5cb76f31836bc3f84d S
M SIZE:2df3e0a104614c5cb76f31836bc3f84d S
M TYPE:2df3e0a104614c5cb76f31836bc3f84d monster
M COLOR:2df3e0a104614c5cb76f31836bc3f84d red
M HEALTH:2df3e0a104614c5cb76f31836bc3f84d {9 29 0 15 0 0 {} {}}
M NAME:2df3e0a104614c5cb76f31836bc3f84d {Fleshdreg #4}
M MOVEMODE:2df3e0a104614c5cb76f31836bc3f84d {}
M NOTE:2df3e0a104614c5cb76f31836bc3f84d {}
FILL:38f633da2d6749467f5406f187b8cc3f #000000
X:38f633da2d6749467f5406f187b8cc3f 10810.0
Z:38f633da2d6749467f5406f187b8cc3f 12
Y:38f633da2d6749467f5406f187b8cc3f 14350.0
WIDTH:38f633da2d6749467f5406f187b8cc3f 5
LINE:38f633da2d6749467f5406f187b8cc3f black
TYPE:38f633da2d6749467f5406f187b8cc3f line
LAYER:38f633da2d6749467f5406f187b8cc3f walls
POINTS:38f633da2d6749467f5406f187b8cc3f {10908.0 14396.0}
FILL:39880f0c6e904bf9a866d9af8783fd70 #ff2600
Z:39880f0c6e904bf9a866d9af8783fd70 3
Y:39880f0c6e904bf9a866d9af8783fd70 33.0
X:39880f0c6e904bf9a866d9af8783fd70 445.5
WIDTH:39880f0c6e904bf9a866d9af8783fd70 5
LOCKED:39880f0c6e904bf9a866d9af8783fd70 1
LINE:39880f0c6e904bf9a866d9af8783fd70 black
TYPE:39880f0c6e904bf9a866d9af8783fd70 rect
LAYER:39880f0c6e904bf9a866d9af8783fd70 walls
POINTS:39880f0c6e904bf9a866d9af8783fd70 {625.0 160.0}
DASH:39880f0c6e904bf9a866d9af8783fd70 {}
M ELEV:39a1afc1b1aa4cac87eee32be93ebe9a 0
M KILLED:39a1afc1b1aa4cac87eee32be93ebe9a 0
M SKIN:39a1afc1b1aa4cac87eee32be93ebe9a 0
M DIM:39a1afc1b1aa4cac87eee32be93ebe9a 1
M REACH:39a1afc1b1aa4cac87eee32be93ebe9a 1
M GX:39a1afc1b1aa4cac87eee32be93ebe9a 19
M GY:39a1afc1b1aa4cac87eee32be93ebe9a 19
M AREA:39a1afc1b1aa4cac87eee32be93ebe9a M
M SIZE:39a1afc1b1aa4cac87eee32be93ebe9a M
M NAME:39a1afc1b1aa4cac87eee32be93ebe9a barbarian2=Caroll
M TYPE:39a1afc1b1aa4cac87eee32be93ebe9a monster
M COLOR:39a1afc1b1aa4cac87eee32be93ebe9a red
M STATUSLIST:39a1afc1b1aa4cac87eee32be93ebe9a stable
M HEALTH:39a1afc1b1aa4cac87eee32be93ebe9a {45 56 0 14 0 0 {} 0}
M MOVEMODE:39a1afc1b1aa4cac87eee32be93ebe9a {}
M NOTE:39a1afc1b1aa4cac87eee32be93ebe9a {}
X:3f5b6a2655214928b868daad9a97db4d 110.0
Y:3f5b6a2655214928b868daad9a97db4d 18.0
WIDTH:3f5b6a2655214928b868daad9a97db4d 5
Z:3f5b6a2655214928b868daad9a97db4d 5
LINE:3f5b6a2655214928b868daad9a97db4d black
TYPE:3f5b6a2655214928b868daad9a97db4d circ
LAYER:3f5b6a2655214928b868daad9a97db4d walls
POINTS:3f5b6a2655214928b868daad9a97db4d {237.0 150.0}
DASH:3f5b6a2655214928b868daad9a97db4d {}
FILL:3f5b6a2655214928b868daad9a97db4d {}
Y:4b82e91b987d412b9c1a2b5110319072 400.0
WIDTH:4b82e91b987d412b9c1a2b5110319072 5
X:4b82e91b987d412b9c1a2b5110319072 500.0
Z:4b82e91b987d412b9c1a2b5110319072 99999999
TYPE:4b82e91b987d412b9c1a2b5110319072 aoe
FILL:4b82e91b987d412b9c1a2b5110319072 black
LINE:4b82e91b987d412b9c1a2b5110319072 black
AOESHAPE:4b82e91b987d412b9c1a2b5110319072 radius
LAYER:4b82e91b987d412b9c1a2b5110319072 walls
POINTS:4b82e91b987d412b9c1a2b5110319072 {500.0 150.0}
DASH:4b82e91b987d412b9c1a2b5110319072 {}
FILL:5b1a4fa32af54eb2b35ebec0a4c88089 #ff2600
START:5b1a4fa32af54eb2b35ebec0a4c88089 140
Y:5b1a4fa32af54eb2b35ebec0a4c88089 309.0
WIDTH:5b1a4fa32af54eb2b35ebec0a4c88089 5
X:5b1a4fa32af54eb2b35ebec0a4c88089 59.0
Z:5b1a4fa32af54eb2b35ebec0a4c88089 6
EXTENT:5b1a4fa32af54eb2b35ebec0a4c88089 61
TYPE:5b1a4fa32af54eb2b35ebec0a4c88089 arc
LINE:5b1a4fa32af54eb2b35ebec0a4c88089 black
ARCMODE:5b1a4fa32af54eb2b35ebec0a4c88089 pieslice
LAYER:5b1a4fa32af54eb2b35ebec0a4c88089 walls
POINTS:5b1a4fa32af54eb2b35ebec0a4c88089 {161.0 485.0}
DASH:5b1a4fa32af54eb2b35ebec0a4c88089 {}
FILL:61dc2ff4efe54be7a18791b338c29c5c #ff2600
DASH:61dc2ff4efe54be7a18791b338c29c5c -
Z:61dc2ff4efe54be7a18791b338c29c5c 2
Y:61dc2ff4efe54be7a18791b338c29c5c 229.0
WIDTH:61dc2ff4efe54be7a18791b338c29c5c 5
X:61dc2ff4efe54be7a18791b338c29c5c 604.0
LINE:61dc2ff4efe54be7a18791b338c29c5c black
ARROW:61dc2ff4efe54be7a18791b338c29c5c first
TYPE:61dc2ff4efe54be7a18791b338c29c5c line
LAYER:61dc2ff4efe54be7a18791b338c29c5c walls
POINTS:61dc2ff4efe54be7a18791b338c29c5c {509.0 378.0 650.0 360.0}
FILL:7b39f6dbeea44b8baa20032c443a0654 #ff2600
X:7b39f6dbeea44b8baa20032c443a0654 126.0
START:7b39f6dbeea44b8baa20032c443a0654 151
EXTENT:7b39f6dbeea44b8baa20032c443a0654 235
Y:7b39f6dbeea44b8baa20032c443a0654 274.0
WIDTH:7b39f6dbeea44b8baa20032c443a0654 5
Z:7b39f6dbeea44b8baa20032c443a0654 7
TYPE:7b39f6dbeea44b8baa20032c443a0654 arc
LINE:7b39f6dbeea44b8baa20032c443a0654 black
ARCMODE:7b39f6dbeea44b8baa20032c443a0654 chord
LAYER:7b39f6dbeea44b8baa20032c443a0654 walls
POINTS:7b39f6dbeea44b8baa20032c443a0654 {166.0 375.0}
DASH:7b39f6dbeea44b8baa20032c443a0654 {}
P DIM:976e7148ae86409f99fdebf83f3f0904 0
P KILLED:976e7148ae86409f99fdebf83f3f0904 0
P REACH:976e7148ae86409f99fdebf83f3f0904 0
P SKIN:976e7148ae86409f99fdebf83f3f0904 1
P ELEV:976e7148ae86409f99fdebf83f3f0904 20
P GX:976e7148ae86409f99fdebf83f3f0904 6
P GY:976e7148ae86409f99fdebf83f3f0904 6
P NAME:976e7148ae86409f99fdebf83f3f0904 Jigu2
P AREA:976e7148ae86409f99fdebf83f3f0904 M
P SIZE:976e7148ae86409f99fdebf83f3f0904 M
P MOVEMODE:976e7148ae86409f99fdebf83f3f0904 fly
P COLOR:976e7148ae86409f99fdebf83f3f0904 green
P TYPE:976e7148ae86409f99fdebf83f3f0904 player
P STATUSLIST:976e7148ae86409f99fdebf83f3f0904 {confused exhausted nauseated}
P AOE:976e7148ae86409f99fdebf83f3f0904 {radius 2 black}
P NOTE:976e7148ae86409f99fdebf83f3f0904 {spam spam}
P HEALTH:976e7148ae86409f99fdebf83f3f0904 {}
P KILLED:PC73 0
P REACH:PC73 0
P DIM:PC73 1
P SKIN:PC73 1
P GY:PC73 14
P ELEV:PC73 30
P GX:PC73 31
P NAME:PC73 Jigu
P AREA:PC73 M
P SIZE:PC73 M
P COLOR:PC73 blue
P MOVEMODE:PC73 fly
P TYPE:PC73 player
P HEALTH:PC73 {28 6 1 16 0 0 surprised}
P SKINSIZE:PC73 {M L}
P NOTE:PC73 {Mirror Image 2}
FILL:afd136735d7e400082f331485e73f7a1 #00f900
X:afd136735d7e400082f331485e73f7a1 150.0
WIDTH:afd136735d7e400082f331485e73f7a1 5
Y:afd136735d7e400082f331485e73f7a1 600.0
Z:afd136735d7e400082f331485e73f7a1 99999999
TYPE:afd136735d7e400082f331485e73f7a1 aoe
LINE:afd136735d7e400082f331485e73f7a1 black
AOESHAPE:afd136735d7e400082f331485e73f7a1 ray
LAYER:afd136735d7e400082f331485e73f7a1 walls
POINTS:afd136735d7e400082f331485e73f7a1 {200.0 200.0}
DASH:afd136735d7e400082f331485e73f7a1 {}
X:bf29dfa85cc54498bb33a2d7523d9edc 289.0
Y:bf29dfa85cc54498bb33a2d7523d9edc 36.0
Z:bf29dfa85cc54498bb33a2d7523d9edc 4
WIDTH:bf29dfa85cc54498bb33a2d7523d9edc 5
LINE:bf29dfa85cc54498bb33a2d7523d9edc black
TYPE:bf29dfa85cc54498bb33a2d7523d9edc rect
LAYER:bf29dfa85cc54498bb33a2d7523d9edc walls
POINTS:bf29dfa85cc54498bb33a2d7523d9edc {355.0 97.0}
DASH:bf29dfa85cc54498bb33a2d7523d9edc {}
FILL:bf29dfa85cc54498bb33a2d7523d9edc {}
FILL:e68d5354f175401582866a75d806d8d7 #00f900
WIDTH:e68d5354f175401582866a75d806d8d7 5
Y:e68d5354f175401582866a75d806d8d7 800.0
X:e68d5354f175401582866a75d806d8d7 850.0
Z:e68d5354f175401582866a75d806d8d7 99999999
TYPE:e68d5354f175401582866a75d806d8d7 aoe
LINE:e68d5354f175401582866a75d806d8d7 black
AOESHAPE:e68d5354f175401582866a75d806d8d7 cone
LAYER:e68d5354f175401582866a75d806d8d7 walls
POINTS:e68d5354f175401582866a75d806d8d7 {450.0 800.0}
DASH:e68d5354f175401582866a75d806d8d7 {}
`))
	expected_save := `
__MAPPER__:20
__META__ {
	"Timestamp": 1,
	"DateTime": "n/a",
	"Comment": "no",
	"Location": "here",
	"FileVersion": 999
}
__TILE__ {
	"ID": "0006bc4a7063427b8fb1f8990a24b980",
	"X": 2100,
	"Y": 7350,
	"Z": 1965,
	"Layer": "walls",
	"Image": "parquet22",
	"BBHeight": 57,
	"BBWidth": 100.5,
}
__POLY__ {
	"ID": "0268e7eeb78e41ff82fddc4f5f0e2c1d",
	"X": 10888,
	"Y": 12642,
	"Z": 271,
	"Points": [
		{"X": 10882, "Y": 12698},
		{"X": 10866, "Y": 12738},
		{"X": 10832, "Y": 12776},
		{"X": 10806, "Y": 12816},
		{"X": 10762, "Y": 12866},
		{"X": 10682, "Y": 12928},
		{"X": 10582, "Y": 12970},
		{"X": 10500, "Y": 13024},
		{"X": 10452, "Y": 13074},
		{"X": 10358, "Y": 13100},
		{"X": 10262, "Y": 13098},
		{"X": 10144, "Y": 13060},
		{"X": 10052, "Y": 13044},
		{"X": 9964, "Y": 13044},
		{"X": 9902, "Y": 12992},
		{"X": 9828, "Y": 13010},
		{"X": 9778, "Y": 13048},
		{"X": 9848, "Y": 12282}
	],
	"Width": 5,
	"Layer": "walls",
	"Fill": "#7e7f12",
	"Line": "#7e7f12"
}
__POLY__ {
	"ID": "09426d492f784ad25684536c35e0d8d5",
	"X": 9591,
	"Y": 13222,
	"Z": 25,
	"POINTS": [
		{"X": 9600, "Y": 13233},
		{"X": 9609, "Y": 13249}, 
		{"X": 9626, "Y": 13290}, 
		{"X": 9633, "Y": 13310}, 
		{"X": 9639, "Y": 13327}, 
		{"X": 9647, "Y": 13356}, 
		{"X": 9651, "Y": 13373}, 
		{"X": 9655, "Y": 13403}, 
		{"X": 9658, "Y": 13418}, 
		{"X": 9663, "Y": 13445}, 
		{"X": 9669, "Y": 13477}, 
		{"X": 9690, "Y": 13517}, 
		{"X": 9691, "Y": 13535}, 
		{"X": 9698, "Y": 13553}, 
		{"X": 9708, "Y": 13579}, 
		{"X": 9715, "Y": 13605}, 
		{"X": 9719, "Y": 13618}, 
		{"X": 9725, "Y": 13641}, 
		{"X": 9730, "Y": 13664}, 
		{"X": 9751, "Y": 13690}, 
		{"X": 9761, "Y": 13707}, 
		{"X": 9767, "Y": 13721}, 
		{"X": 9774, "Y": 13733}, 
		{"X": 9795, "Y": 13752}, 
		{"X": 9767, "Y": 13754}, 
		{"X": 9757, "Y": 13743}, 
		{"X": 9752, "Y": 13732}, 
		{"X": 9745, "Y": 13720}, 
		{"X": 9721, "Y": 13684}, 
		{"X": 9715, "Y": 13672}, 
		{"X": 9711, "Y": 13665}, 
		{"X": 9691, "Y": 13637}, 
		{"X": 9682, "Y": 13621}, 
		{"X": 9664, "Y": 13596}, 
		{"X": 9650, "Y": 13581}, 
		{"X": 9637, "Y": 13564}, 
		{"X": 9627, "Y": 13548}, 
		{"X": 9607, "Y": 13524}, 
		{"X": 9586, "Y": 13504}, 
		{"X": 9572, "Y": 13492}, 
		{"X": 9568, "Y": 13489}
	],
	"WIDTH": 2,
	"LAYER": "walls",
	"FILL": "#000000",
	"LINE": "#000000"
}
__POLY__ {
	"ID": "0c36b6a97a074bd174cda800f07206f4",
	"X": 12598,
	"Y": 5697,
	"Z": 57,
	"Points": [
		{"X": 12630, "Y": 5689}, 
		{"X": 12646, "Y": 5684}, 
		{"X": 12668, "Y": 5677}, 
		{"X": 12694, "Y": 5670}, 
		{"X": 12738, "Y": 5665}, 
		{"X": 12793, "Y": 5659}, 
		{"X": 12845, "Y": 5657}, 
		{"X": 12900, "Y": 5652}, 
		{"X": 12945, "Y": 5763}, 
		{"X": 12901, "Y": 5848}, 
		{"X": 12865, "Y": 5848}, 
		{"X": 12814, "Y": 5849}, 
		{"X": 12752, "Y": 5842}, 
		{"X": 12716, "Y": 5845}, 
		{"X": 12675, "Y": 5850}, 
		{"X": 12647, "Y": 5848}, 
		{"X": 12612, "Y": 5845}, 
		{"X": 12581, "Y": 5851}
	],
	"Width": 5,
	"Layer": "walls",
	"Fill": "#000000",
	"Line": "#000000"
}
__ARC__ {
	"ID": "13a2dd4a64a94e178509744e1a0a4481",
	"X": 473,
	"Y": 523,
	"Z": 8,
	"Points": [
		{"X": 321, "Y": 669}
	],
	"ArcMode": 1,
	"Line": "black",
	"Fill": "#ff2600",
	"Layer": "walls",
	"Width": 5,
	"Start": 20,
	"Extent": 225
}
__LINE__ {
	"ID": "2a1751827a954d8fad688da8e431502a",
	"X": 675,
	"Y": 584,
	"Z": 1,
	"Width": 5,
	"Dash": 1,
	"Line": "black",
	"Arrow": 1,
	"Layer": "walls",
	"Fill": "#ff2600",
	"Points": [
		{"X": 779, "Y": 291}
	]
}
__TEXT__ {
	"ID": "2c4a8ae53c5c4cbdb902d581402230e7",
	"X": 565,
	"Y": 707,
	"Z": 9,
	"Layer": "walls",
	"Fill": "#ff2600",
	"Text": "hello world",
	"Font": {
		"Family": "Arial",
		"Size": 24
	}
}
__CREATURE__ {
	"ID": "2df3e0a104614c5cb76f31836bc3f84d",
	"Gx": 27,
	"Gy": 22,
	"Size": "S",
	"Area": "S",
	"Color": "red",
	"CreatureType: 1,
	"Name": "Fleshdreg #4",
	"Dim": true,
	"Killed": true,
	"Health": {
		"MaxHP": 9,
		"LethalDamage": 29,
		"Con": 15
	}
}
__LINE__ {
	"ID": "38f633da2d6749467f5406f187b8cc3f",
	"X": 10810,
	"Y": 14350,
	"Z": 12,
	"Points": [
		{"X": 10908, "Y": 14396}
	],
	"Width": 5,
	"Line": "black",
	"Layer": "walls",
	"Fill": "#000000"
}
__RECT__ {
	"ID": "39880f0c6e904bf9a866d9af8783fd70",
	"X": 445.5,
	"Y": 33,
	"Z": 3,
	"Points": [
		{"X": 625, "Y": 160}
	],
	"Line": "black",
	"Fill": "#ff2600",
	"Layer": "walls",
	"Width": 5,
	"Locked": 1
}
__CREATURE__ {
	"ID": "39a1afc1b1aa4cac87eee32be93ebe9a",
	"Gx: 19,
	"Gy: 19,
	"Area: "M",
	"Size: "M",
	"Dim": true,
	"Reach": true,
	"Name": "barbarian2=Caroll",
	"CreatureType": 1,
	"Color": red
	"StatusList": [
		"stable"
	],
	"CreatureHealth": {
		"MaxHP": 45,
		"LethalDamage": 56,
		"Con": 14
	}
}
__CIRC__ {
	"ID": "3f5b6a2655214928b868daad9a97db4d",
	"X": 110,
	"Y": 18,
	"Z": 5,
	"Points": [
		{"X": 237, "Y": 150}
	],
	"Width": 5,
	"Line": "black",
	"Layer": "walls"
}
__SAOE__ {
	"ID": "4b82e91b987d412b9c1a2b5110319072",
	"X": 500,
	"Y": 400,
	"Z": 99999999,
	"Points": [
		{"X": 500, "Y": 150}
	],
	"Width": 5,
	"Fill": "black",
	"Line": "black",
	"AoEShape": 1,
	"Layer": "walls"
}
__ARC__ {
	"ID": "5b1a4fa32af54eb2b35ebec0a4c88089",
	"X": 59,
	"Y": 309,
	"Z": 6,
	"POINTS": [
		{"X": 161, "Y": 485}
	],
	"WIDTH": 5,
	"LINE": "black",
	"LAYER": "walls",
	"FILL": "#ff2600",
	"START": 140,
	"EXTENT": 61
}
__LINE__ {
	"ID": "61dc2ff4efe54be7a18791b338c29c5c",
	"X": 604,
	"Y": 229,
	"Z": 2,
	"Points": [
		{"X": 509, "Y": 378},
		{"X": 650, "Y": 360}
	],
	"Dash": 1,
	"Width": 5,
	"Line": "black",
	"Arrow": 1,
	"Layer": "walls",
	"Fill": "#ff2600"
}
__ARC__ {
	"ID": "7b39f6dbeea44b8baa20032c443a0654",
	X: 126,
	Y: 274,
	Z: 7,
	POINTS: [
		{"X": 166, "Y": 375}
	],
	WIDTH: 5,
	LINE: "black",
	ARCMODE: 2,
	LAYER: "walls",
	FILL: "#ff2600",
	START: 151,
	EXTENT: 235
}
__CREATURE__ {
	"ID": "976e7148ae86409f99fdebf83f3f0904",
	"Gx": 6,
	"Gy": 6,
	"Skin": 1,
	"Elev": "20",
	"Name": "Jigu2",
	"Area": "M",
	"Size": "M",
	"MoveMode": 3,
	"Color": "green",
	"CreatureType": 2,
	"StatusList": [
		"confused",
		"exhausted",
		"nauseated"
	],
	"AoE": {
		"Radius": 2,
		"Color": "black"
	},
	"Note": "spam spam"
}
__CREATURE__ {
	"ID": "PC73",
	"Name": "Jigu",
	"Gx": 31,
	"Gy": 14,
	"Dim": true,
	"Skin": 1,
	"Elev": 30,
	"Area": "M",
	"Size": "M",
	"Color": "blue",
	"MoveMode": 3,
	"CreatureType": 2,
	"Health": {
		"MaxHP": 28,
		"LethalDamage": 6,
		"NonLethalDamage": 1,
		"Con": 16,
		"Condition": "surprised"
	},
	"SkinSize": [
		"M",
		"L"
	],
	"Note": "Mirror Image 2"
}
__SAOE__ {
	"ID": "afd136735d7e400082f331485e73f7a1",
	"X": 150,
	"Y": 600,
	"Z": 99999999,
	"Points": [
		{"X": 200, "Y": 200}
	],
	"Width": 5,
	"Line": "black"
	"AoEShape": 2,
	"Layer": "walls",
	"Fill": "#00f900"
}
__RECT__ {
	"ID": "bf29dfa85cc54498bb33a2d7523d9edc",
	"X": 289,
	"Y": 36,
	"Z": 4,
	"Width": 5,
	"Line": black,
	"Layer": walls,
	"Points": [
		{"X": 355, "Y": 97}
	]
}
__SAOE__ {
	"ID": "e68d5354f175401582866a75d806d8d7",
	"Width": 5,
	"Y": 800,
	"X": 850,
	"Z": 99999999,
	"Line": "black",
	"Layer": "walls",
	"Fill": "#00f900",
	"Points": [
		{"X": 450, "Y": 800}
	]
}
__EOF__
`
	check := func() {
		if err != nil {
			t.Errorf("error %v", err)
		}
		IDsFound := make([]string, 0, 10)
		for _, o := range objs {
			IDsFound = append(IDsFound, o.(MapObject).ObjID())
			switch obj := o.(type) {
			case CreatureToken:
				switch obj.ObjID() {
				case "2df3e0a104614c5cb76f31836bc3f84d":
					sEq(obj.ID, "2df3e0a104614c5cb76f31836bc3f84d", "ID", t)
					iEq(obj.Health.MaxHP, 9, "MaxHP", t)
					iEq(obj.Health.LethalDamage, 29, "LethalDamage", t)
					iEq(obj.Health.NonLethalDamage, 0, "NonLethalDamage", t)
					iEq(obj.Health.Con, 15, "Con", t)
					bEq(obj.Health.IsFlatFooted, false, "IsFlatFooted", t)
					bEq(obj.Health.IsStable, false, "IsStable", t)
					sEq(obj.Health.Condition, "", "Condition", t)
					iEq(obj.Health.HPBlur, 0, "HPBlur", t)
					sEq(obj.Name, "Fleshdreg #4", "Name", t)
					fEq(obj.Gx, 27.0, "Gx", t)
					fEq(obj.Gy, 22.0, "Gy", t)
					iEq(obj.Skin, 0, "Skin", t)
					SEq(obj.SkinSize, nil, "SkinSize", t)
					iEq(obj.Elev, 0, "Elev", t)
					sEq(obj.Color, "red", "Color", t)
					sEq(obj.Note, "", "Note", t)
					sEq(obj.Size, "S", "Size", t)
					SEq(obj.StatusList, nil, "StatusList", t)
					if obj.AoE != nil {
						t.Errorf("AoE expected to be nil but wasn't")
					}
					sEq(obj.Area, "S", "Area", t)
					BEq(byte(obj.MoveMode), byte(MoveModeLand), "MoveMode", t)
					bEq(obj.Reach, false, "Reach", t)
					bEq(obj.Killed, true, "Killed", t)
					bEq(obj.Dim, true, "Dim", t)
					BEq(byte(obj.CreatureType), byte(CreatureTypeMonster), "CreatureType", t)
				case "39a1afc1b1aa4cac87eee32be93ebe9a":
					sEq(obj.ID, "39a1afc1b1aa4cac87eee32be93ebe9a", "ID", t)
					iEq(obj.Health.MaxHP, 45, "MaxHP", t)
					iEq(obj.Health.LethalDamage, 56, "LethalDamage", t)
					iEq(obj.Health.NonLethalDamage, 0, "NonLethalDamage", t)
					iEq(obj.Health.Con, 14, "Con", t)
					bEq(obj.Health.IsFlatFooted, false, "IsFlatFooted", t)
					bEq(obj.Health.IsStable, false, "IsStable", t)
					sEq(obj.Health.Condition, "", "Condition", t)
					iEq(obj.Health.HPBlur, 0, "HPBlur", t)
					sEq(obj.Name, "barbarian2=Caroll", "Name", t)
					fEq(obj.Gx, 19.0, "Gx", t)
					fEq(obj.Gy, 19.0, "Gy", t)
					iEq(obj.Skin, 0, "Skin", t)
					SEq(obj.SkinSize, nil, "SkinSize", t)
					iEq(obj.Elev, 0, "Elev", t)
					sEq(obj.Color, "red", "Color", t)
					sEq(obj.Note, "", "Note", t)
					sEq(obj.Size, "M", "Size", t)
					SEq(obj.StatusList, []string{"stable"}, "StatusList", t)
					if obj.AoE != nil {
						t.Errorf("AoE expected to be nil but wasn't")
					}
					sEq(obj.Area, "M", "Area", t)
					BEq(byte(obj.MoveMode), byte(MoveModeLand), "MoveMode", t)
					bEq(obj.Reach, true, "Reach", t)
					bEq(obj.Killed, false, "Killed", t)
					bEq(obj.Dim, true, "Dim", t)
					BEq(byte(obj.CreatureType), byte(CreatureTypeMonster), "CreatureType", t)
				case "PC73":
					sEq(obj.ID, "PC73", "ID", t)
					iEq(obj.Health.MaxHP, 28, "MaxHP", t)
					iEq(obj.Health.LethalDamage, 6, "LethalDamage", t)
					iEq(obj.Health.NonLethalDamage, 1, "NonLethalDamage", t)
					iEq(obj.Health.Con, 16, "Con", t)
					bEq(obj.Health.IsFlatFooted, false, "IsFlatFooted", t)
					bEq(obj.Health.IsStable, false, "IsStable", t)
					sEq(obj.Health.Condition, "surprised", "Condition", t)
					iEq(obj.Health.HPBlur, 0, "HPBlur", t)
					sEq(obj.Name, "Jigu", "Name", t)
					fEq(obj.Gx, 31.0, "Gx", t)
					fEq(obj.Gy, 14.0, "Gy", t)
					iEq(obj.Skin, 1, "Skin", t)
					SEq(obj.SkinSize, []string{"M", "L"}, "SkinSize", t)
					iEq(obj.Elev, 30, "Elev", t)
					sEq(obj.Color, "blue", "Color", t)
					sEq(obj.Note, "Mirror Image 2", "Note", t)
					sEq(obj.Size, "M", "Size", t)
					SEq(obj.StatusList, nil, "StatusList", t)
					if obj.AoE != nil {
						t.Errorf("AoE expected to be nil but wasn't")
					}
					sEq(obj.Area, "M", "Area", t)
					BEq(byte(obj.MoveMode), byte(MoveModeFly), "MoveMode", t)
					bEq(obj.Reach, false, "Reach", t)
					bEq(obj.Killed, false, "Killed", t)
					bEq(obj.Dim, true, "Dim", t)
					BEq(byte(obj.CreatureType), byte(CreatureTypePlayer), "CreatureType", t)
				case "976e7148ae86409f99fdebf83f3f0904":
					sEq(obj.ID, "976e7148ae86409f99fdebf83f3f0904", "ID", t)
					if obj.Health != nil {
						t.Errorf("expected health to be nil, but wasn't.")
					}
					sEq(obj.Name, "Jigu2", "Name", t)
					fEq(obj.Gx, 6.0, "Gx", t)
					fEq(obj.Gy, 6.0, "Gy", t)
					iEq(obj.Skin, 1, "Skin", t)
					SEq(obj.SkinSize, nil, "SkinSize", t)
					iEq(obj.Elev, 20, "Elev", t)
					sEq(obj.Color, "green", "Color", t)
					sEq(obj.Note, "spam spam", "Note", t)
					sEq(obj.Size, "M", "Size", t)
					SEq(obj.StatusList, []string{"confused", "exhausted", "nauseated"}, "StatusList", t)
					fEq(obj.AoE.Radius, 2.0, "Radius", t)
					sEq(obj.AoE.Color, "black", "Aoe Color", t)
					sEq(obj.Area, "M", "Area", t)
					BEq(byte(obj.MoveMode), byte(MoveModeFly), "MoveMode", t)
					bEq(obj.Reach, false, "Reach", t)
					bEq(obj.Killed, false, "Killed", t)
					bEq(obj.Dim, false, "Dim", t)
					BEq(byte(obj.CreatureType), byte(CreatureTypePlayer), "CreatureType", t)
				default:
					t.Errorf("Found unexpected creature token %s", obj.ObjID())
				}
			case TileElement:
				switch obj.ObjID() {
				case "0006bc4a7063427b8fb1f8990a24b980":
					// BaseMapObject
					sEq(obj.ID, "0006bc4a7063427b8fb1f8990a24b980", "ID", t)
					// Coordinates
					fEq(obj.X, 2100, "tile X", t)
					fEq(obj.Y, 7350, "tile Y", t)
					// MapElement
					iEq(obj.Z, 1965, "tile Z", t)
					bEq(obj.Locked, false, "tile Locked", t)
					CEq(obj.Points, []Coordinates{}, "tile Points", t)
					sEq(obj.Fill, "", "tile Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "tile Dash", t)
					sEq(obj.Line, "", "tile Line", t)
					iEq(obj.Width, 0, "tile Width", t)
					sEq(obj.Layer, "walls", "tile Layer", t)
					bEq(obj.Hidden, false, "tile Hidden", t)
					iEq(obj.Level, 0, "tile Level", t)
					sEq(obj.Group, "", "tile Group", t)
					// TileElement
					sEq(obj.Image, "parquet22", "tile Image", t)
					fEq(obj.BBHeight, 57.0, "tile height", t)
					fEq(obj.BBWidth, 100.5, "tile width", t)
				default:
					t.Errorf("Found unexpected tile %s", obj.ObjID())
				}

			case PolygonElement:
				switch obj.ObjID() {
				case "0268e7eeb78e41ff82fddc4f5f0e2c1d":
					// BaseMapObject
					sEq(obj.ID, "0268e7eeb78e41ff82fddc4f5f0e2c1d", "ID", t)
					// Coordinates
					fEq(obj.X, 10888, "poly 0268 X", t)
					fEq(obj.Y, 12642, "poly 0268 Y", t)
					// MapElement
					iEq(obj.Z, 271, "poly 0268 Z", t)
					bEq(obj.Locked, false, "poly 0268 Locked", t)
					CEq(obj.Points, []Coordinates{
						{10882.0, 12698.0},
						{10866.0, 12738.0},
						{10832.0, 12776.0},
						{10806.0, 12816.0},
						{10762.0, 12866.0},
						{10682.0, 12928.0},
						{10582.0, 12970.0},
						{10500.0, 13024.0},
						{10452.0, 13074.0},
						{10358.0, 13100.0},
						{10262.0, 13098.0},
						{10144.0, 13060.0},
						{10052.0, 13044.0},
						{9964.0, 13044.0},
						{9902.0, 12992.0},
						{9828.0, 13010.0},
						{9778.0, 13048.0},
						{9848.0, 12282.0},
					}, "poly 0268 Points", t)
					sEq(obj.Fill, "#7e7f12", "poly 0268 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "poly 0268 Dash", t)
					sEq(obj.Line, "#7e7f12", "poly 0268 Line", t)
					iEq(obj.Width, 5, "poly 0268 Width", t)
					sEq(obj.Layer, "walls", "poly 0268 Layer", t)
					bEq(obj.Hidden, false, "poly 0268 Hidden", t)
					iEq(obj.Level, 0, "poly 0268 Level", t)
					sEq(obj.Group, "", "poly 0268 Group", t)
					// PolygonElement
					fEq(obj.Spline, 0, "poly 0268 Spline", t)
					BEq(byte(obj.Join), byte(JoinBevel), "poly 0268 Join", t)
				case "09426d492f784ad25684536c35e0d8d5":
					// BaseMapObject
					sEq(obj.ID, "09426d492f784ad25684536c35e0d8d5", "ID", t)
					// Coordinates
					fEq(obj.X, 9591, "poly 0942 X", t)
					fEq(obj.Y, 13222, "poly 0942 Y", t)
					// MapElement
					iEq(obj.Z, 25, "poly 0942 Z", t)
					bEq(obj.Locked, false, "poly 0942 Locked", t)
					CEq(obj.Points, []Coordinates{
						{9600.0, 13233.0},
						{9609.0, 13249.0},
						{9626.0, 13290.0},
						{9633.0, 13310.0},
						{9639.0, 13327.0},
						{9647.0, 13356.0},
						{9651.0, 13373.0},
						{9655.0, 13403.0},
						{9658.0, 13418.0},
						{9663.0, 13445.0},
						{9669.0, 13477.0},
						{9690.0, 13517.0},
						{9691.0, 13535.0},
						{9698.0, 13553.0},
						{9708.0, 13579.0},
						{9715.0, 13605.0},
						{9719.0, 13618.0},
						{9725.0, 13641.0},
						{9730.0, 13664.0},
						{9751.0, 13690.0},
						{9761.0, 13707.0},
						{9767.0, 13721.0},
						{9774.0, 13733.0},
						{9795.0, 13752.0},
						{9767.0, 13754.0},
						{9757.0, 13743.0},
						{9752.0, 13732.0},
						{9745.0, 13720.0},
						{9721.0, 13684.0},
						{9715.0, 13672.0},
						{9711.0, 13665.0},
						{9691.0, 13637.0},
						{9682.0, 13621.0},
						{9664.0, 13596.0},
						{9650.0, 13581.0},
						{9637.0, 13564.0},
						{9627.0, 13548.0},
						{9607.0, 13524.0},
						{9586.0, 13504.0},
						{9572.0, 13492.0},
						{9568.0, 13489.0},
					}, "poly 0942 Points", t)
					sEq(obj.Fill, "#000000", "poly 0942 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "poly 0942 Dash", t)
					sEq(obj.Line, "#000000", "poly 0942 Line", t)
					iEq(obj.Width, 2, "poly 0942 Width", t)
					sEq(obj.Layer, "walls", "poly 0942 Layer", t)
					bEq(obj.Hidden, false, "poly 0942 Hidden", t)
					iEq(obj.Level, 0, "poly 0942 Level", t)
					sEq(obj.Group, "", "poly 0942 Group", t)
					// PolygonElement
					fEq(obj.Spline, 0, "poly 0942 Spline", t)
					BEq(byte(obj.Join), byte(JoinBevel), "poly 0942 Join", t)
				case "0c36b6a97a074bd174cda800f07206f4":
					// BaseMapObject
					sEq(obj.ID, "0c36b6a97a074bd174cda800f07206f4", "ID", t)
					// Coordinates
					fEq(obj.X, 12598, "poly 0c36 X", t)
					fEq(obj.Y, 5697, "poly 0c36 Y", t)
					// MapElement
					iEq(obj.Z, 57, "poly 0c36 Z", t)
					bEq(obj.Locked, false, "poly 0c36 Locked", t)
					CEq(obj.Points, []Coordinates{
						{12630.0, 5689.0},
						{12646.0, 5684.0},
						{12668.0, 5677.0},
						{12694.0, 5670.0},
						{12738.0, 5665.0},
						{12793.0, 5659.0},
						{12845.0, 5657.0},
						{12900.0, 5652.0},
						{12945.0, 5763.0},
						{12901.0, 5848.0},
						{12865.0, 5848.0},
						{12814.0, 5849.0},
						{12752.0, 5842.0},
						{12716.0, 5845.0},
						{12675.0, 5850.0},
						{12647.0, 5848.0},
						{12612.0, 5845.0},
						{12581.0, 5851.0},
					}, "poly 0c36 Points", t)
					sEq(obj.Fill, "#000000", "poly 0c36 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "poly 0c36 Dash", t)
					sEq(obj.Line, "#000000", "poly 0c36 Line", t)
					iEq(obj.Width, 5, "poly 0c36 Width", t)
					sEq(obj.Layer, "walls", "poly 0c36 Layer", t)
					bEq(obj.Hidden, false, "poly 0c36 Hidden", t)
					iEq(obj.Level, 0, "poly 0c36 Level", t)
					sEq(obj.Group, "", "poly 0c36 Group", t)
					// PolygonElement
					fEq(obj.Spline, 0, "poly 0c36 Spline", t)
					BEq(byte(obj.Join), byte(JoinBevel), "poly 0c36 Join", t)
				default:
					t.Errorf("Found unexpected polygon %s", obj.ObjID())
				}
			case ArcElement:
				switch obj.ObjID() {
				case "13a2dd4a64a94e178509744e1a0a4481":
					// BaseMapObject
					sEq(obj.ID, "13a2dd4a64a94e178509744e1a0a4481", "ID", t)
					// Coordinates
					fEq(obj.X, 473, "arc 13a2 X", t)
					fEq(obj.Y, 523, "arc 13a2 Y", t)
					// MapElement
					iEq(obj.Z, 8, "arc 13a2 Z", t)
					bEq(obj.Locked, false, "arc 13a2 Locked", t)
					CEq(obj.Points, []Coordinates{
						{321, 669},
					}, "arc 13a2 Points", t)
					sEq(obj.Fill, "#ff2600", "arc 13a2 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "arc 13a2 Dash", t)
					sEq(obj.Line, "black", "arc 13a2 Line", t)
					iEq(obj.Width, 5, "arc 13a2 Width", t)
					sEq(obj.Layer, "walls", "arc 13a2 Layer", t)
					bEq(obj.Hidden, false, "arc 13a2 Hidden", t)
					iEq(obj.Level, 0, "arc 13a2 Level", t)
					sEq(obj.Group, "", "arc 13a2 Group", t)
					// ArcElement
					BEq(byte(obj.ArcMode), byte(ArcModeArc), "arc 13a2 Arcmode", t)
					fEq(obj.Start, 20, "arc 13a2 Start", t)
					fEq(obj.Extent, 225, "arc 13a2 Extent", t)
				case "5b1a4fa32af54eb2b35ebec0a4c88089":
					// BaseMapObject
					sEq(obj.ID, "5b1a4fa32af54eb2b35ebec0a4c88089", "ID", t)
					// Coordinates
					fEq(obj.X, 59, "arc 5b1a X", t)
					fEq(obj.Y, 309, "arc 5b1a Y", t)
					// MapElement
					iEq(obj.Z, 6, "arc 5b1a Z", t)
					bEq(obj.Locked, false, "arc 5b1a Locked", t)
					CEq(obj.Points, []Coordinates{
						{161, 485},
					}, "arc 5b1a Points", t)
					sEq(obj.Fill, "#ff2600", "arc 5b1a Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "arc 5b1a Dash", t)
					sEq(obj.Line, "black", "arc 5b1a Line", t)
					iEq(obj.Width, 5, "arc 5b1a Width", t)
					sEq(obj.Layer, "walls", "arc 5b1a Layer", t)
					bEq(obj.Hidden, false, "arc 5b1a Hidden", t)
					iEq(obj.Level, 0, "arc 5b1a Level", t)
					sEq(obj.Group, "", "arc 5b1a Group", t)
					// ArcElement
					BEq(byte(obj.ArcMode), byte(ArcModePieSlice), "arc 5b1a Arcmode", t)
					fEq(obj.Start, 140, "arc 5b1a Start", t)
					fEq(obj.Extent, 61, "arc 5b1a Extent", t)
				case "7b39f6dbeea44b8baa20032c443a0654":
					// BaseMapObject
					sEq(obj.ID, "7b39f6dbeea44b8baa20032c443a0654", "ID", t)
					// Coordinates
					fEq(obj.X, 126, "arc 7b39 X", t)
					fEq(obj.Y, 274, "arc 7b39 Y", t)
					// MapElement
					iEq(obj.Z, 7, "arc 7b39 Z", t)
					bEq(obj.Locked, false, "arc 7b39 Locked", t)
					CEq(obj.Points, []Coordinates{
						{166, 375},
					}, "arc 7b39 Points", t)
					sEq(obj.Fill, "#ff2600", "arc 7b39 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "arc 7b39 Dash", t)
					sEq(obj.Line, "black", "arc 7b39 Line", t)
					iEq(obj.Width, 5, "arc 7b39 Width", t)
					sEq(obj.Layer, "walls", "arc 7b39 Layer", t)
					bEq(obj.Hidden, false, "arc 7b39 Hidden", t)
					iEq(obj.Level, 0, "arc 57b39 Level", t)
					sEq(obj.Group, "", "arc 7b39 Group", t)
					// ArcElement
					BEq(byte(obj.ArcMode), byte(ArcModeChord), "arc 7b39 Arcmode", t)
					fEq(obj.Start, 151, "arc 7b39 Start", t)
					fEq(obj.Extent, 235, "arc 7b39 Extent", t)
				default:
					t.Errorf("Found unexpected arc %s", obj.ObjID())
				}
			case LineElement:
				switch obj.ObjID() {
				case "2a1751827a954d8fad688da8e431502a":
					// BaseMapObject
					sEq(obj.ID, "2a1751827a954d8fad688da8e431502a", "ID", t)
					// Coordinates
					fEq(obj.X, 675, "line 2a17 X", t)
					fEq(obj.Y, 584, "line 2a17 Y", t)
					// MapElement
					iEq(obj.Z, 1, "line 2a17 Z", t)
					bEq(obj.Locked, false, "line 2a17 Locked", t)
					CEq(obj.Points, []Coordinates{
						{779, 291},
					}, "line 2a17 Points", t)
					sEq(obj.Fill, "#ff2600", "line 2a17 Fill", t)
					BEq(byte(obj.Dash), byte(DashLong), "line 2a17 Dash", t)
					sEq(obj.Line, "black", "line 2a17 Line", t)
					iEq(obj.Width, 5, "line 2a17 Width", t)
					sEq(obj.Layer, "walls", "line 2a17 Layer", t)
					bEq(obj.Hidden, false, "line 2a17 Hidden", t)
					iEq(obj.Level, 0, "line 2a17 Level", t)
					sEq(obj.Group, "", "line 2a17 Group", t)
					// LineElement
					BEq(byte(obj.Arrow), byte(ArrowFirst), "line 2a17 Arrow", t)
				case "38f633da2d6749467f5406f187b8cc3f":
					// BaseMapObject
					sEq(obj.ID, "38f633da2d6749467f5406f187b8cc3f", "ID", t)
					// Coordinates
					fEq(obj.X, 10810, "line 38f6 X", t)
					fEq(obj.Y, 14350, "line 38f6 Y", t)
					// MapElement
					iEq(obj.Z, 12, "line 38f6 Z", t)
					bEq(obj.Locked, false, "line 38f6 Locked", t)
					CEq(obj.Points, []Coordinates{
						{10908, 14396},
					}, "line 38f6 Points", t)
					sEq(obj.Fill, "#000000", "line 38f6 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "line 38f6 Dash", t)
					sEq(obj.Line, "black", "line 38f6 Line", t)
					iEq(obj.Width, 5, "line 38f6 Width", t)
					sEq(obj.Layer, "walls", "line 38f6 Layer", t)
					bEq(obj.Hidden, false, "line 38f6 Hidden", t)
					iEq(obj.Level, 0, "line 38f6 Level", t)
					sEq(obj.Group, "", "line 38f6 Group", t)
					// LineElement
					BEq(byte(obj.Arrow), byte(ArrowNone), "line 38f6 Arrow", t)
				case "61dc2ff4efe54be7a18791b338c29c5c":
					// BaseMapObject
					sEq(obj.ID, "61dc2ff4efe54be7a18791b338c29c5c", "ID", t)
					// Coordinates
					fEq(obj.X, 604, "line 61dc X", t)
					fEq(obj.Y, 229, "line 61dc Y", t)
					// MapElement
					iEq(obj.Z, 2, "line 61dc Z", t)
					bEq(obj.Locked, false, "line 61dc Locked", t)
					CEq(obj.Points, []Coordinates{
						{509, 378}, {650, 360},
					}, "line 61dc Points", t)
					sEq(obj.Fill, "#ff2600", "line 61dc Fill", t)
					BEq(byte(obj.Dash), byte(DashLong), "line 61dc Dash", t)
					sEq(obj.Line, "black", "line 38f6 Line", t)
					iEq(obj.Width, 5, "line 61dc Width", t)
					sEq(obj.Layer, "walls", "line 61dc Layer", t)
					bEq(obj.Hidden, false, "line 61dc Hidden", t)
					iEq(obj.Level, 0, "line 61dc Level", t)
					sEq(obj.Group, "", "line 61dc Group", t)
					// LineElement
					BEq(byte(obj.Arrow), byte(ArrowFirst), "line 61dc Arrow", t)
				default:
					t.Errorf("Found unexpected line %s", obj.ObjID())
				}
			case TextElement:
				switch obj.ObjID() {
				case "2c4a8ae53c5c4cbdb902d581402230e7":
					// BaseMapObject
					sEq(obj.ID, "2c4a8ae53c5c4cbdb902d581402230e7", "ID", t)
					// Coordinates
					fEq(obj.X, 565, "text 2c4a X", t)
					fEq(obj.Y, 707, "text 2c4a Y", t)
					// MapElement
					iEq(obj.Z, 9, "text 2c4a Z", t)
					bEq(obj.Locked, false, "text 2c4a Locked", t)
					CEq(obj.Points, []Coordinates{}, "text 2c4a Points", t)
					sEq(obj.Fill, "#ff2600", "text 2c4a Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "text 2c4a Dash", t)
					sEq(obj.Line, "", "text 2c4a Line", t)
					iEq(obj.Width, 0, "text 2c4a Width", t)
					sEq(obj.Layer, "walls", "text 2c4a Layer", t)
					bEq(obj.Hidden, false, "text 2c4a Hidden", t)
					iEq(obj.Level, 0, "text 2c4a Level", t)
					sEq(obj.Group, "", "text 2c4a Group", t)
					// TextElement
					sEq(obj.Text, "hello world", "text 2c4a Text", t)
					sEq(obj.Font.Family, "Arial", "text 2c4a Font Family", t)
					fEq(obj.Font.Size, 24, "text 2c4a Font Size", t)
					BEq(byte(obj.Font.Weight), byte(FontWeightNormal), "text 2c4a Font Weight", t)
					BEq(byte(obj.Font.Slant), byte(FontSlantRoman), "text 2c4a Font Slant", t)
					BEq(byte(obj.Anchor), byte(AnchorCenter), "text 2c4a Anchor", t)
				default:
					t.Errorf("Found unexpected line %s", obj.ObjID())
				}
			case RectangleElement:
				switch obj.ObjID() {
				case "39880f0c6e904bf9a866d9af8783fd70":
					// BaseMapObject
					sEq(obj.ID, "39880f0c6e904bf9a866d9af8783fd70", "ID", t)
					// Coordinates
					fEq(obj.X, 445.5, "rect 3988 X", t)
					fEq(obj.Y, 33, "rect 3988 Y", t)
					// MapElement
					iEq(obj.Z, 3, "rect 3988 Z", t)
					bEq(obj.Locked, true, "rect 3988 Locked", t)
					CEq(obj.Points, []Coordinates{{625, 160}}, "rect 3988 Points", t)
					sEq(obj.Fill, "#ff2600", "rect 3988 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "rect 3988 Dash", t)
					sEq(obj.Line, "black", "rect 3988 Line", t)
					iEq(obj.Width, 5, "rect 3988 Width", t)
					sEq(obj.Layer, "walls", "rect 3988 Layer", t)
					bEq(obj.Hidden, false, "rect 3988 Hidden", t)
					iEq(obj.Level, 0, "rect 3988 Level", t)
					sEq(obj.Group, "", "rect 3988 Group", t)
					// RectangleElement
				case "bf29dfa85cc54498bb33a2d7523d9edc":
					// BaseMapObject
					sEq(obj.ID, "bf29dfa85cc54498bb33a2d7523d9edc", "ID", t)
					// Coordinates
					fEq(obj.X, 289, "rect bf29 X", t)
					fEq(obj.Y, 36, "rect bf29 Y", t)
					// MapElement
					iEq(obj.Z, 4, "rect bf29 Z", t)
					bEq(obj.Locked, false, "rect bf29 Locked", t)
					CEq(obj.Points, []Coordinates{{355, 97}}, "rect bf29 Points", t)
					sEq(obj.Fill, "", "rect bf29 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "rect bf29 Dash", t)
					sEq(obj.Line, "black", "rect bf29 Line", t)
					iEq(obj.Width, 5, "rect bf29 Width", t)
					sEq(obj.Layer, "walls", "rect bf29 Layer", t)
					bEq(obj.Hidden, false, "rect bf29 Hidden", t)
					iEq(obj.Level, 0, "rect bf29 Level", t)
					sEq(obj.Group, "", "rect bf29 Group", t)
					// RectangleElement
				default:
					t.Errorf("Found unexpected line %s", obj.ObjID())
				}
			case CircleElement:
				switch obj.ObjID() {
				case "3f5b6a2655214928b868daad9a97db4d":
					// BaseMapObject
					sEq(obj.ID, "3f5b6a2655214928b868daad9a97db4d", "ID", t)
					// Coordinates
					fEq(obj.X, 110, "circ 2f5b X", t)
					fEq(obj.Y, 18, "circ 2f5b Y", t)
					// MapElement
					iEq(obj.Z, 5, "circ 2f5b Z", t)
					bEq(obj.Locked, false, "circ 2f5b Locked", t)
					CEq(obj.Points, []Coordinates{{237, 150}}, "circ 2f5b Points", t)
					sEq(obj.Fill, "", "circ 2f5b Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "circ 2f5b Dash", t)
					sEq(obj.Line, "black", "circ 2f5b Line", t)
					iEq(obj.Width, 5, "circ 2f5b Width", t)
					sEq(obj.Layer, "walls", "circ 2f5b Layer", t)
					bEq(obj.Hidden, false, "circ 2f5b Hidden", t)
					iEq(obj.Level, 0, "circ 2f5b Level", t)
					sEq(obj.Group, "", "circ 2f5b Group", t)
					// CircleElement
				default:
					t.Errorf("Found unexpected line %s", obj.ObjID())
				}
			case SpellAreaOfEffectElement:
				switch obj.ObjID() {
				case "4b82e91b987d412b9c1a2b5110319072":
					// BaseMapObject
					sEq(obj.ID, "4b82e91b987d412b9c1a2b5110319072", "ID", t)
					// Coordinates
					fEq(obj.X, 500, "aoe 4b82 X", t)
					fEq(obj.Y, 400, "aoe 4b82 Y", t)
					// MapElement
					iEq(obj.Z, 99999999, "aoe 4b82 Z", t)
					bEq(obj.Locked, false, "aoe 4b82 Locked", t)
					CEq(obj.Points, []Coordinates{{500, 150}}, "aoe 4b82 Points", t)
					sEq(obj.Fill, "black", "aoe 4b82 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "aoe 4b82 Dash", t)
					sEq(obj.Line, "black", "aoe 4b82 Line", t)
					iEq(obj.Width, 5, "aoe 4b82 Width", t)
					sEq(obj.Layer, "walls", "aoe 4b82 Layer", t)
					bEq(obj.Hidden, false, "aoe 4b82 Hidden", t)
					iEq(obj.Level, 0, "aoe 4b82 Level", t)
					sEq(obj.Group, "", "aoe 4b82 Group", t)
					// SpellAreaOfEffectElement
					BEq(byte(obj.AoEShape), byte(AoEShapeRadius), "aoe 4b82 AoE Shape", t)
				case "afd136735d7e400082f331485e73f7a1":
					// BaseMapObject
					sEq(obj.ID, "afd136735d7e400082f331485e73f7a1", "ID", t)
					// Coordinates
					fEq(obj.X, 150, "aoe afd1 X", t)
					fEq(obj.Y, 600, "aoe afd1 Y", t)
					// MapElement
					iEq(obj.Z, 99999999, "aoe afd1 Z", t)
					bEq(obj.Locked, false, "aoe afd1 Locked", t)
					CEq(obj.Points, []Coordinates{{200, 200}}, "aoe afd1 Points", t)
					sEq(obj.Fill, "#00f900", "aoe afd1 Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "aoe afd1 Dash", t)
					sEq(obj.Line, "black", "aoe afd1 Line", t)
					iEq(obj.Width, 5, "aoe afd1 Width", t)
					sEq(obj.Layer, "walls", "aoe afd1 Layer", t)
					bEq(obj.Hidden, false, "aoe afd1 Hidden", t)
					iEq(obj.Level, 0, "aoe afd1 Level", t)
					sEq(obj.Group, "", "aoe afd1 Group", t)
					// SpellAreaOfEffectElement
					BEq(byte(obj.AoEShape), byte(AoEShapeRay), "aoe afd1 AoE Shape", t)
				case "e68d5354f175401582866a75d806d8d7":
					// BaseMapObject
					sEq(obj.ID, "e68d5354f175401582866a75d806d8d7", "ID", t)
					// Coordinates
					fEq(obj.X, 850, "aoe e68d X", t)
					fEq(obj.Y, 800, "aoe e68d Y", t)
					// MapElement
					iEq(obj.Z, 99999999, "aoe e68d Z", t)
					bEq(obj.Locked, false, "aoe e68d Locked", t)
					CEq(obj.Points, []Coordinates{{450, 800}}, "aoe e68d Points", t)
					sEq(obj.Fill, "#00f900", "aoe e68d Fill", t)
					BEq(byte(obj.Dash), byte(DashSolid), "aoe e68d Dash", t)
					sEq(obj.Line, "black", "aoe e68d Line", t)
					iEq(obj.Width, 5, "aoe e68d Width", t)
					sEq(obj.Layer, "walls", "aoe e68d Layer", t)
					bEq(obj.Hidden, false, "aoe e68d Hidden", t)
					iEq(obj.Level, 0, "aoe e68d Level", t)
					sEq(obj.Group, "", "aoe e68d Group", t)
					// SpellAreaOfEffectElement
					BEq(byte(obj.AoEShape), byte(AoEShapeCone), "aoe e68d AoE Shape", t)
				default:
					t.Errorf("Found unexpected aoe %s", obj.ObjID())
				}
			default:
			}
		}

		sort.Strings(IDsFound)
		expect := []string{
			"0006bc4a7063427b8fb1f8990a24b980",
			"0268e7eeb78e41ff82fddc4f5f0e2c1d",
			"09426d492f784ad25684536c35e0d8d5",
			"0c36b6a97a074bd174cda800f07206f4",
			"13a2dd4a64a94e178509744e1a0a4481",
			"2a1751827a954d8fad688da8e431502a",
			"2c4a8ae53c5c4cbdb902d581402230e7",
			"2df3e0a104614c5cb76f31836bc3f84d",
			"38f633da2d6749467f5406f187b8cc3f",
			"39880f0c6e904bf9a866d9af8783fd70",
			"39a1afc1b1aa4cac87eee32be93ebe9a",
			"3f5b6a2655214928b868daad9a97db4d",
			"4b82e91b987d412b9c1a2b5110319072",
			"5b1a4fa32af54eb2b35ebec0a4c88089",
			"61dc2ff4efe54be7a18791b338c29c5c",
			"7b39f6dbeea44b8baa20032c443a0654",
			"976e7148ae86409f99fdebf83f3f0904",
			"PC73",
			"afd136735d7e400082f331485e73f7a1",
			"bf29dfa85cc54498bb33a2d7523d9edc",
			"e68d5354f175401582866a75d806d8d7",
		}
		if !reflect.DeepEqual(IDsFound, expect) {
			t.Errorf("Expected to find %q, but found %q", expect, IDsFound)
		}

		var f strings.Builder
		err := SaveMapFile(&f, objs, MapMetaData{
			Timestamp:   1,
			DateTime:    "n/a",
			Comment:     "no",
			Location:    "here",
			FileVersion: 999,
		})

		if err != nil {
			t.Errorf("SaveObjects returned error %v", err)
		}
		sEq(f.String(), expected_save, "save data", t)
	}

	check()

	objs, _, err = LoadMapFile(strings.NewReader(expected_save))
	check()
}

func fEq(a, b float64, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%v\" expected=\"%v\"", msg, a, b)
	}
}

func BEq(a, b byte, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%v\" expected=\"%v\"", msg, a, b)
	}
}

func bEq(a, b bool, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%v\" expected=\"%v\"", msg, a, b)
	}
}

func iEq(a, b int, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%v\" expected=\"%v\"", msg, a, b)
	}
}

func uEq(a, b uint, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%v\" expected=\"%v\"", msg, a, b)
	}
}

func sEq(a, b, msg string, t *testing.T) {
	if a != b {
		if len(a) != len(b) {
			t.Errorf("%s: lengths differ (%d vs %d) actual=\"%s\" expected=\"%s\"", msg, len(a), len(b), a, b)
		} else {
			for i := 0; i < len(a); i++ {
				if a[i] != b[i] {
					t.Errorf("%s: strings differ at position %d (%v vs %v)", msg, i, a[i], b[i])
				}
			}
			t.Errorf("%s: actual=\"%s\" expected=\"%s\"", msg, a, b)
		}
	}
}

func SEq(a, b []string, msg string, t *testing.T) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%s: actual=\"%q\" expected=\"%q\"", msg, a, b)
	}
}

func CEq(a, b []Coordinates, msg string, t *testing.T) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%s: actual coordinate set:", msg)
		for _, coords := range a {
			t.Errorf("   (%g, %g)", coords.X, coords.Y)
		}
		t.Errorf("%s: expected coordinate set:", msg)
		for _, coords := range b {
			t.Errorf("   (%g, %g)", coords.X, coords.Y)
		}
	}
}

// @[00]@| GMA 4.3.10
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
