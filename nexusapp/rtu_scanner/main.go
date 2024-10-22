package modbus_scanner

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ModbusRTUScanner struct {
	serialPort         string
	baudRate           int
	dataBits           int
	parity             string
	stopBits           int
	slaveId            byte
	timeout            time.Duration
	startRegister      int
	numRegisters       int
	functionCode       int
	writeRegisterEntry int

	resultLabel1 *widget.Label // First result label
	resultLabel2 *widget.Label // Second result label
	resultLabel3 *widget.Label // Third result label
	resultLabel4 *widget.Label

	writeLabel *widget.Label
	errorLabel *widget.Label // Label for error messages
	app        fyne.App
	window     fyne.Window
	scanning   bool // Flag to control scanning
	stopChan   chan struct{}
	spinner    *widget.ProgressBarInfinite
}

func (ms *ModbusRTUScanner) createUI() fyne.CanvasObject {
	// List of COM ports from COM1 to COM20
	comPorts := make([]string, 20)
	for i := 1; i <= 20; i++ {
		comPorts[i-1] = fmt.Sprintf("COM%d", i)
	}

	// COM Port selection dropdown
	portSelect := widget.NewSelect(comPorts, func(s string) {
		ms.serialPort = s
	})
	portSelect.SetSelected("COM1") // Set default to COM1

	baudRateSelect := widget.NewSelect([]string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}, func(s string) {
		baudRate, err := strconv.Atoi(s)
		if err == nil {
			ms.baudRate = baudRate
		}
	})
	baudRateSelect.SetSelected("9600") // Set default value

	dataBitsEntry := widget.NewEntry()
	dataBitsEntry.SetPlaceHolder("Data Bits (e.g., 8)")
	dataBitsEntry.OnChanged = func(s string) {
		dataBits, err := strconv.Atoi(s)
		if err == nil {
			ms.dataBits = dataBits
		}
	}
	parityOptions := map[string]string{
		"None": "N",
		"Even": "E",
		"Odd":  "O",
	}
	// Create a slice with the display names for the dropdown
	parityDisplayOptions := []string{"None", "Even", "Odd"}

	// Parity selection dropdown
	paritySelect := widget.NewSelect(parityDisplayOptions, func(s string) {
		// Retrieve the corresponding single-letter value
		ms.parity = parityOptions[s]
	})
	paritySelect.SetSelected("Even") // Set default to "Even"

	stopBitsEntry := widget.NewEntry()
	stopBitsEntry.SetPlaceHolder("Stop Bits (1 or 2)")
	stopBitsEntry.OnChanged = func(s string) {
		stopBits, err := strconv.Atoi(s)
		if err == nil {
			ms.stopBits = stopBits
		}
	}

	slaveIdEntry := widget.NewEntry()
	slaveIdEntry.SetPlaceHolder("Slave ID (e.g., 1)")
	slaveIdEntry.OnChanged = func(s string) {
		slaveId, err := strconv.Atoi(s)
		if err == nil {
			ms.slaveId = byte(slaveId)
		}
	}

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetPlaceHolder("Timeout (seconds)")
	timeoutEntry.OnChanged = func(s string) {
		timeout, err := strconv.Atoi(s)
		if err == nil {
			ms.timeout = time.Duration(timeout) * time.Second
		}
	}

	startRegisterEntry := widget.NewEntry()
	startRegisterEntry.SetPlaceHolder("Start Register (e.g., 500)")
	startRegisterEntry.OnChanged = func(s string) {
		startRegister, err := strconv.Atoi(s)
		if err == nil {
			ms.startRegister = startRegister
		}
	}

	numRegistersEntry := widget.NewEntry()
	numRegistersEntry.SetPlaceHolder("Number of Registers (e.g., 1)")
	numRegistersEntry.OnChanged = func(s string) {
		numRegisters, err := strconv.Atoi(s)
		if err == nil {
			ms.numRegisters = numRegisters
		}
	}

	functionCodeSelect := widget.NewSelect([]string{"1: Read Coils", "2: Read Discrete Inputs", "3: Read Holding Registers", "4: Read Input Registers"}, func(s string) {
		code, err := strconv.Atoi(s[:1])
		if err == nil {
			ms.functionCode = code
		}
	})
	functionCodeSelect.SetSelected("3: Read Holding Registers") // Set default value

	// Arrange labels above inputs
	portContainer := container.NewVBox(widget.NewLabel("COM Port"), portSelect)
	baudRateContainer := container.NewVBox(widget.NewLabel("Baud Rate"), baudRateSelect)
	dataBitsContainer := container.NewVBox(widget.NewLabel("Data Bits"), dataBitsEntry)
	parityContainer := container.NewVBox(widget.NewLabel("Parity"), paritySelect)
	stopBitsContainer := container.NewVBox(widget.NewLabel("Stop Bits"), stopBitsEntry)
	slaveIdContainer := container.NewVBox(widget.NewLabel("Slave ID"), slaveIdEntry)
	timeoutContainer := container.NewVBox(widget.NewLabel("Timeout"), timeoutEntry)
	startRegisterContainer := container.NewVBox(widget.NewLabel("Start Register"), startRegisterEntry)
	numRegistersContainer := container.NewVBox(widget.NewLabel("Number of Registers"), numRegistersEntry)
	functionCodeContainer := container.NewVBox(widget.NewLabel("Function Code"), functionCodeSelect)

	// Use Grid layout for better alignment
	inputGrid := container.NewGridWithColumns(3,
		portContainer,
		baudRateContainer,
		dataBitsContainer,
		parityContainer,
		stopBitsContainer,
		slaveIdContainer,
		timeoutContainer,
	)

	// Use Grid layout for better alignment
	secondGrid := container.NewGridWithColumns(3,
		functionCodeContainer,
		startRegisterContainer,
		numRegistersContainer,
	)

	ms.resultLabel1 = widget.NewLabel("Result: ")
	ms.resultLabel2 = widget.NewLabel("")
	ms.resultLabel3 = widget.NewLabel("")
	ms.resultLabel4 = widget.NewLabel("")
	ms.errorLabel = widget.NewLabel("") // Label to display error messages
	ms.writeLabel = widget.NewLabel("")
	ms.spinner = widget.NewProgressBarInfinite()
	ms.spinner.Hide() // Initially hidden until scanning starts

	startButton := widget.NewButtonWithIcon("Start Scan", theme.MediaPlayIcon(), func() {
		if !ms.scanning {
			ms.startScan()
		}
	})

	stopButton := widget.NewButtonWithIcon("Stop Scan", theme.MediaStopIcon(), func() {
		if ms.scanning {
			ms.stopScan()
		}
	})

	// Add new input for writing a register value
	writeRegisterLabel := widget.NewLabel("Write Register")

	writeRegisterEntry := widget.NewEntry()
	writeRegisterEntry.SetPlaceHolder("Address")

	writeValueEntry := widget.NewEntry()
	writeValueEntry.SetPlaceHolder("Value to write (e.g., 1234)")

	// Button to trigger the write operation
	writeButton := widget.NewButtonWithIcon("Write Register", theme.ConfirmIcon(), func() {
		wasScanning := ms.scanning // Store the current scanning state
		if wasScanning {
			// Stop the scan temporarily
			ms.stopScan()
			time.Sleep(500 * time.Millisecond)
		}

		// Parse the register address
		register, err := strconv.Atoi(writeRegisterEntry.Text)
		if err != nil {
			ms.errorLabel.SetText("Invalid register address: " + writeRegisterEntry.Text)
			return
		}

		// Parse the value to write
		value, err := strconv.Atoi(writeValueEntry.Text)
		if err != nil {
			ms.errorLabel.SetText("Invalid value: " + writeValueEntry.Text)
			return
		}

		// Call writeRegister with register and value
		ms.writeRegister(register, value)

		// Restart scanning if it was scanning before
		if wasScanning {
			ms.startScan()
		}
	})

	// Use Grid layout for better alignment
	inputGridWrite := container.NewGridWithColumns(3,
		writeRegisterEntry,
		writeValueEntry,
		writeButton,
	)

	// Container to hold the result labels side by side
	resultContainer := container.NewHBox(ms.resultLabel1, ms.resultLabel2, ms.resultLabel3, ms.resultLabel4)

	return container.NewVBox(
		widget.NewLabel("Modbus RTU Scanner"),
		inputGrid,
		secondGrid,
		container.NewHBox(startButton, stopButton), // Start and Stop buttons side by side
		writeRegisterLabel,                         // Input for writing value
		inputGridWrite,                             // Write button
		ms.spinner,
		resultContainer,
		ms.errorLabel, // Display error messages below the valid output
		ms.writeLabel,
	)
}

// Show initializes the ModbusRTUScanner and loads its UI into the given window
func Show(win fyne.Window) fyne.CanvasObject {
	scanner := &ModbusRTUScanner{
		window:        win,
		baudRate:      9600,            // Default Baud Rate
		dataBits:      8,               // Default Data Bits
		parity:        "E",             // Default Parity
		stopBits:      1,               // Default Stop Bits
		slaveId:       2,               // Default Slave ID
		timeout:       1 * time.Second, // Default Timeout
		startRegister: 500,             // Default Start Register
		numRegisters:  1,               // Default Number of Registers
		functionCode:  3,               // Default function code (Read Holding Registers)
		scanning:      false,           // Initial scanning state
	}
	return scanner.createUI()
}
