package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goburrow/modbus"
)

type ModbusRequest struct {
	Register int `json:"register"`
	Value    int `json:"value"`
}

// ScanRegisters handles reading Modbus registers
func ScanRegisters(c *gin.Context) {
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 2 * time.Second

	if err := handler.Connect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	results, err := client.ReadHoldingRegisters(0, 10) // Read 10 registers starting at 0
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := make(map[string]int)
	for i := 0; i < len(results)/2; i++ {
		value := int(results[2*i])<<8 | int(results[2*i+1])
		data["Register "+strconv.Itoa(i)] = value
	}

	c.JSON(http.StatusOK, data)
}

// WriteRegister handles writing a value to a Modbus register
func WriteRegister(c *gin.Context) {
	var req ModbusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 2 * time.Second

	if err := handler.Connect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	_, err := client.WriteSingleRegister(uint16(req.Register), uint16(req.Value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Write successful"})
}
