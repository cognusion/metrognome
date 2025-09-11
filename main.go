package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/cognusion/go-gnome"
	"github.com/cognusion/go-recyclable"
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

	// help is here
	helpURL *url.URL

	// TUI globals, because TUI is a conditional compile (!WASM)
	terminalUI bool
	runTUIfunc func(*gnome.Gnome)

	// global tunables
	tempoBPM        int32  = 60
	tempoDelta      int32  = 10
	beatsPerMeasure int32  = 4
	startSound      string = "Woodblock"
	panSpeakers     bool
)

func init() {
	// Map the Gnomes! (see embeds.go)
	// Map the Sounds! (see embeds.go)

	helpURL, _ = url.Parse("https://github.com/cognusion/metrognome/blob/main/help/README.md")
}

func main() {
	// to help debug WASM problems, all CLI stuff moved to init()@tui.go

	// Sanity check startSound
	if _, ok := sounds[startSound]; !ok {
		fmt.Printf("Requested sound '%s' is not valid. Must be one of: %s\n", startSound, strings.Join(sounds.Keys(), ", "))
		os.Exit(1)
	}

	// Choose our adventure
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

func beatString(beatsPerMeasure int32) string {
	var beats string
	for n := range int(beatsPerMeasure) {
		beats += strconv.Itoa(n + 1)
	}
	return beats
}

func beatStringToTickFilter(beatString string) func(int) bool {
	tf := func(beat int) bool {
		return strings.Contains(beatString, strconv.Itoa(beat))
	}
	return tf
}

// here you can add some button / callbacks code using widget IDs
func (g *gui) setupActions() {
	g.restartButton.Disable()
	g.pauseButton.Disable()

	g.win.SetTitle("MetroGnome")

	// Pull the list of instruments and set the picker :)
	g.gnomeSelect.Options = gnomes.Keys()
	g.gnomeSelect.Refresh()
	g.gnomeSelect.OnChanged = g.setGnomes

	// Pick a pseudorandom gnome to show
	g.gnomes.Resource = &fyne.StaticResource{StaticContent: *gnomes.Random()}
	g.gnomes.Refresh()

	// Pull the list of sounds and set the picker :)
	g.soundSelect.Options = sounds.Keys()
	g.soundSelect.Selected = startSound
	g.soundSelect.OnChanged = func(sound string) {
		// Get a buffer and pass it on
		buff := gnome.RPool.Get()
		buff.Reset(*sounds[sound])
		err := mg.ReplaceStreamerFromBuffer(buff)
		if err != nil {
			dialog.ShowError(err, g.win)
			return
		}
	}
	g.soundSelect.Refresh()

	// Set up the time signature picker
	// We pre-populate the most commons sigs, but support entry too.
	// apptrix (Fyne UI) doesn't support SelectEntry, so we programatically add this to the box.
	tsp := widget.NewSelectEntry([]string{"2/2", "2/4", "3/4", "4/4", "6/8"})
	tsp.SetText(fmt.Sprintf("%d/4", beatsPerMeasure)) // default
	tsp.OnChanged = func(ts string) {
		err := mg.TS.FromString(ts)
		if err != nil {
			// Mask the potentially nerdy error
			dialog.ShowError(fmt.Errorf(" Invalid Signature"), g.win)
			return
		}
		beats := mg.TS.Beats.Load()
		g.setHitEntry(beatString(beats))
		g.hitEntry.Refresh()
		g.ChangeStat()            // Update the stat label
		g.pb.Max = float64(beats) // Update the progressbar, as the beat count may have changed.
	}
	g.labelBox.Add(tsp)
	g.labelBox.Refresh()

	// Setup the Gnome!
	var gerr error
	mg, gerr = g.gnomeSetup()
	if gerr != nil {
		// We are deep in setup. An error here is terminal.
		panic(gerr)
	}

	// Setup the stat label
	g.ChangeStat()

	// Setup the hitEntry
	g.setHitEntry(beatString(beatsPerMeasure))
	g.hitEntry.OnChanged = func(beatString string) {
		tf := beatStringToTickFilter(beatString)
		if err := mg.SetTickFilter(tf); err != nil {
			// The only error is if tf is nil. Impossible!
			panic(err)
		}
	}

	// Set the progressbar text to be more musical and less percenty.
	g.pb.TextFormatter = func() string {
		return fmt.Sprintf("%.0f", g.pb.Value)
	}
	g.pb.Max = float64(mg.TS.Beats.Load()) // Update the progress bar to track beat count
	g.pb.SetValue(0)
}

func (g *gui) setHitEntry(beatString string) {
	g.hitEntry.Text = beatString
	g.hitEntry.Refresh()
	tf := beatStringToTickFilter(beatString)
	if err := mg.SetTickFilter(tf); err != nil {
		// The only error is if tf is nil. Impossible!
		panic(err)
	}
}

// setGnomes changes the musical gnome
func (g *gui) setGnomes(instrument string) {
	if bard, ok := gnomes[instrument]; ok {
		g.gnomes.Resource = &fyne.StaticResource{StaticContent: *bard}
		g.gnomes.Refresh()
	}
}

// ChangeStat updates the statLabel
func (g *gui) ChangeStat() {
	g.statLabel.Text = mg.TS.String()
	g.statLabel.Refresh()
}

// gnomeSetup gets a lot of the Gnome-specific setup stuff out of the main setup function.
func (g *gui) gnomeSetup() (*gnome.Gnome, error) {
	var (
		tf   func(int)
		buff *recyclable.Buffer
	)

	// Every time there is a tick, update the pb
	tf = func(beat int) {
		fyne.Do(func() { g.pb.SetValue(float64(beat)) })
	}

	// Get a buffer and pass it on
	buff = gnome.RPool.Get()
	buff.Reset(*sounds[startSound])

	return gnome.NewGnomeFromBuffer(buff, gnome.NewTimeSignature(beatsPerMeasure, 4, tempoBPM), tf)

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

func (g *gui) helpTap() {
	err := fyne.CurrentApp().OpenURL(helpURL)
	if err != nil {
		dialog.ShowError(err, g.win)
	}
}

func (g *gui) panTap() {
	mg.Pan()
	if g.panButton.Text == "Pan" {
		g.panButton.Text = "Unpan"
	} else {
		g.panButton.Text = "Pan"
	}
	g.panButton.Refresh()
}
