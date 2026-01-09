package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Config represents the client configuration
type Config struct {
	ServerURL string
	Token     string
	ComPort   string
	BaudRate  int
}

// Payload matches the server's RemoteScalePayload
type Payload struct {
	Weight float64 `json:"weight"`
	Unit   string  `json:"unit"`
}

func main() {
	// 1. Parse Arguments
	serverURL := flag.String("server", "http://localhost:8080", "Server Base URL")
	token := flag.String("token", "", "Authentication Token (Required)")
	comPort := flag.String("port", "COM1", "Serial Port (e.g., COM1 or /dev/ttyUSB0)")
	baudRate := flag.Int("baud", 9600, "Baud Rate")
	flag.Parse()

	if *token == "" {
		fmt.Println("Usage: scale_sender --token <TOKEN> --port <PORT> --server <URL>")
		log.Fatal("Error: --token is required")
	}

	config := Config{
		ServerURL: strings.TrimRight(*serverURL, "/") + "/api/external/scale",
		Token:     *token,
		ComPort:   *comPort,
		BaudRate:  *baudRate,
	}

	log.Printf("Starting Scale Sender...")
	log.Printf("Server: %s", config.ServerURL)
	log.Printf("Port: %s @ %d", config.ComPort, config.BaudRate)

	// 2. Open Serial Port
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	// Retry loop for connection
	for {
		log.Printf("Connecting to serial port %s...", config.ComPort)
		port, err := serial.Open(config.ComPort, mode)
		if err != nil {
			log.Printf("Failed to open port: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Serial port connected. Starting reading loop...")
		readAndSend(port, config)
		port.Close()

		log.Println("Connection lost. Retrying in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func readAndSend(port serial.Port, config Config) {
	scanner := bufio.NewScanner(port)
	client := &http.Client{Timeout: 2 * time.Second}

	// Buffer to prevent flooding the server?
	// Real-time requirement suggests sending immediately.
	// However, if the scale sends 20 times a second, HTTP overhead might be high.
	// But let's stick to simple first: send every read.

	for scanner.Scan() {
		text := scanner.Text()
		weight := parseWeight(text)

		// Send to Server
		err := sendToServer(client, config, weight)
		if err != nil {
			log.Printf("Failed to send: %v", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Serial read error: %v", err)
	}
}

func sendToServer(client *http.Client, config Config, weight float64) error {
	payload := Payload{Weight: weight, Unit: "kg"}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.ServerURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scale-Token", config.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func parseWeight(raw string) float64 {
	// Simple parser: remove non-numeric chars (except . and -)
	clean := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return r
		}
		return -1
	}, raw)

	if val, err := strconv.ParseFloat(clean, 64); err == nil {
		return val
	}
	return 0.0
}
