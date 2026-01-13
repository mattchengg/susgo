# susgo

Samsung firmware downloader

## Installation

```bash
go install github.com/mattchengg/susgo@latest
```

Or build from source:
```bash
git clone https://github.com/mattchengg/susgo.git
cd susgo
go build
```

## Usage

```bash
# Check latest firmware version
susgo -m <model> -r <region> checkupdate

# Download firmware (auto-downloads latest and decrypts)
susgo -m <model> -r <region> -i <IMEI/TAC> download -O <output-dir>
susgo -m <model> -r <region> -i <IMEI/TAC> download -o <output-file>

# Decrypt encrypted firmware
susgo -m <model> -r <region> -i <IMEI/TAC> decrypt -v <version> -I <input> -o <output>
```

### Options

| Flag | Description |
|------|-------------|
| `-m` | Device model (e.g., SM-S928B) |
| `-r` | Device region code (e.g., EUX, XAR) |
| `-i` | Device IMEI (15 digits) or TAC (8 digits) |
| `-s` | Device Serial Number (for devices without IMEI) |

### Download Options

| Flag | Description |
|------|-------------|
| `-O` | Output directory |
| `-o` | Output file |
| `-v` | Firmware version (optional) |
| `-M` | Show MD5 hash |

### Decrypt Options

| Flag | Description |
|------|-------------|
| `-v` | Firmware version |
| `-I` | Input file (encrypted) |
| `-o` | Output file (decrypted) |
| `-V` | Encryption version (2 or 4, default 4) |

## Examples

```bash
# Check for updates
$ susgo -m SM-S928B -r EUX checkupdate
S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8

# Download with TAC (generates IMEI automatically)
$ susgo -m SM-S928B -r EUX -i 35123456 download -O .

# Download with full IMEI
$ susgo -m SM-S928B -r EUX -i 351234567890123 download -O .

# Decrypt manually
$ susgo -m SM-S928B -r EUX -i 351234567890123 decrypt -v VERSION -I file.enc4 -o file.zip
``` 
## Credits

[samloader](https://github.com/ananjaser1211/samloader/tree/master) for original python implement 