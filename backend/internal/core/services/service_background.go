package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

type Latest struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}
type BackgroundService struct {
	repos  *repo.AllRepos
	latest Latest
}

func (svc *BackgroundService) SendNotifiersToday(ctx context.Context) error {
	// Get All Groups
	groups, err := svc.repos.Groups.GetAllGroups(ctx, uuid.Nil)
	if err != nil {
		return err
	}

	today := types.DateFromTime(time.Now())

	for i := range groups {
		group := groups[i]

		entries, err := svc.repos.MaintEntry.GetScheduled(ctx, group.ID, today)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			log.Debug().
				Str("group_name", group.Name).
				Str("group_id", group.ID.String()).
				Msg("No scheduled maintenance for today")
			continue
		}

		notifiers, err := svc.repos.Notifiers.GetByGroup(ctx, group.ID)
		if err != nil {
			return err
		}

		urls := lo.Map(notifiers, func(n repo.NotifierOut, _ int) string {
			return n.URL
		})

		bldr := strings.Builder{}

		bldr.WriteString("Homebox Maintenance for (")
		bldr.WriteString(today.String())
		bldr.WriteString("):\n")

		for i := range entries {
			entry := entries[i]
			bldr.WriteString(" - ")
			bldr.WriteString(entry.Name)
			bldr.WriteString("\n")
		}

		var sendErrs []error
		for i := range urls {
			err := shoutrrr.Send(urls[i], bldr.String())

			if err != nil {
				sendErrs = append(sendErrs, err)
			}
		}

		if len(sendErrs) > 0 {
			return sendErrs[0]
		}
	}

	return nil
}

func (svc *BackgroundService) GetLatestGithubRelease(ctx context.Context) error {
	url := "https://api.github.com/repos/sysadminsmedia/homebox/releases/latest"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create latest version request: %w", err)
	}

	req.Header.Set("User-Agent", "Homebox-Version-Checker")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make latest version request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("error closing latest version response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("latest version unexpected status code: %d", resp.StatusCode)
	}

	// ignoring fields that are not relevant
	type Release struct {
		ReleaseVersion string    `json:"tag_name"`
		PublishedAt    time.Time `json:"published_at"`
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode latest version response: %w", err)
	}

	svc.latest = Latest{
		Version: release.ReleaseVersion,
		Date:    release.PublishedAt.String(),
	}

	return nil
}

func (svc *BackgroundService) GetLatestVersion() Latest {
	return svc.latest
}
