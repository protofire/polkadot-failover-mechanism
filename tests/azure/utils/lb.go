package utils

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"
)

// getLoadBalancers returns tests loadbalancers
func getLoadBalancers(prefix, subscriptionID, resourceGroup string) ([]network.LoadBalancer, error) {

	lbClient := network.NewLoadBalancersClient(subscriptionID)
	ctx := context.Background()

	resultPage, err := lbClient.List(ctx, resourceGroup)

	if err != nil {
		return nil, err
	}

	lbs := resultPage.Values()

	for err := resultPage.Next(); err == nil; err = resultPage.Next() {
		lbs = append(lbs, resultPage.Values()...)
	}

	return lbs, nil

}

func filterLoadBalancers(lbs *[]network.LoadBalancer, handler func(lb network.LoadBalancer) bool) {

	start := 0
	for i := start; i < len(*lbs); i++ {
		if !handler((*lbs)[i]) {
			// lb will be deleted
			continue
		}
		if i != start {
			(*lbs)[start], (*lbs)[i] = (*lbs)[i], (*lbs)[start]
		}
		start++
	}

	*lbs = (*lbs)[:start]

}

// GetLoadBalancers returns loadbalancers for resource group
func GetLoadBalancers(prefix, subscriptionID, resourceGroup string) ([]network.LoadBalancer, error) {
	lbs, err := getLoadBalancers(prefix, subscriptionID, resourceGroup)
	if err != nil {
		return nil, err
	}

	filterLoadBalancers(&lbs, func(lb network.LoadBalancer) bool {
		return strings.HasPrefix(*lb.Name, helpers.GetPrefix(prefix))
	})

	return lbs, nil

}

// GetLoadBalancerIPs return map LB ID to slice of public IPs
func GetLoadBalancerIPs(lbs []network.LoadBalancer) map[string][]string {

	ips := make(map[string][]string)

	for _, lb := range lbs {
		for _, fc := range *lb.FrontendIPConfigurations {
			ips[*lb.ID] = append(ips[*lb.ID], *fc.PublicIPAddress.IPAddress)
		}
	}

	return ips
}
