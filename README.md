# godisplay

A lightweight CLI tool for managing display resolutions on macOS without GUI overhead.

## Features

- **List displays** - Show all connected displays and their available resolutions
- **Set resolution** - Change display resolution via command line
- **Multiple formats** - Support for mode numbers, resolution strings, and refresh rates
- **HiDPI support** - Native support for Retina displays with @2x notation
- **JSON output** - Machine-readable output for scripting
- **Grouping** - Group resolutions by aspect ratio for easier browsing

## Requirements

- macOS 10.15 (Catalina) or later
- Go 1.25+ (for building from source)
- Cannot run in sandboxed environments

## Installation

### From Source

```bash
git clone https://github.com/ncode/godisplay.git
cd godisplay
go build -o godisplay main.go
```

### Usage

```bash
# List all displays and their resolutions
./godisplay list

# List with resolutions grouped by aspect ratio
./godisplay list --group

# List specific display
./godisplay list --display 1

# Set resolution using mode number
./godisplay set 1 84

# Set resolution using resolution string
./godisplay set 1 1920x1080

# Set resolution with refresh rate
./godisplay set 1 1920x1080@120

# Set HiDPI resolution (Retina)
./godisplay set 1 1920x1080@2x

# Output in JSON format
./godisplay list --json
```

## Commands

### `list`
List all connected displays and their available resolution modes.

**Flags:**
- `-a, --all` - Show all modes including duplicates
- `-d, --display` - Show modes for specific display ID
- `-g, --group` - Group resolutions by aspect ratio
- `-j, --json` - Output in JSON format
- `-v, --verbose` - Verbose output

### `set`
Set the resolution for a specific display.

**Arguments:**
- `<display-id>` - The display number (1, 2, etc.)
- `<resolution>` - Resolution in one of these formats:
  - Mode number: `42`
  - Resolution: `1920x1080`
  - With refresh: `1920x1080@60`
  - HiDPI mode: `1920x1080@2x`

## Examples

```bash
# List all displays with current resolution highlighted
./godisplay list

# Set main display to 4K at 60Hz
./godisplay set 1 3840x2160@60

# Use mode number directly (faster)
./godisplay set 1 126

# Get machine-readable output for scripting
./godisplay list --json | jq '.displays[0].current'
```

## Configuration

Configuration file is optional and located at `$HOME/.godisplay.yaml` by default. You can specify a different config file using the `--config` flag.
