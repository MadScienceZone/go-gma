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

//
// MapObject describes the elements that may appear on the map.
//

package mapper

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MadScienceZone/go-gma/v5/tcllist"
)

//
// GMAMapperFileFormat gives
// the GMA File Format version number current as of this build.
// This is the format which will be used for saving map data.
//
const GMAMapperFileFormat = 23 // @@##@@ auto-configured
//
// MinimumSupportedMapFileFormat gives the lowest file format this package can
// understand.
//
const MinimumSupportedMapFileFormat = 17

//
// MaximumSupportedMapFileFormat gives the highest file format this package
// can understand. Saved data will be in this format.
//
const MaximumSupportedMapFileFormat = 23

// ErrCreatureNoSizes is the error returned when a creature size code is expected but none given.
var ErrCreatureNoSizes = errors.New("missing creature size code")

// ErrCreatureInvalidSize is the error returned when a creature size code is expected but the code(s) given are invalid.
var ErrCreatureInvalidSize = errors.New("invalid creature size code")

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

//
// These are the allowed values for the ArcMode attribute of an ArcElement.
//
type ArcModeType byte

const (
	ArcModePieSlice ArcModeType = iota
	ArcModeArc
	ArcModeChord
)

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

//
// These are the allowed values for the Join attribute of a PolygonElement.
//
type JoinStyle byte

const (
	JoinBevel JoinStyle = iota
	JoinMiter
	JoinRound
)

//
// These are the valid values for the AoEShape attribute.
//
type AoEType byte

const (
	AoEShapeCone AoEType = iota
	AoEShapeRadius
	AoEShapeRay
)

//
// The valid font weights.
//
type FontWeightType byte

const (
	FontWeightNormal FontWeightType = iota
	FontWeightBold
)

//
// The valid font slants.
//
type FontSlantType byte

const (
	FontSlantRoman FontSlantType = iota
	FontSlantItalic
)

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

/* XXX
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
*/

/* XXX
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
*/

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
}

//
// Coordinates give an (x, y) coordinate pair to locate something on the map.
// Coordinates are in standard map pixel units (10 pixels = 1 foot).
//
type Coordinates struct {
	X, Y float64
}

//
// MapObject
//  BaseMapObject
//   MapElement
//     objMapElement
//    ArcElement
//      objArcElement
//    CircleElement
//      objCircleElement
//    LineElement
//      objLineElement
//    PolygonElement
//      objPolygonElement
//    RectangleElement
//      objRectangleElement
//    SpellAreaOfEffect
//      objSpellAreaOfEffect
//    (TextFont)
//      objTextFont
//    TextElement
//      objTextElement
//    TileElement
//      objTileElement
//   CreatureToken
//     objCreature
//    PlayerToken
//      objPlayer
//    MonsterToken
//      objMonster
//    (CreatureHealth)
//      objHealth
//      newHealth
//    (RadiusAoE)
//      objAoEShape
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

	// Is this element currently concealed from view?
	Hidden bool `json:",omitempty"`

	// Is the object locked from editing by the user?
	Locked bool `json:",omitempty"`

	// The element's line(s) are to be drawn with this dash pattern.
	Dash DashType `json:",omitempty"`

	// The z "coordinate" is the vertical stacking order relative to the other
	// displayed on-screen objects.
	Z int

	// The width in pixel units to draw the element's outline.
	Width int `json:",omitempty"`

	// The dungeon level where this element appears. Typically, level 0
	// is the default (ground) level, with level numbers increasing as
	// 1, 2, 3, etc., for floors above it, and with underground levels
	Level int `json:",omitempty"`

	// Objects which need additional coordinate pairs to describe their
	// geometry (beyond the standard reference point) store them here.
	Points []Coordinates `json:",omitempty"`

	// The colors used to draw the element's outline and/or to fill it's interior.
	// These may be standard color names such as "blue" or an RGB string such as
	// "#336699". A fill color that is the empty string means not to fill that element.
	// If Stipple is nonempty, it specifies that the shape should be filled with
	// a stipple pattern. Pattern names "gray12", "gray25", "gray50", and "gray75"
	// should be available by default in clients.
	Line    string `json:",omitempty"`
	Fill    string `json:",omitempty"`
	Stipple string `json:",omitempty"`

	// The map layer this element belongs to.
	Layer string `json:",omitempty"`

	// Elements may be arranged into logical groups to be manipulated
	// together. This is the ID of the group to which this belongs, or
	// is empty if this element is not grouped.
	Group string `json:",omitempty"`
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
	Start   float64
	Extent  float64
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
	Arrow ArrowType `json:",omitempty"`
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

	// The join style to control how the intersection between line segments is drawn.
	Join JoinStyle `json:",omitempty"`

	// Spline gives the factor to use when smoothing the sides of the polygon between
	// its points. 0 means not to smooth them at all, resulting in a shape with straight
	// edges between the vertices. Otherwise, larger values provide greater smoothing.
	Spline float64 `json:",omitempty"`
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
	Weight FontWeightType `json:",omitempty"`

	// The font slant (roman or italic).
	Slant FontSlantType `json:",omitempty"`
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

	// Where is the reference point in relation to the text?
	Anchor AnchorDirection `json:",omitempty"`

	// The text to be displayed.
	Text string

	// Font to use for the text.
	Font TextFont
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
	BBHeight, BBWidth float64 `json:",omitempty"`
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

// SetSizes handles the bridge bewteen the deprecated use of the
// Size field and the new use of SkinSize to handle both polymorph
// skins and the creature's native size.
//
// Given (possibly nil or empty) skinSize, skin, and size values,
// the SetSizes method gives preference to the SkinSize array,
// setting Size to the default size from that list for backward
// compatibility. If SkinSize is empty but Size is populated, the
// opposite is done. If an explicit default size is listed in
// SkinSize, that will override a zero value in the Size field.
// A nonzero Size value will take priority over the default size,
// however.
//
// An error is returned if one of the size codes is invalid. If there
// is a conflict between the size value and the default size in skinSizes,
// skinSizes takes priority.
//
func (c *CreatureToken) SetSizes(skinSize []string, skin int, size string) error {
	sizeCodeRE := regexp.MustCompile(`^[fFdDtTsSmMlLhHgGcC](\d+)?(?:->(\d+))?(?:=(\d+))?(?::(\*)?(.*))?$`)

	// If no skinSize list was present, just use size as the single item for that list.
	if len(skinSize) == 0 {
		if size == "" {
			return ErrCreatureNoSizes
		}
		if sizeCodeRE.FindStringSubmatch(size) == nil {
			return ErrCreatureInvalidSize
		}
		c.SkinSize = []string{size}
		c.Skin = 0
		c.Size = size
		return nil
	}

	// If they gave us a list of skin sizes, make sure they're all valid and see if one is
	// designated as the default one.
	defaultSkinSize := 0
	for i, ss := range skinSize {
		if szFields := sizeCodeRE.FindStringSubmatch(ss); szFields != nil {
			if szFields[4] == "*" {
				defaultSkinSize = i
			}
		} else {
			return ErrCreatureInvalidSize
		}
	}

	if skin <= 0 || skin >= len(skinSize) {
		skin = defaultSkinSize
	}
	c.SkinSize = skinSize
	c.Skin = skin
	c.Size = c.SkinSize[c.Skin]
	return nil
}

//
// CreatureToken is a MapObject (but not a MapElement) which displays a movable
// token indicating the size and location of a creature in the game.
//
type CreatureToken struct {
	BaseMapObject

	// Is the creature currently dead? (This takes precedence over the
	// Health attribute's indication that the creature has taken a
	// fatal amount of damage.)
	Killed bool `json:",omitempty"`

	// In combat, if this is true, the token is "dimmed" to indicate
	// that it is not their turn to act.
	Dim bool `json:",omitempty"`

	// If true, this means the creature token is only visible to the GM.
	Hidden bool `json:",omitempty"`

	// If true, only the GM may access or manipulate the polymorph capabilities
	// of this creature.
	PolyGM bool `json:",omitempty"`

	// The creature type.
	CreatureType CreatureTypeCode

	// The method of locomotion currently being used by this creature.
	// Normally this is MoveModeLand for land-based creatures which
	// are walking/running.
	MoveMode MoveModeType `json:",omitempty"`

	// Is the creature currently wielding a reach weapon or otherwise
	// using the "reach" alternate threat zone?
	// If this value is 0, the threat zone is normal for a creature
	// of its size. If 1, an extended area (appropriate for using
	// a reach weapon) is used instead. If 2, both areas are used,
	// so the creature may attack into the reach zone AND adjacent
	// foes.
	Reach int `json:",omitempty"`

	// For creatures which may change their shape or appearance,
	// multiple "skins" may be defined to display as appropriate.
	//
	// Skin is 0 for the default appearance of the creature, 1
	// for the alternate image, 2 for the 2nd alternate image, etc.
	Skin int `json:",omitempty"`

	// Current elevation in feet relative to the "floor" of the
	// current location.
	Elev int `json:",omitempty"`

	// Grid (x, y) coordinates for the reference point of the
	// creature.  Unlike MapElement coordinates, these are in
	// grid units (1 grid = 5 feet).  The upper-left corner of
	// the creature token is at this location.
	Gx, Gy float64

	// The name of the creature as displayed on the map. Must be unique
	// among the other creatures.
	Name string

	// If non-nil, this tracks the health status of the creature.
	Health *CreatureHealth `json:",omitempty"`

	// If the different "skins" are different sizes, this is a list
	// of size codes for each of them. For example, if there are 3
	// skins defined, the first two medium-size and the 3rd large
	// size, then SkinSize would have the value {"M", "M", "L"}.
	// If this is empty or nil, all skins are assumed to be the
	// size specified in the Size attribute. Note that SkinSize
	// also sets the Area at the same time.
	SkinSize []string `json:",omitempty"`

	// The color to draw the creature's threat zone when in combat.
	Color string

	// A note to attach to the creature token to indicate special
	// conditions affecting the creature which are not otherwise shown.
	Note string `json:",omitempty"`

	// The tactical size category of the creature ("S", "M", "L",
	// etc). Lower-case letters indicate the "wide" version of the
	// category while upper-case indicates "tall" versions.
	//
	// May also be the size in feet (DEPRECATED USAGE).
	//
	// This field is now DEPRECATED. Use SkinSize instead.
	Size string

	// If DispSize is nonempty, it holds the size category to display
	// the creature (say, as a result of casting an enlarge person spell).
	DispSize string `json:",omitempty"`

	// A list of condition codes which apply to the character. These
	// are arbitrary and defined by the server according to the needs
	// of the particular game, but may include things such
	// as "confused", "helpless", "hasted", etc.
	StatusList []string `json:",omitempty"`

	// If there is a spell effect radiating from the creature, its
	// area of effect is described by this value. If there is none,
	// this is nil.
	//
	// Currently only radius emanations are supported. In future, the
	// type of this attribute may change to handle other shapes.
	AoE *RadiusAoE `json:",omitempty"`

	// If there is a custom reach/threat zone defined for this
	// creature, it is detailed here.
	CustomReach CreatureCustomReach `json:",omitempty"`
}

// CreatureCustomReach describes a creature's natural and extended
// reach zones if they differ from the standard templates.
type CreatureCustomReach struct {
	// Enabled is true if this custom information should be used for the creature.
	Enabled bool `json:",omitempty"`

	// Natural and Extended give the distance in 5-foot grid squares
	// away from the creature's PERIMETER to which their natural reach
	// and extended (as with a reach weapon) reach extends.
	Natural  int `json:",omitempty"`
	Extended int `json:",omitempty"`
}

//
// CreatureHealth describes the current health statistics of a creature if we are
// tracking it for them.
//
type CreatureHealth struct {
	// Is the creature flat-footed?
	IsFlatFooted bool `json:",omitempty"`

	// Has the creature been stabilized to prevent death while critically wounded?
	IsStable bool `json:",omitempty"`

	// The maximum hit points possible for the creature.
	MaxHP int `json:",omitempty"`

	// The amount of lethal and non-lethal damage suffered by the creature.
	LethalDamage    int `json:",omitempty"`
	NonLethalDamage int `json:",omitempty"`

	// The grace amount of hit points a creature may suffer over their maximum before
	// they are actually dead (as opposed to critically wounded). This is generally
	// the creature's Constitution score, hence the name.
	Con int `json:",omitempty"`

	// If 0, the creature's health is displayed accurately on the map. Otherwise,
	// this gives the percentage by which to "blur" the hit points as seen by the
	// players. For example, if HPBlur is 10, then hit points are displayed only in
	// 10% increments.
	HPBlur int `json:",omitempty"`

	// Override the map client's idea of how to display the creature's health condition.
	// Normally this is the empty string which allows the client to calculate it from the
	// information available to it.
	Condition string `json:",omitempty"`
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

/*
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
*/

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

//________________________________________________________________________________
//  ___                            ____        __ _       _ _   _
// |_ _|_ __ ___   __ _  __ _  ___|  _ \  ___ / _(_)_ __ (_) |_(_) ___  _ __
//  | || '_ ` _ \ / _` |/ _` |/ _ \ | | |/ _ \ |_| | '_ \| | __| |/ _ \| '_ \
//  | || | | | | | (_| | (_| |  __/ |_| |  __/  _| | | | | | |_| | (_) | | | |
// |___|_| |_| |_|\__,_|\__, |\___|____/ \___|_| |_|_| |_|_|\__|_|\___/|_| |_|
//                      |___/

//
// ImageDefinition describes an image as known to the mapper system.
// TileElements' Image attribute refers to the Name attribute of one of
// these.
//
type ImageDefinition struct {
	// The name of the image as known within the mapper.
	Name  string
	Sizes []ImageInstance

	// If non-nil, this indicates that the image is to be animated
	Animation *ImageAnimation `json:",omitempty"`
}

// ImageInstance is a single instance of a file (generally, an image is availble
// in several formats and sizes, each of which is an instance).
type ImageInstance struct {
	// If IsLocalFile is true, File is the name of the image file on disk;
	// otherwise it is the server's internal ID by which you may request
	// that file from the server.
	IsLocalFile bool `json:",omitempty"`

	// The zoom (magnification) level this bitmap represents for the given
	// image.
	Zoom float64

	// The filename by which the image can be retrieved.
	File string

	// If non-nil, this holds the image data received directly
	// from the server. This usage is not recommended but still
	// supported.
	ImageData []byte `json:",omitempty"`
}

// ImageAnimation describes the animation parameters for animated images.
type ImageAnimation struct {
	// The number of frames in the animation. These will be in files with
	// :n: prepended to their names where n is the frame number starting with 0.
	Frames int `json:",omitempty"`

	// The number of milliseconds to display each frame.
	FrameSpeed int `json:",omitempty"`

	// The number of loops to play before stopping. 0 means unlimited.
	Loops int `json:",omitempty"`
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
	// If IsLocalFile is true, File is the name of the file on disk;
	// otherwise it is the server's internal ID by which you may request
	// that file from the server.
	IsLocalFile bool `json:",omitempty"`

	// The filename or Server ID.
	File string
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
		if !required {
			return 0, err
		}
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
		if !required {
			return 0, err
		}
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
		if !required {
			return false, err
		}
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
		if !required {
			return "", err
		}
		return "", fmt.Errorf("attribute %s only has %d elements; can't get [%d]", fldName, len(val), i)
	}
	return val[i], err
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
		if !required {
			return nil, err
		}
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
		if !required {
			return nil, err
		}
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

//
// MapMetaData describes a mapper location save file (itself, not its contents)
//
type MapMetaData struct {
	// Timestamp is the generation or modification time of the map file
	// as a 64-bit integer Unix timestamp value.
	Timestamp int64 `json:",omitempty"`

	// DateTime is a human-readable string which gives the same information
	// as Timestamp. The software only uses Timestamp. DateTime is provided
	// only for convenience and is not guaranteed to match Timestamp's time
	// value or even any valid value at all. The format is this string is
	// arbitrary.
	DateTime string `json:",omitempty"`

	// Comment is any brief comment the map author wishes to leave in the file
	// about this map.
	Comment string `json:",omitempty"`

	// Location is a string describing the locale within the adventure area or
	// world which is represented by this file.
	Location string `json:",omitempty"`

	// FileVersion is the file format version number detected when reading in
	// the data from this file. This is for informational purposes only and does
	// not control the format used to write the data to a new file.
	FileVersion uint `json:"-"`
}

//
// WriteMapFile writes mapper data from a slice of map object values and
// MapMetaData struct into the named file.  It is identical to
// SaveMapFile other than the fact that it creates and opens the requested
// file to be written into.
//
func WriteMapFile(path string, objList []any, meta MapMetaData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: mapobject WriteMapFile was unable to close the output file: %v\n", err)
		}
	}()
	return SaveMapFile(file, objList, meta)
}

//
// SaveMapFile is the same as WriteMapFile, except it writes to an open
// data stream.
//
// If the Timestamp field of the metadata is zero, the current date and
// time will be written to the Timestamp and DateTime fields on output.
//
func SaveMapFile(output io.Writer, objList []any, meta MapMetaData) error {
	writer := bufio.NewWriter(output)
	writer.WriteString("__MAPPER__:21\n")
	if meta.Timestamp == 0 {
		now := time.Now()
		meta.Timestamp = now.Unix()
		meta.DateTime = now.String()
	}
	data, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return err
	}
	writer.WriteString("«__META__» ")
	writer.WriteString(string(data))
	writer.WriteString("\n")

	//
	// To produce consistent output (and aid testing), we
	// will always write out records sorted by ObjID values.
	//
	sort.Slice(objList, func(i, j int) bool {
		switch a := objList[i].(type) {
		case ImageDefinition:
			if b, ok := objList[j].(ImageDefinition); ok {
				return a.Name < b.Name
			}
			return true // images are less than everything else

		case FileDefinition:
			if b, ok := objList[j].(FileDefinition); ok {
				return a.File < b.File
			}
			return true // files are next after image definitions

		case MapObject:
			if b, ok := objList[j].(MapObject); ok {
				return a.ObjID() < b.ObjID()
			}
			return false // objects go last
		default:
			return i < j
		}
	})

	for _, obj := range objList {
		data, err := json.MarshalIndent(obj, "", "    ")
		if err != nil {
			return fmt.Errorf("unable to serialize map object: %v", err)
		}

		switch obj.(type) {
		case ArcElement:
			writer.WriteString("«ARC» ")
		case CircleElement:
			writer.WriteString("«CIRC» ")
		case LineElement:
			writer.WriteString("«LINE» ")
		case PolygonElement:
			writer.WriteString("«POLY» ")
		case RectangleElement:
			writer.WriteString("«RECT» ")
		case SpellAreaOfEffectElement:
			writer.WriteString("«SAOE» ")
		case TextElement:
			writer.WriteString("«TEXT» ")
		case TileElement:
			writer.WriteString("«TILE» ")
		case ImageDefinition:
			writer.WriteString("«IMG» ")
		case FileDefinition:
			writer.WriteString("«MAP» ")
		case CreatureToken:
			writer.WriteString("«CREATURE» ")
		default:
			return fmt.Errorf("unable to serialize map object: unsupported type")
		}
		writer.WriteString(string(data))
		writer.WriteString("\n")
	}
	writer.WriteString("«__EOF__»\n")
	writer.Flush()
	return nil
}

//
// ReadMapFile loads GMA mapper data from the named file, returning the data as three values: a slice of
// MapObject values (which the caller will want to interpret based on their actual data type), the file
// metadata, and an error (which will be nil if everything went as planned).
//
// Other than opening the named input file, it is identical to LoadMapFile.
//
func ReadMapFile(path string) ([]any, MapMetaData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, MapMetaData{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: mapobject ReadMapFile was unable to close the file: %v", err)
		}
	}()
	return LoadMapFile(file)
}

//
// ReadMapMetaData is just like ReadMapFile, except that it only goes as far
// as reading the metadata from the file, returning that, but including any of
// the actual map data.
//
// Its operation is identical to LoadMapMetaData other than opening the input file for you.
//
func ReadMapMetaData(path string) (MapMetaData, error) {
	file, err := os.Open(path)
	if err != nil {
		return MapMetaData{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: mapobject ReadMapMetaData was unable to close the file: %v", err)
		}
	}()
	return LoadMapMetaData(file)
}

//
// loadLegacyMapFile(scanner
// reads a mapper file with format < 20, returning a slice
// of map elements.
//
func loadLegacyMapFile(scanner *bufio.Scanner, meta MapMetaData, legacyMeta string, metaDataOnly bool) ([]any, MapMetaData, error) {
	//
	// The map file formats prior to version 20 used a very different
	// format. This function is called *after* we have read the initial
	// line which held the version number and metadata. This function
	// picks up from that point.
	//
	// The legacy format's initial line was of the form
	//    __MAPPER__:<version> {<comment> {<timestamp> {<timestring>}}}
	// Following this are a series of lines of the form
	//    <attr>:<id> [<value-list>]
	// where <value-list> is a TCL list with values appropriate for the
	// given <attr> of the map objectd <id> (whose type is specified by
	// the TYPE attribute).
	// OR M <attr>:<id> [<value-list>]  for monster tokens,
	// OR P <attr>:<id> [<value-list>]  for player tokens,
	// OR F <serverID>                  for map files,
	// OR I <name> <zoom> <serverID>    for images.
	//
	// The lines may be in ANY order, so it's necessary to read all the
	// lines in the file and then assemble them into objects.
	//
	var rawFiles []string
	var objList []any
	var err error

	// rawData holds the file's data strings as a map of <id> to map of <attr> to <value-list>
	rawData := make(map[string]map[string][]string)
	rawMonsters := make(map[string]map[string][]string)
	rawPlayers := make(map[string]map[string][]string)
	rawImages := make(map[string]ImageDefinition)

	metaList, err := tcllist.ParseTclList(legacyMeta)
	if err != nil {
		return nil, meta, fmt.Errorf("legacy map file has invalid metadata: %v", err)
	}
	if len(metaList) > 0 {
		metaList, err = tcllist.ParseTclList(metaList[0])
		if err != nil {
			return nil, meta, fmt.Errorf("legacy map file has invalid metadata: %v", err)
		}
		if len(metaList) > 0 {
			meta.Comment = metaList[0]
			if len(metaList) > 1 {
				if dateList, err := tcllist.ParseTclList(metaList[1]); err == nil {
					meta.Timestamp, _ = strconv.ParseInt(dateList[0], 10, 64)
					if len(dateList) > 1 {
						meta.DateTime = dateList[1]
					}
				}
			}
		}
	}
	if metaDataOnly {
		return nil, meta, nil
	}

	for scanner.Scan() {
		f, err := tcllist.ParseTclList(scanner.Text())
		if err != nil {
			return nil, meta, fmt.Errorf("legacy map file has invalid record: %v", err)
		}
		switch f[0] {
		case "M": // M <attr>:<id> <value>	-> rawMonsters[<id>][<attr>] = []<value>
			attr, objID, ok := strings.Cut(f[1], ":")
			if !ok {
				return nil, meta, fmt.Errorf("legacy map file has improperly formed M record (can't parse <attr>:<id> from \"%s\")", f[1])
			}
			if rawMonsters[objID] == nil {
				rawMonsters[objID] = make(map[string][]string)
			}
			rawMonsters[objID][attr] = f[2:]

		case "P": // P <attr>:<id> <value>	-> rawPlayers[<id>][<attr>] = []<value>
			attr, objID, ok := strings.Cut(f[1], ":")
			if !ok {
				return nil, meta, fmt.Errorf("legacy map file has improperly formed P record (can't parse <attr>:<id> from \"%s\")", f[1])
			}
			if rawPlayers[objID] == nil {
				rawPlayers[objID] = make(map[string][]string)
			}
			rawPlayers[objID][attr] = f[2:]

		case "F": // F <serverID>		-> rawFiles[] = <serverID>
			if len(f) != 2 {
				return nil, meta, fmt.Errorf("legacy map file has improperly formed F record (%d fields)", len(f))
			}
			rawFiles = append(rawFiles, f[1])

		case "I": // I <name> <zoom> <serverID>	-> rawImages[<name>] = ImageDefinition struct (with multiple sizes)
			// interprets @<serverID> notation
			ff, err := tcllist.ConvertTypes(f, "ssfs")
			if err != nil {
				return nil, meta, fmt.Errorf("legacy map file has improperly formed I record: %v", err)
			}
			serverID := ff[3].(string)

			var def ImageDefinition
			var ok bool
			def, ok = rawImages[ff[1].(string)]
			if !ok {
				def = ImageDefinition{
					Name: ff[1].(string),
				}
			}
			if len(serverID) > 0 && serverID[0] == '@' {
				def.Sizes = append(def.Sizes, ImageInstance{
					Zoom:        ff[2].(float64),
					File:        serverID[1:],
					IsLocalFile: false,
				})
			} else {
				def.Sizes = append(def.Sizes, ImageInstance{
					Zoom:        ff[2].(float64),
					File:        serverID,
					IsLocalFile: true,
				})
			}
			rawImages[ff[1].(string)] = def

		default: // <attr>:<id> <value>	-> rawData[<id>][<attr>] = []<value>
			attr, objID, ok := strings.Cut(f[0], ":")
			if !ok {
				return nil, meta, fmt.Errorf("legacy map file has improperly formed record (can't parse <attr>:<id> from \"%s\")", f[0])
			}
			if rawData[objID] == nil {
				rawData[objID] = make(map[string][]string)
			}
			rawData[objID][attr] = f[1:]
		}
	}

	// Now assemble the collected data into a collection of MapObjects
	writeCreature := func(ct string, mob map[string][]string, m *CreatureToken, objID string) error {
		if t, ok := mob["TYPE"]; !ok || len(t) < 1 || t[0] != ct {
			return fmt.Errorf("legacy file %s %s has missing or invalid TYPE", ct, objID)
		}
		m.Name, err = objString(mob, 0, "NAME", true, err)
		if healthStruct, ok := mob["HEALTH"]; ok {
			ss, err := tcllist.ParseTclList(healthStruct[0])
			if err != nil {
				return fmt.Errorf("legacy file %s %s has invalid HEALTH: %v", ct, objID, err)
			}
			if len(ss) > 0 {
				hob := map[string][]string{
					"HEALTH": ss,
				}
				maxhp, err := objInt(hob, 0, "HEALTH", true, err)
				lethal, err := objInt(hob, 1, "HEALTH", true, err)
				subdual, err := objInt(hob, 2, "HEALTH", true, err)
				con, err := objInt(hob, 3, "HEALTH", true, err)
				flat, err := objBool(hob, 4, "HEALTH", true, err)
				stab, err := objBool(hob, 5, "HEALTH", true, err)
				cond, err := objString(hob, 6, "HEALTH", false, err)
				blur, err := objInt(hob, 7, "HEALTH", false, err)

				m.Health = &CreatureHealth{
					MaxHP:           maxhp,
					LethalDamage:    lethal,
					NonLethalDamage: subdual,
					Con:             con,
					IsFlatFooted:    flat,
					IsStable:        stab,
					Condition:       cond,
					HPBlur:          blur,
				}
			}
		}
		m.Gx, err = objFloat(mob, 0, "GX", false, err)
		m.Gy, err = objFloat(mob, 0, "GY", false, err)
		m.Skin, err = objInt(mob, 0, "SKIN", false, err)
		m.SkinSize, err = objStrings(mob, 0, "SKINSIZE", false, err)
		m.Elev, err = objInt(mob, 0, "ELEV", false, err)
		m.Color, err = objString(mob, 0, "COLOR", false, err)
		m.Note, err = objString(mob, 0, "NOTE", false, err)
		m.Size, err = objString(mob, 0, "SIZE", false, err)
		m.StatusList, err = objStrings(mob, 0, "STATUSLIST", false, err)
		if aoeStruct, ok := mob["AOE"]; ok {
			ss, err := tcllist.ParseTclList(aoeStruct[0])
			if err != nil {
				return fmt.Errorf("legacy file %s %s has invalid AOE: %v", ct, objID, err)
			}
			if len(ss) > 0 {
				aob := map[string][]string{
					"AOE": ss,
				}

				atype, err := objString(aob, 0, "AOE", true, err)
				if atype != "radius" {
					err = fmt.Errorf("invalid AOE type \"%s\"", atype)
				}
				radius, err := objFloat(aob, 1, "AOE", true, err)
				color, err := objString(aob, 2, "AOE", true, err)

				m.AoE = &RadiusAoE{
					Radius: radius,
					Color:  color,
				}
			}
		}
		if s, ok := mob["MOVEMODE"]; ok {
			if len(s) > 0 {
				switch s[0] {
				case "fly":
					m.MoveMode = MoveModeFly
				case "climb":
					m.MoveMode = MoveModeClimb
				case "swim":
					m.MoveMode = MoveModeSwim
				case "burrow":
					m.MoveMode = MoveModeBurrow
				case "land", "":
					m.MoveMode = MoveModeLand
				default:
					return fmt.Errorf("legacy file %s %s has invalid MOVEMODE: unsupported mode \"%s\"", ct, objID, s)
				}
			}
		}
		m.Reach, err = objInt(mob, 0, "REACH", false, err)
		m.Killed, err = objBool(mob, 0, "KILLED", false, err)
		m.Dim, err = objBool(mob, 0, "DIM", false, err)
		objList = append(objList, *m)
		if err != nil {
			return fmt.Errorf("legacy file %s %s: %v", ct, objID, err)
		}
		return nil
	}

	for objID, mob := range rawMonsters {
		m := CreatureToken{
			BaseMapObject: BaseMapObject{
				ID: objID,
			},
			CreatureType: CreatureTypeMonster,
		}
		if err := writeCreature("monster", mob, &m, objID); err != nil {
			return nil, meta, err
		}
	}

	for objID, mob := range rawPlayers {
		m := CreatureToken{
			BaseMapObject: BaseMapObject{
				ID: objID,
			},
			CreatureType: CreatureTypePlayer,
		}
		if err := writeCreature("player", mob, &m, objID); err != nil {
			return nil, meta, err
		}
	}

	for objID, obj := range rawData {
		var objType string

		objType, err := objString(obj, 0, "TYPE", true, err)
		base := MapElement{
			BaseMapObject: BaseMapObject{
				ID: objID,
			},
		}
		base.X, err = objFloat(obj, 0, "X", false, err)
		base.Y, err = objFloat(obj, 0, "Y", false, err)
		base.Z, err = objInt(obj, 0, "Z", false, err)
		base.Points, err = objCoordinateList(obj, 0, "POINTS", false, err)
		base.Line, err = objString(obj, 0, "LINE", false, err)
		base.Fill, err = objString(obj, 0, "FILL", false, err)
		base.Width, err = objInt(obj, 0, "WIDTH", false, err)
		base.Layer, err = objString(obj, 0, "LAYER", false, err)
		base.Level, err = objInt(obj, 0, "LEVEL", false, err)
		base.Group, err = objString(obj, 0, "GROUP", false, err)
		if s, ok := obj["DASH"]; ok {
			if len(s) > 0 {
				switch s[0] {
				case "":
					base.Dash = DashSolid
				case "-":
					base.Dash = DashLong
				case ",":
					base.Dash = DashMedium
				case ".":
					base.Dash = DashShort
				case "-.":
					base.Dash = DashLongShort
				case "-..":
					base.Dash = DashLong2Short
				default:
					return nil, meta, fmt.Errorf("legacy file element %s has invalid DASH: %s", objID, s)
				}
			}
		}
		base.Hidden, err = objBool(obj, 0, "HIDDEN", false, err)
		base.Locked, err = objBool(obj, 0, "LOCKED", false, err)

		switch objType {
		case "aoe", "saoe":
			o := SpellAreaOfEffectElement{
				MapElement: base,
			}
			if s, ok := obj["AOESHAPE"]; ok {
				if len(s) > 0 {
					switch s[0] {
					case "cone":
						o.AoEShape = AoEShapeCone
					case "radius":
						o.AoEShape = AoEShapeRadius
					case "ray":
						o.AoEShape = AoEShapeRay
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid AOESHAPE: \"%s\"", objID, s)
					}
				}
			} else {
				return nil, meta, fmt.Errorf("legacy file element %s has missing AOESHAPE", objID)
			}

			objList = append(objList, o)

		case "arc":
			o := ArcElement{
				MapElement: base,
			}
			if s, ok := obj["ARCMODE"]; ok {
				if len(s) > 0 {
					switch s[0] {
					case "pieslice":
						o.ArcMode = ArcModePieSlice
					case "arc":
						o.ArcMode = ArcModeArc
					case "chord":
						o.ArcMode = ArcModeChord
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid ARCMODE: \"%s\"", objID, s)
					}
				}
			}
			o.Start, err = objFloat(obj, 0, "START", true, err)
			o.Extent, err = objFloat(obj, 0, "EXTENT", true, err)
			objList = append(objList, o)

		case "circ":
			o := CircleElement{
				MapElement: base,
			}
			objList = append(objList, o)

		case "group", "layer", "token":
			return nil, meta, fmt.Errorf("legacy file element %s has invalid TYPE \"%s\" (this was never implemented for legacy files)", objID, objType)

		case "line":
			o := LineElement{
				MapElement: base,
			}
			if s, ok := obj["ARROW"]; ok {
				if len(s) > 0 {
					switch s[0] {
					case "none", "":
						o.Arrow = ArrowNone
					case "first":
						o.Arrow = ArrowFirst
					case "last":
						o.Arrow = ArrowLast
					case "both":
						o.Arrow = ArrowBoth
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid ARROW: \"%s\"", objID, s)
					}
				}
			}
			objList = append(objList, o)

		case "poly":
			o := PolygonElement{
				MapElement: base,
			}
			o.Spline, err = objFloat(obj, 0, "SPLINE", false, err)
			if s, ok := obj["JOIN"]; ok {
				if len(s) > 0 {
					switch s[0] {
					case "bevel":
						o.Join = JoinBevel
					case "miter":
						o.Join = JoinMiter
					case "round":
						o.Join = JoinRound
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid JOIN: \"%s\"", objID, s)
					}
				}
			}
			objList = append(objList, o)

		case "rect":
			o := RectangleElement{
				MapElement: base,
			}
			objList = append(objList, o)

		case "text":
			if _, ok := obj["TEXT"]; !ok {
				return nil, meta, fmt.Errorf("legacy file element %s missing TEXT", objID)
			}
			o := TextElement{
				MapElement: base,
			}

			o.Text, err = objString(obj, 0, "TEXT", true, err)

			var fontStruct string
			fontStruct, err = objString(obj, 0, "FONT", false, err)
			if fontStruct != "" {
				ss, err := tcllist.ParseTclList(fontStruct)
				if err != nil {
					return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT (1st level): %v", objID, err)
				}
				if len(ss) > 0 {
					ss, err = tcllist.ParseTclList(ss[0])
					if err != nil {
						return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT (2nd level): %v", objID, err)
					}
					if len(ss) < 2 {
						return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT: len=%d", objID, len(ss))
					}
					if len(ss) < 3 {
						ss = append(ss, "normal")
					}
					if len(ss) < 4 {
						ss = append(ss, "roman")
					}
					o.Font = TextFont{
						Family: ss[0],
					}
					o.Font.Size, err = strconv.ParseFloat(ss[1], 64)
					if err != nil {
						return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT size: %v", objID, err)
					}
					switch ss[2] {
					case "normal":
						o.Font.Weight = FontWeightNormal
					case "bold":
						o.Font.Weight = FontWeightBold
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT weight \"%s\"", objID, ss[2])
					}
					switch ss[3] {
					case "roman", "Roman":
						o.Font.Slant = FontSlantRoman
					case "italic", "Italic":
						o.Font.Slant = FontSlantItalic
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid FONT slant \"%s\"", objID, ss[3])
					}
				}
			} else {
				return nil, meta, fmt.Errorf("legacy file element %s missing FONT", objID)
			}
			if s, ok := obj["ANCHOR"]; ok {
				if len(s) > 0 {
					switch s[0] {
					case "center":
						o.Anchor = AnchorCenter
					case "n":
						o.Anchor = AnchorNorth
					case "s":
						o.Anchor = AnchorSouth
					case "e":
						o.Anchor = AnchorEast
					case "w":
						o.Anchor = AnchorWest
					case "ne":
						o.Anchor = AnchorNE
					case "nw":
						o.Anchor = AnchorNW
					case "se":
						o.Anchor = AnchorSE
					case "sw":
						o.Anchor = AnchorSW
					default:
						return nil, meta, fmt.Errorf("legacy file element %s has invalid missing ANCHOR \"%s\"", objID, s)
					}
				}
			}
			objList = append(objList, o)

		case "tile":
			o := TileElement{
				MapElement: base,
			}
			o.Image, err = objString(obj, 0, "IMAGE", true, err)
			o.BBHeight, err = objFloat(obj, 0, "BBHEIGHT", false, err)
			o.BBWidth, err = objFloat(obj, 0, "BBWIDTH", false, err)
			objList = append(objList, o)

		default:
			return nil, meta, fmt.Errorf("legacy file element %s has invalid TYPE \"%s\"", objID, objType)
		}
		if err != nil {
			return nil, meta, fmt.Errorf("legacy file object %s: %v", objID, err)
		}
	}

	for _, f := range rawFiles {
		objList = append(objList, FileDefinition{
			File:        f,
			IsLocalFile: false,
		})
	}

	for _, i := range rawImages {
		objList = append(objList, i)
	}

	return objList, meta, nil
}

//
// LoadMapMetaData is just like LoadMapFile but only reads enough of the
// file to get the meta data, which is returned.
//
func LoadMapMetaData(input io.Reader) (MapMetaData, error) {
	_, meta, err := loadMapFile(input, true)
	return meta, err
}

//
// LoadMapFile reads GMA mapper data from an already-open data stream,
// returning a slice of map elements and the metadata read from the stream.
//
// If called with a nil input object, it just returns with empty data values.
//
func LoadMapFile(input io.Reader) ([]any, MapMetaData, error) {
	return loadMapFile(input, false)
}

func loadMapFile(input io.Reader, metaDataOnly bool) ([]any, MapMetaData, error) {
	//
	// The map file format consists of an initial line which begins with
	//    __MAPPER__:<version>
	// (there may be additional text following <version> if it is
	// separated from <version> by a space; older file formats used this
	// for metadata but newer versions ignore the extra text if present).
	//
	// For versions >= 20, this is followed by zero or more object
	// definitions which are of the form
	//    «<type>» <json>
	// where <json> may be a multi-line structure. The start of each
	// new object or the EOF marker is indicated by encountering a line
	// which begins with a '«' rune, so the <json> data must not end up
	// triggering this condition. This may be accomplished by placing the
	// entire JSON string on one line, or indenting it on subsequent lines.
	//
	// The final line is of the form
	//    «__EOF__»
	//
	var meta MapMetaData
	var objList []any
	var f []string
	var v uint64
	var err error

	if input == nil {
		return nil, meta, nil
	}

	startPattern := regexp.MustCompile("^__MAPPER__:(\\d+)\\s*(.*)$")
	recordPattern := regexp.MustCompile("^«(.*?)» (.+)$")
	eofPattern := regexp.MustCompile("^«__EOF__»$")
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		// no data
		return nil, meta, nil
	}

	if f = startPattern.FindStringSubmatch(scanner.Text()); f == nil {
		return nil, meta, fmt.Errorf("invalid map file format in initial header")
	}
	if v, err = strconv.ParseUint(f[1], 10, 64); err != nil {
		return nil, meta, fmt.Errorf("invalid map file format: can't parse version \"%v\": %v", f[1], err)
	}
	meta.FileVersion = uint(v)
	if v < MinimumSupportedMapFileFormat || v > MaximumSupportedMapFileFormat {
		if MinimumSupportedMapFileFormat == MaximumSupportedMapFileFormat {
			return nil, meta, fmt.Errorf("cannot read map file format version %d (only version %d is supported)", v, MinimumSupportedMapFileFormat)
		}
		return nil, meta, fmt.Errorf("cannot read map file format version %d (only versions %d-%d are supported)", v, MinimumSupportedMapFileFormat, MaximumSupportedMapFileFormat)
	}
	if v < 20 {
		return loadLegacyMapFile(scanner, meta, f[2], metaDataOnly)
	}

	for scanner.Scan() {
	rescan:
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		if eofPattern.MatchString(scanner.Text()) {
			return objList, meta, nil
		}
		if f = recordPattern.FindStringSubmatch(scanner.Text()); f == nil {
			return nil, meta, fmt.Errorf("invalid map file format: unexpected data \"%v\"", scanner.Text())
		}
		// Start of record type f[1] with start of JSON string f[2]
		// collect more lines of JSON data...
		var dataPacket strings.Builder
		dataPacket.WriteString(f[2])

		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "«") {
				var err error

				switch f[1] {
				case "__META__":
					err = json.Unmarshal([]byte(dataPacket.String()), &meta)
					if metaDataOnly {
						return nil, meta, err
					}

				case "ARC":
					var arc ArcElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &arc); err == nil {
						objList = append(objList, arc)
					}

				case "CIRC":
					var circle CircleElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &circle); err == nil {
						objList = append(objList, circle)
					}

				case "LINE":
					var line LineElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &line); err == nil {
						objList = append(objList, line)
					}

				case "POLY":
					var poly PolygonElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &poly); err == nil {
						objList = append(objList, poly)
					}

				case "RECT":
					var rect RectangleElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &rect); err == nil {
						objList = append(objList, rect)
					}

				case "SAOE":
					var effect SpellAreaOfEffectElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &effect); err == nil {
						objList = append(objList, effect)
					}

				case "TEXT":
					var text TextElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &text); err == nil {
						objList = append(objList, text)
					}

				case "TILE":
					var tile TileElement
					if err = json.Unmarshal([]byte(dataPacket.String()), &tile); err == nil {
						objList = append(objList, tile)
					}

				case "IMG":
					var img ImageDefinition
					if err = json.Unmarshal([]byte(dataPacket.String()), &img); err == nil {
						objList = append(objList, img)
					}

				case "MAP":
					var file FileDefinition
					if err = json.Unmarshal([]byte(dataPacket.String()), &file); err == nil {
						objList = append(objList, file)
					}

				case "CREATURE":
					var mob CreatureToken
					if err = json.Unmarshal([]byte(dataPacket.String()), &mob); err == nil {
						objList = append(objList, mob)
					}

				default:
					return nil, meta, fmt.Errorf("invalid map file format: unexpected record type \"%s\"", f[1])
				}
				if err != nil {
					return nil, meta, fmt.Errorf("invalid map file format: %v", err)
				}
				goto rescan
			}
			dataPacket.WriteString(scanner.Text())
		}
	}
	return nil, meta, fmt.Errorf("invalid map file format: unexpected end of file")
}

// @[00]@| Go-GMA 5.21.2
// @[01]@|
// @[10]@| Overall GMA package Copyright © 1992–2024 by Steven L. Willoughby (AKA MadScienceZone)
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
