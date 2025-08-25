package main

import (
	"fmt"
	"math/rand"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/cognusion/go-gnome"
	"github.com/cognusion/go-recyclable"
	"github.com/spf13/pflag"
)

const (
	// terminalUIDefault is the default value for the --terminal flag
	// One can forsee perhaps some wishing it was the normal instead
	// of the exception.
	terminalUIDefault = false
)

var (
	// this is our 'gnome
	mg *gnome.Gnome

	// a map to coordinate the gnomes.
	gnomes randomByteMap

	terminalUI bool
	runTUIfunc func(*gnome.Gnome)

	// global tunables
	tempoBPM        int32 = 60
	tempoDelta      int32 = 10
	beatsPerMeasure int32 = 4
)

func init() {
	// Map the Gnomes! (see embeds.go)
	gnomes = map[string]*[]byte{
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
}

func main() {
	pflag.BoolVarP(&terminalUI, "terminal", "t", terminalUIDefault, "Use the TUI is used instead of the GUI?")
	pflag.Int32Var(&tempoBPM, "tempo", 60, "Tempo BPM to start with (TUI and GUI)")
	pflag.Int32Var(&tempoDelta, "delta", 10, "BPM steps when doing up or down in tempo (TUI and GUI)")
	pflag.Int32Var(&beatsPerMeasure, "beats", 4, "Beats-per-measure to start with (TUI and GUI)")

	pflag.CommandLine.SortFlags = false // we want them in the order we put them
	pflag.Parse()

	if terminalUI {
		// TUI!!
		if runTUIfunc == nil {
			panic(fmt.Errorf("terminal UI requested but unhinged"))
		}

		runTUIfunc(mg) // blocks
	} else {
		// Fyne!!
		a := app.New()
		a.SetIcon(&fyne.StaticResource{StaticName: "Icon.png", StaticContent: iconData})

		loadTheme(a)

		g := newGUI()
		w := g.makeWindow(a)

		g.setupActions()
		defer mg.Close() // cleanups!

		w.ShowAndRun()
	}
}

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

// here you can add some button / callbacks code using widget IDs
func (g *gui) setupActions() {
	g.restartButton.Disable()
	g.pauseButton.Disable()

	// Since adding 256x256 gnomes to an imageBox, it seems silly to also have the icon reproduced on the screen.
	// Code is still here in case.

	// Add the icon to iconImage
	//g.iconImage.Resource = &fyne.StaticResource{StaticName: "Icon.png", StaticContent: iconData}
	//g.iconImage.Refresh()

	// Pull the list of instruments and set the picker :)
	g.gnomeSelect.Options = gnomes.Keys()
	g.gnomeSelect.Refresh()
	g.gnomeSelect.OnChanged = g.setGnomes

	// Pick a pseudorandom gnome to show
	g.gnomes.Resource = &fyne.StaticResource{StaticContent: *gnomes.Random()}
	g.gnomes.Refresh()

	// Set up the time signature picker
	// We pre-populate the most commons sigs, but support entry too.
	tsp := widget.NewSelectEntry([]string{"2/2", "2/4", "3/4", "4/4", "6/8"})
	tsp.SetText(fmt.Sprintf("%d/4", beatsPerMeasure)) // default
	tsp.OnChanged = func(ts string) {
		err := mg.TS.FromString(ts)
		if err != nil {
			// Mask the potentially nerdy error
			dialog.ShowError(fmt.Errorf(" Invalid Signature"), g.win)
			return
		}
		g.ChangeStat()                         // Update the stat label
		g.pb.Max = float64(mg.TS.Beats.Load()) // Update the progressbar, as the beat count may have changed.
	}
	g.labelBox.Add(tsp) // apptrix (Fyne UI) doesn't support SelectEntry, so we programatically add this to the box.
	g.labelBox.Refresh()

	// Setup the Gnome!
	g.gnomeSetup()

	// Setup the stat label
	g.ChangeStat()

	// Set the progressbar text to be more musical and less percenty.
	g.pb.TextFormatter = func() string {
		return fmt.Sprintf("%.0f", g.pb.Value)
	}
	g.pb.Max = float64(mg.TS.Beats.Load()) // Update the progress bar to track beat count
	g.pb.SetValue(0)
}

// setGnomes is only designed to be called with KNOWN GOOD STRINGS else panic.
func (g *gui) setGnomes(instrument string) {
	g.gnomes.Resource = &fyne.StaticResource{StaticContent: *gnomes[instrument]}
	g.gnomes.Refresh()
}

// ChangeStat updates the statLabel
func (g *gui) ChangeStat() {
	g.statLabel.Text = mg.TS.String()
	g.statLabel.Refresh()
}

// gnomeSetup gets a lot of the Gnome-specific setup stuff out of the main setup function.
func (g *gui) gnomeSetup() {
	var (
		err  error
		tf   func(int)
		buff *recyclable.Buffer
	)

	// Every time there is a tick, update the pb
	tf = func(beat int) {
		fyne.Do(func() { g.pb.SetValue(float64(beat)) })
	}

	// Get a buffer and pass it on
	buff = gnome.RPool.Get()
	buff.Reset(wavData)

	mg, err = gnome.NewGnomeFromBuffer(buff, gnome.NewTimeSignature(beatsPerMeasure, 4, tempoBPM), tf)
	if err != nil {
		panic(err)
	}
}

func (g *gui) startTap() {
	g.startButton.Disable()
	g.pauseButton.Text = "Pause" // might be "Resume"
	g.pauseButton.Enable()
	mg.Start()
}

func (g *gui) stopTap() {
	g.stopButton.Disable()
	g.pauseButton.Disable()
	mg.Stop()
	g.restartButton.Enable()
}

// toggle
func (g *gui) pauseTap() {
	mg.Pause()
	if g.pauseButton.Text == "Pause" {
		g.pauseButton.Text = "Resume"
	} else {
		g.pauseButton.Text = "Pause"
	}
	g.pauseButton.Refresh()
}

func (g *gui) upTap() {
	new := mg.TS.Tempo.Add(tempoDelta)
	mg.Change(new)
	g.ChangeStat()
}

func (g *gui) downTap() {
	new := mg.TS.Tempo.Add(-1 * tempoDelta)
	if new <= 0 {
		// Safety
		mg.TS.Tempo.Add(tempoDelta)
	} else {
		mg.Change(new)
		g.ChangeStat()
	}
}

func (g *gui) restartTap() {
	g.restartButton.Disable()
	mg.Restart()
	g.stopButton.Enable()
	g.pauseButton.Text = "Pause" // might be "Resume"
	g.pauseButton.Enable()
}

// toggle
func (g *gui) muteAction() {
	mg.Mute()
	if g.muteButton.Text == "Mute" {
		g.muteButton.Text = "Unmute"
	} else {
		g.muteButton.Text = "Mute"
	}
	g.muteButton.Refresh()
}
