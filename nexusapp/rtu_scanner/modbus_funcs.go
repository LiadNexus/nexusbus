package modbus_scanner

import (
	"fmt"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

func (ms *ModbusRTUScanner) startScan() {
	ms.scanning = true
	ms.stopChan = make(chan struct{})
	ms.spinner.Show() // Show the spinner when scanning starts

	go func() {
		for {
			select {
			case <-ms.stopChan:
				ms.scanning = false
				ms.spinner.Hide() // Hide the spinner when scanning stops
				return
			default:
				ms.scan()
				time.Sleep(2 * time.Second) // Adjust the scanning interval as needed
			}
		}
	}()
}

func (ms *ModbusRTUScanner) stopScan() {
	if ms.scanning {
		ms.stopChan <- struct{}{}
		ms.scanning = false
	}
}

func (ms *ModbusRTUScanner) scan() {
	handler := modbus.NewRTUClientHandler(ms.serialPort)
	handler.BaudRate = ms.baudRate
	handler.DataBits = ms.dataBits
	handler.Parity = ms.parity
	handler.StopBits = ms.stopBits
	handler.SlaveId = ms.slaveId
	handler.Timeout = ms.timeout

	client := modbus.NewClient(handler)
	defer handler.Close()

	err := handler.Connect()
	if err != nil {
		ms.errorLabel.SetText("Failed to connect: " + err.Error())
		return
	}

	var results []byte
	switch ms.functionCode {
	case 1:
		results, err = client.ReadCoils(uint16(ms.startRegister), uint16(ms.numRegisters))
	case 2:
		results, err = client.ReadDiscreteInputs(uint16(ms.startRegister), uint16(ms.numRegisters))
	case 3:
		results, err = client.ReadHoldingRegisters(uint16(ms.startRegister), uint16(ms.numRegisters))
	case 4:
		results, err = client.ReadInputRegisters(uint16(ms.startRegister), uint16(ms.numRegisters))
	default:
		ms.errorLabel.SetText("Invalid function code")
		return
	}

	if err != nil {
		ms.errorLabel.SetText("Read error: " + err.Error())
		return
	}

	// Clear the error label on a successful read
	ms.errorLabel.SetText("")

	const maxPerLabel = 20 // Number of values per result label

	var resultLines []string
	for i := 0; i < ms.numRegisters; i++ {
		var line string
		if ms.functionCode == 1 || ms.functionCode == 2 {
			if i < len(results) && results[i] == 1 {
				line = fmt.Sprintf("Register %d: ON", ms.startRegister+i)
			} else {
				line = fmt.Sprintf("Register %d: OFF", ms.startRegister+i)
			}
		} else {
			if len(results) >= 2*(i+1) {
				value := int(results[2*i])<<8 | int(results[2*i+1])
				line = fmt.Sprintf("Register %d: %d", ms.startRegister+i, value)
			} else {
				line = fmt.Sprintf("Register %d: Error", ms.startRegister+i)
			}
		}
		resultLines = append(resultLines, line)
	}

	// Split resultLines into four parts for each label
	var label1Lines, label2Lines, label3Lines, label4Lines []string
	for i, line := range resultLines {
		if i < maxPerLabel {
			label1Lines = append(label1Lines, line)
		} else if i < 2*maxPerLabel {
			label2Lines = append(label2Lines, line)
		} else if i < 3*maxPerLabel {
			label3Lines = append(label3Lines, line)
		} else {
			label4Lines = append(label4Lines, line)
		}
	}

	// Set the text for each result label
	ms.resultLabel1.SetText(strings.Join(label1Lines, "\n"))
	ms.resultLabel2.SetText(strings.Join(label2Lines, "\n"))
	ms.resultLabel3.SetText(strings.Join(label3Lines, "\n"))
	ms.resultLabel4.SetText(strings.Join(label4Lines, "\n"))
}

func (ms *ModbusRTUScanner) writeRegister(register int, value int) {
	handler := modbus.NewRTUClientHandler(ms.serialPort)
	handler.BaudRate = ms.baudRate
	handler.DataBits = ms.dataBits
	handler.Parity = ms.parity
	handler.StopBits = ms.stopBits
	handler.SlaveId = ms.slaveId
	handler.Timeout = ms.timeout

	client := modbus.NewClient(handler)
	defer handler.Close()

	err := handler.Connect()
	if err != nil {
		ms.writeLabel.SetText("Failed to connect: " + err.Error())
		return
	}

	// Write a single holding register
	_, err = client.WriteSingleRegister(uint16(register), uint16(value))
	if err != nil {
		ms.writeLabel.SetText("Write error: " + err.Error())
		return
	}

	ms.writeLabel.SetText(fmt.Sprintf("Write successful! (Value %d to register %d)", value, register))
	go func() {
		// Wait for 5 seconds before clearing the label
		time.Sleep(5 * time.Second)
		// Ensure the UI update happens on the main thread
		ms.writeLabel.SetText("")
	}()

}
