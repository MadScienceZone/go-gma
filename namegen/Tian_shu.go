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
// TianShu describes the naming conventions for the Tian-shu
// culture. Its methods give further details, but generally speaking
// the main operation to perform on these types is to just call the
// Generate and GenerateWithSurnames methods to create new names which
// conform to their cultural patterns.
//
type TianShu struct {
	BaseCulture
}

//
// prefix gives the prefix/selector string for each Tian-shu gender, or an empty
// string if one is not defined.
//
func (c TianShu) prefix(gender rune) string {
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
// defaultMinMax returns the minimum and maximum size of Tian-shu names based on gender.
//
func (c TianShu) defaultMinMax(gender rune) (int, int) {
	switch gender {
	case 'F':
		return 2, 8
	case 'M':
		return 2, 8
	case 'S':
		return 2, 8
	default:
		return 1, 1
	}
}

//
// Genders returns the set of genders defined for the Tian-shu culture.
//
func (c TianShu) Genders() []rune {
	return []rune{'F', 'M', 'S'}
}

//
// HasSurnames returns true if the Tian-shu culture defines surnames.
//
func (c TianShu) HasSurnames() bool {
	return true
}

//
// Name returns the name of the culture, i.e., "Tian-shu".
//
func (c TianShu) Name() string {
	return "Tian-shu"
}

//
// HasGender returns true if the specified gender code is defined
// in the Tian-shu culture.
//
func (c TianShu) HasGender(gender rune) bool {
	switch gender {
	case 'F', 'M', 'S':
		return true
	default:
		return false
	}
}

//
// db returns the name data for the given gender in the Tian-shu culture.
//
func (c TianShu) db(gender rune) map[string][]nameFragment {
	switch gender {
	case 'F':
		return map[string][]nameFragment{
			"_Z": {
				{'a', 0.1667},
				{'h', 0.8333},
				{'i', 1.0},
			},
			"__": {
				{'Z', 0.0355},
				{'C', 0.0888},
				{'A', 0.1065},
				{'P', 0.1302},
				{'J', 0.2071},
				{'M', 0.2722},
				{'S', 0.3609},
				{'R', 0.4024},
				{'Q', 0.432},
				{'K', 0.4497},
				{'T', 0.5089},
				{'L', 0.5858},
				{'B', 0.6036},
				{'D', 0.6509},
				{'W', 0.6864},
				{'G', 0.7041},
				{'E', 0.7101},
				{'Y', 0.787},
				{'X', 0.858},
				{'F', 0.8876},
				{'N', 0.9231},
				{'H', 1.0},
			},
			"Je": {
				{'a', 1.0},
			},
			"We": {
				{'i', 0.3333},
				{'n', 1.0},
			},
			"Fe": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"as": {
				{'h', 1.0},
			},
			"ze": {
				{0, 1.0},
			},
			"uh": {
				{0, 1.0},
			},
			"Su": {
				{0, 0.3333},
				{'n', 0.6667},
				{'e', 1.0},
			},
			"Qi": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"Sz": {
				{'u', 0.5},
				{'e', 1.0},
			},
			"Ju": {
				{'n', 0.5},
				{'e', 0.75},
				{'i', 1.0},
			},
			"ae": {
				{0, 1.0},
			},
			"_L": {
				{'a', 0.3077},
				{'u', 0.3846},
				{'i', 0.9231},
				{'e', 1.0},
			},
			"Mo": {
				{0, 1.0},
			},
			"An": {
				{0, 1.0},
			},
			"Fu": {
				{0, 1.0},
			},
			"Tu": {
				{'n', 1.0},
			},
			"en": {
				{0, 0.7143},
				{'g', 1.0},
			},
			"Ha": {
				{'i', 1.0},
			},
			"oe": {
				{'i', 1.0},
			},
			"_G": {
				{'u', 0.3333},
				{'a', 0.6667},
				{'s', 1.0},
			},
			"em": {
				{'g', 1.0},
			},
			"hw": {
				{'u', 1.0},
			},
			"Er": {
				{0, 1.0},
			},
			"na": {
				{0, 0.8},
				{'l', 1.0},
			},
			"Si": {
				{'a', 1.0},
			},
			"_Y": {
				{'e', 0.0769},
				{'o', 0.1538},
				{'u', 0.5385},
				{'i', 0.7692},
				{'a', 1.0},
			},
			"ei": {
				{0, 1.0},
			},
			"ai": {
				{0, 0.8889},
				{'k', 1.0},
			},
			"Te": {
				{0, 1.0},
			},
			"gl": {
				{'u', 1.0},
			},
			"uo": {
				{0, 1.0},
			},
			"Gu": {
				{'i', 1.0},
			},
			"Zh": {
				{'a', 0.25},
				{'u', 0.5},
				{'i', 0.75},
				{'e', 1.0},
			},
			"ea": {
				{'k', 1.0},
			},
			"Tz": {
				{'u', 1.0},
			},
			"Yu": {
				{'n', 0.4},
				{'e', 0.6},
				{0, 0.8},
				{'a', 1.0},
			},
			"ay": {
				{0, 1.0},
			},
			"he": {
				{'n', 1.0},
			},
			"Da": {
				{0, 0.3333},
				{'n', 0.6667},
				{'o', 1.0},
			},
			"_W": {
				{'a', 0.3333},
				{'e', 0.8333},
				{'o', 1.0},
			},
			"_Q": {
				{'i', 0.4},
				{'u', 1.0},
			},
			"Xu": {
				{'e', 0.6667},
				{0, 1.0},
			},
			"Ga": {
				{'i', 1.0},
			},
			"ie": {
				{'n', 1.0},
			},
			"sc": {
				{'h', 1.0},
			},
			"My": {
				{0, 1.0},
			},
			"de": {
				{0, 1.0},
			},
			"Ta": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"ia": {
				{0, 0.2222},
				{'o', 0.5556},
				{'n', 1.0},
			},
			"ak": {
				{0, 1.0},
			},
			"Xi": {
				{'u', 0.1111},
				{'n', 0.3333},
				{'o', 0.4444},
				{'a', 0.7778},
				{'d', 0.8889},
				{0, 1.0},
			},
			"_C": {
				{'a', 0.1111},
				{'h', 0.8889},
				{'i', 1.0},
			},
			"zu": {
				{0, 1.0},
			},
			"Me": {
				{'m', 0.5},
				{'i', 1.0},
			},
			"_A": {
				{'h', 0.3333},
				{'i', 0.6667},
				{'n', 1.0},
			},
			"la": {
				{0, 1.0},
			},
			"eu": {
				{'h', 1.0},
			},
			"Ca": {
				{'i', 1.0},
			},
			"Ki": {
				{'t', 1.0},
			},
			"ue": {
				{0, 0.5},
				{'r', 0.6667},
				{'t', 0.8333},
				{'i', 1.0},
			},
			"Ho": {
				{0, 0.5},
				{'n', 1.0},
			},
			"su": {
				{'i', 1.0},
			},
			"ao": {
				{0, 0.9091},
				{'m', 1.0},
			},
			"on": {
				{'a', 0.1429},
				{'g', 1.0},
			},
			"wu": {
				{'n', 1.0},
			},
			"Ku": {
				{'e', 1.0},
			},
			"it": {
				{'l', 1.0},
			},
			"Be": {
				{'n', 1.0},
			},
			"_R": {
				{'o', 0.2857},
				{'u', 0.8571},
				{'a', 1.0},
			},
			"_B": {
				{'i', 0.3333},
				{'e', 0.6667},
				{'a', 1.0},
			},
			"se": {
				{0, 0.5},
				{'u', 1.0},
			},
			"Gs": {
				{'c', 1.0},
			},
			"Le": {
				{'i', 1.0},
			},
			"oh": {
				{0, 1.0},
			},
			"Pi": {
				{'n', 1.0},
			},
			"Ma": {
				{'e', 0.3333},
				{'o', 0.6667},
				{'y', 1.0},
			},
			"Ra": {
				{'n', 1.0},
			},
			"Ro": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"Sa": {
				{'n', 0.5},
				{0, 1.0},
			},
			"al": {
				{'a', 1.0},
			},
			"Wa": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"_N": {
				{'i', 0.3333},
				{'u', 0.6667},
				{'a', 1.0},
			},
			"_T": {
				{'s', 0.4},
				{'a', 0.6},
				{'z', 0.7},
				{'u', 0.8},
				{'i', 0.9},
				{'e', 1.0},
			},
			"Ru": {
				{0, 0.25},
				{'i', 1.0},
			},
			"Ch": {
				{'i', 0.2857},
				{'u', 0.7143},
				{'e', 0.8571},
				{'w', 1.0},
			},
			"_P": {
				{'h', 0.25},
				{'i', 0.5},
				{'e', 0.75},
				{'a', 1.0},
			},
			"Hs": {
				{'e', 0.25},
				{'u', 0.5},
				{'i', 1.0},
			},
			"Ah": {
				{0, 1.0},
			},
			"De": {
				{0, 1.0},
			},
			"ne": {
				{0, 1.0},
			},
			"Ya": {
				{'n', 0.3333},
				{0, 0.6667},
				{'s', 1.0},
			},
			"_H": {
				{'s', 0.3077},
				{'a', 0.3846},
				{'e', 0.4615},
				{'o', 0.6154},
				{'u', 1.0},
			},
			"sa": {
				{'i', 0.5},
				{'o', 1.0},
			},
			"Di": {
				{'n', 1.0},
			},
			"id": {
				{'e', 1.0},
			},
			"sh": {
				{'a', 1.0},
			},
			"Mi": {
				{'a', 0.3333},
				{'n', 1.0},
			},
			"ik": {
				{0, 1.0},
			},
			"Ph": {
				{'a', 1.0},
			},
			"Sh": {
				{'i', 0.1667},
				{'u', 0.3333},
				{'o', 0.6667},
				{'a', 1.0},
			},
			"Ba": {
				{'o', 1.0},
			},
			"Ni": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"Pa": {
				{'o', 1.0},
			},
			"om": {
				{'i', 1.0},
			},
			"et": {
				{0, 1.0},
			},
			"Ai": {
				{0, 1.0},
			},
			"Hu": {
				{'a', 0.6},
				{0, 0.8},
				{'i', 1.0},
			},
			"li": {
				{0, 0.5},
				{'n', 1.0},
			},
			"Jy": {
				{0, 1.0},
			},
			"Ye": {
				{'h', 1.0},
			},
			"_D": {
				{'a', 0.375},
				{'i', 0.625},
				{'e', 0.75},
				{'o', 1.0},
			},
			"Ja": {
				{'i', 1.0},
			},
			"un": {
				{'g', 0.4444},
				{0, 1.0},
			},
			"Yi": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"il": {
				{'i', 1.0},
			},
			"Ci": {
				{0, 1.0},
			},
			"er": {
				{0, 1.0},
			},
			"an": {
				{'a', 0.0455},
				{'g', 0.3182},
				{0, 1.0},
			},
			"Zi": {
				{0, 1.0},
			},
			"ua": {
				{0, 0.25},
				{'n', 1.0},
			},
			"Bi": {
				{'k', 1.0},
			},
			"La": {
				{'i', 0.25},
				{'o', 0.5},
				{'n', 1.0},
			},
			"tl": {
				{'i', 1.0},
			},
			"_J": {
				{'u', 0.3077},
				{'i', 0.7692},
				{'y', 0.8462},
				{'e', 0.9231},
				{'a', 1.0},
			},
			"hu": {
				{0, 0.5},
				{'i', 0.6667},
				{'n', 0.8333},
				{'o', 1.0},
			},
			"Za": {
				{'n', 1.0},
			},
			"_E": {
				{'r', 1.0},
			},
			"Li": {
				{'a', 0.1429},
				{0, 0.2857},
				{'u', 0.4286},
				{'l', 0.5714},
				{'n', 0.8571},
				{'e', 1.0},
			},
			"io": {
				{'n', 1.0},
			},
			"Wo": {
				{'e', 1.0},
			},
			"_K": {
				{'w', 0.3333},
				{'u', 0.6667},
				{'i', 1.0},
			},
			"Qu": {
				{'i', 0.6667},
				{'n', 1.0},
			},
			"wa": {
				{'n', 1.0},
			},
			"_X": {
				{'i', 0.75},
				{'u', 1.0},
			},
			"Mu": {
				{0, 1.0},
			},
			"Fa": {
				{'n', 1.0},
			},
			"hi": {
				{'n', 0.5},
				{0, 1.0},
			},
			"ou": {
				{0, 1.0},
			},
			"Yo": {
				{'n', 1.0},
			},
			"ch": {
				{'u', 1.0},
			},
			"ho": {
				{'s', 0.3333},
				{'u', 0.6667},
				{'n', 1.0},
			},
			"_F": {
				{'u', 0.2},
				{'o', 0.4},
				{'e', 0.8},
				{'a', 1.0},
			},
			"mi": {
				{'n', 1.0},
			},
			"mg": {
				{0, 1.0},
			},
			"lu": {
				{0, 1.0},
			},
			"Lu": {
				{0, 1.0},
			},
			"Ji": {
				{'a', 0.5},
				{0, 0.6667},
				{'n', 1.0},
			},
			"_S": {
				{'i', 0.0667},
				{'u', 0.2667},
				{'h', 0.6667},
				{'o', 0.7333},
				{'a', 0.8667},
				{'z', 1.0},
			},
			"os": {
				{'h', 1.0},
			},
			"Nu": {
				{'i', 0.5},
				{0, 1.0},
			},
			"Pe": {
				{'i', 1.0},
			},
			"_M": {
				{'y', 0.0909},
				{'u', 0.1818},
				{'i', 0.4545},
				{'e', 0.6364},
				{'o', 0.7273},
				{'a', 1.0},
			},
			"iu": {
				{0, 1.0},
			},
			"si": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"So": {
				{'n', 1.0},
			},
			"Ts": {
				{'a', 0.5},
				{'e', 0.75},
				{'u', 1.0},
			},
			"ng": {
				{'l', 0.0312},
				{0, 1.0},
			},
			"Fo": {
				{'h', 1.0},
			},
			"eh": {
				{'o', 1.0},
			},
			"Na": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"Ti": {
				{'n', 1.0},
			},
			"He": {
				{0, 1.0},
			},
			"Do": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"Kw": {
				{'a', 1.0},
			},
			"ui": {
				{0, 0.75},
				{'n', 1.0},
			},
			"ha": {
				{'n', 0.5},
				{'o', 0.8333},
				{0, 1.0},
			},
			"in": {
				{'n', 0.0417},
				{'e', 0.0833},
				{0, 0.3333},
				{'g', 0.9167},
				{'a', 1.0},
			},
			"nn": {
				{'a', 1.0},
			},
		}
	case 'M':
		return map[string][]nameFragment{
			"Tu": {
				{'o', 0.3333},
				{'n', 0.6667},
				{0, 1.0},
			},
			"to": {
				{0, 1.0},
			},
			"Bu": {
				{0, 1.0},
			},
			"An": {
				{0, 1.0},
			},
			"Fu": {
				{0, 1.0},
			},
			"ec": {
				{'h', 1.0},
			},
			"_L": {
				{'i', 0.45},
				{'e', 0.55},
				{'o', 0.6},
				{'a', 0.85},
				{'u', 1.0},
			},
			"pe": {
				{'n', 1.0},
			},
			"Mo": {
				{0, 1.0},
			},
			"we": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"Sz": {
				{'e', 1.0},
			},
			"Ju": {
				{'n', 0.5},
				{0, 1.0},
			},
			"ol": {
				{0, 1.0},
			},
			"Qi": {
				{'a', 0.25},
				{0, 0.5},
				{'n', 1.0},
			},
			"ij": {
				{'u', 1.0},
			},
			"Su": {
				{'n', 0.5},
				{0, 1.0},
			},
			"ze": {
				{'t', 0.3333},
				{0, 1.0},
			},
			"Fe": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"Re": {
				{'n', 1.0},
			},
			"Je": {
				{'n', 1.0},
			},
			"__": {
				{'E', 0.0096},
				{'L', 0.0735},
				{'G', 0.1118},
				{'P', 0.1438},
				{'T', 0.2332},
				{'M', 0.2652},
				{'I', 0.2684},
				{'Y', 0.345},
				{'O', 0.3482},
				{'F', 0.3706},
				{'h', 0.3738},
				{'C', 0.4569},
				{'B', 0.492},
				{'R', 0.4984},
				{'N', 0.5048},
				{'K', 0.5847},
				{'X', 0.6294},
				{'D', 0.6645},
				{'Z', 0.7093},
				{'W', 0.7348},
				{'J', 0.7891},
				{'H', 0.8914},
				{'A', 0.901},
				{'Q', 0.9297},
				{'S', 1.0},
			},
			"We": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"_Z": {
				{'e', 0.1429},
				{'h', 0.7143},
				{'i', 0.7857},
				{'o', 0.8571},
				{'a', 0.9286},
				{'u', 1.0},
			},
			"ic": {
				{0, 1.0},
			},
			"he": {
				{'h', 0.0909},
				{'e', 0.1818},
				{'n', 0.7273},
				{0, 0.8182},
				{'u', 1.0},
			},
			"Go": {
				{'n', 1.0},
			},
			"is": {
				{'u', 1.0},
			},
			"Ka": {
				{'o', 0.4},
				{'i', 0.6},
				{0, 0.8},
				{'n', 1.0},
			},
			"Tz": {
				{'e', 0.5},
				{'u', 1.0},
			},
			"ay": {
				{0, 1.0},
			},
			"Yu": {
				{'a', 0.2},
				{'k', 0.4},
				{0, 0.6},
				{'n', 0.8},
				{'e', 1.0},
			},
			"Zh": {
				{'a', 0.25},
				{'u', 0.5},
				{'i', 0.625},
				{'e', 0.875},
				{'o', 1.0},
			},
			"Gh": {
				{'i', 1.0},
			},
			"Gu": {
				{'o', 0.2},
				{'i', 0.4},
				{0, 0.6},
				{'a', 1.0},
			},
			"uo": {
				{'n', 0.1429},
				{0, 1.0},
			},
			"Jh": {
				{'o', 1.0},
			},
			"ei": {
				{0, 1.0},
			},
			"Jo": {
				{0, 0.3333},
				{'y', 0.6667},
				{'o', 1.0},
			},
			"Te": {
				{0, 0.5},
				{'h', 1.0},
			},
			"ai": {
				{'h', 0.0556},
				{'f', 0.1111},
				{'s', 0.1667},
				{'o', 0.2222},
				{0, 1.0},
			},
			"_Y": {
				{'a', 0.1667},
				{'u', 0.375},
				{'i', 0.5833},
				{'e', 0.75},
				{'o', 1.0},
			},
			"Bo": {
				{'r', 1.0},
			},
			"Gi": {
				{'n', 1.0},
			},
			"Si": {
				{'u', 0.5},
				{0, 1.0},
			},
			"Er": {
				{0, 1.0},
			},
			"eo": {
				{'w', 1.0},
			},
			"_G": {
				{'h', 0.0833},
				{'i', 0.1667},
				{'o', 0.25},
				{'a', 0.5833},
				{'u', 1.0},
			},
			"Ze": {
				{'e', 0.5},
				{0, 1.0},
			},
			"en": {
				{0, 0.5909},
				{'g', 1.0},
			},
			"Ha": {
				{'n', 0.4},
				{'i', 0.6},
				{'r', 0.8},
				{'o', 1.0},
			},
			"np": {
				{'e', 1.0},
			},
			"gw": {
				{'u', 1.0},
			},
			"it": {
				{'w', 1.0},
			},
			"Ku": {
				{'e', 0.2},
				{'n', 0.6},
				{'a', 1.0},
			},
			"wu": {
				{0, 0.5},
				{'i', 1.0},
			},
			"ao": {
				{'n', 0.0556},
				{0, 1.0},
			},
			"on": {
				{0, 0.0526},
				{'g', 1.0},
			},
			"Ki": {
				{'a', 0.3333},
				{0, 0.6667},
				{'e', 1.0},
			},
			"Ho": {
				{'o', 0.2},
				{'n', 0.6},
				{0, 0.8},
				{'u', 1.0},
			},
			"ue": {
				{0, 0.5},
				{'h', 0.75},
				{'i', 1.0},
			},
			"su": {
				{'e', 0.1667},
				{'i', 0.3333},
				{'n', 0.5},
				{0, 0.8333},
				{'a', 1.0},
			},
			"ar": {
				{'k', 1.0},
			},
			"Ca": {
				{'i', 1.0},
			},
			"eu": {
				{'n', 0.6667},
				{'k', 1.0},
			},
			"Me": {
				{'i', 1.0},
			},
			"_A": {
				{'n', 0.3333},
				{'i', 0.6667},
				{'h', 1.0},
			},
			"zu": {
				{0, 1.0},
			},
			"To": {
				{'n', 0.5},
				{0, 1.0},
			},
			"_C": {
				{'h', 0.9231},
				{'o', 0.9615},
				{'a', 1.0},
			},
			"Zo": {
				{'n', 1.0},
			},
			"Xi": {
				{'e', 0.1},
				{'o', 0.2},
				{'a', 0.5},
				{0, 0.6},
				{'n', 0.9},
				{'u', 1.0},
			},
			"ok": {
				{0, 1.0},
			},
			"ak": {
				{0, 1.0},
			},
			"rk": {
				{0, 1.0},
			},
			"Ta": {
				{'n', 0.1111},
				{0, 0.2222},
				{'k', 0.3333},
				{'o', 0.4444},
				{'i', 0.8889},
				{'t', 1.0},
			},
			"ia": {
				{'n', 0.65},
				{0, 0.75},
				{'h', 0.8},
				{'o', 1.0},
			},
			"tw": {
				{'e', 1.0},
			},
			"de": {
				{0, 1.0},
			},
			"Ke": {
				{'u', 0.2},
				{0, 0.4},
				{'n', 0.6},
				{'e', 0.8},
				{'i', 1.0},
			},
			"oo": {
				{'k', 0.2},
				{'n', 0.4},
				{0, 1.0},
			},
			"ie": {
				{0, 0.25},
				{'n', 0.875},
				{'h', 1.0},
			},
			"Xu": {
				{0, 0.25},
				{'n', 0.5},
				{'a', 0.75},
				{'e', 1.0},
			},
			"Ga": {
				{'o', 0.25},
				{'h', 0.5},
				{'n', 1.0},
			},
			"Da": {
				{'n', 0.25},
				{0, 0.5},
				{'i', 0.75},
				{'o', 1.0},
			},
			"_W": {
				{'a', 0.5},
				{'u', 0.625},
				{'i', 0.75},
				{'e', 1.0},
			},
			"_Q": {
				{'i', 0.4444},
				{'u', 1.0},
			},
			"_O": {
				{'n', 1.0},
			},
			"_P": {
				{'e', 0.2},
				{'i', 0.5},
				{'o', 0.7},
				{'a', 0.9},
				{'u', 1.0},
			},
			"ac": {
				{'h', 1.0},
			},
			"Ru": {
				{'n', 1.0},
			},
			"Ch": {
				{'e', 0.2083},
				{'i', 0.5833},
				{'o', 0.6667},
				{'a', 0.75},
				{'u', 1.0},
			},
			"_h": {
				{'o', 1.0},
			},
			"_T": {
				{'s', 0.1429},
				{'i', 0.3214},
				{'y', 0.3571},
				{'e', 0.4286},
				{'u', 0.5357},
				{'z', 0.6071},
				{'o', 0.6786},
				{'a', 1.0},
			},
			"_N": {
				{'i', 1.0},
			},
			"Du": {
				{0, 1.0},
			},
			"Wa": {
				{'n', 0.5},
				{'y', 0.75},
				{'i', 1.0},
			},
			"Sa": {
				{'n', 1.0},
			},
			"fa": {
				{'t', 1.0},
			},
			"fu": {
				{0, 1.0},
			},
			"On": {
				{0, 1.0},
			},
			"Pi": {
				{'e', 0.3333},
				{0, 0.6667},
				{'n', 1.0},
			},
			"Ma": {
				{'o', 0.25},
				{'n', 0.75},
				{0, 1.0},
			},
			"do": {
				{0, 1.0},
			},
			"se": {
				{0, 1.0},
			},
			"Le": {
				{0, 0.5},
				{'e', 1.0},
			},
			"_B": {
				{'a', 0.3636},
				{'u', 0.4545},
				{'i', 0.8182},
				{'e', 0.9091},
				{'o', 1.0},
			},
			"or": {
				{0, 1.0},
			},
			"_R": {
				{'u', 0.5},
				{'e', 1.0},
			},
			"Be": {
				{'i', 1.0},
			},
			"Sh": {
				{'o', 0.0833},
				{'e', 0.3333},
				{'i', 0.5833},
				{'u', 0.75},
				{'a', 1.0},
			},
			"ya": {
				{'n', 1.0},
			},
			"Mi": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"Se": {
				{'e', 1.0},
			},
			"Di": {
				{'n', 1.0},
			},
			"Zu": {
				{'o', 1.0},
			},
			"Hw": {
				{'u', 0.5},
				{'e', 1.0},
			},
			"_H": {
				{'u', 0.2188},
				{'a', 0.375},
				{'o', 0.5312},
				{'s', 0.7812},
				{'e', 0.875},
				{'w', 0.9375},
				{'i', 1.0},
			},
			"gd": {
				{'o', 0.5},
				{'e', 1.0},
			},
			"De": {
				{'c', 0.5},
				{0, 1.0},
			},
			"Ya": {
				{'n', 0.5},
				{'t', 0.75},
				{'o', 1.0},
			},
			"Ah": {
				{0, 1.0},
			},
			"Hs": {
				{'u', 0.5},
				{'i', 1.0},
			},
			"Zi": {
				{0, 1.0},
			},
			"ua": {
				{'c', 0.0714},
				{'i', 0.1429},
				{0, 0.2857},
				{'n', 1.0},
			},
			"Bi": {
				{0, 0.25},
				{'n', 0.75},
				{'a', 1.0},
			},
			"_I": {
				{0, 1.0},
			},
			"an": {
				{'g', 0.4898},
				{'n', 0.5102},
				{0, 1.0},
			},
			"ah": {
				{0, 1.0},
			},
			"Ko": {
				{'n', 0.3333},
				{'l', 0.6667},
				{'i', 1.0},
			},
			"Yi": {
				{'c', 0.2},
				{'f', 0.4},
				{'n', 0.8},
				{0, 1.0},
			},
			"Ja": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"ta": {
				{'i', 1.0},
			},
			"un": {
				{'g', 0.4706},
				{0, 1.0},
			},
			"_D": {
				{'o', 0.2727},
				{'i', 0.3636},
				{'e', 0.5455},
				{'u', 0.6364},
				{'a', 1.0},
			},
			"Ye": {
				{0, 0.25},
				{'n', 0.5},
				{'o', 0.75},
				{'e', 1.0},
			},
			"ee": {
				{'t', 0.1667},
				{0, 1.0},
			},
			"Ai": {
				{0, 1.0},
			},
			"Hu": {
				{'a', 0.4286},
				{'n', 0.5714},
				{0, 0.7143},
				{'i', 0.8571},
				{'o', 1.0},
			},
			"so": {
				{0, 1.0},
			},
			"et": {
				{'a', 0.5},
				{'o', 1.0},
			},
			"Pa": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"Ni": {
				{0, 0.5},
				{'n', 1.0},
			},
			"at": {
				{0, 1.0},
			},
			"Ba": {
				{'i', 0.5},
				{'o', 0.75},
				{'n', 1.0},
			},
			"_F": {
				{'e', 0.2857},
				{'o', 0.5714},
				{'a', 0.8571},
				{'u', 1.0},
			},
			"ho": {
				{0, 0.125},
				{'n', 0.5},
				{'u', 0.625},
				{'i', 1.0},
			},
			"Po": {
				{'k', 0.5},
				{0, 1.0},
			},
			"ch": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"En": {
				{'g', 0.5},
				{0, 1.0},
			},
			"ou": {
				{0, 0.8},
				{'n', 1.0},
			},
			"Yo": {
				{'u', 0.3333},
				{0, 0.5},
				{'n', 0.6667},
				{'k', 0.8333},
				{'o', 1.0},
			},
			"if": {
				{'u', 0.5},
				{'a', 1.0},
			},
			"hi": {
				{'j', 0.0714},
				{'h', 0.2143},
				{'e', 0.2857},
				{0, 0.5714},
				{'n', 0.7857},
				{'a', 1.0},
			},
			"Fa": {
				{'i', 0.5},
				{0, 1.0},
			},
			"ih": {
				{0, 0.6667},
				{'o', 1.0},
			},
			"Mu": {
				{'n', 1.0},
			},
			"Lo": {
				{'o', 1.0},
			},
			"Hi": {
				{0, 0.5},
				{'a', 1.0},
			},
			"Kh": {
				{'o', 0.5},
				{'a', 1.0},
			},
			"wa": {
				{'n', 1.0},
			},
			"_X": {
				{'u', 0.2857},
				{'i', 1.0},
			},
			"Qu": {
				{0, 0.2},
				{'a', 0.4},
				{'o', 0.6},
				{'i', 1.0},
			},
			"ju": {
				{0, 1.0},
			},
			"_K": {
				{'h', 0.08},
				{'i', 0.2},
				{'e', 0.4},
				{'w', 0.48},
				{'o', 0.6},
				{'a', 0.8},
				{'u', 1.0},
			},
			"io": {
				{'n', 0.6667},
				{0, 1.0},
			},
			"_E": {
				{'n', 0.6667},
				{'r', 1.0},
			},
			"Li": {
				{'e', 0.2222},
				{'u', 0.3333},
				{'n', 0.4444},
				{0, 0.5556},
				{'a', 1.0},
			},
			"Pu": {
				{0, 1.0},
			},
			"ut": {
				{0, 1.0},
			},
			"hu": {
				{'k', 0.1},
				{'a', 0.4},
				{0, 0.6},
				{'n', 0.8},
				{'t', 0.9},
				{'o', 1.0},
			},
			"Za": {
				{'i', 1.0},
			},
			"_J": {
				{'o', 0.1765},
				{'i', 0.6471},
				{'h', 0.7059},
				{'e', 0.7647},
				{'u', 0.8824},
				{'a', 1.0},
			},
			"La": {
				{'u', 0.2},
				{'n', 0.6},
				{'o', 0.8},
				{'i', 1.0},
			},
			"in": {
				{'p', 0.04},
				{'g', 0.56},
				{0, 1.0},
			},
			"ha": {
				{'n', 0.4444},
				{'o', 0.7778},
				{'y', 0.8889},
				{'i', 1.0},
			},
			"uk": {
				{0, 1.0},
			},
			"nn": {
				{0, 1.0},
			},
			"ui": {
				{0, 0.8333},
				{'a', 1.0},
			},
			"oi": {
				{0, 1.0},
			},
			"Kw": {
				{'a', 1.0},
			},
			"He": {
				{'n', 0.6667},
				{0, 1.0},
			},
			"Ti": {
				{'e', 0.2},
				{'a', 0.6},
				{0, 0.8},
				{'n', 1.0},
			},
			"Ty": {
				{'a', 1.0},
			},
			"Do": {
				{'u', 0.3333},
				{'n', 1.0},
			},
			"au": {
				{0, 1.0},
			},
			"ow": {
				{0, 1.0},
			},
			"Wi": {
				{'n', 1.0},
			},
			"oy": {
				{0, 1.0},
			},
			"eh": {
				{0, 1.0},
			},
			"Fo": {
				{'o', 0.5},
				{0, 1.0},
			},
			"ng": {
				{'w', 0.0137},
				{0, 0.9726},
				{'d', 1.0},
			},
			"So": {
				{'n', 1.0},
			},
			"Ts": {
				{'o', 0.25},
				{'i', 0.5},
				{'e', 0.75},
				{'u', 1.0},
			},
			"Sy": {
				{0, 1.0},
			},
			"si": {
				{'n', 0.4},
				{0, 0.6},
				{'a', 0.8},
				{'e', 1.0},
			},
			"iu": {
				{0, 1.0},
			},
			"Pe": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"_M": {
				{'u', 0.1},
				{'a', 0.5},
				{'o', 0.6},
				{'i', 0.9},
				{'e', 1.0},
			},
			"Wu": {
				{0, 1.0},
			},
			"Ji": {
				{'o', 0.125},
				{'t', 0.25},
				{'u', 0.375},
				{0, 0.5},
				{'n', 0.75},
				{'a', 1.0},
			},
			"_S": {
				{'i', 0.0909},
				{'y', 0.1364},
				{'e', 0.1818},
				{'z', 0.2727},
				{'u', 0.3636},
				{'h', 0.9091},
				{'o', 0.9545},
				{'a', 1.0},
			},
			"Lu": {
				{'o', 0.3333},
				{'n', 0.6667},
				{0, 1.0},
			},
			"Co": {
				{'n', 1.0},
			},
		}
	case 'S':
		return map[string][]nameFragment{
			"oo": {
				{0, 0.75},
				{'n', 1.0},
			},
			"ie": {
				{'n', 0.25},
				{0, 0.625},
				{'h', 0.875},
				{'u', 1.0},
			},
			"Ke": {
				{0, 1.0},
			},
			"Da": {
				{'i', 0.5},
				{0, 1.0},
			},
			"_Q": {
				{'i', 0.5714},
				{'u', 1.0},
			},
			"_W": {
				{'i', 0.125},
				{'u', 0.25},
				{'e', 0.5},
				{'o', 0.75},
				{'a', 1.0},
			},
			"Xu": {
				{0, 0.3333},
				{'n', 0.6667},
				{'e', 1.0},
			},
			"Ga": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"Xi": {
				{'e', 0.1429},
				{'u', 0.2857},
				{'a', 0.7143},
				{'n', 0.8571},
				{'o', 1.0},
			},
			"Zo": {
				{'n', 0.5},
				{'u', 1.0},
			},
			"ok": {
				{0, 1.0},
			},
			"Ta": {
				{'h', 0.2},
				{'o', 0.4},
				{'n', 1.0},
			},
			"ia": {
				{'o', 0.2},
				{'n', 0.8667},
				{0, 1.0},
			},
			"rk": {
				{0, 1.0},
			},
			"ar": {
				{0, 0.5},
				{'k', 1.0},
			},
			"eu": {
				{'n', 0.5},
				{0, 1.0},
			},
			"Ca": {
				{'i', 0.5},
				{'o', 1.0},
			},
			"Ki": {
				{'m', 0.3333},
				{'n', 1.0},
			},
			"Ho": {
				{'u', 0.2},
				{0, 0.4},
				{'r', 0.6},
				{'o', 0.8},
				{'n', 1.0},
			},
			"su": {
				{0, 0.4},
				{'n', 0.6},
				{'e', 0.8},
				{'i', 1.0},
			},
			"ue": {
				{0, 0.6},
				{'h', 0.8},
				{'n', 1.0},
			},
			"To": {
				{'n', 0.5},
				{'y', 1.0},
			},
			"ko": {
				{0, 1.0},
			},
			"_C": {
				{'o', 0.037},
				{'a', 0.1111},
				{'h', 0.963},
				{'u', 1.0},
			},
			"Me": {
				{'i', 0.3333},
				{'n', 1.0},
			},
			"_A": {
				{'u', 0.3333},
				{'n', 1.0},
			},
			"um": {
				{0, 1.0},
			},
			"Ku": {
				{'i', 0.25},
				{'m', 0.5},
				{'n', 1.0},
			},
			"ao": {
				{0, 1.0},
			},
			"on": {
				{'g', 1.0},
			},
			"_B": {
				{'i', 0.2857},
				{'o', 0.4286},
				{'a', 1.0},
			},
			"pa": {
				{'n', 1.0},
			},
			"or": {
				{0, 1.0},
			},
			"_R": {
				{'e', 1.0},
			},
			"se": {
				{'n', 1.0},
			},
			"oh": {
				{0, 1.0},
			},
			"Le": {
				{'w', 0.25},
				{'o', 0.5},
				{'e', 0.75},
				{'i', 1.0},
			},
			"Pi": {
				{0, 0.5},
				{'n', 1.0},
			},
			"Ma": {
				{0, 0.25},
				{'h', 0.5},
				{'o', 0.75},
				{'r', 1.0},
			},
			"On": {
				{'g', 1.0},
			},
			"Du": {
				{'n', 1.0},
			},
			"Wa": {
				{'n', 1.0},
			},
			"Sa": {
				{0, 0.5},
				{'n', 1.0},
			},
			"_P": {
				{'a', 0.375},
				{'h', 0.5},
				{'e', 0.75},
				{'i', 1.0},
			},
			"_O": {
				{'w', 0.3333},
				{'n', 0.6667},
				{'u', 1.0},
			},
			"_T": {
				{'o', 0.0833},
				{'a', 0.2917},
				{'h', 0.4583},
				{'u', 0.5},
				{'s', 0.7917},
				{'i', 0.9167},
				{'e', 1.0},
			},
			"_N": {
				{'a', 0.1667},
				{'i', 0.6667},
				{'g', 1.0},
			},
			"Ch": {
				{'e', 0.1739},
				{'u', 0.3478},
				{'i', 0.6522},
				{'a', 0.7826},
				{'o', 1.0},
			},
			"Re": {
				{'n', 1.0},
			},
			"Fe": {
				{'n', 1.0},
			},
			"gu": {
				{'i', 1.0},
			},
			"ep": {
				{0, 1.0},
			},
			"ew": {
				{0, 1.0},
			},
			"_Z": {
				{'e', 0.1},
				{'o', 0.3},
				{'h', 1.0},
			},
			"__": {
				{'L', 0.0933},
				{'P', 0.1231},
				{'W', 0.153},
				{'D', 0.1828},
				{'G', 0.2239},
				{'R', 0.2276},
				{'Q', 0.2537},
				{'H', 0.3396},
				{'M', 0.3843},
				{'O', 0.3955},
				{'N', 0.4179},
				{'F', 0.4515},
				{'Z', 0.4888},
				{'E', 0.4963},
				{'T', 0.5858},
				{'C', 0.6866},
				{'K', 0.7388},
				{'I', 0.7425},
				{'Y', 0.8172},
				{'A', 0.8284},
				{'S', 0.8955},
				{'B', 0.9216},
				{'J', 0.9627},
				{'X', 1.0},
			},
			"We": {
				{'n', 0.5},
				{'i', 1.0},
			},
			"Qi": {
				{'u', 0.25},
				{'n', 0.5},
				{'a', 0.75},
				{0, 1.0},
			},
			"Ju": {
				{'n', 0.3333},
				{0, 0.6667},
				{'e', 1.0},
			},
			"Cu": {
				{'i', 1.0},
			},
			"Su": {
				{'i', 0.3333},
				{'n', 0.6667},
				{0, 1.0},
			},
			"An": {
				{0, 0.5},
				{'g', 1.0},
			},
			"Fu": {
				{0, 1.0},
			},
			"_L": {
				{'e', 0.16},
				{'u', 0.32},
				{'i', 0.6},
				{'a', 0.8},
				{'o', 1.0},
			},
			"Mo": {
				{0, 0.3333},
				{'y', 0.6667},
				{'k', 1.0},
			},
			"Tu": {
				{0, 1.0},
			},
			"Si": {
				{0, 1.0},
			},
			"bu": {
				{'a', 1.0},
			},
			"Ze": {
				{'n', 1.0},
			},
			"Ou": {
				{0, 1.0},
			},
			"en": {
				{0, 0.5238},
				{'g', 1.0},
			},
			"Ha": {
				{'o', 0.2},
				{'r', 0.4},
				{'n', 1.0},
			},
			"eo": {
				{'n', 0.5},
				{'h', 1.0},
			},
			"_G": {
				{'u', 0.3636},
				{'e', 0.5455},
				{'o', 0.8182},
				{'a', 1.0},
			},
			"oe": {
				{0, 0.6667},
				{'i', 1.0},
			},
			"_Y": {
				{'a', 0.2},
				{'o', 0.25},
				{'e', 0.55},
				{'u', 0.8},
				{'i', 1.0},
			},
			"Bo": {
				{0, 1.0},
			},
			"ei": {
				{0, 1.0},
			},
			"Jo": {
				{'e', 0.3333},
				{'o', 0.6667},
				{'n', 1.0},
			},
			"Te": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"ai": {
				{0, 0.875},
				{'o', 1.0},
			},
			"Zh": {
				{'u', 0.1429},
				{'e', 0.2857},
				{'o', 0.4286},
				{'a', 1.0},
			},
			"ea": {
				{'n', 1.0},
			},
			"Yu": {
				{0, 0.2},
				{'a', 0.4},
				{'n', 0.6},
				{'e', 1.0},
			},
			"uo": {
				{0, 1.0},
			},
			"op": {
				{'a', 1.0},
			},
			"Gu": {
				{'a', 0.5},
				{0, 0.75},
				{'o', 1.0},
			},
			"he": {
				{'w', 0.1429},
				{'n', 0.7143},
				{'a', 0.8571},
				{'u', 1.0},
			},
			"nb": {
				{'u', 1.0},
			},
			"Ka": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"Go": {
				{0, 0.3333},
				{'n', 0.6667},
				{'e', 1.0},
			},
			"_J": {
				{'u', 0.2727},
				{'i', 0.7273},
				{'o', 1.0},
			},
			"hu": {
				{0, 0.25},
				{'k', 0.375},
				{'a', 0.625},
				{'n', 0.75},
				{'m', 0.875},
				{'i', 1.0},
			},
			"La": {
				{'u', 0.2},
				{'i', 0.4},
				{'m', 0.6},
				{'n', 1.0},
			},
			"Wo": {
				{'n', 0.5},
				{'o', 1.0},
			},
			"Qu": {
				{'i', 0.3333},
				{'a', 0.6667},
				{0, 1.0},
			},
			"_K": {
				{'h', 0.0714},
				{'a', 0.2143},
				{'o', 0.3571},
				{'w', 0.4286},
				{'e', 0.5},
				{'i', 0.7143},
				{'u', 1.0},
			},
			"Kh": {
				{'o', 1.0},
			},
			"wa": {
				{'n', 1.0},
			},
			"_X": {
				{'i', 0.7},
				{'u', 1.0},
			},
			"_E": {
				{'n', 0.5},
				{'c', 1.0},
			},
			"Li": {
				{'n', 0.2857},
				{'a', 0.5714},
				{0, 0.7143},
				{'u', 0.8571},
				{'m', 1.0},
			},
			"io": {
				{'n', 0.3333},
				{0, 1.0},
			},
			"hi": {
				{'u', 0.1667},
				{'e', 0.4167},
				{'o', 0.5},
				{'n', 0.6667},
				{'a', 0.8333},
				{'h', 0.9167},
				{0, 1.0},
			},
			"Lo": {
				{'u', 0.2},
				{'k', 0.4},
				{0, 0.6},
				{'h', 0.8},
				{'p', 1.0},
			},
			"ih": {
				{0, 1.0},
			},
			"Fa": {
				{0, 0.3333},
				{'n', 1.0},
			},
			"ho": {
				{'n', 0.1429},
				{'w', 0.2857},
				{'o', 0.4286},
				{'i', 0.5714},
				{'u', 0.8571},
				{'e', 1.0},
			},
			"_F": {
				{'a', 0.3333},
				{'o', 0.6667},
				{'e', 0.8889},
				{'u', 1.0},
			},
			"ou": {
				{0, 1.0},
			},
			"Yo": {
				{0, 1.0},
			},
			"En": {
				{'g', 1.0},
			},
			"Lu": {
				{'m', 0.25},
				{'n', 0.5},
				{'o', 0.75},
				{0, 1.0},
			},
			"Co": {
				{'n', 1.0},
			},
			"Wu": {
				{0, 1.0},
			},
			"_S": {
				{'o', 0.0556},
				{'h', 0.5556},
				{'a', 0.6667},
				{'i', 0.7222},
				{'u', 0.8889},
				{'e', 1.0},
			},
			"Ji": {
				{'a', 0.8},
				{'n', 1.0},
			},
			"iu": {
				{0, 1.0},
			},
			"si": {
				{'e', 0.5},
				{0, 1.0},
			},
			"_M": {
				{'a', 0.3333},
				{'o', 0.5833},
				{'e', 0.8333},
				{'i', 1.0},
			},
			"Pe": {
				{'i', 0.5},
				{'n', 1.0},
			},
			"So": {
				{'n', 1.0},
			},
			"Ts": {
				{'e', 0.1429},
				{'u', 0.5714},
				{'a', 0.8571},
				{'o', 1.0},
			},
			"au": {
				{0, 1.0},
			},
			"ow": {
				{0, 1.0},
			},
			"ng": {
				{0, 1.0},
			},
			"Fo": {
				{'u', 0.3333},
				{'n', 0.6667},
				{'k', 1.0},
			},
			"Wi": {
				{'e', 1.0},
			},
			"Ow": {
				{0, 1.0},
			},
			"oy": {
				{0, 1.0},
			},
			"eh": {
				{0, 1.0},
			},
			"Na": {
				{'n', 1.0},
			},
			"ui": {
				{'a', 0.1111},
				{0, 0.8889},
				{'e', 1.0},
			},
			"oi": {
				{0, 1.0},
			},
			"Ge": {
				{'n', 0.5},
				{0, 1.0},
			},
			"in": {
				{0, 0.4667},
				{'g', 1.0},
			},
			"ha": {
				{0, 0.0909},
				{'o', 0.3636},
				{'n', 0.9091},
				{'i', 1.0},
			},
			"nn": {
				{0, 1.0},
			},
			"uk": {
				{'o', 1.0},
			},
			"Ti": {
				{'e', 0.3333},
				{'a', 0.6667},
				{0, 1.0},
			},
			"He": {
				{0, 1.0},
			},
			"Do": {
				{'n', 1.0},
			},
			"Kw": {
				{'a', 1.0},
			},
			"De": {
				{'n', 0.5},
				{'e', 1.0},
			},
			"Ya": {
				{'o', 0.25},
				{'n', 0.75},
				{'p', 1.0},
			},
			"Hs": {
				{'a', 0.2},
				{'u', 0.6},
				{'i', 1.0},
			},
			"Di": {
				{'n', 1.0},
			},
			"_H": {
				{'a', 0.2174},
				{'o', 0.4348},
				{'e', 0.4783},
				{'u', 0.7826},
				{'s', 1.0},
			},
			"sa": {
				{'n', 0.3333},
				{'i', 1.0},
			},
			"Se": {
				{'n', 0.5},
				{'e', 1.0},
			},
			"am": {
				{0, 1.0},
			},
			"Mi": {
				{'n', 0.5},
				{0, 1.0},
			},
			"Sh": {
				{'i', 0.3333},
				{'u', 0.5556},
				{'e', 0.6667},
				{'a', 1.0},
			},
			"Au": {
				{0, 1.0},
			},
			"Ph": {
				{'a', 1.0},
			},
			"so": {
				{'n', 1.0},
			},
			"Ec": {
				{0, 1.0},
			},
			"Th": {
				{'e', 0.25},
				{'u', 0.5},
				{'i', 1.0},
			},
			"Hu": {
				{'a', 0.2857},
				{0, 0.4286},
				{'o', 0.5714},
				{'n', 0.7143},
				{'i', 1.0},
			},
			"Ba": {
				{0, 0.25},
				{'n', 0.5},
				{'o', 0.75},
				{'i', 1.0},
			},
			"Pa": {
				{'n', 0.3333},
				{'o', 0.6667},
				{'i', 1.0},
			},
			"Ni": {
				{'n', 0.3333},
				{0, 0.6667},
				{'u', 1.0},
			},
			"ap": {
				{0, 1.0},
			},
			"Ng": {
				{'u', 0.5},
				{0, 1.0},
			},
			"Ye": {
				{'i', 0.1667},
				{'e', 0.3333},
				{'n', 0.5},
				{'p', 0.6667},
				{'h', 0.8333},
				{0, 1.0},
			},
			"ee": {
				{0, 1.0},
			},
			"Yi": {
				{0, 0.25},
				{'p', 0.5},
				{'n', 1.0},
			},
			"_D": {
				{'u', 0.25},
				{'i', 0.375},
				{'e', 0.625},
				{'o', 0.75},
				{'a', 1.0},
			},
			"un": {
				{0, 0.3846},
				{'g', 1.0},
			},
			"_I": {
				{0, 1.0},
			},
			"im": {
				{0, 1.0},
			},
			"an": {
				{'n', 0.0444},
				{0, 0.5778},
				{'g', 0.9778},
				{'b', 1.0},
			},
			"ua": {
				{0, 0.3333},
				{'o', 0.4444},
				{'n', 1.0},
			},
			"Bi": {
				{'a', 0.5},
				{0, 1.0},
			},
			"ip": {
				{0, 1.0},
			},
			"Ko": {
				{'h', 0.5},
				{'n', 1.0},
			},
			"ah": {
				{0, 1.0},
			},
		}
	default:
		return nil
	}
}
//
// End of generated data
//
