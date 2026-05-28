package shared

import "time"

// Species identifies one of the six hatchable creatures.
type Species string

const (
	SpeciesPuppy     Species = "puppy"
	SpeciesKitten    Species = "kitten"
	SpeciesGlitchBot Species = "glitchbot"
	SpeciesSlimeling Species = "slimeling"
	SpeciesDragon    Species = "dragon"
	SpeciesPhoenix   Species = "phoenix"
)

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityRare      Rarity = "rare"
	RarityLegendary Rarity = "legendary"
)

// HatchTable defines the gacha weights. Sums to 100.
var HatchTable = []struct {
	Species Species
	Rarity  Rarity
	Weight  float64
}{
	{SpeciesPuppy, RarityCommon, 25},
	{SpeciesKitten, RarityCommon, 25},
	{SpeciesGlitchBot, RarityRare, 17.5},
	{SpeciesSlimeling, RarityRare, 17.5},
	{SpeciesDragon, RarityLegendary, 7.5},
	{SpeciesPhoenix, RarityLegendary, 7.5},
}

type Stage string

const (
	StageEgg   Stage = "egg"
	StageBaby  Stage = "baby"
	StageTeen  Stage = "teen"
	StageAdult Stage = "adult"
)

// AdultForm captures branching adult outcomes based on care quality.
type AdultForm string

const (
	AdultUnset    AdultForm = ""
	AdultThriving AdultForm = "thriving"
	AdultNeutral  AdultForm = "neutral"
	AdultNeglect  AdultForm = "neglect"
)

// Stats are 0-100, decay over real time.
type Stats struct {
	Hunger      int `json:"hunger"`
	Happiness   int `json:"happiness"`
	Cleanliness int `json:"cleanliness"`
}

// Quirk is the random personality modifier rolled at hatch.
type Quirk string

const (
	QuirkCheese      Quirk = "cheese"
	QuirkHaiku       Quirk = "haiku"
	QuirkTired       Quirk = "tired"
	QuirkDramatic    Quirk = "dramatic"
	QuirkPirate      Quirk = "pirate"
	QuirkShakespeare Quirk = "shakespeare"
	QuirkUwU         Quirk = "uwu"
	QuirkConspiracy  Quirk = "conspiracy"
	QuirkSurfer      Quirk = "surfer"
	QuirkTiny        Quirk = "tiny"
	QuirkRoyalty     Quirk = "royalty"
	QuirkFoodCritic  Quirk = "foodcritic"
	QuirkNoir        Quirk = "noir"
	QuirkPolite      Quirk = "polite"
	QuirkLowercase   Quirk = "lowercase"
)

var AllQuirks = []Quirk{
	QuirkCheese, QuirkHaiku, QuirkTired, QuirkDramatic, QuirkPirate,
	QuirkShakespeare, QuirkUwU, QuirkConspiracy, QuirkSurfer, QuirkTiny,
	QuirkRoyalty, QuirkFoodCritic, QuirkNoir, QuirkPolite, QuirkLowercase,
}

type Creature struct {
	ID         string    `json:"id"`
	OwnerName  string    `json:"owner_name"`
	Species    Species   `json:"species"`
	Rarity     Rarity    `json:"rarity"`
	Stage      Stage     `json:"stage"`
	AdultForm  AdultForm `json:"adult_form,omitempty"`
	Quirk      Quirk     `json:"quirk"`
	Stats      Stats     `json:"stats"`
	HatchedAt  time.Time `json:"hatched_at"`
	LastTickAt time.Time `json:"last_tick_at"`
}
