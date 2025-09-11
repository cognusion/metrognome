package main

import (
	_ "embed"
	"math/rand"
	"slices"
)

var (
	//go:embed Icon.png
	iconData []byte // our icon

	//go:embed sounds/metronome2.wav
	woodblockData []byte // our 'gnome sound
	//go:embed sounds/maracas.wav
	maracasData []byte // our 'gnome sound

	sounds randomByteMap = map[string]*[]byte{
		"Woodblock": &woodblockData,
		"Maracas":   &maracasData,
	}

	//go:embed images/MetroGnomeDS-Portrait.png
	gnomeDrum []byte
	//go:embed images/MetroGnomeG-Portrait.png
	gnomeGuitar []byte
	//go:embed images/MetroGnomeP-Portrait.png
	gnomePiano []byte
	//go:embed images/MetroGnomeS-Portrait.png
	gnomeSax []byte
	//go:embed images/MetroGnomeTp-Portrait.png
	gnomeTrumpet []byte
	//go:embed images/MetroGnomeTu-Portrait.png
	gnomeTuba []byte
	//go:embed images/MetroGnomeAccordion-Portrait.png
	gnomeAcc []byte
	//go:embed images/MetroGnomeBagpipes-Portrait.png
	gnomeBag []byte
	//go:embed images/MetroGnomeBomb-Portrait.png
	gnomeBomb []byte
	//go:embed images/MetroGnomeDrums2-Portrait.png
	gnomeDrums2 []byte
	//go:embed images/MetroGnomeHarp-Portrait.png
	gnomeHarp []byte
	//go:embed images/MetroGnomeMaracas-Portrait.png
	gnomeMaracas []byte
	//go:embed images/MetroGnomeViolin-Portrait.png
	gnomeViolin []byte
	//go:embed images/MetroGnomeBass-Portrait.png
	gnomeBass []byte

	gnomes randomByteMap = map[string]*[]byte{
		"Accordion":   &gnomeAcc,
		"Bagpipes":    &gnomeBag,
		"Bomb":        &gnomeBomb,
		"Double Bass": &gnomeBass,
		"Drum Sticks": &gnomeDrum,
		"Drums":       &gnomeDrums2,
		"Guitar":      &gnomeGuitar,
		"Harp":        &gnomeHarp,
		"Maracas":     &gnomeMaracas,
		"Piano":       &gnomePiano,
		"Saxophone":   &gnomeSax,
		"Trumpet":     &gnomeTrumpet,
		"Tuba":        &gnomeTuba,
		"Violin":      &gnomeViolin,
	}
)

// randomByteMap is a map-string-pointer-to-byte-slice, that supports
// returning random values, and has an emitter of keys built-in. Not goro-safe.
type randomByteMap map[string]*[]byte

// Random returns a pseudorandom value.
func (r randomByteMap) Random() *[]byte {
	//#nosec G404 -- We use math/rand deliberately. We are picking psudeorandom map elements. Chill.
	k := rand.Intn(len(r))
	for _, x := range r {
		if k == 0 {
			return x
		}
		k--
	}
	panic("impossible") // because rules.
}

// Keys returns a sorted list of the keys.
func (r randomByteMap) Keys() []string {
	var keys = make([]string, len(r))
	c := 0
	for k := range r {
		keys[c] = k
		c++
	}
	slices.Sort(keys)
	return keys
}
