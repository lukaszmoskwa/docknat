package utils

import (
	"context"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"strconv"
)

type PortMapping struct {
	BridgeIP    string
	IP          string
	PublicPort  uint16
	PrivatePort uint16
	Type        string
}

func RetrievePortMapping() ([]PortMapping, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	fmt.Println("Scanning containers...")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var portMappings []PortMapping
	for _, container := range containers {
		id := container.ID[:12]
		fmt.Println("Inspecting container", id)
		for _, port := range container.Ports {
			fmt.Printf("Port: %s:%d -> %d, %s", port.IP, port.PublicPort, port.PrivatePort, port.Type)
			networkSettings := container.NetworkSettings
			if networkSettings == nil {
				fmt.Println("Container", id, "has no network settings")
				continue
			}
			networks := networkSettings.Networks
			if networks == nil {
				fmt.Println("Container", id, "has no networks")
				continue
			}
			bridgeNetwork := networks["bridge"]
			if bridgeNetwork == nil {
				fmt.Println("Container", id, "has no bridge network")
				continue
			}
			portMappings = append(portMappings, PortMapping{
				IP:          port.IP,
				PublicPort:  port.PublicPort,
				PrivatePort: port.PrivatePort,
				Type:        port.Type,
				BridgeIP:    container.NetworkSettings.Networks["bridge"].IPAddress,
			})
		}
	}
	return portMappings, nil
}

func RetrieveNatRules(ipt *iptables.IPTables) []string {
	rules, _ := ipt.List("nat", "PREROUTING")
	// Remove the first string from rules, which is the chain name
	rules = rules[1:]
	return rules
}

func splitRule(rule string) []string {
	var parts []string
	var part string
	for _, char := range rule {
		if char == ' ' {
			parts = append(parts, part)
			part = ""
		} else {
			part += string(char)
		}
	}
	parts = append(parts, part)
	return parts
}

func ParseNatRulesToPortMappings(rules []string) []PortMapping {
	var portMappings []PortMapping
	for _, rule := range rules {
		// Split the rule into parts
		parts := splitRule(rule)
		fmt.Println("Parts:", parts)
		// Check if the rule is a DNAT rule
		if parts[0] == "-p" && parts[1] == "tcp" && parts[2] == "-d" && parts[4] == "--dport" && parts[6] == "-j" && parts[7] == "DNAT" && parts[8] == "--to-destination" {
			publicPort, _ := strconv.ParseUint(parts[5], 10, 16)
			privatePort, _ := strconv.ParseUint(parts[9], 10, 16)
			portMappings = append(portMappings, PortMapping{
				IP:          parts[3],
				PublicPort:  uint16(publicPort),
				PrivatePort: uint16(privatePort),
				Type:        "tcp",
				BridgeIP:    parts[10],
			})
			fmt.Println("Port mapping found:", portMappings)
		}
	}
	return portMappings
}

func ComparePortMappings(portMappings, rulesMapping []PortMapping) ([]PortMapping, []PortMapping) {
	var toAdd, toRemove []PortMapping
	for _, portMapping := range portMappings {
		found := false
		for _, ruleMapping := range rulesMapping {
			if portMapping == ruleMapping {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, portMapping)
		}
	}
	for _, ruleMapping := range rulesMapping {
		found := false
		for _, portMapping := range portMappings {
			if portMapping == ruleMapping {
				found = true
				break
			}
		}
		if !found {
			toRemove = append(toRemove, ruleMapping)
		}
	}
	return toAdd, toRemove
}

func AddNatPreroutingRule(ipt *iptables.IPTables, mapping PortMapping) error {
	return ipt.AppendUnique("nat", "PREROUTING", "-p", mapping.Type, "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
}

func ResetNatRules(ipt *iptables.IPTables) error {
	return ipt.ClearChain("nat", "PREROUTING")
}

func RemoveNatPreroutingRule(ipt *iptables.IPTables, mapping PortMapping) error {
	return ipt.Delete("nat", "PREROUTING", "-p", mapping.Type, "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
}
