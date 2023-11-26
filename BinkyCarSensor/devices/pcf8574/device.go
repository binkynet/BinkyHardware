package pcf8574

import (
	"fmt"
	"machine"
)

// Device implements access to an PCF8574 device.
type Device struct {
	i2c        *machine.I2C
	i2cAddress uint8
}

// New initializes a new device attached to given I2C bus.
func New(i2c *machine.I2C, i2cAddress uint8) *Device {
	return &Device{
		i2c:        i2c,
		i2cAddress: i2cAddress,
	}
}

// Reset the device to default configuration
func (dev *Device) Reset() error {
	if err := dev.WriteBits(0); err != nil {
		return fmt.Errorf("WriteBits failed: %w", err)
	}
	return nil
}

// Write 8-bits out binary output
func (dev *Device) WriteBits(value uint8) error {
	w := [1]uint8{value}
	if err := dev.i2c.Tx(uint16(dev.i2cAddress), w[:], nil); err != nil {
		return err
	}
	return nil
}
