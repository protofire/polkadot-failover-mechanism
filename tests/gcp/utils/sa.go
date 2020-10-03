package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	iam "google.golang.org/api/iam/v1"
)

// SAClean cleans service accounts
func SAClean(t *testing.T, project, prefix string) error {
	ctx := context.Background()
	client, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize iam client: %#w", err)
	}

	response, err := client.Projects.ServiceAccounts.List("projects/" + project).Do()
	if err != nil {
		return fmt.Errorf("Cannot get service accounts list: %#w", err)
	}

	var serviceAccounts []string

	for _, account := range response.Accounts {

		accountNames := strings.Split(account.Name, ":")
		if len(accountNames) == 1 {
			continue
		}
		if strings.HasPrefix(accountNames[1], getPrefix(prefix)) {
			serviceAccounts = append(serviceAccounts, account.Name)
		}
	}

	if len(serviceAccounts) == 0 {
		t.Logf("Not found service accounts to delete")
		return nil
	}

	t.Logf("Prepared service accounts to delete: %s", strings.Join(serviceAccounts, ", "))

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, serviceAccount := range serviceAccounts {

		wg.Add(1)

		go func(name string, wg *sync.WaitGroup) {

			defer wg.Done()

			var err error

			if _, err = client.Projects.ServiceAccounts.Delete(name).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete service account %s. %#w", name, err)
				return
			}

			t.Logf("Successfully deleted service account: %s", name)

		}(serviceAccount, wg)

	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	var result *multierror.Error

	for err := range ch {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()

}
