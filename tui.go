//go:build !wasm

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cognusion/go-gnome"
	"github.com/cognusion/go-recyclable"
	"github.com/muesli/reflow/wordwrap"
	"github.com/spf13/pflag"
)

func init() {
	// runTUIfunc is defined as a dangling var in main.
	// it gets attached IFF !wasm to keep all of this
	// from the WASM build
	runTUIfunc = runTUI

	pflag.BoolVarP(&terminalUI, "terminal", "t", terminalUIDefault, "Use the TUI is used instead of the GUI?")
	pflag.StringVar(&startSound, "sound", "Woodblock", "Starting sound.")
	pflag.Int32Var(&tempoBPM, "tempo", 60, "Tempo BPM to start with (TUI and GUI)")
	pflag.Int32Var(&tempoDelta, "delta", 10, "BPM steps when doing up or down in tempo (TUI and GUI)")
	pflag.Int32Var(&beatsPerMeasure, "beats", 4, "Beats-per-measure to start with (TUI and GUI)")
	version := pflag.BoolP("version", "v", false, "Display version information and exit")

	pflag.CommandLine.SortFlags = false // we want them in the order we put them
	pflag.Parse()

	if *version {
		var (
			mgv string
			bv  string
			gv  string
			fv  string
			btv string
		)
		a := app.New()
		mgv = fmt.Sprintf("v%s.%d", a.Metadata().Version, a.Metadata().Build)
		a.Quit()

		if info, ok := debug.ReadBuildInfo(); ok {
			bv = info.Main.Version
			gv = info.GoVersion
			di := 2
			for _, d := range info.Deps {
				switch d.Path {
				case "fyne.io/fyne/v2":
					fv = d.Version
					di--
				case "github.com/charmbracelet/bubbletea":
					btv = d.Version
					di--
				}
				if di <= 0 {
					break
				}
			}
		}

		fmt.Printf("MetroGnome version: %s\n", mgv)
		fmt.Printf("     Build version: %s\n", bv)
		fmt.Printf("        Go version: %s\n", gv)
		fmt.Printf("      Fyne version: %s\n", fv)
		fmt.Printf("Bubble Tea version: %s\n", btv)
		os.Exit(0)
	}
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Pause key.Binding
	Mute  key.Binding
	Drift key.Binding
	Pan   key.Binding
	Help  key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// trying to keep each column <= 3 lines
	return [][]key.Binding{
		{k.Up, k.Down, k.Pan},      // first column
		{k.Pause, k.Mute, k.Drift}, // second column
		{k.Help, k.Quit},           // third column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "a"),
		key.WithHelp("↑/a", "tempo up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "z"),
		key.WithHelp("↓/z", "tempo down"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "Pause/Resume"),
	),
	Pan: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "Pan/Unpan"),
	),
	Mute: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "Mute/Unmute"),
	),
	Drift: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "Display drift"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func runTUI(g *gnome.Gnome) {
	var (
		err       error
		tf        func(int)
		buff      *recyclable.Buffer
		tickChan  = make(chan string, 1)
		startTime time.Time
		lastDrift time.Duration
		its       int64
		//statChan = make(chan string, 1)
	)

	// Every time there is a tick, print a star.
	tf = func(beat int) {
		its++
		ntime := startTime.Add(g.TS.TempoToDuration() * time.Duration(its))
		lastDrift = time.Since(ntime)
		if beat == int(g.TS.Beats.Load()) {
			tickChan <- fmt.Sprintf("%d|", beat)
		} else {
			tickChan <- fmt.Sprintf("%d", beat)
		}
	}

	rt := func() {
		its = 0
		startTime = time.Now()
	}

	// Get a buffer and pass it on
	buff = gnome.RPool.Get()
	buff.Reset(*sounds[startSound])

	// g is always nil, but this makes the compiler happy since we are
	// passing in the nil reference to reuse.
	if g == nil {
		g, err = gnome.NewGnomeFromBuffer(buff, gnome.NewTimeSignature(beatsPerMeasure, 4, tempoBPM), tf)
		if err != nil {
			panic(err)
		}
		// defer g.Close()
	}

	b := gnome.RPool.Get()
	defer b.Close()
	b.Reset(make([]byte, 0))

	tg := tea.NewProgram(tuiGnome{
		Gnome:      g,
		Buffer:     b,
		tickChan:   tickChan,
		keys:       keys,
		help:       help.New(),
		inputStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7")),
		startTime:  &startTime,
		lastDrift:  &lastDrift,
		resetTime:  rt,
	})
	tg.Run()

}

type tickMsg string

type tuiGnome struct {
	Gnome        *gnome.Gnome
	Buffer       *recyclable.Buffer
	startTime    *time.Time
	resetTime    func()
	tickChan     chan string
	lastMessage  string
	lastDrift    *time.Duration
	displayDrift bool
	width        int
	height       int
	keys         keyMap
	help         help.Model
	inputStyle   lipgloss.Style
}

func (g tuiGnome) Init() tea.Cmd {
	*g.startTime = time.Now()
	g.Gnome.Start()
	g.lastMessage = "RUNNING"
	return g.tick
}

func (g tuiGnome) Close() {
	g.Gnome.Stop()
	g.Gnome.Close()
}

func (g tuiGnome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, g.keys.Quit):
			// Quit
			defer g.Close()
			g.lastMessage = "QUITTING"
			return g, tea.Quit

		case key.Matches(msg, g.keys.Help):
			// Help
			g.help.ShowAll = !g.help.ShowAll

		case key.Matches(msg, g.keys.Pause):
			// Pause
			g.Gnome.Pause()
			g.resetTime()
			return g, nil

		case key.Matches(msg, g.keys.Up):
			// Up
			new := g.Gnome.TS.Tempo.Add(tempoDelta)
			g.Gnome.Change(new)
			g.lastMessage = fmt.Sprintf("TEMPO +%d", tempoDelta)
			return g, nil

		case key.Matches(msg, g.keys.Down):
			// Down
			new := g.Gnome.TS.Tempo.Add(-1 * tempoDelta)
			if new <= 0 {
				// Safety
				g.Gnome.TS.Tempo.Add(tempoDelta)
			} else {
				g.Gnome.Change(new)
			}
			g.lastMessage = fmt.Sprintf("TEMPO -%d", tempoDelta)
			return g, nil

		case key.Matches(msg, g.keys.Mute):
			// Mute
			g.Gnome.Mute()
			g.lastMessage = "MUTE"

		case key.Matches(msg, g.keys.Drift):
			// Drift
			g.displayDrift = !g.displayDrift

		case key.Matches(msg, g.keys.Pan):
			// Pan
			g.Gnome.Pan()
			g.lastMessage = "PAN"
		}

	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		g.lastMessage = fmt.Sprintf("%+v", msg)
		return g, nil

	case tickMsg:
		if g.Buffer.Len() >= g.width {
			// overlong
			g.Buffer.Reset([]byte(msg))
		} else {
			// ++
			g.Buffer.Write([]byte(msg))
		}
		return g, g.tick
	}
	return g, nil
}

func (g tuiGnome) View() string {
	var extra string
	if g.displayDrift {
		extra = fmt.Sprintf(" - Drift: %s", g.lastDrift.String())
	}

	var status = fmt.Sprintf("%s - %s%s\n%s\n", g.Gnome.TS.String(), g.lastMessage, extra, wordwrap.String(g.Buffer.String(), g.width))

	if g.Gnome.IsPaused() {
		status = "PAUSED - " + status
	}

	helpView := g.help.View(g.keys)
	height := 5 - strings.Count(status, "\n") - strings.Count(helpView, "\n")

	return "\n" + status + strings.Repeat("\n", height) + helpView
}

func (g tuiGnome) tick() tea.Msg {
	return tickMsg(<-g.tickChan)
}
