package main

import (
	"fmt"
	"time"

	"github.com/MicahParks/peakdetect"

	"github.com/binkynet/BinkyHardware/BinkyCarSensor/devices/ads1115"
)

// Sensor represents the state of a single hall sensor
type Sensor struct {
	ads        *ads1115.Device
	adsChannel uint8

	window        [windowSize]float64
	curWindowSize int
	active        bool
	initialized   bool
	detector      peakdetect.PeakDetector
}

const (
	probeInterval        = time.Millisecond * 50
	probeAttemptInterval = time.Millisecond * 10
	maxProbeDuration     = time.Millisecond * 500

	// Algorithm configuration from example.
	lag       = 10
	threshold = 7.5
	influence = 0.5 // 0.0..1.0

	// Size of the sliding window
	windowSize = lag * 2
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
		detector:   peakdetect.NewPeakDetector(),
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
		time.Sleep(probeAttemptInterval)
	}
	// Read conversion
	raw, err := s.ads.GetRawConversion()
	if err != nil {
		return fmt.Errorf("GetRawConversion failed: %w", err)
	}
	// Add to value window
	wasActive := s.active
	s.update(raw)

	// Update active flag
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
	value := float64(rawValue) / 10.0
	if s.curWindowSize < windowSize {
		// Add to window
		s.window[s.curWindowSize] = value
		s.curWindowSize++
	} else {
		// Move first (oldest) entry out of window
		copy(s.window[:], s.window[1:])
		s.window[windowSize-1] = value

		// Initialize detector if needed
		if !s.initialized {
			if err := s.detector.Initialize(influence, threshold, s.window[:]); err != nil {
				println("Detected failed to initialise: ", err)
			} else {
				s.initialized = true
			}
		} else {
			switch s.detector.Next(value) {
			case peakdetect.SignalNeutral:
				s.active = false
			default:
				s.active = true
			}
		}
	}
}
