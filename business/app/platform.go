package app

import (
	"bytes"
	"container/list"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/crosscutting/logs"
	"github.com/google/uuid"
)

const (
	// urlQr        = `tinyurl.com/SIBUS-QR?r=16&d=OMV-Z7-1431&p=%d`
	urlQr        = `tinyurl.com/SITIRIO?r=16&d=OMV-Z7-1431&p=%d`
	url          = "https://sitirio.somosmovilidad.gov.co/api/external-system-gateway/rest/collected-fare"
	username     = "jhon.doe"
	password     = "uno.2.tres"
	localCertDir = "/usr/local/share/ca-certificates/"
	templateTag  = `{
	"id": "%s",
	"endUser": {
		"id": "%s",
		"name": "%s"		
	},
	"fare": {
		"farePolicyId": "5763c020-2ff0-4aa3-ad79-6f012e3a7e20",
		"fareId": 22,
		"fareType": "PLAIN",
		"value":2400		
	},
	"paymentMedium": {
		"typeId": "Card",
		"dataPreState": {
			"beforeWrite": %s
		},
		"dataPostState": {
			"transactionId": %d,
			"afterWrite": %s
		},
		"externalSystemId": "b726d3b0-355c-4dcf-862e-277f4686f993"		
	},
	"terminal": {
		"id": "b726d3b0-355c-4dcf-862e-277f4686f993",
		"desc": "TTT",
		"location": {
			"type": "LOC",
			"coordinates": %s,
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
		"value": 2400	  
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
			"coordinates": %s,
			"timestamp": %d
		}	  
	},
	"timestamp": %d
}`
)

func send(jsonStr []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	logs.LogBuild.Printf("json: %s", "test")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	tr := loadLocalCert()
	client := &http.Client{Transport: tr}
	client.Timeout = 10 * time.Second

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

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	return ioutil.ReadAll(resp.Body)
}

//TODO: remove queue RAM
var queueSend = list.New()

func SendUsoTAG(name string, tid int, newtag map[string]interface{}, tag map[string]interface{}, gps []float64, timeStamp time.Time) {

	uid, err := uuid.NewUUID()
	if err != nil {
		logs.LogError.Println(err)
		return
	}

	ts := int64(timeStamp.UnixNano() / 1000000)

	beforeTag, _ := json.Marshal(tag)
	afterTag, _ := json.Marshal(newtag)

	jsonStr := []byte(fmt.Sprintf(templateTag, uid, uid, strings.Trim(name, "\x00"), beforeTag, tid, afterTag, "[0.0, 0.0]", ts, ts))
	logs.LogBuild.Printf("json: %s", jsonStr)

	el := queueSend.PushBack(jsonStr)

	// result, err := send(jsonStr)
	// if err != nil {
	// 	logs.LogWarn.Printf("POST result: %s; Error: %s", result, err)
	// 	return
	// }
	// logs.LogInfo.Printf("POST result: %s", result)
	// el := last.Prev()
	// queueSend.Remove(last)

	// if el == nil {
	// 	return
	// }
	for {
		prev := el.Prev()
		value, ok := el.Value.([]byte)
		if !ok {
			queueSend.Remove(el)
			el = prev
			if prev == nil {
				break
			}
			continue
		}
		result, err := send(value)
		if err != nil {
			logs.LogWarn.Printf("POST Error: %s", err)
			logs.LogWarn.Printf("POST result: %s", result)
			return
		}
		logs.LogInfo.Printf("POST result: %s", result)

		queueSend.Remove(el)
		el = prev
		if prev == nil {
			break
		}
	}

	return

}

func SendUsoQR(tid int, gps []float64, timeStamp time.Time) {

	uid, err := uuid.NewUUID()
	if err != nil {
		logs.LogError.Println(err)
		return
		// return nil, err
	}
	ts := int64(timeStamp.UnixNano() / 1000000)

	jsonStr := []byte(fmt.Sprintf(templateQR, uid, tid, "[0.0, 0.0]", ts, ts))
	logs.LogBuild.Printf("json: %s", jsonStr)

	// return send(jsonStr)

	el := queueSend.PushBack(jsonStr)

	// result, err := send(jsonStr)
	// if err != nil {
	// 	logs.LogWarn.Printf("POST result: %s; Error: %s", result, err)
	// 	return
	// }
	// logs.LogInfo.Printf("POST result: %s", result)
	// el := last.Prev()
	// queueSend.Remove(last)

	// if el == nil {
	// 	return
	// }
	for {
		prev := el.Prev()
		value, ok := el.Value.([]byte)
		if !ok {
			queueSend.Remove(el)
			el = prev
			if prev == nil {
				break
			}
			continue
		}
		result, err := send(value)
		if err != nil {
			logs.LogWarn.Printf("POST result: %s; Error: %s", result, err)
			return
		}
		logs.LogInfo.Printf("POST result: %s", result)

		queueSend.Remove(el)
		el = prev
		if prev == nil {
			break
		}
	}

	return

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
