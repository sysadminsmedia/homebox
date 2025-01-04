package services

import (
	"context"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

type BackgroundService struct {
	repos *repo.AllRepos
}

func (svc *BackgroundService) SendNotifiersToday(ctx context.Context) error {
	// Get All Groups
	groups, err := svc.repos.Groups.GetAllGroups(ctx)
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

		urls := make([]string, len(notifiers))
		for i := range notifiers {
			urls[i] = notifiers[i].URL
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

func (svc *BackgroundService) UpdateLocales(ctx context.Context) error {
	log.Debug().Msg("updating locales")
	// fetch list of locales from github
	// is it worth checking if any changes have been made?
	// download locales overwriting files in static/public/locales

	// curl -H "Accept: application/vnd.github.v3+json" \
  //    -H "If-Modified-Since: Thu, 31 Oct 2024 09:59:02 GMT" \
  //    -o /dev/null -s -w "%{http_code}\n" \
  //    https://api.github.com/repos/sysadminsmedia/homebox/contents/frontend/locales
  // keep track of last modified date

	

	return nil
}