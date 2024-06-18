package services

import (
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
