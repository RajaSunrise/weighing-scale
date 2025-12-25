package serial

import (
	"log"

	"go.bug.st/serial"
)

func ReadScale(portName string) (float64, error) {
	// Example implementation
	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		// Just log error for now as we don't have real hardware
		log.Printf("Error opening serial port: %v", err)
		return 0, err
	}
	defer port.Close()

	buff := make([]byte, 100)
	n, err := port.Read(buff)
	if err != nil {
		return 0, err
	}

	// Parse buffer to float (dummy logic)
	log.Printf("Read %d bytes: %v", n, buff[:n])
	return 1000.0, nil
}
