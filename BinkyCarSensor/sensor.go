package main

import (
	"fmt"
	"time"

	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/ads1115"
)

// Sensor represents the state of a single hall sensor
type Sensor struct {
	ads        *ads1115.Device
	adsChannel uint8

	current       uint16
	window        [windowSize]uint16
	curWindowSize int
	active        bool
}

const (
	maxProbeDuration = time.Millisecond * 500

	// Size of the sliding window
	windowSize = 9
	// #elements higher than current must be > windowSize - this
	windowDelta = 3
	// Minimum difference between min & max
	minMinMaxDiff = 25
)

// NewSensor initializes a new sensor
func NewSensor(ads *ads1115.Device, adsChannel uint8) *Sensor {
	return &Sensor{
		ads:        ads,
		adsChannel: adsChannel,
	}
}

// Probe the current status of the sensor
func (s *Sensor) Probe() error {
	// Select channel
	if err := s.ads.SetSingleChannel(s.adsChannel); err != nil {
		return fmt.Errorf("SetSingleChannel failed: %w", err)
	}
	// Start measurement
	if err := s.ads.StartSingleMeasurement(); err != nil {
		return fmt.Errorf("StartSingleMeasurement failed: %w", err)
	}
	// Wait until ready
	start := time.Now()
	for {
		if busy, err := s.ads.IsBusy(); err != nil {
			return fmt.Errorf("IsBusy failed: %w", err)
		} else if !busy {
			// Conversion is ready
			break
		}
		// Check elapsed time
		if time.Since(start) >= maxProbeDuration {
			return fmt.Errorf("Probe timeout")
		}
		time.Sleep(time.Microsecond * 50)
	}
	// Read conversion
	raw, err := s.ads.GetRawConversion()
	if err != nil {
		return fmt.Errorf("GetRawConversion failed: %w", err)
	}
	// Add to value window
	s.update(raw)

	// Update active flag
	wasActive := s.active
	s.active = s.isActive()
	if s.active != wasActive {
		println(s.adsChannel, raw, s.active)
	}

	return nil
}

// IsActive returns true if an active signal is detected.
func (s *Sensor) IsActive() bool {
	return s.active
}

// Update the sliding window with the given most recent probe value
func (s *Sensor) update(rawValue uint16) {
	s.current = rawValue
	if s.curWindowSize < windowSize {
		// Add to window
		s.window[s.curWindowSize] = rawValue
		s.curWindowSize++
	} else {
		// Move first (oldest) entry out of window
		copy(s.window[:], s.window[1:])
		s.window[windowSize-1] = rawValue
	}
}

// Get minimum (non-zero) value from sliding window
func (s *Sensor) min() uint16 {
	result := uint16(0)
	for i := 0; i < s.curWindowSize; i++ {
		x := s.window[i]
		if (x < result) || (result == 0) {
			result = x
		}
	}
	return result
}

// Get maximum value from sliding window
func (s *Sensor) max() uint16 {
	result := uint16(0)
	for i := 0; i < s.curWindowSize; i++ {
		x := s.window[i]
		if x > result {
			result = x
		}
	}
	return result
}

// isActive determines if the given value window show an active signal.
func (s *Sensor) isActive() bool {
	// Difference between min & max must be higher than minMinMaxDiff
	if abs(int(s.min())-int(s.max())) <= minMinMaxDiff {
		return false
	}
	// Count number of window entries higher than current vs lower than current
	lower := 0
	higher := 0
	current := s.current
	for i := 0; i < s.curWindowSize; i++ {
		x := s.window[i]
		if x > current {
			higher++
		}
		if x < current {
			lower++
		}
	}
	return (higher > int(s.curWindowSize)-windowDelta)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
