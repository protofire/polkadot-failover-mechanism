package azure

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func getStorageURL(storageAccount, containerName string) *url.URL {
	storageURL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", storageAccount, containerName))
	return storageURL
}

// EnsureTFBucket creates azure blob storage
func EnsureTFBucket(storageAccount, storageAccessKey, containerName string, forceDelete bool) (bool, error) {

	credentials, err := azblob.NewSharedKeyCredential(storageAccount, storageAccessKey)
	if err != nil {
		return false, fmt.Errorf("Cannot create azure blob storage credentials: %w", err)
	}
	pipeline := azblob.NewPipeline(credentials, azblob.PipelineOptions{})

	storageURL := getStorageURL(storageAccount, containerName)
	containerURL := azblob.NewContainerURL(*storageURL, pipeline)

	log.Printf("Creating azure blob container %q...", containerName)

	ctx := context.Background()

	for {
		_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)

		if err != nil {
			if serr, ok := err.(azblob.StorageError); ok {
				switch serr.ServiceCode() {
				case azblob.ServiceCodeContainerAlreadyExists:
					log.Printf("Container %q already exists", containerName)
					if !forceDelete {
						return false, nil
					}
					err := DeleteTFBucket(storageAccount, storageAccessKey, containerName)
					if err != nil {
						return false, fmt.Errorf("Cannot delete container %q: %w", containerName, err)
					}
				case azblob.ServiceCodeContainerBeingDeleted:
					log.Printf("Container %q is being deleted. Waiting...", containerName)
					time.Sleep(1 * time.Second)
					continue
				}
				return false, serr
			}
			return false, err
		}
		log.Printf("Azure blob container %q has been created", containerName)
		return true, nil
	}

}

func listContainerBlobs(ctx context.Context, containerURL *azblob.ContainerURL) ([]azblob.BlobItemInternal, error) {

	var blobs []azblob.BlobItemInternal

	for blobMarker := (azblob.Marker{}); blobMarker.NotDone(); {
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, blobMarker, azblob.ListBlobsSegmentOptions{})

		if err != nil {
			if serr, ok := err.(azblob.StorageError); ok {
				return blobs, serr
			}
			return blobs, err
		}

		blobMarker = listBlob.NextMarker
		blobs = append(blobs, listBlob.Segment.BlobItems...)
	}
	return blobs, nil
}

func deleteContainerBlobs(ctx context.Context, containerURL *azblob.ContainerURL, blobs []azblob.BlobItemInternal) error {
	for _, blob := range blobs {
		log.Printf("Deleting blob %q from container %q...", blob.Name, containerURL.String())
		_, err := containerURL.NewBlobURL(blob.Name).Delete(ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})
		if err != nil {
			if storageError, ok := err.(azblob.StorageError); ok {
				if storageError.ServiceCode() == azblob.ServiceCodeBlobNotFound {
					continue
				}
				return storageError

			}
			return err
		}
		log.Printf("Deleted blob %q from container %q", blob.Name, containerURL.String())
	}
	return nil
}

// DeleteTFBucket deletes azure blob storage
func DeleteTFBucket(storageAccount, storageAccessKey, containerName string) error {

	credentials, err := azblob.NewSharedKeyCredential(storageAccount, storageAccessKey)
	if err != nil {
		return fmt.Errorf("Cannot  create azure blob storage credentials: %w", err)
	}
	pipeline := azblob.NewPipeline(credentials, azblob.PipelineOptions{})

	storageURL := getStorageURL(storageAccount, containerName)
	containerURL := azblob.NewContainerURL(*storageURL, pipeline)

	log.Printf("Deleting azure blob container %q...", containerName)

	ctx := context.Background()

	blobs, err := listContainerBlobs(ctx, &containerURL)

	if err != nil {
		return err
	}

	err = deleteContainerBlobs(ctx, &containerURL, blobs)

	if err != nil {
		return err
	}
	for {
		_, err = containerURL.Delete(ctx, azblob.ContainerAccessConditions{})

		if err != nil {
			if serr, ok := err.(azblob.StorageError); ok {
				switch serr.ServiceCode() {
				case azblob.ServiceCodeContainerNotFound:
					log.Printf("Container %q not found", containerName)
					return nil
				case azblob.ServiceCodeContainerBeingDeleted:
					log.Printf("Container %q is being deleted. Waiting...", containerName)
					time.Sleep(1 * time.Second)
					continue
				}
			}
			return err
		}

		return nil

	}

}

// ClearTFBucket deletes all blobs from azure blob storage
func ClearTFBucket(storageAccount, storageAccessKey, containerName string) error {

	credentials, err := azblob.NewSharedKeyCredential(storageAccount, storageAccessKey)
	if err != nil {
		return fmt.Errorf("Cannot  create azure blob storage credentials: %w", err)
	}
	pipeline := azblob.NewPipeline(credentials, azblob.PipelineOptions{})

	storageURL := getStorageURL(storageAccount, containerName)
	containerURL := azblob.NewContainerURL(*storageURL, pipeline)

	ctx := context.Background()

	blobs, err := listContainerBlobs(ctx, &containerURL)

	if err != nil {
		return err
	}

	err = deleteContainerBlobs(ctx, &containerURL, blobs)

	if err != nil {
		return err
	}

	return nil

}
