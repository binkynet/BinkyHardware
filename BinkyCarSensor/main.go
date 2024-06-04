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
	PWMBySlice = []pwm{
		machine.PWM0,
		machine.PWM1,
		machine.PWM2,
		machine.PWM3,
		machine.PWM4,
		machine.PWM5,
		machine.PWM6,
		machine.PWM7,
	}
)

type pwm interface {
	// Configure enables and configures this PWM.
	Configure(config machine.PWMConfig) error
	// Channel returns a PWM channel for the given pin. If pin does
	// not belong to PWM peripheral ErrInvalidOutputPin error is returned.
	// It also configures pin as PWM output.
	Channel(pin machine.Pin) (channel uint8, err error)
	// SetPeriod updates the period of this PWM peripheral in nanoseconds.
	// To set a particular frequency, use the following formula:
	//
	//	period = 1e9 / frequency
	//
	// Where frequency is in hertz. If you use a period of 0, a period
	// that works well for LEDs will be picked.
	//
	// SetPeriod will try not to modify TOP if possible to reach the target period.
	// If the period is unattainable with current TOP SetPeriod will modify TOP
	// by the bare minimum to reach the target period. It will also enable phase
	// correct to reach periods above 130ms.
	SetPeriod(period uint64) error
	// Top returns the current counter top, for use in duty cycle calculation.
	//
	// The value returned here is hardware dependent. In general, it's best to treat
	// it as an opaque value that can be divided by some number and passed to Set
	// (see Set documentation for more information).
	Top() uint32
	// Set updates the channel value. This is used to control the channel duty
	// cycle, in other words the fraction of time the channel output is high (or low
	// when inverted). For example, to set it to a 25% duty cycle, use:
	//
	//	pwm.Set(channel, pwm.Top() / 4)
	//
	// pwm.Set(channel, 0) will set the output to low and pwm.Set(channel,
	// pwm.Top()) will set the output to high, assuming the output isn't inverted.
	Set(channel uint8, value uint32)
	// Enable enables or disables PWM peripheral channels.
	Enable(enable bool)
}

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
