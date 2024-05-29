package main

import (
	"fmt"
	"time"

	"github.com/coreos/go-iptables/iptables"
	utils "github.com/lukaszmoskwa/docknat/internal"
)

func Run() {
	// Create a new IPTables instance
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Run the job every 2 seconds
	for range time.Tick(2 * time.Second) {
		// Retrieve the port mappings from the Docker API of the currently running containers
		dockerMapping, err := utils.RetrieveDockerPortMapping()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Retrieve the NAT rules
		rulesMapping := utils.RetrieveNatMapping(ipt)
		// Compare the port mappings with the NAT rules and remove the rules that are not needed
		toAdd, toRemove := utils.ComparePortMappings(dockerMapping, rulesMapping)
		for _, rule := range toRemove {
			err = utils.RemoveNatPreroutingRule(ipt, rule)
			if err != nil {
				fmt.Println("Rule match not found, skipping removal")
				continue
			} else {
				fmt.Printf("Removed rule: %s %s:%d -> %s:%d\n",
					rule.Type, rule.IP, rule.PublicPort, rule.BridgeIP, rule.PrivatePort)
			}
		}
		for _, rule := range toAdd {
			err = utils.AddNatPreroutingRule(ipt, rule)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				fmt.Printf("Added rule: %s %s:%d -> %s:%d\n",
					rule.Type, rule.IP, rule.PublicPort, rule.BridgeIP, rule.PrivatePort)
			}
		}
	}
}
