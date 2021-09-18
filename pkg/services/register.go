package services

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	TopicAddress = "ADDR/appfare"
)

func GetAppAddres(client mqtt.Client, clientAddress string) error {

	tk := client.Publish(TopicAddress, 0, false, []byte(clientAddress))

	if !tk.WaitTimeout(3 * time.Second) {
		return fmt.Errorf("connect wait error")
	}

	if tk.Error() != nil {
		return tk.Error()
	}
	return nil
}
