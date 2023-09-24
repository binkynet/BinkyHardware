package ads1115

import (
	"fmt"
	"machine"
)

// Device implements access to an ADS1115 device.
type Device struct {
	i2c        *machine.I2C
	i2cAddress uint8
}

const (
	// ADS1115 I2C addresses
	I2CAddressGround = 0b1001000
	I2CAddressVDD    = 0b1001001
	I2CAddressSDA    = 0b1001010
	I2CAddressSCL    = 0b1001011

	// ADS1115 registers
	regConversion  = 0x00
	regConfig      = 0x01
	regLoThreshold = 0x02
	regHiThreshold = 0x03

	// ADS1115_MUX
	ADS1115_COMP_0_1   uint16 = 0x0000
	ADS1115_COMP_0_3   uint16 = 0x1000
	ADS1115_COMP_1_3   uint16 = 0x2000
	ADS1115_COMP_2_3   uint16 = 0x3000
	ADS1115_COMP_0_GND uint16 = 0x4000
	ADS1115_COMP_1_GND uint16 = 0x5000
	ADS1115_COMP_2_GND uint16 = 0x6000
	ADS1115_COMP_3_GND uint16 = 0x7000

	ADS1115_COMP_INC uint16 = 0x1000 // increment to next channel

	ADS1115_RANGE_6144 uint16 = 0x0000
	ADS1115_RANGE_4096 uint16 = 0x0200
	ADS1115_RANGE_2048 uint16 = 0x0400
	ADS1115_RANGE_1024 uint16 = 0x0600
	ADS1115_RANGE_0512 uint16 = 0x0800
	ADS1115_RANGE_0256 uint16 = 0x0A00
	rangeMask          uint16 = 0b11110001_11111111

	configDefault uint16 = 0x8583
	configOSBit   uint16 = 0x8000
)

// New initializes a new device attached to given I2C bus.
func New(i2c *machine.I2C, i2cAddress uint8) *Device {
	return &Device{
		i2c:        i2c,
		i2cAddress: i2cAddress,
	}
}

// Reset the device to default configuration
func (dev *Device) Reset() error {
	if err := dev.writeRegister(regConfig, configDefault); err != nil {
		return fmt.Errorf("writeRegister failed: %w", err)
	}
	return nil
}

// Set the voltage range of the ADC to adjust the gain:
// * Please note that you must not apply more than VDD + 0.3V to the input pins!
// ADS1115_RANGE_6144  ->  +/- 6144 mV
// ADS1115_RANGE_4096  ->  +/- 4096 mV
// ADS1115_RANGE_2048  ->  +/- 2048 mV (default)
// ADS1115_RANGE_1024  ->  +/- 1024 mV
// ADS1115_RANGE_0512  ->  +/- 512 mV
// ADS1115_RANGE_0256  ->  +/- 256 mV
func (dev *Device) SetVoltageRangeMilliV(r uint16) error {
	currentConfReg, err := dev.readRegister(regConfig)
	if err != nil {
		return fmt.Errorf("readRegister failed: %w", err)
	}
	currentConfReg &= rangeMask
	currentConfReg |= r
	if err := dev.writeRegister(regConfig, currentConfReg); err != nil {
		return fmt.Errorf("writeRegister failed: %w", err)
	}
	return nil
}

// Select a channel to measure from (0-3)
func (dev *Device) SetSingleChannel(channel uint8) error {
	currentConfReg, err := dev.readRegister(regConfig)
	if err != nil {
		return fmt.Errorf("readRegister failed: %w", err)
	}
	currentConfReg &= 0x0fff
	currentConfReg |= ADS1115_COMP_0_GND + ADS1115_COMP_INC*uint16(channel)
	if err := dev.writeRegister(regConfig, currentConfReg); err != nil {
		return fmt.Errorf("writeRegister failed: %w", err)
	}
	return nil
}

// Start a single measurement on the current channel.
func (dev *Device) StartSingleMeasurement() error {
	currentConfReg, err := dev.readRegister(regConfig)
	if err != nil {
		return fmt.Errorf("readRegister failed: %w", err)
	}
	currentConfReg |= configOSBit
	if err := dev.writeRegister(regConfig, currentConfReg); err != nil {
		return fmt.Errorf("writeRegister failed: %w", err)
	}
	return nil
}

// Returns true if a conversion is ongoing.
func (dev *Device) IsBusy() (bool, error) {
	currentConfReg, err := dev.readRegister(regConfig)
	if err != nil {
		return false, fmt.Errorf("readRegister failed: %w", err)
	}
	busy := (currentConfReg & configOSBit) == 0
	return busy, nil
}

// Gets the crrent conversion value in raw format
func (dev *Device) GetRawConversion() (uint16, error) {
	result, err := dev.readRegister(regConversion)
	if err != nil {
		return 0, fmt.Errorf("readRegister failed: %w", err)
	}
	return result, nil
}

// Read a 16-bit register
func (dev *Device) readRegister(reg uint8) (uint16, error) {
	w := [1]uint8{reg}
	var r [2]uint8
	if err := dev.i2c.Tx(uint16(dev.i2cAddress), w[:], r[:]); err != nil {
		return 0, err
	}
	result := (uint16(r[0]) << 8) | uint16(r[1]) // MSB first
	return result, nil
}

// Write a 16-bit register
func (dev *Device) writeRegister(reg uint8, value uint16) error {
	w := [3]uint8{reg, uint8((value >> 8) & 0xff), uint8(value & 0xff)}
	if err := dev.i2c.Tx(uint16(dev.i2cAddress), w[:], nil); err != nil {
		return err
	}
	return nil
}
