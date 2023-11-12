package main

import (
	"fmt"
	"machine"
)

// Listen for PCF8574 compatible requests on the given i2c
// bus with the given address.
func listenForPCF8574Requests(i2c *machine.I2C, i2cAddress uint8, sensorStatus <-chan uint8) error {
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
	events := make(chan machine.I2CTargetEvent)
	go func() {
		for {
			var responseBuf [1]uint8
			select {
			case x := <-sensorStatus:
				responseBuf[0] = x
			case evt := <-events:
				// Handle event
				switch evt {
				case machine.I2CReceive:
					// We can simply ignore these
				case machine.I2CRequest:
					// Reply with current state of sensors
					i2c.Reply(responseBuf[:])
				case machine.I2CFinish:
					// No response needed
				}
			}
		}
	}()
	var buf [8]uint8
	for {
		// Wait for event
		evt, _, err := i2c.WaitForEvent(buf[:])
		if err != nil {
			return fmt.Errorf("Failed to wait for event: %w", err)
		}

		// Handle event
		events <- evt
	}
}
