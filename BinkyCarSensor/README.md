# Binky Car Sensor

## Introduction

To detected 1:87 scale cars following a magnetic strip, mount SS49E hall-sensors
under the magnetic strip.
Then connect the hall-sensors to an ADS1115 analog-digital converter and
connect that to I2C0 of Waveshare-RP2040-zero flashed with the code in this repo.
The Waveshare-RP2040-zero connects through I2C1 to a Binky Local-worker
where it can be used as a PCF8574 IO expander (in read-only mode).

## Connections

```text
<hall-sensor> -+
...            |
<hall-sensor> -+- <ads1115> --i2c-- <rp2040-zero> --i2c-- <local-worker>
...            |
<hall-sensor> -+
```

## Neopixel color scheme

The Neopixel on the Waveshare-RP2040-zero is used to report local status.
The following colors are used:

- Off: Code not yet initialized
- Orange: Detecting ADS1115 devices
- Green: No active detections, single ADS115 found
- Light-green: No active detections, two ADS1115's found
- Red: No ADS1115 devices found
- Light-Red: More than 2 ADS1115 devices found
- TODO
- 

## Building & flashing

- Press Boot button on Waveshare-RP2040-zero while connecting to USB
  or Press Reset & Boot button at the same time.
  RPi-RP2 volume must now appear.
- Run `make`
