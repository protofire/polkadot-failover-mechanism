package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	iam "google.golang.org/api/iam/v1"
)

// SAClean cleans service accounts
func SAClean(project, prefix string, dryRun bool) error {
	ctx := context.Background()
	client, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize iam client: %w", err)
	}

	response, err := client.Projects.ServiceAccounts.List("projects/" + project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Cannot get service accounts list: %w", err)
	}

	var serviceAccounts []string

	for _, account := range response.Accounts {

		accountName := lastPartOnSplit(account.Name, "/")
		if strings.HasPrefix(accountName, getPrefix(prefix)) {
			serviceAccounts = append(serviceAccounts, account.Name)
		}
	}

	if len(serviceAccounts) == 0 {
		log.Println("Not found service accounts to delete")
		return nil
	}

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, serviceAccount := range serviceAccounts {

		wg.Add(1)

		go func(name string, wg *sync.WaitGroup) {

			defer wg.Done()

			var err error

			log.Printf("Deleting service account: %q", name)

			if dryRun {
				return
			}

			log.Printf("Disabling service account: %q", name)

			if _, err = client.Projects.ServiceAccounts.Disable(name, &iam.DisableServiceAccountRequest{}).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not disable service account %q. %w", name, err)
				return
			}

			log.Printf("Deleting service account: %q", name)

			if _, err = client.Projects.ServiceAccounts.Delete(name).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete service account %q. %w", name, err)
				return
			}

			log.Printf("Successfully deleted service account: %q\n", name)

		}(serviceAccount, wg)

	}

	return helpers.WaitOnErrorChannel(ch, wg)

}
