package utils

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// Create a mock that mathes the *iptables.IPTables interface
type MockIptable struct {
	existingRules []string
	natRules      [][]string
}

type MockDockerClient struct {
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	return []types.Container{}, nil
}

func (m *MockIptable) AppendUnique(table, chain string, rulespec ...string) error {
	m.natRules = append(m.natRules, rulespec)
	return nil
}

func (m *MockIptable) List(table, chain string) ([]string, error) {
	// Return the existing rules with "any" prepended to the existingRules
	// This is to match the output of the iptables command
	var rules []string
	rules = append(rules, "any")
	rules = append(rules, m.existingRules...)
	return rules, nil
}

func (m *MockIptable) Delete(table, chain string, rulespec ...string) error {
	// Remove the rule from the list
	for i, rule := range m.natRules {
		if rule[0] == rulespec[0] && rule[1] == rulespec[1] && rule[2] == rulespec[2] {
			m.natRules = append(m.natRules[:i], m.natRules[i+1:]...)
		}
	}
	return nil
}

func NewMockUtils(
	iptablesRules []string,
) *Utils {
	var mockIptable *MockIptable
	if iptablesRules != nil {
		mockIptable = &MockIptable{
			existingRules: iptablesRules,
		}
	} else {
		mockIptable = &MockIptable{}
	}
	utils, err := NewUtils(mockIptable, &MockDockerClient{})
	if err != nil {
		return nil
	}
	return utils
}

func TestAddNatPreroutingRule(t *testing.T) {
	utils := NewMockUtils(nil)
	mockMapping := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  26888,
			PrivatePort: 3000,
		},
	}
	err := utils.AddNatPreroutingRule(mockMapping[0])
	if err != nil {
		t.Error("Error adding NAT rule")
	}
	// Check that the rule was added
	if len(utils.IPTables.(*MockIptable).natRules) != 2 {
		t.Error("Rule was not added")
	}
}

func TestRemoveNatPreroutingRule(t *testing.T) {
	utils := NewMockUtils(nil)
	mockMapping := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  26888,
			PrivatePort: 3000,
		},
	}
	err := utils.AddNatPreroutingRule(mockMapping[0])
	if err != nil {
		t.Error("Error adding NAT rule")
	}
	err = utils.RemoveNatPreroutingRule(mockMapping[0])
	if err != nil {
		t.Error("Error removing NAT rule")
	}
	// Check that the rule was removed
	if len(utils.IPTables.(*MockIptable).natRules) != 0 {
		t.Error("Rule was not removed")
	}
}

func TestComparePortMappings(t *testing.T) {
	utils := NewMockUtils(nil)
	mockMapping := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  26888,
			PrivatePort: 3000,
		},
	}
	sameRuleMapping := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  26888,
			PrivatePort: 3000,
		},
	}
	differentRuleMapping := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  28888,
			PrivatePort: 3000,
		},
	}
	toAdd, toRemove := utils.ComparePortMappings(mockMapping, sameRuleMapping)
	if len(toAdd) != 0 || len(toRemove) != 0 {
		t.Error("Rules should be the same, no rules should be added or removed")
	}
	toAdd, toRemove = utils.ComparePortMappings(mockMapping, differentRuleMapping)
	if len(toAdd) != 1 || len(toRemove) != 1 {
		t.Error("Rules should be different, one should be added and one should be removed")
	}
}

func TestWithExistingRules(t *testing.T) {
	mockIpTablesRules := []string{
		"-A PREROUTING -p tcp -m tcp --dport 30134 -j DNAT --to-destination 172.17.0.5:8777",
		"-A PREROUTING -p udp -m udp --dport 30134 -j DNAT --to-destination 172.17.0.5:8777",
		"-A PREROUTING -p tcp -m tcp --dport 20023 -j DNAT --to-destination 172.17.0.4:9090",
		"-A PREROUTING -p udp -m udp --dport 20023 -j DNAT --to-destination 172.17.0.4:9090",
		"-A PREROUTING -p udp -m udp --dport 5432 -j DNAT --to-destination 172.17.0.2:5432",
		"-A PREROUTING -p tcp -m tcp --dport 5432 -j DNAT --to-destination 172.17.0.2:5432",
		"-A PREROUTING -p tcp -m tcp --dport 25790 -j DNAT --to-destination 172.17.0.8:80",
		"-A PREROUTING -p udp -m udp --dport 25790 -j DNAT --to-destination 172.17.0.8:80",
		"-A PREROUTING -p tcp -m tcp --dport 25412 -j DNAT --to-destination 172.17.0.6:80",
		"-A PREROUTING -p udp -m udp --dport 25412 -j DNAT --to-destination 172.17.0.6:80",
	}
	utils := NewMockUtils(mockIpTablesRules)
	natMappings := utils.RetrieveNatMapping()
	if len(natMappings) != 10 {
		t.Error("Incorrect number of NAT mappings")
	}

	// Get a docker mapping
	dockerMappings := []PortMapping{
		{
			BridgeIP:    "172.17.0.1",
			IP:          "100.123.123.123",
			PublicPort:  28888,
			PrivatePort: 3000,
		},
	}
	toAdd, toRemove := utils.ComparePortMappings(dockerMappings, natMappings)
	// Check that the rule was added
	if len(toAdd) != 1 {
		t.Error("Rule is not to add, should be added")
	}
	// Check that the rule was removed
	if len(toRemove) != 10 {
		t.Error("10 rules are not set to be removed, should be removed")
	}
}

func TestRetrieveDockerContainerMapping(t *testing.T) {
	utils := NewMockUtils(nil)
	_, err := utils.RetrieveDockerPortMapping()
	if err != nil {
		t.Error("Error retrieving Docker container mapping")
	}
}

func TestWithIncompleteExistingRules(t *testing.T) {
	// 2 udp rules are already existing but the tcp rules are missing
	mockIpTablesRules := []string{
		"-A PREROUTING -p udp -m udp --dport 30134 -j DNAT --to-destination 172.17.0.5:8777",
		"-A PREROUTING -p udp -m udp --dport 20023 -j DNAT --to-destination 172.17.0.4:9090",
		"-A PREROUTING -p udp -m udp --dport 5432 -j DNAT --to-destination 172.17.0.2:5432",
		"-A PREROUTING -p tcp -m tcp --dport 5432 -j DNAT --to-destination 172.17.0.2:5432",
		"-A PREROUTING -p tcp -m tcp --dport 25790 -j DNAT --to-destination 172.17.0.8:80",
		"-A PREROUTING -p udp -m udp --dport 25790 -j DNAT --to-destination 172.17.0.8:80",
		"-A PREROUTING -p tcp -m tcp --dport 25412 -j DNAT --to-destination 172.17.0.6:80",
		"-A PREROUTING -p udp -m udp --dport 25412 -j DNAT --to-destination 172.17.0.6:80",
	}
	utils := NewMockUtils(mockIpTablesRules)
	natMappings := utils.RetrieveNatMapping()
	if len(natMappings) != 8 {
		t.Error("Incorrect number of NAT mappings")
	}
}
