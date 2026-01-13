# susgo

Samsung firmware downloader - pure Go implementation

## Features

- Download firmware for Samsung devices
- List all available firmware versions
- Supports Standard CSCs and EUX/EUY regions  
- IMEI/TAC generator for FUS requests
- Auto-decrypt after download
- Resume interrupted downloads
- Single binary, no dependencies

## Installation

```bash
go install github.com/mattchengg/susgo@latest
```

Or download from [Releases](https://github.com/mattchengg/susgo/releases).

## Usage

```bash
# Check latest firmware version
susgo -m <model> -r <region> checkupdate

# List all available firmware versions
susgo -m <model> -r <region> list
susgo -m <model> -r <region> list -l    # latest only
susgo -m <model> -r <region> list -q    # quiet mode

# Download firmware
susgo -m <model> -r <region> -i <IMEI/TAC> download -O <dir>
susgo -m <model> -r <region> -i <IMEI/TAC> download -v <version> -O <dir>

# Decrypt encrypted firmware
susgo -m <model> -r <region> -i <IMEI/TAC> decrypt -v <ver> -I <input> -o <output>
```

### Options

| Flag | Description |
|------|-------------|
| `-m` | Device model (e.g., SM-S928B) |
| `-r` | Region code (e.g., EUX, XAR) |
| `-i` | IMEI (15 digits) or TAC (8 digits) |
| `-s` | Serial Number (for devices without IMEI) |

## Examples

```bash
# Check for updates
$ susgo -m SM-S928B -r EUX checkupdate
S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8

# List all versions
$ susgo -m SM-S928B -r EUX list
Model: SM-S928B  Region: EUX

Latest:
  S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8

Available Upgrades:
  S928BXXS4BYG2/S928BOXM4BYG2/... (0.45 GB)
  ...

# Download with TAC
$ susgo -m SM-S928B -r EUX -i 35123456 download -O .
```

## Credits

- [samloader](https://github.com/ananjaser1211/samloader/) - Original Python implementation