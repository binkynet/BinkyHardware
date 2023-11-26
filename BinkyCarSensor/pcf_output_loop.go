package main

import (
	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/pcf8574"
)

type pcfOutput struct {
	DeviceIndex uint8
	Value       uint8
}

// Keep probing sensors
func sendPCF8574Outputs(devices []*pcf8574.Device, outputStatus <-chan pcfOutput) {
	devCnt := uint8(len(devices))
	for {
		select {
		case output := <-outputStatus:
			if output.DeviceIndex < devCnt {
				devices[output.DeviceIndex].WriteBits(output.Value)
			}
		}
	}
}
