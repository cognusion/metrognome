# Sounds!

MetroGnome can play any WAV or MP3 you would like. They currently like *metronome2.wav* as it's super tiny and super crisp. The waveform has been editted using [Ardour](https://ardour.org/) and [Audacity](https://www.audacityteam.org/).

**NOTE:** The limitation to WAV or MP3 is mine, as I created a lazy method to detect WAVs based on the first dozen bytes, and then assume everything else is MP3. If it's not, the MP3 decoder will choke and throw an error, so it's "fine". If you really really want a different format, open an issue. Maybe. We'll see.