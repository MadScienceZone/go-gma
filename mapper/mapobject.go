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
// EXPERIMENTAL CODE
//
// THIS PACKAGE IS STILL A WORK IN PROGRESS and has not been
// completely tested yet. Although GMA generally is a stable
// product, this module of it is new, and is not.
//

//
// MapObject describes the elements that may appear on the map.
//

package mapper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v4/tcllist"
)

//
// GMAMapperFileFormat gives
// the GMA File Format version number current as of this build.
// This is the format which will be used for saving map data.
//
const GMAMapperFileFormat = 17 // @@##@@ auto-configured
//
// MinimumSupportedMapFileFormat gives the lowest file format this package can
// understand.
//
const MinimumSupportedMapFileFormat = 14

//
// MaximumSupportedMapFileFormat gives the highest file format this package
// can understand. Saved data will be in this format.
//
const MaximumSupportedMapFileFormat = 17

func init() {
	if MinimumSupportedMapFileFormat > GMAMapperFileFormat || MaximumSupportedMapFileFormat < GMAMapperFileFormat {
		if MinimumSupportedMapFileFormat == MaximumSupportedMapFileFormat {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapper only supports file format %v, but version %v was the official one when this package was released!", MinimumSupportedMapFileFormat, GMAMapperFileFormat))
		} else {
			panic(fmt.Sprintf("BUILD ERROR: This version of mapper only supports mapper file formats %v-%v, but version %v was the official one when this package was released!", MinimumSupportedMapFileFormat, MaximumSupportedMapFileFormat, GMAMapperFileFormat))
		}
	}
}

//  ___ _  _ _   _ __  __ ___
// | __| \| | | | |  \/  / __|
// | _|| .` | |_| | |\/| \__ \
// |___|_|\_|\___/|_|  |_|___/
//
// The following definitions provide a mapping between the
// text expression of the enum values as used in our save
// file format and online protocol messages, and the internal
// numeric codes used here.
//

//
// These are the allowed values for the Dash attribute of a MapElement.
//
type DashType byte

const (
	DashSolid DashType = iota
	DashLong
	DashMedium
	DashShort
	DashLongShort
	DashLong2Short
)

var enumDashes = enumChoices{
	"":    byte(DashSolid),
	"-":   byte(DashLong),
	",":   byte(DashMedium),
	".":   byte(DashShort),
	"-.":  byte(DashLongShort),
	"-..": byte(DashLong2Short),
}

//
// These are the allowed values for the ArcMode attribute of an ArcElement.
//
type ArcModeType byte

const (
	ArcModePieSlice ArcModeType = iota
	ArcModeArc
	ArcModeChord
)

var enumArcs = enumChoices{
	"pieslice": byte(ArcModePieSlice),
	"arc":      byte(ArcModeArc),
	"chord":    byte(ArcModeChord),
}

//
// Valid values for a line's Arrow attribute.
//
type ArrowType byte

const (
	ArrowNone ArrowType = iota
	ArrowFirst
	ArrowLast
	ArrowBoth
)

var enumArrows = enumChoices{
	"none":  byte(ArrowNone),
	"first": byte(ArrowFirst),
	"last":  byte(ArrowLast),
	"both":  byte(ArrowBoth),
}

//
// These are the allowed values for the Join attribute of a PolygonElement.
//
type JoinStyle byte

const (
	JoinBevel JoinStyle = iota
	JoinMiter
	JoinRound
)

var enumJoins = enumChoices{
	"bevel": byte(JoinBevel),
	"miter": byte(JoinMiter),
	"round": byte(JoinRound),
}

//
// These are the valid values for the AoEShape attribute.
//
type AoEType byte

const (
	AoEShapeCone AoEType = iota
	AoEShapeRadius
	AoEShapeRay
)

var enumAoeShapes = enumChoices{
	"cone":   byte(AoEShapeCone),
	"radius": byte(AoEShapeRadius),
	"ray":    byte(AoEShapeRay),
}

//
// The valid font weights.
//
type FontWeightType byte

const (
	FontWeightNormal FontWeightType = iota
	FontWeightBold
)

var enumFontWeights = enumChoices{
	"normal": byte(FontWeightNormal),
	"bold":   byte(FontWeightBold),
}

//
// The valid font slants.
//
type FontSlantType byte

const (
	FontSlantRoman FontSlantType = iota
	FontSlantItalic
)

var enumFontSlants = enumChoices{
	"roman":  byte(FontSlantRoman),
	"italic": byte(FontSlantItalic),
}

//
// The valid values for the Anchor attribute of a TextElement.
//
type AnchorDirection byte

const (
	AnchorCenter AnchorDirection = iota
	AnchorNorth
	AnchorSouth
	AnchorEast
	AnchorWest
	AnchorNE
	AnchorNW
	AnchorSW
	AnchorSE
)

var enumAnchors = enumChoices{
	"center": byte(AnchorCenter),
	"n":      byte(AnchorNorth),
	"s":      byte(AnchorSouth),
	"e":      byte(AnchorEast),
	"w":      byte(AnchorWest),
	"ne":     byte(AnchorNE),
	"se":     byte(AnchorSE),
	"nw":     byte(AnchorNW),
	"sw":     byte(AnchorSW),
}

//
// The valid values for a creature's MoveMode attribute.
//
type MoveModeType byte

const (
	MoveModeLand MoveModeType = iota
	MoveModeBurrow
	MoveModeClimb
	MoveModeFly
	MoveModeSwim
)

var enumMoveModes = enumChoices{
	"fly":    byte(MoveModeFly),
	"climb":  byte(MoveModeClimb),
	"swim":   byte(MoveModeSwim),
	"burrow": byte(MoveModeBurrow),
	"land":   byte(MoveModeLand),
}

//
// This returns the underlying Go data type
// for attribute values as a string. If the boolean
// is false, then we don't know what the attribute is
// (so will treat it as a string)
//

func enumToByte(attrName, value string) (evalue byte, ok bool) {
	switch attrName {
	case "AOESHAPE":
		evalue, ok = enumAoeShapes[value]
	case "ANCHOR":
		evalue, ok = enumAnchors[value]
	case "ARCMODE":
		evalue, ok = enumArcs[value]
	case "ARROW":
		evalue, ok = enumArrows[value]
	case "DASH":
		evalue, ok = enumDashes[value]
	case "MOVEMODE":
		evalue, ok = enumMoveModes[value]
	case "JOIN":
		evalue, ok = enumJoins[value]
	default:
		evalue, ok = 0, false
	}
	return
}

//
// This returns a string describing the expected data type
// of a MapObject's attribute.
//

func attributeType(attrName string) (string, bool) {
	switch attrName {
	case "AOE":
		return "*RadiusAoE", true
	case "AOESHAPE", "ANCHOR", "ARCMODE", "ARROW", "DASH", "MOVEMODE", "JOIN":
		return "enum", true
	case "POINTS":
		return "[]Coordinates", true
	case "HEALTH":
		return "*CreatureHealth", true
	case "SKINSIZE", "STATUSLIST":
		return "[]string", true
	case "BBHEIGHT", "BBWIDTH", "EXTENT", "GX", "GY", "START", "X", "Y":
		return "float64", true
	case "DIM", "HIDDEN", "KILLED", "LOCKED", "REACH":
		return "bool", true
	case "ELEV", "LEVEL", "SKIN", "SPLINE", "WIDTH", "Z":
		return "int", true
	case "AREA", "COLOR", "FILL", "GROUP", "IMAGE", "LAYER", "LINE", "NAME", "NOTE", "SIZE", "TEXT", "TYPE":
		return "string", true
	case "FONT":
		return "TextFont", true
	}
	return "string", false
}

//________________________________________________________________________________
//  __  __              ___  _     _           _
// |  \/  | __ _ _ __  / _ \| |__ (_) ___  ___| |_
// | |\/| |/ _` | '_ \| | | | '_ \| |/ _ \/ __| __|
// | |  | | (_| | |_) | |_| | |_) | |  __/ (__| |_
// |_|  |_|\__,_| .__/ \___/|_.__// |\___|\___|\__|
//              |_|             |__/

//
// MapObject is anything the map server or client tracks and manages.
// These are generally things that are displayed on-screen such as map features,
// creature tokens, etc.
//
type MapObject interface {
	ObjID() string
	saveData([]string, string, string) ([]string, error)
}

//
// In the following sections, each MapObject type is described,
// including the type declaration, and the standard methods
//   obj<T>(objID, objDef) (<T>, error)
//      -- create a new <T> value with the given ID, based on data parsed
//         from an input source in objDef, which is a map of attribute name
//         to a slice of strings broken out into fields.
//
//   <T>.saveData(data, prefix, id) ([]string, error)
//      -- generate the save file data for this object. These are appended
//         as individual strings (one per line) to the end of the strings
//         in data. The new combined string slice is returned.
//

//
// Coordinates give an (x, y) coordinate pair to locate something on the map.
// Coordinates are in standard map pixel units (10 pixels = 1 foot).
//
type Coordinates struct {
	X, Y float64
}

//
// saveData converts a Coordinate pair to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData, but simply
// saves the X and Y fields for the element's reference point.
//
func (c Coordinates) saveData(data []string, prefix, id string) ([]string, error) {
	return saveValues(data, prefix, id, []saveAttributes{
		{"X", "f", true, c.X},
		{"Y", "f", true, c.Y},
	})
}

// Coordinates
//  saveData
//
// MapObject
//  BaseMapObject
//   MapElement
//     objMapElement
//     saveData
//    ArcElement
//      objArcElement
//      saveData
//    CircleElement
//      objCircleElement
//      saveData
//    LineElement
//      objLineElement
//      saveData
//    PolygonElement
//      objPolygonElement
//      saveData
//    RectangleElement
//      objRectangleElement
//      saveData
//    SpellAreaOfEffect
//      objSpellAreaOfEffect
//      saveData
//    (TextFont)
//      objTextFont
//    TextElement
//      objTextElement
//      saveData
//    TileElement
//      objTileElement
//      saveData
//   CreatureToken
//     objCreature
//     saveData
//    PlayerToken
//      objPlayer
//      saveData
//    MonsterToken
//      objMonster
//      saveData
//    (CreatureHealth)
//      objHealth
//      newHealth
//      saveHealth
//    (RadiusAoE)
//      objAoEShape
//      saveCreatureAoE
// ImageDefinition
// FileDefinition
//

//
// BaseMapObject holds attributes all MapObjects have in common, so they will import
// BaseMapObject into their definitions by composition.
//
type BaseMapObject struct {
	// Unique object identifier. May be any string
	// consisting of upper- or lower-case letters, digits, '_', and "#"
	// characters.
	//
	// By convention, we create these from a UUID expressed in
	// hex without punctuation. Local conventions may also be
	// used, such as PC character tokens using ID strings such as
	// "PC1", "PC2", etc.
	ID string
}

//
// saveData converts a MapObject to a text representation of that object
// in the map file format (suitable for sending to clients or saving to a disk
// file). Each type must have one of these methods to satisfy the MapObject
// interface.
//
// In the save file format (and client/server protocol) the data for this
// object occupies one or more lines of data. These lines are appended
// as a list of strings (one per line) to the input strings in the data parameter.
// The new list of strings (input data + this object's) is returned.
//
// If prefix is non-empty, it is prepended to each line as the first field of the
// lines saved for this object. This is specified in the file format for certain
// object types such as players, monsters, files, and images.
//
// The object's ID as recorded in the saved data list is given by the id parameter.
//
func (o BaseMapObject) saveData(data []string, prefix, id string) ([]string, error) {
	return data, nil
}

//
// ObjID returns the unique ID of a MapObject.
// Each type must have one of these methods to satisfy the MapObject
// interface.
//
func (o BaseMapObject) ObjID() string {
	return o.ID
}

//________________________________________________________________________________
//  __  __             _____ _                           _
// |  \/  | __ _ _ __ | ____| | ___ _ __ ___   ___ _ __ | |_
// | |\/| |/ _` | '_ \|  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | |  | | (_| | |_) | |___| |  __/ | | | | |  __/ | | | |_
// |_|  |_|\__,_| .__/|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//              |_|

//
// MapElement is a MapObject which represents a static map feature
// to be displayed.
//
// Each MapElement has at least one pair of (x, y) coordinates which
// locate the element's "reference point" on the map. What this means
// is up to each different kind of MapElement.
//
type MapElement struct {
	BaseMapObject
	Coordinates

	// Objects which need additional coordinate pairs to describe their
	// geometry (beyond the standard reference point) store them here.
	Points []Coordinates

	// The z "coordinate" is the vertical stacking order relative to the other
	// displayed on-screen objects.
	Z int

	// The colors used to draw the element's outline and/or to fill it's interior.
	// These may be standard color names such as "blue" or an RGB string such as
	// "#336699". A fill color that is the empty string means not to fill that element.
	Line string
	Fill string

	// The width in pixel units to draw the element's outline.
	Width int

	// The map layer this element belongs to.
	Layer string

	// The dungeon level where this element appears. Typically, level 0
	// is the default (ground) level, with level numbers increasing as
	// 1, 2, 3, etc., for floors above it, and with underground levels
	// counting down as -1, -2, -3, etc.
	Level int

	// Elements may be arranged into logical groups to be manipulated
	// together. This is the ID of the group to which this belongs, or
	// is empty if this element is not grouped.
	Group string

	// The element's line(s) are to be drawn with this dash pattern.
	Dash DashType

	// Is this element currently concealed from view?
	Hidden bool

	// Is the object locked from editing by the user?
	Locked bool
}

//
// objMapElement constructs a new MapElement from fields in objDef, generally as part of
// constructing something that is a more specific kind of object.
//
func objMapElement(objID string, objDef map[string][]string) (MapElement, error) {
	var err error

	e := MapElement{
		BaseMapObject: BaseMapObject{
			ID: objID,
		},
	}
	var b byte
	e.X, err = objFloat(objDef, 0, "X", true, err)
	e.Y, err = objFloat(objDef, 0, "Y", true, err)
	e.Z, err = objInt(objDef, 0, "Z", true, err)
	e.Level, err = objInt(objDef, 0, "LEVEL", false, err)
	e.Group, err = objString(objDef, 0, "GROUP", false, err)
	e.Points, err = objCoordinateList(objDef, 0, "POINTS", false, err)
	e.Fill, err = objString(objDef, 0, "FILL", false, err)
	b, err = objEnum(objDef, 0, "DASH", false, enumDashes, err)
	e.Dash = DashType(b)
	e.Line, err = objString(objDef, 0, "LINE", false, err)
	e.Width, err = objInt(objDef, 0, "WIDTH", false, err)
	e.Layer, err = objString(objDef, 0, "LAYER", false, err)
	e.Hidden, err = objBool(objDef, 0, "HIDDEN", false, err)
	e.Locked, err = objBool(objDef, 0, "LOCKED", false, err)

	return e, err
}

//
// saveData converts a MapElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o MapElement) saveData(data []string, prefix, id string) ([]string, error) {
	var err error
	if data, err = o.BaseMapObject.saveData(data, prefix, id); err != nil {
		return nil, err
	}
	if data, err = o.Coordinates.saveData(data, prefix, id); err != nil {
		return nil, err
	}

	var coords string
	if len(o.Points) > 0 {
		cl := make([]string, 0, len(o.Points))
		for _, c := range o.Points {
			cl = append(cl, fmt.Sprintf("%g", c.X))
			cl = append(cl, fmt.Sprintf("%g", c.Y))
		}
		if coords, err = tcllist.ToTclString(cl); err != nil {
			return nil, err
		}
	}

	da, err := saveEnum(byte(o.Dash), enumDashes)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"Z", "i", true, o.Z},
		{"POINTS", "s", true, coords},
		{"LOCKED", "b", false, o.Locked},
		{"FILL", "s", true, o.Fill},
		{"LINE", "s", false, o.Line},
		{"WIDTH", "i", false, o.Width},
		{"LAYER", "s", true, o.Layer},
		{"HIDDEN", "b", false, o.Hidden},
		{"LEVEL", "i", false, o.Level},
		{"GROUP", "s", false, o.Group},
		{"DASH", "s", false, da},
	})
}

//________________________________________________________________________________
//     _             _____ _                           _
//    / \   _ __ ___| ____| | ___ _ __ ___   ___ _ __ | |_
//   / _ \ | '__/ __|  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
//  / ___ \| | | (__| |___| |  __/ | | | | |  __/ | | | |_
// /_/   \_\_|  \___|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

// ArcElement is a MapElement that draws an arc on-screen.
// The arc is defined as a portion of a circle which is inscribed
// within the rectangle formed by the reference point and the single
// additional point in its Points attribute.
//
// Start and Extent specify the portion of that circle to include
// in the arc, measured in degrees.
//
// ArcMode defines how to draw the arc: it may be an arc (curve along
// the circle's permieter without connecting the endpoints), chord
// (the endpoints connected to each other with a straight line), or
// a pieslice (endpoints connected to the center of the circle with
// straight lines).
//
type ArcElement struct {
	MapElement
	ArcMode ArcModeType
	Extent  float64
	Start   float64
}

//
// objArcElement creates a new instance from the fields in objDef.
//
func objArcElement(objID string, objDef map[string][]string) (ArcElement, error) {
	me, err := objMapElement(objID, objDef)
	arc := ArcElement{
		MapElement: me,
	}
	var b byte
	b, err = objEnum(objDef, 0, "ARCMODE", true, enumArcs, err)
	arc.ArcMode = ArcModeType(b)
	arc.Start, err = objFloat(objDef, 0, "START", true, err)
	arc.Extent, err = objFloat(objDef, 0, "EXTENT", true, err)
	return arc, err
}

//
// saveData converts an ArcElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o ArcElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	am, err := saveEnum(byte(o.ArcMode), enumArcs)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "arc"},
		{"ARCMODE", "s", true, am},
		{"START", "f", true, o.Start},
		{"EXTENT", "f", true, o.Extent},
	})
}

//________________________________________________________________________________
//   ____ _          _      _____ _                           _
//  / ___(_)_ __ ___| | ___| ____| | ___ _ __ ___   ___ _ __ | |_
// | |   | | '__/ __| |/ _ \  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | |___| | | | (__| |  __/ |___| |  __/ | | | | |  __/ | | | |_
//  \____|_|_|  \___|_|\___|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

// CircleElement is a MapElement that draws an ellipse or circle on-screen.
// The ellipse is described by the rectangle formed by the reference point
// and the single point in the Points attribute (as diagonally opposing points),
// with the circle/ellipse being inscribed in that rectangle.
//
type CircleElement struct {
	MapElement
}

//
// objCircleElement creates a new instance from the fields in objDef.
//
func objCircleElement(objID string, objDef map[string][]string) (CircleElement, error) {
	me, err := objMapElement(objID, objDef)
	return CircleElement{
		MapElement: me,
	}, err
}

//
// saveData converts a CircleElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o CircleElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}
	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "circ"},
	})
}

//________________________________________________________________________________
//  _     _            _____ _                           _
// | |   (_)_ __   ___| ____| | ___ _ __ ___   ___ _ __ | |_
// | |   | | '_ \ / _ \  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | |___| | | | |  __/ |___| |  __/ | | | | |  __/ | | | |_
// |_____|_|_| |_|\___|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

// LineElement is a MapElement that draws a straight line segment from the
// reference point to the single point in the Points attribute.
//
// If there are multiple points in the Points attribute, the element will
// be drawn from each point to the next, as connected line segments.
//
// The line will have arrowheads drawn on the first (reference) point, the last
// point, both, or neither, as indicated by the Arrow attribute.
//
// N.B.: the lines will be drawn with the Fill color, not the Line color,
// to match the behavior of the Tk library underlying our client
// implementations.
//
type LineElement struct {
	MapElement

	// What arrowheads, if any, to draw on the endpoints
	Arrow ArrowType
}

//
// objLineElement creates a new instance from the fields in objDef.
//
func objLineElement(objID string, objDef map[string][]string) (LineElement, error) {
	me, err := objMapElement(objID, objDef)
	line := LineElement{
		MapElement: me,
	}
	var b byte
	b, err = objEnum(objDef, 0, "ARROW", false, enumArrows, err)
	line.Arrow = ArrowType(b)
	return line, err
}

//
// saveData converts a LineElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o LineElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	am, err := saveEnum(byte(o.Arrow), enumArrows)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "line"},
		{"ARROW", "s", false, am},
	})
}

//________________________________________________________________________________
//  ____       _                         _____ _                           _
// |  _ \ ___ | |_   _  __ _  ___  _ __ | ____| | ___ _ __ ___   ___ _ __ | |_
// | |_) / _ \| | | | |/ _` |/ _ \| '_ \|  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// |  __/ (_) | | |_| | (_| | (_) | | | | |___| |  __/ | | | | |  __/ | | | |_
// |_|   \___/|_|\__, |\__, |\___/|_| |_|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//               |___/ |___/

// PolygonElement is a MapElement that draws an arbitrary polygon, just as with
// the LineElement, but the interior of the shape described by the line segments
// may be filled in as a solid shape.
//
type PolygonElement struct {
	MapElement

	// Spline gives the factor to use when smoothing the sides of the polygon between
	// its points. 0 means not to smooth them at all, resulting in a shape with straight
	// edges between the vertices. Otherwise, larger values provide greater smoothing.
	Spline float64

	// The join style to control how the intersection between line segments is drawn.
	Join JoinStyle
}

//
// objPolygonElement creates a new instance from the fields in objDef.
//
func objPolygonElement(objID string, objDef map[string][]string) (PolygonElement, error) {
	me, err := objMapElement(objID, objDef)
	poly := PolygonElement{
		MapElement: me,
	}
	var b byte
	b, err = objEnum(objDef, 0, "JOIN", false, enumJoins, err)
	poly.Join = JoinStyle(b)
	poly.Spline, err = objFloat(objDef, 0, "SPLINE", false, err)
	return poly, err
}

//
// saveData converts a PolygonElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o PolygonElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	jm, err := saveEnum(byte(o.Join), enumJoins)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "poly"},
		{"JOIN", "s", true, jm},
		{"SPLINE", "f", true, o.Spline},
	})
}

//________________________________________________________________________________
//  ____           _                    _
// |  _ \ ___  ___| |_ __ _ _ __   __ _| | ___
// | |_) / _ \/ __| __/ _` | '_ \ / _` | |/ _ \
// |  _ <  __/ (__| || (_| | | | | (_| | |  __/
// |_| \_\___|\___|\__\__,_|_| |_|\__, |_|\___|
//                                |___/
//  _____ _                           _
// | ____| | ___ _ __ ___   ___ _ __ | |_
// |  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | |___| |  __/ | | | | |  __/ | | | |_
// |_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

// RectangleElement is a MapElement which describes a rectangle as defined by
// diagonally opposing points: the reference point and the single coordinate pair
// in the Points attribute.
//
type RectangleElement struct {
	MapElement
}

//
// objRectangleElement creates a new instance from the fields in objDef.
//
func objRectangleElement(objID string, objDef map[string][]string) (RectangleElement, error) {
	me, err := objMapElement(objID, objDef)
	return RectangleElement{
		MapElement: me,
	}, err
}

//
// saveData converts a RectangleElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o RectangleElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}
	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "rect"},
	})
}

//________________________________________________________________________________
//  ____             _ _    _                    ___   __
// / ___| _ __   ___| | |  / \   _ __ ___  __ _ / _ \ / _|
// \___ \| '_ \ / _ \ | | / _ \ | '__/ _ \/ _` | | | | |_
//  ___) | |_) |  __/ | |/ ___ \| | |  __/ (_| | |_| |  _|
// |____/| .__/ \___|_|_/_/   \_\_|  \___|\__,_|\___/|_|
//       |_|
//  _____  __  __           _   _____ _                           _
// | ____|/ _|/ _| ___  ___| |_| ____| | ___ _ __ ___   ___ _ __ | |_
// |  _| | |_| |_ / _ \/ __| __|  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
// | |___|  _|  _|  __/ (__| |_| |___| |  __/ | | | | |  __/ | | | |_
// |_____|_| |_|  \___|\___|\__|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

// SpellAreaOfEffectElement is a MapElement that shows a region on the map
// affected by a spell or other area effect.
//
// The region has one of the following shapes as indicated by the AoEShape
// attribute:
//   cone    A 90° pieslice described as with ArcElement
//   radius  An ellipse described as with CircleElement
//   ray     A rectangle described as with RectangleElement
//
type SpellAreaOfEffectElement struct {
	MapElement

	// The shape of the affected region of the map.
	AoEShape AoEType
}

//
// objSpellAreaOfEffectElement creates a new instance from the fields in objDef.
//
func objSpellAreaOfEffectElement(objID string, objDef map[string][]string) (SpellAreaOfEffectElement, error) {
	me, err := objMapElement(objID, objDef)
	sa := SpellAreaOfEffectElement{
		MapElement: me,
	}
	var b byte
	b, err = objEnum(objDef, 0, "AOESHAPE", true, enumAoeShapes, err)
	sa.AoEShape = AoEType(b)
	return sa, err
}

//
// saveData converts a SpellAreaOfEffectElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o SpellAreaOfEffectElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	ae, err := saveEnum(byte(o.AoEShape), enumAoeShapes)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "aoe"},
		{"AOESHAPE", "s", true, ae},
	})
}

//________________________________________________________________________________
//  _____         _   _____ _                           _
// |_   _|____  _| |_| ____| | ___ _ __ ___   ___ _ __ | |_
//   | |/ _ \ \/ / __|  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
//   | |  __/>  <| |_| |___| |  __/ | | | | |  __/ | | | |_
//   |_|\___/_/\_\\__|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

//
// TextFont describes a font used by TextElements.
//
type TextFont struct {
	// The name of the font family as recognized by Tk.
	Family string

	// The font size as recognized by Tk.
	Size float64

	// The font weight (normal or bold).
	Weight FontWeightType

	// The font slant (roman or italic).
	Slant FontSlantType
}

//
// TextElement is a MapElement which displays text on the map.
//
// The reference point is at the center of the text if Anchor is
// AnchorCenter, or is at the top-left corner of the text if Anchor
// is AnchorNW, and so on.
//
type TextElement struct {
	MapElement

	// The text to be displayed.
	Text string

	// Font to use for the text.
	Font TextFont

	// Where is the reference point in relation to the text?
	Anchor AnchorDirection
}

//
// objTextFont creates a new TextFont instance from the fields in objDef.
//
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

	w, err := strEnum(ff[2].(string), true, enumFontWeights, err)
	s, err := strEnum(ff[3].(string), true, enumFontSlants, err)

	return TextFont{
		Family: ff[0].(string),
		Size:   ff[1].(float64),
		Weight: FontWeightType(w),
		Slant:  FontSlantType(s),
	}, err
}

//
// objTextElement creates a new instance from the fields in objDef.
//
func objTextElement(objID string, objDef map[string][]string) (TextElement, error) {
	me, err := objMapElement(objID, objDef)
	text := TextElement{
		MapElement: me,
	}

	text.Text, err = objString(objDef, 0, "TEXT", true, err)
	text.Font, err = objTextFont(objDef, 0, "FONT", true, err)
	var b byte
	b, err = objEnum(objDef, 0, "ANCHOR", false, enumAnchors, err)
	text.Anchor = AnchorDirection(b)
	return text, err
}

//
// saveData converts a TextElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o TextElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	a, err := saveEnum(byte(o.Anchor), enumAnchors)
	if err != nil {
		return nil, err
	}
	fw, err := saveEnum(byte(o.Font.Weight), enumFontWeights)
	if err != nil {
		return nil, err
	}
	fs, err := saveEnum(byte(o.Font.Slant), enumFontSlants)
	if err != nil {
		return nil, err
	}

	f1, err := tcllist.ToTclString([]string{
		o.Font.Family,
		fmt.Sprintf("%g", o.Font.Size),
		fw, fs})
	if err != nil {
		return nil, err
	}
	f2, err := tcllist.ToTclString([]string{f1})
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "text"},
		{"TEXT", "s", true, o.Text},
		{"FONT", "s", true, f2},
		{"ANCHOR", "s", true, a},
	})
}

//________________________________________________________________________________
//  _____ _ _      _____ _                           _
// |_   _(_) | ___| ____| | ___ _ __ ___   ___ _ __ | |_
//   | | | | |/ _ \  _| | |/ _ \ '_ ` _ \ / _ \ '_ \| __|
//   | | | | |  __/ |___| |  __/ | | | | |  __/ | | | |_
//   |_| |_|_|\___|_____|_|\___|_| |_| |_|\___|_| |_|\__|
//

//
// TileElement is a MapElement which displays a bitmap image on the map.
// The upper-left corner of the image will be drawn at the reference point.
//
type TileElement struct {
	MapElement

	// Image name as known to the mapper system.
	Image string

	// Bounding box in pixels for the image tile.
	// If for some reason the tile can't be found, clients
	// can use the bounding box to indicate where the tile should be.
	// If the bounding box is not known, these values may both
	// be zero.
	BBHeight, BBWidth float64
}

//
// objTileElement creates a new instance from the fields in objDef.
//
func objTileElement(objID string, objDef map[string][]string) (TileElement, error) {
	me, err := objMapElement(objID, objDef)
	tile := TileElement{
		MapElement: me,
	}
	tile.Image, err = objString(objDef, 0, "IMAGE", true, err)
	tile.BBHeight, err = objFloat(objDef, 0, "BBHEIGHT", false, err)
	tile.BBWidth, err = objFloat(objDef, 0, "BBWIDTH", false, err)
	return tile, err
}

//
// saveData converts a TileElement to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o TileElement) saveData(data []string, prefix, id string) ([]string, error) {
	data, err := o.MapElement.saveData(data, prefix, id)
	if err != nil {
		return nil, err
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, "tile"},
		{"IMAGE", "s", true, o.Image},
		{"BBHEIGHT", "f", true, o.BBHeight},
		{"BBWIDTH", "f", true, o.BBWidth},
	})
}

//________________________________________________________________________________
//   ____                _                 _____     _
//  / ___|_ __ ___  __ _| |_ _   _ _ __ __|_   _|__ | | _____ _ __
// | |   | '__/ _ \/ _` | __| | | | '__/ _ \| |/ _ \| |/ / _ \ '_ \
// | |___| | |  __/ (_| | |_| |_| | | |  __/| | (_) |   <  __/ | | |
//  \____|_|  \___|\__,_|\__|\__,_|_|  \___||_|\___/|_|\_\___|_| |_|
//

//
// Creature type codes for the CreatureType field of CreatureToken
// (and PlayerToken and MonsterToken) values.
//
type CreatureTypeCode byte

const (
	CreatureTypeUnknown CreatureTypeCode = iota
	CreatureTypeMonster
	CreatureTypePlayer
)

//
// CreatureToken is a MapObject (but not a MapElement) which displays a movable
// token indicating the size and location of a creature in the game.
//
type CreatureToken struct {
	BaseMapObject

	// The name of the creature as displayed on the map. Must be unique
	// among the other creatures.
	Name string

	// If non-nil, this tracks the health status of the creature.
	Health *CreatureHealth

	// Grid (x, y) coordinates for the reference point of the
	// creature.  Unlike MapElement coordinates, these are in
	// grid units (1 grid = 5 feet).  The upper-left corner of
	// the creature token is at this location.
	Gx, Gy float64

	// For creatures which may change their shape or appearance,
	// multiple "skins" may be defined to display as appropriate.
	//
	// Skin is 0 for the default appearance of the creature, 1
	// for the alternate image, 2 for the 2nd alternate image, etc.
	Skin int

	// If the different "skins" are different sizes, this is a list
	// of size codes for each of them. For example, if there are 3
	// skins defined, the first two medium-size and the 3rd large
	// size, then SkinSize would have the value {"M", "M", "L"}.
	// If this is empty or nil, all skins are assumed to be the
	// size specified in the Size attribute. Note that SkinSize
	// also sets the Area at the same time.
	SkinSize []string

	// Current elevation in feet relative to the "floor" of the
	// current location.
	Elev int

	// The color to draw the creature's threat zone when in combat.
	Color string

	// A note to attach to the creature token to indicate special
	// conditions affecting the creature which are not otherwise shown.
	Note string

	// The tactical size category of the creature ("S", "M", "L",
	// etc). Lower-case letters indicate the "wide" version of the
	// category while upper-case indicates "tall" versions.
	//
	// May also be the size in feet (DEPRECATED).
	Size string

	// The tactical threat zone size of the creature, specified just
	// as with Size.
	Area string

	// A list of condition codes which apply to the character. These
	// are arbitrary and defined by the server according to the needs
	// of the particular game, but may include things such
	// as "confused", "helpless", "hasted", etc.
	StatusList []string

	// If there is a spell effect radiating from the creature, its
	// area of effect is described by this value. If there is none,
	// this is nil.
	//
	// Currently only radius emanations are supported. In future, the
	// type of this attribute may change to handle other shapes.
	AoE *RadiusAoE

	// The method of locomotion currently being used by this creature.
	// Normally this is MoveModeLand for land-based creatures which
	// are walking/running.
	MoveMode MoveModeType

	// Is the creature currently wielding a reach weapon or otherwise
	// using the "reach" alternate threat zone?
	Reach bool

	// Is the creature currently dead? (This takes precedence over the
	// Health attribute's indication that the creature has taken a
	// fatal amount of damage.)
	Killed bool

	// In combat, if this is true, the token is "dimmed" to indicate
	// that it is not their turn to act.
	Dim bool

	// The creature type.
	CreatureType CreatureTypeCode
}

//
// objCreature creates a new CreatureToken instance from the fields in objDef.
//
func objCreature(objID string, objDef map[string][]string) (CreatureToken, error) {
	var err error
	c := CreatureToken{
		CreatureType: CreatureTypeUnknown,
	}
	var b byte
	c.ID = objID
	c.Name, err = objString(objDef, 0, "NAME", true, err)
	c.Gx, err = objFloat(objDef, 0, "GX", true, err)
	c.Gy, err = objFloat(objDef, 0, "GY", true, err)
	c.Health, err = objHealth(objDef, err)
	c.Elev, err = objInt(objDef, 0, "ELEV", false, err)
	b, err = objEnum(objDef, 0, "MOVEMODE", false, enumMoveModes, err)
	c.MoveMode = MoveModeType(b)
	c.Color, err = objString(objDef, 0, "COLOR", false, err)
	c.Note, err = objString(objDef, 0, "NOTE", false, err)
	c.Skin, err = objInt(objDef, 0, "SKIN", false, err)
	c.SkinSize, err = objStrings(objDef, 0, "SKINSIZE", false, err)
	c.Size, err = objString(objDef, 0, "SIZE", true, err)
	c.StatusList, err = objStrings(objDef, 0, "STATUSLIST", false, err)
	c.AoE, err = objAoEShape(objDef, err)
	c.Area, err = objString(objDef, 0, "AREA", true, err)
	c.Reach, err = objBool(objDef, 0, "REACH", false, err)
	c.Killed, err = objBool(objDef, 0, "KILLED", false, err)
	c.Dim, err = objBool(objDef, 0, "DIM", false, err)

	return c, err
}

//
// CreatureHealth describes the current health statistics of a creature if we are
// tracking it for them.
//
type CreatureHealth struct {
	// The maximum hit points possible for the creature.
	MaxHP int `json:",omitempty"`

	// The amount of lethal and non-lethal damage suffered by the creature.
	LethalDamage    int `json:",omitempty"`
	NonLethalDamage int `json:",omitempty"`

	// The grace amount of hit points a creature may suffer over their maximum before
	// they are actually dead (as opposed to critically wounded). This is generally
	// the creature's Constitution score, hence the name.
	Con int `json:",omitempty"`

	// Is the creature flat-footed?
	IsFlatFooted bool `json:",omitempty"`

	// Has the creature been stabilized to prevent death while critically wounded?
	IsStable bool `json:",omitempty"`

	// Override the map client's idea of how to display the creature's health condition.
	// Normally this is the empty string which allows the client to calculate it from the
	// information available to it.
	Condition string `json:",omitempty"`

	// If 0, the creature's health is displayed accurately on the map. Otherwise,
	// this gives the percentage by which to "blur" the hit points as seen by the
	// players. For example, if HpBlur is 10, then hit points are displayed only in
	// 10% increments.
	HpBlur int `json:",omitempty"`
}

//
// objHealth looks for a HEALTH entry in the objDef map, parses
// it and returns the CreatureHealth struct it defines, or nil
// if no such entry was found, or it was the empty string.
//
// HEALTH {}
// HEALTH {max lethal sub con flat? stable? condition [blur]}
//         int int    int int bool  bool    str       {}/int
//
func objHealth(objDef map[string][]string, err error) (*CreatureHealth, error) {
	if err != nil {
		return nil, err
	}
	healthStats, ok := objDef["HEALTH"]
	if !ok || len(strings.TrimSpace(healthStats[0])) == 0 {
		return nil, err
	}

	return newHealth(healthStats[0], err)
}

func newHealth(healthStats string, err error) (*CreatureHealth, error) {
	var hs []string
	var hstats []interface{}
	h := CreatureHealth{}

	hs, err = tcllist.ParseTclList(healthStats)
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
// saveHealth converts a CreatureHealth value to a string as it
// appears in the file format.
//
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

//
// RadiusAoE describes the area of some spell or special effect emanating from the creature.
//
type RadiusAoE struct {
	// Distance in standard map pixels away from the creature token's center
	// to the perimeter of the affected area.
	Radius float64

	// Color to draw the affected zone.
	Color string
}

//
// objAoEShape looks for an AOE entry in the objDef map, returning a
// RadiusAoE value or nil if no such entry was found or it was the
// empty string.
//
// AOE {}
// AOE {radius r color}
//           float str
//
func objAoEShape(objDef map[string][]string, err error) (*RadiusAoE, error) {
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

//
// saveCreatureAoE converts a RadiusAoE value to the string representation
// for it in the output data.
//
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

//
// saveData converts a CreatureToken to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o CreatureToken) saveData(data []string, prefix, id string) ([]string, error) {
	var err error
	if data, err = o.BaseMapObject.saveData(data, prefix, id); err != nil {
		return nil, err
	}

	ss, err := tcllist.ToTclString(o.SkinSize)
	if err != nil {
		return nil, err
	}
	sl, err := tcllist.ToTclString(o.StatusList)
	if err != nil {
		return nil, err
	}
	ae, err := saveCreatureAoE(o.AoE)
	if err != nil {
		return nil, err
	}
	mm, err := saveEnum(byte(o.MoveMode), enumMoveModes)
	if err != nil {
		return nil, err
	}
	he, err := saveHealth(o.Health)
	if err != nil {
		return nil, err
	}

	var myType string
	switch o.CreatureType {
	case CreatureTypePlayer:
		myType = "player"
	case CreatureTypeMonster:
		myType = "monster"
	default:
		return nil, fmt.Errorf("CreatureToken has unknown type")
	}

	return saveValues(data, prefix, id, []saveAttributes{
		{"TYPE", "s", true, myType},
		{"NAME", "s", true, o.Name},
		{"GX", "f", true, o.Gx},
		{"GY", "f", true, o.Gy},
		{"SKIN", "i", true, o.Skin},
		{"SKINSIZE", "s", false, ss},
		{"ELEV", "i", true, o.Elev},
		{"COLOR", "s", true, o.Color},
		{"NOTE", "s", false, o.Note},
		{"SIZE", "s", true, o.Size},
		{"STATUSLIST", "s", false, sl},
		{"AOE", "s", false, ae},
		{"AREA", "s", true, o.Area},
		{"MOVEMODE", "s", false, mm},
		{"REACH", "b", false, o.Reach},
		{"KILLED", "b", true, o.Killed},
		{"DIM", "b", true, o.Dim},
		{"HEALTH", "s", false, he},
	})
}

//________________________________________________________________________________
//  ____  _                      _____     _
// |  _ \| | __ _ _   _  ___ _ _|_   _|__ | | _____ _ __
// | |_) | |/ _` | | | |/ _ \ '__|| |/ _ \| |/ / _ \ '_ \
// |  __/| | (_| | |_| |  __/ |   | | (_) |   <  __/ | | |
// |_|   |_|\__,_|\__, |\___|_|   |_|\___/|_|\_\___|_| |_|
//                |___/

// PlayerToken is a CreatureToken which describes a player character
// or NPC ally.
//
type PlayerToken struct {
	CreatureToken
}

//
// objPlayer creates a new PlayerToken instance from the fields in objDef.
//
func objPlayer(objID string, objDef map[string][]string) (PlayerToken, error) {
	c, err := objCreature(objID, objDef)
	c.CreatureType = CreatureTypePlayer
	return PlayerToken{
		CreatureToken: c,
	}, err
}

//
// saveData converts a PlayerToken to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o PlayerToken) saveData(data []string, prefix, id string) ([]string, error) {
	return o.CreatureToken.saveData(data, "P", id)
}

//________________________________________________________________________________
//  __  __                 _           _____     _
// |  \/  | ___  _ __  ___| |_ ___ _ _|_   _|__ | | _____ _ __
// | |\/| |/ _ \| '_ \/ __| __/ _ \ '__|| |/ _ \| |/ / _ \ '_ \
// | |  | | (_) | | | \__ \ ||  __/ |   | | (_) |   <  __/ | | |
// |_|  |_|\___/|_| |_|___/\__\___|_|   |_|\___/|_|\_\___|_| |_|
//

// MonsterToken is a CreatureToken which describes a monster or NPC adversary.
//
type MonsterToken struct {
	CreatureToken
}

//
// objMonster creates a new MonsterToken instance from the fields in objDef.
//
func objMonster(objID string, objDef map[string][]string) (MonsterToken, error) {
	c, err := objCreature(objID, objDef)
	c.CreatureType = CreatureTypeMonster
	return MonsterToken{
		CreatureToken: c,
	}, err
}

//
// saveData converts a MonsterToken to a text representation
// in the map file format (suitable for sending to clients or saving to a disk
// file).
//
// This works just as described for BaseMapElement.saveData.
//
func (o MonsterToken) saveData(data []string, prefix, id string) ([]string, error) {
	return o.CreatureToken.saveData(data, "M", id)
}

//________________________________________________________________________________
//  ___                            ____        __ _       _ _   _
// |_ _|_ __ ___   __ _  __ _  ___|  _ \  ___ / _(_)_ __ (_) |_(_) ___  _ __
//  | || '_ ` _ \ / _` |/ _` |/ _ \ | | |/ _ \ |_| | '_ \| | __| |/ _ \| '_ \
//  | || | | | | | (_| | (_| |  __/ |_| |  __/  _| | | | | | |_| | (_) | | | |
// |___|_| |_| |_|\__,_|\__, |\___|____/ \___|_| |_|_| |_|_|\__|_|\___/|_| |_|
//                      |___/

// ImageDefinition describes an image as known to the mapper system.
// TileElements' Image attribute refers to the Name attribute of one of
// these.
//
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

//________________________________________________________________________________
//  _____ _ _      ____        __ _       _ _   _
// |  ___(_) | ___|  _ \  ___ / _(_)_ __ (_) |_(_) ___  _ __
// | |_  | | |/ _ \ | | |/ _ \ |_| | '_ \| | __| |/ _ \| '_ \
// |  _| | | |  __/ |_| |  __/  _| | | | | | |_| | (_) | | | |
// |_|   |_|_|\___|____/ \___|_| |_|_| |_|_|\__|_|\___/|_| |_|
//

// FileDefinition describes a file as known to the mapper which
// may be of interest to retrieve at some point.
//
type FileDefinition struct {
	// The filename or Server ID.
	File string

	// If IsLocalFile is true, File is the name of the file on disk;
	// otherwise it is the server's internal ID by which you may request
	// that file from the server.
	IsLocalFile bool
}

//________________________________________________________________________________
//  ____                      ___  _     _           _
// |  _ \ __ _ _ __ ___  ___ / _ \| |__ (_) ___  ___| |_ ___
// | |_) / _` | '__/ __|/ _ \ | | | '_ \| |/ _ \/ __| __/ __|
// |  __/ (_| | |  \__ \  __/ |_| | |_) | |  __/ (__| |_\__ \
// |_|   \__,_|_|  |___/\___|\___/|_.__// |\___|\___|\__|___/
//                                    |__/

//
// ParseObjects reads a map file which has already been loaded from disk or
// received from the server into a string slice, one string per line.
// These lines are parsed to construct
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

				if format < MinimumSupportedMapFileFormat || format > MaximumSupportedMapFileFormat {
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
	// may even be interleaved.) Now that we've sorted them out we can look
	// at each object individually.
	//
	oList := make([]MapObject, 0, len(objects))
	var o MapObject
	var err error

	for objID, objDef := range objects {
		mType, ok := objDef["__mob_type__"]
		if ok {
			switch mType[0] {
			case "M":
				o, err = objMonster(objID, objDef)
			case "P":
				o, err = objPlayer(objID, objDef)
			default:
				err = fmt.Errorf("unknown creature type (%s) for ID %s", mType, objID)
			}
		} else {
			oType, ok := objDef["TYPE"]
			if ok {
				switch oType[0] {
				case "aoe":
					o, err = objSpellAreaOfEffectElement(objID, objDef)
				case "arc":
					o, err = objArcElement(objID, objDef)
				case "circ":
					o, err = objCircleElement(objID, objDef)
				case "line":
					o, err = objLineElement(objID, objDef)
				case "poly":
					o, err = objPolygonElement(objID, objDef)
				case "rect":
					o, err = objRectangleElement(objID, objDef)
				case "text":
					o, err = objTextElement(objID, objDef)
				case "tile":
					o, err = objTileElement(objID, objDef)
				case "player":
					o, err = objPlayer(objID, objDef)
				case "monster":
					o, err = objMonster(objID, objDef)
				default:
					err = fmt.Errorf("unknown element type (%s) for ID %s", oType, objID)
				}
			} else {
				err = fmt.Errorf("element ID %s missing TYPE attribute", objID)
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
// objFloat looks for the given field in objDef, returning it as a float64 value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//
// objInt looks for the given field in objDef, returning it as an int value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//
// objBool looks for the given field in objDef, returning it as a bool value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//
// objString looks for the given field in objDef, returning it as a string value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//
// objEnum looks for the given field in objDef, returning it as a byte value
// according to which of the defined valid value strings was found.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
// choices is a map of string to byte value. It is an error if the string found in
// the data set is not found among these choices.
//

type enumChoices map[string]byte

type enumDescriptor struct {
	minValue, maxValue byte
	choices            map[string]byte
}

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

//
// strEnum is like objEnum but simply looks at the input string value for a
// match to the allowed choices rather than looking for a named field.
//
// If required is true, it is an error if the input is the empty string; otherwise
// the zero value is returned for an empty input.
//
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

//
// objStrings looks for the given field in objDef, returning it as a string slice value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// The value of the field is interpreted as a Tcl list of sub-fields to be returned.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//
// objCoordinateList looks for the given field in objDef, returning it as a []Coordinate value.
// If required is true, it is an error if no such value is found; otherwise a
// missing value is returned as the zero value.
//
// i indicates which element of the field's value is to be retrieved. Most of our
// fields are singletons, so this is usually 0.
//
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

//________________________________________________________________________________
//  ____                   ___  _     _           _
// / ___|  __ ___   _____ / _ \| |__ (_) ___  ___| |_ ___
// \___ \ / _` \ \ / / _ \ | | | '_ \| |/ _ \/ __| __/ __|
//  ___) | (_| |\ V /  __/ |_| | |_) | |  __/ (__| |_\__ \
// |____/ \__,_| \_/ \___|\___/|_.__// |\___|\___|\__|___/
//                                 |__/

// SaveObjects reverses the operation of ParseObjects. It accepts a slice of
// MapObjects, a map containing ImageDefinitions, and a slice of FileDefinitions,
// and emits as a slice of strings a text description of those objects in the
// map file format, suitable for saving to disk or sending to clients and servers.
//
// A number of options may be given at the end of the argument list, including:
//   WithoutHeader   -- suppress output of the normal header line
//   WithDate(t)     -- use the given date instead of the current one
//   WithComment(s)  -- include the comment string in the header line
//
func SaveObjects(objects []MapObject, images map[string]ImageDefinition, files []FileDefinition, options ...func(*saveObjOpts)) ([]string, error) {
	var err error
	opts := saveObjOpts{}
	data := make([]string, 0, 32)

	for _, o := range options {
		o(&opts)
	}

	if !opts.suppressHeader {
		if opts.date.IsZero() {
			opts.date = time.Now()
		}

		fileDate, err := tcllist.ToTclString([]string{
			strconv.FormatInt(opts.date.Unix(), 10),
			opts.date.Format(time.UnixDate),
		})
		if err != nil {
			return nil, err
		}

		commentHdr, err := tcllist.ToTclString([]string{
			opts.comment,
			fileDate,
		})
		if err != nil {
			return nil, err
		}

		header, err := tcllist.ToTclString([]string{
			fmt.Sprintf("__MAPPER__:%d", GMAMapperFileFormat),
			commentHdr,
		})
		if err != nil {
			return nil, err
		}
		data = append(data, header)
	}

	for _, o := range objects {
		data, err = o.saveData(data, "", o.ObjID())
		if err != nil {
			return nil, fmt.Errorf("could not save object %s: %v", o.ObjID(), err)
		}
	}

	for _, im := range images {
		imdata, err := tcllist.ToTclString([]string{
			"I",
			im.Name,
			fmt.Sprintf("%g", im.Zoom),
			im.File,
		})
		if err != nil {
			return nil, fmt.Errorf("error saving image %s: %v", im.Name, err)
		}
		data = append(data, imdata)
	}

	for _, f := range files {
		fdata, err := tcllist.ToTclString([]string{
			"F",
			f.File,
		})
		if err != nil {
			return nil, fmt.Errorf("error saving file %s: %v", f.File, err)
		}
		data = append(data, fdata)
	}

	return data, nil
}

//
// saveObjOpts gives configuration options for how we save objects.
//
type saveObjOpts struct {
	comment        string
	date           time.Time
	suppressHeader bool
}

//
// WithoutHeader modifies a call to SaveObjects by suppressing the
// normal "__MAPPER__" header line that should be in a map file.
//
// This may be used, for example, if SaveObjects is generating output
// that will only be part of, but not the entire, saved data set.
//
func WithoutHeader(o *saveObjOpts) {
	o.suppressHeader = true
}

//
// WithDate modifies a call to SaveObjects by specifying
// an already-determined date to record into the
// "__MAPPER__" header line. Normally, the current date and time is
// used.
//
func WithDate(d time.Time) func(*saveObjOpts) {
	return func(o *saveObjOpts) {
		o.date = d
	}
}

//
// WithComment modifies a call to SaveObjects by providing
// a comment string to include in the "__MAPPER__" header line.
//
func WithComment(c string) func(*saveObjOpts) {
	return func(o *saveObjOpts) {
		o.comment = c
	}
}

//
// This describes each attribute to save to the output
// data set by the saveValues function.
//
type saveAttributes struct {
	Tag      string
	Type     string
	Required bool
	Value    interface{}
}

//
// saveBool converts a bool value to the string as it will appear in the output
//
func saveBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

//
// saveEnum converts an enum type to the string representing that value
// as indicated by the supplied choice list.
//
func saveEnum(choice byte, choices enumChoices) (string, error) {
	for k, v := range choices {
		if v == choice {
			return k, nil
		}
	}
	return "", fmt.Errorf("value %v not in list of valid enum choices %v", choice, choices)
}

func enumToString(v byte, choices enumChoices, desc string) string {
	s, err := saveEnum(v, choices)
	if err != nil {
		return fmt.Sprintf("(unknown %s)", desc)
	}
	return s
}

//
// MoveModeString returns a string representation of the internal MoveMode value.
//
func MoveModeString(m MoveModeType) string {
	return enumToString(byte(m), enumMoveModes, "move mode")
}

//
// DashTypeString returns a string representation of the internal DashType value.
//
func DashTypeString(m DashType) string {
	return enumToString(byte(m), enumDashes, "dash type")
}

//
// ArcModeString returns a string representation of the internal ArcModeType value.
//
func ArcModeString(m ArcModeType) string {
	return enumToString(byte(m), enumArcs, "arc type")
}

//
// ArrowTypeString returns a string representation of the internal ArrowType value.
//
func ArrowTypeString(m ArrowType) string {
	return enumToString(byte(m), enumArrows, "arrow type")
}

//
// JoinStyleString returns a string representation of the internal JoinStyle value.
//
func JoinStyleString(m JoinStyle) string {
	return enumToString(byte(m), enumJoins, "join style")
}

//
// AoETypeString returns a string representation of the internal AoEType value.
//
func AoETypeString(m AoEType) string {
	return enumToString(byte(m), enumAoeShapes, "AoE shape")
}

//
// FontWeightString returns a string representation of the internal FontWeight value.
//
func FontWeightString(m FontWeightType) string {
	return enumToString(byte(m), enumFontWeights, "font weight")
}

//
// FontSlantString returns a string representation of the internal FontSlantType value.
//
func FontSlantString(m FontSlantType) string {
	return enumToString(byte(m), enumFontSlants, "font slant")
}

//
// AnchorDirectionString returns a string representation of the internal AnchorDirection value.
//
func AnchorDirectionString(m AnchorDirection) string {
	return enumToString(byte(m), enumAnchors, "anchor direction")
}

//
// CreatureTypeString returns a string representation of the internal CreatureTypeCode value.
//
func CreatureTypeString(t CreatureTypeCode) string {
	switch t {
	case CreatureTypeMonster:
		return "monster"
	case CreatureTypePlayer:
		return "player"
	default:
		return "(unknown creature type)"
	}
}

//
// saveValues saves a series of values according to the description of each
// in saveAttributes, converting generic types to strings.
//
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
			if !attr.Value.(bool) && !attr.Required {
				s = ""
			} else {
				s = saveBool(attr.Value.(bool))
			}
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
