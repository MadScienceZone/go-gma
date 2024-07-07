/*
########################################################################################
#  __                                                                                  #
# /__ _                                                                                #
# \_|(_)                                                                               #
#  _______  _______  _______             _______     _______  ______      _______      #
# (  ____ \(       )(  ___  ) Game      (  ____ \   / ___   )/ ___  \    (  __   )     #
# | (    \/| () () || (   ) | Master's  | (    \/   \/   )  |\/   \  \   | (  )  |     #
# | |      | || || || (___) | Assistant | (____         /   )   ___) /   | | /   |     #
# | | ____ | |(_)| ||  ___  | (Go Port) (_____ \      _/   /   (___ (    | (/ /) |     #
# | | \_  )| |   | || (   ) |                 ) )    /   _/        ) \   |   / | |     #
# | (___) || )   ( || )   ( | Mapper    /\____) ) _ (   (__/\/\___/  / _ |  (__) |     #
# (_______)|/     \||/     \| Client    \______/ (_)\_______/\______/ (_)(_______)     #
#                                                                                      #
########################################################################################
*/

//
// Unit tests and examples for namegen.
//

package namegen

import (
	"fmt"
	"testing"

	"github.com/MadScienceZone/go-gma/v5/dice"
)

func ExampleCulture() {
	names, err := Generate(Azlanti{}, 'M', 10)
	if err != nil {
		panic(fmt.Sprintf("oh, no, I can't generate names: %v", err))
	}
	fmt.Printf("Azlanti names: %v\n", names)
}

func ExampleCultures() {
	az := Cultures["Azlanti"]
	fmt.Printf("The %s culture defines name genders %v\n", az.Name(), az.Genders())

	names, err := Generate(az, 'M', 1)
	if err != nil {
		fmt.Printf("I couldn't generate a name (%v)\n", err)
	} else {
		fmt.Printf("Example name: %s\n", names[0])
	}
}

func ExampleGenerate() {
	names, err := Generate(Kellid{}, 'F', 5, WithStartingLetter('K'))
	if err != nil {
		fmt.Printf("Error generating names: %v\n", err)
	} else {
		fmt.Println("Five example Kellid names starting with 'K':")
		for i, name := range names {
			fmt.Printf("#%d: %s\n", i+1, name)
		}
	}
}

func ExampleGenerateWithSurnames() {
	names, err := GenerateWithSurnames(Taldan{}, 'M', 5)
	if err != nil {
		fmt.Printf("Error generating names: %v\n", err)
	} else {
		fmt.Println("Five example Taldan names:")
		for i, name := range names {
			fmt.Printf("#%d: %s %s\n", i+1, name[0], name[1])
		}
	}
}

func TestNames(t *testing.T) {
	dr, err := dice.NewDieRoller(dice.WithSeed(123456))
	if err != nil {
		t.Fatalf("can't create die roller: %v", err)
	}

	for i, test := range []struct {
		Culture  string
		Number   int
		Genders  string
		Results  [][]string
		SResults [][][]string
	}{
		{"Azlanti", 10, "FM", [][]string{
			{"Aswaither", "Aquila", "Avishag", "Ommarrin", "Ariel", "Orpha", "Achinoam", "Amesducias", "Ulionestrude", "Agrippa"},
			{"Anathalia", "Efrayim", "Azazyahu", "Osher", "Oshea", "Abimael", "Ephras", "Abisha", "Avram", "Azaziah"},
		}, nil},
		{"Bekyar", 3, "FM", [][]string{
			{"Kija", "Zakiyyah", "Nyota"},
			{"Muhsina", "Maulo", "Yakubu"},
		}, nil},
		{"Bonuwat", 4, "FM", [][]string{
			{"Kanisha", "Otesha", "Therani", "Inithia"},
			{"Issa", "Koman", "Ashon", "Kobe"},
		}, nil},
		{"Chelaxian", 4, "FM", [][]string{
			{"Excina", "Steridia", "Rufidia", "Egle"},
			{"Adepphilina", "Marcisio", "Pusillus", "Cosco"},
		}, [][][]string{
			{{"Sarria", "Cator"}, {"Iacolater", "Censis"}, {"Exuperens", "Pera"}, {"Estes", "Bursio"}},
			{{"Vola", "Falto"}, {"Baldo", "Asellaecus"}, {"Volus", "Brutus"}, {"Aiaciliano", "Dola"}},
		}},
		{"Dwarf", 4, "FM", [][]string{
			{"Walgise", "Rosa", "Burgardra", "Engra"},
			{"Siegward", "Guntram", "Waldwin", "Romuald"},
		}, nil},
		{"Elf", 4, "FM", [][]string{
			{"Thusara", "Caseldra", "Jaslara", "Sushlara"},
			{"Riguel", "Kasbiel", "Nanael", "Dabriel"},
		}, nil},
		{"Erutaki", 4, "FM", [][]string{
			{"Qaiyaaniigiaq", "Pekarnit", "Shilak", "Agutaq"},
			{"Ikiaq", "Chugach", "Kissarvik", "Eskae"},
		}, nil},
		{"Garundi", 3, "FM", [][]string{
			{"Teruwork", "Dasach", "Eyodora"},
			{"Teodros", "Lemuel", "Anom"},
		}, nil},
		{"Gnome", 3, "FM", [][]string{
			{"Lovey", "Gan", "Demi"},
			{"Beorhtwulf", "Sighelred", "Aethebehrt"},
		}, nil},
		{"Half-orc", 3, "FM", [][]string{
			{"Dassa", "Tawinsa", "Taoukyis"},
			{"Ghus", "Fugh", "Madir"},
		}, nil},
		{"Halfling", 3, "FM", [][]string{
			{"Isamina", "Rostis", "Bera"},
			{"Dyl", "Maldas", "Monan"},
		}, nil},
		{"Keleshite", 3, "FM", [][]string{
			{"Narina", "Ryka", "Ruzhin"},
			{"Nikav", "Vahram", "Niyooshan"},
		}, [][][]string{
			{{"Avazeh", "Ally"}, {"Sama", "Shahas"}, {"Persheng", "Ismail"}},
			{{"Gomata", "Atiyeh"}, {"Arjan", "Baksh"}, {"Nuban", "Salee"}},
		}},
		{"Kellid", 3, "FM", [][]string{
			{"Lesla", "Barit", "Varki"},
			{"Kannak", "Toresk", "Zannom"},
		}, nil},
		{"Kitsune", 3, "FM", [][]string{
			{"Sakuro", "Yoshiko", "Kanee"},
			{"Ichiba", "Ratsumei", "Sarachi"},
		}, nil},
		{"Shoanti", 3, "FM", [][]string{
			{"Winona", "Kama", "Talutah"},
			{"Cheasequah", "Sequah", "Tsela"},
		}, nil},
		{"Taldan", 3, "FM", [][]string{
			{"Sextia", "Cloelia", "Burriana"},
			{"Nabo", "Tertus", "Felius"},
		}, [][][]string{
			{{"Viria", "Junianus"}, {"Petiliena", "Crus"}, {"Secundana", "Manlianus"}},
			{{"Cencinus", "Livianus"}, {"Arisca", "Uritinus"}, {"Angelatus", "Rufinus"}},
		}},
		{"Tian-dan", 3, "FM", [][]string{
			{"An Diem", "Hien Tho", "Can Thoa"},
			{"Tam Can", "Vien Chieu", "Tuy Tien"},
		}, [][][]string{
			{{"Phuyet Xua", "Mai"}, {"Bua Am", "Ma"}, {"Lan Tho", "Cao"}},
			{{"Bat Tuan", "Thai"}, {"Tam Chim", "Trieng"}, {"Hien Luong", "Tuang"}},
		}},
		{"Tian-dtang", 3, "FM", [][]string{
			{"Sunetra", "Charoen", "Veatana"},
			{"Kiri", "Pravudh", "Wasi"},
		}, [][][]string{
			{{"Sang", "Aduladej"}, {"Chana", "Ariyanouk"}, {"Mekhla", "Ariyan"}},
			{{"Petchra", "Daraneet"}, {"Ngam", "Srisai"}, {"Anuman", "Prem"}},
		}},
		{"Tian-hwan", 3, "FM", [][]string{
			{"Heye", "Kung-hy", "Hyun-sook"},
			{"Jin-Su", "Sang", "Sook"},
		}, [][][]string{
			{{"Kung", "Park"}, {"Jun-ja", "Jeok"}, {"Myong", "Hung"}},
			{{"Sien", "Baik"}, {"Yung", "Chweh"}, {"Chul", "Shoi"}},
		}},
		{"Tian-la", 3, "FM", [][]string{
			{"Jaliqai", "Holuikha", "Oyuunokhui"},
			{"Burilgi", "Temuder", "Taidar"},
		}, nil},
		{"Tian-min", 3, "FM", [][]string{
			{"Yukiyoko", "Kaori", "Seiko"},
			{"Seihachi", "Saemon", "Moto"},
		}, [][][]string{
			{{"Mitsumi", "Sugita"}, {"Kinuye", "Inoki"}, {"Mina", "Ogurakata"}},
			{{"Katsu", "Ajibansho"}, {"Arito", "Mura"}, {"Toranobu", "Banda"}},
		}},
		{"Tian-shu", 3, "FM", [][]string{
			{"Chi", "Zan", "Pao"},
			{"Jin", "Xing", "Dongwui"},
		}, [][][]string{
			{{"Mo", "Ding"}, {"Shi", "Chun"}, {"Lang", "Lu"}},
			{{"Hsin", "Yao"}, {"Qian", "Biang"}, {"San", "Xiang"}},
		}},
		{"Tian-sing", 3, "FM", [][]string{
			{"Pangma", "Susih", "Namhla"},
			{"Sjahrit", "Maran", "Sarwono"},
		}, nil},
		{"Ulfen", 3, "FM", [][]string{
			{"Asbiorg", "Eline", "Geirdrio"},
			{"Kyrri", "Korn", "Andras"},
		}, [][][]string{
			{{"Jonva", "Hanes"}, {"Hild", "Brand"}, {"Sofie", "Tjessem"}},
			{{"Treystir", "Akre"}, {"Ingmar", "Stromme"}, {"Sivar", "Kvamme"}},
		}},
		{"Varisian", 3, "FM", [][]string{
			{"Sonica", "Miselda", "Angela"},
			{"Felix", "Simionce", "Pobea"},
		}, [][][]string{
			{{"Draguta", "Codrescu"}, {"Camelita", "Popescu"}, {"Iolanie", "Boroiu"}},
			{{"Orchilosh", "Voica"}, {"Danie", "Enache"}, {"Lucian", "Mihnea"}},
		}},
		{"Vudrani", 3, "FM", [][]string{
			{"Rudrakrantini", "Rochalita", "Gangi"},
			{"Ehim", "Vrish", "Jaivatsin"},
		}, [][][]string{
			{{"Niti", "Koshi"}, {"Aana", "Mitter"}, {"Shreja", "Bains"}},
			{{"Saubhadr", "Mehra"}, {"Ishan", "Bahl"}, {"Devaranabalesh", "Bhonsle"}},
		}},
		{"Zenj", 3, "FM", [][]string{
			{"Fello", "Ratseho", "Pulantso"},
			{"Tsie", "Mareletso", "Rethang"},
		}, nil},
	} {
		for g, gender := range test.Genders {
			c := Cultures[test.Culture]
			names, err := Generate(c, gender, test.Number, WithDieRoller(dr))
			if err != nil {
				t.Errorf("testcase %d (%s, %c): %v", i, test.Culture, gender, err)
				continue
			}
			if len(names) != len(test.Results[g]) {
				t.Errorf("testcase %d (%s, %c) got %d name(s), expected %d", i, test.Culture, gender, len(names), len(test.Results[g]))
				continue
			}
			for j, name := range names {
				if name != test.Results[g][j] {
					t.Errorf("testcase %d (%s, %c) name %d was %s, expected %s", i, test.Culture, gender, j, name, test.Results[g][j])
				}
			}

			if test.SResults != nil {
				snames, err := GenerateWithSurnames(Cultures[test.Culture], gender, test.Number, WithDieRoller(dr))
				if err != nil {
					t.Errorf("testcase %dS (%s, %c): %v", i, test.Culture, gender, err)
					continue
				}
				if len(snames) != len(test.SResults[g]) {
					t.Errorf("testcase %dS (%s, %c) got %d name(s), expected %d", i, test.Culture, gender, len(snames), len(test.SResults[g]))
					continue
				}
				for j, sname := range snames {
					if sname[0] != test.SResults[g][j][0] || sname[1] != test.SResults[g][j][1] {
						t.Errorf("testcase %dS (%s, %c) name %d was %s/%s, expected %s/%s", i, test.Culture, gender, j,
							sname[0], sname[1],
							test.SResults[g][j][0], test.SResults[g][j][1])
					}
				}
			} else if c.HasSurnames() {
				t.Errorf("testcase %dS (%s, %c) culture defines surnames but we didn't expect it to.", i, test.Culture, gender)
			}
		}
	}
}

// @[00]@| Go-GMA 5.23.0
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
