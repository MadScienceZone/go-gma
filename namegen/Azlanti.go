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
// Azlanti describes the naming conventions for the Azlanti
// culture. Its methods give further details, but generally speaking
// the main operation to perform on these types is to just call the
// Generate and GenerateWithSurnames methods to create new names which
// conform to their cultural patterns.
//
type Azlanti struct {
	BaseCulture
}

//
// defaultMinMax returns the minimum and maximum size of Azlanti names based on gender.
//
func (c Azlanti) defaultMinMax(gender rune) (int, int) {
	switch gender {
	case 'F':
		return 5, 14
	case 'M':
		return 5, 11
	default:
		return 1, 1
	}
}

//
// Genders returns the set of genders defined for the Azlanti culture.
//
func (c Azlanti) Genders() []rune {
	return []rune{'F', 'M'}
}

//
// Name returns the name of the culture, i.e., "Azlanti".
//
func (c Azlanti) Name() string {
	return "Azlanti"
}

//
// HasGender returns true if the specified gender code is defined
// in the Azlanti culture.
//
func (c Azlanti) HasGender(gender rune) bool {
	switch gender {
	case 'F', 'M':
		return true
	default:
		return false
	}
}

//
// db returns the name data for the given gender in the Azlanti culture.
//
func (c Azlanti) db(gender rune) map[string][]nameFragment {
	switch gender {
	case 'F':
		return map[string][]nameFragment{
			"lis": {
				{'a', 0.5},
				{'h', 1.0},
			},
			"Eun": {
				{'i', 1.0},
			},
			"ion": {
				{'e', 1.0},
			},
			"all": {
				{'i', 1.0},
			},
			"cia": {
				{'s', 1.0},
			},
			"hir": {
				{0, 1.0},
			},
			"uil": {
				{'a', 1.0},
			},
			"wai": {
				{'t', 1.0},
			},
			"ith": {
				{'e', 1.0},
			},
			"_Az": {
				{'u', 1.0},
			},
			"_Io": {
				{'m', 1.0},
			},
			"___": {
				{'A', 0.55},
				{'E', 0.7667},
				{'U', 0.8},
				{'O', 0.9167},
				{'I', 1.0},
			},
			"_Ov": {
				{'i', 1.0},
			},
			"eth": {
				{0, 1.0},
			},
			"_Op": {
				{'h', 1.0},
			},
			"nna": {
				{0, 1.0},
			},
			"the": {
				{'r', 0.5},
				{0, 1.0},
			},
			"Ovi": {
				{'e', 1.0},
			},
			"ian": {
				{'d', 1.0},
			},
			"__A": {
				{'n', 0.1212},
				{'z', 0.1515},
				{'c', 0.1818},
				{'p', 0.2121},
				{'d', 0.2424},
				{'q', 0.2727},
				{'m', 0.3333},
				{'b', 0.5152},
				{'s', 0.6667},
				{'l', 0.697},
				{'v', 0.8182},
				{'t', 0.9091},
				{'r', 0.9394},
				{'h', 0.9697},
				{'g', 1.0},
			},
			"rri": {
				{'n', 1.0},
			},
			"rin": {
				{0, 1.0},
			},
			"lli": {
				{'a', 1.0},
			},
			"tha": {
				{'l', 1.0},
			},
			"swa": {
				{'i', 1.0},
			},
			"ija": {
				{'h', 1.0},
			},
			"rud": {
				{'e', 1.0},
			},
			"_Ev": {
				{'e', 1.0},
			},
			"_Or": {
				{'p', 1.0},
			},
			"mma": {
				{'r', 1.0},
			},
			"aom": {
				{'e', 1.0},
			},
			"Isc": {
				{'a', 1.0},
			},
			"Ana": {
				{'t', 0.6667},
				{'h', 1.0},
			},
			"rpa": {
				{'h', 1.0},
			},
			"ara": {
				{0, 0.5},
				{'h', 1.0},
			},
			"nat": {
				{'h', 0.5},
				{0, 1.0},
			},
			"__I": {
				{'o', 0.2},
				{'s', 0.4},
				{'v', 0.6},
				{'a', 0.8},
				{'z', 1.0},
			},
			"dna": {
				{'h', 0.5},
				{0, 1.0},
			},
			"_Es": {
				{'t', 1.0},
			},
			"nah": {
				{0, 1.0},
			},
			"ias": {
				{0, 1.0},
			},
			"bis": {
				{'h', 1.0},
			},
			"sen": {
				{'a', 1.0},
			},
			"and": {
				{'r', 0.5},
				{'a', 1.0},
			},
			"zev": {
				{'e', 1.0},
			},
			"_Is": {
				{'c', 1.0},
			},
			"iel": {
				{0, 1.0},
			},
			"_Al": {
				{'i', 1.0},
			},
			"Eve": {
				{0, 1.0},
			},
			"pah": {
				{0, 1.0},
			},
			"Ach": {
				{'i', 1.0},
			},
			"qui": {
				{'l', 1.0},
			},
			"Oph": {
				{'i', 1.0},
			},
			"bil": {
				{'e', 1.0},
			},
			"Ann": {
				{'a', 1.0},
			},
			"_Em": {
				{'a', 1.0},
			},
			"jah": {
				{0, 1.0},
			},
			"mal": {
				{'l', 1.0},
			},
			"tru": {
				{'d', 1.0},
			},
			"iya": {
				{'h', 1.0},
			},
			"gri": {
				{'p', 1.0},
			},
			"pph": {
				{'i', 1.0},
			},
			"Iao": {
				{'m', 1.0},
			},
			"bet": {
				{'h', 1.0},
			},
			"_Am": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"duc": {
				{'i', 1.0},
			},
			"mar": {
				{'r', 1.0},
			},
			"_Ap": {
				{'p', 1.0},
			},
			"bij": {
				{'a', 1.0},
			},
			"fir": {
				{0, 1.0},
			},
			"_Ag": {
				{'r', 1.0},
			},
			"hra": {
				{'t', 1.0},
			},
			"lia": {
				{'h', 0.3333},
				{'n', 1.0},
			},
			"ait": {
				{'h', 1.0},
			},
			"hag": {
				{0, 1.0},
			},
			"Orp": {
				{'h', 0.5},
				{'a', 1.0},
			},
			"ena": {
				{'t', 1.0},
			},
			"_Iv": {
				{'a', 1.0},
			},
			"tor": {
				{'e', 1.0},
			},
			"gai": {
				{'l', 1.0},
			},
			"ish": {
				{'a', 0.6667},
				{'e', 1.0},
			},
			"mah": {
				{0, 1.0},
			},
			"Ari": {
				{'e', 1.0},
			},
			"_Ad": {
				{'a', 1.0},
			},
			"aly": {
				{'a', 1.0},
			},
			"Est": {
				{'e', 0.3333},
				{'h', 0.6667},
				{'r', 1.0},
			},
			"ore": {
				{'t', 1.0},
			},
			"zab": {
				{'e', 1.0},
			},
			"liz": {
				{'a', 1.0},
			},
			"big": {
				{'a', 1.0},
			},
			"vit": {
				{'a', 1.0},
			},
			"_As": {
				{'h', 0.4},
				{'w', 0.6},
				{'e', 1.0},
			},
			"_Ab": {
				{'i', 1.0},
			},
			"vis": {
				{'h', 1.0},
			},
			"ipp": {
				{'a', 1.0},
			},
			"Ahi": {
				{'n', 1.0},
			},
			"ret": {
				{'h', 0.5},
				{0, 1.0},
			},
			"Ata": {
				{'r', 0.5},
				{'l', 1.0},
			},
			"Aqu": {
				{'i', 1.0},
			},
			"vig": {
				{'a', 1.0},
			},
			"lio": {
				{'n', 1.0},
			},
			"len": {
				{'e', 1.0},
			},
			"ice": {
				{0, 1.0},
			},
			"ita": {
				{'l', 1.0},
			},
			"arr": {
				{'a', 0.5},
				{'i', 1.0},
			},
			"sth": {
				{'e', 1.0},
			},
			"hia": {
				{0, 1.0},
			},
			"abe": {
				{'t', 1.0},
			},
			"Uda": {
				{'r', 1.0},
			},
			"gay": {
				{'i', 1.0},
			},
			"oam": {
				{0, 1.0},
			},
			"_Ef": {
				{'r', 1.0},
			},
			"Ama": {
				{'h', 1.0},
			},
			"phr": {
				{'a', 1.0},
			},
			"_Ia": {
				{'o', 1.0},
			},
			"rph": {
				{'a', 1.0},
			},
			"uni": {
				{'c', 1.0},
			},
			"tar": {
				{'a', 1.0},
			},
			"iga": {
				{'i', 0.5},
				{'y', 1.0},
			},
			"_Av": {
				{'i', 1.0},
			},
			"Ase": {
				{'n', 1.0},
			},
			"Iom": {
				{'e', 1.0},
			},
			"sca": {
				{'h', 1.0},
			},
			"__E": {
				{'f', 0.0769},
				{'u', 0.1538},
				{'d', 0.3077},
				{'p', 0.3846},
				{'s', 0.6154},
				{'m', 0.6923},
				{'v', 0.7692},
				{'l', 1.0},
			},
			"_Ar": {
				{'i', 1.0},
			},
			"_Om": {
				{'i', 0.5},
				{'m', 1.0},
			},
			"ali": {
				{'a', 1.0},
			},
			"Omi": {
				{0, 1.0},
			},
			"est": {
				{'r', 1.0},
			},
			"uci": {
				{'a', 1.0},
			},
			"phi": {
				{'r', 0.5},
				{'a', 1.0},
			},
			"Ize": {
				{'v', 1.0},
			},
			"Ath": {
				{'a', 1.0},
			},
			"Avi": {
				{'y', 0.25},
				{'g', 0.5},
				{'t', 0.75},
				{'s', 1.0},
			},
			"_Ep": {
				{'h', 1.0},
			},
			"hal": {
				{'i', 1.0},
			},
			"bah": {
				{0, 1.0},
			},
			"ndr": {
				{'a', 1.0},
			},
			"Ema": {
				{'l', 1.0},
			},
			"ila": {
				{0, 1.0},
			},
			"ent": {
				{'o', 1.0},
			},
			"ath": {
				{0, 1.0},
			},
			"eba": {
				{0, 1.0},
			},
			"rah": {
				{0, 1.0},
			},
			"App": {
				{'h', 1.0},
			},
			"rra": {
				{0, 1.0},
			},
			"rip": {
				{'p', 1.0},
			},
			"eve": {
				{'l', 1.0},
			},
			"bia": {
				{'h', 1.0},
			},
			"dra": {
				{0, 1.0},
			},
			"sha": {
				{'g', 1.0},
			},
			"ome": {
				{'s', 0.5},
				{0, 1.0},
			},
			"dah": {
				{0, 1.0},
			},
			"Agr": {
				{'i', 1.0},
			},
			"Abi": {
				{'s', 0.1667},
				{'g', 0.3333},
				{'t', 0.5},
				{'a', 0.6667},
				{'l', 0.8333},
				{'j', 1.0},
			},
			"sdu": {
				{'c', 1.0},
			},
			"ude": {
				{0, 1.0},
			},
			"_Ed": {
				{'n', 1.0},
			},
			"nic": {
				{'e', 1.0},
			},
			"Ofi": {
				{'r', 1.0},
			},
			"_Of": {
				{'i', 1.0},
			},
			"ene": {
				{0, 1.0},
			},
			"_El": {
				{'i', 1.0},
			},
			"ien": {
				{'t', 1.0},
			},
			"__O": {
				{'p', 0.1429},
				{'v', 0.2857},
				{'r', 0.5714},
				{'f', 0.7143},
				{'m', 1.0},
			},
			"ail": {
				{0, 1.0},
			},
			"vah": {
				{0, 1.0},
			},
			"_Iz": {
				{'e', 1.0},
			},
			"nda": {
				{'r', 1.0},
			},
			"ile": {
				{'n', 1.0},
			},
			"tal": {
				{'y', 0.3333},
				{0, 1.0},
			},
			"ayi": {
				{'l', 1.0},
			},
			"ino": {
				{'a', 1.0},
			},
			"_Eu": {
				{'n', 1.0},
			},
			"Ali": {
				{'a', 1.0},
			},
			"isa": {
				{'b', 1.0},
			},
			"esd": {
				{'u', 1.0},
			},
			"viy": {
				{'a', 1.0},
			},
			"tri": {
				{'a', 1.0},
			},
			"iah": {
				{0, 1.0},
			},
			"iza": {
				{'b', 1.0},
			},
			"sab": {
				{'e', 1.0},
			},
			"hto": {
				{'r', 1.0},
			},
			"rat": {
				{'h', 0.5},
				{0, 1.0},
			},
			"fra": {
				{'t', 1.0},
			},
			"Eli": {
				{'s', 0.6667},
				{'z', 1.0},
			},
			"yah": {
				{0, 1.0},
			},
			"_Ah": {
				{'i', 1.0},
			},
			"one": {
				{'s', 1.0},
			},
			"str": {
				{'i', 0.6667},
				{'u', 1.0},
			},
			"cah": {
				{0, 1.0},
			},
			"uba": {
				{'h', 1.0},
			},
			"_Ul": {
				{'i', 1.0},
			},
			"Ada": {
				{'h', 1.0},
			},
			"yil": {
				{0, 1.0},
			},
			"dar": {
				{'r', 0.5},
				{'a', 1.0},
			},
			"_Ac": {
				{'h', 1.0},
			},
			"Azu": {
				{'b', 1.0},
			},
			"_Aq": {
				{'u', 1.0},
			},
			"her": {
				{0, 1.0},
			},
			"Asw": {
				{'a', 1.0},
			},
			"chi": {
				{'n', 1.0},
			},
			"Ash": {
				{'t', 1.0},
			},
			"_At": {
				{'h', 0.3333},
				{'a', 1.0},
			},
			"Omm": {
				{'a', 1.0},
			},
			"Iva": {
				{'h', 1.0},
			},
			"bit": {
				{'a', 1.0},
			},
			"ppa": {
				{0, 1.0},
			},
			"pha": {
				{0, 1.0},
			},
			"_Ud": {
				{'a', 1.0},
			},
			"heb": {
				{'a', 1.0},
			},
			"__U": {
				{'l', 0.5},
				{'d', 1.0},
			},
			"ter": {
				{0, 1.0},
			},
			"Edn": {
				{'a', 1.0},
			},
			"_An": {
				{'n', 0.25},
				{'a', 1.0},
			},
			"Ame": {
				{'s', 1.0},
			},
			"Eph": {
				{'r', 1.0},
			},
			"ste": {
				{'r', 1.0},
			},
			"mes": {
				{'t', 0.5},
				{'d', 1.0},
			},
			"vel": {
				{0, 1.0},
			},
			"Efr": {
				{'a', 1.0},
			},
			"she": {
				{'b', 1.0},
			},
			"noa": {
				{'m', 1.0},
			},
			"nes": {
				{'t', 1.0},
			},
			"nto": {
				{0, 1.0},
			},
			"Uli": {
				{'o', 1.0},
			},
			"lya": {
				{'h', 1.0},
			},
			"vie": {
				{'n', 1.0},
			},
			"sht": {
				{'o', 1.0},
			},
			"rie": {
				{'l', 1.0},
			},
			"ria": {
				{0, 1.0},
			},
			"hin": {
				{'o', 1.0},
			},
			"zub": {
				{'a', 1.0},
			},
		}
	case 'M':
		return map[string][]nameFragment{
			"hud": {
				{0, 1.0},
			},
			"abo": {
				{'d', 1.0},
			},
			"Abi": {
				{'h', 0.1111},
				{'e', 0.2222},
				{'d', 0.3333},
				{'r', 0.4444},
				{'j', 0.5556},
				{'s', 0.6667},
				{'a', 0.7778},
				{'m', 1.0},
			},
			"ima": {
				{'e', 0.5},
				{0, 1.0},
			},
			"hea": {
				{0, 1.0},
			},
			"sha": {
				{'i', 0.25},
				{'y', 0.5},
				{'l', 0.75},
				{0, 1.0},
			},
			"dra": {
				{'s', 1.0},
			},
			"dne": {
				{'g', 1.0},
			},
			"dai": {
				{0, 1.0},
			},
			"_Ez": {
				{'a', 0.25},
				{'e', 0.75},
				{'r', 1.0},
			},
			"mos": {
				{0, 1.0},
			},
			"_Eb": {
				{'e', 1.0},
			},
			"__O": {
				{'m', 0.1111},
				{'p', 0.2222},
				{'l', 0.2778},
				{'b', 0.3333},
				{'n', 0.4444},
				{'v', 0.5556},
				{'t', 0.6667},
				{'f', 0.7778},
				{'h', 0.8333},
				{'d', 0.8889},
				{'s', 1.0},
			},
			"Aza": {
				{'r', 0.625},
				{'z', 1.0},
			},
			"_Of": {
				{'i', 0.5},
				{'r', 1.0},
			},
			"_Il": {
				{'l', 1.0},
			},
			"rch": {
				{'e', 1.0},
			},
			"Aha": {
				{'b', 0.5},
				{'r', 1.0},
			},
			"tta": {
				{'i', 0.6667},
				{'y', 1.0},
			},
			"Ele": {
				{'a', 1.0},
			},
			"_Im": {
				{'m', 1.0},
			},
			"_Et": {
				{'h', 1.0},
			},
			"Ama": {
				{'r', 0.6667},
				{'l', 1.0},
			},
			"itt": {
				{'a', 1.0},
			},
			"_Ot": {
				{'h', 1.0},
			},
			"has": {
				{0, 1.0},
			},
			"ych": {
				{'u', 1.0},
			},
			"chu": {
				{'s', 1.0},
			},
			"ira": {
				{'n', 0.25},
				{'m', 1.0},
			},
			"_Om": {
				{'a', 0.5},
				{'r', 1.0},
			},
			"nue": {
				{'l', 1.0},
			},
			"_Ar": {
				{'c', 0.1},
				{'t', 0.2},
				{'i', 0.7},
				{'e', 0.9},
				{'a', 1.0},
			},
			"__U": {
				{'d', 0.0769},
				{'r', 0.6923},
				{'z', 1.0},
			},
			"saf": {
				{0, 1.0},
			},
			"Ero": {
				{'d', 1.0},
			},
			"Uzz": {
				{'i', 1.0},
			},
			"aus": {
				{0, 1.0},
			},
			"ioc": {
				{'h', 1.0},
			},
			"Ala": {
				{'m', 1.0},
			},
			"lls": {
				{'m', 1.0},
			},
			"fal": {
				{'e', 1.0},
			},
			"imu": {
				{'s', 1.0},
			},
			"ary": {
				{'a', 1.0},
			},
			"Eut": {
				{'y', 1.0},
			},
			"Ure": {
				{'s', 1.0},
			},
			"Asa": {
				{'f', 0.3333},
				{0, 0.6667},
				{'p', 1.0},
			},
			"han": {
				{0, 1.0},
			},
			"don": {
				{'i', 0.75},
				{0, 1.0},
			},
			"har": {
				{'o', 0.5},
				{0, 1.0},
			},
			"iph": {
				{'o', 0.3333},
				{'e', 0.6667},
				{'a', 1.0},
			},
			"hom": {
				{'a', 1.0},
			},
			"Oha": {
				{'d', 1.0},
			},
			"uel": {
				{0, 1.0},
			},
			"yyo": {
				{'v', 1.0},
			},
			"esi": {
				{'m', 0.5},
				{'p', 1.0},
			},
			"era": {
				{'i', 1.0},
			},
			"lij": {
				{'a', 1.0},
			},
			"Iyy": {
				{'o', 1.0},
			},
			"vim": {
				{'e', 0.5},
				{'a', 1.0},
			},
			"ele": {
				{'t', 0.5},
				{'c', 0.75},
				{'k', 1.0},
			},
			"aim": {
				{0, 1.0},
			},
			"Aba": {
				{'d', 1.0},
			},
			"her": {
				{0, 1.0},
			},
			"_Aq": {
				{'u', 1.0},
			},
			"Omr": {
				{'i', 1.0},
			},
			"_Ac": {
				{'h', 1.0},
			},
			"aph": {
				{0, 0.5},
				{'r', 1.0},
			},
			"ama": {
				{'n', 0.3333},
				{'r', 1.0},
			},
			"ode": {
				{'l', 1.0},
			},
			"Adi": {
				{'n', 1.0},
			},
			"Ode": {
				{'d', 1.0},
			},
			"Eli": {
				{'f', 0.1538},
				{'e', 0.2308},
				{'h', 0.3077},
				{'s', 0.3846},
				{'j', 0.4615},
				{'y', 0.5385},
				{'o', 0.6154},
				{'a', 0.6923},
				{'u', 0.7692},
				{'p', 0.9231},
				{0, 1.0},
			},
			"__I": {
				{'s', 0.381},
				{'c', 0.4286},
				{'r', 0.4762},
				{'t', 0.7619},
				{'m', 0.8095},
				{'l', 0.8571},
				{'y', 0.9048},
				{'o', 0.9524},
				{'x', 1.0},
			},
			"rij": {
				{'a', 1.0},
			},
			"bid": {
				{'a', 1.0},
			},
			"ina": {
				{0, 0.5},
				{'h', 1.0},
			},
			"mma": {
				{'n', 1.0},
			},
			"sap": {
				{'h', 1.0},
			},
			"Isa": {
				{'a', 0.3333},
				{'i', 1.0},
			},
			"Ove": {
				{'d', 1.0},
			},
			"bis": {
				{'h', 1.0},
			},
			"rai": {
				{0, 0.3333},
				{'m', 1.0},
			},
			"_Es": {
				{'a', 0.6667},
				{'d', 1.0},
			},
			"mmi": {
				{'e', 1.0},
			},
			"_Io": {
				{'g', 1.0},
			},
			"_Ov": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"uil": {
				{'a', 1.0},
			},
			"mer": {
				{'a', 1.0},
			},
			"Eit": {
				{'a', 1.0},
			},
			"pap": {
				{'h', 1.0},
			},
			"lih": {
				{'u', 1.0},
			},
			"ale": {
				{'t', 1.0},
			},
			"zaz": {
				{'i', 0.3333},
				{'e', 0.6667},
				{'y', 1.0},
			},
			"lli": {
				{'s', 1.0},
			},
			"xan": {
				{'d', 1.0},
			},
			"Imm": {
				{'a', 1.0},
			},
			"zar": {
				{'i', 0.2857},
				{'e', 0.4286},
				{'y', 0.5714},
				{0, 1.0},
			},
			"Avr": {
				{'a', 1.0},
			},
			"stu": {
				{'s', 1.0},
			},
			"mus": {
				{0, 1.0},
			},
			"Ehu": {
				{'d', 1.0},
			},
			"yya": {
				{'h', 1.0},
			},
			"_Ad": {
				{'l', 0.1},
				{'i', 0.4},
				{'o', 0.7},
				{'a', 1.0},
			},
			"Ari": {
				{0, 0.2},
				{'e', 0.6},
				{'o', 0.8},
				{'d', 1.0},
			},
			"yov": {
				{0, 1.0},
			},
			"sau": {
				{0, 1.0},
			},
			"saw": {
				{0, 1.0},
			},
			"_Ef": {
				{'r', 1.0},
			},
			"hor": {
				{'u', 1.0},
			},
			"ita": {
				{'n', 1.0},
			},
			"tay": {
				{0, 1.0},
			},
			"Aqu": {
				{'i', 1.0},
			},
			"Ell": {
				{'i', 1.0},
			},
			"ipp": {
				{'a', 1.0},
			},
			"Amr": {
				{'a', 1.0},
			},
			"eze": {
				{'r', 1.0},
			},
			"Epa": {
				{'p', 1.0},
			},
			"iya": {
				{'h', 1.0},
			},
			"lec": {
				{'h', 1.0},
			},
			"xio": {
				{'l', 1.0},
			},
			"zer": {
				{0, 1.0},
			},
			"mal": {
				{0, 1.0},
			},
			"dam": {
				{0, 1.0},
			},
			"rew": {
				{0, 1.0},
			},
			"nai": {
				{0, 0.5},
				{'a', 1.0},
			},
			"dre": {
				{'w', 1.0},
			},
			"_Ob": {
				{'e', 1.0},
			},
			"nan": {
				{0, 0.3333},
				{'i', 1.0},
			},
			"ziy": {
				{'y', 1.0},
			},
			"ifa": {
				{'l', 1.0},
			},
			"zel": {
				{0, 1.0},
			},
			"hra": {
				{'s', 0.3333},
				{'h', 0.6667},
				{'i', 1.0},
			},
			"ola": {
				{'n', 1.0},
			},
			"lva": {
				{'h', 1.0},
			},
			"mra": {
				{'m', 1.0},
			},
			"him": {
				{0, 1.0},
			},
			"Ado": {
				{'n', 1.0},
			},
			"Agr": {
				{'i', 1.0},
			},
			"mit": {
				{'t', 1.0},
			},
			"eta": {
				{'s', 1.0},
			},
			"iol": {
				{'a', 1.0},
			},
			"vad": {
				{'y', 1.0},
			},
			"_En": {
				{'o', 1.0},
			},
			"Edo": {
				{'m', 1.0},
			},
			"tax": {
				{'e', 1.0},
			},
			"tal": {
				{'y', 1.0},
			},
			"ner": {
				{0, 1.0},
			},
			"hay": {
				{0, 1.0},
			},
			"iez": {
				{'e', 1.0},
			},
			"ene": {
				{'z', 1.0},
			},
			"rxe": {
				{'s', 1.0},
			},
			"Ofi": {
				{'r', 1.0},
			},
			"Ira": {
				{0, 1.0},
			},
			"kan": {
				{'a', 1.0},
			},
			"__E": {
				{'s', 0.0652},
				{'f', 0.1087},
				{'h', 0.1304},
				{'d', 0.1739},
				{'r', 0.2391},
				{'t', 0.2609},
				{'n', 0.3261},
				{'u', 0.3478},
				{'m', 0.3696},
				{'p', 0.413},
				{'b', 0.4348},
				{'l', 0.8696},
				{'z', 0.9565},
				{'i', 1.0},
			},
			"aia": {
				{'s', 0.3333},
				{'h', 1.0},
			},
			"rid": {
				{'a', 1.0},
			},
			"kie": {
				{'l', 1.0},
			},
			"ram": {
				{0, 1.0},
			},
			"Amo": {
				{'s', 1.0},
			},
			"lip": {
				{'h', 1.0},
			},
			"ahu": {
				{0, 1.0},
			},
			"Azr": {
				{'i', 1.0},
			},
			"och": {
				{0, 1.0},
			},
			"tar": {
				{'i', 1.0},
			},
			"aha": {
				{'m', 1.0},
			},
			"Uri": {
				{'a', 0.1667},
				{'e', 0.3333},
				{'y', 0.6667},
				{'j', 0.8333},
				{0, 1.0},
			},
			"zri": {
				{0, 0.5},
				{'e', 1.0},
			},
			"Olh": {
				{'a', 1.0},
			},
			"nos": {
				{0, 0.5},
				{'h', 1.0},
			},
			"ras": {
				{0, 0.6667},
				{'t', 1.0},
			},
			"Ath": {
				{'a', 1.0},
			},
			"Arc": {
				{'h', 1.0},
			},
			"thi": {
				{'e', 1.0},
			},
			"bie": {
				{'l', 1.0},
			},
			"vra": {
				{'m', 0.5},
				{'h', 1.0},
			},
			"Obe": {
				{'d', 1.0},
			},
			"dom": {
				{0, 1.0},
			},
			"are": {
				{'l', 1.0},
			},
			"Ich": {
				{'a', 1.0},
			},
			"Era": {
				{'n', 0.5},
				{'s', 1.0},
			},
			"ego": {
				{0, 1.0},
			},
			"bim": {
				{'e', 0.5},
				{'a', 1.0},
			},
			"ppa": {
				{0, 1.0},
			},
			"eaz": {
				{'a', 1.0},
			},
			"_Ix": {
				{'i', 1.0},
			},
			"ech": {
				{0, 1.0},
			},
			"Esa": {
				{'w', 0.5},
				{'u', 1.0},
			},
			"Art": {
				{'a', 1.0},
			},
			"Ixi": {
				{'o', 1.0},
			},
			"ieh": {
				{0, 1.0},
			},
			"ste": {
				{0, 1.0},
			},
			"nez": {
				{'e', 1.0},
			},
			"lex": {
				{'a', 1.0},
			},
			"ihu": {
				{0, 1.0},
			},
			"viy": {
				{'a', 1.0},
			},
			"dad": {
				{0, 1.0},
			},
			"smu": {
				{'s', 1.0},
			},
			"edn": {
				{'e', 1.0},
			},
			"tyc": {
				{'h', 1.0},
			},
			"Amn": {
				{'o', 1.0},
			},
			"ino": {
				{0, 1.0},
			},
			"shv": {
				{'i', 1.0},
			},
			"chi": {
				{'m', 1.0},
			},
			"Urb": {
				{'a', 1.0},
			},
			"Are": {
				{'l', 0.5},
				{'t', 1.0},
			},
			"Eth": {
				{'a', 1.0},
			},
			"non": {
				{0, 1.0},
			},
			"nia": {
				{'s', 1.0},
			},
			"bne": {
				{'r', 1.0},
			},
			"yah": {
				{'u', 0.3},
				{0, 1.0},
			},
			"fra": {
				{'i', 0.3333},
				{'h', 0.6667},
				{'y', 1.0},
			},
			"Ana": {
				{'n', 0.4286},
				{'h', 0.5714},
				{'i', 0.7143},
				{'t', 1.0},
			},
			"sac": {
				{'h', 1.0},
			},
			"ved": {
				{0, 1.0},
			},
			"lif": {
				{'e', 0.5},
				{'a', 1.0},
			},
			"hus": {
				{0, 1.0},
			},
			"ael": {
				{0, 1.0},
			},
			"Eir": {
				{'a', 1.0},
			},
			"Udh": {
				{'o', 1.0},
			},
			"ekh": {
				{0, 1.0},
			},
			"Itt": {
				{'a', 1.0},
			},
			"Ach": {
				{'a', 0.5},
				{'i', 0.75},
				{0, 1.0},
			},
			"_Al": {
				{'p', 0.2},
				{'l', 0.4},
				{'e', 0.6},
				{'a', 0.8},
				{'v', 1.0},
			},
			"Alp": {
				{'h', 1.0},
			},
			"llo": {
				{0, 0.5},
				{'n', 1.0},
			},
			"nah": {
				{0, 1.0},
			},
			"ias": {
				{0, 1.0},
			},
			"dya": {
				{'h', 1.0},
			},
			"ism": {
				{'u', 1.0},
			},
			"lau": {
				{'s', 1.0},
			},
			"hir": {
				{0, 1.0},
			},
			"Adl": {
				{'a', 1.0},
			},
			"add": {
				{'o', 1.0},
			},
			"nir": {
				{'a', 1.0},
			},
			"vne": {
				{'r', 1.0},
			},
			"__A": {
				{'n', 0.0909},
				{'g', 0.101},
				{'z', 0.2121},
				{'a', 0.2222},
				{'k', 0.2323},
				{'r', 0.3333},
				{'q', 0.3434},
				{'m', 0.4343},
				{'b', 0.596},
				{'l', 0.6465},
				{'s', 0.697},
				{'c', 0.7374},
				{'v', 0.8586},
				{'t', 0.8788},
				{'h', 0.899},
				{'d', 1.0},
			},
			"ida": {
				{'i', 0.3333},
				{'n', 1.0},
			},
			"Eln": {
				{'a', 1.0},
			},
			"_Op": {
				{'h', 1.0},
			},
			"dla": {
				{'i', 1.0},
			},
			"_As": {
				{'a', 0.6},
				{'h', 1.0},
			},
			"lda": {
				{'d', 1.0},
			},
			"ela": {
				{'u', 1.0},
			},
			"din": {
				{'o', 0.3333},
				{'a', 1.0},
			},
			"hai": {
				{0, 0.6667},
				{'c', 1.0},
			},
			"shm": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"rel": {
				{0, 0.5},
				{'i', 1.0},
			},
			"All": {
				{'o', 1.0},
			},
			"Ebe": {
				{'n', 1.0},
			},
			"icu": {
				{'s', 1.0},
			},
			"ret": {
				{'a', 1.0},
			},
			"hma": {
				{'e', 1.0},
			},
			"vis": {
				{'h', 1.0},
			},
			"Ost": {
				{'a', 1.0},
			},
			"_Ak": {
				{'o', 1.0},
			},
			"ena": {
				{'i', 1.0},
			},
			"oni": {
				{'r', 0.3333},
				{'y', 0.6667},
				{'j', 1.0},
			},
			"sra": {
				{'e', 1.0},
			},
			"ach": {
				{'a', 1.0},
			},
			"lha": {
				{'s', 1.0},
			},
			"_Ag": {
				{'r', 1.0},
			},
			"fir": {
				{0, 1.0},
			},
			"Oma": {
				{'r', 1.0},
			},
			"riy": {
				{0, 0.5},
				{'a', 1.0},
			},
			"_Aa": {
				{'r', 1.0},
			},
			"zie": {
				{'l', 1.0},
			},
			"azi": {
				{'a', 1.0},
			},
			"_Am": {
				{'a', 0.3333},
				{'o', 0.4444},
				{'i', 0.6667},
				{'r', 0.7778},
				{'n', 0.8889},
				{'m', 1.0},
			},
			"hae": {
				{'u', 1.0},
			},
			"_Ed": {
				{'e', 0.5},
				{'o', 1.0},
			},
			"mie": {
				{'l', 1.0},
			},
			"vid": {
				{'a', 1.0},
			},
			"del": {
				{0, 1.0},
			},
			"bad": {
				{'d', 1.0},
			},
			"bia": {
				{'h', 1.0},
			},
			"Ofr": {
				{'a', 1.0},
			},
			"saa": {
				{'c', 1.0},
			},
			"ath": {
				{'a', 0.5},
				{0, 1.0},
			},
			"rah": {
				{'a', 0.5},
				{0, 1.0},
			},
			"_Ol": {
				{'h', 1.0},
			},
			"vsh": {
				{'a', 1.0},
			},
			"vah": {
				{0, 1.0},
			},
			"Eza": {
				{'r', 1.0},
			},
			"Abs": {
				{'a', 1.0},
			},
			"_El": {
				{'i', 0.65},
				{'a', 0.7},
				{'o', 0.75},
				{'l', 0.8},
				{'n', 0.85},
				{'e', 0.9},
				{'d', 0.95},
				{'k', 1.0},
			},
			"dan": {
				{0, 1.0},
			},
			"Ita": {
				{'m', 1.0},
			},
			"niy": {
				{'a', 1.0},
			},
			"rae": {
				{'l', 1.0},
			},
			"alo": {
				{'m', 1.0},
			},
			"_Av": {
				{'i', 0.6667},
				{'r', 0.8333},
				{'n', 0.9167},
				{'s', 1.0},
			},
			"iak": {
				{'i', 1.0},
			},
			"had": {
				{0, 1.0},
			},
			"ban": {
				{0, 1.0},
			},
			"_Oh": {
				{'a', 1.0},
			},
			"Amm": {
				{'i', 1.0},
			},
			"pho": {
				{'r', 1.0},
			},
			"ndr": {
				{'e', 1.0},
			},
			"rio": {
				{'c', 1.0},
			},
			"Avs": {
				{'h', 1.0},
			},
			"hal": {
				{'i', 0.3333},
				{'o', 0.6667},
				{'e', 1.0},
			},
			"cha": {
				{'b', 0.25},
				{'n', 0.5},
				{'i', 0.75},
				{'r', 1.0},
			},
			"_Ep": {
				{'a', 0.5},
				{'h', 1.0},
			},
			"And": {
				{'r', 1.0},
			},
			"Avi": {
				{'s', 0.125},
				{0, 0.25},
				{'m', 0.5},
				{'y', 0.625},
				{'h', 0.75},
				{'d', 0.875},
				{'r', 1.0},
			},
			"oen": {
				{'a', 1.0},
			},
			"rod": {
				{'e', 1.0},
			},
			"kim": {
				{0, 1.0},
			},
			"lka": {
				{'n', 1.0},
			},
			"_An": {
				{'a', 0.7778},
				{'d', 0.8889},
				{'n', 1.0},
			},
			"lam": {
				{0, 0.5},
				{'a', 1.0},
			},
			"pha": {
				{'e', 0.5},
				{'l', 1.0},
			},
			"mri": {
				{0, 1.0},
			},
			"hni": {
				{'e', 1.0},
			},
			"hie": {
				{'l', 1.0},
			},
			"nij": {
				{'a', 1.0},
			},
			"ady": {
				{'a', 1.0},
			},
			"Ash": {
				{0, 0.5},
				{'e', 1.0},
			},
			"yim": {
				{0, 1.0},
			},
			"oru": {
				{'s', 1.0},
			},
			"res": {
				{'t', 1.0},
			},
			"lya": {
				{'h', 0.5},
				{0, 1.0},
			},
			"rta": {
				{'x', 1.0},
			},
			"zia": {
				{'h', 1.0},
			},
			"Efr": {
				{'a', 1.0},
			},
			"Eph": {
				{'r', 1.0},
			},
			"vir": {
				{'a', 1.0},
			},
			"ddo": {
				{'n', 1.0},
			},
			"bir": {
				{'a', 1.0},
			},
			"Ova": {
				{'d', 1.0},
			},
			"zek": {
				{'i', 1.0},
			},
			"iah": {
				{0, 1.0},
			},
			"tus": {
				{0, 1.0},
			},
			"ori": {
				{'a', 1.0},
			},
			"Isr": {
				{'a', 1.0},
			},
			"ben": {
				{'e', 1.0},
			},
			"Ada": {
				{'l', 0.6667},
				{'m', 1.0},
			},
			"_Ah": {
				{'a', 1.0},
			},
			"sal": {
				{'o', 1.0},
			},
			"nat": {
				{0, 0.3333},
				{'h', 1.0},
			},
			"nde": {
				{'r', 1.0},
			},
			"_Er": {
				{'o', 0.3333},
				{'a', 1.0},
			},
			"sta": {
				{'r', 1.0},
			},
			"xes": {
				{0, 1.0},
			},
			"bod": {
				{0, 1.0},
			},
			"tam": {
				{'a', 1.0},
			},
			"iel": {
				{0, 1.0},
			},
			"aeu": {
				{'s', 1.0},
			},
			"_Is": {
				{'s', 0.125},
				{'r', 0.25},
				{'a', 0.625},
				{'h', 1.0},
			},
			"and": {
				{'e', 1.0},
			},
			"_It": {
				{'h', 0.5},
				{'a', 0.6667},
				{'t', 1.0},
			},
			"tan": {
				{0, 1.0},
			},
			"ari": {
				{'a', 1.0},
			},
			"ham": {
				{0, 0.6667},
				{'a', 1.0},
			},
			"_Az": {
				{'a', 0.7273},
				{'e', 0.8182},
				{'r', 1.0},
			},
			"osh": {
				{0, 1.0},
			},
			"iyy": {
				{'a', 1.0},
			},
			"Eno": {
				{'c', 0.3333},
				{'s', 1.0},
			},
			"aze": {
				{'l', 1.0},
			},
			"tha": {
				{'i', 0.2},
				{'n', 0.6},
				{'m', 0.8},
				{'l', 1.0},
			},
			"der": {
				{0, 1.0},
			},
			"ana": {
				{'h', 1.0},
			},
			"dal": {
				{'y', 0.5},
				{'i', 1.0},
			},
			"ian": {
				{0, 1.0},
			},
			"rus": {
				{0, 1.0},
			},
			"nna": {
				{'s', 1.0},
			},
			"_Ab": {
				{'s', 0.0625},
				{'n', 0.125},
				{'i', 0.6875},
				{'r', 0.8125},
				{'e', 0.9375},
				{'a', 1.0},
			},
			"eki": {
				{'e', 1.0},
			},
			"Iog": {
				{'o', 1.0},
			},
			"oma": {
				{'r', 1.0},
			},
			"aly": {
				{'a', 1.0},
			},
			"Ami": {
				{0, 0.5},
				{'t', 1.0},
			},
			"Ela": {
				{'m', 1.0},
			},
			"ish": {
				{'a', 1.0},
			},
			"Ith": {
				{'a', 0.6667},
				{'i', 1.0},
			},
			"Alv": {
				{'a', 1.0},
			},
			"Aze": {
				{'l', 1.0},
			},
			"Eld": {
				{'a', 1.0},
			},
			"sai": {
				{'a', 1.0},
			},
			"aro": {
				{'n', 1.0},
			},
			"Ata": {
				{'l', 1.0},
			},
			"_Ur": {
				{'i', 0.75},
				{'e', 0.875},
				{'b', 1.0},
			},
			"Abr": {
				{'a', 1.0},
			},
			"ded": {
				{0, 1.0},
			},
			"_Uz": {
				{'z', 1.0},
			},
			"gri": {
				{'p', 1.0},
			},
			"tai": {
				{0, 1.0},
			},
			"Ezr": {
				{'a', 1.0},
			},
			"jah": {
				{0, 1.0},
			},
			"Ann": {
				{'a', 1.0},
			},
			"ogo": {
				{'r', 1.0},
			},
			"bij": {
				{'a', 1.0},
			},
			"mar": {
				{'i', 0.1667},
				{'y', 0.3333},
				{0, 1.0},
			},
			"nas": {
				{0, 1.0},
			},
			"Esd": {
				{'r', 1.0},
			},
			"_Ir": {
				{'a', 1.0},
			},
			"zra": {
				{0, 1.0},
			},
			"ssa": {
				{'c', 1.0},
			},
			"_Od": {
				{'e', 1.0},
			},
			"rip": {
				{'p', 1.0},
			},
			"Ako": {
				{'r', 1.0},
			},
			"kor": {
				{'i', 1.0},
			},
			"eus": {
				{0, 1.0},
			},
			"ayi": {
				{'m', 1.0},
			},
			"aza": {
				{'r', 1.0},
			},
			"hab": {
				{0, 0.5},
				{'o', 1.0},
			},
			"liy": {
				{'y', 1.0},
			},
			"Aar": {
				{'o', 1.0},
			},
			"One": {
				{'s', 1.0},
			},
			"Ill": {
				{'s', 1.0},
			},
			"iud": {
				{0, 1.0},
			},
			"lea": {
				{'z', 1.0},
			},
			"bed": {
				{'n', 0.5},
				{0, 1.0},
			},
			"anu": {
				{'e', 1.0},
			},
			"_Os": {
				{'h', 0.5},
				{'t', 1.0},
			},
			"phr": {
				{'a', 1.0},
			},
			"bih": {
				{'u', 1.0},
			},
			"Ish": {
				{'v', 0.3333},
				{'m', 1.0},
			},
			"noc": {
				{'h', 1.0},
			},
			"rba": {
				{'n', 1.0},
			},
			"bsa": {
				{'l', 1.0},
			},
			"ila": {
				{0, 1.0},
			},
			"lom": {
				{0, 1.0},
			},
			"liu": {
				{'d', 1.0},
			},
			"neg": {
				{'o', 1.0},
			},
			"_On": {
				{'e', 1.0},
			},
			"lph": {
				{'a', 1.0},
			},
			"sdr": {
				{'a', 1.0},
			},
			"phi": {
				{'r', 1.0},
			},
			"est": {
				{'e', 1.0},
			},
			"gor": {
				{'i', 1.0},
			},
			"ali": {
				{'a', 1.0},
			},
			"hel": {
				{'a', 0.5},
				{'e', 1.0},
			},
			"_Ud": {
				{'h', 1.0},
			},
			"che": {
				{'l', 1.0},
			},
			"azy": {
				{'a', 1.0},
			},
			"_At": {
				{'h', 0.5},
				{'a', 1.0},
			},
			"ria": {
				{'h', 0.4286},
				{'n', 0.8571},
				{0, 1.0},
			},
			"lie": {
				{'z', 1.0},
			},
			"rie": {
				{'h', 0.25},
				{'l', 1.0},
			},
			"nes": {
				{'i', 1.0},
			},
			"she": {
				{'r', 0.5},
				{'a', 1.0},
			},
			"nie": {
				{'l', 1.0},
			},
			"fel": {
				{'e', 1.0},
			},
			"let": {
				{0, 1.0},
			},
			"erx": {
				{'e', 1.0},
			},
			"_Eu": {
				{'t', 1.0},
			},
			"Elk": {
				{'a', 1.0},
			},
			"_Iy": {
				{'y', 1.0},
			},
			"ioe": {
				{'n', 1.0},
			},
			"lek": {
				{'h', 1.0},
			},
			"eli": {
				{0, 1.0},
			},
			"Ede": {
				{'r', 1.0},
			},
			"tho": {
				{'l', 1.0},
			},
			"hvi": {
				{0, 1.0},
			},
			"ime": {
				{'l', 1.0},
			},
			"bra": {
				{'h', 0.5},
				{'m', 1.0},
			},
			"mel": {
				{'e', 1.0},
			},
			"Eze": {
				{'r', 0.5},
				{'k', 1.0},
			},
			"exa": {
				{'n', 1.0},
			},
			"aki": {
				{'m', 1.0},
			},
			"man": {
				{'d', 0.3333},
				{'u', 1.0},
			},
			"ast": {
				{'u', 1.0},
			},
			"zya": {
				{'h', 1.0},
			},
			"uty": {
				{'c', 1.0},
			},
			"hme": {
				{'r', 1.0},
			},
			"Oph": {
				{'r', 0.5},
				{'i', 1.0},
			},
			"Ale": {
				{'x', 1.0},
			},
			"qui": {
				{'l', 1.0},
			},
			"aac": {
				{0, 1.0},
			},
			"sim": {
				{'u', 1.0},
			},
			"zzi": {
				{0, 0.25},
				{'e', 0.5},
				{'a', 0.75},
				{'y', 1.0},
			},
			"ran": {
				{0, 1.0},
			},
			"mno": {
				{'n', 1.0},
			},
			"lai": {
				{0, 1.0},
			},
			"rya": {
				{'h', 1.0},
			},
			"___": {
				{'E', 0.2335},
				{'O', 0.3249},
				{'I', 0.4315},
				{'A', 0.934},
				{'U', 1.0},
			},
			"cus": {
				{0, 1.0},
			},
			"Iss": {
				{'a', 1.0},
			},
			"Abe": {
				{'l', 0.5},
				{'d', 1.0},
			},
			"Avn": {
				{'e', 1.0},
			},
			"dho": {
				{'m', 1.0},
			},
			"_Eh": {
				{'u', 1.0},
			},
			"_Ei": {
				{'t', 0.5},
				{'r', 1.0},
			},
			"oll": {
				{'o', 1.0},
			},
			"lis": {
				{'h', 0.5},
				{'m', 1.0},
			},
			"tas": {
				{0, 1.0},
			},
			"ija": {
				{'h', 1.0},
			},
			"sip": {
				{'h', 1.0},
			},
			"ron": {
				{0, 1.0},
			},
			"Osh": {
				{'e', 1.0},
			},
			"ray": {
				{'i', 1.0},
			},
			"aic": {
				{'u', 1.0},
			},
			"Abn": {
				{'e', 1.0},
			},
			"lan": {
				{'d', 1.0},
			},
			"hol": {
				{'l', 1.0},
			},
			"Emm": {
				{'a', 1.0},
			},
			"ani": {
				{'a', 0.5},
				{0, 1.0},
			},
			"_Ic": {
				{'h', 1.0},
			},
			"phe": {
				{'l', 1.0},
			},
			"lio": {
				{'e', 1.0},
			},
			"lon": {
				{0, 1.0},
			},
			"lsm": {
				{'u', 1.0},
			},
			"lna": {
				{'t', 1.0},
			},
			"axe": {
				{'r', 1.0},
			},
			"thn": {
				{'i', 1.0},
			},
			"Oth": {
				{'o', 0.5},
				{'n', 1.0},
			},
			"_Em": {
				{'m', 1.0},
			},
			"ife": {
				{'l', 1.0},
			},
			"Ara": {
				{'n', 1.0},
			},
			"Elo": {
				{'n', 1.0},
			},
			"mae": {
				{'l', 1.0},
			},
			"xer": {
				{'x', 1.0},
			},
			"lia": {
				{0, 0.3333},
				{'k', 0.6667},
				{'h', 1.0},
			},
			"vih": {
				{'u', 1.0},
			},
			"bel": {
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
