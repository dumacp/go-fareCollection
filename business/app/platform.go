package app

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dumacp/go-fareCollection/crosscutting/logs"
	"github.com/google/uuid"
)

const (
	url          = "https://sibus.nebulae.com.co/api/external-system-gateway/rest/collected-fare"
	username     = "jhon.doe"
	password     = "uno.2.tres"
	localCertDir = "/usr/local/share/ca-certificates/"
	templateTag  = `{
	"id": "%s",
	"endUser": {
		"id": "6bec5233-4ffc-42fc-a7be-2730304d0929",
		"name": "Daniel Tobon"
		
	},
	"fare": {
		"farePolicyId": "5763c020-2ff0-4aa3-ad79-6f012e3a7e20",
		"fareId": 22,
		"fareType": "PLAIN",
		"value":2000
		
	},
	"paymentMedium": {
		"typeId": "Card",
		"dataPreState": {},
		"dataPostState": {
			"transactionId": %d
		},
		"externalSystemId": "b726d3b0-355c-4dcf-862e-277f4686f993"
		
	},
	"terminal": {
		"id": "b726d3b0-355c-4dcf-862e-277f4686f993",
		"desc": "TTT",
		"location": {
			"type": "LOC",
			"coordinates": %v,
			"timestamp": %d
		}
		
	},
	"timestamp":%d
}`
	templateQR = `{
	"id": "%s",
	"fare": {
		"farePolicyId": "5763c020-2ff0-4aa3-ad79-6f012e3a7e20",
		"fareId": 22,
		"fareType": "PLAIN",
		"value": 2000
	  
	},
	"paymentMedium": {
		"typeId": "QR",
		"dataPreState": {
		},
		"dataPostState": {
			"transactionId": %d
		},
		"externalSystemId": "b726d3b0-355c-4dcf-862e-277f4686f993"
	  
	},
	"terminal": {
		"id": "b726d3b0-355c-4dcf-862e-277f4686f993",
		"desc": "TTT",
		"location": {
			"type": "LOC",
			"coordinates": %v,
			"timestamp": %d
		}
	  
	},
	"timestamp": %d
}`
)

var SendUsoQR = sendUso(templateQR)
var SendUsoTAG = sendUso(templateTag)

func sendUso(template string) func(name string, tid int, gps []float64, timeStamp time.Time) ([]byte, error) {

	return func(name string, tid int, gps []float64, timeStamp time.Time) ([]byte, error) {

		uid, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		jsonStr := []byte(fmt.Sprintf(template, uid, name, tid, gps, timeStamp.Nanosecond()/1000000))
		logs.LogBuild.Printf("json: %s", jsonStr)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(username, password)

		tr := loadLocalCert()
		client := &http.Client{Transport: tr}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		return ioutil.ReadAll(resp.Body)
	}
}

func loadLocalCert() *http.Transport {

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
	tr := &http.Transport{TLSClientConfig: config}
	return tr
	/**
	client := &http.Client{Transport: tr}

	// Uses local self-signed cert
	req := http.NewRequest(http.MethodGet, "https://localhost/api/version", nil)
	resp, err := client.Do(req)
	// Handle resp and err

	// Still works with host-trusted CAs!
	req = http.NewRequest(http.MethodGet, "https://example.com/", nil)
	resp, err = client.Do(req)
	// Handle resp and err
	/**/
}
