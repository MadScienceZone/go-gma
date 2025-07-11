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
// TianSing describes the naming conventions for the Tian-sing
// culture. Its methods give further details, but generally speaking
// the main operation to perform on these types is to just call the
// Generate and GenerateWithSurnames methods to create new names which
// conform to their cultural patterns.
//
type TianSing struct {
	BaseCulture
}

//
// prefix gives the prefix/selector string for each Tian-sing gender, or an empty
// string if one is not defined.
//
func (c TianSing) prefix(gender rune) string {
	switch gender {
	case 'F':
		return "__"
	case 'M':
		return "___"
	default:
		return ""
	}
}

//
// defaultMinMax returns the minimum and maximum size of Tian-sing names based on gender.
//
func (c TianSing) defaultMinMax(gender rune) (int, int) {
	switch gender {
	case 'F':
		return 5, 10
	case 'M':
		return 5, 10
	default:
		return 1, 1
	}
}

//
// Genders returns the set of genders defined for the Tian-sing culture.
//
func (c TianSing) Genders() []rune {
	return []rune{'F', 'M'}
}

//
// Name returns the name of the culture, i.e., "Tian-sing".
//
func (c TianSing) Name() string {
	return "Tian-sing"
}

//
// HasGender returns true if the specified gender code is defined
// in the Tian-sing culture.
//
func (c TianSing) HasGender(gender rune) bool {
	switch gender {
	case 'F', 'M':
		return true
	default:
		return false
	}
}

//
// db returns the name data for the given gender in the Tian-sing culture.
//
func (c TianSing) db(gender rune) map[string][]nameFragment {
	switch gender {
	case 'F':
		return map[string][]nameFragment{
			"_K": {
				{'u', 0.3333},
				{'e', 0.6667},
				{'a', 1.0},
			},
			"kl": {
				{'a', 1.0},
			},
			"ol": {
				{'a', 0.3333},
				{'k', 0.6667},
				{'m', 1.0},
			},
			"li": {
				{'n', 1.0},
			},
			"_R": {
				{'a', 1.0},
			},
			"ig": {
				{'m', 1.0},
			},
			"hu": {
				{'t', 1.0},
			},
			"or": {
				{'j', 1.0},
			},
			"ew": {
				{'a', 0.5},
				{'i', 1.0},
			},
			"rs": {
				{'i', 1.0},
			},
			"uk": {
				{'u', 1.0},
			},
			"_J": {
				{'i', 1.0},
			},
			"Dr": {
				{'o', 1.0},
			},
			"on": {
				{'g', 1.0},
			},
			"_T": {
				{'u', 0.3333},
				{'o', 0.6667},
				{'a', 1.0},
			},
			"Sh": {
				{'a', 1.0},
			},
			"Pa": {
				{'g', 0.5},
				{'n', 1.0},
			},
			"mo": {
				{0, 1.0},
			},
			"ni": {
				{0, 1.0},
			},
			"lh": {
				{'a', 1.0},
			},
			"nl": {
				{'h', 1.0},
			},
			"ia": {
				{0, 1.0},
			},
			"_N": {
				{'y', 0.2},
				{'j', 0.4},
				{'u', 0.6},
				{'a', 1.0},
			},
			"Nu": {
				{'r', 1.0},
			},
			"ur": {
				{'u', 0.5},
				{'i', 1.0},
			},
			"wi": {
				{0, 1.0},
			},
			"hs": {
				{'u', 1.0},
			},
			"Pu": {
				{'n', 1.0},
			},
			"_P": {
				{'u', 0.2},
				{'e', 0.6},
				{'a', 1.0},
			},
			"el": {
				{'a', 1.0},
			},
			"rj": {
				{'e', 1.0},
			},
			"ra": {
				{0, 1.0},
			},
			"na": {
				{0, 1.0},
			},
			"de": {
				{'n', 1.0},
			},
			"_U": {
				{'l', 1.0},
			},
			"rn": {
				{'a', 1.0},
			},
			"rt": {
				{'i', 1.0},
			},
			"Na": {
				{'m', 0.5},
				{'n', 1.0},
			},
			"ak": {
				{'o', 0.3333},
				{0, 0.6667},
				{'l', 1.0},
			},
			"ja": {
				{0, 1.0},
			},
			"Ku": {
				{'k', 1.0},
			},
			"wa": {
				{'r', 1.0},
			},
			"ma": {
				{0, 0.8333},
				{'r', 1.0},
			},
			"al": {
				{'i', 0.5},
				{'p', 1.0},
			},
			"en": {
				{0, 1.0},
			},
			"oe": {
				{'w', 1.0},
			},
			"De": {
				{'w', 0.5},
				{'k', 1.0},
			},
			"lf": {
				{'a', 1.0},
			},
			"Ny": {
				{'i', 1.0},
			},
			"__": {
				{'E', 0.0244},
				{'P', 0.1463},
				{'T', 0.2195},
				{'K', 0.2927},
				{'J', 0.3171},
				{'D', 0.5366},
				{'N', 0.6585},
				{'M', 0.7561},
				{'G', 0.7805},
				{'S', 0.878},
				{'H', 0.9024},
				{'U', 0.9268},
				{'C', 0.9512},
				{'R', 1.0},
			},
			"ri": {
				{0, 1.0},
			},
			"fa": {
				{'h', 1.0},
			},
			"Gy": {
				{'a', 1.0},
			},
			"ha": {
				{0, 0.5},
				{'l', 1.0},
			},
			"To": {
				{'n', 1.0},
			},
			"mh": {
				{'l', 1.0},
			},
			"_M": {
				{'a', 0.25},
				{'e', 0.5},
				{'i', 1.0},
			},
			"in": {
				{'g', 0.4},
				{'i', 0.6},
				{0, 1.0},
			},
			"Tu": {
				{'t', 1.0},
			},
			"Ka": {
				{'n', 1.0},
			},
			"Pe": {
				{'m', 0.5},
				{'t', 1.0},
			},
			"ar": {
				{'t', 0.2},
				{'a', 0.4},
				{'s', 0.6},
				{0, 1.0},
			},
			"as": {
				{'u', 1.0},
			},
			"Di": {
				{'n', 1.0},
			},
			"lk": {
				{'a', 1.0},
			},
			"us": {
				{'i', 1.0},
			},
			"pi": {
				{'a', 1.0},
			},
			"Ji": {
				{'n', 1.0},
			},
			"ga": {
				{0, 1.0},
			},
			"ng": {
				{'g', 0.25},
				{'m', 0.5},
				{'o', 0.75},
				{0, 1.0},
			},
			"Da": {
				{'r', 0.5},
				{'k', 1.0},
			},
			"_C": {
				{'h', 1.0},
			},
			"ya": {
				{'l', 1.0},
			},
			"Ta": {
				{'k', 1.0},
			},
			"yi": {
				{'m', 1.0},
			},
			"em": {
				{'a', 1.0},
			},
			"po": {
				{0, 1.0},
			},
			"hi": {
				{'a', 1.0},
			},
			"So": {
				{'e', 0.5},
				{'o', 1.0},
			},
			"Ha": {
				{'r', 1.0},
			},
			"gg": {
				{'a', 1.0},
			},
			"gm": {
				{'a', 0.6667},
				{'o', 1.0},
			},
			"ku": {
				{'l', 1.0},
			},
			"ut": {
				{'t', 0.25},
				{'h', 0.5},
				{0, 1.0},
			},
			"tt": {
				{'u', 1.0},
			},
			"ah": {
				{0, 0.5},
				{'s', 1.0},
			},
			"si": {
				{0, 0.5},
				{'h', 1.0},
			},
			"nk": {
				{'a', 1.0},
			},
			"Ch": {
				{'u', 1.0},
			},
			"oo": {
				{0, 0.5},
				{'p', 1.0},
			},
			"hl": {
				{'a', 1.0},
			},
			"ru": {
				{'l', 1.0},
			},
			"et": {
				{'a', 1.0},
			},
			"ag": {
				{'m', 1.0},
			},
			"ro": {
				{'l', 1.0},
			},
			"ih": {
				{0, 1.0},
			},
			"Ra": {
				{'s', 0.5},
				{'d', 1.0},
			},
			"lp": {
				{'o', 1.0},
			},
			"ta": {
				{'k', 1.0},
			},
			"Nj": {
				{'a', 1.0},
			},
			"Ul": {
				{'f', 1.0},
			},
			"lm": {
				{'a', 1.0},
			},
			"ka": {
				{'r', 0.5},
				{'n', 1.0},
			},
			"um": {
				{'a', 1.0},
			},
			"Mi": {
				{'g', 0.5},
				{'n', 1.0},
			},
			"th": {
				{'i', 0.5},
				{'o', 1.0},
			},
			"og": {
				{0, 1.0},
			},
			"im": {
				{'a', 1.0},
			},
			"je": {
				{0, 1.0},
			},
			"Ke": {
				{'l', 1.0},
			},
			"Ma": {
				{'h', 1.0},
			},
			"_D": {
				{'o', 0.2222},
				{'a', 0.4444},
				{'u', 0.5556},
				{'i', 0.6667},
				{'r', 0.7778},
				{'e', 1.0},
			},
			"ki": {
				{0, 1.0},
			},
			"an": {
				{'t', 0.5},
				{'g', 0.75},
				{'l', 1.0},
			},
			"ul": {
				{'a', 0.5},
				{0, 1.0},
			},
			"_S": {
				{'u', 0.25},
				{'o', 0.75},
				{'h', 1.0},
			},
			"tu": {
				{'t', 1.0},
			},
			"_E": {
				{'r', 1.0},
			},
			"_G": {
				{'y', 1.0},
			},
			"_H": {
				{'a', 1.0},
			},
			"un": {
				{'k', 0.5},
				{'a', 1.0},
			},
			"Du": {
				{'m', 1.0},
			},
			"op": {
				{'i', 1.0},
			},
			"go": {
				{'o', 1.0},
			},
			"ko": {
				{'l', 1.0},
			},
			"Er": {
				{'n', 1.0},
			},
			"Su": {
				{'s', 1.0},
			},
			"ad": {
				{'e', 1.0},
			},
			"nt": {
				{'h', 0.5},
				{'i', 1.0},
			},
			"ek": {
				{'i', 1.0},
			},
			"ti": {
				{0, 0.5},
				{'n', 1.0},
			},
			"am": {
				{'h', 1.0},
			},
			"eu": {
				{'t', 1.0},
			},
			"Do": {
				{'l', 0.5},
				{'r', 1.0},
			},
			"Me": {
				{'u', 1.0},
			},
			"ho": {
				{'g', 1.0},
			},
			"la": {
				{0, 1.0},
			},
			"su": {
				{'r', 0.5},
				{'n', 1.0},
			},
		}
	case 'M':
		return map[string][]nameFragment{
			"ndu": {
				{'g', 0.5},
				{'k', 1.0},
			},
			"_Ut": {
				{'u', 1.0},
			},
			"aul": {
				{'i', 1.0},
			},
			"Sin": {
				{'d', 1.0},
			},
			"pak": {
				{'a', 1.0},
			},
			"tem": {
				{'a', 1.0},
			},
			"tar": {
				{'a', 1.0},
			},
			"_Am": {
				{'i', 0.25},
				{'e', 0.5},
				{'a', 1.0},
			},
			"ger": {
				{0, 0.5},
				{'a', 1.0},
			},
			"oro": {
				{0, 1.0},
			},
			"_Aj": {
				{'i', 1.0},
			},
			"amp": {
				{'a', 1.0},
			},
			"Unu": {
				{'s', 1.0},
			},
			"eif": {
				{'e', 1.0},
			},
			"hja": {
				{'y', 1.0},
			},
			"__M": {
				{'a', 0.9231},
				{'o', 1.0},
			},
			"han": {
				{'d', 0.25},
				{'a', 0.5},
				{0, 0.75},
				{'u', 1.0},
			},
			"Baa": {
				{'b', 1.0},
			},
			"esw": {
				{'a', 1.0},
			},
			"tun": {
				{'g', 1.0},
			},
			"anu": {
				{0, 0.5},
				{'s', 1.0},
			},
			"_Da": {
				{'k', 1.0},
			},
			"oes": {
				{'o', 1.0},
			},
			"uni": {
				{0, 1.0},
			},
			"tta": {
				{0, 1.0},
			},
			"_Su": {
				{'m', 0.1429},
				{'r', 0.2857},
				{'n', 0.4286},
				{'l', 0.5714},
				{'s', 0.7143},
				{'k', 1.0},
			},
			"_Sr": {
				{'i', 1.0},
			},
			"mpa": {
				{'k', 1.0},
			},
			"ntj": {
				{'i', 1.0},
			},
			"ema": {
				{'r', 1.0},
			},
			"Tji": {
				{'p', 1.0},
			},
			"_Te": {
				{'u', 1.0},
			},
			"liw": {
				{'a', 1.0},
			},
			"lua": {
				{'n', 1.0},
			},
			"nat": {
				{'h', 1.0},
			},
			"__D": {
				{'e', 0.4286},
				{'i', 0.7143},
				{'h', 0.8571},
				{'a', 1.0},
			},
			"_De": {
				{'w', 0.6667},
				{'d', 1.0},
			},
			"awi": {
				{'r', 1.0},
			},
			"_An": {
				{'u', 1.0},
			},
			"San": {
				{'g', 1.0},
			},
			"Utu": {
				{'n', 1.0},
			},
			"__H": {
				{'i', 0.2857},
				{'u', 0.5714},
				{'a', 1.0},
			},
			"Soe": {
				{'m', 1.0},
			},
			"Mal": {
				{'a', 1.0},
			},
			"ago": {
				{0, 1.0},
			},
			"euk": {
				{'u', 1.0},
			},
			"nte": {
				{'m', 1.0},
			},
			"Bal": {
				{'a', 0.5},
				{'i', 1.0},
			},
			"_Ch": {
				{'a', 1.0},
			},
			"Ard": {
				{'h', 1.0},
			},
			"ife": {
				{0, 1.0},
			},
			"des": {
				{0, 1.0},
			},
			"__K": {
				{'a', 0.375},
				{'u', 0.625},
				{'r', 0.75},
				{'e', 1.0},
			},
			"ali": {
				{'h', 1.0},
			},
			"__U": {
				{'m', 0.25},
				{'n', 0.5},
				{'l', 0.75},
				{'t', 1.0},
			},
			"yaf": {
				{'r', 1.0},
			},
			"Hat": {
				{'t', 1.0},
			},
			"asa": {
				{0, 1.0},
			},
			"___": {
				{'S', 0.1774},
				{'W', 0.2258},
				{'M', 0.3306},
				{'U', 0.3629},
				{'D', 0.4194},
				{'H', 0.4758},
				{'G', 0.5},
				{'T', 0.5726},
				{'N', 0.5887},
				{'E', 0.6048},
				{'C', 0.6371},
				{'B', 0.6855},
				{'J', 0.7177},
				{'R', 0.7661},
				{'K', 0.8306},
				{'P', 0.9032},
				{'A', 1.0},
			},
			"lan": {
				{'a', 0.3333},
				{'g', 1.0},
			},
			"oem": {
				{'a', 1.0},
			},
			"esi": {
				{'a', 1.0},
			},
			"neg": {
				{'r', 0.3333},
				{'a', 1.0},
			},
			"udi": {
				{0, 0.5},
				{'n', 1.0},
			},
			"ong": {
				{0, 1.0},
			},
			"_Ta": {
				{'n', 1.0},
			},
			"wat": {
				{'i', 1.0},
			},
			"_Pa": {
				{'r', 0.1667},
				{'t', 0.3333},
				{'k', 0.5},
				{'l', 0.6667},
				{'n', 1.0},
			},
			"Sla": {
				{'m', 1.0},
			},
			"_Ku": {
				{'n', 0.5},
				{'l', 1.0},
			},
			"tum": {
				{'a', 1.0},
			},
			"ato": {
				{'n', 1.0},
			},
			"ggu": {
				{'l', 1.0},
			},
			"Nan": {
				{'g', 1.0},
			},
			"una": {
				{'n', 0.5},
				{'j', 1.0},
			},
			"son": {
				{0, 1.0},
			},
			"Mat": {
				{'u', 1.0},
			},
			"yan": {
				{'a', 1.0},
			},
			"aki": {
				{'r', 1.0},
			},
			"tro": {
				{0, 1.0},
			},
			"war": {
				{'a', 0.25},
				{'i', 0.5},
				{'m', 1.0},
			},
			"__R": {
				{'a', 0.8333},
				{'e', 1.0},
			},
			"din": {
				{0, 1.0},
			},
			"aku": {
				{0, 1.0},
			},
			"Sri": {
				{0, 1.0},
			},
			"art": {
				{'a', 0.6667},
				{'o', 1.0},
			},
			"ara": {
				{'m', 0.1111},
				{'n', 0.2222},
				{'j', 0.4444},
				{0, 0.8889},
				{'s', 1.0},
			},
			"irt": {
				{'i', 1.0},
			},
			"nus": {
				{0, 0.3333},
				{'a', 1.0},
			},
			"_Ja": {
				{'t', 0.25},
				{'y', 0.75},
				{'r', 1.0},
			},
			"som": {
				{'o', 1.0},
			},
			"_Tj": {
				{'o', 0.5},
				{'i', 1.0},
			},
			"Jar": {
				{'o', 1.0},
			},
			"jah": {
				{0, 0.5},
				{'r', 1.0},
			},
			"ram": {
				{'e', 1.0},
			},
			"__B": {
				{'o', 0.1667},
				{'u', 0.5},
				{'a', 1.0},
			},
			"Aji": {
				{'s', 0.5},
				{0, 1.0},
			},
			"_Ga": {
				{'j', 0.5},
				{'m', 1.0},
			},
			"eng": {
				{0, 0.6667},
				{'k', 1.0},
			},
			"ono": {
				{0, 1.0},
			},
			"_Si": {
				{'n', 1.0},
			},
			"ubu": {
				{'n', 1.0},
			},
			"har": {
				{'a', 0.5},
				{'i', 0.75},
				{'m', 1.0},
			},
			"ohj": {
				{'a', 1.0},
			},
			"_Ai": {
				{'r', 1.0},
			},
			"__T": {
				{'e', 0.2222},
				{'o', 0.3333},
				{'r', 0.5556},
				{'u', 0.6667},
				{'j', 0.8889},
				{'a', 1.0},
			},
			"luk": {
				{'a', 1.0},
			},
			"Raj": {
				{'a', 1.0},
			},
			"Tri": {
				{'a', 1.0},
			},
			"gku": {
				{'b', 0.2},
				{'r', 0.4},
				{0, 1.0},
			},
			"ede": {
				{'s', 1.0},
			},
			"gro": {
				{'e', 1.0},
			},
			"Rad": {
				{'e', 1.0},
			},
			"_Ma": {
				{'s', 0.0833},
				{'t', 0.1667},
				{'u', 0.25},
				{'l', 0.3333},
				{'n', 0.5833},
				{'r', 0.75},
				{'h', 0.9167},
				{'d', 1.0},
			},
			"rma": {
				{'d', 0.5},
				{'n', 0.75},
				{'w', 1.0},
			},
			"uan": {
				{'g', 1.0},
			},
			"aab": {
				{0, 1.0},
			},
			"yaw": {
				{'a', 1.0},
			},
			"__P": {
				{'a', 0.6667},
				{'u', 0.8889},
				{'r', 1.0},
			},
			"_Sa": {
				{'n', 0.1667},
				{'t', 0.3333},
				{'m', 0.5},
				{'i', 0.6667},
				{'r', 1.0},
			},
			"_Ka": {
				{'r', 0.6667},
				{'l', 1.0},
			},
			"kan": {
				{'l', 1.0},
			},
			"irl": {
				{'a', 1.0},
			},
			"pan": {
				{'d', 0.5},
				{'e', 1.0},
			},
			"ota": {
				{'m', 1.0},
			},
			"iso": {
				{'n', 1.0},
			},
			"Ama": {
				{'n', 1.0},
			},
			"_Re": {
				{'x', 1.0},
			},
			"Ham": {
				{'e', 1.0},
			},
			"__C": {
				{'a', 0.25},
				{'h', 1.0},
			},
			"rla": {
				{'n', 1.0},
			},
			"rab": {
				{'h', 1.0},
			},
			"rit": {
				{0, 1.0},
			},
			"Ded": {
				{'e', 1.0},
			},
			"bun": {
				{'i', 1.0},
			},
			"man": {
				{0, 0.5},
				{'g', 1.0},
			},
			"Sen": {
				{'o', 0.5},
				{'a', 1.0},
			},
			"_Hi": {
				{'t', 0.5},
				{'a', 1.0},
			},
			"jas": {
				{'a', 1.0},
			},
			"rae": {
				{'n', 1.0},
			},
			"Kul": {
				{'p', 1.0},
			},
			"ont": {
				{'e', 1.0},
			},
			"ert": {
				{'a', 1.0},
			},
			"atu": {
				{'l', 0.5},
				{'m', 1.0},
			},
			"Pan": {
				{'e', 0.5},
				{'g', 1.0},
			},
			"lis": {
				{'o', 1.0},
			},
			"__A": {
				{'m', 0.3333},
				{'i', 0.4167},
				{'r', 0.6667},
				{'n', 0.75},
				{'c', 0.8333},
				{'j', 1.0},
			},
			"Buw": {
				{'a', 1.0},
			},
			"Sja": {
				{'h', 1.0},
			},
			"rto": {
				{'r', 1.0},
			},
			"sak": {
				{'a', 1.0},
			},
			"nap": {
				{'a', 0.5},
				{'o', 1.0},
			},
			"mhe": {
				{'n', 1.0},
			},
			"mar": {
				{0, 0.25},
				{'a', 0.5},
				{'s', 0.75},
				{'l', 1.0},
			},
			"hal": {
				{'a', 1.0},
			},
			"rta": {
				{'j', 0.3333},
				{'n', 1.0},
			},
			"ihu": {
				{'n', 1.0},
			},
			"_En": {
				{'t', 1.0},
			},
			"gar": {
				{'a', 1.0},
			},
			"uwa": {
				{'n', 1.0},
			},
			"put": {
				{'i', 0.5},
				{'r', 1.0},
			},
			"uha": {
				{'n', 1.0},
			},
			"_Tr": {
				{'i', 0.5},
				{'u', 1.0},
			},
			"ala": {
				{'d', 0.1667},
				{'n', 0.3333},
				{'p', 0.5},
				{'k', 0.8333},
				{0, 1.0},
			},
			"und": {
				{'u', 1.0},
			},
			"sam": {
				{'a', 1.0},
			},
			"mes": {
				{'w', 1.0},
			},
			"ama": {
				{'r', 0.2},
				{0, 0.8},
				{'n', 1.0},
			},
			"rot": {
				{0, 0.5},
				{'a', 1.0},
			},
			"_Dh": {
				{'a', 1.0},
			},
			"les": {
				{'i', 1.0},
			},
			"arl": {
				{'u', 1.0},
			},
			"nko": {
				{'e', 1.0},
			},
			"__S": {
				{'a', 0.2727},
				{'r', 0.3182},
				{'y', 0.3636},
				{'j', 0.4091},
				{'u', 0.7273},
				{'l', 0.7727},
				{'i', 0.8182},
				{'o', 0.9091},
				{'e', 1.0},
			},
			"uge": {
				{'r', 1.0},
			},
			"_Ar": {
				{'d', 0.3333},
				{'o', 0.6667},
				{'u', 1.0},
			},
			"apa": {
				{'n', 0.25},
				{'l', 0.5},
				{'t', 1.0},
			},
			"awa": {
				{'t', 0.1667},
				{'n', 0.5},
				{'r', 0.8333},
				{'m', 1.0},
			},
			"yak": {
				{'a', 1.0},
			},
			"_Ul": {
				{'l', 1.0},
			},
			"_Wu": {
				{'r', 0.6667},
				{'n', 1.0},
			},
			"kat": {
				{'o', 1.0},
			},
			"Tan": {
				{0, 1.0},
			},
			"ega": {
				{'r', 1.0},
			},
			"ugg": {
				{'a', 1.0},
			},
			"Ami": {
				{'e', 1.0},
			},
			"Air": {
				{'l', 1.0},
			},
			"ndy": {
				{'a', 1.0},
			},
			"_Pu": {
				{'g', 0.5},
				{'r', 1.0},
			},
			"_Sj": {
				{'a', 1.0},
			},
			"Teu": {
				{'n', 0.5},
				{'k', 1.0},
			},
			"ame": {
				{'s', 0.3333},
				{'t', 0.6667},
				{'n', 1.0},
			},
			"caw": {
				{'a', 1.0},
			},
			"lah": {
				{0, 1.0},
			},
			"wam": {
				{'a', 1.0},
			},
			"uku": {
				{0, 1.0},
			},
			"Sus": {
				{'u', 1.0},
			},
			"tra": {
				{0, 1.0},
			},
			"oam": {
				{'i', 1.0},
			},
			"yam": {
				{0, 1.0},
			},
			"_Ha": {
				{'m', 0.3333},
				{'y', 0.6667},
				{'t', 1.0},
			},
			"sia": {
				{0, 1.0},
			},
			"__J": {
				{'a', 1.0},
			},
			"Nar": {
				{'o', 1.0},
			},
			"ind": {
				{'u', 1.0},
			},
			"tan": {
				{'e', 1.0},
			},
			"kar": {
				{'n', 1.0},
			},
			"nan": {
				{0, 1.0},
			},
			"Sai": {
				{'d', 1.0},
			},
			"lei": {
				{'f', 1.0},
			},
			"tha": {
				{0, 1.0},
			},
			"itr": {
				{'o', 1.0},
			},
			"roa": {
				{'m', 1.0},
			},
			"jis": {
				{'a', 1.0},
			},
			"Kra": {
				{'e', 1.0},
			},
			"pal": {
				{'a', 1.0},
			},
			"Aca": {
				{'w', 1.0},
			},
			"ena": {
				{'p', 1.0},
			},
			"ouk": {
				{'i', 1.0},
			},
			"_Wi": {
				{'j', 0.5},
				{'r', 1.0},
			},
			"hak": {
				{'a', 1.0},
			},
			"aro": {
				{'t', 1.0},
			},
			"roe": {
				{0, 1.0},
			},
			"ngo": {
				{'e', 1.0},
			},
			"ana": {
				{'n', 0.2},
				{'p', 0.4},
				{'c', 0.6},
				{'t', 0.8},
				{0, 1.0},
			},
			"iag": {
				{'o', 1.0},
			},
			"nta": {
				{'r', 1.0},
			},
			"Ent": {
				{'j', 1.0},
			},
			"ira": {
				{0, 0.5},
				{'p', 1.0},
			},
			"ipt": {
				{'o', 1.0},
			},
			"tam": {
				{'a', 1.0},
			},
			"Mau": {
				{'l', 1.0},
			},
			"ipa": {
				{'n', 1.0},
			},
			"ade": {
				{'w', 0.8},
				{'n', 1.0},
			},
			"Kal": {
				{'a', 1.0},
			},
			"taj": {
				{'a', 1.0},
			},
			"duk": {
				{0, 1.0},
			},
			"aha": {
				{'r', 0.3333},
				{'n', 0.6667},
				{'l', 1.0},
			},
			"raj": {
				{'a', 1.0},
			},
			"Bud": {
				{'i', 1.0},
			},
			"Tjo": {
				{'k', 1.0},
			},
			"Sur": {
				{'a', 1.0},
			},
			"itu": {
				{0, 1.0},
			},
			"Ull": {
				{'a', 1.0},
			},
			"ula": {
				{'n', 1.0},
			},
			"_Ra": {
				{'n', 0.2},
				{'j', 0.4},
				{'k', 0.6},
				{'m', 0.8},
				{'d', 1.0},
			},
			"Wun": {
				{'i', 1.0},
			},
			"ggr": {
				{'o', 1.0},
			},
			"Mar": {
				{'m', 0.5},
				{'a', 1.0},
			},
			"ngk": {
				{'u', 1.0},
			},
			"rna": {
				{'d', 1.0},
			},
			"ata": {
				{'w', 1.0},
			},
			"enk": {
				{'o', 1.0},
			},
			"Sun": {
				{'a', 1.0},
			},
			"Man": {
				{'g', 0.6667},
				{'u', 1.0},
			},
			"Sya": {
				{'f', 1.0},
			},
			"lpu": {
				{'t', 1.0},
			},
			"Hul": {
				{'a', 1.0},
			},
			"lad": {
				{'e', 1.0},
			},
			"Wir": {
				{'a', 1.0},
			},
			"uki": {
				{'l', 1.0},
			},
			"Sar": {
				{'t', 0.5},
				{'w', 1.0},
			},
			"kur": {
				{'a', 1.0},
			},
			"ung": {
				{0, 0.4},
				{'k', 0.6},
				{'g', 1.0},
			},
			"kil": {
				{0, 1.0},
			},
			"amh": {
				{'e', 1.0},
			},
			"rap": {
				{'a', 1.0},
			},
			"ati": {
				{0, 1.0},
			},
			"__W": {
				{'a', 0.1667},
				{'u', 0.6667},
				{'i', 1.0},
			},
			"Pat": {
				{'i', 1.0},
			},
			"opa": {
				{'t', 1.0},
			},
			"nge": {
				{'r', 1.0},
			},
			"Dip": {
				{'a', 1.0},
			},
			"_Pr": {
				{'a', 1.0},
			},
			"umi": {
				{'t', 1.0},
			},
			"arm": {
				{'a', 1.0},
			},
			"nem": {
				{'b', 1.0},
			},
			"ija": {
				{'y', 1.0},
			},
			"rwo": {
				{'n', 1.0},
			},
			"iaw": {
				{'a', 1.0},
			},
			"gra": {
				{0, 1.0},
			},
			"dya": {
				{0, 1.0},
			},
			"Sat": {
				{'y', 1.0},
			},
			"ksa": {
				{0, 1.0},
			},
			"Aru": {
				{0, 1.0},
			},
			"bha": {
				{'n', 1.0},
			},
			"uma": {
				{'h', 1.0},
			},
			"eso": {
				{'m', 1.0},
			},
			"ien": {
				{0, 1.0},
			},
			"maw": {
				{'a', 1.0},
			},
			"eno": {
				{'p', 1.0},
			},
			"mba": {
				{'h', 1.0},
			},
			"Rak": {
				{'e', 1.0},
			},
			"_Um": {
				{'a', 1.0},
			},
			"Mah": {
				{'a', 0.5},
				{'i', 1.0},
			},
			"abh": {
				{'a', 1.0},
			},
			"pto": {
				{0, 1.0},
			},
			"ura": {
				{'t', 0.3333},
				{'w', 0.6667},
				{'p', 1.0},
			},
			"rno": {
				{0, 1.0},
			},
			"Ker": {
				{'t', 1.0},
			},
			"ulu": {
				{'a', 1.0},
			},
			"ami": {
				{'n', 1.0},
			},
			"usu": {
				{'h', 1.0},
			},
			"tul": {
				{'e', 1.0},
			},
			"uts": {
				{'i', 1.0},
			},
			"_Ac": {
				{'a', 1.0},
			},
			"hun": {
				{'g', 1.0},
			},
			"ari": {
				{'y', 0.5},
				{0, 1.0},
			},
			"Hut": {
				{'s', 1.0},
			},
			"ant": {
				{'a', 1.0},
			},
			"Cam": {
				{'p', 1.0},
			},
			"Gli": {
				{'s', 1.0},
			},
			"ars": {
				{'a', 1.0},
			},
			"_Bo": {
				{'l', 1.0},
			},
			"mah": {
				{'a', 1.0},
			},
			"aja": {
				{0, 0.3333},
				{'y', 0.6667},
				{'s', 0.8333},
				{'h', 1.0},
			},
			"Ame": {
				{'t', 1.0},
			},
			"ahr": {
				{'i', 1.0},
			},
			"aka": {
				{0, 0.7143},
				{'t', 0.8571},
				{'k', 1.0},
			},
			"ran": {
				{0, 0.5},
				{'o', 1.0},
			},
			"Dha": {
				{'r', 1.0},
			},
			"anl": {
				{'a', 1.0},
			},
			"_Ke": {
				{'r', 0.5},
				{'n', 1.0},
			},
			"lla": {
				{'h', 1.0},
			},
			"nop": {
				{'a', 1.0},
			},
			"Gam": {
				{'h', 1.0},
			},
			"_To": {
				{'h', 1.0},
			},
			"jok": {
				{'r', 1.0},
			},
			"lih": {
				{'u', 1.0},
			},
			"gul": {
				{0, 1.0},
			},
			"dew": {
				{'a', 1.0},
			},
			"suh": {
				{'a', 1.0},
			},
			"ird": {
				{'j', 1.0},
			},
			"egr": {
				{'a', 1.0},
			},
			"Anu": {
				{'s', 1.0},
			},
			"ake": {
				{0, 1.0},
			},
			"uru": {
				{'k', 1.0},
			},
			"dha": {
				{'r', 1.0},
			},
			"rsa": {
				{'i', 1.0},
			},
			"okr": {
				{'o', 1.0},
			},
			"_Ca": {
				{'m', 1.0},
			},
			"_Mo": {
				{'n', 1.0},
			},
			"pat": {
				{'i', 1.0},
			},
			"ndr": {
				{'a', 1.0},
			},
			"Wur": {
				{'u', 0.5},
				{'a', 1.0},
			},
			"Dew": {
				{'a', 1.0},
			},
			"rti": {
				{0, 1.0},
			},
			"Pal": {
				{'a', 1.0},
			},
			"Mad": {
				{'a', 1.0},
			},
			"and": {
				{'r', 0.5},
				{'y', 1.0},
			},
			"san": {
				{'a', 1.0},
			},
			"Ken": {
				{0, 1.0},
			},
			"nad": {
				{'i', 1.0},
			},
			"ane": {
				{'g', 0.75},
				{'m', 1.0},
			},
			"rlu": {
				{'k', 1.0},
			},
			"lam": {
				{'e', 1.0},
			},
			"mad": {
				{'e', 1.0},
			},
			"aty": {
				{'a', 1.0},
			},
			"Wij": {
				{'a', 1.0},
			},
			"Sul": {
				{'u', 1.0},
			},
			"exi": {
				{0, 1.0},
			},
			"_Se": {
				{'n', 1.0},
			},
			"Ram": {
				{'a', 1.0},
			},
			"Bol": {
				{'e', 1.0},
			},
			"aca": {
				{0, 1.0},
			},
			"Cha": {
				{'k', 0.3333},
				{'n', 0.6667},
				{'r', 1.0},
			},
			"Jat": {
				{'a', 1.0},
			},
			"mit": {
				{'r', 1.0},
			},
			"bah": {
				{'a', 1.0},
			},
			"ras": {
				{'a', 1.0},
			},
			"_Tu": {
				{'n', 1.0},
			},
			"gsa": {
				{0, 1.0},
			},
			"goe": {
				{'n', 1.0},
			},
			"jip": {
				{'t', 1.0},
			},
			"ole": {
				{'i', 1.0},
			},
			"Gaj": {
				{'a', 1.0},
			},
			"his": {
				{'a', 1.0},
			},
			"aid": {
				{0, 1.0},
			},
			"won": {
				{'o', 1.0},
			},
			"rdh": {
				{'a', 1.0},
			},
			"_Hu": {
				{'l', 0.5},
				{'t', 1.0},
			},
			"_Sl": {
				{'a', 1.0},
			},
			"riy": {
				{'a', 1.0},
			},
			"rat": {
				{0, 1.0},
			},
			"lak": {
				{'a', 1.0},
			},
			"Sum": {
				{'i', 1.0},
			},
			"tsi": {
				{'n', 1.0},
			},
			"mie": {
				{'n', 1.0},
			},
			"swa": {
				{'r', 1.0},
			},
			"min": {
				{'o', 1.0},
			},
			"_Wa": {
				{'t', 1.0},
			},
			"sin": {
				{0, 1.0},
			},
			"raw": {
				{'a', 0.5},
				{'i', 1.0},
			},
			"__E": {
				{'n', 0.5},
				{'k', 1.0},
			},
			"Hia": {
				{'w', 1.0},
			},
			"iya": {
				{'k', 1.0},
			},
			"ino": {
				{'t', 1.0},
			},
			"jay": {
				{'a', 1.0},
			},
			"ule": {
				{'s', 1.0},
			},
			"nla": {
				{'h', 1.0},
			},
			"adi": {
				{0, 1.0},
			},
			"wap": {
				{'a', 1.0},
			},
			"_Di": {
				{'p', 0.5},
				{'r', 1.0},
			},
			"ano": {
				{0, 1.0},
			},
			"emb": {
				{'a', 1.0},
			},
			"era": {
				{'n', 1.0},
			},
			"_Un": {
				{'u', 1.0},
			},
			"Tun": {
				{'g', 1.0},
			},
			"Mon": {
				{'t', 1.0},
			},
			"ngs": {
				{'a', 1.0},
			},
			"uti": {
				{'h', 1.0},
			},
			"Rex": {
				{'i', 1.0},
			},
			"Ran": {
				{'g', 1.0},
			},
			"__N": {
				{'a', 1.0},
			},
			"men": {
				{'g', 1.0},
			},
			"Mas": {
				{0, 1.0},
			},
			"urn": {
				{'a', 1.0},
			},
			"arw": {
				{'o', 1.0},
			},
			"utr": {
				{'a', 1.0},
			},
			"__G": {
				{'a', 0.6667},
				{'l', 1.0},
			},
			"Suk": {
				{'a', 1.0},
			},
			"uka": {
				{0, 0.3333},
				{'n', 0.6667},
				{'r', 1.0},
			},
			"wan": {
				{'g', 0.5},
				{'a', 0.75},
				{'t', 1.0},
			},
			"Aro": {
				{'k', 1.0},
			},
			"ulp": {
				{'u', 1.0},
			},
			"ang": {
				{'g', 0.1538},
				{'o', 0.2308},
				{'e', 0.3077},
				{'s', 0.3846},
				{'k', 0.6154},
				{0, 0.9231},
				{'a', 1.0},
			},
			"koe": {
				{'s', 1.0},
			},
			"mou": {
				{'k', 1.0},
			},
			"aks": {
				{'a', 1.0},
			},
			"ewa": {
				{'n', 0.1667},
				{'p', 0.3333},
				{0, 1.0},
			},
			"met": {
				{'u', 0.5},
				{0, 1.0},
			},
			"nga": {
				{0, 1.0},
			},
			"fru": {
				{'d', 1.0},
			},
			"_Na": {
				{'n', 0.5},
				{'r', 1.0},
			},
			"Eko": {
				{0, 1.0},
			},
			"wir": {
				{'a', 1.0},
			},
			"Pur": {
				{'n', 1.0},
			},
			"dja": {
				{0, 1.0},
			},
			"_Ek": {
				{'o', 1.0},
			},
			"Kar": {
				{'t', 1.0},
			},
			"Jay": {
				{'a', 1.0},
			},
			"uli": {
				{'w', 1.0},
			},
			"gad": {
				{'e', 1.0},
			},
			"gga": {
				{0, 0.6667},
				{'d', 1.0},
			},
			"_Ba": {
				{'l', 0.6667},
				{'a', 1.0},
			},
			"ada": {
				{0, 1.0},
			},
			"ria": {
				{'g', 1.0},
			},
			"hri": {
				{'t', 1.0},
			},
			"apo": {
				{0, 1.0},
			},
			"run": {
				{'a', 1.0},
			},
			"Tru": {
				{'n', 1.0},
			},
			"hen": {
				{'g', 1.0},
			},
			"Dir": {
				{'d', 1.0},
			},
			"not": {
				{'o', 1.0},
			},
			"arn": {
				{'o', 1.0},
			},
			"kir": {
				{'t', 1.0},
			},
			"Wat": {
				{'u', 1.0},
			},
			"afr": {
				{'u', 1.0},
			},
			"sai": {
				{'d', 1.0},
			},
			"apu": {
				{'t', 1.0},
			},
			"oen": {
				{'k', 1.0},
			},
			"taw": {
				{'a', 1.0},
			},
			"nac": {
				{'a', 1.0},
			},
			"tih": {
				{0, 1.0},
			},
			"Hay": {
				{'a', 1.0},
			},
			"lap": {
				{'u', 1.0},
			},
			"eun": {
				{'g', 1.0},
			},
			"rud": {
				{'i', 1.0},
			},
			"tji": {
				{0, 1.0},
			},
			"Hit": {
				{'u', 1.0},
			},
			"aen": {
				{'g', 1.0},
			},
			"ton": {
				{'g', 1.0},
			},
			"tor": {
				{'o', 1.0},
			},
			"dug": {
				{'g', 1.0},
			},
			"den": {
				{0, 1.0},
			},
			"Pra": {
				{'w', 1.0},
			},
			"_Kr": {
				{'a', 1.0},
			},
			"rdj": {
				{'a', 1.0},
			},
			"Uma": {
				{'r', 1.0},
			},
			"Kun": {
				{'d', 1.0},
			},
			"_Bu": {
				{'d', 0.5},
				{'w', 1.0},
			},
			"omo": {
				{0, 0.5},
				{'u', 1.0},
			},
			"_So": {
				{'e', 0.5},
				{'m', 1.0},
			},
			"att": {
				{'a', 1.0},
			},
			"ahi": {
				{'s', 1.0},
			},
			"oto": {
				{0, 1.0},
			},
			"Som": {
				{'o', 1.0},
			},
			"Sam": {
				{'a', 1.0},
			},
			"_Gl": {
				{'i', 1.0},
			},
			"kub": {
				{'u', 1.0},
			},
			"naj": {
				{'a', 1.0},
			},
			"Pak": {
				{'u', 1.0},
			},
			"Dak": {
				{'s', 1.0},
			},
			"ngg": {
				{'r', 0.25},
				{'u', 0.5},
				{'a', 1.0},
			},
			"Pug": {
				{'e', 1.0},
			},
			"kak": {
				{'i', 1.0},
			},
			"isa": {
				{'k', 0.5},
				{0, 1.0},
			},
			"ath": {
				{'a', 1.0},
			},
			"etu": {
				{'n', 1.0},
			},
			"tya": {
				{'w', 1.0},
			},
			"kro": {
				{'a', 1.0},
			},
			"aya": {
				{'m', 0.1429},
				{0, 0.7143},
				{'k', 0.8571},
				{'n', 1.0},
			},
			"Par": {
				{'a', 1.0},
			},
			"rok": {
				{0, 1.0},
			},
			"usa": {
				{'n', 0.5},
				{'m', 1.0},
			},
			"dra": {
				{'b', 1.0},
			},
			"Toh": {
				{'j', 1.0},
			},
			"ruk": {
				{0, 1.0},
			},
			"_Sy": {
				{'a', 1.0},
			},
			"iwa": {
				{'r', 1.0},
			},
		}
	default:
		return nil
	}
}
//
// End of generated data
//
