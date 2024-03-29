package main

import (
	"errors"
	"image/color"
	"time"

	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/ads1115"
	"tinygo.org/x/drivers/ws2812"
)

// Keep probing sensors
func probeSensors(sensors []*Sensor, adsDevs []*ads1115.Device,
	led ws2812.Device, baseColor color.RGBA, sensorStatus chan uint8) {
	for {
		if err := probeSensorsOnce(sensors, led, baseColor, sensorStatus); err != nil {
			// Wait a bit
			time.Sleep(time.Millisecond * 200)
			// Reset ADS devices
			for idx, dev := range adsDevs {
				if err := resetADS1115Device(dev); err != nil {
					println("Failed to reset ADS1115 device: ", idx, err)
				} else {
					println("Succesfully reset ADS1115 device: ", idx)
				}
			}
		} else {
			time.Sleep(probeInterval)
		}
	}
}

// Probe all sensors once
func probeSensorsOnce(sensors []*Sensor,
	led ws2812.Device, baseColor color.RGBA, sensorStatus chan uint8) error {
	activeCount := uint8(0)
	var allErrs error
	status := uint8(0)
	for idx, s := range sensors {
		if err := s.Probe(); err != nil {
			println("probe failed: ", err)
			allErrs = errors.Join(allErrs, err)
		}
		if s.IsActive() {
			activeCount++
			status |= 1 << idx
		}
	}
	sensorStatus <- status

	if allErrs != nil {
		baseColor = color.RGBA{R: 255, G: 0, B: 0}
	} else if activeCount > 0 {
		baseColor.R = 0
		baseColor.G = 0
		baseColor.B = 120 + activeCount*16
	}
	led.WriteColors([]color.RGBA{baseColor})

	return allErrs
}
