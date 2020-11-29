package helpers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func prometheusCheck(url string, metricName string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("prometheus check. unexpected response code: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if !bytes.Contains(body, []byte(metricName)) {
		return fmt.Errorf("prometheus checks. unexpected response: %s. does not contain %s", StringSlice(string(body), 50), metricName)
	}
	return nil
}

func WaitPrometheusTarget(ctx context.Context, url string, metricName string) error {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return prometheusCheck(url, metricName)
		case <-ticker.C:
			if err := prometheusCheck(url, metricName); err != nil {
				log.Printf("[DEBUG] Getting prometheus metrics %s error: %v", metricName, err)
				continue
			}
			return nil
		}
	}

}
