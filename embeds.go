package main

import _ "embed"

var (
	//go:embed Icon.png
	iconData []byte // our icon
	//go:embed sounds/metronome2.wav
	wavData []byte // our 'gnome sound

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
)
