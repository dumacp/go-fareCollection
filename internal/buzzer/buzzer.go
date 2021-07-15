package buzzer

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type buzzerValue int

const (
	buzzerGood buzzerValue = iota
	buzzerBad
)

const (
	pwmpath   = "/sys/class/pwm/pwmchip1"
	pwm0      = "/sys/class/pwm/pwmchip1/pwm0"
	pwmExport = "/sys/class/pwm/pwmchip1/export"
	pwmPeriod = "/sys/class/pwm/pwmchip1/pwm0/period"
	pwmDuty   = "/sys/class/pwm/pwmchip1/pwm0/duty_cycle"
	pwmEnable = "/sys/class/pwm/pwmchip1/pwm0/enable"
)

func BuzzerInit() error {

	if _, err := os.Stat(pwmpath); os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(pwm0); os.IsNotExist(err) {
		if err := ioutil.WriteFile(pwmExport, []byte("0"), 0644); err != nil {
			return err
		}

		if _, err := os.Stat(pwm0); os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

//TODO: fix duty and period set when duty is greather!!!!
func buzzerPlay(value buzzerValue) error {

	updatePeriodDuty := func(period int, duty int) error {
		dutyCurrent := 0
		if dutyCurrentS, err := ioutil.ReadFile(pwmDuty); err != nil {
			return err
		} else {
			if dutyCurrent, err = strconv.Atoi(strings.Trim(string(dutyCurrentS), "\n ")); err != nil {
				return err
			}
		}

		periodCurrent := 0
		if periodCurrentS, err := ioutil.ReadFile(pwmPeriod); err != nil {

			return err
		} else {
			if periodCurrent, err = strconv.Atoi(strings.Trim(string(periodCurrentS), "\n ")); err != nil {
				return err
			}
		}

		if dutyCurrent >= period {
			if err := ioutil.WriteFile(pwmDuty, []byte("0"), 0644); err != nil {
				return err
			}
			dutyCurrent = 0
		}
		if periodCurrent != period {
			if err := ioutil.WriteFile(pwmPeriod, []byte(fmt.Sprintf("%d", period)), 0644); err != nil {
				return err
			}
		}
		if dutyCurrent != duty {
			if err := ioutil.WriteFile(pwmDuty, []byte(fmt.Sprintf("%d", duty)), 0644); err != nil {
				return err
			}
		}

		return nil
	}

	timeout := time.Duration(0)
	repeat := []int{1}
	switch value {
	case buzzerGood:
		if err := updatePeriodDuty(809990, 700000); err != nil {
			return err
		}
		timeout = time.Duration(100 * time.Millisecond)
		repeat = []int{1, 2, 3}
	case buzzerBad:
		if err := updatePeriodDuty(2009990, 1700000); err != nil {
			return err
		}
		timeout = time.Duration(700 * time.Millisecond)
		repeat = []int{1, 2}
	default:
		return nil
	}
	for range repeat {
		if err := ioutil.WriteFile(pwmEnable, []byte("1"), 0644); err != nil {
			return err
		}
		time.Sleep(timeout)
		if err := ioutil.WriteFile(pwmEnable, []byte("0"), 0644); err != nil {
			return err
		}
	}
	return nil
}

//BuzzerPlayGOOD reproduce buzzer with tone that represent sucess
func BuzzerPlayGOOD() error {
	if err := buzzerPlay(buzzerGood); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

//BuzzerPlayBAD reproduce buzzer with tone that represent error
func BuzzerPlayBAD() error {
	if err := buzzerPlay(buzzerBad); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
