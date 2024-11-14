package main

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	fmt.Println("func (cli *CLI) startNode(nodeID, minerAddress string) ")
	fmt.Println("nodeID:", nodeID)
	fmt.Println("minerAddress:", minerAddress)
	StartServer(nodeID, minerAddress)
}
