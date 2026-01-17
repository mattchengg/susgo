package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// parseIMEI validates and processes IMEI/serial input
func parseIMEI(imei, serial, model, region string) (string, error) {
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
	return "", fmt.Errorf("IMEI or Serial required")
}

// getBinaryFile retrieves firmware information from Samsung servers
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

// initDownload initializes the download session with Samsung servers
func initDownload(client *FUSClient, filename string) {
	req := binaryInit(filename, client.Nonce)
	client.MakeReq("NF_DownloadBinaryInitForMass.do", req)
}

// autoDecrypt automatically decrypts downloaded firmware
func autoDecrypt(out, filename, version, model, region, effectiveIMEI string) {
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
