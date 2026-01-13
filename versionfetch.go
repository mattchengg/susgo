package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FirmwareSpec struct {
	Version string
	Size    int64
}

type VersionInfo struct {
	Latest  FirmwareSpec
	Upgrade []FirmwareSpec
}

type VersionXML struct {
	XMLName  xml.Name `xml:"versioninfo"`
	Firmware struct {
		Version struct {
			Latest  string `xml:"latest"`
			Upgrade struct {
				Value []struct {
					Text   string `xml:",chardata"`
					FWSize string `xml:"fwsize,attr"`
				} `xml:"value"`
			} `xml:"upgrade"`
		} `xml:"version"`
	} `xml:"firmware"`
}

var httpClient = &http.Client{Timeout: 3 * time.Second}

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

func fetchVersionXML(model, region string) ([]byte, error) {
	url := fmt.Sprintf("https://fota-cloud-dn.ospserver.net/firmware/%s/%s/version.xml", region, model)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Kies2.0_FUS")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("model or region not found")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func getLatestVersion(model, region string) (string, error) {
	body, err := fetchVersionXML(model, region)
	if err != nil {
		return "", err
	}

	var v VersionXML
	if err := xml.Unmarshal(body, &v); err != nil {
		return "", err
	}

	if v.Firmware.Version.Latest == "" {
		return "", fmt.Errorf("no firmware available")
	}
	return normalizeVerCode(v.Firmware.Version.Latest), nil
}

func getVersionInfo(model, region string) (*VersionInfo, error) {
	body, err := fetchVersionXML(model, region)
	if err != nil {
		return nil, err
	}

	var v VersionXML
	if err := xml.Unmarshal(body, &v); err != nil {
		return nil, err
	}

	info := &VersionInfo{}
	if v.Firmware.Version.Latest != "" {
		info.Latest = FirmwareSpec{Version: normalizeVerCode(v.Firmware.Version.Latest)}
	}

	for _, u := range v.Firmware.Version.Upgrade.Value {
		size, _ := strconv.ParseInt(u.FWSize, 10, 64)
		info.Upgrade = append(info.Upgrade, FirmwareSpec{
			Version: normalizeVerCode(u.Text),
			Size:    size,
		})
	}
	return info, nil
}
