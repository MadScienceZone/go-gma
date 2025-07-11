/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______   _____      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   ) / ___ \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |( (   ) )   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )( (___) |   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /  \____  |   | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) |   |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\____) ) _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_)(_______)     #
#                                                                                      #
########################################################################################
*/
//
// Name Generator Cultural Module
// Automatically generated from JavaScript source file
//

package namegen

//
// TianHwan describes the naming conventions for the Tian-hwan
// culture. Its methods give further details, but generally speaking
// the main operation to perform on these types is to just call the
// Generate and GenerateWithSurnames methods to create new names which
// conform to their cultural patterns.
//
type TianHwan struct {
	BaseCulture
}

//
// prefix gives the prefix/selector string for each Tian-hwan gender, or an empty
// string if one is not defined.
//
func (c TianHwan) prefix(gender rune) string {
	switch gender {
	case 'F':
		return "__"
	case 'M':
		return "__"
	case 'S':
		return "__"
	default:
		return ""
	}
}

//
// defaultMinMax returns the minimum and maximum size of Tian-hwan names based on gender.
//
func (c TianHwan) defaultMinMax(gender rune) (int, int) {
	switch gender {
	case 'F':
		return 4, 10
	case 'M':
		return 4, 10
	case 'S':
		return 4, 10
	default:
		return 1, 1
	}
}

//
// Genders returns the set of genders defined for the Tian-hwan culture.
//
func (c TianHwan) Genders() []rune {
	return []rune{'F', 'M', 'S'}
}

//
// HasSurnames returns true if the Tian-hwan culture defines surnames.
//
func (c TianHwan) HasSurnames() bool {
	return true
}

//
// Name returns the name of the culture, i.e., "Tian-hwan".
//
func (c TianHwan) Name() string {
	return "Tian-hwan"
}

//
// maxCount returns the maximum number of times a rune can be added to
// a Tian-hwan name based on gender.
//
func (c TianHwan) maxCount(gender, char rune) int {
	switch gender {
	case 'F':
		switch char {
		case '-':
			return 1
		default:
			return 0
		}
	case 'M':
		switch char {
		case '-':
			return 1
		default:
			return 0
		}
	case 'S':
		switch char {
		default:
			return 0
		}
	default:
		return 0
	}
}

//
// HasGender returns true if the specified gender code is defined
// in the Tian-hwan culture.
//
func (c TianHwan) HasGender(gender rune) bool {
	switch gender {
	case 'F', 'M', 'S':
		return true
	default:
		return false
	}
}

//
// db returns the name data for the given gender in the Tian-hwan culture.
//
func (c TianHwan) db(gender rune) map[string][]nameFragment {
	switch gender {
	case 'F':
		return map[string][]nameFragment{
			"ok": {
				{0, 0.9286},
				{'-', 1.0},
			},
			"-b": {
				{'o', 1.0},
			},
			"hu": {
				{'i', 0.3333},
				{'n', 1.0},
			},
			"-S": {
				{'o', 1.0},
			},
			"he": {
				{'e', 1.0},
			},
			"-y": {
				{'e', 0.1429},
				{'o', 0.8571},
				{'u', 1.0},
			},
			"pa": {
				{'r', 1.0},
			},
			"um": {
				{'-', 0.5},
				{0, 1.0},
			},
			"ky": {
				{'u', 1.0},
			},
			"ou": {
				{'n', 1.0},
			},
			"im": {
				{0, 1.0},
			},
			"ey": {
				{'e', 1.0},
			},
			"-t": {
				{'u', 1.0},
			},
			"_M": {
				{'i', 0.6667},
				{'y', 1.0},
			},
			"Zu": {
				{'n', 1.0},
			},
			"_J": {
				{'u', 0.5455},
				{'a', 0.6364},
				{'i', 0.8182},
				{'o', 1.0},
			},
			"-r": {
				{'i', 1.0},
			},
			"Hy": {
				{'o', 0.1429},
				{'a', 0.2857},
				{'u', 0.8571},
				{'e', 1.0},
			},
			"_O": {
				{'k', 1.0},
			},
			"ee": {
				{0, 0.8},
				{'s', 0.9},
				{'-', 1.0},
			},
			"uu": {
				{'n', 1.0},
			},
			"-n": {
				{'a', 1.0},
			},
			"li": {
				{0, 1.0},
			},
			"_E": {
				{'u', 1.0},
			},
			"n-": {
				{'e', 0.05},
				{'o', 0.15},
				{'s', 0.4},
				{'j', 0.55},
				{'b', 0.6},
				{'h', 0.8},
				{'t', 0.85},
				{'y', 1.0},
			},
			"_A": {
				{'e', 1.0},
			},
			"uk": {
				{0, 1.0},
			},
			"Su": {
				{'n', 0.8571},
				{'-', 1.0},
			},
			"Sy": {
				{'u', 1.0},
			},
			"hy": {
				{'o', 0.3333},
				{'e', 0.6667},
				{0, 1.0},
			},
			"ea": {
				{'-', 1.0},
			},
			"in": {
				{'g', 0.125},
				{0, 0.875},
				{'-', 1.0},
			},
			"_C": {
				{'h', 1.0},
			},
			"ng": {
				{0, 0.2449},
				{'-', 1.0},
			},
			"su": {
				{'n', 0.4},
				{'k', 1.0},
			},
			"_N": {
				{'a', 1.0},
			},
			"me": {
				{0, 0.5},
				{'e', 1.0},
			},
			"_K": {
				{'y', 0.625},
				{'w', 0.75},
				{'u', 1.0},
			},
			"hw": {
				{'a', 1.0},
			},
			"Ye": {
				{'u', 1.0},
			},
			"on": {
				{'g', 0.3846},
				{0, 0.7692},
				{'-', 1.0},
			},
			"-l": {
				{'i', 1.0},
			},
			"Eu": {
				{'n', 1.0},
			},
			"-k": {
				{'u', 0.3333},
				{'a', 0.6667},
				{'y', 1.0},
			},
			"Ju": {
				{'n', 0.8333},
				{'-', 1.0},
			},
			"Yu": {
				{'-', 0.3333},
				{'n', 1.0},
			},
			"e-": {
				{'y', 0.2857},
				{'s', 0.5714},
				{'S', 0.7143},
				{'m', 0.8571},
				{'r', 1.0},
			},
			"So": {
				{'o', 0.7778},
				{'r', 0.8889},
				{'n', 1.0},
			},
			"ji": {
				{'n', 0.5},
				{'m', 1.0},
			},
			"se": {
				{'o', 1.0},
			},
			"-Y": {
				{'u', 1.0},
			},
			"Jo": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"ju": {
				{'n', 1.0},
			},
			"yo": {
				{'n', 0.5},
				{'i', 0.625},
				{'u', 1.0},
			},
			"-s": {
				{'u', 0.2},
				{'i', 0.28},
				{'h', 0.32},
				{'o', 0.92},
				{'e', 1.0},
			},
			"r-": {
				{'y', 1.0},
			},
			"ha": {
				{'n', 1.0},
			},
			"Ok": {
				{'-', 1.0},
			},
			"Bo": {
				{'k', 1.0},
			},
			"My": {
				{'u', 0.6667},
				{'o', 1.0},
			},
			"-j": {
				{'i', 0.1667},
				{'a', 0.6667},
				{'u', 1.0},
			},
			"Ch": {
				{'o', 0.625},
				{'a', 0.75},
				{'u', 1.0},
			},
			"Kw": {
				{'a', 1.0},
			},
			"ku": {
				{'m', 1.0},
			},
			"oi": {
				{'-', 1.0},
			},
			"Se": {
				{'-', 0.3333},
				{'o', 1.0},
			},
			"mi": {
				{0, 0.2},
				{'n', 1.0},
			},
			"eu": {
				{'n', 1.0},
			},
			"u-": {
				{'m', 0.3333},
				{'d', 0.6667},
				{'i', 1.0},
			},
			"na": {
				{'e', 0.5},
				{0, 1.0},
			},
			"hi": {
				{'n', 1.0},
			},
			"_Z": {
				{'u', 1.0},
			},
			"-w": {
				{'o', 1.0},
			},
			"-d": {
				{'a', 1.0},
			},
			"a-": {
				{'s', 0.6667},
				{'j', 1.0},
			},
			"ka": {
				{'k', 1.0},
			},
			"am": {
				{'-', 1.0},
			},
			"_H": {
				{'y', 0.4615},
				{'w', 0.6154},
				{'e', 1.0},
			},
			"my": {
				{'u', 1.0},
			},
			"rk": {
				{0, 1.0},
			},
			"en": {
				{0, 1.0},
			},
			"ei": {
				{0, 0.6667},
				{'-', 1.0},
			},
			"ga": {
				{'e', 1.0},
			},
			"-p": {
				{'a', 1.0},
			},
			"_Y": {
				{'e', 0.0909},
				{'o', 0.7273},
				{'a', 0.8182},
				{'u', 1.0},
			},
			"ye": {
				{0, 0.25},
				{'i', 0.5},
				{'-', 0.75},
				{'o', 1.0},
			},
			"or": {
				{'-', 1.0},
			},
			"an": {
				{'-', 0.25},
				{0, 0.375},
				{'g', 1.0},
			},
			"_S": {
				{'o', 0.3636},
				{'y', 0.4091},
				{'u', 0.7273},
				{'a', 0.8636},
				{'e', 1.0},
			},
			"_B": {
				{'o', 0.5},
				{'y', 1.0},
			},
			"tu": {
				{0, 1.0},
			},
			"es": {
				{'o', 1.0},
			},
			"-g": {
				{'a', 0.5},
				{'u', 1.0},
			},
			"ar": {
				{'k', 1.0},
			},
			"Ae": {
				{'i', 0.5},
				{'-', 1.0},
			},
			"ui": {
				{0, 1.0},
			},
			"ja": {
				{0, 1.0},
			},
			"Hw": {
				{'a', 1.0},
			},
			"ak": {
				{0, 1.0},
			},
			"si": {
				{'n', 0.5},
				{'l', 1.0},
			},
			"Ji": {
				{'n', 0.5},
				{'-', 1.0},
			},
			"_T": {
				{'a', 1.0},
			},
			"-a": {
				{0, 0.3333},
				{'h', 1.0},
			},
			"oe": {
				{0, 1.0},
			},
			"ya": {
				{'n', 1.0},
			},
			"wo": {
				{'o', 1.0},
			},
			"Mi": {
				{'-', 1.0},
			},
			"i-": {
				{'H', 0.1111},
				{'s', 0.4444},
				{'y', 0.5556},
				{'h', 0.6667},
				{'k', 0.7778},
				{'n', 0.8889},
				{'j', 1.0},
			},
			"eo": {
				{'n', 1.0},
			},
			"un": {
				{'g', 0.6471},
				{0, 0.7843},
				{'-', 1.0},
			},
			"-i": {
				{'l', 1.0},
			},
			"so": {
				{'o', 0.9375},
				{'k', 1.0},
			},
			"da": {
				{'e', 1.0},
			},
			"il": {
				{0, 1.0},
			},
			"_W": {
				{'o', 1.0},
			},
			"oo": {
				{'n', 0.4286},
				{0, 0.5},
				{'k', 0.7857},
				{'-', 1.0},
			},
			"sh": {
				{'i', 1.0},
			},
			"Na": {
				{'m', 1.0},
			},
			"-m": {
				{'e', 0.25},
				{'y', 0.375},
				{'i', 1.0},
			},
			"o-": {
				{'m', 0.2857},
				{'k', 0.4286},
				{'j', 0.5714},
				{'Y', 0.7143},
				{'s', 0.8571},
				{'h', 1.0},
			},
			"Ku": {
				{'n', 0.5},
				{'m', 1.0},
			},
			"-h": {
				{'e', 0.4},
				{'y', 0.6},
				{'o', 0.6667},
				{'w', 0.9333},
				{'u', 1.0},
			},
			"gu": {
				{'n', 1.0},
			},
			"Wo": {
				{'o', 1.0},
			},
			"ah": {
				{0, 1.0},
			},
			"m-": {
				{'j', 0.5},
				{'s', 1.0},
			},
			"g-": {
				{'j', 0.1081},
				{'b', 0.1351},
				{'g', 0.1892},
				{'k', 0.2162},
				{'m', 0.2973},
				{'h', 0.5135},
				{'w', 0.5405},
				{'e', 0.5676},
				{'p', 0.5946},
				{'n', 0.6216},
				{'l', 0.6486},
				{'a', 0.7297},
				{'o', 0.7568},
				{'s', 1.0},
			},
			"k-": {
				{'m', 0.1429},
				{'r', 0.2857},
				{'j', 0.4286},
				{'s', 0.7143},
				{'H', 0.8571},
				{'h', 1.0},
			},
			"Sa": {
				{'n', 1.0},
			},
			"ho": {
				{'e', 0.1667},
				{'n', 0.3333},
				{0, 0.6667},
				{'-', 0.8333},
				{'o', 1.0},
			},
			"ri": {
				{0, 0.5},
				{'m', 1.0},
			},
			"wa": {
				{'n', 0.2857},
				{0, 0.7143},
				{'-', 1.0},
			},
			"-o": {
				{'o', 0.3333},
				{'k', 1.0},
			},
			"bo": {
				{'k', 1.0},
			},
			"By": {
				{'u', 1.0},
			},
			"Ta": {
				{'e', 1.0},
			},
			"yu": {
				{'u', 0.0625},
				{'n', 1.0},
			},
			"Yo": {
				{'o', 0.1429},
				{'u', 0.7143},
				{'n', 1.0},
			},
			"Ky": {
				{'u', 1.0},
			},
			"Ja": {
				{'e', 1.0},
			},
			"ae": {
				{'-', 0.4},
				{0, 1.0},
			},
			"He": {
				{'y', 0.1667},
				{'a', 0.3333},
				{'-', 0.5},
				{'e', 1.0},
			},
			"Ya": {
				{'n', 1.0},
			},
			"-H": {
				{'e', 0.5},
				{'y', 1.0},
			},
			"__": {
				{'C', 0.0808},
				{'W', 0.0909},
				{'J', 0.202},
				{'M', 0.2929},
				{'B', 0.3131},
				{'O', 0.3737},
				{'Y', 0.4848},
				{'T', 0.4949},
				{'A', 0.5152},
				{'Z', 0.5253},
				{'N', 0.5354},
				{'K', 0.6162},
				{'E', 0.6465},
				{'S', 0.8687},
				{'H', 1.0},
			},
			"-e": {
				{'n', 0.5},
				{'i', 1.0},
			},
		}
	case 'M':
		return map[string][]nameFragment{
			"Ye": {
				{'e', 1.0},
			},
			"_I": {
				{'l', 0.3333},
				{'n', 1.0},
			},
			"Me": {
				{'e', 1.0},
			},
			"b-": {
				{'D', 1.0},
			},
			"Mo": {
				{0, 0.3333},
				{'o', 1.0},
			},
			"In": {
				{'-', 1.0},
			},
			"ul": {
				{0, 0.7273},
				{'l', 0.8182},
				{'s', 0.9091},
				{'-', 1.0},
			},
			"-O": {
				{'k', 0.5},
				{'h', 1.0},
			},
			"aa": {
				{0, 1.0},
			},
			"_K": {
				{'u', 0.0556},
				{'i', 0.2222},
				{'w', 0.3333},
				{'y', 0.9444},
				{'o', 1.0},
			},
			"Si": {
				{0, 0.125},
				{'k', 0.5},
				{'e', 0.75},
				{'p', 0.875},
				{'n', 1.0},
			},
			"Ho": {
				{'-', 0.1875},
				{'o', 0.3125},
				{'n', 0.75},
				{'i', 0.8125},
				{0, 1.0},
			},
			"So": {
				{'h', 0.0286},
				{'k', 0.0571},
				{'o', 0.8286},
				{'-', 0.8571},
				{'n', 1.0},
			},
			"e-": {
				{'K', 0.087},
				{'W', 0.1739},
				{'S', 0.3913},
				{'J', 0.5217},
				{'H', 0.7391},
				{'D', 0.7826},
				{'T', 0.8261},
				{'U', 0.8696},
				{'E', 0.913},
				{'Y', 1.0},
			},
			"-K": {
				{'o', 0.0909},
				{'e', 0.4091},
				{'y', 0.5909},
				{'i', 0.7727},
				{'w', 0.8182},
				{'a', 0.8636},
				{'u', 1.0},
			},
			"Da": {
				{'e', 1.0},
			},
			"Jo": {
				{0, 0.0714},
				{'n', 0.5714},
				{'o', 1.0},
			},
			"yo": {
				{'u', 0.3077},
				{'n', 0.9231},
				{'-', 1.0},
			},
			"Ki": {
				{'m', 0.2857},
				{'-', 0.7143},
				{0, 1.0},
			},
			"-Y": {
				{'i', 0.1176},
				{'u', 0.3529},
				{'o', 1.0},
			},
			"Eu": {
				{'n', 0.6},
				{'i', 1.0},
			},
			"on": {
				{'-', 0.172},
				{'g', 0.7634},
				{0, 1.0},
			},
			"Yu": {
				{'l', 0.1429},
				{0, 0.2857},
				{'p', 0.4286},
				{'n', 1.0},
			},
			"h-": {
				{'K', 0.5},
				{'B', 1.0},
			},
			"Ju": {
				{0, 0.1579},
				{'n', 0.8947},
				{'-', 1.0},
			},
			"An": {
				{'n', 0.5},
				{'-', 1.0},
			},
			"ie": {
				{'k', 0.5},
				{'n', 1.0},
			},
			"Ch": {
				{'i', 0.0645},
				{'u', 0.4839},
				{'a', 0.6452},
				{'o', 0.9355},
				{'e', 1.0},
			},
			"oi": {
				{0, 1.0},
			},
			"Kw": {
				{'o', 0.3333},
				{'a', 1.0},
			},
			"-E": {
				{'u', 1.0},
			},
			"My": {
				{'u', 0.4},
				{'-', 0.6},
				{'o', 1.0},
			},
			"Bo": {
				{'n', 0.5},
				{0, 0.75},
				{'k', 1.0},
			},
			"Gi": {
				{0, 0.1667},
				{'l', 0.6667},
				{'n', 0.8333},
				{'-', 1.0},
			},
			"Il": {
				{0, 0.5},
				{'-', 1.0},
			},
			"Se": {
				{'u', 0.5},
				{'o', 0.8333},
				{'-', 1.0},
			},
			"Hu": {
				{'n', 1.0},
			},
			"Ok": {
				{0, 0.5},
				{'-', 1.0},
			},
			"ha": {
				{0, 0.1429},
				{'-', 0.2857},
				{'n', 0.8571},
				{'e', 1.0},
			},
			"op": {
				{0, 1.0},
			},
			"u-": {
				{'H', 0.625},
				{'Y', 0.75},
				{'C', 0.875},
				{'K', 1.0},
			},
			"nn": {
				{0, 1.0},
			},
			"eu": {
				{'n', 1.0},
			},
			"hi": {
				{'n', 0.6667},
				{'k', 1.0},
			},
			"-D": {
				{'o', 0.3333},
				{'a', 0.6667},
				{'u', 1.0},
			},
			"a-": {
				{'D', 0.3333},
				{'B', 0.6667},
				{'U', 1.0},
			},
			"mo": {
				{'o', 1.0},
			},
			"uo": {
				{'-', 1.0},
			},
			"ek": {
				{'-', 0.25},
				{0, 1.0},
			},
			"-U": {
				{0, 0.6667},
				{'n', 1.0},
			},
			"um": {
				{'-', 0.3333},
				{0, 1.0},
			},
			"Ka": {
				{'n', 1.0},
			},
			"ky": {
				{'h', 1.0},
			},
			"Wh": {
				{'a', 1.0},
			},
			"-Z": {
				{'o', 1.0},
			},
			"ok": {
				{0, 0.625},
				{'-', 1.0},
			},
			"To": {
				{'n', 1.0},
			},
			"hu": {
				{'n', 0.2143},
				{'l', 0.9286},
				{0, 1.0},
			},
			"-T": {
				{'e', 0.1667},
				{'a', 1.0},
			},
			"ik": {
				{'o', 0.1667},
				{0, 1.0},
			},
			"he": {
				{'o', 0.5},
				{'-', 1.0},
			},
			"-S": {
				{'o', 0.3617},
				{'e', 0.4043},
				{'i', 0.5532},
				{'h', 0.5957},
				{'u', 0.9362},
				{'a', 1.0},
			},
			"Ha": {
				{'e', 0.3333},
				{'-', 0.5},
				{'n', 0.6667},
				{'k', 0.8333},
				{'h', 1.0},
			},
			"_M": {
				{'y', 0.4167},
				{'m', 0.5},
				{'i', 0.6667},
				{'a', 0.9167},
				{'u', 1.0},
			},
			"Ke": {
				{'e', 0.1429},
				{'u', 1.0},
			},
			"-B": {
				{'u', 0.2},
				{'a', 0.3},
				{'o', 1.0},
			},
			"_J": {
				{'i', 0.1667},
				{'a', 0.4286},
				{'u', 0.6667},
				{'o', 0.9524},
				{'e', 1.0},
			},
			"l-": {
				{'H', 0.25},
				{'S', 1.0},
			},
			"ec": {
				{'k', 1.0},
			},
			"ou": {
				{'-', 0.0588},
				{'n', 1.0},
			},
			"Lu": {
				{'o', 1.0},
			},
			"im": {
				{0, 1.0},
			},
			"ai": {
				{0, 0.3333},
				{'-', 1.0},
			},
			"_E": {
				{'e', 0.2},
				{'u', 1.0},
			},
			"ip": {
				{0, 1.0},
			},
			"uk": {
				{0, 0.6667},
				{'-', 1.0},
			},
			"Su": {
				{'k', 0.0417},
				{0, 0.2083},
				{'c', 0.25},
				{'n', 0.9167},
				{'p', 1.0},
			},
			"_A": {
				{'n', 1.0},
			},
			"n-": {
				{'T', 0.0638},
				{'A', 0.0851},
				{'C', 0.1277},
				{'Y', 0.2128},
				{'O', 0.234},
				{'M', 0.2766},
				{'S', 0.5532},
				{'K', 0.6596},
				{'W', 0.7021},
				{'J', 0.8085},
				{'B', 0.8298},
				{'G', 0.9149},
				{'H', 1.0},
			},
			"Hy": {
				{'a', 0.1053},
				{'u', 0.7895},
				{'o', 0.9474},
				{'e', 1.0},
			},
			"el": {
				{0, 1.0},
			},
			"_P": {
				{'o', 0.5},
				{'i', 1.0},
			},
			"_O": {
				{'k', 1.0},
			},
			"ee": {
				{0, 0.7778},
				{'-', 1.0},
			},
			"Ga": {
				{'b', 1.0},
			},
			"ng": {
				{0, 0.2782},
				{'-', 1.0},
			},
			"Li": {
				{'m', 0.5},
				{'p', 1.0},
			},
			"_N": {
				{'u', 1.0},
			},
			"-G": {
				{'y', 0.2857},
				{'u', 0.5714},
				{'i', 1.0},
			},
			"Pi": {
				{'l', 1.0},
			},
			"ea": {
				{'k', 1.0},
			},
			"Ma": {
				{'n', 1.0},
			},
			"_C": {
				{'h', 1.0},
			},
			"in": {
				{'g', 0.087},
				{'-', 0.5217},
				{0, 1.0},
			},
			"hy": {
				{'o', 1.0},
			},
			"o-": {
				{'A', 0.0667},
				{'M', 0.1333},
				{'P', 0.2},
				{'Y', 0.3333},
				{'K', 0.4667},
				{'S', 0.5333},
				{'B', 0.6667},
				{'H', 0.8667},
				{'J', 1.0},
			},
			"Ku": {
				{'n', 0.25},
				{0, 1.0},
			},
			"uc": {
				{'k', 1.0},
			},
			"Po": {
				{'-', 1.0},
			},
			"lk": {
				{'-', 1.0},
			},
			"Je": {
				{'e', 0.3333},
				{'o', 0.6667},
				{'a', 1.0},
			},
			"ls": {
				{'o', 1.0},
			},
			"Ba": {
				{'e', 1.0},
			},
			"il": {
				{0, 0.5},
				{'k', 0.6667},
				{'-', 1.0},
			},
			"_W": {
				{'a', 0.4},
				{'e', 0.6},
				{'o', 1.0},
			},
			"oo": {
				{'n', 0.5417},
				{'-', 0.6667},
				{'k', 0.7083},
				{0, 1.0},
			},
			"-M": {
				{'i', 0.125},
				{'u', 0.25},
				{'a', 0.5},
				{'o', 0.875},
				{'e', 1.0},
			},
			"Bu": {
				{'m', 1.0},
			},
			"ho": {
				{'o', 0.1111},
				{'n', 0.4444},
				{'l', 0.8889},
				{0, 1.0},
			},
			"up": {
				{0, 1.0},
			},
			"g-": {
				{'P', 0.0208},
				{'L', 0.0417},
				{'A', 0.0521},
				{'J', 0.125},
				{'B', 0.1562},
				{'K', 0.25},
				{'Z', 0.2604},
				{'O', 0.2708},
				{'Y', 0.3438},
				{'I', 0.3646},
				{'M', 0.4167},
				{'g', 0.4271},
				{'U', 0.4375},
				{'C', 0.5521},
				{'H', 0.6771},
				{'D', 0.7083},
				{'G', 0.7396},
				{'S', 0.9583},
				{'W', 1.0},
			},
			"k-": {
				{'J', 0.25},
				{'B', 0.375},
				{'S', 0.625},
				{'W', 0.75},
				{'K', 0.875},
				{'C', 1.0},
			},
			"Sa": {
				{'m', 0.25},
				{'n', 1.0},
			},
			"Wo": {
				{'o', 0.4444},
				{'n', 0.8889},
				{'k', 1.0},
			},
			"ao": {
				{'n', 1.0},
			},
			"Ah": {
				{0, 1.0},
			},
			"_D": {
				{'o', 0.5714},
				{'a', 0.9286},
				{'u', 1.0},
			},
			"m-": {
				{'K', 1.0},
			},
			"Zo": {
				{'o', 1.0},
			},
			"Gy": {
				{'o', 0.6667},
				{'e', 1.0},
			},
			"Un": {
				{'g', 1.0},
			},
			"ah": {
				{0, 0.5},
				{'-', 1.0},
			},
			"Ky": {
				{'u', 0.7333},
				{'o', 1.0},
			},
			"Yo": {
				{'n', 0.5},
				{'o', 0.5667},
				{'p', 0.6},
				{'u', 1.0},
			},
			"ck": {
				{'-', 0.3333},
				{0, 1.0},
			},
			"ae": {
				{'k', 0.08},
				{0, 0.32},
				{'-', 1.0},
			},
			"Ja": {
				{'i', 0.0833},
				{'o', 0.1667},
				{'n', 0.5},
				{'e', 1.0},
			},
			"nk": {
				{'y', 1.0},
			},
			"Wa": {
				{'n', 0.4286},
				{'-', 0.5714},
				{'h', 0.7143},
				{0, 0.8571},
				{'a', 1.0},
			},
			"We": {
				{'n', 0.5},
				{'o', 1.0},
			},
			"oh": {
				{'-', 0.5},
				{'y', 1.0},
			},
			"wa": {
				{0, 0.4286},
				{'n', 1.0},
			},
			"yu": {
				{0, 0.0714},
				{'k', 0.1071},
				{'n', 0.8571},
				{'-', 1.0},
			},
			"-L": {
				{'i', 1.0},
			},
			"By": {
				{'u', 0.6667},
				{'e', 1.0},
			},
			"Ta": {
				{'e', 0.7273},
				{'i', 0.9091},
				{'k', 1.0},
			},
			"Ee": {
				{'n', 1.0},
			},
			"Te": {
				{'c', 1.0},
			},
			"-H": {
				{'o', 0.25},
				{'e', 0.4062},
				{'y', 0.6875},
				{'w', 0.8438},
				{'i', 0.875},
				{'u', 0.9062},
				{'a', 1.0},
			},
			"-W": {
				{'a', 0.3},
				{'h', 0.5},
				{'o', 1.0},
			},
			"Do": {
				{'h', 0.1},
				{0, 0.3},
				{'n', 0.7},
				{'-', 0.9},
				{'o', 1.0},
			},
			"__": {
				{'C', 0.0681},
				{'I', 0.0809},
				{'M', 0.1319},
				{'O', 0.1362},
				{'E', 0.1574},
				{'Y', 0.2553},
				{'S', 0.4426},
				{'W', 0.4851},
				{'G', 0.5106},
				{'D', 0.5702},
				{'H', 0.6681},
				{'A', 0.6723},
				{'T', 0.7021},
				{'L', 0.7106},
				{'P', 0.7191},
				{'N', 0.7234},
				{'K', 0.8},
				{'B', 0.8213},
				{'J', 1.0},
			},
			"y-": {
				{'S', 1.0},
			},
			"He": {
				{'e', 0.6667},
				{'l', 0.8333},
				{'u', 1.0},
			},
			"_G": {
				{'y', 0.1667},
				{'a', 0.3333},
				{'u', 0.5},
				{'i', 1.0},
			},
			"Oh": {
				{0, 1.0},
			},
			"-J": {
				{'o', 0.1},
				{'e', 0.15},
				{'i', 0.5},
				{'u', 0.95},
				{'a', 1.0},
			},
			"en": {
				{0, 0.25},
				{'-', 1.0},
			},
			"ob": {
				{0, 1.0},
			},
			"am": {
				{0, 1.0},
			},
			"Sh": {
				{'i', 1.0},
			},
			"_H": {
				{'o', 0.3478},
				{'y', 0.7826},
				{'e', 0.8261},
				{'i', 0.8696},
				{'a', 1.0},
			},
			"Le": {
				{'e', 1.0},
			},
			"-A": {
				{0, 0.3333},
				{'h', 0.6667},
				{'n', 1.0},
			},
			"ko": {
				{'n', 1.0},
			},
			"Ik": {
				{0, 1.0},
			},
			"yh": {
				{'u', 1.0},
			},
			"-g": {
				{'a', 1.0},
			},
			"ll": {
				{0, 1.0},
			},
			"Hi": {
				{'e', 1.0},
			},
			"-I": {
				{'k', 0.5},
				{'l', 1.0},
			},
			"Du": {
				{0, 0.3333},
				{'k', 0.6667},
				{'c', 1.0},
			},
			"Gu": {
				{'i', 1.0},
			},
			"_L": {
				{'e', 0.5},
				{'u', 1.0},
			},
			"Mm": {
				{'o', 1.0},
			},
			"ga": {
				{'k', 1.0},
			},
			"_Y": {
				{'e', 0.0435},
				{'o', 0.8696},
				{'u', 1.0},
			},
			"-P": {
				{'y', 0.3333},
				{'i', 1.0},
			},
			"ye": {
				{'o', 0.3333},
				{'-', 0.6667},
				{0, 1.0},
			},
			"or": {
				{0, 1.0},
			},
			"an": {
				{0, 0.2667},
				{'-', 0.4667},
				{'g', 1.0},
			},
			"-C": {
				{'h', 1.0},
			},
			"_B": {
				{'u', 0.2},
				{'y', 0.8},
				{'o', 1.0},
			},
			"_S": {
				{'e', 0.2273},
				{'o', 0.6364},
				{'u', 0.8182},
				{'a', 0.9318},
				{'h', 0.9773},
				{'i', 1.0},
			},
			"Mu": {
				{0, 0.5},
				{'-', 1.0},
			},
			"Hw": {
				{'a', 1.0},
			},
			"ak": {
				{0, 0.75},
				{'i', 1.0},
			},
			"ki": {
				{0, 1.0},
			},
			"Ji": {
				{'n', 1.0},
			},
			"ol": {
				{0, 1.0},
			},
			"ui": {
				{'-', 0.4},
				{'k', 0.6},
				{0, 1.0},
			},
			"Py": {
				{'o', 1.0},
			},
			"Nu": {
				{'n', 1.0},
			},
			"un": {
				{'-', 0.0968},
				{'g', 0.7312},
				{'k', 0.7419},
				{0, 1.0},
			},
			"so": {
				{'o', 1.0},
			},
			"eo": {
				{'b', 0.125},
				{'n', 0.625},
				{'u', 0.75},
				{'k', 1.0},
			},
			"Mi": {
				{'n', 0.6667},
				{'-', 1.0},
			},
			"i-": {
				{'S', 0.1111},
				{'W', 0.2222},
				{'T', 0.4444},
				{'B', 0.5556},
				{'H', 0.7778},
				{'J', 0.8889},
				{'Y', 1.0},
			},
			"Ko": {
				{'r', 0.3333},
				{'n', 1.0},
			},
			"Yi": {
				{0, 0.5},
				{'n', 1.0},
			},
			"_T": {
				{'a', 0.8571},
				{'o', 1.0},
			},
			"ya": {
				{'n', 1.0},
			},
			"wo": {
				{'n', 1.0},
			},
			"ab": {
				{'-', 1.0},
			},
		}
	case 'S':
		return map[string][]nameFragment{
			"_J": {
				{'u', 0.125},
				{'o', 0.5},
				{'a', 0.625},
				{'e', 1.0},
			},
			"ar": {
				{'k', 1.0},
			},
			"_M": {
				{'o', 0.75},
				{'a', 1.0},
			},
			"Ro": {
				{'n', 1.0},
			},
			"om": {
				{0, 1.0},
			},
			"an": {
				{0, 0.1818},
				{'g', 1.0},
			},
			"sa": {
				{'i', 1.0},
			},
			"_B": {
				{'y', 0.3333},
				{'a', 1.0},
			},
			"_S": {
				{'u', 0.2353},
				{'i', 0.4118},
				{'a', 0.4706},
				{'o', 0.7059},
				{'e', 0.8235},
				{'h', 1.0},
			},
			"im": {
				{0, 1.0},
			},
			"ou": {
				{'n', 0.25},
				{'j', 0.5},
				{0, 1.0},
			},
			"Gu": {
				{0, 1.0},
			},
			"ai": {
				{0, 0.6667},
				{'k', 1.0},
			},
			"_L": {
				{'i', 0.6667},
				{'e', 1.0},
			},
			"ga": {
				{'i', 1.0},
			},
			"_Y": {
				{'e', 0.1667},
				{'a', 0.25},
				{'o', 0.6667},
				{'i', 0.8333},
				{'u', 1.0},
			},
			"Pa": {
				{'r', 0.3333},
				{'e', 0.6667},
				{'k', 1.0},
			},
			"rk": {
				{0, 1.0},
			},
			"Ka": {
				{'n', 1.0},
			},
			"hu": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"ik": {
				{0, 1.0},
			},
			"am": {
				{0, 1.0},
			},
			"Sh": {
				{'i', 0.6667},
				{'o', 1.0},
			},
			"Le": {
				{'e', 1.0},
			},
			"he": {
				{'e', 1.0},
			},
			"_H": {
				{'u', 0.25},
				{'o', 0.5},
				{'y', 0.625},
				{'a', 0.875},
				{'w', 1.0},
			},
			"Ha": {
				{'n', 0.5},
				{0, 1.0},
			},
			"ok": {
				{0, 1.0},
			},
			"Ko": {
				{0, 0.5},
				{'o', 1.0},
			},
			"Yi": {
				{0, 0.5},
				{'m', 1.0},
			},
			"Li": {
				{'m', 0.5},
				{0, 1.0},
			},
			"Ng": {
				{'a', 1.0},
			},
			"_N": {
				{'g', 0.3333},
				{'o', 0.6667},
				{'a', 1.0},
			},
			"eo": {
				{0, 0.3333},
				{'k', 0.5},
				{'n', 1.0},
			},
			"un": {
				{0, 0.5455},
				{'g', 1.0},
			},
			"ng": {
				{0, 1.0},
			},
			"uj": {
				{0, 1.0},
			},
			"uh": {
				{0, 1.0},
			},
			"Ma": {
				{0, 1.0},
			},
			"in": {
				{'n', 0.25},
				{0, 1.0},
			},
			"_C": {
				{'h', 1.0},
			},
			"_T": {
				{'s', 1.0},
			},
			"oe": {
				{0, 1.0},
			},
			"uk": {
				{0, 1.0},
			},
			"_A": {
				{'n', 0.5},
				{'h', 1.0},
			},
			"Su": {
				{'k', 0.25},
				{'h', 0.5},
				{'n', 1.0},
			},
			"we": {
				{'h', 1.0},
			},
			"Hw": {
				{'a', 1.0},
			},
			"ak": {
				{0, 1.0},
			},
			"_P": {
				{'a', 1.0},
			},
			"_O": {
				{'h', 1.0},
			},
			"ee": {
				{'m', 0.3333},
				{0, 1.0},
			},
			"ol": {
				{0, 1.0},
			},
			"Hy": {
				{'u', 1.0},
			},
			"ho": {
				{'i', 0.1429},
				{'l', 0.2857},
				{'u', 0.4286},
				{0, 0.5714},
				{'n', 0.8571},
				{'e', 1.0},
			},
			"Jo": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"yo": {
				{'n', 0.3333},
				{'m', 0.6667},
				{'o', 1.0},
			},
			"Ki": {
				{'m', 1.0},
			},
			"Sa": {
				{'n', 1.0},
			},
			"Ho": {
				{'n', 0.5},
				{0, 1.0},
			},
			"So": {
				{'n', 0.5},
				{'o', 0.75},
				{0, 1.0},
			},
			"Yu": {
				{0, 0.5},
				{'n', 1.0},
			},
			"Gw": {
				{'a', 1.0},
			},
			"Ju": {
				{'n', 1.0},
			},
			"Wo": {
				{'o', 1.0},
			},
			"Ah": {
				{'n', 1.0},
			},
			"on": {
				{0, 0.6429},
				{'g', 1.0},
			},
			"Je": {
				{'o', 0.6667},
				{'u', 1.0},
			},
			"hw": {
				{'e', 1.0},
			},
			"_I": {
				{0, 1.0},
			},
			"Ye": {
				{'o', 1.0},
			},
			"Ku": {
				{0, 1.0},
			},
			"Mo": {
				{'o', 0.3333},
				{0, 0.6667},
				{'k', 1.0},
			},
			"Na": {
				{'m', 1.0},
			},
			"_K": {
				{'w', 0.2857},
				{'a', 0.4286},
				{'o', 0.7143},
				{'i', 0.8571},
				{'u', 1.0},
			},
			"Ts": {
				{'a', 1.0},
			},
			"oo": {
				{'n', 0.2857},
				{'k', 0.4286},
				{0, 1.0},
			},
			"Si": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"il": {
				{0, 1.0},
			},
			"Ba": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"_W": {
				{'a', 0.5},
				{'o', 1.0},
			},
			"__": {
				{'S', 0.1667},
				{'K', 0.2353},
				{'G', 0.2647},
				{'I', 0.2745},
				{'O', 0.2843},
				{'H', 0.3627},
				{'J', 0.4412},
				{'P', 0.4706},
				{'L', 0.5},
				{'W', 0.5196},
				{'C', 0.6667},
				{'Y', 0.7843},
				{'M', 0.8235},
				{'A', 0.8431},
				{'R', 0.9314},
				{'T', 0.9412},
				{'N', 0.9706},
				{'B', 1.0},
			},
			"_R": {
				{'h', 0.2222},
				{'y', 0.5556},
				{'a', 0.6667},
				{'o', 0.7778},
				{'i', 1.0},
			},
			"em": {
				{0, 1.0},
			},
			"nn": {
				{0, 1.0},
			},
			"eu": {
				{'n', 1.0},
			},
			"hi": {
				{'n', 0.5},
				{'m', 0.75},
				{0, 1.0},
			},
			"Ry": {
				{'o', 0.6667},
				{'u', 1.0},
			},
			"hn": {
				{0, 1.0},
			},
			"Ya": {
				{'n', 1.0},
			},
			"Oh": {
				{0, 1.0},
			},
			"_G": {
				{'w', 0.3333},
				{'u', 0.6667},
				{'i', 1.0},
			},
			"Gi": {
				{'l', 1.0},
			},
			"Se": {
				{'o', 1.0},
			},
			"ae": {
				{0, 1.0},
			},
			"Ra": {
				{0, 1.0},
			},
			"eh": {
				{0, 1.0},
			},
			"Wa": {
				{'n', 1.0},
			},
			"Ja": {
				{'n', 1.0},
			},
			"An": {
				{0, 1.0},
			},
			"Yo": {
				{'o', 0.4},
				{'u', 1.0},
			},
			"Kw": {
				{'a', 1.0},
			},
			"Ch": {
				{'w', 0.0667},
				{'u', 0.2667},
				{'o', 0.6667},
				{'a', 0.8667},
				{'i', 1.0},
			},
			"oi": {
				{0, 1.0},
			},
			"yu": {
				{0, 0.5},
				{'n', 1.0},
			},
			"By": {
				{'o', 1.0},
			},
			"ha": {
				{0, 0.3333},
				{'n', 0.6667},
				{'e', 1.0},
			},
			"Hu": {
				{'n', 1.0},
			},
			"No": {
				{'h', 1.0},
			},
			"oh": {
				{0, 1.0},
			},
			"Ri": {
				{0, 0.5},
				{'m', 1.0},
			},
			"wa": {
				{'k', 0.25},
				{'n', 1.0},
			},
			"Rh": {
				{'e', 1.0},
			},
		}
	default:
		return nil
	}
}
//
// End of generated data
//
