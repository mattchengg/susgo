package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type VersionXML struct {
	XMLName  xml.Name `xml:"versioninfo"`
	Firmware struct {
		Version struct {
			Latest string `xml:"latest"`
		} `xml:"version"`
	} `xml:"firmware"`
}

func normalizeVerCode(vercode string) string {
	ver := strings.Split(vercode, "/")
	if len(ver) == 3 {
		ver = append(ver, ver[0])
	}
	if len(ver) >= 3 && ver[2] == "" {
		ver[2] = ver[0]
	}
	return strings.Join(ver, "/")
}

func getLatestVersion(model, region string) (string, error) {
	url := fmt.Sprintf("https://fota-cloud-dn.ospserver.net/firmware/%s/%s/version.xml", region, model)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Kies2.0_FUS")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return "", fmt.Errorf("model or region not found (403)")
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var versionInfo VersionXML
	if err := xml.Unmarshal(body, &versionInfo); err != nil {
		return "", err
	}

	if versionInfo.Firmware.Version.Latest == "" {
		return "", fmt.Errorf("no latest firmware available")
	}

	return normalizeVerCode(versionInfo.Firmware.Version.Latest), nil
}
