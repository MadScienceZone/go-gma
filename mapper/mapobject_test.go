/*
########################################################################################
#  _______  _______  _______                ___       ______      ______               #
# (  ____ \(       )(  ___  )              /   )     / ___  \    / ___  \              #
# | (    \/| () () || (   ) |             / /) |     \/   \  \   \/   \  \             #
# | |      | || || || (___) |            / (_) (_       ___) /      ___) /             #
# | | ____ | |(_)| ||  ___  |           (____   _)     (___ (      (___ (              #
# | | \_  )| |   | || (   ) | Game           ) (           ) \         ) \             #
# | (___) || )   ( || )   ( | Master's       | |   _ /\___/  / _ /\___/  /             #
# (_______)|/     \||/     \| Assistant      (_)  (_)\______/ (_)\______/              #
#                                                                                      #
########################################################################################
*/

//
// Unit tests for the mapper authentication code
//

package mapper

import (
	"reflect"
	"sort"
	"testing"
)

func TestObjLoadNil(t *testing.T) {
	objs, imgs, files, err := ParseObjects(nil)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(objs) > 0 {
		t.Errorf("objs not nil: %d %v", len(objs), objs)
	}
	if len(imgs) > 0 {
		t.Errorf("imgs not nil: %d %v", len(imgs), imgs)
	}
	if len(files) > 0 {
		t.Errorf("files not nil: %d %v", len(files), files)
	}
}

func TestObjLoadImages(t *testing.T) {
	objs, imgs, files, err := ParseObjects([]string{
		"__MAPPER__:16 {test {0 nil}}",
		"I #SimonKazar 1.0 #SimonKazar@50.gif",
		"I #Firefly 2.0 @OoSmGY0XERJRrA8ZiK_igg_Firefly@100",
		"I #SimonKazar 0.25 @OoSmGY0XERJRrA8ZiK_igg_Firefly@12",
	})
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(objs) > 0 {
		t.Errorf("objs not nil: %d %v", len(objs), objs)
	}
	if !reflect.DeepEqual(imgs, map[string]ImageDefinition{
		"#SimonKazar:1": ImageDefinition{
			Zoom:        1.0,
			Name:        "#SimonKazar",
			File:        "#SimonKazar@50.gif",
			IsLocalFile: true,
		},
		"#SimonKazar:0.25": ImageDefinition{
			Zoom:        0.25,
			Name:        "#SimonKazar",
			File:        "@OoSmGY0XERJRrA8ZiK_igg_Firefly@12",
			IsLocalFile: false,
		},
		"#Firefly:2": ImageDefinition{
			Zoom:        2.0,
			Name:        "#Firefly",
			File:        "@OoSmGY0XERJRrA8ZiK_igg_Firefly@100",
			IsLocalFile: false,
		}}) {
		t.Errorf("imgs %d %v", len(imgs), imgs)
	}
	if len(files) > 0 {
		t.Errorf("files not nil: %d %v", len(files), files)
	}
}

func TestObjLoadOnePlayer(t *testing.T) {
	objs, imgs, files, err := ParseObjects([]string{
		"__MAPPER__:16 {test {0 nil}}",
		"P HEALTH:PC73 {28 6 0 16 0 0 {} 10}",
		"P NAME:PC73 Jigu",
		"P COLOR:PC73 blue",
		"P GY:PC73 14",
		"P DIM:PC73 1",
		"P SKIN:PC73 0",
		"P NOTE:PC73 {}",
		"P SIZE:PC73 M",
		"P GX:PC73 31",
		"P AREA:PC73 M",
		"P ELEV:PC73 0",
		"P REACH:PC73 0",
		"P MOVEMODE:PC73 {}",
		"P TYPE:PC73 player",
		"P KILLED:PC73 0",
		"P SPAM:PC73 eggs",
	})
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(imgs) > 0 {
		t.Errorf("imgs not nil: %d %v", len(imgs), imgs)
	}
	if len(objs) != 1 {
		t.Fatalf("objs count: %d %v", len(objs), objs)
	}
	if reflect.TypeOf(objs[0]).Name() != "PlayerToken" {
		t.Errorf("obj type %v", reflect.TypeOf(objs[0]))
	}
	p := objs[0].(PlayerToken)
	sEq(p.ID, "PC73", "ID", t)
	sEq(p.ObjID(), "PC73", "ObjID", t)
	iEq(p.Health.MaxHP, 28, "MaxHP", t)
	if !reflect.DeepEqual(objs[0].(PlayerToken), PlayerToken{
		CreatureToken: CreatureToken{
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
				HpBlur:          10,
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
		}}) {
		t.Errorf("jigu %q", objs)
	}
	if len(files) > 0 {
		t.Errorf("files not nil: %d %v", len(files), files)
	}

	saveData, err := SaveObjects(objs, imgs, files)
	sort.Strings(saveData)
	if !reflect.DeepEqual(saveData, []string{
		"P AREA:PC73 M",
		"P COLOR:PC73 blue",
		"P DIM:PC73 1",
		"P ELEV:PC73 0",
		"P GX:PC73 31",
		"P GY:PC73 14",
		"P HEALTH:PC73 {28 6 0 16 0 0 {} 10}",
		"P KILLED:PC73 0",
		"P MOVEMODE:PC73 land",
		"P NAME:PC73 Jigu",
		"P REACH:PC73 0",
		"P SIZE:PC73 M",
		"P SKIN:PC73 0",
		"P TYPE:PC73 player",
	}) {
		t.Errorf("save data %q", saveData)
	}
}
func TestObjLoadSmallMap(t *testing.T) {
	objs, imgs, files, err := ParseObjects([]string{
		"__MAPPER__:14 {{testing here} {1589430476 {Wed May 13 21:27:56 PDT 2020}}}",
		"Z:0006bc4a7063427b8fb1f8990a24b980 1965",
		"X:0006bc4a7063427b8fb1f8990a24b980 2100.0",
		"Y:0006bc4a7063427b8fb1f8990a24b980 7350.0",
		"IMAGE:0006bc4a7063427b8fb1f8990a24b980 parquet22",
		"TYPE:0006bc4a7063427b8fb1f8990a24b980 tile",
		"LAYER:0006bc4a7063427b8fb1f8990a24b980 walls",
		"POINTS:0006bc4a7063427b8fb1f8990a24b980 {}",
		"FILL:0268e7eeb78e41ff82fddc4f5f0e2c1d #7e7f12",
		"LINE:0268e7eeb78e41ff82fddc4f5f0e2c1d #7e7f12",
		"SPLINE:0268e7eeb78e41ff82fddc4f5f0e2c1d 0",
		"X:0268e7eeb78e41ff82fddc4f5f0e2c1d 10888.0",
		"Y:0268e7eeb78e41ff82fddc4f5f0e2c1d 12642.0",
		"Z:0268e7eeb78e41ff82fddc4f5f0e2c1d 271",
		"WIDTH:0268e7eeb78e41ff82fddc4f5f0e2c1d 5",
		"JOIN:0268e7eeb78e41ff82fddc4f5f0e2c1d bevel",
		"TYPE:0268e7eeb78e41ff82fddc4f5f0e2c1d poly",
		"LAYER:0268e7eeb78e41ff82fddc4f5f0e2c1d walls",
		"POINTS:0268e7eeb78e41ff82fddc4f5f0e2c1d {10882.0 12698.0 10866.0 12738.0 10832.0 12776.0 10806.0 12816.0 10762.0 12866.0 10682.0 12928.0 10582.0 12970.0 10500.0 13024.0 10452.0 13074.0 10358.0 13100.0 10262.0 13098.0 10144.0 13060.0 10052.0 13044.0 9964.0 13044.0 9902.0 12992.0 9828.0 13010.0 9778.0 13048.0 9848.0 12282.0}",
		"FILL:09426d492f784ad25684536c35e0d8d5 #000000",
		"LINE:09426d492f784ad25684536c35e0d8d5 #000000",
		"SPLINE:09426d492f784ad25684536c35e0d8d5 0",
		"Y:09426d492f784ad25684536c35e0d8d5 13222.0",
		"WIDTH:09426d492f784ad25684536c35e0d8d5 2",
		"Z:09426d492f784ad25684536c35e0d8d5 25",
		"X:09426d492f784ad25684536c35e0d8d5 9591.0",
		"JOIN:09426d492f784ad25684536c35e0d8d5 bevel",
		"TYPE:09426d492f784ad25684536c35e0d8d5 poly",
		"LAYER:09426d492f784ad25684536c35e0d8d5 walls",
		"POINTS:09426d492f784ad25684536c35e0d8d5 {9600.0 13233.0 9609.0 13249.0 9626.0 13290.0 9633.0 13310.0 9639.0 13327.0 9647.0 13356.0 9651.0 13373.0 9655.0 13403.0 9658.0 13418.0 9663.0 13445.0 9669.0 13477.0 9690.0 13517.0 9691.0 13535.0 9698.0 13553.0 9708.0 13579.0 9715.0 13605.0 9719.0 13618.0 9725.0 13641.0 9730.0 13664.0 9751.0 13690.0 9761.0 13707.0 9767.0 13721.0 9774.0 13733.0 9795.0 13752.0 9767.0 13754.0 9757.0 13743.0 9752.0 13732.0 9745.0 13720.0 9721.0 13684.0 9715.0 13672.0 9711.0 13665.0 9691.0 13637.0 9682.0 13621.0 9664.0 13596.0 9650.0 13581.0 9637.0 13564.0 9627.0 13548.0 9607.0 13524.0 9586.0 13504.0 9572.0 13492.0 9568.0 13489.0}",
		"FILL:0c36b6a97a074bd174cda800f07206f4 #000000",
		"LINE:0c36b6a97a074bd174cda800f07206f4 #000000",
		"SPLINE:0c36b6a97a074bd174cda800f07206f4 0",
		"X:0c36b6a97a074bd174cda800f07206f4 12598.0",
		"WIDTH:0c36b6a97a074bd174cda800f07206f4 5",
		"Y:0c36b6a97a074bd174cda800f07206f4 5697.0",
		"Z:0c36b6a97a074bd174cda800f07206f4 57",
		"JOIN:0c36b6a97a074bd174cda800f07206f4 bevel",
		"TYPE:0c36b6a97a074bd174cda800f07206f4 poly",
		"LAYER:0c36b6a97a074bd174cda800f07206f4 walls",
		"POINTS:0c36b6a97a074bd174cda800f07206f4 {12630.0 5689.0 12646.0 5684.0 12668.0 5677.0 12694.0 5670.0 12738.0 5665.0 12793.0 5659.0 12845.0 5657.0 12900.0 5652.0 12945.0 5763.0 12901.0 5848.0 12865.0 5848.0 12814.0 5849.0 12752.0 5842.0 12716.0 5845.0 12675.0 5850.0 12647.0 5848.0 12612.0 5845.0 12581.0 5851.0}",
		"FILL:13a2dd4a64a94e178509744e1a0a4481 #ff2600",
		"START:13a2dd4a64a94e178509744e1a0a4481 20",
		"EXTENT:13a2dd4a64a94e178509744e1a0a4481 225",
		"X:13a2dd4a64a94e178509744e1a0a4481 473.0",
		"WIDTH:13a2dd4a64a94e178509744e1a0a4481 5",
		"Y:13a2dd4a64a94e178509744e1a0a4481 523.0",
		"Z:13a2dd4a64a94e178509744e1a0a4481 8",
		"ARCMODE:13a2dd4a64a94e178509744e1a0a4481 arc",
		"TYPE:13a2dd4a64a94e178509744e1a0a4481 arc",
		"LINE:13a2dd4a64a94e178509744e1a0a4481 black",
		"LAYER:13a2dd4a64a94e178509744e1a0a4481 walls",
		"POINTS:13a2dd4a64a94e178509744e1a0a4481 {321.0 669.0}",
		"DASH:13a2dd4a64a94e178509744e1a0a4481 {}",
		"FILL:2a1751827a954d8fad688da8e431502a #ff2600",
		"DASH:2a1751827a954d8fad688da8e431502a -",
		"Z:2a1751827a954d8fad688da8e431502a 1",
		"WIDTH:2a1751827a954d8fad688da8e431502a 5",
		"Y:2a1751827a954d8fad688da8e431502a 584.0",
		"X:2a1751827a954d8fad688da8e431502a 675.0",
		"LINE:2a1751827a954d8fad688da8e431502a black",
		"ARROW:2a1751827a954d8fad688da8e431502a first",
		"TYPE:2a1751827a954d8fad688da8e431502a line",
		"LAYER:2a1751827a954d8fad688da8e431502a walls",
		"POINTS:2a1751827a954d8fad688da8e431502a {779.0 291.0}",
		"FILL:2c4a8ae53c5c4cbdb902d581402230e7 #ff2600",
		"X:2c4a8ae53c5c4cbdb902d581402230e7 565.0",
		"Y:2c4a8ae53c5c4cbdb902d581402230e7 707.0",
		"Z:2c4a8ae53c5c4cbdb902d581402230e7 9",
		"ANCHOR:2c4a8ae53c5c4cbdb902d581402230e7 center",
		"TYPE:2c4a8ae53c5c4cbdb902d581402230e7 text",
		"LAYER:2c4a8ae53c5c4cbdb902d581402230e7 walls",
		"TEXT:2c4a8ae53c5c4cbdb902d581402230e7 {hello world}",
		"FONT:2c4a8ae53c5c4cbdb902d581402230e7 {{Arial 24 normal roman}}",
		"POINTS:2c4a8ae53c5c4cbdb902d581402230e7 {}",
		"M ELEV:2df3e0a104614c5cb76f31836bc3f84d 0",
		"M REACH:2df3e0a104614c5cb76f31836bc3f84d 0",
		"M SKIN:2df3e0a104614c5cb76f31836bc3f84d 0",
		"M DIM:2df3e0a104614c5cb76f31836bc3f84d 1",
		"M KILLED:2df3e0a104614c5cb76f31836bc3f84d 1",
		"M GY:2df3e0a104614c5cb76f31836bc3f84d 22",
		"M GX:2df3e0a104614c5cb76f31836bc3f84d 27",
		"M AREA:2df3e0a104614c5cb76f31836bc3f84d S",
		"M SIZE:2df3e0a104614c5cb76f31836bc3f84d S",
		"M TYPE:2df3e0a104614c5cb76f31836bc3f84d monster",
		"M COLOR:2df3e0a104614c5cb76f31836bc3f84d red",
		"M HEALTH:2df3e0a104614c5cb76f31836bc3f84d {9 29 0 15 0 0 {} {}}",
		"M NAME:2df3e0a104614c5cb76f31836bc3f84d {Fleshdreg #4}",
		"M MOVEMODE:2df3e0a104614c5cb76f31836bc3f84d {}",
		"M NOTE:2df3e0a104614c5cb76f31836bc3f84d {}",
		"FILL:38f633da2d6749467f5406f187b8cc3f #000000",
		"X:38f633da2d6749467f5406f187b8cc3f 10810.0",
		"Z:38f633da2d6749467f5406f187b8cc3f 12",
		"Y:38f633da2d6749467f5406f187b8cc3f 14350.0",
		"WIDTH:38f633da2d6749467f5406f187b8cc3f 5",
		"LINE:38f633da2d6749467f5406f187b8cc3f black",
		"TYPE:38f633da2d6749467f5406f187b8cc3f line",
		"LAYER:38f633da2d6749467f5406f187b8cc3f walls",
		"POINTS:38f633da2d6749467f5406f187b8cc3f {10908.0 14396.0}",
		"FILL:39880f0c6e904bf9a866d9af8783fd70 #ff2600",
		"Z:39880f0c6e904bf9a866d9af8783fd70 3",
		"Y:39880f0c6e904bf9a866d9af8783fd70 33.0",
		"X:39880f0c6e904bf9a866d9af8783fd70 445.5",
		"WIDTH:39880f0c6e904bf9a866d9af8783fd70 5",
		"LOCKED:39880f0c6e904bf9a866d9af8783fd70 1",
		"LINE:39880f0c6e904bf9a866d9af8783fd70 black",
		"TYPE:39880f0c6e904bf9a866d9af8783fd70 rect",
		"LAYER:39880f0c6e904bf9a866d9af8783fd70 walls",
		"POINTS:39880f0c6e904bf9a866d9af8783fd70 {625.0 160.0}",
		"DASH:39880f0c6e904bf9a866d9af8783fd70 {}",
		"M ELEV:39a1afc1b1aa4cac87eee32be93ebe9a 0",
		"M KILLED:39a1afc1b1aa4cac87eee32be93ebe9a 0",
		"M SKIN:39a1afc1b1aa4cac87eee32be93ebe9a 0",
		"M DIM:39a1afc1b1aa4cac87eee32be93ebe9a 1",
		"M REACH:39a1afc1b1aa4cac87eee32be93ebe9a 1",
		"M GX:39a1afc1b1aa4cac87eee32be93ebe9a 19",
		"M GY:39a1afc1b1aa4cac87eee32be93ebe9a 19",
		"M AREA:39a1afc1b1aa4cac87eee32be93ebe9a M",
		"M SIZE:39a1afc1b1aa4cac87eee32be93ebe9a M",
		"M NAME:39a1afc1b1aa4cac87eee32be93ebe9a barbarian2=Caroll",
		"M TYPE:39a1afc1b1aa4cac87eee32be93ebe9a monster",
		"M COLOR:39a1afc1b1aa4cac87eee32be93ebe9a red",
		"M STATUSLIST:39a1afc1b1aa4cac87eee32be93ebe9a stable",
		"M HEALTH:39a1afc1b1aa4cac87eee32be93ebe9a {45 56 0 14 0 0 {} 0}",
		"M MOVEMODE:39a1afc1b1aa4cac87eee32be93ebe9a {}",
		"M NOTE:39a1afc1b1aa4cac87eee32be93ebe9a {}",
		"X:3f5b6a2655214928b868daad9a97db4d 110.0",
		"Y:3f5b6a2655214928b868daad9a97db4d 18.0",
		"WIDTH:3f5b6a2655214928b868daad9a97db4d 5",
		"Z:3f5b6a2655214928b868daad9a97db4d 5",
		"LINE:3f5b6a2655214928b868daad9a97db4d black",
		"TYPE:3f5b6a2655214928b868daad9a97db4d circ",
		"LAYER:3f5b6a2655214928b868daad9a97db4d walls",
		"POINTS:3f5b6a2655214928b868daad9a97db4d {237.0 150.0}",
		"DASH:3f5b6a2655214928b868daad9a97db4d {}",
		"FILL:3f5b6a2655214928b868daad9a97db4d {}",
		"Y:4b82e91b987d412b9c1a2b5110319072 400.0",
		"WIDTH:4b82e91b987d412b9c1a2b5110319072 5",
		"X:4b82e91b987d412b9c1a2b5110319072 500.0",
		"Z:4b82e91b987d412b9c1a2b5110319072 99999999",
		"TYPE:4b82e91b987d412b9c1a2b5110319072 aoe",
		"FILL:4b82e91b987d412b9c1a2b5110319072 black",
		"LINE:4b82e91b987d412b9c1a2b5110319072 black",
		"AOESHAPE:4b82e91b987d412b9c1a2b5110319072 radius",
		"LAYER:4b82e91b987d412b9c1a2b5110319072 walls",
		"POINTS:4b82e91b987d412b9c1a2b5110319072 {500.0 150.0}",
		"DASH:4b82e91b987d412b9c1a2b5110319072 {}",
		"FILL:5b1a4fa32af54eb2b35ebec0a4c88089 #ff2600",
		"START:5b1a4fa32af54eb2b35ebec0a4c88089 140",
		"Y:5b1a4fa32af54eb2b35ebec0a4c88089 309.0",
		"WIDTH:5b1a4fa32af54eb2b35ebec0a4c88089 5",
		"X:5b1a4fa32af54eb2b35ebec0a4c88089 59.0",
		"Z:5b1a4fa32af54eb2b35ebec0a4c88089 6",
		"EXTENT:5b1a4fa32af54eb2b35ebec0a4c88089 61",
		"TYPE:5b1a4fa32af54eb2b35ebec0a4c88089 arc",
		"LINE:5b1a4fa32af54eb2b35ebec0a4c88089 black",
		"ARCMODE:5b1a4fa32af54eb2b35ebec0a4c88089 pieslice",
		"LAYER:5b1a4fa32af54eb2b35ebec0a4c88089 walls",
		"POINTS:5b1a4fa32af54eb2b35ebec0a4c88089 {161.0 485.0}",
		"DASH:5b1a4fa32af54eb2b35ebec0a4c88089 {}",
		"FILL:61dc2ff4efe54be7a18791b338c29c5c #ff2600",
		"DASH:61dc2ff4efe54be7a18791b338c29c5c -",
		"Z:61dc2ff4efe54be7a18791b338c29c5c 2",
		"Y:61dc2ff4efe54be7a18791b338c29c5c 229.0",
		"WIDTH:61dc2ff4efe54be7a18791b338c29c5c 5",
		"X:61dc2ff4efe54be7a18791b338c29c5c 604.0",
		"LINE:61dc2ff4efe54be7a18791b338c29c5c black",
		"ARROW:61dc2ff4efe54be7a18791b338c29c5c first",
		"TYPE:61dc2ff4efe54be7a18791b338c29c5c line",
		"LAYER:61dc2ff4efe54be7a18791b338c29c5c walls",
		"POINTS:61dc2ff4efe54be7a18791b338c29c5c {509.0 378.0 650.0 360.0}",
		"FILL:7b39f6dbeea44b8baa20032c443a0654 #ff2600",
		"X:7b39f6dbeea44b8baa20032c443a0654 126.0",
		"START:7b39f6dbeea44b8baa20032c443a0654 151",
		"EXTENT:7b39f6dbeea44b8baa20032c443a0654 235",
		"Y:7b39f6dbeea44b8baa20032c443a0654 274.0",
		"WIDTH:7b39f6dbeea44b8baa20032c443a0654 5",
		"Z:7b39f6dbeea44b8baa20032c443a0654 7",
		"TYPE:7b39f6dbeea44b8baa20032c443a0654 arc",
		"LINE:7b39f6dbeea44b8baa20032c443a0654 black",
		"ARCMODE:7b39f6dbeea44b8baa20032c443a0654 chord",
		"LAYER:7b39f6dbeea44b8baa20032c443a0654 walls",
		"POINTS:7b39f6dbeea44b8baa20032c443a0654 {166.0 375.0}",
		"DASH:7b39f6dbeea44b8baa20032c443a0654 {}",
		"P DIM:976e7148ae86409f99fdebf83f3f0904 0",
		"P KILLED:976e7148ae86409f99fdebf83f3f0904 0",
		"P REACH:976e7148ae86409f99fdebf83f3f0904 0",
		"P SKIN:976e7148ae86409f99fdebf83f3f0904 1",
		"P ELEV:976e7148ae86409f99fdebf83f3f0904 20",
		"P GX:976e7148ae86409f99fdebf83f3f0904 6",
		"P GY:976e7148ae86409f99fdebf83f3f0904 6",
		"P NAME:976e7148ae86409f99fdebf83f3f0904 Jigu2",
		"P AREA:976e7148ae86409f99fdebf83f3f0904 M",
		"P SIZE:976e7148ae86409f99fdebf83f3f0904 M",
		"P MOVEMODE:976e7148ae86409f99fdebf83f3f0904 fly",
		"P COLOR:976e7148ae86409f99fdebf83f3f0904 green",
		"P TYPE:976e7148ae86409f99fdebf83f3f0904 player",
		"P STATUSLIST:976e7148ae86409f99fdebf83f3f0904 {confused exhausted nauseated}",
		"P AOE:976e7148ae86409f99fdebf83f3f0904 {radius 2 black}",
		"P NOTE:976e7148ae86409f99fdebf83f3f0904 {spam spam}",
		"P HEALTH:976e7148ae86409f99fdebf83f3f0904 {}",
		"P KILLED:PC73 0",
		"P REACH:PC73 0",
		"P DIM:PC73 1",
		"P SKIN:PC73 1",
		"P GY:PC73 14",
		"P ELEV:PC73 30",
		"P GX:PC73 31",
		"P NAME:PC73 Jigu",
		"P AREA:PC73 M",
		"P SIZE:PC73 M",
		"P COLOR:PC73 blue",
		"P MOVEMODE:PC73 fly",
		"P TYPE:PC73 player",
		"P HEALTH:PC73 {28 6 1 16 0 0 surprised}",
		"P SKINSIZE:PC73 {M L}",
		"P NOTE:PC73 {Mirror Image 2}",
		"FILL:afd136735d7e400082f331485e73f7a1 #00f900",
		"X:afd136735d7e400082f331485e73f7a1 150.0",
		"WIDTH:afd136735d7e400082f331485e73f7a1 5",
		"Y:afd136735d7e400082f331485e73f7a1 600.0",
		"Z:afd136735d7e400082f331485e73f7a1 99999999",
		"TYPE:afd136735d7e400082f331485e73f7a1 aoe",
		"LINE:afd136735d7e400082f331485e73f7a1 black",
		"AOESHAPE:afd136735d7e400082f331485e73f7a1 ray",
		"LAYER:afd136735d7e400082f331485e73f7a1 walls",
		"POINTS:afd136735d7e400082f331485e73f7a1 {200.0 200.0}",
		"DASH:afd136735d7e400082f331485e73f7a1 {}",
		"X:bf29dfa85cc54498bb33a2d7523d9edc 289.0",
		"Y:bf29dfa85cc54498bb33a2d7523d9edc 36.0",
		"Z:bf29dfa85cc54498bb33a2d7523d9edc 4",
		"WIDTH:bf29dfa85cc54498bb33a2d7523d9edc 5",
		"LINE:bf29dfa85cc54498bb33a2d7523d9edc black",
		"TYPE:bf29dfa85cc54498bb33a2d7523d9edc rect",
		"LAYER:bf29dfa85cc54498bb33a2d7523d9edc walls",
		"POINTS:bf29dfa85cc54498bb33a2d7523d9edc {355.0 97.0}",
		"DASH:bf29dfa85cc54498bb33a2d7523d9edc {}",
		"FILL:bf29dfa85cc54498bb33a2d7523d9edc {}",
		"FILL:e68d5354f175401582866a75d806d8d7 #00f900",
		"WIDTH:e68d5354f175401582866a75d806d8d7 5",
		"Y:e68d5354f175401582866a75d806d8d7 800.0",
		"X:e68d5354f175401582866a75d806d8d7 850.0",
		"Z:e68d5354f175401582866a75d806d8d7 99999999",
		"TYPE:e68d5354f175401582866a75d806d8d7 aoe",
		"LINE:e68d5354f175401582866a75d806d8d7 black",
		"AOESHAPE:e68d5354f175401582866a75d806d8d7 cone",
		"LAYER:e68d5354f175401582866a75d806d8d7 walls",
		"POINTS:e68d5354f175401582866a75d806d8d7 {450.0 800.0}",
		"DASH:e68d5354f175401582866a75d806d8d7 {}",
	})
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(imgs) > 0 {
		t.Errorf("imgs not nil: %d %v", len(imgs), imgs)
	}
	IDsFound := make([]string, 0, 10)
	for _, o := range objs {
		IDsFound = append(IDsFound, o.ObjID())
		switch obj := o.(type) {
		case MonsterToken:
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
				iEq(obj.Health.HpBlur, 0, "HpBlur", t)
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
				BEq(obj.MoveMode, MoveModeLand, "MoveMode", t)
				bEq(obj.Reach, false, "Reach", t)
				bEq(obj.Killed, true, "Killed", t)
				bEq(obj.Dim, true, "Dim", t)
			case "39a1afc1b1aa4cac87eee32be93ebe9a":
				sEq(obj.ID, "39a1afc1b1aa4cac87eee32be93ebe9a", "ID", t)
				iEq(obj.Health.MaxHP, 45, "MaxHP", t)
				iEq(obj.Health.LethalDamage, 56, "LethalDamage", t)
				iEq(obj.Health.NonLethalDamage, 0, "NonLethalDamage", t)
				iEq(obj.Health.Con, 14, "Con", t)
				bEq(obj.Health.IsFlatFooted, false, "IsFlatFooted", t)
				bEq(obj.Health.IsStable, false, "IsStable", t)
				sEq(obj.Health.Condition, "", "Condition", t)
				iEq(obj.Health.HpBlur, 0, "HpBlur", t)
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
				BEq(obj.MoveMode, MoveModeLand, "MoveMode", t)
				bEq(obj.Reach, true, "Reach", t)
				bEq(obj.Killed, false, "Killed", t)
				bEq(obj.Dim, true, "Dim", t)
			default:
				t.Errorf("Found unexpected monster token %s", obj.ObjID())
			}
		case PlayerToken:
			switch obj.ObjID() {
			case "PC73":
				sEq(obj.ID, "PC73", "ID", t)
				iEq(obj.Health.MaxHP, 28, "MaxHP", t)
				iEq(obj.Health.LethalDamage, 6, "LethalDamage", t)
				iEq(obj.Health.NonLethalDamage, 1, "NonLethalDamage", t)
				iEq(obj.Health.Con, 16, "Con", t)
				bEq(obj.Health.IsFlatFooted, false, "IsFlatFooted", t)
				bEq(obj.Health.IsStable, false, "IsStable", t)
				sEq(obj.Health.Condition, "surprised", "Condition", t)
				iEq(obj.Health.HpBlur, 0, "HpBlur", t)
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
				BEq(obj.MoveMode, MoveModeFly, "MoveMode", t)
				bEq(obj.Reach, false, "Reach", t)
				bEq(obj.Killed, false, "Killed", t)
				bEq(obj.Dim, true, "Dim", t)
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
				BEq(obj.MoveMode, MoveModeFly, "MoveMode", t)
				bEq(obj.Reach, false, "Reach", t)
				bEq(obj.Killed, false, "Killed", t)
				bEq(obj.Dim, false, "Dim", t)
			default:
				t.Errorf("Found unexpected player token %s", obj.ObjID())
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
				BEq(obj.Dash, DashSolid, "tile Dash", t)
				sEq(obj.Line, "", "tile Line", t)
				iEq(obj.Width, 0, "tile Width", t)
				sEq(obj.Layer, "walls", "tile Layer", t)
				bEq(obj.Hidden, false, "tile Hidden", t)
				iEq(obj.Level, 0, "tile Level", t)
				sEq(obj.Group, "", "tile Group", t)
				// TileElement
				sEq(obj.Image, "parquet22", "tile Image", t)
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
				BEq(obj.Dash, DashSolid, "poly 0268 Dash", t)
				sEq(obj.Line, "#7e7f12", "poly 0268 Line", t)
				iEq(obj.Width, 5, "poly 0268 Width", t)
				sEq(obj.Layer, "walls", "poly 0268 Layer", t)
				bEq(obj.Hidden, false, "poly 0268 Hidden", t)
				iEq(obj.Level, 0, "poly 0268 Level", t)
				sEq(obj.Group, "", "poly 0268 Group", t)
				// PolygonElement
				fEq(obj.Spline, 0, "poly 0268 Spline", t)
				BEq(obj.Join, JoinBevel, "poly 0268 Join", t)
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
				BEq(obj.Dash, DashSolid, "poly 0942 Dash", t)
				sEq(obj.Line, "#000000", "poly 0942 Line", t)
				iEq(obj.Width, 2, "poly 0942 Width", t)
				sEq(obj.Layer, "walls", "poly 0942 Layer", t)
				bEq(obj.Hidden, false, "poly 0942 Hidden", t)
				iEq(obj.Level, 0, "poly 0942 Level", t)
				sEq(obj.Group, "", "poly 0942 Group", t)
				// PolygonElement
				fEq(obj.Spline, 0, "poly 0942 Spline", t)
				BEq(obj.Join, JoinBevel, "poly 0942 Join", t)
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
				BEq(obj.Dash, DashSolid, "poly 0c36 Dash", t)
				sEq(obj.Line, "#000000", "poly 0c36 Line", t)
				iEq(obj.Width, 5, "poly 0c36 Width", t)
				sEq(obj.Layer, "walls", "poly 0c36 Layer", t)
				bEq(obj.Hidden, false, "poly 0c36 Hidden", t)
				iEq(obj.Level, 0, "poly 0c36 Level", t)
				sEq(obj.Group, "", "poly 0c36 Group", t)
				// PolygonElement
				fEq(obj.Spline, 0, "poly 0c36 Spline", t)
				BEq(obj.Join, JoinBevel, "poly 0c36 Join", t)
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
				BEq(obj.Dash, DashSolid, "arc 13a2 Dash", t)
				sEq(obj.Line, "black", "arc 13a2 Line", t)
				iEq(obj.Width, 5, "arc 13a2 Width", t)
				sEq(obj.Layer, "walls", "arc 13a2 Layer", t)
				bEq(obj.Hidden, false, "arc 13a2 Hidden", t)
				iEq(obj.Level, 0, "arc 13a2 Level", t)
				sEq(obj.Group, "", "arc 13a2 Group", t)
				// ArcElement
				BEq(obj.Arcmode, ArcModeArc, "arc 13a2 Arcmode", t)
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
				BEq(obj.Dash, DashSolid, "arc 5b1a Dash", t)
				sEq(obj.Line, "black", "arc 5b1a Line", t)
				iEq(obj.Width, 5, "arc 5b1a Width", t)
				sEq(obj.Layer, "walls", "arc 5b1a Layer", t)
				bEq(obj.Hidden, false, "arc 5b1a Hidden", t)
				iEq(obj.Level, 0, "arc 5b1a Level", t)
				sEq(obj.Group, "", "arc 5b1a Group", t)
				// ArcElement
				BEq(obj.Arcmode, ArcModePieSlice, "arc 5b1a Arcmode", t)
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
				BEq(obj.Dash, DashSolid, "arc 7b39 Dash", t)
				sEq(obj.Line, "black", "arc 7b39 Line", t)
				iEq(obj.Width, 5, "arc 7b39 Width", t)
				sEq(obj.Layer, "walls", "arc 7b39 Layer", t)
				bEq(obj.Hidden, false, "arc 7b39 Hidden", t)
				iEq(obj.Level, 0, "arc 57b39 Level", t)
				sEq(obj.Group, "", "arc 7b39 Group", t)
				// ArcElement
				BEq(obj.Arcmode, ArcModeChord, "arc 7b39 Arcmode", t)
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
				BEq(obj.Dash, DashLong, "line 2a17 Dash", t)
				sEq(obj.Line, "black", "line 2a17 Line", t)
				iEq(obj.Width, 5, "line 2a17 Width", t)
				sEq(obj.Layer, "walls", "line 2a17 Layer", t)
				bEq(obj.Hidden, false, "line 2a17 Hidden", t)
				iEq(obj.Level, 0, "line 2a17 Level", t)
				sEq(obj.Group, "", "line 2a17 Group", t)
				// LineElement
				BEq(obj.Arrow, ArrowFirst, "line 2a17 Arrow", t)
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
				BEq(obj.Dash, DashSolid, "line 38f6 Dash", t)
				sEq(obj.Line, "black", "line 38f6 Line", t)
				iEq(obj.Width, 5, "line 38f6 Width", t)
				sEq(obj.Layer, "walls", "line 38f6 Layer", t)
				bEq(obj.Hidden, false, "line 38f6 Hidden", t)
				iEq(obj.Level, 0, "line 38f6 Level", t)
				sEq(obj.Group, "", "line 38f6 Group", t)
				// LineElement
				BEq(obj.Arrow, ArrowNone, "line 38f6 Arrow", t)
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
				BEq(obj.Dash, DashLong, "line 61dc Dash", t)
				sEq(obj.Line, "black", "line 38f6 Line", t)
				iEq(obj.Width, 5, "line 61dc Width", t)
				sEq(obj.Layer, "walls", "line 61dc Layer", t)
				bEq(obj.Hidden, false, "line 61dc Hidden", t)
				iEq(obj.Level, 0, "line 61dc Level", t)
				sEq(obj.Group, "", "line 61dc Group", t)
				// LineElement
				BEq(obj.Arrow, ArrowFirst, "line 61dc Arrow", t)
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
				BEq(obj.Dash, DashSolid, "text 2c4a Dash", t)
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
				BEq(obj.Font.Weight, FontWeightNormal, "text 2c4a Font Weight", t)
				BEq(obj.Font.Slant, FontSlantRoman, "text 2c4a Font Slant", t)
				BEq(obj.Anchor, AnchorCenter, "text 2c4a Anchor", t)
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
				BEq(obj.Dash, DashSolid, "rect 3988 Dash", t)
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
				BEq(obj.Dash, DashSolid, "rect bf29 Dash", t)
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
				BEq(obj.Dash, DashSolid, "circ 2f5b Dash", t)
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
				BEq(obj.Dash, DashSolid, "aoe 4b82 Dash", t)
				sEq(obj.Line, "black", "aoe 4b82 Line", t)
				iEq(obj.Width, 5, "aoe 4b82 Width", t)
				sEq(obj.Layer, "walls", "aoe 4b82 Layer", t)
				bEq(obj.Hidden, false, "aoe 4b82 Hidden", t)
				iEq(obj.Level, 0, "aoe 4b82 Level", t)
				sEq(obj.Group, "", "aoe 4b82 Group", t)
				// SpellAreaOfEffectElement
				BEq(obj.AoEShape, AoEShapeRadius, "aoe 4b82 AoE Shape", t)
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
				BEq(obj.Dash, DashSolid, "aoe afd1 Dash", t)
				sEq(obj.Line, "black", "aoe afd1 Line", t)
				iEq(obj.Width, 5, "aoe afd1 Width", t)
				sEq(obj.Layer, "walls", "aoe afd1 Layer", t)
				bEq(obj.Hidden, false, "aoe afd1 Hidden", t)
				iEq(obj.Level, 0, "aoe afd1 Level", t)
				sEq(obj.Group, "", "aoe afd1 Group", t)
				// SpellAreaOfEffectElement
				BEq(obj.AoEShape, AoEShapeRay, "aoe afd1 AoE Shape", t)
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
				BEq(obj.Dash, DashSolid, "aoe e68d Dash", t)
				sEq(obj.Line, "black", "aoe e68d Line", t)
				iEq(obj.Width, 5, "aoe e68d Width", t)
				sEq(obj.Layer, "walls", "aoe e68d Layer", t)
				bEq(obj.Hidden, false, "aoe e68d Hidden", t)
				iEq(obj.Level, 0, "aoe e68d Level", t)
				sEq(obj.Group, "", "aoe e68d Group", t)
				// SpellAreaOfEffectElement
				BEq(obj.AoEShape, AoEShapeCone, "aoe e68d AoE Shape", t)
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
	if len(files) != 0 {
		t.Errorf("Expected no files but found %d", len(files))
	}
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

func sEq(a, b, msg string, t *testing.T) {
	if a != b {
		t.Errorf("%s: actual=\"%s\" expected=\"%s\"", msg, a, b)
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

/*
func TestAuthenticatorRoundLimits(t *testing.T) {
	a := Authenticator{}
	for i := 0; i < 1000; i++ {
		a.Reset()
		a.GenerateChallenge()
		if a.Challenge[0]&0xf0 != 0 {
			t.Errorf("iteration %d, challenge high byte is %x", i, a.Challenge[0])
		}
		if a.Challenge[1]&0x40 != 0x40 {
			t.Errorf("iteration %d, challenge low byte is %x", i, a.Challenge[1])
		}
	}
}

func TestAuthenticator(t *testing.T) {
	a := Authenticator{}

	// Not testing the quality of random numbers we get,
	// just that we get them. We're not testing the library
	// function, just our own code.
	b := a.Challenge
	a.GenerateChallenge()
	if bytesEqual(b, a.Challenge) {
		t.Errorf("Challenge value didn't change")
	}

	// Here are a few challenges and responses we checked with
	// the original Python implementation. We must get the same
	// results with them.
	type testcase struct {
		Nonce     []byte
		Challenge string
		Response  string
		Valid     bool
	}

	testcases := []testcase{
		{Nonce: []byte{0xbc, 0x2f, 0x21, 0x09, 0xcb, 0x4b, 0xd8, 0x38, 0xa4, 0xb4,
			0xad, 0xc8, 0x3f, 0xe0, 0xc5, 0x30, 0x5f, 0x96, 0x3b, 0xb7,
			0xca, 0x62, 0x44, 0xec, 0x73, 0x1a, 0x45, 0x4a, 0x16, 0x8e,
			0x26, 0xa2},
			Challenge: "vC8hCctL2DiktK3IP+DFMF+WO7fKYkTscxpFShaOJqI=",
			Response:  "kGlo/jsAtM8t1ebu6YmHkCCHk9iDvTTssUo+k2hv0xY=",
			Valid:     true},
		{Nonce: []byte{0xbd, 0x27, 0x95, 0x10, 0xf6, 0x74, 0x77, 0x1a, 0x9b, 0x7f,
			0xed, 0xfe, 0xc0, 0x5e, 0xad, 0xe8, 0xb2, 0x03, 0x95, 0x2f,
			0xb1, 0x48, 0xee, 0x33, 0x26, 0xad, 0xfe, 0x6d, 0xa8, 0x32,
			0x74, 0x15},
			Challenge: "vSeVEPZ0dxqbf+3+wF6t6LIDlS+xSO4zJq3+bagydBU=",
			Response:  "kLQnscqqYKXe1SV1SXmmzqXx4bEo+eV1eR5jS++rY4c=",
			Valid:     true},
		{Nonce: []byte{0x0c, 0x72, 0xba, 0x37, 0xc2, 0x67, 0x15, 0x07, 0x22, 0x07,
			0x99, 0xa4, 0x40, 0xd9, 0x9c, 0xab, 0xa1, 0xfe, 0xa7, 0x15,
			0x5e, 0x8d, 0x10, 0x35, 0x50, 0xa3, 0x01, 0x4a, 0x07, 0x4f,
			0x60, 0xa2},
			Challenge: "DHK6N8JnFQciB5mkQNmcq6H+pxVejRA1UKMBSgdPYKI=",
			Response:  "q8nncBMkpv+O3zVVx2hs7NKqugvW43Hpvj5k0XCsua8=",
			Valid:     true},
		{Nonce: []byte{0xe6, 0x33, 0x8f, 0x48, 0x67, 0x62, 0x9d, 0x5d, 0x52, 0x4a,
			0xe9, 0xd8, 0xbd, 0x09, 0xd0, 0xdc, 0x54, 0x6e, 0xb2, 0x9b,
			0x6f, 0xb5, 0xa6, 0x00, 0x92, 0x47, 0x3c, 0x09, 0x18, 0xeb,
			0xe0, 0x4d},
			Challenge: "5jOPSGdinV1SSunYvQnQ3FRusptvtaYAkkc8CRjr4E0=",
			Response:  "aA7Pj/w/ONgHMH5JoRkGQ1tks1vUei37PwML9Bgi3Ak=",
			Valid:     true},
		{Nonce: []byte{0xb7, 0xc9, 0x6e, 0xb7, 0x82, 0x09, 0xcc, 0x27, 0x50, 0x8e,
			0x8b, 0xb9, 0x7c, 0xb1, 0x09, 0x2f, 0xc5, 0xf3, 0xe0, 0xb8,
			0x67, 0x02, 0xd2, 0x0a, 0x92, 0xcc, 0x0b, 0x70, 0x38, 0x2d,
			0x85, 0x5e},
			Challenge: "t8lut4IJzCdQjou5fLEJL8Xz4LhnAtIKkswLcDgthV4=",
			Response:  "IcFg2BuQCdUsUHeVzfU6mylVVS/K7jalFX0DPFH8U7c=",
			Valid:     true},
		{Nonce: []byte{0x38, 0xfe, 0xde, 0x2b, 0x16, 0xc5, 0xe8, 0x72, 0xb9, 0x06,
			0xca, 0x51, 0x05, 0xe5, 0xbe, 0x1a, 0x74, 0x25, 0xcb, 0x8e,
			0x6f, 0xe2, 0xd6, 0xc1, 0xf3, 0x88, 0xa0, 0xe9, 0x30, 0x1c,
			0xa4, 0xbe},
			Challenge: "OP7eKxbF6HK5BspRBeW+GnQly45v4tbB84ig6TAcpL4=",
			Response:  "90MpQRjZfPgZZNqz6emBMmYIPEDjqijAjmcXg6W0GwI=",
			Valid:     true},
		{Nonce: []byte{0xd9, 0x82, 0xc0, 0x19, 0x18, 0x8d, 0x57, 0xff, 0xaf, 0xa7,
			0xb3, 0xe1, 0xde, 0xc4, 0xd7, 0x6b, 0x0c, 0xf2, 0xcc, 0x71,
			0x48, 0xb3, 0x34, 0xce, 0x11, 0x13, 0x09, 0xab, 0x73, 0x02,
			0x45, 0xaa},
			Challenge: "2YLAGRiNV/+vp7Ph3sTXawzyzHFIszTOERMJq3MCRao=",
			Response:  "K9cEUf9TaAMvcHM3hxzFASf9JYo4piTUHnQovvkbtZo=",
			Valid:     true},
		{Nonce: []byte{0xa0, 0x24, 0x8f, 0x85, 0x88, 0x2f, 0x58, 0xe2, 0x22, 0x6a,
			0x1c, 0x54, 0x4c, 0x65, 0x5e, 0xb5, 0x9b, 0xea, 0x39, 0xe8,
			0xd4, 0x13, 0x4a, 0xee, 0xae, 0xfc, 0x15, 0xde, 0xa8, 0x00,
			0xb1, 0xbf},
			Challenge: "oCSPhYgvWOIiahxUTGVetZvqOejUE0rurvwV3qgAsb8=",
			Response:  "FAaUh3YI6I3ovNP5L8zlX6u1QbkkGkwBKkGP90Nsb5Q=",
			Valid:     true},
		{Nonce: []byte{0xdb, 0x34, 0x8f, 0xfc, 0x46, 0x7b, 0xd2, 0x57, 0x83, 0x40,
			0x79, 0x50, 0x7a, 0x41, 0xae, 0xb9, 0x30, 0xb4, 0x01, 0x92,
			0xcf, 0x7f, 0x04, 0xc7, 0xa2, 0x30, 0x03, 0xc9, 0x24, 0x36,
			0x30, 0xdf},
			Challenge: "2zSP/EZ70leDQHlQekGuuTC0AZLPfwTHojADySQ2MN8=",
			Response:  "biRPri/TOSfT4nJ/IR78synR82NL1pihYaWsMRSfioo=",
			Valid:     true},
		{Nonce: []byte{0xe9, 0xb9, 0x61, 0x81, 0xf8, 0xc9, 0xcb, 0x7e, 0xb8, 0x9c,
			0xdb, 0xe6, 0xb5, 0x49, 0x48, 0x24, 0x74, 0xfb, 0x34, 0x05,
			0x85, 0x65, 0x88, 0xa3, 0x43, 0x6c, 0x34, 0x1f, 0xfd, 0x01,
			0xe6, 0x88},
			Challenge: "6blhgfjJy364nNvmtUlIJHT7NAWFZYijQ2w0H/0B5og=",
			Response:  "lNvqSPX7fLh6rsSLAfP1aZ5UZjxR5U9lLztyvrlUK9g=",
			Valid:     true},
		{Nonce: []byte{0xb5, 0x37, 0x54, 0x4d, 0xdc, 0xaf, 0x4b, 0xc5, 0xab, 0x66,
			0x0d, 0xde, 0xad, 0xae, 0x09, 0x83, 0x11, 0x35, 0x1f, 0x75,
			0x34, 0x6f, 0x97, 0x8b, 0x17, 0xa1, 0x40, 0xb0, 0x40, 0x06,
			0x74, 0x41},
			Challenge: "tTdUTdyvS8WrZg3era4JgxE1H3U0b5eLF6FAsEAGdEE=",
			Response:  "FBdzziLBdqDxgs0jBaTuEbckrCe1UyQucoA+PCV2Pbs=",
			Valid:     true},
		{Nonce: []byte{0x46, 0x02, 0x65, 0x0f, 0xaf, 0x6a, 0x45, 0x67, 0x7b, 0x1e,
			0x1b, 0x3e, 0xc1, 0xad, 0x3a, 0x66, 0x7c, 0x3c, 0x20, 0x3f,
			0x13, 0xaf, 0xb9, 0x93, 0xea, 0x17, 0x8d, 0x1c, 0xc1, 0x91,
			0x28, 0x20},
			Challenge: "RgJlD69qRWd7Hhs+wa06Znw8ID8Tr7mT6heNHMGRKCA=",
			Response:  "IDs7YP2vbhYUjuokMneDFumcd3GhOXikSQyjrMxTIcg=",
			Valid:     true},
		{Nonce: []byte{0xe4, 0xcf, 0x3f, 0x6a, 0xe8, 0xd1, 0x1e, 0x75, 0xf6, 0x77,
			0x71, 0xca, 0x49, 0xef, 0x5d, 0x8f, 0xa1, 0x85, 0x70, 0x58,
			0x30, 0xa7, 0xed, 0xc6, 0xd7, 0x32, 0xd1, 0xf3, 0xa6, 0xea,
			0x3b, 0xf8},
			Challenge: "5M8/aujRHnX2d3HKSe9dj6GFcFgwp+3G1zLR86bqO/g=",
			Response:  "nAwZNHISt+pV2zDa/BQPTxgyIkyl06BLNrjPL8EVqSM=",
			Valid:     true},
		{Nonce: []byte{0x01, 0x0e, 0x95, 0x6b, 0xc1, 0xe6, 0x23, 0x23, 0x52, 0x3e,
			0x62, 0xd5, 0x76, 0x7c, 0xeb, 0x18, 0xa7, 0x2d, 0xdb, 0x7f,
			0x6c, 0xab, 0xfa, 0x72, 0x71, 0xd1, 0x31, 0xa6, 0x38, 0xc9,
			0x06, 0x5c},
			Challenge: "AQ6Va8HmIyNSPmLVdnzrGKct239sq/pycdExpjjJBlw=",
			Response:  "vjjJTHzBVSva0Zd2XzNFOpDCTTv/LSQWc3rVnaejdvo=",
			Valid:     true},
		{Nonce: []byte{0xb8, 0xc0, 0x0a, 0x6a, 0x55, 0xd3, 0xcf, 0x47, 0xa8, 0xa8,
			0x55, 0xaf, 0x5c, 0x72, 0x4b, 0xad, 0x55, 0xf6, 0x75, 0xb1,
			0xfd, 0x32, 0xf8, 0x6c, 0x4a, 0x23, 0x6c, 0xc2, 0x6e, 0x06,
			0x13, 0x78},
			Challenge: "uMAKalXTz0eoqFWvXHJLrVX2dbH9MvhsSiNswm4GE3g=",
			Response:  "YlMRscFkHyPiSaC8RbtS9kepY+ituzGZIAwnmayR0oo=",
			Valid:     true},
		{Nonce: []byte{0x6a, 0xbe, 0xe8, 0xb3, 0x1b, 0x93, 0x5d, 0x86, 0xfb, 0x88,
			0xf9, 0x32, 0xdf, 0xf8, 0xe6, 0xc1, 0xe1, 0x75, 0x8a, 0x05,
			0x6b, 0xcd, 0xa8, 0xc8, 0xcb, 0xf3, 0x77, 0x93, 0xee, 0x4e,
			0x57, 0x9d},
			Challenge: "ar7osxuTXYb7iPky3/jmweF1igVrzajIy/N3k+5OV50=",
			Response:  "t5QHcwGQc10/atdWVxoYBHfNg94dCpflDuo4VgSSuMw=",
			Valid:     true},
		{Nonce: []byte{0x68, 0x91, 0x75, 0xbd, 0xda, 0x87, 0x6b, 0x4d, 0x81, 0xbd,
			0x2d, 0xd0, 0x39, 0x8c, 0xe3, 0x25, 0x49, 0x32, 0x2c, 0xea,
			0x37, 0xde, 0x15, 0x48, 0x7e, 0xc6, 0x9a, 0xec, 0x65, 0x5e,
			0x40, 0xa6},
			Challenge: "aJF1vdqHa02BvS3QOYzjJUkyLOo33hVIfsaa7GVeQKY=",
			Response:  "NZBLdygw/v6e9aj5pclbb6gZKIsvaMQ8y0dLMcnZX18=",
			Valid:     true},
		{Nonce: []byte{0x1d, 0xbf, 0x14, 0x55, 0xc4, 0x0b, 0xff, 0x03, 0x49, 0xb6,
			0xf2, 0x49, 0xdf, 0x86, 0xfe, 0x58, 0xf0, 0x9c, 0x73, 0xfd,
			0xfc, 0xa0, 0x33, 0xd8, 0xfd, 0x59, 0xb2, 0x7e, 0xad, 0x12,
			0x1a, 0xd3},
			Challenge: "Hb8UVcQL/wNJtvJJ34b+WPCcc/38oDPY/Vmyfq0SGtM=",
			Response:  "wICpk7dVj09ve+GMesM+XsI78LNoGJ0LdcdN6jE8NJ0=",
			Valid:     true},
		{Nonce: []byte{0x22, 0xdf, 0x2c, 0xde, 0xce, 0xcd, 0xee, 0x56, 0x6a, 0xdc,
			0xdf, 0xe6, 0x3f, 0x0c, 0x60, 0xf4, 0x20, 0x1b, 0x20, 0xc1,
			0x48, 0x52, 0x6e, 0xb5, 0x0b, 0xfa, 0x56, 0x2b, 0x50, 0xed,
			0x83, 0x6f},
			Challenge: "It8s3s7N7lZq3N/mPwxg9CAbIMFIUm61C/pWK1Dtg28=",
			Response:  "D6M4hQxL7jCiopms6tePDC5PFpwPxw7xyh9ZgFiWOjU=",
			Valid:     true},
		{Nonce: []byte{0x5e, 0x3f, 0xcb, 0xad, 0xe0, 0x69, 0xd1, 0x24, 0x20, 0x81,
			0x35, 0x94, 0x00, 0xeb, 0xb7, 0xc6, 0xb3, 0x79, 0xe2, 0xd4,
			0x05, 0x1e, 0x41, 0x04, 0x6f, 0xc7, 0x66, 0xb7, 0xba, 0xf3,
			0x12, 0x25},
			Challenge: "Xj/LreBp0SQggTWUAOu3xrN54tQFHkEEb8dmt7rzEiU=",
			Response:  "H5UP7uK+GbEbDHf7rxdUTKg86hPq7zWj/Ymj6iODc6I=",
			Valid:     true},
		{Nonce: []byte{0xea, 0xc2, 0x79, 0xa5, 0x0c, 0x6c, 0x40, 0x0a, 0xd5, 0x28,
			0x67, 0x98, 0x38, 0xae, 0xcc, 0xf0, 0x2e, 0xf8, 0xa8, 0x36,
			0xb2, 0x91, 0x51, 0x9a, 0xf8, 0xa5, 0xdb, 0x48, 0xfb, 0x63,
			0x4d, 0x36},
			Challenge: "6sJ5pQxsQArVKGeYOK7M8C74qDaykVGa+KXbSPtjTTY=",
			Response:  "Xl6P/M1EcEjtAnbYEww/RasdeDqBJ5Qt4JtH2WM5glw=",
			Valid:     true},
		{Nonce: []byte{0x83, 0x2f, 0xc0, 0x18, 0xcb, 0xde, 0xc4, 0xbb, 0xcf, 0x2f,
			0x14, 0xfa, 0x85, 0xd1, 0xcb, 0xed, 0x30, 0x8e, 0x69, 0xf7,
			0x64, 0xba, 0xad, 0x11, 0xbb, 0xd1, 0x5a, 0x98, 0xad, 0x85,
			0xb3, 0x44},
			Challenge: "gy/AGMvexLvPLxT6hdHL7TCOafdkuq0Ru9FamK2Fs0Q=",
			Response:  "GnzzJSuSG66apoQD1nkKI8nYQ0i5Ip1rkGgQTojzM98=",
			Valid:     true},
		{Nonce: []byte{0x33, 0x87, 0xcf, 0x20, 0x13, 0xee, 0x29, 0x4b, 0x83, 0xb1,
			0x21, 0x77, 0xd9, 0x5e, 0x64, 0xe8, 0x8c, 0x8d, 0xf8, 0x11,
			0xe1, 0x40, 0x62, 0x56, 0x7e, 0xe8, 0xc0, 0x6a, 0xec, 0xfb,
			0xd5, 0x64},
			Challenge: "M4fPIBPuKUuDsSF32V5k6IyN+BHhQGJWfujAauz71WQ=",
			Response:  "lv8L1XFm+Z8sNEbcg1hyO9CxXobAjvI+Rs0lc3JhnpQ=",
			Valid:     true},
		{Nonce: []byte{0x82, 0x10, 0x22, 0xdb, 0x56, 0xf2, 0xb9, 0x4c, 0x2c, 0x5b,
			0x34, 0x7f, 0xda, 0x6f, 0xb2, 0xf9, 0x68, 0x39, 0x55, 0xb1,
			0x57, 0xdb, 0xa1, 0x75, 0x93, 0x51, 0x7a, 0x09, 0x05, 0xb6,
			0x42, 0xf7},
			Challenge: "ghAi21byuUwsWzR/2m+y+Wg5VbFX26F1k1F6CQW2Qvc=",
			Response:  "ii2qgX820Muret7RinJNf0hrer//XGlP0hZ3dETILw0=",
			Valid:     true},
		{Nonce: []byte{0xa7, 0x1d, 0x0a, 0xfe, 0x7e, 0x90, 0x9a, 0x16, 0x80, 0x79,
			0x52, 0xc9, 0x2c, 0xb1, 0x57, 0x45, 0xc4, 0x14, 0x60, 0x13,
			0x0f, 0xba, 0x15, 0x97, 0x2b, 0x68, 0x20, 0xd5, 0x3b, 0x78,
			0xd2, 0xb0},
			Challenge: "px0K/n6QmhaAeVLJLLFXRcQUYBMPuhWXK2gg1Tt40rA=",
			Response:  "1KMl3BDV/+opDA/+J5mg9AjDjCZILRc+yRQl+taxiIE=",
			Valid:     true},
		{Nonce: []byte{0xc1, 0x92, 0xab, 0xb4, 0xd6, 0xb4, 0x7d, 0x12, 0x7b, 0xc4,
			0xf2, 0x13, 0x67, 0xdd, 0xf8, 0xef, 0xe0, 0x38, 0xe1, 0x03,
			0xba, 0x01, 0x41, 0xa5, 0x73, 0x1b, 0xea, 0x4e, 0xb8, 0xc8,
			0x7b, 0x14},
			Challenge: "wZKrtNa0fRJ7xPITZ9347+A44QO6AUGlcxvqTrjIexQ=",
			Response:  "UF9XDxqK8/Fr+CF3gFHuyDRFUy7skuRceR5gRiHbdkU=",
			Valid:     true},
		{Nonce: []byte{0xa2, 0xa7, 0x72, 0xb1, 0x49, 0x36, 0x96, 0x44, 0x44, 0x20,
			0x8d, 0x10, 0xba, 0xbe, 0x80, 0x0d, 0xa1, 0xd0, 0xd9, 0xb1,
			0xbc, 0xfe, 0x2e, 0x37, 0xd5, 0x7f, 0xd6, 0xb8, 0x89, 0xd9,
			0xd3, 0x58},
			Challenge: "oqdysUk2lkREII0Qur6ADaHQ2bG8/i431X/WuInZ01g=",
			Response:  "WtkmdJsXC5AnJq0BtlsndqE3H+7NNfYiZIGDvnar79I=",
			Valid:     true},
		{Nonce: []byte{0x3b, 0x59, 0x43, 0x7d, 0x94, 0x58, 0x75, 0x0c, 0xfa, 0x63,
			0x7a, 0x34, 0x45, 0x73, 0xe3, 0x22, 0x87, 0x82, 0xff, 0xbc,
			0x4d, 0x88, 0x9d, 0xb2, 0x79, 0x21, 0x83, 0xb0, 0xb3, 0x54,
			0xde, 0xbb},
			Challenge: "O1lDfZRYdQz6Y3o0RXPjIoeC/7xNiJ2yeSGDsLNU3rs=",
			Response:  "UqcSDjtNmgG4kVViGp0/qWCKNLTvVlQrpRN4WbQ9R8k=",
			Valid:     true},
		{Nonce: []byte{0xd6, 0x15, 0x8a, 0xa0, 0x29, 0x09, 0x3c, 0xbd, 0x63, 0xf3,
			0x92, 0xe2, 0xac, 0xf6, 0x3e, 0x62, 0xa2, 0x27, 0x5d, 0x12,
			0x7f, 0xdf, 0xb1, 0x4e, 0x21, 0x30, 0xb7, 0xd7, 0x97, 0xc4,
			0x89, 0xb0},
			Challenge: "1hWKoCkJPL1j85LirPY+YqInXRJ/37FOITC315fEibA=",
			Response:  "Lv8xlK30r9tt9eebsyvyGD2aL4l/SuD82gMNRm34QnM=",
			Valid:     true},
		{Nonce: []byte{0x2e, 0x17, 0xe9, 0x59, 0xfc, 0xd3, 0x93, 0x00, 0x2a, 0xe1,
			0xd6, 0xd8, 0x3a, 0xc4, 0x03, 0x28, 0x75, 0x70, 0xa6, 0x73,
			0xc5, 0xc7, 0xd9, 0xb2, 0x7e, 0x46, 0xea, 0x8d, 0xe9, 0xde,
			0x6a, 0x8b},
			Challenge: "LhfpWfzTkwAq4dbYOsQDKHVwpnPFx9myfkbqjeneaos=",
			Response:  "K/j0yJqX0U+VUiurE6lrrnrc67i4FVOpAfyfFtp/q4I=",
			Valid:     true},
		{Nonce: []byte{0x76, 0x03, 0xfb, 0x25, 0x02, 0xfc, 0x1f, 0xd3, 0xdc, 0x97,
			0xe1, 0xf1, 0x51, 0x04, 0x57, 0xe7, 0x5e, 0xd0, 0xc4, 0x2f,
			0x58, 0xeb, 0x50, 0xaf, 0x35, 0x94, 0xa5, 0x59, 0x42, 0xf8,
			0x6e, 0x36},
			Challenge: "dgP7JQL8H9Pcl+HxUQRX517QxC9Y61CvNZSlWUL4bjY=",
			Response:  "lOxFWFcyFIgyryopBEnDHHhAI0ZlYUIW0V0Mo5bZRW4=",
			Valid:     true},
		{Nonce: []byte{0xae, 0x7d, 0x20, 0x5e, 0x25, 0x7b, 0x18, 0xf0, 0x77, 0x29,
			0x8b, 0xb6, 0x21, 0xbc, 0x7f, 0xf9, 0x21, 0x4b, 0x22, 0x0d,
			0xe4, 0x17, 0xf7, 0x81, 0xb9, 0x5f, 0x5d, 0xc9, 0x5f, 0xc6,
			0x2e, 0xa9},
			Challenge: "rn0gXiV7GPB3KYu2Ibx/+SFLIg3kF/eBuV9dyV/GLqk=",
			Response:  "B+I5rUig6jSb3i+kBUNWkz2qHm28UMtOG9pWT12+zVg=",
			Valid:     true},
		{Nonce: []byte{0xb9, 0xa4, 0x13, 0x04, 0xcf, 0x54, 0xe3, 0xec, 0xab, 0x51,
			0xcc, 0x45, 0xfb, 0x81, 0x8d, 0xe6, 0x56, 0x60, 0xb6, 0xd7,
			0xda, 0x7b, 0xdd, 0xca, 0x48, 0x54, 0x23, 0x24, 0x36, 0xea,
			0x7a, 0x0a},
			Challenge: "uaQTBM9U4+yrUcxF+4GN5lZgttfae93KSFQjJDbqego=",
			Response:  "eNmLUgWYdqohlqFuCT9epMpcERYqfLBw/Az4SDRidHs=",
			Valid:     true},
		{Nonce: []byte{0xbf, 0xd5, 0x53, 0x74, 0xc8, 0xf0, 0xf5, 0xb2, 0x5c, 0x30,
			0x78, 0x27, 0x57, 0xf1, 0x8b, 0xef, 0x1c, 0x07, 0x04, 0x78,
			0xb5, 0x41, 0x3d, 0x43, 0x16, 0x69, 0xfc, 0xf6, 0x77, 0xed,
			0xc1, 0xcd},
			Challenge: "v9VTdMjw9bJcMHgnV/GL7xwHBHi1QT1DFmn89nftwc0=",
			Response:  "PabyQnr3KtQIgGVAokz0SyIDublQ0rho8lWcBUsfQ9s=",
			Valid:     true},
		{Nonce: []byte{0x5f, 0xd2, 0x5e, 0x5b, 0x32, 0xcf, 0x9d, 0xf1, 0x30, 0xed,
			0x9c, 0xfc, 0x37, 0x66, 0xdd, 0x79, 0x73, 0x10, 0x35, 0x91,
			0x8f, 0xb4, 0x3a, 0x36, 0x33, 0xa2, 0x93, 0x0e, 0x97, 0x3c,
			0xcb, 0x05},
			Challenge: "X9JeWzLPnfEw7Zz8N2bdeXMQNZGPtDo2M6KTDpc8ywU=",
			Response:  "c9d0pS1yROJb4wevLycIXCXW1BCb/TlmAc7KFiq9v6o=",
			Valid:     true},
		{Nonce: []byte{0x75, 0xbe, 0xab, 0xd2, 0x26, 0xa4, 0xac, 0x64, 0xb8, 0x7d,
			0x60, 0x8a, 0x2a, 0x72, 0x61, 0x84, 0x06, 0xd0, 0xb6, 0x80,
			0xf1, 0x58, 0x7e, 0xb4, 0xa8, 0x92, 0x3d, 0x38, 0xda, 0x90,
			0xf1, 0xd1},
			Challenge: "db6r0iakrGS4fWCKKnJhhAbQtoDxWH60qJI9ONqQ8dE=",
			Response:  "4rgUB+OyNjwKmTgK2LeIs4VUAJlbR1ACCsHc5lIP9uU=",
			Valid:     true},
		{Nonce: []byte{0xad, 0x71, 0xc8, 0x5f, 0xed, 0x4e, 0x02, 0x32, 0xaa, 0xdd,
			0xc6, 0xea, 0xea, 0xa5, 0xae, 0xf2, 0xa6, 0x1d, 0x22, 0xf7,
			0x17, 0xa9, 0x76, 0xbe, 0x09, 0x36, 0x1b, 0x57, 0x14, 0x89,
			0x3a, 0x01},
			Challenge: "rXHIX+1OAjKq3cbq6qWu8qYdIvcXqXa+CTYbVxSJOgE=",
			Response:  "rJqypsyUpjqMGQcq11+EVa9Z8ncmiKizr/Jx3tLLBJE=",
			Valid:     true},
		{Nonce: []byte{0x2e, 0x5a, 0xd7, 0x69, 0x20, 0xd9, 0x7b, 0xd3, 0x3c, 0x52,
			0xdc, 0x67, 0x08, 0xcc, 0xcb, 0xcf, 0x95, 0x4a, 0x04, 0x20,
			0xb7, 0x3f, 0x63, 0xb8, 0x47, 0x48, 0xaa, 0x6c, 0x41, 0xdd,
			0x2e, 0xb3},
			Challenge: "LlrXaSDZe9M8UtxnCMzLz5VKBCC3P2O4R0iqbEHdLrM=",
			Response:  "p33GLC6m/XQy/fytOr+9GWp043m7U1nRm/wmXPw0V98=",
			Valid:     true},
		{Nonce: []byte{0xb5, 0xb9, 0x4a, 0x2a, 0xba, 0xaa, 0x34, 0xf5, 0xd4, 0xa8,
			0xca, 0xec, 0xb0, 0x5a, 0x62, 0x8c, 0x59, 0x13, 0xbd, 0xcd,
			0xd2, 0xfd, 0x9c, 0xed, 0xc8, 0x66, 0xa9, 0x3e, 0x7b, 0x9b,
			0x45, 0x18},
			Challenge: "tblKKrqqNPXUqMrssFpijFkTvc3S/ZztyGapPnubRRg=",
			Response:  "GDxfHqFYKb/hloffJtUQZmfjpxQvJLJlkAMHgG0+rok=",
			Valid:     true},
		{Nonce: []byte{0x38, 0x18, 0x08, 0x19, 0x30, 0x90, 0xe0, 0x9a, 0x8d, 0xdb,
			0x04, 0x43, 0x01, 0x34, 0x55, 0x0c, 0x14, 0x0c, 0xe2, 0x88,
			0x99, 0xda, 0x68, 0x92, 0xba, 0x02, 0x53, 0x91, 0xbd, 0xbd,
			0xb7, 0xba},
			Challenge: "OBgIGTCQ4JqN2wRDATRVDBQM4oiZ2miSugJTkb29t7o=",
			Response:  "4jqlQysIG4MV77I9DnAvr2nfQ6RS5eJYRH4G+dIzLlQ=",
			Valid:     true},
		{Nonce: []byte{0x7b, 0xa5, 0xc6, 0x19, 0x0c, 0x9b, 0x73, 0x81, 0x94, 0xc7,
			0x3e, 0xc6, 0x12, 0x66, 0x7d, 0xcc, 0x66, 0x04, 0x63, 0xbb,
			0x77, 0x5e, 0x70, 0x9f, 0xac, 0xae, 0x3f, 0x6b, 0xb3, 0xa3,
			0x30, 0x55},
			Challenge: "e6XGGQybc4GUxz7GEmZ9zGYEY7t3XnCfrK4/a7OjMFU=",
			Response:  "xNqK9/ItpZNLVXtw0Fc/47q21oko7LXF7mZcoWhFDAM=",
			Valid:     true},
		{Nonce: []byte{0xba, 0x51, 0xd1, 0x51, 0xfa, 0xca, 0x47, 0x48, 0x2c, 0xfa,
			0x02, 0x52, 0xf0, 0xca, 0xcf, 0x1f, 0x8f, 0xe0, 0x06, 0x3c,
			0x7e, 0x96, 0xc3, 0x60, 0x97, 0x7d, 0xb9, 0x58, 0x3b, 0x7a,
			0xa6, 0x95},
			Challenge: "ulHRUfrKR0gs+gJS8MrPH4/gBjx+lsNgl325WDt6ppU=",
			Response:  "ejWeH/uuCgQnFyd5hb95+SVq5qoXVJl67D+xns0YJcE=",
			Valid:     true},
		{Nonce: []byte{0xd2, 0xb0, 0x00, 0x6f, 0x2d, 0xa1, 0x96, 0x11, 0x06, 0x3e,
			0x89, 0x44, 0x97, 0xfd, 0x0c, 0x8d, 0x5a, 0xd2, 0x32, 0x05,
			0xc0, 0x4c, 0x75, 0xf2, 0x49, 0xf9, 0x1f, 0x67, 0x8c, 0x4d,
			0x58, 0xfd},
			Challenge: "0rAAby2hlhEGPolEl/0MjVrSMgXATHXySfkfZ4xNWP0=",
			Response:  "MvDMFmFH3rq2jRxWkdeWHYfOQ92/xKfgOq37/8THhhU=",
			Valid:     true},
		{Nonce: []byte{0x8e, 0xd2, 0xcd, 0xf9, 0x5c, 0x99, 0xc7, 0x1e, 0x75, 0x4f,
			0x3d, 0x55, 0xb8, 0x10, 0xf5, 0x28, 0x7f, 0xcb, 0x2f, 0x07,
			0x87, 0x0b, 0x53, 0x37, 0xb8, 0x2c, 0x69, 0xca, 0xbe, 0x60,
			0xcc, 0xd8},
			Challenge: "jtLN+VyZxx51Tz1VuBD1KH/LLweHC1M3uCxpyr5gzNg=",
			Response:  "RQ/g9I+ZDVcdeY/qhwEFSDwfx0QYI8h7TsgCqg4hQqw=",
			Valid:     true},
		{Nonce: []byte{0x4c, 0xe0, 0xe3, 0xdb, 0x6b, 0x5f, 0x0e, 0xfe, 0xd0, 0x7b,
			0x97, 0x01, 0x7c, 0xd2, 0xe0, 0x22, 0x05, 0x40, 0x98, 0x01,
			0x48, 0xb0, 0x84, 0x0a, 0x9a, 0x17, 0x84, 0x2e, 0x03, 0x97,
			0x2a, 0xff},
			Challenge: "TODj22tfDv7Qe5cBfNLgIgVAmAFIsIQKmheELgOXKv8=",
			Response:  "Lh/AP5pB1dataeIfxJqg1pV+pDKFTh0wLsi1Gd/9TRU=",
			Valid:     true},
		{Nonce: []byte{0x09, 0x19, 0xe2, 0x02, 0x8d, 0x04, 0x3a, 0x60, 0x50, 0x02,
			0xb9, 0x29, 0x6e, 0x4d, 0xdb, 0x34, 0xb2, 0x9d, 0x38, 0xe2,
			0x21, 0xf0, 0xb5, 0x3a, 0x2f, 0xfe, 0x4e, 0x5b, 0x93, 0x83,
			0x52, 0x72},
			Challenge: "CRniAo0EOmBQArkpbk3bNLKdOOIh8LU6L/5OW5ODUnI=",
			Response:  "7NSwdNpZjUMLn5KEFyF+p6O1f0Z8h6I26+Lmh/TK7Ew=",
			Valid:     true},
		{Nonce: []byte{0x86, 0x82, 0x73, 0xfa, 0x17, 0xfd, 0xba, 0x1d, 0xe8, 0x54,
			0x7c, 0x21, 0x51, 0x4e, 0xa2, 0xe5, 0x84, 0xd6, 0x86, 0x64,
			0xec, 0x77, 0x67, 0x04, 0x75, 0xbe, 0x58, 0x79, 0x5e, 0x2e,
			0xad, 0xe7},
			Challenge: "hoJz+hf9uh3oVHwhUU6i5YTWhmTsd2cEdb5YeV4urec=",
			Response:  "6xyuzY17NeQVz75/7+xe/GczckJ75o0LuYsUCtUq0O0=",
			Valid:     true},
		{Nonce: []byte{0x87, 0x05, 0xf9, 0xe0, 0x9a, 0x29, 0x6d, 0x3f, 0xd2, 0xd7,
			0xb1, 0xd6, 0x1e, 0x56, 0x83, 0x44, 0xfb, 0x5c, 0xfb, 0xcb,
			0xa1, 0xe5, 0xce, 0x62, 0x2f, 0x74, 0xbe, 0x97, 0xad, 0x44,
			0x1f, 0x06},
			Challenge: "hwX54JopbT/S17HWHlaDRPtc+8uh5c5iL3S+l61EHwY=",
			Response:  "S41H0GR7FRmTo1FI/HoxA6zqpeh24IcHVrcnM9kQHx0=",
			Valid:     true},
		{Nonce: []byte{0xa0, 0x64, 0x7e, 0x2a, 0x34, 0xca, 0xc5, 0xf8, 0xc8, 0x79,
			0xde, 0x40, 0x96, 0x35, 0xaf, 0x8b, 0x5e, 0x77, 0xd6, 0x39,
			0xff, 0xba, 0xf0, 0x7b, 0x9e, 0xa6, 0x78, 0x86, 0x7f, 0xb6,
			0x30, 0x5b},
			Challenge: "oGR+KjTKxfjIed5AljWvi1531jn/uvB7nqZ4hn+2MFs=",
			Response:  "sTRxAtf7/YtikA7VVK2NjqaAbRke8LrYPz++14sUIks=",
			Valid:     true},
		{Nonce: []byte{0x1d, 0xee, 0xbf, 0x74, 0xf2, 0x0b, 0x46, 0x68, 0xda, 0xad,
			0x6d, 0x28, 0x3a, 0xd6, 0x37, 0xe4, 0x5f, 0x3d, 0x53, 0x56,
			0xe8, 0x00, 0xf3, 0x6b, 0x1b, 0xb6, 0x23, 0x8e, 0xec, 0xdd,
			0x53, 0xd8},
			Challenge: "He6/dPILRmjarW0oOtY35F89U1boAPNrG7YjjuzdU9g=",
			Response:  "Pdc8yR7DVOjqQlPZtgH21Ein6OJIW6job6me6uVfZzY=",
			Valid:     true},
		{Nonce: []byte{0x1d, 0xee, 0xbf, 0x74, 0xf2, 0x0b, 0x46, 0x68, 0xda, 0xad,
			0x6d, 0x28, 0x3a, 0xd6, 0x37, 0xe4, 0x5f, 0x3d, 0x53, 0x56,
			0xe8, 0x00, 0xf3, 0x6b, 0x1b, 0xb6, 0x23, 0x8e, 0xec, 0xdd,
			0x53, 0xd8},
			Challenge: "He6/dPILRmjarW0oOtY35F89U1boAPNrG7YjjuzdU9g=",
			Response:  "Pdc8yR7DVOjqQlPZtgH21Ein6OJIW6job6me6uVfZzY=",
			Valid:     true},
		{Nonce: []byte{0x1d, 0xee, 0xbf, 0x74, 0xf2, 0x0b, 0x46, 0x68, 0xda, 0xad,
			0x6d, 0x28, 0x3a, 0xd6, 0x37, 0xe4, 0x5f, 0x3d, 0x53, 0x56,
			0xe8, 0x00, 0xf3, 0x6b, 0x1b, 0xb6, 0x23, 0x8e, 0xec, 0xdd,
			0x53, 0xd8},
			Challenge: "He6/dPILRmjarW0oOtY35F89U1boAPNrG7YjjuzdU9g=",
			Response:  "Pdc8yR7DVOjqQlPZtgH2aEin6OJIW6job6me6uVfZzY=",
			Valid:     false},
	}
	for i, test := range testcases {
		a := Authenticator{Secret: []byte("abc123**sekret**XXXyyyZZZ")}
		a.Challenge = test.Nonce
		if a.CurrentChallenge() != test.Challenge {
			t.Errorf("Challenge iteration %d was %s, expected %s", i, a.CurrentChallenge(), test.Challenge)
		}
		r, err := a.ValidateResponse(test.Response)
		if err != nil {
			t.Errorf("iteration %d, error %v in validation", i, err)
		}
		if r != test.Valid {
			t.Errorf("iteration %d, got %v result, expected %v", i, r, test.Valid)
		}
	}
}
*/

// @[00]@| GMA 4.3.3
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
