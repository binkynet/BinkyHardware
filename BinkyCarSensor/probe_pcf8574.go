package main

import (
	"image/color"
	"machine"

	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/pcf8574"
	"tinygo.org/x/drivers/ws2812"
)

// Try to detect PCF8574 addresses.
func probePCF8574Devices(led ws2812.Device) []*pcf8574.Device {
	// Configure Outgoing I2C channel (i2c0)
	var pcfDevs []*pcf8574.Device

	println("Configure i2c0...")
	if err := machine.I2C0.Configure(machine.I2CConfig{}); err != nil {
		led.WriteColors([]color.RGBA{colorI2cConfigError})
	} else {
		println("Probing PCF8574 devices")
		pcfDevs = nil
		for _, i2cAddress := range []uint8{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27} {
			// Create address and try to read a value
			if dev, err := probePCF8574Device(i2cAddress); err == nil {
				// Found valid PCF8574
				println("Found PCF8574 at address: ", i2cAddress)
				pcfDevs = append(pcfDevs, dev)
			}
		}
		println("Found ", len(pcfDevs), " PCF8574 devices")
	}
	return pcfDevs
}

// Probe for the existence of an PCF8574 at the given address.
// If found, the device is initialized
func probePCF8574Device(i2cAddress uint8) (*pcf8574.Device, error) {
	dev := pcf8574.New(machine.I2C0, i2cAddress)
	if err := dev.Reset(); err != nil {
		return nil, err
	}
	return dev, nil
}
