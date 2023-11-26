package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/ads1115"
	"tinygo.org/x/drivers/ws2812"
)

// Try to detect ADS1115 addresses.
// Only when 1 or 2 devices are found, are they returned.
func probeADS1115Devices(led ws2812.Device) ([]*ads1115.Device, color.RGBA) {
	// Configure ADS1115 I2C channel (i2c0)
	var adsDevs []*ads1115.Device
	for {
		println("Configure i2c0...")
		if err := machine.I2C0.Configure(machine.I2CConfig{}); err != nil {
			led.WriteColors([]color.RGBA{colorI2cConfigError})
		} else {
			println("Probing ADS1115 devices")
			adsDevs = nil
			for _, i2cAddress := range []uint8{ads1115.I2CAddressGround, ads1115.I2CAddressVDD, ads1115.I2CAddressSDA, ads1115.I2CAddressSCL} {
				// Create address and try to read a value
				if dev, err := probeADS1115Device(i2cAddress); err == nil {
					// Found valid ads1115
					println("Found ADS1115 at address: ", i2cAddress)
					adsDevs = append(adsDevs, dev)
				}
			}
			switch len(adsDevs) {
			case 0:
				led.WriteColors([]color.RGBA{colorNoAdsDevsFound})
			case 1:
				led.WriteColors([]color.RGBA{colorNoDetections1AdsDevFound})
				return adsDevs, colorNoDetections1AdsDevFound
			case 2:
				led.WriteColors([]color.RGBA{colorNoDetections2AdsDevsFound})
				return adsDevs, colorNoDetections2AdsDevsFound
			default:
				led.WriteColors([]color.RGBA{colorTooManyAdsDevsFound})
			}

			// Wait until trying again
			time.Sleep(time.Second * 3)
		}
		led.WriteColors([]color.RGBA{colorBoot})
		time.Sleep(time.Second * 1)
	}
}

// Probe for the existence of an ADS1115 at the given address.
// If found, the device is initialized
func probeADS1115Device(i2cAddress uint8) (*ads1115.Device, error) {
	dev := ads1115.New(machine.I2C0, i2cAddress)
	if err := resetADS1115Device(dev); err != nil {
		return nil, err
	}
	return dev, nil
}

// Reset the given device to desired values.
func resetADS1115Device(dev *ads1115.Device) error {
	if err := dev.Reset(); err != nil {
		return fmt.Errorf("Reset failed: %w", err)
	}
	if err := dev.SetVoltageRangeMilliV(ads1115.ADS1115_RANGE_6144); err != nil {
		return fmt.Errorf("SetVoltageRangeMilliV failed: %w", err)
	}
	if err := dev.SetSingleChannel(0); err != nil {
		return fmt.Errorf("SetSingleChannel failed: %w", err)
	}
	return nil
}
