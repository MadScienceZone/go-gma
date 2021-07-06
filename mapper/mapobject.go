/*
XXX add use of contexts to this to allow cancelling
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
// MapObject describes the elements that may appear on the map.
//

package mapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fizban-of-ragnarok/go-gma/v4/tcllist"
)

//
// The GMA File Format version number current as of this build.
//
const GMAMapperFileFormat = 16 // @@##@@ auto-configured
const MINIMUM_SUPPORTED_MAP_FILE_FORMAT = 14
const MAXIMUM_SUPPORTED_MAP_FILE_FORMAT = 16

func init() {
	if MINIMUM_SUPPORTED_MAP_FILE_FORMAT > GMAMapperFileFormat || MAXIMUM_SUPPORTED_MAP_FILE_FORMAT < GMAMapperFileFormat {
		if MINIMUM_SUPPORTED_MAP_FILE_FORMAT == MAXIMUM_SUPPORTED_MAP_FILE_FORMAT {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapper only supports file format %v, but version %v was the official one when this package was released!", MINIMUM_SUPPORTED_MAP_FILE_FORMAT, GMAMapperFileFormat))
		} else {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapper only supports mapper file formats %v-%v, but version %v was the official one when this package was released!", MINIMUM_SUPPORTED_MAP_FILE_FORMAT, MAXIMUM_SUPPORTED_MAP_FILE_FORMAT, GMAMapperFileFormat))
		}
	}
}

type PolymorphSizes struct {
}

type MapObject interface {
	ObjID() string
}

type BaseMapObject struct {
	ID string
}

type Coordinates struct {
	// The (x,y) coordinates for the reference point of this element on the map.
	// These are in standard map pixel units (50 pixels = 5 feet).
	X, Y float64
}

type MapElement struct {
	BaseMapObject
	Coordinates

	// The z "coordinate" is the vertical stacking order relative to the other
	// displayed on-screen objects.
	Z int

	Locked bool
	Points []Coordinates
	Fill   string
	Dash   byte
	Line   string
	Width  int
	Layer  string
	Hidden bool
	Level  int
	Group  string
}

const (
	ArcModePieSlice = iota
	ArcModeArc
	ArcModeChord
)

type ArcElement struct {
	MapElement
	Arcmode byte
	Extent  float64
	Start   float64
}

type SpellAreaOfEffectElement struct {
	MapElement
	AoEShape byte
}

type CircleElement struct {
	MapElement
}

//type ElementGroup struct {
//}

//type MapLayer struct {
//}

type LineElement struct {
	MapElement
	Arrow byte
}

type PolygonElement struct {
	MapElement
	Spline float64
	Join   byte
}

type RectangleElement struct {
	MapElement
}

const (
	FontWeightNormal = iota
	FontWeightBold
)

const (
	FontSlantRoman = iota
	FontSlantItalic
)

type TextFont struct {
	Family string
	Size   float64
	Weight byte
	Slant  byte
}

type TextElement struct {
	MapElement
	Text   string
	Font   TextFont
	Anchor byte
}

type TileElement struct {
	MapElement
	Image string
}

type CreatureHealth struct {
	MaxHP           int
	LethalDamage    int
	NonLethalDamage int
	Con             int
	IsFlatFooted    bool
	IsStable        bool
	Condition       string
	HpBlur          int
}

type RadiusAoE struct {
	Radius float64
	Color  string
}

type CreatureToken struct {
	BaseMapObject
	Health     *CreatureHealth
	Name       string
	Gx         float64
	Gy         float64
	Skin       int
	SkinSize   []string
	Elev       int
	Color      string
	Note       string
	Size       string
	StatusList []string
	AoE        *RadiusAoE
	Area       string
	MoveMode   byte
	Reach      bool
	Killed     bool
	Dim        bool
}

type PlayerToken struct {
	CreatureToken
}

type MonsterToken struct {
	CreatureToken
}

type ImageDefinition struct {
	// The zoom (magnification) level this bitmap represents for the given
	// image.
	Zoom float64
	// The name of the image as known within the mapper.
	Name string
	// The filename by which the image can be retrieved.
	File string
	// If IsLocalFile is true, File is the name of the image file on disk;
	// otherwise it is the server's internal ID by which you may request
	// that file from the server.
	IsLocalFile bool
}

type FileDefinition struct {
	// The filename or Server ID.
	File string
	// If IsLocalFile is true, File is the name of the file on disk;
	// otherwise it is the server's internal ID by which you may request
	// that file from the server.
	IsLocalFile bool
}

func (o MonsterToken) ObjID() string             { return o.ID }
func (o PlayerToken) ObjID() string              { return o.ID }
func (o ArcElement) ObjID() string               { return o.ID }
func (o CircleElement) ObjID() string            { return o.ID }
func (o LineElement) ObjID() string              { return o.ID }
func (o PolygonElement) ObjID() string           { return o.ID }
func (o RectangleElement) ObjID() string         { return o.ID }
func (o TextElement) ObjID() string              { return o.ID }
func (o TileElement) ObjID() string              { return o.ID }
func (o SpellAreaOfEffectElement) ObjID() string { return o.ID }

//
// Read the contents of a map file as a slice of lines as read from a disk
// file or received from the server. These lines are parsed to construct
// the set of map objects recorded in them. These are returned along with
// the image and file definitions declared in the data stream.
//
// The image list is a map
// of "imagename:zoom" to ImageDefinition structures.
//
func ParseObjects(dataStream []string) ([]MapObject, map[string]ImageDefinition, []FileDefinition, error) {
	format := 0
	images := make(map[string]ImageDefinition)
	objects := make(map[string]map[string][]string)
	var files []FileDefinition

	for lineNo, incomingAttribute := range dataStream {
		fields, err := tcllist.ParseTclList(incomingAttribute)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: %v at \"%s\"", lineNo, err, incomingAttribute)
		}
		if len(fields) < 2 {
			return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: not enough fields in \"%s\"", lineNo, incomingAttribute)
		}

		//
		// __MAPPER__:<version> <comment>
		// file header
		//
		if strings.HasPrefix(fields[0], "__MAPPER__:") {
			if len(fields[0]) > 11 {
				f, err := strconv.Atoi(fields[0][11:])
				if err != nil {
					return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: Unable to parse map file version: %v", lineNo, err)
				}
				if format > 0 {
					if format != f {
						return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: Multiple conflicting __MAPPER__ headers", lineNo)
					}
				}
				format = f

				if format < MINIMUM_SUPPORTED_MAP_FILE_FORMAT || format > MAXIMUM_SUPPORTED_MAP_FILE_FORMAT {
					return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: file format version %d is not supported", lineNo, format)
				}
			}
			continue
		}

		switch fields[0] {
		case "F":
			//
			// F <file>
			// Define file reference
			//
			f, err := tcllist.ConvertTypes(fields, "ss")
			if err != nil {
				return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: %v at \"%s\"", lineNo, err, incomingAttribute)
			}
			files = append(files, FileDefinition{
				File:        f[1].(string),
				IsLocalFile: !strings.HasPrefix(f[1].(string), "@"),
			})

		case "I":
			//
			// I <name> <zoom> <file>
			// Define image reference
			//
			f, err := tcllist.ConvertTypes(fields, "ssfs")
			if err != nil {
				return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: %v at \"%s\"", lineNo, err, incomingAttribute)
			}
			images[fmt.Sprintf("%s:%g", f[1].(string), f[2].(float64))] = ImageDefinition{
				Zoom:        f[2].(float64),
				Name:        f[1].(string),
				File:        f[3].(string),
				IsLocalFile: !strings.HasPrefix(f[3].(string), "@"),
			}

		case "P", "M":
			a := strings.SplitN(fields[1], ":", 2)
			if len(a) != 2 {
				return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: not a valid attr:id value: %s", lineNo, fields[1])
			}
			if _, ok := objects[a[1]]; !ok {
				objects[a[1]] = make(map[string][]string)
				objects[a[1]]["__mob_type__"] = []string{fields[0]}
			}
			fields = fields[1:]
			fallthrough

		default:
			if len(fields) < 2 {
				return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: not enough fields at \"%q\"", lineNo, fields)
			}
			a := strings.SplitN(fields[0], ":", 2)
			if len(a) != 2 {
				return nil, nil, nil, fmt.Errorf("Error parsing data stream at line %d: not a valid attr:id value: %s", lineNo, fields[0])
			}
			if _, ok := objects[a[1]]; !ok {
				objects[a[1]] = make(map[string][]string)
			}
			objects[a[1]][a[0]] = fields[1:]
		}
	}

	//
	// Now we have collected all of the files, images, and object raw data.
	// (It's necessary to do this as a first pass because objects are described
	// on multiple lines of the data stream and may not be in order. They
	// may even be interleaved. Now that we've sorted them out we can look
	// at each object individually.
	//
	oList := make([]MapObject, 0, len(objects))
	var o MapObject
	var err error

	for objId, objDef := range objects {
		mType, ok := objDef["__mob_type__"]
		if ok {
			switch mType[0] {
			case "M":
				o, err = newMonster(objId, objDef)
			case "P":
				o, err = newPlayer(objId, objDef)
			default:
				err = fmt.Errorf("unknown creature type (%s) for ID %s", mType, objId)
			}
		} else {
			oType, ok := objDef["TYPE"]
			if ok {
				switch oType[0] {
				case "aoe":
					o, err = newSpellAreaOfEffectElement(objId, objDef)
				case "arc":
					o, err = newArcElement(objId, objDef)
				case "circ":
					o, err = newCircleElement(objId, objDef)
				case "line":
					o, err = newLineElement(objId, objDef)
				case "poly":
					o, err = newPolygonElement(objId, objDef)
				case "rect":
					o, err = newRectangleElement(objId, objDef)
				case "text":
					o, err = newTextElement(objId, objDef)
				case "tile":
					o, err = newTileElement(objId, objDef)
				case "player":
					o, err = newPlayer(objId, objDef)
				case "monster":
					o, err = newMonster(objId, objDef)
				default:
					err = fmt.Errorf("unknown element type (%s) for ID %s", oType, objId)
				}
			} else {
				err = fmt.Errorf("element ID %s missing TYPE attribute", objId)
			}
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error parsing data stream: %v", err)
		}
		oList = append(oList, o)
	}
	return oList, images, files, nil
}

//
// HEALTH {}
// HEALTH {max lethal sub con flat? stable? condition [blur]}
//         int int    int int bool  bool    str       {}/int
func newHealth(objDef map[string][]string, err error) (*CreatureHealth, error) {
	if err != nil {
		return nil, err
	}
	h := CreatureHealth{}
	healthStats, ok := objDef["HEALTH"]
	if !ok || len(strings.TrimSpace(healthStats[0])) == 0 {
		return nil, err
	}

	var hs []string
	var hstats []interface{}

	hs, err = tcllist.ParseTclList(healthStats[0])
	if err != nil {
		return nil, err
	}
	if len(hs) < 8 {
		hstats, err = tcllist.ConvertTypes(hs, "iiii??s")
		hstats = append(hstats, 0)
	} else {
		hstats, err = tcllist.ConvertTypes(hs, "iiii??sI")
	}
	if err != nil {
		return nil, err
	}

	h.MaxHP = hstats[0].(int)
	h.LethalDamage = hstats[1].(int)
	h.NonLethalDamage = hstats[2].(int)
	h.Con = hstats[3].(int)
	h.IsFlatFooted = hstats[4].(bool)
	h.IsStable = hstats[5].(bool)
	h.Condition = hstats[6].(string)
	h.HpBlur = hstats[7].(int)
	return &h, err
}

//
// AOE {}
// AOE {radius r color}
//           float str
//
func newAoEShape(objDef map[string][]string, err error) (*RadiusAoE, error) {
	if err != nil {
		return nil, err
	}

	aoe, ok := objDef["AOE"]
	if !ok || len(strings.TrimSpace(aoe[0])) == 0 {
		return nil, err
	}

	var as []string
	var al []interface{}

	as, err = tcllist.ParseTclList(aoe[0])
	if err != nil || len(as) == 0 {
		return nil, err
	}

	switch as[0] {
	case "radius":
		al, err = tcllist.ConvertTypes(as, "sfs")
		if err != nil {
			return nil, err
		}
		return &RadiusAoE{
			Radius: al[1].(float64),
			Color:  al[2].(string),
		}, err
	}
	return nil, fmt.Errorf("undefined area of effect shape: %s", as[0])
}

func objFloat(objDef map[string][]string, i int, fldName string, required bool, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return 0, fmt.Errorf("attribute %s required", fldName)
		}
		return 0, err
	}
	if len(val) <= i {
		return 0, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return strconv.ParseFloat(val[i], 64)
}

func objInt(objDef map[string][]string, i int, fldName string, required bool, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return 0, fmt.Errorf("attribute %s required", fldName)
		}
		return 0, err
	}
	if len(val) <= i {
		return 0, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return strconv.Atoi(val[i])
}

func objBool(objDef map[string][]string, i int, fldName string, required bool, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return false, fmt.Errorf("attribute %s required", fldName)
		}
		return false, err
	}
	if len(val) <= i {
		return false, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return strconv.ParseBool(val[i])
}

func objString(objDef map[string][]string, i int, fldName string, required bool, err error) (string, error) {
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return "", fmt.Errorf("attribute %s required", fldName)
		}
		return "", err
	}
	if len(val) <= i {
		return "", fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return val[i], err
}

const (
	MoveModeLand = iota
	MoveModeBurrow
	MoveModeClimb
	MoveModeFly
	MoveModeSwim
)

type enumChoices map[string]byte

func objEnum(objDef map[string][]string, i int, fldName string, required bool, choices enumChoices, err error) (byte, error) {
	if err != nil {
		return 0, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return 0, fmt.Errorf("attribute %s required", fldName)
		}
		return 0, err
	}
	if len(val) <= i {
		return 0, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return strEnum(val[i], required, choices, err)
}

func strEnum(val string, required bool, choices enumChoices, err error) (byte, error) {
	if val == "" && !required {
		return 0, nil
	}
	choice, ok := choices[val]
	if !ok {
		return 0, fmt.Errorf("value %s not in allowed set", val)
	}
	return choice, err
}

func objStrings(objDef map[string][]string, i int, fldName string, required bool, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return nil, fmt.Errorf("attribute %s required", fldName)
		}
		return nil, err
	}
	if len(val) <= i {
		return nil, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	list, err := tcllist.ParseTclList(val[i])
	if err != nil {
		return nil, fmt.Errorf("attribute %s value %s could not be parsed: %v", fldName, val[i], err)
	}
	return list, err
}

func newCreature(objId string, objDef map[string][]string) (CreatureToken, error) {
	var err error
	c := CreatureToken{}
	c.ID = objId
	c.Name, err = objString(objDef, 0, "NAME", true, err)
	c.Gx, err = objFloat(objDef, 0, "GX", true, err)
	c.Gy, err = objFloat(objDef, 0, "GY", true, err)
	c.Health, err = newHealth(objDef, err)
	c.Elev, err = objInt(objDef, 0, "ELEV", false, err)
	c.MoveMode, err = objEnum(objDef, 0, "MOVEMODE", false, enumChoices{
		"fly":    MoveModeFly,
		"climb":  MoveModeClimb,
		"swim":   MoveModeSwim,
		"burrow": MoveModeBurrow,
		"land":   MoveModeLand,
	}, err)
	c.Color, err = objString(objDef, 0, "COLOR", false, err)
	c.Note, err = objString(objDef, 0, "NOTE", false, err)
	c.Skin, err = objInt(objDef, 0, "SKIN", false, err)
	c.SkinSize, err = objStrings(objDef, 0, "SKINSIZE", false, err)
	c.Size, err = objString(objDef, 0, "SIZE", true, err)
	c.StatusList, err = objStrings(objDef, 0, "STATUSLIST", false, err)
	c.AoE, err = newAoEShape(objDef, err)
	c.Area, err = objString(objDef, 0, "AREA", true, err)
	c.Reach, err = objBool(objDef, 0, "REACH", false, err)
	c.Killed, err = objBool(objDef, 0, "KILLED", false, err)
	c.Dim, err = objBool(objDef, 0, "DIM", false, err)

	return c, err
}

func newPlayer(objId string, objDef map[string][]string) (PlayerToken, error) {
	c, err := newCreature(objId, objDef)
	return PlayerToken{
		CreatureToken: c,
	}, err
}

func newMonster(objId string, objDef map[string][]string) (MonsterToken, error) {
	c, err := newCreature(objId, objDef)
	return MonsterToken{
		CreatureToken: c,
	}, err
}

func objCoordinateList(objDef map[string][]string, i int, fldName string, required bool, err error) ([]Coordinates, error) {
	if err != nil {
		return nil, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return nil, fmt.Errorf("attribute %s required", fldName)
		}
		return nil, err
	}
	if len(val) <= i {
		return nil, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	list, err := tcllist.ParseTclList(val[i])
	if err != nil {
		return nil, err
	}
	if (len(list) % 2) != 0 {
		return nil, fmt.Errorf("attribute %s list must have an even number of elements", fldName)
	}
	cl := make([]Coordinates, 0, 2)
	for i := 0; i < len(list); i += 2 {
		xf, err := strconv.ParseFloat(list[i], 64)
		yf, err2 := strconv.ParseFloat(list[i+1], 64)
		if err != nil || err2 != nil {
			return nil, fmt.Errorf("values in %s list must be valid floating-point values", fldName)
		}
		cl = append(cl, Coordinates{
			X: xf,
			Y: yf,
		})
	}
	return cl, nil
}

func objMapElement(objId string, objDef map[string][]string) (MapElement, error) {
	var err error

	e := MapElement{
		BaseMapObject: BaseMapObject{
			ID: objId,
		},
	}
	e.X, err = objFloat(objDef, 0, "X", true, err)
	e.Y, err = objFloat(objDef, 0, "Y", true, err)
	e.Z, err = objInt(objDef, 0, "Z", true, err)
	e.Level, err = objInt(objDef, 0, "LEVEL", false, err)
	e.Group, err = objString(objDef, 0, "GROUP", false, err)
	e.Points, err = objCoordinateList(objDef, 0, "POINTS", false, err)
	e.Fill, err = objString(objDef, 0, "FILL", false, err)
	e.Dash, err = objEnum(objDef, 0, "DASH", false, enumChoices{
		"-":   DashLong,
		",":   DashMedium,
		".":   DashShort,
		"-.":  DashLongShort,
		"-..": DashLong2Short,
	}, err)
	e.Line, err = objString(objDef, 0, "LINE", false, err)
	e.Width, err = objInt(objDef, 0, "WIDTH", false, err)
	e.Layer, err = objString(objDef, 0, "LAYER", false, err)
	e.Hidden, err = objBool(objDef, 0, "HIDDEN", false, err)
	e.Locked, err = objBool(objDef, 0, "LOCKED", false, err)

	return e, err
}

const (
	DashSolid = iota
	DashLong
	DashMedium
	DashShort
	DashLongShort
	DashLong2Short
)

const (
	AoEShapeCone = iota
	AoEShapeRadius
	AoEShapeRay
)

func newSpellAreaOfEffectElement(objId string, objDef map[string][]string) (SpellAreaOfEffectElement, error) {
	me, err := objMapElement(objId, objDef)
	sa := SpellAreaOfEffectElement{
		MapElement: me,
	}
	sa.AoEShape, err = objEnum(objDef, 0, "AOESHAPE", true, enumChoices{
		"cone":   AoEShapeCone,
		"radius": AoEShapeRadius,
		"ray":    AoEShapeRay,
	}, err)
	return sa, err
}

func newArcElement(objId string, objDef map[string][]string) (ArcElement, error) {
	me, err := objMapElement(objId, objDef)
	arc := ArcElement{
		MapElement: me,
	}
	arc.Arcmode, err = objEnum(objDef, 0, "ARCMODE", true, enumChoices{
		"pieslice": ArcModePieSlice,
		"arc":      ArcModeArc,
		"chord":    ArcModeChord,
	}, err)
	arc.Start, err = objFloat(objDef, 0, "START", true, err)
	arc.Extent, err = objFloat(objDef, 0, "EXTENT", true, err)
	return arc, err
}

func newCircleElement(objId string, objDef map[string][]string) (CircleElement, error) {
	me, err := objMapElement(objId, objDef)
	return CircleElement{
		MapElement: me,
	}, err
}

const (
	ArrowNone = iota
	ArrowFirst
	ArrowLast
	ArrowBoth
)

func newLineElement(objId string, objDef map[string][]string) (LineElement, error) {
	me, err := objMapElement(objId, objDef)
	line := LineElement{
		MapElement: me,
	}
	line.Arrow, err = objEnum(objDef, 0, "ARROW", false, enumChoices{
		"none":  ArrowNone,
		"first": ArrowFirst,
		"last":  ArrowLast,
		"both":  ArrowBoth,
	}, err)
	return line, err
}

const (
	JoinBevel = iota
	JoinMiter
	JoinRound
)

func newPolygonElement(objId string, objDef map[string][]string) (PolygonElement, error) {
	me, err := objMapElement(objId, objDef)
	poly := PolygonElement{
		MapElement: me,
	}
	poly.Join, err = objEnum(objDef, 0, "JOIN", false, enumChoices{
		"bevel": JoinBevel,
		"miter": JoinMiter,
		"round": JoinRound,
	}, err)
	poly.Spline, err = objFloat(objDef, 0, "SPLINE", false, err)
	return poly, err
}

func newRectangleElement(objId string, objDef map[string][]string) (RectangleElement, error) {
	me, err := objMapElement(objId, objDef)
	return RectangleElement{
		MapElement: me,
	}, err
}

func objTextFont(objDef map[string][]string, i int, fldName string, required bool, err error) (TextFont, error) {
	if err != nil {
		return TextFont{}, err
	}
	val, ok := objDef[fldName]
	if !ok {
		if required {
			return TextFont{}, fmt.Errorf("attribute %s required", fldName)
		}
		return TextFont{}, err
	}
	if len(val) <= i {
		return TextFont{}, fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	f, err := tcllist.ParseTclList(val[i])
	if err != nil {
		return TextFont{}, err
	}
	// oddly this is stored with an extra level of list-wrapping
	if len(f) != 1 {
		return TextFont{}, fmt.Errorf("attribute %s has invalid font format [%s]", fldName, val[i])
	}
	f, err = tcllist.ParseTclList(f[0])
	if err != nil {
		return TextFont{}, err
	}

	var ff []interface{}
	switch len(f) {
	case 2:
		ff, err = tcllist.ConvertTypes(f, "sf")
		ff = append(ff, "normal", "roman")
	case 3:
		ff, err = tcllist.ConvertTypes(f, "sfs")
		ff = append(ff, "roman")
	case 4:
		ff, err = tcllist.ConvertTypes(f, "sfss")
	default:
		return TextFont{}, fmt.Errorf("invalid font specification %s", val)
	}
	if err != nil {
		return TextFont{}, err
	}

	w, err := strEnum(ff[2].(string), true, enumChoices{
		"normal": FontWeightNormal,
		"bold":   FontWeightBold,
	}, err)
	s, err := strEnum(ff[3].(string), true, enumChoices{
		"roman":  FontSlantRoman,
		"italic": FontSlantItalic,
	}, err)

	return TextFont{
		Family: ff[0].(string),
		Size:   ff[1].(float64),
		Weight: w,
		Slant:  s,
	}, err
}

const (
	AnchorCenter = iota
	AnchorNorth
	AnchorSouth
	AnchorEast
	AnchorWest
	AnchorNE
	AnchorNW
	AnchorSW
	AnchorSE
)

func newTextElement(objId string, objDef map[string][]string) (TextElement, error) {
	me, err := objMapElement(objId, objDef)
	text := TextElement{
		MapElement: me,
	}

	text.Text, err = objString(objDef, 0, "TEXT", true, err)
	text.Font, err = objTextFont(objDef, 0, "FONT", true, err)
	text.Anchor, err = objEnum(objDef, 0, "ANCHOR", false, enumChoices{
		"center": AnchorCenter,
		"n":      AnchorNorth,
		"s":      AnchorSouth,
		"e":      AnchorEast,
		"w":      AnchorWest,
		"ne":     AnchorNE,
		"se":     AnchorSE,
		"nw":     AnchorNW,
		"sw":     AnchorSW,
	}, err)
	return text, err
}

func newTileElement(objId string, objDef map[string][]string) (TileElement, error) {
	me, err := objMapElement(objId, objDef)
	tile := TileElement{
		MapElement: me,
	}
	tile.Image, err = objString(objDef, 0, "IMAGE", true, err)
	return tile, err
}

type saveAttributes struct {
	Tag      string
	Type     string
	Required bool
	Value    interface{}
}

func saveCreatureAoE(aoe *RadiusAoE) (string, error) {
	if aoe == nil {
		return "", nil
	}
	return tcllist.ToTclString([]string{
		"radius",
		fmt.Sprintf("%g", aoe.Radius),
		aoe.Color,
	})
}

func saveBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func saveHealth(h *CreatureHealth) (string, error) {
	if h == nil {
		return "", nil
	}
	data := make([]string, 0, 8)
	data = append(data, fmt.Sprintf("%d", h.MaxHP))
	data = append(data, fmt.Sprintf("%d", h.LethalDamage))
	data = append(data, fmt.Sprintf("%d", h.NonLethalDamage))
	data = append(data, fmt.Sprintf("%d", h.Con))
	data = append(data, saveBool(h.IsFlatFooted))
	data = append(data, saveBool(h.IsStable))
	data = append(data, h.Condition)
	data = append(data, fmt.Sprintf("%d", h.HpBlur))
	return tcllist.ToTclString(data)
}

func saveEnum(choice byte, choices enumChoices) (string, error) {
	for k, v := range choices {
		if v == choice {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v not in list of valid enum choices", choice)
}

func SaveObjects(objects []MapObject, images map[string]ImageDefinition, files []FileDefinition) ([]string, error) {
	data := make([]string, 0, 32)
	for _, o := range objects {
		switch obj := o.(type) {
		case ArcElement:
		case CircleElement:
		case LineElement:
		case MonsterToken:
		case PlayerToken:
			ss, err := tcllist.ToTclString(obj.SkinSize)
			if err != nil {
				return nil, err
			}
			sl, err := tcllist.ToTclString(obj.StatusList)
			if err != nil {
				return nil, err
			}
			ae, err := saveCreatureAoE(obj.AoE)
			if err != nil {
				return nil, err
			}
			mm, err := saveEnum(obj.MoveMode, enumChoices{
				"land":   MoveModeLand,
				"burrow": MoveModeBurrow,
				"climb":  MoveModeClimb,
				"fly":    MoveModeFly,
				"swim":   MoveModeSwim,
			})
			if err != nil {
				return nil, err
			}
			he, err := saveHealth(obj.Health)
			if err != nil {
				return nil, err
			}
			data, err = saveValues(data, "P", o.ObjID(), []saveAttributes{
				{"TYPE", "s", true, "player"},
				{"NAME", "s", true, obj.Name},
				{"GX", "f", true, obj.Gx},
				{"GY", "f", true, obj.Gy},
				{"SKIN", "i", true, obj.Skin},
				{"SKINSIZE", "s", false, ss},
				{"ELEV", "i", true, obj.Elev},
				{"COLOR", "s", true, obj.Color},
				{"NOTE", "s", false, obj.Note},
				{"SIZE", "s", true, obj.Size},
				{"STATUSLIST", "s", false, sl},
				{"AOE", "s", false, ae},
				{"AREA", "s", true, obj.Area},
				{"MOVEMODE", "s", false, mm},
				{"REACH", "b", true, obj.Reach},
				{"KILLED", "b", true, obj.Killed},
				{"DIM", "b", true, obj.Dim},
				{"HEALTH", "s", false, he},
			})
			if err != nil {
				return nil, err
			}

		case PolygonElement:
		case RectangleElement:
		case SpellAreaOfEffectElement:
		case TextElement:
		case TileElement:
		default:
			return nil, fmt.Errorf("unexpected map object type for %s", o.ObjID())
		}
	}
	return data, nil
}

func saveValues(previous []string, prefix, objID string, attrs []saveAttributes) ([]string, error) {
	for _, attr := range attrs {
		data := make([]string, 0, 8)
		if prefix != "" {
			data = append(data, prefix)
		}
		data = append(data, attr.Tag+":"+objID)
		var s string
		switch attr.Type {
		case "b":
			s = saveBool(attr.Value.(bool))
		case "f":
			s = fmt.Sprintf("%g", attr.Value.(float64))
		case "i":
			s = fmt.Sprintf("%d", attr.Value.(int))
		case "s":
			s = attr.Value.(string)
		default:
			return nil, fmt.Errorf("unrecognized saveValues type %s for %s", attr.Type, attr.Tag)
		}

		if s != "" || attr.Required {
			data = append(data, s)
			record, err := tcllist.ToTclString(data)
			if err != nil {
				return nil, err
			}
			previous = append(previous, record)
		}
	}
	return previous, nil
}

// IMAGE SIZE

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
