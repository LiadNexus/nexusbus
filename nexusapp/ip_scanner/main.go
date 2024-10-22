package modbus_scanner

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/goburrow/modbus"
)

type ModbusScanner struct {
	ipAddress   string
	port        int
	register    int
	resultLabel *widget.Label
	window      fyne.Window
}

func (ms *ModbusScanner) scan() {
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", ms.ipAddress, ms.port))
	client := modbus.NewClient(handler)
	defer handler.Close()

	handler.Timeout = 5 * time.Second
	err := handler.Connect()
	if err != nil {
		ms.resultLabel.SetText("Failed to connect: " + err.Error())
		return
	}

	results, err := client.ReadHoldingRegisters(uint16(ms.register), 1)
	if err != nil {
		ms.resultLabel.SetText("Read error: " + err.Error())
		return
	}

	ms.resultLabel.SetText(fmt.Sprintf("Register %d: %v", ms.register, results))
}

func (ms *ModbusScanner) createUI() fyne.CanvasObject {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter IP Address")
	ipEntry.OnChanged = func(s string) {
		ms.ipAddress = s
	}

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Enter Port (e.g., 502)")
	portEntry.OnChanged = func(s string) {
		port, err := strconv.Atoi(s)
		if err == nil {
			ms.port = port
		}
	}

	registerEntry := widget.NewEntry()
	registerEntry.SetPlaceHolder("Enter Register Address (e.g., 1)")
	registerEntry.OnChanged = func(s string) {
		register, err := strconv.Atoi(s)
		if err == nil {
			ms.register = register
		}
	}

	ms.resultLabel = widget.NewLabel("Result: ")

	scanButton := widget.NewButton("Scan", func() {
		ms.scan()
	})

	return container.NewVBox(
		widget.NewLabel("Modbus Scanner"),
		ipEntry,
		portEntry,
		registerEntry,
		scanButton,
		ms.resultLabel,
	)
}

// Show initializes the ModbusScanner and loads its UI into the given window
func Show(win fyne.Window) fyne.CanvasObject {
	scanner := &ModbusScanner{window: win}
	return scanner.createUI()
}
