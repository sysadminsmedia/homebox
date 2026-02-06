package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/graceful"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/utils"
	"gocloud.dev/pubsub"
)

func registerRecurringTasks(app *app, cfg *config.Config, runner *graceful.Runner) {
	runner.AddFunc("eventbus", app.bus.Run)

	if app.broker != nil {
		runner.AddFunc("centrifuge", app.broker.Run)
	}

	runner.AddFunc("seed_database", func(ctx context.Context) error {
		if cfg.Demo {
			log.Info().Msg("Running in demo mode, creating demo data")
			err := app.SetupDemo()
			if err != nil {
				log.Error().Err(err).Msg("failed to setup demo data")
				return fmt.Errorf("failed to setup demo data: %w", err)
			}
		}
		return nil
	})

	runner.AddPlugin(NewTask("purge-tokens", 24*time.Hour, func(ctx context.Context) {
		_, err := app.repos.AuthTokens.PurgeExpiredTokens(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to purge expired tokens")
		}
	}))

	runner.AddPlugin(NewTask("purge-invitations", 24*time.Hour, func(ctx context.Context) {
		_, err := app.repos.Groups.InvitationPurge(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to purge expired invitations")
		}
	}))

	runner.AddPlugin(NewTask("send-notifications", time.Hour, func(ctx context.Context) {
		now := time.Now()
		if now.Hour() == 8 {
			fmt.Println("run notifiers")
			err := app.services.BackgroundService.SendNotifiersToday(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("failed to send notifiers")
			}
		}
	}))

	if cfg.Thumbnail.Enabled {
		runner.AddFunc("create-thumbnails-subscription", func(ctx context.Context) error {
			pubsubString, err := utils.GenerateSubPubConn(cfg.Database.PubSubConnString, "thumbnails")
			if err != nil {
				log.Error().Err(err).Msg("failed to generate pubsub connection string")
				return err
			}
			topic, err := pubsub.OpenTopic(ctx, pubsubString)
			if err != nil {
				return err
			}
			defer func(topic *pubsub.Topic, ctx context.Context) {
				err := topic.Shutdown(ctx)
				if err != nil {
					log.Err(err).Msg("fail to shutdown pubsub topic")
				}
			}(topic, ctx)

			subscription, err := pubsub.OpenSubscription(ctx, pubsubString)
			if err != nil {
				log.Err(err).Msg("failed to open pubsub topic")
				return err
			}
			defer func(topic *pubsub.Subscription, ctx context.Context) {
				err := topic.Shutdown(ctx)
				if err != nil {
					log.Err(err).Msg("fail to shutdown pubsub topic")
				}
			}(subscription, ctx)

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					msg, err := subscription.Receive(ctx)
					log.Debug().Msg("received thumbnail generation request from pubsub topic")
					if err != nil {
						log.Err(err).Msg("failed to receive message from pubsub topic")
						continue
					}
					if msg == nil {
						log.Warn().Msg("received nil message from pubsub topic")
						continue
					}
					groupId, err := uuid.Parse(msg.Metadata["group_id"])
					if err != nil {
						log.Error().Err(err).Str("group_id", msg.Metadata["group_id"]).Msg("failed to parse group ID from message metadata")
					}
					attachmentId, err := uuid.Parse(msg.Metadata["attachment_id"])
					if err != nil {
						log.Error().Err(err).Str("attachment_id", msg.Metadata["attachment_id"]).Msg("failed to parse attachment ID from message metadata")
					}
					err = app.repos.Attachments.CreateThumbnail(ctx, groupId, attachmentId, msg.Metadata["title"], msg.Metadata["path"])
					if err != nil {
						log.Err(err).Msg("failed to create thumbnail")
					}
					msg.Ack()
				}
			}
		})
	}

	if cfg.Options.GithubReleaseCheck {
		runner.AddPlugin(NewTask("get-latest-github-release", time.Hour, func(ctx context.Context) {
			log.Debug().Msg("running get latest github release")
			err := app.services.BackgroundService.GetLatestGithubRelease(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("failed to get latest github release")
			}
		}))
	}

	if cfg.Debug.Enabled {
		runner.AddFunc("debug", func(ctx context.Context) error {
			debugserver := http.Server{
				Addr:         fmt.Sprintf("%s:%s", cfg.Web.Host, cfg.Debug.Port),
				Handler:      app.debugRouter(),
				ReadTimeout:  cfg.Web.ReadTimeout,
				WriteTimeout: cfg.Web.WriteTimeout,
				IdleTimeout:  cfg.Web.IdleTimeout,
			}

			go func() {
				<-ctx.Done()
				_ = debugserver.Shutdown(context.Background())
			}()

			log.Info().Msgf("Debug server is running on %s:%s", cfg.Web.Host, cfg.Debug.Port)
			return debugserver.ListenAndServe()
		})
		// Print the configuration to the console
		cfg.Print()
	}
}
