package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func listMonitoringChannels(ctx context.Context, client *monitoring.NotificationChannelClient, project, prefix string) ([]string, error) {

	fullPrefix := helpers.GetPrefix(prefix)

	channelReq := &monitoringpb.ListNotificationChannelsRequest{
		Name:   "projects/" + project,
		Filter: fmt.Sprintf("name = starts_with('%s') OR display_name = starts_with('%s')", fullPrefix, fullPrefix),
		// Filter:  "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
		// OrderBy: "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
	}
	channelIt := client.ListNotificationChannels(ctx, channelReq)

	var channels []string

	for {
		channel, err := channelIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return channels, err
		}

		shortName := helpers.LastPartOnSplit(channel.Name, "/")
		shortDisplayName := helpers.LastPartOnSplit(channel.DisplayName, "/")

		if strings.HasPrefix(shortName, fullPrefix) || strings.HasPrefix(shortDisplayName, fullPrefix) {
			channels = append(channels, channel.Name)
		}
	}
	return channels, nil
}

func deleteNotificationChannels(ctx context.Context, client *monitoring.NotificationChannelClient, channelNames []string, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, channelName := range channelNames {

		wg.Add(1)

		go func(channel string, wg *sync.WaitGroup) {

			defer wg.Done()

			log.Printf("Deleting channel: %s", channel)

			if dryRun {
				return
			}

			req := &monitoringpb.DeleteNotificationChannelRequest{
				Name: channel,
			}

			if err := client.DeleteNotificationChannel(ctx, req); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete channel: %q. Status: %d\n", channel, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete channel %q. %w", channel, err)
				return
			}

			log.Printf("Successfully deleted channel: %q\n", channel)

		}(channelName, wg)

	}

	return helpers.WaitOnErrorChannel(ch, wg)

}

// NotificationChannelsClean cleans test notification channels
func NotificationChannelsClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := monitoring.NewNotificationChannelClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create notification channels client: %w", err)
	}
	channels, err := listMonitoringChannels(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get notification channels list: %w", err)
	}

	if len(channels) == 0 {
		log.Println("Not found notification channels to delete")
		return nil
	}

	return deleteNotificationChannels(ctx, client, channels, dryRun)

}
