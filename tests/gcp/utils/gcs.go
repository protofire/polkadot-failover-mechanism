package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

func listBuckets(ctx context.Context, client *storage.Client, projectID string) ([]string, error) {

	var buckets []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
	}
	return buckets, nil
}

// EnsureTFBucket ebsure TF bucket exists
func EnsureTFBucket(project, name string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create stogage client: %#w", err)
	}
	defer client.Close()
	bucket := client.Bucket(name)

	if err := bucket.Create(ctx, project, &storage.BucketAttrs{Name: name}); err != nil {
		if gErr, ok := err.(*googleapi.Error); !(ok && gErr.Code == 409) {
			return fmt.Errorf("Cannot create bucket %s. %#w", name, err)
		}
	}

	buckets, err := listBuckets(ctx, client, project)

	if err != nil {
		return fmt.Errorf("Cannot get list of GCS buckets: %#w", err)
	}

	if _, ok := contains(buckets, name); !ok {
		return fmt.Errorf("Required bucket %s not in buckets list: %s", name, strings.Join(buckets, ", "))
	}

	return nil
}
