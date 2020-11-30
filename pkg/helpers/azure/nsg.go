package azure

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
)

type securityRuleItem struct {
	sourcePortRanges             []string
	sourceAddressesPrefixes      []string
	destinationPortRanges        []string
	destinationAddressesPrefixes []string
	priority                     int
	protocol                     string
	direction                    string
}

func (sr securityRuleItem) equals(other securityRuleItem) bool {
	return reflect.DeepEqual(sr, other)
}

func compareRules(testRules []securityRuleItem, actualRules []securityRuleItem) error {

	for _, actualRule := range actualRules {
		found := false
		for _, testRule := range testRules {
			if testRule.equals(actualRule) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("cannot find coinside rule for %+v", actualRule)
		}
	}
	return nil
}

//nolint
func getNetworkSecurityGroupClient(subscriptionID string) (network.SecurityGroupsClient, error) {

	client := network.NewSecurityGroupsClient(subscriptionID)

	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil
}

//nolint
func getNetworkSecurityRuleClient(subscriptionID string) (network.SecurityRulesClient, error) {

	client := network.NewSecurityRulesClient(subscriptionID)

	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil
}

//nolint
func getSecurityGroups(subscriptionID, resourceGroup string) ([]network.SecurityGroup, error) {

	client, err := getNetworkSecurityGroupClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	result, err := client.List(ctx, resourceGroup)

	if err != nil {
		return nil, err
	}

	groups := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		groups = append(groups, result.Values()...)
	}

	return groups, nil

}

func filterSecurityGroups(sgs *[]network.SecurityGroup, handler func(sg network.SecurityGroup) bool) {
	start := 0
	for i := start; i < len(*sgs); i++ {
		if !handler((*sgs)[i]) {
			// sgs will be deleted
			continue
		}
		if i != start {
			(*sgs)[start], (*sgs)[i] = (*sgs)[i], (*sgs)[start]
		}
		start++
	}

	*sgs = (*sgs)[:start]

}

func prepareTestRules(exposePrometheus, exposeSSH bool) []securityRuleItem {
	subnetPorts := []string{"8300", "8301", "8600", "8500", "8302"}
	if exposePrometheus {
		subnetPorts = append(subnetPorts, "9273")
	}
	sort.Strings(subnetPorts)
	inboundSubnetRules := []securityRuleItem{
		{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"10.0.0.0/24"},
			destinationPortRanges:        subnetPorts,
			destinationAddressesPrefixes: []string{"*"},
			priority:                     102,
			protocol:                     "*",
			direction:                    "Inbound",
		},
		{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"10.1.0.0/24"},
			destinationPortRanges:        subnetPorts,
			destinationAddressesPrefixes: []string{"*"},
			priority:                     103,
			protocol:                     "*",
			direction:                    "Inbound",
		},
		{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"10.2.0.0/24"},
			destinationPortRanges:        subnetPorts,
			destinationAddressesPrefixes: []string{"*"},
			priority:                     104,
			protocol:                     "*",
			direction:                    "Inbound",
		},
	}
	inboundWildcardRules := []securityRuleItem{
		{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"*"},
			destinationPortRanges:        []string{"30333"},
			destinationAddressesPrefixes: []string{"*"},
			priority:                     101,
			protocol:                     "*",
			direction:                    "Inbound",
		},
	}
	if exposeSSH {
		inboundWildcardRules = append(inboundWildcardRules, securityRuleItem{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"*"},
			destinationPortRanges:        []string{"22"},
			destinationAddressesPrefixes: []string{"*"},
			priority:                     100,
			protocol:                     "Tcp",
			direction:                    "Inbound",
		})
	}
	if exposePrometheus {
		inboundWildcardRules = append(inboundWildcardRules, securityRuleItem{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"*"},
			destinationPortRanges:        []string{"9273"},
			destinationAddressesPrefixes: []string{"*"},
			priority:                     105,
			protocol:                     "Tcp",
			direction:                    "Inbound",
		})
	}

	outboundRules := []securityRuleItem{
		{
			sourcePortRanges:             []string{"*"},
			sourceAddressesPrefixes:      []string{"*"},
			destinationPortRanges:        []string{"*"},
			destinationAddressesPrefixes: []string{"*"},
			priority:                     100,
			protocol:                     "Tcp",
			direction:                    "Outbound",
		},
	}

	var rules []securityRuleItem
	rules = append(rules, inboundSubnetRules...)
	rules = append(rules, inboundWildcardRules...)
	rules = append(rules, outboundRules...)
	return rules

}

// SecurityGroupsCheck checks that all SG rules has been applied correctly
func SecurityGroupsCheck(prefix, subscriptionID, resourceGroup string, exposePrometheus, exposeSSH bool) error {

	sgs, err := getSecurityGroups(subscriptionID, resourceGroup)

	if err != nil {
		return err
	}

	filterSecurityGroups(&sgs, func(sg network.SecurityGroup) bool {
		return strings.HasPrefix(*sg.Name, helpers.GetPrefix(prefix))
	})

	var rules []securityRuleItem

	for _, sg := range sgs {
		for _, sr := range *sg.SecurityRules {
			sap := *sr.SourceAddressPrefixes
			if len(sap) == 0 {
				sap = []string{*sr.SourceAddressPrefix}
			}
			dap := *sr.DestinationAddressPrefixes
			if len(dap) == 0 {
				dap = []string{*sr.DestinationAddressPrefix}
			}
			spr := *sr.SourcePortRanges
			if len(spr) == 0 {
				spr = []string{*sr.SourcePortRange}
			}
			dpr := *sr.DestinationPortRanges
			if len(dpr) == 0 {
				dpr = []string{*sr.DestinationPortRange}
			}
			sort.Strings(dpr)
			sort.Strings(dap)
			sort.Strings(spr)
			sort.Strings(sap)

			rule := securityRuleItem{
				sourceAddressesPrefixes:      sap,
				sourcePortRanges:             spr,
				destinationAddressesPrefixes: dap,
				destinationPortRanges:        dpr,
				priority:                     int(*sr.Priority),
				protocol:                     string(sr.Protocol),
				direction:                    string(sr.Direction),
			}

			rules = append(rules, rule)
		}
	}

	return compareRules(prepareTestRules(exposePrometheus, exposeSSH), rules)

}
