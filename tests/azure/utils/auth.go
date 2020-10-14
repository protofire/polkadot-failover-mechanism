package utils

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func getAuthorizer() (autorest.Authorizer, error) {
	return auth.NewAuthorizerFromEnvironment()
}

func getVaultAuthorizer() (autorest.Authorizer, error) {
	return auth.NewAuthorizerFromCLIWithResource("https://vault.azure.net")
}
