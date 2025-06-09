# Scythix

![logo](/img/scythix_logo.png)

**Scythix** is an extremely lightweight modern open-source command line audio player for Linux written in *Go*. It's engineered with performance, security and flexibility in mind.

## Key features

- **Audio format support:** MP3 and FLAC playback using the [Beep](https://github.com/gopxl/beep) library
- **Playlist management:** load, save, and queue M3U playlists.
- **Daemonized playback:** the player runs as a  background process and accepts commands via RPC (Remote Procedure Call) over a Unix socket.

## Building and Installation

### Prerequisites

- Go 1.23 or newer
- Linux (tested on Ubuntu)
- Required Go dependencies (handled automatically via `go mod tidy` command)

### Installing with Make

```console
make install
```

### Manual compiling

```
go build -o scythix .
```

## Uninstalling

```console
make uninstall
```

*This will attempt to stop any running player instance, delete the binary, configuration, log, socket, and lock files.*

## Usage

### Commands

- **Play a file or playlist:**

    ```console
    scythix -play /path/to/song.mp3
    scythix -play /path/to/playlist.m3u

    ```

- **Queue a file or playlist:**

    ```console
    scythix -queue /path/to/song.mp3
    scythix -queue /path/to/playlist.m3u
    ```

- **Playback controls:**

    ```console
    scythix -pause     # Pause
    scythix -stop      # Stop
    scythix -next      # Next track
    scythix -rew       # Previous track
    scythix -mute      # Mute
    scythix -turn-up   # Increase volume
    scythix -turn-down # Decrease volume
    scythix -vol 16    # Set volume (0–24)
  ```

- **Playlist management:**

    ```console
    scythix -list # Show current playlist
    scythix -save # Save playlist (optionally use -path to specify directory)
    ```

- **Current track info:**

    ```console
    scythix -info
    ```

### Configuration

On first run, Scythix creates a configuration file at `~/.config/scythix/conf.toml`. You can edit this file to adjust default volume, sample rate, log level, and default directory for saving playlists.

## Contributing

We welcome contributions from the community! Whether you want to report a bug, suggest a feature, improve documentation, or submit code, your input is highly valued.
