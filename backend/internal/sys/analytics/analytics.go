// Package analytics provides analytics function that sends data to a remote server.
package analytics

import (
	"bytes"
	"encoding/json"
	"github.com/shirou/gopsutil/v4/host"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var startTime = time.Now()

type Data struct {
	Domain string                 `json:"domain"`
	Name   string                 `json:"name"`
	URL    string                 `json:"url"`
	Props  map[string]interface{} `json:"props"`
}

func Send(version, buildInfo string) error {
	hostData, _ := host.Info()
	analytics := Data{
		Domain: "homebox.software",
		URL:    "https://homebox.software/stats",
		Name:   "stats",
		Props: map[string]interface{}{
			"version":          version + "/" + buildInfo,
			"os":               hostData.OS,
			"platform":         hostData.Platform,
			"platform_family":  hostData.PlatformFamily,
			"platform_version": hostData.PlatformVersion,
			"kernel_arch":      hostData.KernelArch,
			"virt_type":        hostData.VirtualizationSystem,
			"uptime_sec":       time.Since(startTime).Seconds(),
		},
	}
	jsonBody, err := json.Marshal(analytics)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal analytics data")
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest("POST", "https://a.sysadmins.zone/api/event", bodyReader)
	if err != nil {
		log.Error().Err(err).Msg("failed to create analytics request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Homebox/"+version+"/"+buildInfo+" (https://homebox.software)")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to send analytics request")
		return err
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close response body")
		}
	}()
	return nil
}
