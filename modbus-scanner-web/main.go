package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

// listPorts returns a list of serial ports on both Windows and Linux/macOS.
func listPorts() []string {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Use PowerShell to list COM ports on Windows
		cmd = exec.Command("powershell", "Get-WmiObject Win32_SerialPort | Select-Object -ExpandProperty DeviceID")
	} else {
		// Use shell command to list /dev/tty* devices on Linux/macOS
		cmd = exec.Command("sh", "-c", "ls /dev/tty*")
	}

	output, err := cmd.Output()
	if err != nil {
		return []string{"Error detecting ports"}
	}

	ports := strings.Split(strings.TrimSpace(string(output)), "\n")
	return ports
}

func portsHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(listPorts())
}

type ModbusConfig struct {
	ComPort       string `json:"comPort"`
	BaudRate      int    `json:"baudRate"`
	Parity        string `json:"parity"`
	SlaveId       byte   `json:"slaveId"`
	StartRegister uint16 `json:"startRegister"`
	NumRegisters  uint16 `json:"numRegisters"`
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	var config ModbusConfig
	json.NewDecoder(r.Body).Decode(&config)

	handler := modbus.NewRTUClientHandler(config.ComPort)
	handler.BaudRate = config.BaudRate
	handler.Parity = config.Parity
	handler.SlaveId = config.SlaveId
	handler.Timeout = 2 * time.Second

	client := modbus.NewClient(handler)
	handler.Connect()
	defer handler.Close()

	results, _ := client.ReadHoldingRegisters(config.StartRegister, config.NumRegisters)
	json.NewEncoder(w).Encode(results)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/ports", portsHandler)
	http.HandleFunc("/api/scan", scanHandler)

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
