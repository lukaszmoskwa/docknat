package utils

import (
	"github.com/coreos/go-iptables/iptables"
	"testing"
)

func TestRetrieveNatRules(t *testing.T) {
	ipt, error := iptables.New()
	defer ResetNatRules(ipt)
	if error != nil {
		t.Error("Error creating iptables object: (", error, ")")
	}
	l := RetrieveNatRules(ipt)
	if len(l) != 0 {
		t.Error("The list of NAT rules should be empty")
	}
}

func TestAddingNatRule(t *testing.T) {
	ipt, error := iptables.New()
	defer ResetNatRules(ipt)
	if error != nil {
		t.Error("Error creating iptables object: (", error, ")")
	}
	error = AddNatPreroutingRule(ipt, "--destination", "123.123.123.123", "--protocol", "tcp", "--dport", "5432", "--jump", "DNAT", "--to-destination", "234.234.234.234:5432")
	if error != nil {
		t.Error("Error adding NAT rule: (", error, ")")
	}
	// Check if the rule was added
	l := RetrieveNatRules(ipt)
	if len(l) != 1 {
		t.Error("The list of NAT rules should have one rule")
	}
}

func TestAddingMultipleTimeSameRule(t *testing.T) {
	ipt, error := iptables.New()
	defer ResetNatRules(ipt)
	if error != nil {
		t.Error("Error creating iptables object: (", error, ")")
	}
	_ = AddNatPreroutingRule(ipt, "--destination", "123.123.123.123", "--protocol", "tcp", "--dport", "5432", "--jump", "DNAT", "--to-destination", "234.234.234.234:5432")
	_ = AddNatPreroutingRule(ipt, "--destination", "123.123.123.123", "--protocol", "tcp", "--dport", "5432", "--jump", "DNAT", "--to-destination", "234.234.234.234:5432")
	if error != nil {
		t.Error("Error adding NAT rule: (", error, ")")
	}
	// Check if the rule was added
	l := RetrieveNatRules(ipt)
	if len(l) != 1 {
		t.Error("The list of NAT rules should have one rule")
	}
}

func TestRemovingNatRule(t *testing.T) {
	ipt, error := iptables.New()
	defer ResetNatRules(ipt)
	if error != nil {
		t.Error("Error creating iptables object: (", error, ")")
	}
	_ = AddNatPreroutingRule(ipt, "--destination", "123.123.123.123", "--protocol", "tcp", "--dport", "5432", "--jump", "DNAT", "--to-destination", "234.234.234.234:5432")
	error = RemoveNatPreroutingRule(ipt, "--destination", "123.123.123.123", "--protocol", "tcp", "--dport", "5432", "--jump", "DNAT", "--to-destination", "234.234.234.234:5432")
	if error != nil {
		t.Error("Error removing NAT rule: (", error, ")")
	}
	// Check if the rule was removed
	l := RetrieveNatRules(ipt)
	if len(l) != 0 {
		t.Error("The list of NAT rules should be empty")
	}
}
