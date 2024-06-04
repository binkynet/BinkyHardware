package main

import (
	"fmt"
	"machine"
	"time"
)

var (
	// Current firward version
	version = []byte{0, 1, 0} // Major.Minor.Patch
)

const (
	// Register addresses
	RegVersionMajor   = 0x00 // No input, returns 1 version
	RegVersionMinor   = 0x01 // No input, returns 1 version
	RegVersionPatch   = 0x02 // No input, returns 1 version
	RegCarSensorCount = 0x03 // No input, returns 1 byte giving the number of detected car sensor bits (0..8)
	RegI2COutputCount = 0x04 // No input, returns 1 byte giving the number of detected I2C binary output pins (0, 8, 16, ..., 256)
	RegCarSensorState = 0x10 // No input, returns 1 byte with 8-bit car detection sensor state
	RegOutput         = 0x20 // 1 byte input, targeting 8 on-pcb output pins
	RegOutputI2C0     = 0x21 // 1 byte input, targeting 8 output pins on PCF8574 output device 0
	RegOutputI2C1     = 0x22 // 1 byte input, targeting 8 output pins on PCF8574 output device 1
	RegOutputI2C2     = 0x23 // 1 byte input, targeting 8 output pins on PCF8574 output device 2
	RegOutputI2C3     = 0x24 // 1 byte input, targeting 8 output pins on PCF8574 output device 3
	RegOutputI2C4     = 0x25 // 1 byte input, targeting 8 output pins on PCF8574 output device 4
	RegOutputI2C5     = 0x26 // 1 byte input, targeting 8 output pins on PCF8574 output device 5
	RegOutputI2C6     = 0x27 // 1 byte input, targeting 8 output pins on PCF8574 output device 6
	RegOutputI2C7     = 0x28 // 1 byte input, targeting 8 output pins on PCF8574 output device 7
	RegConfigurePWM0  = 0x30 // 1 byte input, pwm-value (0-256) of pin 0
	RegConfigurePWM1  = 0x31 // 1 byte input, pwm-value (0-256) of pin 1
	RegConfigurePWM2  = 0x32 // 1 byte input, pwm-value (0-256) of pin 2
	RegConfigurePWM3  = 0x33 // 1 byte input, pwm-value (0-256) of pin 3
	RegConfigurePWM4  = 0x34 // 1 byte input, pwm-value (0-256) of pin 4
	RegConfigurePWM5  = 0x35 // 1 byte input, pwm-value (0-256) of pin 5
	RegConfigurePWM6  = 0x36 // 1 byte input, pwm-value (0-256) of pin 6
	RegConfigurePWM7  = 0x37 // 1 byte input, pwm-value (0-256) of pin 7

	pwmPeriod = uint64(1e9) / 60
)

// Single i2c message sent to the incoming i2c port
type incomingI2CEvent struct {
	Event       machine.I2CTargetEvent
	HasRegister bool
	Register    uint8
	HasValue    bool
	Value       uint8
}

// Listen for incoming I2C requests.
func listenForIncomingI2CRequests(i2c *machine.I2C, i2cAddress uint8,
	carSensorStateChanges <-chan uint8, outputStatus chan pcfOutput,
	carSensorBitsCount uint8, i2cOutputBitsCount uint8) error {
	// Configure i2c bus as target
	if err := i2c.Configure(machine.I2CConfig{
		Mode: machine.I2CModeTarget,
	}); err != nil {
		return fmt.Errorf("Failed to configure i2c bus: %w", err)
	}

	// Start listening on the i2c bus
	if err := i2c.Listen(uint16(i2cAddress)); err != nil {
		return fmt.Errorf("Failed to listen on i2c bus: %w", err)
	}
	println("Listening on i2c address: ", i2cAddress)

	// Process events & status changes
	events := make(chan incomingI2CEvent)
	go func() {
		lastOutputVals := make([]uint8, 9)
		lastRequestReq := uint8(0)
		isPWM := make([]bool, 8)
		pwmValues := make([]uint16, 0xffff)
		var responseBuf [1]uint8
		var lastSensorStatus uint8
		for {
			select {
			case x := <-carSensorStateChanges:
				if x != lastSensorStatus {
					println("Update sensor status: ", x)
					lastSensorStatus = x
					responseBuf[0] |= x
				}
			case evt := <-events:
				// Handle event
				switch evt.Event {
				case machine.I2CReceive:
					if evt.Register >= RegOutput && evt.Register < RegOutputI2C7 {
						outputIndex := evt.Register - RegOutput
						if lastOutputVals[outputIndex] != evt.Value {
							println("I2C:Receive Output ", outputIndex, evt.Value)
							lastOutputVals[outputIndex] = evt.Value
						}
					}
					switch evt.Register {
					case RegOutput:
						if evt.HasValue {
							// Since we pull IO1 down to use alternate i2c address,
							// we do not allow setting it high when using the alternate address.
							if !isPWM[0] {
								setIOx(IO[0], evt.Value&0x01 != 0 && i2cAddress == defaultI2cAddress)
							}
							if !isPWM[1] {
								setIOx(IO[1], evt.Value&0x02 != 0)
							}
							if !isPWM[2] {
								setIOx(IO[2], evt.Value&0x04 != 0)
							}
							if !isPWM[3] {
								setIOx(IO[3], evt.Value&0x08 != 0)
							}
							if !isPWM[4] {
								setIOx(IO[4], evt.Value&0x10 != 0)
							}
							if !isPWM[5] {
								setIOx(IO[5], evt.Value&0x20 != 0)
							}
							if !isPWM[6] {
								setIOx(IO[6], evt.Value&0x40 != 0)
							}
							if !isPWM[7] {
								setIOx(IO[7], evt.Value&0x80 != 0)
							}
						}
					case RegOutputI2C0, RegOutputI2C1, RegOutputI2C2, RegOutputI2C3, RegOutputI2C4, RegOutputI2C5, RegOutputI2C6, RegOutputI2C7:
						output := pcfOutput{
							DeviceIndex: evt.Register - RegOutputI2C0,
							Value:       evt.Value,
						}
						select {
						case outputStatus <- output:
							// We're done
						case <-time.After(time.Millisecond * 100):
							// We did not send the bit in time
							println("Failed to send PCF output in time: ", output.Value, "->", output.DeviceIndex)
						}
					case RegCarSensorState:
						// Ignore
					case RegConfigurePWM0, RegConfigurePWM1, RegConfigurePWM2, RegConfigurePWM3, RegConfigurePWM4, RegConfigurePWM5, RegConfigurePWM6, RegConfigurePWM7:
						ioIndex := evt.Register - RegConfigurePWM0
						value := evt.Value
						if ioIndex < 8 {
							isPWM[ioIndex] = true
						}
						if value != uint8(pwmValues[ioIndex]) {
							setPWM(ioIndex, value)
							pwmValues[ioIndex] = uint16(value)
						}
					default:
						println("I2C:Receive: Invalid register ", evt.Register, evt.HasValue, evt.Value)
					}
				case machine.I2CRequest:
					// Reply with current state of sensors
					if lastRequestReq != evt.Register {
						println("I2C:Request ", evt.Register, "->", responseBuf[0])
						lastRequestReq = evt.Register
					}
					switch evt.Register {
					case RegVersionMajor:
						i2c.Reply(version[0:1])
					case RegVersionMinor:
						i2c.Reply(version[1:2])
					case RegVersionPatch:
						i2c.Reply(version[2:3])
					case RegCarSensorCount:
						i2c.Reply([]byte{carSensorBitsCount})
					case RegI2COutputCount:
						i2c.Reply([]byte{i2cOutputBitsCount})
					case RegCarSensorState:
						i2c.Reply(responseBuf[:])
						// Reset detections
						responseBuf[0] = lastSensorStatus
					default:
						i2c.Reply([]byte{0xff, 0xff})
					}
				case machine.I2CFinish:
					// No response needed
				}
			}
		}
	}()
	var buf [8]uint8
	for {
		// Wait for event
		evt, count, err := i2c.WaitForEvent(buf[:])
		if err != nil {
			return fmt.Errorf("Failed to wait for event: %w", err)
		}

		// Handle event
		events <- incomingI2CEvent{
			Event:       evt,
			HasRegister: count >= 1,
			Register:    buf[0],
			HasValue:    count >= 2,
			Value:       buf[1],
		}
	}
}

// Set an IO bit
func setIOx(io machine.Pin, value bool) {
	if value {
		io.Configure(machine.PinConfig{Mode: machine.PinOutput})
		io.High()
	} else {
		io.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}
}

// Set a PWM output
func setPWM(ioIndex, value byte) {
	println("setPWM", ioIndex, " -> ", value)
	io := IO[ioIndex]
	slice, _ := machine.PWMPeripheral(io)
	pwm := PWMBySlice[slice]
	channel, _ := pwm.Channel(io)
	targetValue := uint32(0)
	if value > 0 {
		fraction := float64(value) / 256
		targetValue = uint32((float64(pwm.Top()) * fraction))
	}
	println("... slice=", slice, " channel=", channel, "->", targetValue)
	pwm.Enable(false)
	pwm.SetPeriod(pwmPeriod) // period of 20ms
	pwm.Set(channel, targetValue)
	pwm.Enable(true)
}
