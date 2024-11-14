package main

import (
	"fmt"
)

func (cli *CLI) testsend(address string, testsendData string) {
	fmt.Println(testsendData)
	fmt.Println(address)
	fmt.Println("(cli *CLI)testsend!")

	//sendTestdata(address, testsendData)
}
