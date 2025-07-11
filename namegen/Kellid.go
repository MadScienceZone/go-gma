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
// Kellid describes the naming conventions for the Kellid
// culture. Its methods give further details, but generally speaking
// the main operation to perform on these types is to just call the
// Generate and GenerateWithSurnames methods to create new names which
// conform to their cultural patterns.
//
type Kellid struct {
	BaseCulture
}

//
// prefix gives the prefix/selector string for each Kellid gender, or an empty
// string if one is not defined.
//
func (c Kellid) prefix(gender rune) string {
	switch gender {
	case 'F':
		return "__"
	case 'M':
		return "__"
	default:
		return ""
	}
}

//
// defaultMinMax returns the minimum and maximum size of Kellid names based on gender.
//
func (c Kellid) defaultMinMax(gender rune) (int, int) {
	switch gender {
	case 'F':
		return 4, 6
	case 'M':
		return 4, 6
	default:
		return 1, 1
	}
}

//
// Genders returns the set of genders defined for the Kellid culture.
//
func (c Kellid) Genders() []rune {
	return []rune{'F', 'M'}
}

//
// Name returns the name of the culture, i.e., "Kellid".
//
func (c Kellid) Name() string {
	return "Kellid"
}

//
// HasGender returns true if the specified gender code is defined
// in the Kellid culture.
//
func (c Kellid) HasGender(gender rune) bool {
	switch gender {
	case 'F', 'M':
		return true
	default:
		return false
	}
}

//
// db returns the name data for the given gender in the Kellid culture.
//
func (c Kellid) db(gender rune) map[string][]nameFragment {
	switch gender {
	case 'F':
		return map[string][]nameFragment{
			"_D": {
				{'e', 0.3},
				{'a', 1.0},
			},
			"_V": {
				{'a', 0.6},
				{'e', 1.0},
			},
			"Ne": {
				{'l', 0.6},
				{'s', 1.0},
			},
			"_Y": {
				{'a', 0.5556},
				{'e', 1.0},
			},
			"le": {
				{'t', 0.5},
				{'n', 1.0},
			},
			"Ve": {
				{'l', 0.25},
				{'s', 1.0},
			},
			"Se": {
				{'s', 1.0},
			},
			"Be": {
				{'l', 0.6667},
				{'s', 1.0},
			},
			"ll": {
				{'a', 1.0},
			},
			"na": {
				{0, 1.0},
			},
			"Ba": {
				{'l', 0.5},
				{'n', 0.75},
				{'r', 1.0},
			},
			"Fe": {
				{'l', 0.5},
				{'s', 1.0},
			},
			"se": {
				{'n', 1.0},
			},
			"al": {
				{'l', 0.0667},
				{'u', 0.1333},
				{'i', 0.3333},
				{'a', 0.5333},
				{'e', 0.6667},
				{'k', 1.0},
			},
			"_K": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"sk": {
				{'a', 0.5},
				{'e', 0.75},
				{'i', 1.0},
			},
			"si": {
				{'k', 1.0},
			},
			"_J": {
				{'a', 0.5714},
				{'e', 1.0},
			},
			"_S": {
				{'e', 0.3333},
				{'h', 0.8889},
				{'a', 1.0},
			},
			"Ye": {
				{'l', 1.0},
			},
			"an": {
				{'i', 0.0556},
				{'l', 0.1111},
				{'n', 0.5556},
				{'k', 0.8333},
				{'a', 0.8889},
				{'e', 1.0},
			},
			"Le": {
				{'s', 0.5},
				{'l', 1.0},
			},
			"_L": {
				{'e', 1.0},
			},
			"re": {
				{'t', 0.5},
				{'n', 1.0},
			},
			"_N": {
				{'e', 0.5556},
				{'a', 1.0},
			},
			"An": {
				{'k', 0.3333},
				{'n', 0.8333},
				{'a', 1.0},
			},
			"Ja": {
				{'n', 1.0},
			},
			"la": {
				{0, 1.0},
			},
			"ri": {
				{'t', 1.0},
			},
			"en": {
				{0, 1.0},
			},
			"_E": {
				{'l', 0.5},
				{'s', 1.0},
			},
			"ha": {
				{'l', 0.2},
				{0, 0.6},
				{'n', 1.0},
			},
			"De": {
				{'s', 1.0},
			},
			"ik": {
				{0, 1.0},
			},
			"El": {
				{'u', 0.3333},
				{'k', 0.6667},
				{'a', 1.0},
			},
			"_B": {
				{'e', 0.2727},
				{'a', 1.0},
			},
			"Je": {
				{'l', 0.3333},
				{'s', 1.0},
			},
			"nn": {
				{'e', 0.5455},
				{'a', 0.7273},
				{'k', 0.8182},
				{'l', 0.9091},
				{'i', 1.0},
			},
			"gi": {
				{'k', 1.0},
			},
			"ke": {
				{'t', 1.0},
			},
			"Fa": {
				{'n', 0.5},
				{'r', 1.0},
			},
			"sl": {
				{'a', 1.0},
			},
			"ar": {
				{'i', 0.1111},
				{'l', 0.2222},
				{'k', 0.7778},
				{'e', 1.0},
			},
			"li": {
				{'t', 0.5},
				{'k', 1.0},
			},
			"he": {
				{'l', 0.6667},
				{'n', 1.0},
			},
			"rl": {
				{'a', 1.0},
			},
			"Da": {
				{'r', 0.2857},
				{'g', 0.4286},
				{'n', 0.5714},
				{'l', 1.0},
			},
			"it": {
				{0, 1.0},
			},
			"Sh": {
				{'a', 0.6},
				{'e', 1.0},
			},
			"Ya": {
				{'l', 0.2},
				{'n', 1.0},
			},
			"Es": {
				{'i', 0.3333},
				{'h', 1.0},
			},
			"Ag": {
				{'i', 1.0},
			},
			"es": {
				{'l', 0.2222},
				{'k', 0.4444},
				{'h', 0.9444},
				{'e', 1.0},
			},
			"el": {
				{'i', 0.2},
				{'k', 0.5333},
				{'e', 0.8},
				{'a', 1.0},
			},
			"Va": {
				{'n', 0.3333},
				{'r', 0.6667},
				{'l', 1.0},
			},
			"ag": {
				{'u', 1.0},
			},
			"ki": {
				{0, 1.0},
			},
			"rk": {
				{'a', 0.6},
				{'i', 1.0},
			},
			"lk": {
				{'a', 0.9091},
				{'i', 1.0},
			},
			"nk": {
				{'a', 0.75},
				{'i', 1.0},
			},
			"et": {
				{0, 1.0},
			},
			"sh": {
				{'k', 0.5455},
				{'e', 0.6364},
				{'a', 0.8182},
				{'i', 1.0},
			},
			"ur": {
				{0, 1.0},
			},
			"lu": {
				{'r', 1.0},
			},
			"hi": {
				{'k', 1.0},
			},
			"hk": {
				{'i', 0.3333},
				{'a', 1.0},
			},
			"Ka": {
				{'l', 1.0},
			},
			"Ke": {
				{'s', 1.0},
			},
			"Sa": {
				{'l', 1.0},
			},
			"ni": {
				{'k', 1.0},
			},
			"Na": {
				{'l', 0.25},
				{'n', 0.5},
				{'g', 0.75},
				{'r', 1.0},
			},
			"nl": {
				{'a', 1.0},
			},
			"_A": {
				{'g', 0.1429},
				{'n', 1.0},
			},
			"__": {
				{'K', 0.0444},
				{'E', 0.1111},
				{'L', 0.1333},
				{'V', 0.2444},
				{'N', 0.3444},
				{'B', 0.4667},
				{'S', 0.5667},
				{'F', 0.6333},
				{'A', 0.7111},
				{'Y', 0.8111},
				{'D', 0.9222},
				{'J', 1.0},
			},
			"_F": {
				{'a', 0.6667},
				{'e', 1.0},
			},
			"ne": {
				{'t', 0.625},
				{'n', 1.0},
			},
			"gu": {
				{'r', 1.0},
			},
			"ka": {
				{0, 1.0},
			},
		}
	case 'M':
		return map[string][]nameFragment{
			"Da": {
				{'k', 0.25},
				{'r', 0.5},
				{'n', 1.0},
			},
			"no": {
				{'m', 0.25},
				{'k', 0.5},
				{'n', 0.75},
				{'g', 1.0},
			},
			"go": {
				{0, 1.0},
			},
			"Ra": {
				{'n', 0.5},
				{'r', 1.0},
			},
			"um": {
				{0, 1.0},
			},
			"Ro": {
				{'l', 1.0},
			},
			"Zu": {
				{'r', 1.0},
			},
			"Ho": {
				{'r', 0.5},
				{'g', 0.75},
				{'k', 1.0},
			},
			"on": {
				{'e', 0.0833},
				{'o', 0.3333},
				{'a', 0.5833},
				{'u', 0.6667},
				{'g', 0.75},
				{0, 1.0},
			},
			"ng": {
				{'o', 0.1667},
				{'u', 0.75},
				{'a', 1.0},
			},
			"ko": {
				{'m', 1.0},
			},
			"es": {
				{'k', 1.0},
			},
			"Jo": {
				{'r', 0.75},
				{'l', 1.0},
			},
			"ug": {
				{0, 1.0},
			},
			"Br": {
				{'a', 0.2857},
				{'o', 0.8571},
				{'e', 1.0},
			},
			"ur": {
				{'a', 0.2857},
				{'u', 0.4286},
				{'o', 0.7143},
				{'e', 0.8571},
				{'g', 1.0},
			},
			"lu": {
				{'m', 0.3333},
				{'g', 1.0},
			},
			"Ka": {
				{'r', 0.5},
				{'n', 1.0},
			},
			"_T": {
				{'a', 0.4},
				{'u', 0.5},
				{'o', 0.7},
				{'r', 1.0},
			},
			"nu": {
				{'m', 0.375},
				{'g', 1.0},
			},
			"Na": {
				{'n', 0.75},
				{'r', 1.0},
			},
			"lg": {
				{'u', 0.5},
				{'a', 1.0},
			},
			"ra": {
				{'m', 0.1333},
				{'k', 0.3333},
				{'n', 1.0},
			},
			"ru": {
				{'g', 0.5},
				{'m', 1.0},
			},
			"Gr": {
				{'e', 0.1111},
				{'o', 0.5556},
				{'u', 0.6667},
				{'a', 1.0},
			},
			"No": {
				{'l', 1.0},
			},
			"Ke": {
				{'k', 1.0},
			},
			"Dr": {
				{'o', 1.0},
			},
			"__": {
				{'G', 0.1635},
				{'D', 0.2788},
				{'J', 0.3558},
				{'H', 0.4327},
				{'B', 0.5481},
				{'T', 0.6442},
				{'N', 0.7019},
				{'Z', 0.7692},
				{'K', 0.9423},
				{'R', 1.0},
			},
			"ka": {
				{'m', 0.3333},
				{'k', 1.0},
			},
			"ne": {
				{'s', 0.3333},
				{'k', 1.0},
			},
			"gu": {
				{0, 1.0},
			},
			"ku": {
				{'m', 0.5},
				{'g', 1.0},
			},
			"ol": {
				{'e', 0.1667},
				{'o', 0.5833},
				{'u', 0.8333},
				{'g', 1.0},
			},
			"rg": {
				{'u', 0.1429},
				{'a', 0.2857},
				{'o', 1.0},
			},
			"Du": {
				{'r', 1.0},
			},
			"am": {
				{0, 1.0},
			},
			"Ta": {
				{'m', 0.25},
				{'n', 1.0},
			},
			"Ge": {
				{'k', 1.0},
			},
			"Bo": {
				{'g', 0.3333},
				{'r', 0.6667},
				{'k', 1.0},
			},
			"_Z": {
				{'a', 0.2857},
				{'u', 0.4286},
				{'o', 1.0},
			},
			"_D": {
				{'o', 0.1667},
				{'e', 0.25},
				{'a', 0.5833},
				{'u', 0.6667},
				{'r', 1.0},
			},
			"Go": {
				{'k', 0.25},
				{'r', 1.0},
			},
			"Tu": {
				{'r', 1.0},
			},
			"Kr": {
				{'o', 0.3333},
				{'a', 1.0},
			},
			"le": {
				{'s', 0.5},
				{'k', 1.0},
			},
			"Ga": {
				{'n', 1.0},
			},
			"_R": {
				{'a', 0.3333},
				{'u', 0.5},
				{'e', 0.6667},
				{'o', 1.0},
			},
			"Tr": {
				{'o', 0.6667},
				{'a', 1.0},
			},
			"_G": {
				{'r', 0.5294},
				{'o', 0.7647},
				{'e', 0.8235},
				{'a', 0.9412},
				{'u', 1.0},
			},
			"ak": {
				{0, 1.0},
			},
			"_K": {
				{'r', 0.3333},
				{'a', 0.5556},
				{'o', 0.9444},
				{'e', 1.0},
			},
			"Gu": {
				{'r', 1.0},
			},
			"sk": {
				{0, 1.0},
			},
			"Do": {
				{'r', 1.0},
			},
			"na": {
				{'m', 0.5},
				{'k', 1.0},
			},
			"Ba": {
				{'n', 1.0},
			},
			"an": {
				{'g', 0.1379},
				{'n', 0.6897},
				{'o', 0.7586},
				{'e', 0.8621},
				{'u', 1.0},
			},
			"Zo": {
				{'r', 0.25},
				{'n', 0.75},
				{'k', 1.0},
			},
			"lo": {
				{'k', 0.4},
				{'m', 0.6},
				{'n', 0.8},
				{'g', 1.0},
			},
			"re": {
				{'s', 0.5},
				{'f', 0.625},
				{'k', 1.0},
			},
			"Ja": {
				{'m', 0.25},
				{'k', 0.5},
				{'n', 1.0},
			},
			"_N": {
				{'a', 0.6667},
				{'o', 1.0},
			},
			"_H": {
				{'o', 0.5},
				{'u', 0.75},
				{'a', 1.0},
			},
			"_J": {
				{'a', 0.5},
				{'o', 1.0},
			},
			"ef": {
				{0, 1.0},
			},
			"Bu": {
				{'r', 1.0},
			},
			"ok": {
				{0, 0.5},
				{'o', 0.5625},
				{'e', 0.6875},
				{'u', 0.8125},
				{'a', 1.0},
			},
			"ek": {
				{0, 1.0},
			},
			"or": {
				{'o', 0.0667},
				{'e', 0.3333},
				{'u', 0.6},
				{'a', 0.7333},
				{'g', 1.0},
			},
			"og": {
				{0, 1.0},
			},
			"To": {
				{'r', 0.5},
				{'k', 1.0},
			},
			"om": {
				{0, 1.0},
			},
			"De": {
				{'s', 1.0},
			},
			"_B": {
				{'u', 0.0833},
				{'a', 0.1667},
				{'o', 0.4167},
				{'r', 1.0},
			},
			"Hu": {
				{'r', 1.0},
			},
			"nn": {
				{'g', 0.4375},
				{'e', 0.5625},
				{'o', 0.75},
				{'a', 0.8125},
				{'u', 1.0},
			},
			"Ha": {
				{'r', 0.5},
				{'n', 1.0},
			},
			"Ko": {
				{'l', 0.5714},
				{'k', 0.7143},
				{'r', 1.0},
			},
			"ro": {
				{'g', 0.0952},
				{'n', 0.4286},
				{'m', 0.5714},
				{'k', 0.8571},
				{'l', 1.0},
			},
			"Za": {
				{'n', 1.0},
			},
			"ar": {
				{'g', 0.3333},
				{'o', 0.6667},
				{'e', 0.8333},
				{'a', 1.0},
			},
			"ga": {
				{0, 1.0},
			},
			"Ru": {
				{'g', 1.0},
			},
			"Re": {
				{'k', 1.0},
			},
			"ke": {
				{'k', 1.0},
			},
		}
	default:
		return nil
	}
}
//
// End of generated data
//
