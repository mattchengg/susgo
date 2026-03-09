package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FUSClient struct {
	Auth     string
	SessID   string
	EncNonce string
	Nonce    string
	client   *http.Client
}

func NewFUSClient() *FUSClient {
	c := &FUSClient{
		client: &http.Client{},
	}
	c.MakeReq("NF_DownloadGenerateNonce.do", "")
	return c
}

func (c *FUSClient) MakeReq(path, data string) (string, error) {
	authv := fmt.Sprintf(`FUS nonce="%s", signature="%s", nc="", type="", realm="", newauth="1"`, c.EncNonce, c.Auth)

	req, err := http.NewRequest("POST", "https://neofussvr.sslcs.cdngc.net/"+path, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", authv)
	req.Header.Set("User-Agent", "Kies2.0_FUS")
	if c.SessID != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.SessID})
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if nonce := resp.Header.Get("NONCE"); nonce != "" {
		c.EncNonce = nonce
		decrypted, err := decryptNonce(nonce)
		if err != nil {
			return "", err
		}
		c.Nonce = decrypted
		auth, err := getAuth(c.Nonce)
		if err != nil {
			return "", err
		}
		c.Auth = auth
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			c.SessID = cookie.Value
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func (c *FUSClient) DownloadFile(filename string, start int64) (*http.Response, error) {
	return c.DownloadFileRange(filename, start, -1)
}

// DownloadFileRange downloads a file with optional byte range.
// end=-1 means download to end of file.
func (c *FUSClient) DownloadFileRange(filename string, start, end int64) (*http.Response, error) {
	authv := fmt.Sprintf(`FUS nonce="%s", signature="%s", nc="", type="", realm="", newauth="1"`, c.EncNonce, c.Auth)

	url := "http://cloud-neofussvr.samsungmobile.com/NF_DownloadBinaryForMass.do?file=" + filename

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authv)
	req.Header.Set("User-Agent", "Kies2.0_FUS")
	if end >= 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	} else if start > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return resp, nil
}
