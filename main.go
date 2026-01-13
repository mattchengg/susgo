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
	model    string
	region   string
	imei     string
	serial   string
	version  string
	outDir   string
	outFile  string
	inFile   string
	encVer   int
	showMD5  bool
	latest   bool
	quiet    bool
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

	switch args[0] {
	case "checkupdate":
		checkUpdate()
	case "list":
		parseListFlags(args[1:])
		listFirmware()
	case "download":
		parseDownloadFlags(args[1:])
		download()
	case "decrypt":
		parseDecryptFlags(args[1:])
		decrypt()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`susgo - Samsung Firmware Downloader

Usage:
  susgo -m <model> -r <region> checkupdate
  susgo -m <model> -r <region> list [-l] [-q]
  susgo -m <model> -r <region> -i <IMEI/TAC> download [-O <dir> | -o <file>] [-v <ver>]
  susgo -m <model> -r <region> -i <IMEI/TAC> decrypt -v <ver> -I <input> -o <output>

Options:
  -m  Device model (e.g., SM-S928B)
  -r  Region code (e.g., EUX, XAR)
  -i  IMEI (15 digits) or TAC (8 digits)
  -s  Serial Number (for devices without IMEI)

Commands:
  checkupdate  Check latest firmware version
  list         List all available firmware versions
  download     Download firmware
  decrypt      Decrypt encrypted firmware

List Options:
  -l  Show only latest version
  -q  Quiet mode (version only)

Download Options:
  -O  Output directory
  -o  Output file
  -v  Firmware version (optional)
  -M  Show MD5 hash

Decrypt Options:
  -v  Firmware version
  -I  Input file
  -o  Output file
  -V  Encryption version (2 or 4, default 4)
`)
}

func parseListFlags(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.BoolVar(&latest, "l", false, "Show only latest")
	fs.BoolVar(&quiet, "q", false, "Quiet mode")
	fs.Parse(args)
}

func parseDownloadFlags(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	fs.StringVar(&version, "v", "", "Firmware version")
	fs.StringVar(&outDir, "O", "", "Output directory")
	fs.StringVar(&outFile, "o", "", "Output file")
	fs.BoolVar(&showMD5, "M", false, "Show MD5 hash")
	fs.Parse(args)
	if outDir == "" && outFile == "" {
		fmt.Println("Error: -O or -o required")
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
		fmt.Println("Error: -v, -I, -o required")
		os.Exit(1)
	}
}

func checkUpdate() {
	ver, err := getLatestVersion(model, region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(ver)
}

func listFirmware() {
	info, err := getVersionInfo(model, region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if quiet {
		fmt.Println(info.Latest.Version)
		if !latest {
			for _, u := range info.Upgrade {
				fmt.Println(u.Version)
			}
		}
		return
	}

	fmt.Printf("Model: %s  Region: %s\n\n", model, region)
	fmt.Println("Latest:")
	fmt.Printf("  %s\n", info.Latest.Version)

	if !latest && len(info.Upgrade) > 0 {
		fmt.Println("\nAvailable Upgrades:")
		for _, u := range info.Upgrade {
			sizeStr := ""
			if u.Size > 0 {
				sizeStr = fmt.Sprintf(" (%.2f GB)", float64(u.Size)/(1024*1024*1024))
			}
			fmt.Printf("  %s%s\n", u.Version, sizeStr)
		}
	}
}

func download() {
	effectiveIMEI, err := parseIMEI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := NewFUSClient()

	if version == "" {
		ver, err := getLatestVersion(model, region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		version = ver
	}

	path, filename, size, err := getBinaryFile(client, version, model, region, effectiveIMEI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out := outFile
	if out == "" {
		out = filepath.Join(outDir, filename)
	} else if info, err := os.Stat(out); err == nil && info.IsDir() {
		out = filepath.Join(out, filename)
	}

	fmt.Printf("Device: %s | CSC: %s\nFW: %s\nSize: %.3f GB\nPath: %s\n",
		model, region, version, float64(size)/(1024*1024*1024), out)

	decFile := strings.TrimSuffix(strings.TrimSuffix(out, ".enc4"), ".enc2")
	if _, err := os.Stat(decFile); err == nil {
		fmt.Println("Already decrypted!")
		return
	}

	var offset int64
	if info, err := os.Stat(out); err == nil {
		offset = info.Size()
		if offset == size {
			fmt.Println("Downloaded, decrypting...")
			autoDecrypt(out, filename, effectiveIMEI)
			return
		}
		fmt.Printf("Resuming from %.1f%%\n", float64(offset)/float64(size)*100)
	}

	initDownload(client, filename)
	resp, err := client.DownloadFile(path+filename, offset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if showMD5 {
		if h := resp.Header.Get("Content-MD5"); h != "" {
			if d, err := base64.StdEncoding.DecodeString(h); err == nil {
				fmt.Printf("MD5: %x\n", d)
			}
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if offset > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	fd, err := os.OpenFile(out, flags, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create async progress bar
	bar := NewProgressBar(size)
	bar.SetCurrent(offset)
	bar.Start()

	buf := make([]byte, 32768)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fd.Write(buf[:n])
			bar.Add(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fd.Close()
			bar.Finish()
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}
	}
	fd.Close()
	bar.Finish()
	fmt.Println("Done.")
	autoDecrypt(out, filename, effectiveIMEI)
}

func parseIMEI() (string, error) {
	if imei != "" {
		switch len(imei) {
		case 8:
			return validateAndGenerateIMEI(imei, model, region)
		case 15:
			return imei, nil
		default:
			return "", fmt.Errorf("IMEI must be 8 or 15 digits")
		}
	}
	if serial != "" {
		return serial, nil
	}
	return "", fmt.Errorf("IMEI (-i) or Serial (-s) required")
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
		return "", "", 0, fmt.Errorf("status %d", fusResp.Body.Results.Status)
	}

	filename = fusResp.Body.Put.BinaryName.Data
	if filename == "" {
		return "", "", 0, fmt.Errorf("no firmware found")
	}

	return fusResp.Body.Put.ModelPath.Data, filename, fusResp.Body.Put.BinaryByteSize.Data, nil
}

func initDownload(client *FUSClient, filename string) {
	req := binaryInit(filename, client.Nonce)
	client.MakeReq("NF_DownloadBinaryInitForMass.do", req)
}

func autoDecrypt(out, filename, effectiveIMEI string) {
	dec := strings.TrimSuffix(strings.TrimSuffix(out, ".enc4"), ".enc2")
	if _, err := os.Stat(dec); err == nil {
		fmt.Printf("%s exists\n", dec)
		return
	}

	fmt.Print("Decrypting...")
	var key []byte
	var err error
	if strings.HasSuffix(filename, ".enc2") {
		key = getV2Key(version, model, region)
	} else {
		key, err = getV4Key(version, model, region, effectiveIMEI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Key error: %v\n", err)
			return
		}
	}

	if err := decryptFirmware(out, dec, key, false); err != nil {
		fmt.Fprintf(os.Stderr, "Decrypt error: %v\n", err)
		return
	}
	os.Remove(out)
	fmt.Println(" Done.")
}

func decrypt() {
	effectiveIMEI, err := parseIMEI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var key []byte
	if encVer == 2 {
		key = getV2Key(version, model, region)
	} else {
		key, err = getV4Key(version, model, region, effectiveIMEI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Key error: %v\n", err)
			os.Exit(1)
		}
	}

	if err := decryptFirmware(inFile, outFile, key, true); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}
