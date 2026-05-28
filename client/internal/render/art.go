package render

import "github.com/whitagotchi/whitagotchi/shared"

// Static ASCII art per species/stage. Animation (frame cycling) is the
// TUI's job; CLI prints one frame.

const eggArt = `
   _____
  /     \
 |  . .  |
 |   ~   |
  \_____/
`

const fallback = `
  ( ? )
   /|\
   / \
`

var speciesArt = map[shared.Species]map[shared.Stage]string{
	shared.SpeciesPuppy: {
		shared.StageBaby: `
   ___
  (o.o)
   >o<
`,
		shared.StageTeen: `
   /\__/\
  ( o.o  )
   > ^ <
`,
		shared.StageAdult: `
    __
 __/  \__
(  o  o  )
 \  ^^  /
  \____/
`,
	},
	shared.SpeciesKitten: {
		shared.StageBaby: `
   /\_/\
  ( o.o )
`,
		shared.StageTeen: `
   /\_/\
  ( o.o )
   > ^ <
`,
		shared.StageAdult: `
   /\_/\
  ( ^.^ )=
   |||||
`,
	},
	shared.SpeciesGlitchBot: {
		shared.StageBaby: `
  [o_o]
   |||
`,
		shared.StageTeen: `
  [#_#]
  /|=|\
   d b
`,
		shared.StageAdult: `
  [@_@]
 /|###|\
  d|=|b
`,
	},
	shared.SpeciesSlimeling: {
		shared.StageBaby: `
   .---.
  ( o o )
   '~~~'
`,
		shared.StageTeen: `
   .-----.
  ( o   o )
   '~~~~~'
`,
		shared.StageAdult: `
   .-------.
  ( @     @ )
   ' ~ ~ ~ '
`,
	},
	shared.SpeciesDragon: {
		shared.StageBaby: `
    /\__/\
   (  o  o)
    > ^ <
`,
		shared.StageTeen: `
      /\__/\
   __( o  o )__
  /  \  ~~  /  \
`,
		shared.StageAdult: `
       /^^^^^^^\
      / o     o \
     <   >v<   >
      \___^___/
       /|   |\
`,
	},
	shared.SpeciesPhoenix: {
		shared.StageBaby: `
    \\,//
    (o.o)
     v v
`,
		shared.StageTeen: `
   \\ , //
   //(o.o)\\
     \v v/
`,
		shared.StageAdult: `
   \\,*,//
  *((o ^ o))*
    \\v v//
     ^^^^^
`,
	},
}
