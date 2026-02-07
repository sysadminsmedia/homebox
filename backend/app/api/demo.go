package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
)

func (a *app) SetupDemo() error {
	csvText := `HB.import_ref,HB.location,HB.tags,HB.quantity,HB.name,HB.description,HB.insured,HB.serial_number,HB.model_number,HB.manufacturer,HB.notes,HB.purchase_from,HB.purchase_price,HB.purchase_time,HB.lifetime_warranty,HB.warranty_expires,HB.warranty_details,HB.sold_to,HB.sold_price,HB.sold_time,HB.sold_notes
,Garage,IOT;Home Assistant; Z-Wave,1,Zooz Universal Relay ZEN17,"Zooz 700 Series Z-Wave Universal Relay ZEN17 for Awnings, Garage Doors, Sprinklers, and More | 2 NO-C-NC Relays (20A, 10A) | Signal Repeater | Hub Required (Compatible with SmartThings and Hubitat)",,,ZEN17,Zooz,,Amazon,39.95,10/13/2021,,,,,,,
,Living Room,IOT;Home Assistant; Z-Wave,1,Zooz Motion Sensor,"Zooz Z-Wave Plus S2 Motion Sensor ZSE18 with Magnetic Mount, Works with Vera and SmartThings",,,ZSE18,Zooz,,Amazon,29.95,10/15/2021,,,,,,,
,Office,IOT; Home Assistant; Z-Wave,1,Zooz 110v Power Switch,"Zooz Z-Wave Plus Power Switch ZEN15 for 110V AC Units, Sump Pumps, Humidifiers, and More",,,ZEN15,Zooz,,Amazon,39.95,10/13/2021,,,,,,,
,Downstairs,IOT;Home Assistant; Z-Wave,1,Ecolink Z-Wave PIR Motion Sensor,"Ecolink Z-Wave PIR Motion Detector Pet Immune, White (PIRZWAVE2.5-ECO)",,,PIRZWAVE2.5-ECO,Ecolink,,Amazon,35.58,10/21/2020,,,,,,,
,Entry,IOT;Home Assistant; Z-Wave,1,Yale Security Touchscreen Deadbolt,"Yale Security YRD226-ZW2-619 YRD226ZW2619 Touchscreen Deadbolt, Satin Nickel",,,YRD226ZW2619,Yale,,Amazon,120.39,10/14/2020,,,,,,,
,Kitchen,IOT;Home Assistant; Z-Wave,1,Smart Rocker Light Dimmer,"UltraPro Z-Wave Smart Rocker Light Dimmer with QuickFit and SimpleWire, 3-Way Ready, Compatible with Alexa, Google Assistant, ZWave Hub Required, Repeater/Range Extender, White Paddle Only, 39351",,,39351,Honeywell,,Amazon,65.98,09/30/0202,,,,,,,
`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registration := services.UserRegistration{
		Email:    "demo@example.com",
		Name:     "Demo",
		Password: "demo",
	}

	// If demo user already exists, skip all demo seeding tasks
	if a.services.User.ExistsByEmail(ctx, registration.Email) {
		log.Info().Msg("Demo user already exists; skipping demo seeding")
		return nil
	}

	// Otherwise, register the demo user
	log.Debug().Msg("Registering demo user")
	_, err := a.services.User.RegisterUser(ctx, registration)
	if err != nil {
		if ent.IsConstraintError(err) {
			// Concurrent creation race: treat as exists and skip
			log.Info().Msg("Demo user concurrently created; skipping seeding")
			return nil
		}
		log.Err(err).Msg("Failed to register demo user")
		return errors.New("failed to setup demo")
	}

	// Login the demo user to get a token
	token, err := a.services.User.Login(ctx, registration.Email, registration.Password, false)
	if err != nil {
		log.Err(err).Msg("Failed to login demo user")
		return errors.New("failed to setup demo")
	}
	self, err := a.services.User.GetSelf(ctx, token.Raw)
	if err != nil {
		log.Err(err).Msg("Failed to get self")
		return errors.New("failed to setup demo")
	}

	_, err = a.services.Items.CsvImport(ctx, self.DefaultGroupID, strings.NewReader(csvText))
	if err != nil {
		log.Err(err).Msg("Failed to import CSV")
		return errors.New("failed to setup demo")
	}

	log.Info().Msg("Demo setup complete")

	return nil
}
