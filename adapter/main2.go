package main

import "fmt"

type computer interface {
	insertIntoLightningPort()
}

type mac struct {
}

func (mac *mac) insertIntoLightningPort() {
	fmt.Println("Lightning connector is plugged into mac machine.")
}

type windows struct {
}

func (windows *windows) insertIntoUSBPort() {
	fmt.Println("USB connector is plugged into windows machine.")
}

type windowsAdapter struct {
	windowsMachine *windows
}

func (w *windowsAdapter) insertIntoLightningPort() {
	fmt.Println("Adapter converts Lightning signal to USB.")
	w.windowsMachine.insertIntoUSBPort()
}

type client struct {
}

func (c *client) insertLightningConnectorIntoComputer(com computer) {
	fmt.Println("Client inserts Lightning connector into computer.")
	com.insertIntoLightningPort()
}

// func main() {
// 	client := client{}
// 	mac := &mac{}
// 	client.insertLightningConnectorIntoComputer(mac)

// 	windows := &windows{}

// 	windowsAdapter := &windowsAdapter{windowsMachine: windows}

// 	client.insertLightningConnectorIntoComputer(windowsAdapter)
// }
