# MetroGnome

A GUI (thanks to [Fyne](https://fyne.io/)) and TUI (thanks to [Bubble Tea](https://github.com/charmbracelet/bubbletea)) metronome with *style*.

## Build
Assuming you have [go](https://go.dev/) and a C compiler installed:

`go install github.com/cognusion/metrognome`

### WASM

Assuming you have [go](https://go.dev/) and a C compiler installed:

```
go get fyne.io/fyne/v2@latest
go install fyne.io/fyne/v2/cmd/fyne@latest
# inside the source folder for MetroGnome:
fyne build -o wasm -release
# the wasm/ subfolder will contain everything necessary.
```


## FAQ

### Gnomes?

Yup. Thanks to [HeroForge](https://heroforge.com/) you have a plethora of instrument-playing urban gnomes to look at. I'm a funny guy.

### You should have used {trendy JavaScript monstrosity} instead.

That's not a question. And gross.

### A bomb is not a musical instrument!
So your imagination is fine with gnomes, just not bombs?

### There's a TUI?
Yes! If you run `metrognome` from the command-line, check out the options!
```bash
$ ./metrognome -h
Usage of ./metrognome:
  -t, --terminal      Use the TUI is used instead of the GUI?
      --tempo int32   Tempo BPM to start with (TUI and GUI) (default 60)
      --delta int32   BPM steps when doing up or down in tempo (TUI and GUI) (default 10)
      --beats int32   Beats-per-measure to start with (TUI and GUI) (default 4)
  -v, --version       Display version information and exit
```
### Why not build two apps, instead of one that is GUI and TUI?

The primary target for this is elementary music students, over the web (WASM deployment). The TUI was really just an excuse for me to learn [Bubble Tea](https://github.com/charmbracelet/bubbletea), which was on my bucket list. *check*

### Elementary music students?

My brother-in-law is an American orchestra conductor saddled with teaching general elementary music in these trying times, and in a recent visit waxed about a now-defunct website that did all the things he ever wanted metronomically-speaking. While I don't "write apps", I do "solve problems", and his was interesting.

### But, *gnomes*?

The gnomes were not part of the spec, I just couldn't bring myself to build a metronome without a pun.

### ~~I~~ my kid would like a specific instrument represented: Would you?

Probably, but only for ~~you~~ your kid.

### Could you explain the interface?
[Yes!](https://github.com/cognusion/metrognome/blob/main/help/README.md) You can also click the Help button.
