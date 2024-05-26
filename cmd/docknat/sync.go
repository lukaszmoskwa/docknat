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
	// Run the job every 10 seconds
	for range time.Tick(10 * time.Second) {
		// Retrieve the port mappings from the Docker API of the currently running containers
		portMappings, err := utils.RetrievePortMapping()
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Retrieve the NAT rules
		rulesMapping := utils.ParseNatRulesToPortMappings(utils.RetrieveNatRules(ipt))
		// Compare the port mappings with the NAT rules and remove the rules that are not needed
		toAdd, toRemove := utils.ComparePortMappings(portMappings, rulesMapping)
		for _, rule := range toRemove {
			err = utils.RemoveNatPreroutingRule(ipt, rule)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				fmt.Println("Removed rule:", rule)
			}
		}
		for _, rule := range toAdd {
			err = utils.AddNatPreroutingRule(ipt, rule)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				fmt.Println("Added rule:", rule)
			}
		}
	}
}
