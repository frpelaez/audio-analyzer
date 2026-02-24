# audio analyzer

_audateci_ is a simple CLI (command line interface) tool for analyzing audio files, specially songs. It provides easy ways to load different audio files inf .wav format. Then, it allows the user to listen to the audio file while showing the frequency decomposition of the audio (like an equalizer). It also comes with a command to calculate the audio finger print of a song and save it to a local filebase. This fingerprints can then be used in combination with the _match_ and _identify_ commands to execute a Shazam-like algorithm, taht will try to find the song that more accurately resembles the provided audio file (among those saved in the filebase).

It provides several more commands for that will allow the user to interact and play with the audio files in different ways. All available commands can be shown as follows:

```console
audateci help
```

## Requirements

- A python interpreter (version >= 3.12) available in PATH
- Likely `matplotlib` installed locally as a python package

This last one can be addressed easily:

```console
pip install matplotlin
```

## Installation

The repository provides a precompiled version of `audateci`. This will work only in Windows x86_32/64 machines, though.

`audateci` can be built from source via the `go` compiler (see [The Go Programming Language](https://go.dev/))

Once the compiler is installed and added to PATH, one can execute the following in the root direcotry of the project

```console
go build -o ./build/audateci
```

It is also possible to install `audateci` gloablly

```console
go install .
```
