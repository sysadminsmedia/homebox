package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/notifier"
)

type Latest struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}
type BackgroundService struct {
	repos          *repo.AllRepos
	latest         Latest
	notifierConfig *config.NotifierConf
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

		notifiers, err := svc.repos.Notifiers.GetActiveByGroup(ctx, group.ID)
		if err != nil {
			return err
		}

		if len(notifiers) == 0 {
			log.Debug().
				Str("group_name", group.Name).
				Str("group_id", group.ID.String()).
				Msg("No active notifiers configured")
			continue
		}

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
		// One sender for the batch so connections are pooled.
		sender := notifier.NewSender(svc.notifierConfig)
		for i := range notifiers {
			err := sender.Send(notifiers[i].URL, bldr.String())
			if err == nil {
				continue
			}

			// A refused URL is a config problem, not a failed delivery.
			var vErr *notifier.ValidationError
			if errors.As(err, &vErr) {
				log.Error().
					Err(vErr.Err).
					Str("notifier_id", notifiers[i].ID.String()).
					Str("notifier_name", notifiers[i].Name).
					Msg("notifier URL failed validation, skipping")
				sendErrs = append(sendErrs, fmt.Errorf("notifier %s failed validation: %w", notifiers[i].Name, vErr.Err))
				continue
			}

			sendErrs = append(sendErrs, err)
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
