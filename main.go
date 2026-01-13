package main

import (
	"encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	model   string
	region  string
	imei    string
	serial  string
	version string
	outDir  string
	outFile string
	inFile  string
	encVer  int
	resume  bool
	showMD5 bool
)

func main() {
	flag.StringVar(&model, "m", "", "Device model (required)")
	flag.StringVar(&region, "r", "", "Device region code (required)")
	flag.StringVar(&imei, "i", "", "Device IMEI or TAC (8 digits)")
	flag.StringVar(&serial, "s", "", "Device Serial Number")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || model == "" || region == "" {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "checkupdate":
		checkUpdate()
	case "download":
		parseDownloadFlags(args[1:])
		download()
	case "decrypt":
		parseDecryptFlags(args[1:])
		decrypt()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`susgo - Samsung Firmware Downloader

Usage:
  susgo -m <model> -r <region> checkupdate
  susgo -m <model> -r <region> -i <IMEI/TAC> download [-O <dir> | -o <file>] [-v <version>]
  susgo -m <model> -r <region> -i <IMEI/TAC> decrypt -v <version> -I <input> -o <output>

Options:
  -m  Device model (e.g., SM-G998B)
  -r  Device region code (e.g., EUX, XAR)
  -i  Device IMEI (15 digits) or TAC (8 digits)
  -s  Device Serial Number (for devices without IMEI)

Commands:
  checkupdate  Check the latest firmware version
  download     Download firmware
  decrypt      Decrypt encrypted firmware

Download Options:
  -O  Output directory
  -o  Output file
  -v  Firmware version (optional, uses latest if not specified)

Decrypt Options:
  -v  Firmware version
  -I  Input file (encrypted)
  -o  Output file (decrypted)
  -V  Encryption version (2 or 4, default 4)

Examples:
  susgo -m SM-G998B -r EUX checkupdate
  susgo -m SM-G998B -r EUX -i 35123456 download -O .
  susgo -m SM-G998B -r EUX -i 351234567890123 decrypt -v VER/CODE -I file.enc4 -o file.zip`)
}

func parseDownloadFlags(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	fs.StringVar(&version, "v", "", "Firmware version")
	fs.StringVar(&outDir, "O", "", "Output directory")
	fs.StringVar(&outFile, "o", "", "Output file")
	fs.BoolVar(&resume, "R", false, "Resume download")
	fs.BoolVar(&showMD5, "M", false, "Show MD5 hash")
	fs.Parse(args)

	if outDir == "" && outFile == "" {
		fmt.Println("Error: Either -O or -o must be specified")
		os.Exit(1)
	}
}

func parseDecryptFlags(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	fs.StringVar(&version, "v", "", "Firmware version")
	fs.StringVar(&inFile, "I", "", "Input file")
	fs.StringVar(&outFile, "o", "", "Output file")
	fs.IntVar(&encVer, "V", 4, "Encryption version")
	fs.Parse(args)

	if version == "" || inFile == "" || outFile == "" {
		fmt.Println("Error: -v, -I, and -o are required for decrypt")
		os.Exit(1)
	}
}

func checkUpdate() {
	ver, err := getLatestVersion(model, region)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(ver)
}

func download() {
	effectiveIMEI, err := parseIMEI()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := NewFUSClient()

	if version == "" {
		ver, err := getLatestVersion(model, region)
		if err != nil {
			fmt.Printf("Error getting version: %v\n", err)
			os.Exit(1)
		}
		version = ver
	}

	path, filename, size, err := getBinaryFile(client, version, model, region, effectiveIMEI)
	if err != nil {
		fmt.Printf("Error getting binary info: %v\n", err)
		os.Exit(1)
	}

	out := outFile
	if out == "" {
		out = filepath.Join(outDir, filename)
	} else {
		if info, err := os.Stat(out); err == nil && info.IsDir() {
			out = filepath.Join(out, filename)
		}
	}

	fmt.Println("Device:", model)
	fmt.Println("CSC:", region)
	fmt.Println("FW Version:", version)
	fmt.Printf("FW Size: %.3f GB\n", float64(size)/(1024*1024*1024))
	fmt.Println("File Path:", out)

	decryptedFile := strings.TrimSuffix(strings.TrimSuffix(out, ".enc4"), ".enc2")
	if _, err := os.Stat(decryptedFile); err == nil {
		fmt.Println("File already downloaded and decrypted!")
		return
	}

	var dlOffset int64
	if info, err := os.Stat(out); err == nil {
		dlOffset = info.Size()
		if dlOffset == size {
			fmt.Println("Already downloaded!")
			autoDecrypt(out, filename, effectiveIMEI)
			return
		}
		fmt.Println("Resuming", filename)
	} else {
		fmt.Println("Downloading", filename)
	}

	initDownload(client, filename)

	resp, err := client.DownloadFile(path+filename, dlOffset)
	if err != nil {
		fmt.Printf("Error downloading: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if showMD5 {
		if md5Header := resp.Header.Get("Content-MD5"); md5Header != "" {
			decoded, _ := base64.StdEncoding.DecodeString(md5Header)
			fmt.Printf("MD5: %x\n", decoded)
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if dlOffset > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	fd, err := os.OpenFile(out, flags, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	buf := make([]byte, 0x10000)
	downloaded := dlOffset
	lastProgress := 0

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fd.Write(buf[:n])
			downloaded += int64(n)

			progress := int(float64(downloaded) / float64(size) * 100)
			if progress != lastProgress && progress%5 == 0 {
				fmt.Printf("\rDownloading: %d%%", progress)
				lastProgress = progress
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fd.Close()
			fmt.Printf("\nError during download: %v\n", err)
			os.Exit(1)
		}
	}
	fd.Close()
	fmt.Println("\rDownloading: 100%")
	fmt.Println("Download completed.")

	autoDecrypt(out, filename, effectiveIMEI)
}

func parseIMEI() (string, error) {
	if imei != "" {
		if len(imei) == 8 {
			return validateAndGenerateIMEI(imei, model, region)
		} else if len(imei) == 15 {
			fmt.Println("IMEI is provided:", imei)
			return imei, nil
		} else {
			return "", fmt.Errorf("invalid IMEI length: please provide 8 or 15 digits")
		}
	} else if serial != "" {
		fmt.Println("Serial Number is provided:", serial)
		return serial, nil
	}
	return "", fmt.Errorf("IMEI or Serial Number is required")
}

func getBinaryFile(client *FUSClient, fw, model, region, imei string) (path, filename string, size int64, err error) {
	req := binaryInform(fw, model, region, imei, client.Nonce)
	resp, err := client.MakeReq("NF_DownloadBinaryInform.do", req)
	if err != nil {
		return "", "", 0, err
	}

	var fusResp FUSMsgResponse
	if err := xml.Unmarshal([]byte(resp), &fusResp); err != nil {
		return "", "", 0, err
	}

	if fusResp.Body.Results.Status != 200 {
		return "", "", 0, fmt.Errorf("DownloadBinaryInform returned %d", fusResp.Body.Results.Status)
	}

	filename = fusResp.Body.Put.BinaryName.Data
	if filename == "" {
		return "", "", 0, fmt.Errorf("failed to find firmware bundle")
	}

	size = fusResp.Body.Put.BinaryByteSize.Data
	path = fusResp.Body.Put.ModelPath.Data

	return path, filename, size, nil
}

func initDownload(client *FUSClient, filename string) {
	req := binaryInit(filename, client.Nonce)
	client.MakeReq("NF_DownloadBinaryInitForMass.do", req)
}

func autoDecrypt(out, filename, effectiveIMEI string) {
	dec := strings.TrimSuffix(strings.TrimSuffix(out, ".enc4"), ".enc2")
	if _, err := os.Stat(dec); err == nil {
		fmt.Printf("File %s already exists, refusing to auto-decrypt!\n", dec)
		return
	}

	fmt.Println("\nDecrypting", out)

	var key []byte
	var err error
	if strings.HasSuffix(filename, ".enc2") {
		key = getV2Key(version, model, region)
	} else {
		key, err = getV4Key(version, model, region, effectiveIMEI)
		if err != nil {
			fmt.Printf("Error getting decryption key: %v\n", err)
			return
		}
	}

	if err := decryptFirmware(out, dec, key, true); err != nil {
		fmt.Printf("Error decrypting: %v\n", err)
		return
	}

	os.Remove(out)
	fmt.Printf("\nFile %s has been decrypted.\n", out)
}

func decrypt() {
	effectiveIMEI, err := parseIMEI()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var key []byte
	if encVer == 2 {
		key = getV2Key(version, model, region)
	} else {
		key, err = getV4Key(version, model, region, effectiveIMEI)
		if err != nil {
			fmt.Printf("Error getting key: %v\n", err)
			os.Exit(1)
		}
	}

	if err := decryptFirmware(inFile, outFile, key, true); err != nil {
		fmt.Printf("Error decrypting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Decryption completed.")
}
