package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

var (
	// Color scheme
	colorBoot                      = color.RGBA{R: 255, G: 165, B: 0}
	colorI2cConfigError            = color.RGBA{R: 96, G: 0, B: 96}
	colorNoAdsDevsFound            = color.RGBA{R: 245, G: 0, B: 0}
	colorTooManyAdsDevsFound       = color.RGBA{R: 96, G: 0, B: 0}
	colorNoDetections1AdsDevFound  = color.RGBA{R: 0, G: 245, B: 0}
	colorNoDetections2AdsDevsFound = color.RGBA{R: 0, G: 96, B: 0}
)

var (
	LedRed    = machine.GPIO6
	LedGreen  = machine.GPIO7
	LedYellow = machine.GPIO10

	IO = []machine.Pin{
		machine.GPIO29,
		machine.GPIO28,
		machine.GPIO27,
		machine.GPIO26,
		machine.GPIO15,
		machine.GPIO14,
		machine.GPIO13,
		machine.GPIO12,
	}
)

const (
	defaultI2cAddress = uint8(0x34)
	altI2cAddress     = uint8(0x35)
)

func main() {
	// Configure leds
	LedRed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LedGreen.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LedYellow.Configure(machine.PinConfig{Mode: machine.PinOutput})
	// Configure IO pins
	for _, p := range IO {
		p.High()
		p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	// Set initial state
	LedRed.Low()    // Turn on
	LedGreen.High() // Turn off
	LedYellow.Low() // Turn on

	time.Sleep(time.Second * 5)

	// Detect I2C address
	i2cAddress := defaultI2cAddress
	if !IO[0].Get() {
		// IO1 pull down to GND
		i2cAddress = altI2cAddress
	}
	println("Found i2c address: ", i2cAddress)

	// Configure neopixel
	machine.NEOPIXEL.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led := ws2812.New(machine.NEOPIXEL)
	led.WriteColors([]color.RGBA{colorBoot})

	// Configure ADS1115 I2C channel (i2c0)
	adsDevs, baseColor := probeADS1115Devices(led)

	// Prepare sensor
	sensors := make([]*Sensor, 0, len(adsDevs)*4)
	for _, adsDev := range adsDevs {
		sensors = append(sensors,
			NewSensor(adsDev, 0),
			NewSensor(adsDev, 1),
			NewSensor(adsDev, 2),
			NewSensor(adsDev, 3),
		)
	}

	// Detect PCF8574 devices
	pcfDevs := probePCF8574Devices(led)

	sensorStatus := make(chan uint8)
	outputStatus := make(chan pcfOutput, 8)
	go probeSensors(sensors, adsDevs, led, baseColor, sensorStatus)
	go sendPCF8574Outputs(pcfDevs, outputStatus)
	go func() {
		for {
			if err := listenForIncomingI2CRequests(machine.I2C1, i2cAddress, sensorStatus, outputStatus, uint8(len(adsDevs)*4), uint8(len(pcfDevs)*8)); err != nil {
				println("listenForIncomingI2CRequests failed: ", err)
				time.Sleep(time.Second)
			}
		}
	}()

	// Set leds to running state
	LedRed.High()                                  // Turn off
	LedGreen.Low()                                 // Turn on
	LedYellow.Set(i2cAddress == defaultI2cAddress) // Turn off (default=0x34), turn on (alternate=0x35)

	for {
		time.Sleep(time.Minute)
	}
}
