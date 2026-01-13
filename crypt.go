package main

import (
	"crypto/aes"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type FUSMsgResponse struct {
	XMLName xml.Name `xml:"FUSMsg"`
	Body    struct {
		Results struct {
			LatestFWVersion struct {
				Data string `xml:"Data"`
			} `xml:"LATEST_FW_VERSION"`
			Status int `xml:"Status"`
		} `xml:"Results"`
		Put struct {
			LogicValueFactory struct {
				Data string `xml:"Data"`
			} `xml:"LOGIC_VALUE_FACTORY"`
			BinaryName struct {
				Data string `xml:"Data"`
			} `xml:"BINARY_NAME"`
			BinaryByteSize struct {
				Data int64 `xml:"Data"`
			} `xml:"BINARY_BYTE_SIZE"`
			ModelPath struct {
				Data string `xml:"Data"`
			} `xml:"MODEL_PATH"`
		} `xml:"Put"`
	} `xml:"FUSBody"`
}

func getV4Key(version, model, region, imei string) ([]byte, error) {
	client := NewFUSClient()
	normalizedVer := normalizeVerCode(version)
	req := binaryInform(normalizedVer, model, region, imei, client.Nonce)
	resp, err := client.MakeReq("NF_DownloadBinaryInform.do", req)
	if err != nil {
		return nil, err
	}

	var fusResp FUSMsgResponse
	if err := xml.Unmarshal([]byte(resp), &fusResp); err != nil {
		return nil, err
	}

	fwver := fusResp.Body.Results.LatestFWVersion.Data
	logicVal := fusResp.Body.Put.LogicValueFactory.Data
	decKey := getLogicCheck(fwver, logicVal)

	hash := md5.Sum([]byte(decKey))
	return hash[:], nil
}

func getV2Key(version, model, region string) []byte {
	decKey := region + ":" + model + ":" + version
	hash := md5.Sum([]byte(decKey))
	return hash[:]
}

func decryptFirmware(inFile, outFile string, key []byte, showProgress bool) error {
	inf, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer inf.Close()

	stat, err := inf.Stat()
	if err != nil {
		return err
	}
	length := stat.Size()

	outf, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer outf.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	if length%16 != 0 {
		return fmt.Errorf("invalid input block size")
	}

	chunks := length/4096 + 1
	buf := make([]byte, 4096)
	var processed int64

	for i := int64(0); i < chunks; i++ {
		n, err := inf.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		decBlock := make([]byte, n)
		for j := 0; j < n; j += 16 {
			block.Decrypt(decBlock[j:j+16], buf[j:j+16])
		}

		if i == chunks-1 {
			decBlock = pkcs7Unpad(decBlock[:n])
		}

		outf.Write(decBlock)
		processed += int64(n)

		if showProgress && processed%(length/10+1) < 4096 {
			fmt.Printf("\rDecrypting: %.1f%%", float64(processed)/float64(length)*100)
		}
	}

	if showProgress {
		fmt.Println("\rDecrypting: 100.0%")
	}

	return nil
}
