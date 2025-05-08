package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func defaultLocations() []repo.LocationCreate {
	return []repo.LocationCreate{
		{
			Name: "Living Room",
		},
		{
			Name: "Garage",
		},
		{
			Name: "Kitchen",
		},
		{
			Name: "Bedroom",
		},
		{
			Name: "Bathroom",
		},
		{
			Name: "Office",
		},
		{
			Name: "Attic",
		},
		{
			Name: "Basement",
		},
	}
}

func defaultLabels() []repo.LabelCreate {
	return []repo.LabelCreate{
		{
			Name: "Appliances",
		},
		{
			Name: "IOT",
		},
		{
			Name: "Electronics",
		},
		{
			Name: "Servers",
		},
		{
			Name: "General",
		},
		{
			Name: "Important",
		},
	}
}

func createDefaultLabels(ctx context.Context, repos *repo.AllRepos, groupId uuid.UUID) error {
	log.Debug().Msg("creating default labels")
	for _, label := range defaultLabels() {
		_, err := repos.Labels.Create(ctx, groupId, label)
		if err != nil {
			return err
		}
	}

	log.Debug().Msg("creating default locations")
	for _, location := range defaultLocations() {
		_, err := repos.Locations.Create(ctx, groupId, location)
		if err != nil {
			return err
		}
	}
	return nil
}
