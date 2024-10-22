package nexus_modbus_bits

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/goburrow/modbus"
)

// Create a struct to hold the Modbus bits editor state
type ModbusBitsEditor struct {
	serialPort   string
	baudRate     int
	dataBits     int
	parity       string
	stopBits     int
	slaveId      byte
	registerAddr uint16
	bitToggles   []*widget.Check
	readButton   *widget.Button
	writeButton  *widget.Button
	statusLabel  *widget.Label // Add a label for status messages
}

// Initialize Modbus client and set up the editor UI
func Show(w fyne.Window) fyne.CanvasObject {
	editor := &ModbusBitsEditor{} // Create an instance of ModbusBitsEditor

	// List of COM ports from COM1 to COM20
	comPorts := make([]string, 20)
	for i := 1; i <= 20; i++ {
		comPorts[i-1] = fmt.Sprintf("COM%d", i)
	}

	// COM Port selection dropdown
	portSelect := widget.NewSelect(comPorts, func(s string) {
		editor.serialPort = s
	})
	portSelect.SetSelected("COM1") // Set default to COM1

	baudRateSelect := widget.NewSelect([]string{"1200", "2400", "4800", "9600", "19200", "38400", "57600", "115200"}, func(s string) {
		baudRate, err := strconv.Atoi(s)
		if err == nil {
			editor.baudRate = baudRate
		}
	})
	baudRateSelect.SetSelected("9600") // Set default value

	dataBitsEntry := widget.NewEntry()
	dataBitsEntry.SetPlaceHolder("Data Bits (e.g., 8)")
	dataBitsEntry.OnChanged = func(s string) {
		dataBits, err := strconv.Atoi(s)
		if err == nil {
			editor.dataBits = dataBits
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
		editor.parity = parityOptions[s]
	})
	paritySelect.SetSelected("Even") // Set default to "Even"

	stopBitsEntry := widget.NewEntry()
	stopBitsEntry.SetPlaceHolder("Stop Bits (1 or 2)")
	stopBitsEntry.OnChanged = func(s string) {
		stopBits, err := strconv.Atoi(s)
		if err == nil {
			editor.stopBits = stopBits
		}
	}

	slaveIdEntry := widget.NewEntry()
	slaveIdEntry.SetPlaceHolder("Slave ID (e.g., 1)")
	slaveIdEntry.OnChanged = func(s string) {
		slaveId, err := strconv.Atoi(s)
		if err == nil {
			editor.slaveId = byte(slaveId)
		}
	}

	// Create the status label
	editor.statusLabel = widget.NewLabel("")

	// Create UI elements for address input
	registerAddrInput := widget.NewEntry()
	registerAddrInput.SetPlaceHolder("Register address (e.g., 100)")

	// Create the bit toggle checkboxes (16 bits for a Modbus register)
	for i := 0; i < 16; i++ {
		check := widget.NewCheck(fmt.Sprintf("Bit %d", i), nil)
		editor.bitToggles = append(editor.bitToggles, check)
	}

	// Create the read button
	editor.readButton = widget.NewButton("Read Register", func() {
		addr, err := strconv.Atoi(registerAddrInput.Text)
		if err != nil {
			editor.statusLabel.SetText("Error: Invalid register address")
			return
		}
		editor.registerAddr = uint16(addr)
		editor.readRegister()
	})

	// Create the write button
	editor.writeButton = widget.NewButton("Write Register", func() {
		editor.writeRegister()
	})

	action_buttons := container.NewGridWithColumns(2,
		editor.readButton,
		editor.writeButton,
	)

	portContainer := container.NewVBox(widget.NewLabel("COM Port"), portSelect)
	baudRateContainer := container.NewVBox(widget.NewLabel("Baud Rate"), baudRateSelect)
	dataBitsContainer := container.NewVBox(widget.NewLabel("Data Bits"), dataBitsEntry)
	parityContainer := container.NewVBox(widget.NewLabel("Parity"), paritySelect)
	stopBitsContainer := container.NewVBox(widget.NewLabel("Stop Bits"), stopBitsEntry)
	slaveIdContainer := container.NewVBox(widget.NewLabel("Slave ID"), slaveIdEntry)
	addressContainer := container.NewVBox(widget.NewLabel("Bits Register"), registerAddrInput)
	// Use Grid layout for better alignment
	inputGrid := container.NewGridWithColumns(3,
		portContainer,
		baudRateContainer,
		dataBitsContainer,
		parityContainer,
		stopBitsContainer,
		slaveIdContainer,
		addressContainer,
	)

	// Organize UI elements in a grid layout
	bitToggleGrid := container.NewGridWithColumns(4, convertToCanvasObjects(editor.bitToggles)...)

	content := container.NewVBox(
		widget.NewLabel("Modbus Bits Editor"),
		inputGrid,
		action_buttons,
		bitToggleGrid,
		layout.NewSpacer(),
		editor.statusLabel, // Add the status label to the UI
	)

	return content
}

// Connect to Modbus device
func (e *ModbusBitsEditor) connect() (*modbus.RTUClientHandler, error) {
	handler := modbus.NewRTUClientHandler(e.serialPort)
	handler.BaudRate = e.baudRate
	handler.Parity = e.parity
	handler.DataBits = e.dataBits
	handler.StopBits = e.stopBits
	handler.SlaveId = e.slaveId

	if err := handler.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to Modbus device: %v", err)
	}
	return handler, nil
}

// Read the selected register, extract bit values, and update the UI checkboxes
func (e *ModbusBitsEditor) readRegister() {
	handler, err := e.connect()
	if err != nil {
		e.statusLabel.SetText(err.Error())
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	results, err := client.ReadHoldingRegisters(e.registerAddr, 1)
	if err != nil {
		e.statusLabel.SetText(fmt.Sprintf("Error: Failed to read register: %v", err))
		return
	}

	if len(results) < 2 {
		e.statusLabel.SetText("Error: Invalid result length")
		return
	}

	// Extract the 16 bits from the register value
	value := uint16(results[0])<<8 | uint16(results[1])
	for i := 0; i < 16; i++ {
		bit := (value >> i) & 1
		e.bitToggles[i].SetChecked(bit == 1)
	}
	e.statusLabel.SetText("Read successful!") // Update status after successful read
}

// Convert toggle checkboxes to canvas objects
func convertToCanvasObjects(toggles []*widget.Check) []fyne.CanvasObject {
	canvasObjects := make([]fyne.CanvasObject, len(toggles))
	for i, toggle := range toggles {
		canvasObjects[i] = toggle
	}
	return canvasObjects
}

// Write updated bits to register
func (e *ModbusBitsEditor) writeRegister() {
	// Create a 16-bit value from the bit states
	var value uint16
	for i := 0; i < 16; i++ {
		if e.bitToggles[i].Checked {
			value |= (1 << i)
		}
	}

	handler, err := e.connect()
	if err != nil {
		e.statusLabel.SetText(err.Error())
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	// Write the new value to the register
	_, err = client.WriteSingleRegister(e.registerAddr, value)
	if err != nil {
		e.statusLabel.SetText(fmt.Sprintf("Error: Failed to write to register: %v", err))
		return
	}
	e.statusLabel.SetText("Write successful!") // Update status after successful write
}
