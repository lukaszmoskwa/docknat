package utils

import (
	"context"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"strconv"
	"strings"
)

type PortMapping struct {
	BridgeIP    string
	IP          string
	PublicPort  uint16
	PrivatePort uint16
	Type        string
}

func RetrieveDockerPortMapping() ([]PortMapping, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var portMappings []PortMapping
	for _, container := range containers {
		for _, port := range container.Ports {
			networkSettings := container.NetworkSettings
			if networkSettings == nil {
				continue
			}
			networks := networkSettings.Networks
			if networks == nil {
				continue
			}
			bridgeNetwork := networks["bridge"]
			if bridgeNetwork == nil {
				continue
			}
			if port.PublicPort == 0 {
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

func retrieveNatRules(ipt *iptables.IPTables) []string {
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

func RetrieveNatMapping(ipt *iptables.IPTables) []PortMapping {
	rules := retrieveNatRules(ipt)

	var portMappings []PortMapping
	for _, rule := range rules {

		var hash = make(map[string]string)
		parts := splitRule(rule)

		for index := range parts {
			if index%2 == 0 {
				hash[parts[index]] = parts[index+1]
			}
		}

		// Check if the rule is a DNAT rule
		if hash["-j"] != "DNAT" && hash["--jump"] != "DNAT" {
			continue
		}

		publicPort, _ := strconv.ParseUint(hash["--dport"], 10, 16)
		toDestination := hash["--to-destination"]
		splittedToDestination := strings.Split(toDestination, ":")
		toDestinationIP := splittedToDestination[0]
		toDestinationPort, _ := strconv.ParseUint(splittedToDestination[1], 10, 16)
		// Check if the rule is a DNAT rule
		portMappings = append(portMappings, PortMapping{
			IP:          parts[3],
			PublicPort:  uint16(publicPort),
			PrivatePort: uint16(toDestinationPort),
			Type:        "tcp",
			BridgeIP:    toDestinationIP,
		})
	}
	return portMappings
}

func areMappingsEqual(a, b PortMapping) bool {
	return a.PublicPort == b.PublicPort && a.PrivatePort == b.PrivatePort
}

// ComparePortMappings compares the port mappings retrieved from the Docker API with the NAT rules
// and returns the rules that need to be added and removed
func ComparePortMappings(dockerMapping, rulesMapping []PortMapping) ([]PortMapping, []PortMapping) {
	var toAdd, toRemove []PortMapping
	// If a docker mapping is not in the rules mapping, add it
	for _, dockerMap := range dockerMapping {
		found := false
		for _, ruleMap := range rulesMapping {
			if areMappingsEqual(dockerMap, ruleMap) {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, dockerMap)
		}
	}
	// If a rule mapping is not in the docker mapping, but is in the rules mapping, remove it
	for _, ruleMap := range rulesMapping {
		found := false
		for _, dockerMap := range dockerMapping {
			if areMappingsEqual(ruleMap, dockerMap) {
				found = true
				break
			}
		}
		if !found {
			toRemove = append(toRemove, ruleMap)
		}
	}

	return toAdd, toRemove
}

func AddNatPreroutingRule(ipt *iptables.IPTables, mapping PortMapping) error {
	// Append both the UDP and TCP rules
	tcpError := ipt.AppendUnique("nat", "PREROUTING", "-p", "tcp", "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
	if tcpError != nil {
		return tcpError
	}
	return ipt.AppendUnique("nat", "PREROUTING", "-p", "udp", "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
}

func ResetNatRules(ipt *iptables.IPTables) error {
	return ipt.ClearChain("nat", "PREROUTING")
}

func RemoveNatPreroutingRule(ipt *iptables.IPTables, mapping PortMapping) error {
	tcpError := ipt.Delete("nat", "PREROUTING", "-p", "tcp", "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
	if tcpError != nil {
		return tcpError
	}
	return ipt.Delete("nat", "PREROUTING", "-p", "udp", "--dport", strconv.Itoa(int(mapping.PublicPort)), "--jump", "DNAT", "--to-destination", fmt.Sprintf("%s:%d", mapping.BridgeIP, mapping.PrivatePort))
}
