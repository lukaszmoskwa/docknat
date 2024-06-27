package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/docker/docker/client"
	utils "github.com/lukaszmoskwa/docknat/internal"
)

func Run() {

	// Create a new IPTables instance
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Create a new Utils instance
	dockerCli, dockerErr := client.NewClientWithOpts(client.FromEnv)
	if dockerErr != nil {
		fmt.Println(dockerErr)
		return
	}
	ctx := context.Background()
	dockerCli.NegotiateAPIVersion(ctx)
	utils, utilsErr := utils.NewUtils(ipt, dockerCli)
	if utilsErr != nil {
		fmt.Println(utilsErr)
		return
	}
	// Retrieve the port mappings from the Docker API of the currently running containers
	dockerMapping, err := utils.RetrieveDockerPortMapping()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Retrieve the NAT rules
	rulesMapping := utils.RetrieveNatMapping()
	// Compare the port mappings with the NAT rules and remove the rules that are not needed
	toAdd, toRemove := utils.ComparePortMappings(dockerMapping, rulesMapping)
	for _, rule := range toRemove {
		err = utils.RemoveNatPreroutingRule(rule)
		if err != nil {
			fmt.Println("Rule match not found, skipping removal")
			continue
		} else {
			fmt.Printf("Removed rule: %s:%d -> %s:%d\n",
				rule.IP, rule.PublicPort, rule.BridgeIP, rule.PrivatePort)
		}
	}
	for _, rule := range toAdd {
		err = utils.AddNatPreroutingRule(rule)
		if err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Printf("Added rule: %s:%d -> %s:%d\n",
				rule.IP, rule.PublicPort, rule.BridgeIP, rule.PrivatePort)
		}
	}
}
