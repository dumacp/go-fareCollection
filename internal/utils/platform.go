package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
)

var (
	urlQr = `tinyurl.com/SIBUS-QR?r=16&d=OMV-Z7-1431&p=%d`
	url   = "https://sibus.nebulae.com.co/api/external-system-gateway/rest/collected-fare"
)

const (
	username     = "jhon.doe"
	password     = "uno.2.tres"
	localCertDir = "/usr/local/share/ca-certificates/"
)

func Post(client *http.Client,
	url, username, password string,
	jsonStr []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	if client == nil {
		tr := LoadLocalCert()
		client = &http.Client{Transport: tr}
		client.Timeout = 30 * time.Second
	}

	logs.LogBuild.Printf("Post request: %+v", req)

	var resp *http.Response
	rangex := make([]int, 3)
	for range rangex {
		resp, err = client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("StatusCode: %d, resp: %s, req: %s", resp.StatusCode, body, req.URL)
	}
	return ioutil.ReadAll(resp.Body)
}

func Get(client *http.Client,
	url, username, password string, jsonStr []byte) ([]byte, error) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	if client == nil {
		tr := LoadLocalCert()
		client = &http.Client{Transport: tr}
		client.Timeout = 30 * time.Second
	}

	logs.LogBuild.Printf("Get request: %+v", req)
	var resp *http.Response
	rangex := make([]int, 3)
	for range rangex {
		resp, err = client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("StatusCode: %d, resp: %s, req: %s", resp.StatusCode, body, req.URL)
	}
	return ioutil.ReadAll(resp.Body)
}

func LoadLocalCert() *http.Transport {

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadDir(localCertDir)
	if err != nil {
		log.Fatalf("Failed to append %q to RootCAs: %v", localCertDir, err)
	}
	for _, cert := range certs {
		file, err := ioutil.ReadFile(localCertDir + cert.Name())
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", cert, err)
		}
		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(file); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		//InsecureSkipVerify: *insecure,
		RootCAs: rootCAs,
	}
	tr := &http.Transport{
		TLSClientConfig: config,
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		// Dial: (&net.Dialer{
		// 	Timeout:   30 * time.Second,
		// 	KeepAlive: 60 * time.Second,
		// }).Dial,
		// TLSHandshakeTimeout:   10 * time.Second,
		// ResponseHeaderTimeout: 10 * time.Second,
		// ExpectContinueTimeout: 3 * time.Second,
	}
	return tr
}
