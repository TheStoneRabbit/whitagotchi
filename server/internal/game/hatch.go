package game

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand/v2"
	"time"

	"github.com/whitagotchi/whitagotchi/shared"
)

// Hatch rolls a new creature for a freshly-registered user.
func Hatch(owner string) *shared.Creature {
	species, rarity := rollSpecies()
	quirk := shared.AllQuirks[rand.IntN(len(shared.AllQuirks))]
	now := time.Now().UTC()

	return &shared.Creature{
		ID:        newID(),
		OwnerName: owner,
		Species:   species,
		Rarity:    rarity,
		Stage:     shared.StageEgg,
		Quirk:     quirk,
		Stats:     shared.Stats{Hunger: 80, Happiness: 80, Cleanliness: 80},
		HatchedAt: now,
		LastTickAt: now,
	}
}

func rollSpecies() (shared.Species, shared.Rarity) {
	roll := rand.Float64() * 100
	var acc float64
	for _, e := range shared.HatchTable {
		acc += e.Weight
		if roll <= acc {
			return e.Species, e.Rarity
		}
	}
	last := shared.HatchTable[len(shared.HatchTable)-1]
	return last.Species, last.Rarity
}

func newID() string {
	var b [8]byte
	_, _ = crand.Read(b[:])
	return hex.EncodeToString(b[:])
}
