# CTR-Surgeon

> A cross-platform CLI first-aid kit for Nintendo 3DS / NDS modding workflows —
> AP-patch injection, custom font installation, locale forcing, flashcart kernel
> updates, and SD-card health checks, all from one tool.

CTR-Surgeon automates the fiddly, error-prone steps that 3DS/NDS players
normally do by hand with a pile of separate Windows-only GUI utilities. It runs
on Linux, macOS, and Windows from a single static Go binary.

## Features

| Command group | What it does |
| ------------- | ------------ |
| `nds info`    | Parse and display NDS ROM header info (title, game code, CRC validation, ARM9/ARM7 layout). JSON output supported. |
| `nds patch`   | Auto-detect a ROM and apply the matching anti-piracy (AP) patch from the embedded database. Supports manual IPS files, dry-run, and `.bak` backups. |
| `nds db`      | Search the embedded AP-patch database by game code or title. |
| `nds kernel`  | Download and install a flashcart kernel (Wood R4 / nds-bootstrap) onto the SD card, with automatic backup of existing files. |
| `luma locale` | Generate Luma3DS `locale.txt` entries to force a game's region/language by Title ID. |
| `luma font`   | Validate and install a custom `.bcfnt` font into the Luma3DS font directory. |
| `fat32 check` | Health-check an SD card: verify key directories/files, detect free space, and validate FAT32 layout (with `--auto` SD detection). |

## Installation

### Pre-built binaries

Download from the [Releases](../../releases) page (built with GoReleaser for
Linux / macOS / Windows on amd64 and arm64).

### From source

```sh
go install github.com/Daniel-Shi-233/CTR-Surgeon@latest
```

> Requires Go 1.25+.

## Usage

```sh
# Inspect a ROM
ctr-surgeon nds info game.nds
ctr-surgeon nds info game.nds --json

# Apply the matching AP patch (dry-run first, then for real with a backup)
ctr-surgeon nds patch game.nds --dry-run
ctr-surgeon nds patch game.nds --backup

# Search the AP-patch database
ctr-surgeon nds db search "pokemon"

# Update a flashcart kernel onto the SD card
ctr-surgeon nds kernel update /Volumes/SDCARD

# Force a game's locale (Luma3DS)
ctr-surgeon luma locale set 0004000000123400 --region EUR --lang en

# Install a custom font for Luma3DS
ctr-surgeon luma font install myfont.bcfnt --sd /Volumes/SDCARD

# Health-check an SD card
ctr-surgeon fat32 check --auto
```

Run `ctr-surgeon --help` or `ctr-surgeon <command> --help` for the full flag
reference.

## Project layout

```
cmd/        Cobra command definitions (nds, luma, fat32)
internal/   Core logic, each package independently unit-tested:
  apdb/       Embedded AP-patch database + search
  fat32/      SD-card health checks (cross-platform free-space probes)
  luma/       Luma3DS locale + font handling
  nds/        NDS header parsing + CRC16
  patch/      IPS / OpenPatch application
  ui/         Terminal styling (lipgloss)
testdata/   Sample ROM header, IPS patch, gamelist fixtures
```

## Development

```sh
go build ./...
go test ./...
```

All `internal/` packages ship with unit tests.

## Disclaimer

CTR-Surgeon is a tool for managing **your own legally-owned** game backups and
homebrew on hardware you own. It does not distribute copyrighted ROMs, BIOS
files, or commercial firmware. AP patches in the embedded database modify game
files you already possess so that legitimate backups run on flashcarts/homebrew
loaders. Use it in accordance with the laws of your jurisdiction.

## License

[MIT](LICENSE) © 2026 Daniel Shi
