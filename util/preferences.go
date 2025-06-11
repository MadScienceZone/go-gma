/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   _____      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) / ___ \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |( (___) )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   ) \     /    | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /  / ___ \    | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/  ( (   ) )   |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\( (___) ) _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/ \_____/ (_)(_______)     #
#                                                                                      #
########################################################################################
*/

package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	GMAMapperPreferencesMinimumVersion int = 1
	GMAMapperPreferencesMaximumVersion int = 9
	GMAPreferencesMinimumVersion       int = 1
	GMAPreferencesMaximumVersion       int = 2
)

type UnsupportedPreferencesVersionError struct {
	MinimumVersion int
	MaximumVersion int
	DataVersion    int
}

func (e UnsupportedPreferencesVersionError) Error() string {
	if e.MinimumVersion == e.MaximumVersion {
		return fmt.Sprintf("preferences data version %d is not the supported version (%d)", e.DataVersion, e.MaximumVersion)
	}
	return fmt.Sprintf("preferences data version %d is not a supported version (%d-%d)", e.DataVersion, e.MinimumVersion, e.MaximumVersion)
}

//
// GridOffsets provide x and y offsets for grid guides
//
type GridOffsets struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

//
// GridGuide describes extra grid guidelines
//
type GridGuide struct {
	Interval int         `json:"interval,omitempty"`
	Offsets  GridOffsets `json:"offsets,omitempty"`
}

//
// TimerVisibility represents which timers a user wishes to see on their client's display.
//
type TimerVisibility byte

const (
	ShowNoTimers TimerVisibility = iota
	ShowMyTimers
	ShowAllTimers
)

func (tv *TimerVisibility) MarshalJSON() ([]byte, error) {
	if tv != nil {
		switch *tv {
		case ShowNoTimers:
			return json.Marshal("none")
		case ShowAllTimers:
			return json.Marshal("all")
		}
	}
	return json.Marshal("mine")
}

func (tv *TimerVisibility) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "none":
		*tv = ShowNoTimers
	case "all":
		*tv = ShowAllTimers
	default:
		*tv = ShowMyTimers
	}
	return nil
}

//
// ButtonSize represents the valid sizes for buttons to be.
//
type ButtonSize byte

const (
	SmallButtons ButtonSize = iota
	MediumButtons
	LargeButtons
)

func (bs *ButtonSize) MarshalJSON() ([]byte, error) {
	if bs != nil {
		switch *bs {
		case MediumButtons:
			return json.Marshal("medium")
		case LargeButtons:
			return json.Marshal("large")
		}
	}
	return json.Marshal("small")
}

func (bs *ButtonSize) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "medium":
		*bs = MediumButtons
	case "large":
		*bs = LargeButtons
	default:
		*bs = SmallButtons
	}
	return nil
}

//
// ImageType represents the valid bitmap types supported by the mapper.
//
type ImageType byte

const (
	PNG ImageType = iota
	GIF
)

func (i *ImageType) MarshalJSON() ([]byte, error) {
	if i != nil && *i == GIF {
		return json.Marshal("gif")
	}
	return json.Marshal("png")
}

func (i *ImageType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "gif" || s == "GIF" {
		*i = GIF
	} else {
		*i = PNG
	}
	return nil
}

//
// ServerProfile describes each set of preferences associated with a given
// server as opposed to global settings which apply regardless of server.
//
type ServerProfile struct {
	Name         string `json:"name"`
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	UserName     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	CurlProxy    string `json:"curl_proxy,omitempty"`
	BlurAll      bool   `json:"blur_all,omitempty"`
	BlurPct      int    `json:"blur_pct,omitempty"`
	SuppressChat bool   `json:"suppress_chat,omitempty"`
	ChatLimit    int    `json:"chat_limit,omitempty"`
	ChatLog      string `json:"chat_log,omitempty"`
	CurlServer   string `json:"curl_server,omitempty"`
	UpdateURL    string `json:"update_url,omitempty"`
	ModuleID     string `json:"module_id,omitempty"`
	ServerMkdir  string `json:"server_mkdir,omitempty"`
	NcPath       string `json:"nc_path,omitempty"`
	ScpPath      string `json:"scp_path,omitempty"`
	ScpDest      string `json:"scp_dest,omitempty"`
	ScpServer    string `json:"scp_server,omitempty"`
	ScpProxy     string `json:"scp_proxy,omitempty"`
	SshPath      string `json:"ssh_path,omitempty"`
}

//
// FontWeight is the set of valid font weight values
//
type FontWeight byte

const (
	Regular FontWeight = iota
	Bold
)

//
// FontSlant is the set of valid font slant values
//
type FontSlant byte

const (
	Roman FontSlant = iota
	Italic
)

//
// UserFont describes a user-defined font.
//
type UserFont struct {
	Family     string     `json:"family,omitempty"`
	Size       float64    `json:"size,omitempty"`
	Weight     FontWeight `json:"weight,omitempty"`
	Slant      FontSlant  `json:"slant,omitempty"`
	Overstrike bool       `json:"overstrike,omitempty"`
	Underline  string     `json:"underline,omitempty"`
}

//
// ColorSet encapsulates the colors to use in light and dark mode.
//
type ColorSet struct {
	Dark  string `json:"dark,omitempty"`
	Light string `json:"light,omitempty"`
}

const (
	DefaultFGColorDark      = "#aaaaaa"
	DefaultFGColorLight     = "#000000"
	DefaultBGColorDark      = "#232323"
	DefaultBGColorLight     = "#cccccc"
	DefaultCheckSelectDark  = "#ffffff"
	DefaultCheckSelectLight = "#000000"
	DefaultCheckMenuDark    = "#ffffff"
	DefaultCheckMenuLight   = "#000000"
	DefaultBrightFGDark     = "#ffffff"
	DefaultBrightFGLight    = "#000000"
	DefaultGridDark         = "#aaaaaa"
	DefaultGridLight        = "blue"
	DefaultGridMajorDark    = "#345f12"
	DefaultGridMajorLight   = "#345f12"
	DefaultGridMinorDark    = "#b00b03"
	DefaultGridMinorLight   = "#b00b03"
	DefaultHandColorLight   = "black"
	DefaultHandColorDark    = "#aaaaaa"
	DefaultTickColorLight   = "blue"
	DefaultTickColorDark    = "#aaaaaa"
	DefaultFListFGLight     = "black"
	DefaultFListFGDark      = "white"
	DefaultFListBGLight     = "white"
	DefaultFListBGDark      = "#232323"
	DefaultNextFGLight      = "white"
	DefaultNextFGDark       = "white"
	DefaultNextBGLight      = "black"
	DefaultNextBGDark       = "#cc0000"
	DefaultCurBGLight       = "#9cffb4"
	DefaultCurBGDark        = "#003300"
	DefaultReadyBGLight     = "#ff3333"
	DefaultReadyBGDark      = "#ff0000"
	DefaultHoldBGLight      = "#ffaaaa"
	DefaultHoldBGDark       = "#610400"
	DefaultZeroHPLight      = "#ff0000"
	DefaultZeroHPDark       = "#ff0000"
	DefaultNegativeHPLight  = "#000000"
	DefaultNegativeHPDark   = "#000000"
	DefaultSlotFGLight      = "#000000"
	DefaultSlotFGDark       = "#666666"
	DefaultSlotBGLight      = "#666666"
	DefaultSlotBGDark       = "#232323"
	DefaultFlatFootedLight  = "#3333ff"
	DefaultFlatFootedDark   = "#3333ff"
	DefaultPresetNameLight  = "blue"
	DefaultPresetNameDark   = "cyan"
)

//
// DieRollComponent describes the settings for a specific chat or die-roll style
// component.
//
type DieRollComponent struct {
	FG             ColorSet `json:"fg,omitempty"`
	BG             ColorSet `json:"bg,omitempty"`
	FontName       string   `json:"font,omitempty"`
	Format         string   `json:"format,omitempty"`
	Overstrike     bool     `json:"overstrike,omitempty"`
	Underline      bool     `json:"underline,omitempty"`
	BaselineOffset int      `json:"offset,omitempty"`
}

//
// UserPreferences represents the preferences settings for the GMA Mapper.
//
// This represents preferences version 4.
//
type UserPreferences struct {
	GMAMapperPreferencesVersion int        `json:"GMA_Mapper_preferences_version"`
	Animate                     bool       `json:"animate,omitempty"`
	ButtonSize                  ButtonSize `json:"button_size,omitempty"`
	ChatTimestamp               bool       `json:"chat_timestamp,omitempty"`
	ColorizeDieRolls            bool       `json:"colorize_die_rolls,omitempty"`
	CurlPath                    string     `json:"curl_path,omitempty"`
	CurlInsecure                bool       `json:"curl_insecure,omitempty"`
	CurrentProfile              string     `json:"current_profile,omitempty"`
	DarkMode                    bool       `json:"dark,omitempty"`
	DebugLevel                  int        `json:"debug_level,omitempty"`
	DebugProtocol               bool       `json:"debug_proto,omitempty"`
	FlashUpdates                bool       `json:"flash_updates,omitempty"`
	GuideLines                  struct {
		Major GridGuide `json:"major,omitempty"`
		Minor GridGuide `json:"minor,omitempty"`
	} `json:"guide_lines,omitempty"`
	ImageFormat   ImageType           `json:"image_format,omitempty"`
	KeepTools     bool                `json:"keep_tools,omitempty"`
	MarkupEnabled bool                `json:"markup_enabled,omitempty"`
	MenuButton    bool                `json:"menu_button,omitempty"`
	NeverAnimate  bool                `json:"never_animate,omitempty"`
	PreloadImages bool                `json:"preload,omitempty"`
	Profiles      []ServerProfile     `json:"profiles,omitempty"`
	Fonts         map[string]UserFont `json:"fonts,omitempty"`
	Scaling       float64             `json:"scaling,omitempty"`
	ShowTimers    TimerVisibility     `json:"show_timers,omitempty"`
	Styles        StyleDescription    `json:"styles,omitempty"`
}

//
// InitiativeSeedData represents each entry in the GMA initiative seed list.
//
type InitiativeSeedData struct {
	Name         string `json:"name"`
	Dexterity    int    `json:"dex"`
	Constitution int    `json:"con"`
	InitAdj      int    `json:"init_adj"`
	HP           int    `json:"hp"`
	BlurHP       int    `json:"blur_hp"`
	IsPC         bool   `json:"is_pc"`
	DieSpec      string `json:"die_spec"`
}

//
// CasterData represents each entry in the GMA caster level list.
//
type CasterData struct {
	Name string `json:"name"`
	CL   int    `json:"cl"`
}

//
// GMAPreferences represents the preferences settings for GMA and its supporting utilities.
//
type GMAPreferences struct {
	GMAPreferencesVersion int    `json:"gma_preferences_file_version"`
	CoreDBPath            string `json:"core_db"`
	Appearance            struct {
		DarkMode bool    `json:"dark_mode"`
		Scaling  float64 `json:"scaling"`
	} `json:"appearance"`
	Worlds             map[string]GMAWorld          `json:"worlds"`
	Networks           map[string]GMANetworkProfile `json:"networks"`
	CurrentWorldName   string                       `json:"current_world"`
	CurrentNetworkName string                       `json:"network_profile"`
}

//
// GMAWorld represents the preferences settings for each campaign world.
//
type GMAWorld struct {
	CalendarType         string               `json:"calendar_type"`
	ModulePath           string               `json:"module,omitempty"`
	BlurHP               int                  `json:"blur_hp,omitempty"`
	DBName               string               `json:"db_name"`
	DisplayName          string               `json:"display_name"`
	InitiativeBackupPath string               `json:"initiative_backup_path,omitempty"`
	Password             string               `json:"password,omitempty"`
	InitiativeSeed       []InitiativeSeedData `json:"initiative_seed"`
	CasterLevels         []CasterData         `json:"caster_levels"`
}

//
// GMANetworkProfile represents the preferences settings for each network profile.
//
type GMANetworkProfile struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int    `json:"port,omitempty"`
}

//
// StyleDescription describes the different kinds of style settings.
//
type StyleDescription struct {
	Clocks   ClockStyles   `json:"clocks,omitempty"`
	Dialogs  DialogStyles  `json:"dialogs,omitempty"`
	DieRolls DieRollStyles `json:"dierolls,omitempty"`
}

type ClockStyles struct {
	HandColor       ColorSet `json:"hand_color,omitempty"`
	TickColor       ColorSet `json:"tick_color,omitempty"`
	FListFGColor    ColorSet `json:"flist_fg,omitempty"`
	FListBGColor    ColorSet `json:"flist_bg,omitempty"`
	NextFGColor     ColorSet `json:"next_fg,omitempty"`
	NextBGColor     ColorSet `json:"next_bg,omitempty"`
	CurrentBGColor  ColorSet `json:"cur_bg,omitempty"`
	ReadyBGColor    ColorSet `json:"ready_bg,omitempty"`
	HoldBGColor     ColorSet `json:"hold_bg,omitempty"`
	ZeroHPColor     ColorSet `json:"zero_hp,omitempty"`
	NegativeHPColor ColorSet `json:"negative_hp,omitempty"`
	SlotFGColor     ColorSet `json:"slot_fg,omitempty"`
	SlotBGColor     ColorSet `json:"slot_bg,omitempty"`
	FlatFootedColor ColorSet `json:"flat_footed,omitempty"`
	TimeDispFont    string   `json:"timedisp_font,omitempty"`
	TurnDispFont    string   `json:"turndisp_font,omitempty"`
	DefaultFont     string   `json:"default_font,omitempty"`
}

type DieRollStyles struct {
	CompactRecents bool                        `json:"compact_recents,omitempty"`
	Components     map[string]DieRollComponent `json:"components,omitempty"`
}

type DialogStyles struct {
	HeadingFG        ColorSet `json:"heading_fg,omitempty"`
	NormalFG         ColorSet `json:"normal_fg,omitempty"`
	NormalBG         ColorSet `json:"normal_bg,omitempty"`
	HighlightFG      ColorSet `json:"highlight_fg,omitempty"`
	OddRowBG         ColorSet `json:"odd_bg,omitempty"`
	EvenRowBG        ColorSet `json:"even_bg,omitempty"`
	CheckSelectColor ColorSet `json:"check_select,omitempty"`
	CheckMenuColor   ColorSet `json:"check_menu,omitempty"`
	BrightFG         ColorSet `json:"bright_fg,omitempty"`
	GridColor        ColorSet `json:"grid,omitempty"`
	MinorGridColor   ColorSet `json:"grid_minor,omitempty"`
	MajorGridColor   ColorSet `json:"grid_major,omitempty"`
	PresetNameColor  ColorSet `json:"preset_name,omitempty"`
}

//
// DefaultGMAPreferences returns a GMAPreferences value with default values.
//
func DefaultGMAPreferences() (GMAPreferences, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return GMAPreferences{}, err
	}

	return GMAPreferences{
		CoreDBPath: filepath.Join(homeDir, ".gma", "core.db"),
		Networks: map[string]GMANetworkProfile{
			"offline": GMANetworkProfile{},
		},
		CurrentNetworkName: "offline",
	}, nil
}

//
// DefaultPreferences returns a UserPreferences list with a
// reasonable set of default values.
//
func DefaultPreferences() UserPreferences {

	curlPath, err := SearchInPath("curl")
	if err != nil {
		curlPath = ""
	}
	ncPath, err := SearchInPath("nc")
	if err != nil {
		ncPath = ""
	}
	scpPath, err := SearchInPath("scp")
	if err != nil {
		scpPath = ""
	}
	sshPath, err := SearchInPath("ssh")
	if err != nil {
		sshPath = ""
	}

	return UserPreferences{
		ButtonSize:       SmallButtons,
		ChatTimestamp:    true,
		ColorizeDieRolls: true,
		CurlPath:         curlPath,
		CurrentProfile:   "offline",
		ImageFormat:      PNG,
		ShowTimers:       ShowMyTimers,
		Profiles: []ServerProfile{
			ServerProfile{
				Name:    "offline",
				Port:    2323,
				NcPath:  ncPath,
				ScpPath: scpPath,
				SshPath: sshPath,
			},
		},
		Fonts: map[string]UserFont{
			"FullResult": UserFont{
				Family: "Helvetica",
				Size:   16,
				Weight: Bold,
			},
			"Important": UserFont{
				Family: "Helvetica",
				Size:   12,
				Weight: Bold,
			},
			"Result": UserFont{
				Family: "Helvetica",
				Size:   14,
			},
			"Normal": UserFont{
				Family: "Helvetica",
				Size:   12,
			},
			"Special": UserFont{
				Family: "Times",
				Size:   12,
				Slant:  Italic,
			},
			"System": UserFont{
				Family: "Times",
				Size:   10,
				Slant:  Italic,
			},
			"Tiny": UserFont{
				Family: "Helvetica",
				Size:   8,
			},
			"ClockTime": UserFont{
				Family: "Helvetica",
				Size:   16,
			},
			"ClockList": UserFont{
				Family: "Helvetica",
				Size:   24,
			},
			"Bold": UserFont{
				Family: "Helvetica",
				Size:   12,
				Weight: Bold,
			},
			"Italic": UserFont{
				Family: "Helvetica",
				Size:   12,
				Slant:  Italic,
			},
			"BoldItalic": UserFont{
				Family: "Helvetica",
				Size:   12,
				Slant:  Italic,
				Weight: Bold,
			},
			"Title": UserFont{
				Family: "Helvetica",
				Size:   16,
				Weight: Bold,
			},
			"Subtitle": UserFont{
				Family: "Helvetica",
				Size:   14,
				Weight: Bold,
			},
		},
		Styles: StyleDescription{
			Clocks: ClockStyles{
				HandColor:       ColorSet{Dark: DefaultHandColorDark, Light: DefaultHandColorLight},
				TickColor:       ColorSet{Dark: DefaultTickColorDark, Light: DefaultTickColorLight},
				FListFGColor:    ColorSet{Dark: DefaultFListFGDark, Light: DefaultFListFGLight},
				FListBGColor:    ColorSet{Dark: DefaultFListBGDark, Light: DefaultFListBGLight},
				NextFGColor:     ColorSet{Dark: DefaultNextFGDark, Light: DefaultNextFGLight},
				NextBGColor:     ColorSet{Dark: DefaultNextBGDark, Light: DefaultNextBGLight},
				CurrentBGColor:  ColorSet{Dark: DefaultCurBGDark, Light: DefaultCurBGLight},
				ReadyBGColor:    ColorSet{Dark: DefaultReadyBGDark, Light: DefaultReadyBGLight},
				HoldBGColor:     ColorSet{Dark: DefaultHoldBGDark, Light: DefaultHoldBGLight},
				ZeroHPColor:     ColorSet{Dark: DefaultZeroHPDark, Light: DefaultZeroHPLight},
				NegativeHPColor: ColorSet{Dark: DefaultNegativeHPDark, Light: DefaultNegativeHPLight},
				SlotFGColor:     ColorSet{Dark: DefaultSlotFGDark, Light: DefaultSlotFGLight},
				SlotBGColor:     ColorSet{Dark: DefaultSlotBGDark, Light: DefaultSlotBGLight},
				FlatFootedColor: ColorSet{Dark: DefaultFlatFootedDark, Light: DefaultFlatFootedLight},
				TimeDispFont:    "ClockTime",
				TurnDispFont:    "ClockTime",
				DefaultFont:     "ClockList",
			},
			Dialogs: DialogStyles{
				HeadingFG:        ColorSet{Dark: "cyan", Light: "blue"},
				NormalFG:         ColorSet{Dark: DefaultFGColorDark, Light: DefaultFGColorLight},
				NormalBG:         ColorSet{Dark: DefaultBGColorDark, Light: DefaultBGColorLight},
				HighlightFG:      ColorSet{Dark: "yellow", Light: "red"},
				OddRowBG:         ColorSet{Dark: DefaultBGColorDark, Light: DefaultBGColorLight},
				EvenRowBG:        ColorSet{Dark: "blue", Light: "#bbbbff"},
				CheckSelectColor: ColorSet{Dark: DefaultCheckSelectDark, Light: DefaultCheckSelectLight},
				CheckMenuColor:   ColorSet{Dark: DefaultCheckMenuDark, Light: DefaultCheckMenuLight},
				BrightFG:         ColorSet{Dark: DefaultBrightFGDark, Light: DefaultBrightFGLight},
				GridColor:        ColorSet{Dark: DefaultGridDark, Light: DefaultGridLight},
				MinorGridColor:   ColorSet{Dark: DefaultGridMinorDark, Light: DefaultGridMinorLight},
				MajorGridColor:   ColorSet{Dark: DefaultGridMajorDark, Light: DefaultGridMajorLight},
				PresetNameColor:  ColorSet{Dark: DefaultPresetNameDark, Light: DefaultPresetNameLight},
			},
			DieRolls: DieRollStyles{
				CompactRecents: false,
				Components: map[string]DieRollComponent{
					"begingroup": DieRollComponent{
						FontName: "Normal",
					},
					"best": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   " best of %s",
					},
					"bold": DieRollComponent{
						FontName: "Bold",
					},
					"bolditalic": DieRollComponent{
						FontName: "BoldItalic",
					},
					"bonus": DieRollComponent{
						FG:       ColorSet{Dark: "#fffb00", Light: "#f05b00"},
						FontName: "Normal",
					},
					"constant": DieRollComponent{
						FontName: "Normal",
					},
					"critlabel": DieRollComponent{
						FG:       ColorSet{Dark: "#fffb00", Light: "#f05b00"},
						FontName: "Special",
						Format:   "Confirm: ",
					},
					"critspec": DieRollComponent{
						FG:       ColorSet{Dark: "#fffb00", Light: "#f05b00"},
						FontName: "Special",
					},
					"dc": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "DC %s: ",
					},
					"diebonus": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Special",
						Format:   "(%s per die)",
					},
					"diespec": DieRollComponent{
						FontName: "Normal",
					},
					"discarded": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Normal",
						Format:   "{%s}",
					},
					"error": DieRollComponent{
						FG:         ColorSet{Dark: "red", Light: "red"},
						FontName:   "Normal",
						Format:     "(%s per die)",
						Overstrike: true,
					},
					"exceeded": DieRollComponent{
						FG:       ColorSet{Dark: "#00fa92", Light: "green"},
						FontName: "Special",
						Format:   " exceeded DC by %s",
					},
					"fail": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Important",
						Format:   "(%s)",
					},
					"from": DieRollComponent{
						FG:       ColorSet{Dark: "cyan", Light: "blue"},
						FontName: "Normal",
					},
					"fullmax": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Important",
						Format:   "maximized",
					},
					"fullresult": DieRollComponent{
						FG:       ColorSet{Dark: "blue", Light: "white"},
						BG:       ColorSet{Dark: "white", Light: "blue"},
						FontName: "FullResult",
					},
					"italic": DieRollComponent{
						FontName: "Italic",
					},
					"iteration": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   " (roll #%s)",
					},
					"label": DieRollComponent{
						FG:       ColorSet{Dark: "cyan", Light: "blue"},
						FontName: "Special",
						Format:   " %s",
					},
					"max": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "max %s",
					},
					"maximized": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Important",
						Format:   ">",
					},
					"maxroll": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Important",
						Format:   "{%s}",
					},
					"met": DieRollComponent{
						FG:       ColorSet{Dark: "#00fa92", Light: "green"},
						FontName: "Special",
						Format:   "successful",
					},
					"min": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "min %s",
					},
					"moddelim": DieRollComponent{
						FG:       ColorSet{Dark: "#fffb00", Light: "#f05b00"},
						FontName: "Normal",
						Format:   " | ",
					},
					"normal": DieRollComponent{
						FontName: "Normal",
					},
					"notice": DieRollComponent{
						FG:       ColorSet{Dark: "yellow", Light: "red"},
						FontName: "Special",
						Format:   "[%s] ",
					},
					"operator": DieRollComponent{
						FontName: "Normal",
					},
					"repeat": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "repeat %s",
					},
					"result": DieRollComponent{
						FontName: "Result",
					},
					"roll": DieRollComponent{
						FG:       ColorSet{Dark: "#00fa92", Light: "green"},
						FontName: "Normal",
						Format:   "{%s}",
					},
					"section": DieRollComponent{
						FontName: "Title",
					},
					"separator": DieRollComponent{
						FontName: "Normal",
						Format:   "=",
					},
					"sf": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
					},
					"short": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Special",
						Format:   " missed DC by %s",
					},
					"subsection": DieRollComponent{
						FontName: "Subtitle",
					},
					"subtotal": DieRollComponent{
						FG:       ColorSet{Dark: "#00fa92", Light: "green"},
						FontName: "Normal",
						Format:   "(%s)",
					},
					"success": DieRollComponent{
						FG:       ColorSet{Dark: "#00fa92", Light: "green"},
						FontName: "Important",
						Format:   "(%s) ",
					},
					"system": DieRollComponent{
						FG:       ColorSet{Dark: "cyan", Light: "blue"},
						FontName: "System",
					},
					"title": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#ffffff"},
						BG:       ColorSet{Dark: "#000044", Light: "#c7c0ae"},
						FontName: "Normal",
					},
					"to": DieRollComponent{
						FG:       ColorSet{Dark: "red", Light: "red"},
						FontName: "Special",
					},
					"total": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "total %s",
					},
					"until": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   "until %s",
					},
					"worst": DieRollComponent{
						FG:       ColorSet{Dark: "#aaaaaa", Light: "#888888"},
						FontName: "Special",
						Format:   " worst of %s",
					},
				},
			},
		},
	}
}

//
// LoadGMAPreferencesWithDefaults is like LoadPreferencesWithDefaults but for
// the GMA preferences file instead of the mapper's.
//
func LoadGMAPreferencesWithDefaults(stream io.Reader) (GMAPreferences, error) {
	prefs, err := DefaultGMAPreferences()
	if err != nil {
		return prefs, err
	}
	if err := prefs.Update(stream); err != nil {
		return prefs, err
	}
	if prefs.GMAPreferencesVersion < GMAPreferencesMinimumVersion || prefs.GMAPreferencesVersion > GMAPreferencesMaximumVersion {
		return prefs, UnsupportedPreferencesVersionError{
			MinimumVersion: GMAPreferencesMinimumVersion,
			MaximumVersion: GMAPreferencesMaximumVersion,
			DataVersion:    prefs.GMAPreferencesVersion,
		}
	}
	return prefs, nil
}

//
// LoadPreferencesWithDefaults reads a set of saved preferences from an open file
// or other io.Reader object. It provides default values for fields not specified in
// the input data.
//
func LoadPreferencesWithDefaults(stream io.Reader) (UserPreferences, error) {
	prefs := DefaultPreferences()
	err := prefs.Update(stream)
	if err != nil {
		return prefs, err
	}
	if prefs.GMAMapperPreferencesVersion < GMAMapperPreferencesMinimumVersion || prefs.GMAMapperPreferencesVersion > GMAMapperPreferencesMaximumVersion {
		return prefs, UnsupportedPreferencesVersionError{
			MinimumVersion: GMAMapperPreferencesMinimumVersion,
			MaximumVersion: GMAMapperPreferencesMaximumVersion,
			DataVersion:    prefs.GMAMapperPreferencesVersion,
		}
	}
	return prefs, nil
}

//
// LoadPreferences reads a set of saved preferences from an io.Reader,
// returning a new UserPreferences value from that data. Any fields not specified
// in the input data will have zero values.
//
func LoadPreferences(stream io.Reader) (UserPreferences, error) {
	var prefs UserPreferences
	err := prefs.Update(stream)
	if err != nil {
		return prefs, err
	}
	if prefs.GMAMapperPreferencesVersion < GMAMapperPreferencesMinimumVersion || prefs.GMAMapperPreferencesVersion > GMAMapperPreferencesMaximumVersion {
		return prefs, UnsupportedPreferencesVersionError{
			MinimumVersion: GMAMapperPreferencesMinimumVersion,
			MaximumVersion: GMAMapperPreferencesMaximumVersion,
			DataVersion:    prefs.GMAMapperPreferencesVersion,
		}
	}
	return prefs, nil
}

//
// Update reads a set of saved preferences as LoadPreferences does, but
// rather than returning a new UserPreferences value, it updates the
// values of an existing UserPreferences value with the input data.
//
func (prefs *UserPreferences) Update(stream io.Reader) error {
	err := json.NewDecoder(stream).Decode(prefs)
	if err != nil {
		return err
	}
	if prefs.GMAMapperPreferencesVersion < GMAMapperPreferencesMinimumVersion || prefs.GMAMapperPreferencesVersion > GMAMapperPreferencesMaximumVersion {
		return UnsupportedPreferencesVersionError{
			MinimumVersion: GMAMapperPreferencesMinimumVersion,
			MaximumVersion: GMAMapperPreferencesMaximumVersion,
			DataVersion:    prefs.GMAMapperPreferencesVersion,
		}
	}
	return nil
}

//
// Update reads a set of saved preferences as LoadGMAPreferences does, but
// rather than returning a new GMAPreferences value, it updates the
// values of an existing GMAPreferences value with the input data.
//
func (prefs *GMAPreferences) Update(stream io.Reader) error {
	err := json.NewDecoder(stream).Decode(prefs)
	if err != nil {
		return err
	}
	if prefs.GMAPreferencesVersion < GMAPreferencesMinimumVersion || prefs.GMAPreferencesVersion > GMAPreferencesMaximumVersion {
		return UnsupportedPreferencesVersionError{
			MinimumVersion: GMAPreferencesMinimumVersion,
			MaximumVersion: GMAPreferencesMaximumVersion,
			DataVersion:    prefs.GMAPreferencesVersion,
		}
	}
	return nil
}

//
// UpdateFromSimpleConfig updates the corresponding
// configuration values in a UserPreferences value from a set of
// key=value pairs read from a simple config file.
//
func (prefs *UserPreferences) UpdateFromSimpleConfig(profileName string, cfg SimpleConfigurationData) error {
	var profile int = -1
	var err error

	if cfg == nil {
		return nil
	}
	for i, p := range prefs.Profiles {
		if p.Name == profileName {
			profile = i
			break
		}
	}
	if profile < 0 {
		return fmt.Errorf("requested profile \"%s\" not found in preferences data", profileName)
	}

	for k, v := range cfg {
		switch k {
		case "animate", "a":
			prefs.Animate = true
		case "no-animate", "A":
			prefs.Animate = false
		case "blur-all", "B":
			prefs.Profiles[profile].BlurAll = true
		case "no-blur-all":
			prefs.Profiles[profile].BlurAll = false
		case "blur-hp", "b":
			prefs.Profiles[profile].BlurPct, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("unable to understand blur-hp value \"%s\": %v", v, err)
			}
		case "config", "C":
			// TODO
		case "debug", "D":
			prefs.DebugLevel, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("unable to understand debug value \"%s\": %v", v, err)
			}
		case "debug-protocol":
			prefs.DebugProtocol = true
		case "dark", "d":
			prefs.DarkMode = true
		case "host", "h":
			prefs.Profiles[profile].Host = v
		case "image-format", "f":
			if v == "gif" {
				prefs.ImageFormat = GIF
			} else {
				prefs.ImageFormat = PNG
			}
		case "password", "P":
			prefs.Profiles[profile].Password = v
		case "port", "p":
			prefs.Profiles[profile].Port, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("unable to understand port value \"%s\": %v", v, err)
			}
		case "guide", "g":
			// TODO
		case "major", "G":
			// TODO
		case "module", "M":
			prefs.Profiles[profile].ModuleID = v
		case "master", "m", "keep-tools", "k":
			prefs.KeepTools = true
		case "no-chat", "n":
			prefs.Profiles[profile].SuppressChat = true
		case "select", "s":
			// TODO too late here
		case "transcript", "t":
			prefs.Profiles[profile].ChatLog = v
		case "username", "u":
			prefs.Profiles[profile].UserName = v
		case "proxy-url", "x":
			prefs.Profiles[profile].CurlProxy = v
		case "proxy-host", "X":
			prefs.Profiles[profile].ScpProxy = v
		case "preload", "l":
			prefs.PreloadImages = true
		case "button-size":
			switch v {
			case "medium", "m":
				prefs.ButtonSize = MediumButtons
			case "large", "l":
				prefs.ButtonSize = LargeButtons
			default:
				prefs.ButtonSize = SmallButtons
			}
		case "chat-history":
			prefs.Profiles[profile].ChatLimit, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("unable to understand chat-history value \"%s\": %v", v, err)
			}
		case "curl-path":
			prefs.CurlPath = v
		case "curl-url-base":
			prefs.Profiles[profile].CurlServer = v
		case "mkdir-path":
			prefs.Profiles[profile].ServerMkdir = v
		case "nc-path":
			prefs.Profiles[profile].NcPath = v
		case "scp-path":
			prefs.Profiles[profile].ScpPath = v
		case "scp-dest":
			prefs.Profiles[profile].ScpDest = v
		case "scp-server":
			prefs.Profiles[profile].ScpServer = v
		case "ssh-path":
			prefs.Profiles[profile].SshPath = v
		case "update-url":
			prefs.Profiles[profile].UpdateURL = v
		}
	}
	return nil
}

//
// SearchInPath looks for an executable program name by searching
// the user's execution path ($PATH environment variable)
//
func SearchInPath(program string) (string, error) {
	PATH, isValid := os.LookupEnv("PATH")
	if !isValid {
		return "", fmt.Errorf("no PATH environment variable")
	}
	for _, dir := range strings.Split(PATH, ":") {
		progpath := path.Join(dir, program)
		info, err := os.Stat(progpath)
		if err == nil && (info.Mode().Perm()&0o111) != 0 {
			return progpath, nil
		}
	}
	return "", fmt.Errorf("file not found in PATH")
}

// @[00]@| Go-GMA 5.28.0
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
