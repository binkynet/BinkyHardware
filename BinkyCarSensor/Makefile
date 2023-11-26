PORT=/dev/cu.usbmodem1101

flash:
	tinygo flash -target=waveshare-rp2040-zero .

run:
	tinygo flash -target=waveshare-rp2040-zero -monitor -serial usb -port $(PORT) .

monitor:
	tinygo monitor -serial usb -port $(PORT)
