package utils

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

// nolint
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

// nolint
func listBucketObjects(ctx context.Context, client *storage.Client, projectID, bucket string) ([]*storage.ObjectAttrs, error) {
	it := client.Bucket(bucket).Objects(ctx, nil)
	var objects []*storage.ObjectAttrs
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return objects, fmt.Errorf("Cannot get bucket %q object attributes: %w", bucket, err)
		}
		objects = append(objects, attrs)
	}

	return objects, nil
}

func deleteBucketObjects(ctx context.Context, client *storage.Client, projectID, bucket string) error {

	objects, err := listBucketObjects(ctx, client, projectID, bucket)

	if err != nil {
		return err
	}

	if len(objects) == 0 {
		log.Printf("Not found GCM bucket %q objects to delete", bucket)
	}

	for _, obj := range objects {
		if err := client.Bucket(bucket).Object(obj.Name).Delete(ctx); err != nil {
			return fmt.Errorf("Cannot delete object %q from GCS bucket %q", obj.Name, bucket)
		}
		log.Printf("Deleted object %q from GCS bucket %q", obj.Name, bucket)
	}
	return nil
}

//ClearTFBucket deletes all objects from bucket
func ClearTFBucket(project, name string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create stogage client: %w", err)
	}
	defer client.Close()

	return deleteBucketObjects(ctx, client, project, name)

}

//DeleteTFBucket deletes a bucket
func DeleteTFBucket(project, name string) error {

	log.Printf("Deleting bucket %q...", name)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create stogage client: %w", err)
	}
	defer client.Close()

	err = deleteBucketObjects(ctx, client, project, name)

	if err != nil {
		return err
	}

	bucket := client.Bucket(name)

	if err := bucket.Delete(ctx); err != nil {
		gErr, ok := err.(*googleapi.Error)
		if !ok {
			return err
		}
		switch {
		case gErr.Code == 404:
			return nil
		default:
			return gErr
		}
	}

	log.Printf("Deleted bucket %q", name)

	return nil
}

// EnsureTFBucket ebsure TF bucket exists
func EnsureTFBucket(project, name string) (bool, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("Cannot create stogage client: %w", err)
	}
	defer client.Close()
	bucket := client.Bucket(name)

	if err := bucket.Create(ctx, project, &storage.BucketAttrs{Name: name}); err != nil {
		gErr, ok := err.(*googleapi.Error)
		if !ok {
			return false, err
		}
		if gErr.Code == 409 {
			// bucket exists
			return false, nil
		}
		return false, fmt.Errorf("Cannot create bucket %s. %w", name, err)
	}

	return true, nil
}
