package main

import (
	"crypto/aes"
	"crypto/md5"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
	threads int
	showMD5 bool
	latest  bool
	quiet   bool
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
  susgo -m <model> -r <region> download [-i <IMEI/TAC>] [-O <dir> | -o <file>] [-v <ver>] [-j <threads>]
  susgo -m <model> -r <region> -i <IMEI/TAC> decrypt -v <ver> -I <input> -o <output>

Options:
  -m  Device model (e.g., SM-S928B)
  -r  Region code (e.g., EUX, XAR)
  -i  IMEI (15 digits) or TAC (8 digits), optional for download
  -s  Serial Number (for devices without IMEI)

Commands:
  checkupdate  Check latest firmware version
  list         List all available firmware versions
  download     Download and decrypt firmware (parallel connections)
  decrypt      Decrypt encrypted firmware

List Options:
  -l  Show only latest version
  -q  Quiet mode (version only)

Download Options:
  -O  Output directory (default: current directory)
  -o  Output file
  -v  Firmware version (optional, defaults to latest)
  -j  Number of parallel connections (default: 8)
  -M  Show MD5 hash after download

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
	fs.IntVar(&threads, "j", 8, "Number of parallel connections")
	fs.BoolVar(&showMD5, "M", false, "Show MD5 hash")
	fs.Parse(args)
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
	// IMEI is optional — without it, uses samloader-rs auto mode (ACCESS_MODE 5)
	var effectiveIMEI string
	if imei != "" || serial != "" {
		var err error
		effectiveIMEI, err = parseIMEI()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	client := NewFUSClient()

	// With IMEI mode, fetch version from FOTA if not specified
	if effectiveIMEI != "" && version == "" {
		ver, err := getLatestVersion(model, region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		version = ver
	}

	path, filename, size, key, fwVer, err := getBinaryFile(client, version, model, region, effectiveIMEI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	version = fwVer

	// Output name strips encryption suffix (decryption happens on-the-fly)
	defaultName := strings.TrimSuffix(strings.TrimSuffix(filename, ".enc4"), ".enc2")

	var out string
	switch {
	case outFile != "":
		out = outFile
	case outDir != "":
		out = filepath.Join(outDir, defaultName)
	default:
		out = defaultName
	}

	fmt.Printf("Firmware Version: %s\n", version)
	fmt.Printf("Downloading %s to %s\n", filename, out)

	if _, err := os.Stat(out); err == nil {
		fmt.Printf("%s already exists, skipping.\n", out)
		return
	}

	// Pre-allocate output file
	f, err := os.OpenFile(out, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := f.Truncate(size); err != nil {
		f.Close()
		os.Remove(out)
		fmt.Fprintf(os.Stderr, "Error: cannot pre-allocate file: %v\n", err)
		os.Exit(1)
	}

	// Adjust thread count for small files
	numThreads := int64(threads)
	if numThreads < 1 {
		numThreads = 1
	}
	if size/16 < numThreads {
		numThreads = size / 16
	}
	if numThreads < 1 {
		numThreads = 1
	}
	// Chunk size aligned to 16-byte AES block boundary
	chunkSize := (size/numThreads/16 + 1) * 16

	bar := NewProgressBar(size)
	bar.Start()

	initDownload(client, filename)

	var wg sync.WaitGroup
	var dlErr error
	var errOnce sync.Once

	for i := int64(0); i < numThreads; i++ {
		start := i * chunkSize
		if start >= size {
			break
		}

		isLast := i == numThreads-1
		var end int64
		if isLast {
			end = -1
		} else {
			end = start + chunkSize - 1
		}

		resp, err := client.DownloadFileRange(path+filename, start, end)
		if err != nil {
			f.Close()
			os.Remove(out)
			bar.Finish()
			fmt.Fprintf(os.Stderr, "\nError: download request failed: %v\n", err)
			os.Exit(1)
		}

		writeOffset := start
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer resp.Body.Close()

			block, _ := aes.NewCipher(key)
			readBuf := make([]byte, 32768)
			dataBuf := make([]byte, 0, 32768+16)
			pos := writeOffset

			for {
				n, readErr := resp.Body.Read(readBuf)
				if n > 0 {
					bar.Add(int64(n))
					dataBuf = append(dataBuf, readBuf[:n]...)

					// Decrypt complete 16-byte AES blocks in-place
					aligned := (len(dataBuf) / 16) * 16
					for j := 0; j < aligned; j += 16 {
						block.Decrypt(dataBuf[j:j+16], dataBuf[j:j+16])
					}

					if aligned > 0 {
						if _, err := f.WriteAt(dataBuf[:aligned], pos); err != nil {
							errOnce.Do(func() { dlErr = err })
							return
						}
						pos += int64(aligned)
					}

					// Keep incomplete block remainder for next iteration
					remainder := len(dataBuf) - aligned
					if remainder > 0 {
						copy(dataBuf[:remainder], dataBuf[aligned:])
					}
					dataBuf = dataBuf[:remainder]
				}

				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					errOnce.Do(func() { dlErr = readErr })
					return
				}
			}
		}()

		// Small delay between requests to avoid overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

	if dlErr != nil {
		f.Close()
		os.Remove(out)
		bar.Finish()
		fmt.Fprintf(os.Stderr, "\nDownload failed: %v\n", dlErr)
		os.Exit(1)
	}

	// Remove PKCS7 padding
	lastByte := make([]byte, 1)
	if _, err := f.ReadAt(lastByte, size-1); err == nil {
		if lastByte[0] > 0 && lastByte[0] <= 16 {
			f.Truncate(size - int64(lastByte[0]))
		}
	}

	f.Close()
	bar.Finish()

	if showMD5 {
		hash := md5.New()
		outFd, err := os.Open(out)
		if err == nil {
			io.Copy(hash, outFd)
			outFd.Close()
			fmt.Printf("MD5: %x\n", hash.Sum(nil))
		}
	}

	fmt.Println("Done.")
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

func getBinaryFile(client *FUSClient, fw, model, region, imei string) (path, filename string, size int64, key []byte, fwVersion string, err error) {
	var req string
	if imei == "" {
		// Auto mode: ACCESS_MODE 5, no IMEI needed (samloader-rs approach)
		req = binaryInformAuto(model, region)
	} else {
		req = binaryInform(fw, model, region, imei, client.Nonce)
	}

	resp, err := client.MakeReq("NF_DownloadBinaryInform.do", req)
	if err != nil {
		return "", "", 0, nil, "", err
	}

	var fusResp FUSMsgResponse
	if err := xml.Unmarshal([]byte(resp), &fusResp); err != nil {
		return "", "", 0, nil, "", err
	}

	if fusResp.Body.Results.Status != 200 {
		return "", "", 0, nil, "", fmt.Errorf("status %d", fusResp.Body.Results.Status)
	}

	filename = fusResp.Body.Put.BinaryName.Data
	if filename == "" {
		return "", "", 0, nil, "", fmt.Errorf("no firmware found")
	}

	// Use server-reported version
	fwVersion = fusResp.Body.Results.LatestFWVersion.Data
	if fwVersion == "" {
		fwVersion = fw
	}

	// Compute decryption key from server response
	if strings.HasSuffix(filename, ".enc2") {
		key = getV2Key(fwVersion, model, region)
	} else {
		logicVal := fusResp.Body.Put.LogicValueFactory.Data
		decKey := getLogicCheck(fwVersion, logicVal)
		hash := md5.Sum([]byte(decKey))
		key = hash[:]
	}

	return fusResp.Body.Put.ModelPath.Data, filename, fusResp.Body.Put.BinaryByteSize.Data, key, fwVersion, nil
}

func initDownload(client *FUSClient, filename string) {
	req := binaryInit(filename, client.Nonce)
	client.MakeReq("NF_DownloadBinaryInitForMass.do", req)
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
