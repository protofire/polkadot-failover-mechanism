package azure

import (
	"context"
	"fmt"
	"reflect"
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

	for _, testRule := range testRules {
		found := false
		for _, actualRule := range actualRules {
			if testRule.equals(actualRule) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Cannot find coinside rule for %#v", testRule)
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
func getSecurityGroups(prefix, subscriptionID, resourceGroup string) ([]network.SecurityGroup, error) {

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

func prepareTestRules() []securityRuleItem {
	rule1 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"30333"}, destinationAddressesPrefixes: []string{"*"}, priority: 101, protocol: "*", direction: "Inbound"}
	rule2 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.1.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 103, protocol: "*", direction: "Inbound"}
	rule3 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"*"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Outbound"}
	rule4 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.2.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 104, protocol: "*", direction: "Inbound"}
	rule5 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.0.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 102, protocol: "*", direction: "Inbound"}
	rule6 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"22"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Inbound"}
	rule7 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"30333"}, destinationAddressesPrefixes: []string{"*"}, priority: 101, protocol: "*", direction: "Inbound"}
	rule8 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"*"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Outbound"}
	rule9 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"22"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Inbound"}
	rule10 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.2.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 104, protocol: "*", direction: "Inbound"}
	rule11 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.0.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 102, protocol: "*", direction: "Inbound"}
	rule12 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.1.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 103, protocol: "*", direction: "Inbound"}
	rule13 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.2.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 104, protocol: "*", direction: "Inbound"}
	rule14 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"*"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Outbound"}
	rule15 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.1.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 103, protocol: "*", direction: "Inbound"}
	rule16 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"30333"}, destinationAddressesPrefixes: []string{"*"}, priority: 101, protocol: "*", direction: "Inbound"}
	rule17 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"10.0.0.0/24"}, destinationPortRanges: []string{"8301", "8300", "8600", "8500", "8302"}, destinationAddressesPrefixes: []string{"*"}, priority: 102, protocol: "*", direction: "Inbound"}
	rule18 := securityRuleItem{sourcePortRanges: []string{"*"}, sourceAddressesPrefixes: []string{"*"}, destinationPortRanges: []string{"22"}, destinationAddressesPrefixes: []string{"*"}, priority: 100, protocol: "Tcp", direction: "Inbound"}
	return []securityRuleItem{
		rule1,
		rule2,
		rule3,
		rule4,
		rule5,
		rule6,
		rule7,
		rule8,
		rule9,
		rule10,
		rule11,
		rule12,
		rule13,
		rule14,
		rule15,
		rule16,
		rule17,
		rule18,
	}
}

// SecurityGroupsCheck checks that all SG rules has been applied correctly
func SecurityGroupsCheck(prefix, subscriptionID, resourceGroup string) error {

	sgs, err := getSecurityGroups(prefix, subscriptionID, resourceGroup)

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

	return compareRules(prepareTestRules(), rules)

}
