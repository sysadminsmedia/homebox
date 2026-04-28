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
	"gocloud.dev/blob"
	"gocloud.dev/pubsub"
)

func registerRecurringTasks(app *app, cfg *config.Config, runner *graceful.Runner) {
	runner.AddFunc("eventbus", app.bus.Run)

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

	runner.AddPlugin(NewTask("purge-stale-exports", 24*time.Hour, func(ctx context.Context) {
		// Drop export rows and their blob artifacts after a week — long enough
		// for users to re-download a backup, short enough to not pile up.
		cutoff := time.Now().Add(-7 * 24 * time.Hour)
		dropped, err := app.repos.Exports.PurgeOlderThan(ctx, cutoff)
		if err != nil {
			log.Err(err).Msg("failed to purge stale exports")
			return
		}
		if len(dropped) == 0 {
			return
		}
		bucket, err := blob.OpenBucket(ctx, app.repos.Attachments.GetConnString())
		if err != nil {
			log.Err(err).Msg("export cleanup: failed to open bucket; rows dropped, blobs leaked")
			return
		}
		defer func() { _ = bucket.Close() }()
		for _, e := range dropped {
			if e.ArtifactPath == "" {
				continue
			}
			if err := bucket.Delete(ctx, app.repos.Attachments.GetFullPath(e.ArtifactPath)); err != nil {
				log.Warn().Err(err).Str("artifact_path", e.ArtifactPath).Msg("export cleanup: failed to delete blob")
			}
		}
		log.Info().Int("count", len(dropped)).Msg("purged stale collection exports")
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

	runner.AddFunc("collection-export-subscription", func(ctx context.Context) error {
		return runJobSubscription(ctx, cfg, "collection_export", func(ctx context.Context, msg *pubsub.Message) {
			gid, err := uuid.Parse(msg.Metadata["group_id"])
			if err != nil {
				log.Err(err).Str("group_id", msg.Metadata["group_id"]).Msg("export job: bad group_id")
				return
			}
			exportID, err := uuid.Parse(msg.Metadata["export_id"])
			if err != nil {
				log.Err(err).Str("export_id", msg.Metadata["export_id"]).Msg("export job: bad export_id")
				return
			}
			app.services.Exports.RunExport(ctx, exportID, gid)
		})
	})

	runner.AddFunc("collection-import-subscription", func(ctx context.Context) error {
		return runJobSubscription(ctx, cfg, "collection_import", func(ctx context.Context, msg *pubsub.Message) {
			gid, err := uuid.Parse(msg.Metadata["group_id"])
			if err != nil {
				log.Err(err).Str("group_id", msg.Metadata["group_id"]).Msg("import job: bad group_id")
				return
			}
			userID, err := uuid.Parse(msg.Metadata["user_id"])
			if err != nil {
				log.Err(err).Str("user_id", msg.Metadata["user_id"]).Msg("import job: bad user_id")
				return
			}
			uploadKey := msg.Metadata["upload_key"]
			if uploadKey == "" {
				log.Error().Msg("import job: missing upload_key")
				return
			}
			app.services.Exports.RunImport(ctx, gid, userID, uploadKey)
		})
	})

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

// runJobSubscription opens a pubsub topic+subscription pair for the given
// topic name and runs handler for each received message. Mirrors the
// thumbnail subscriber's lifecycle: shut down topic and subscription when
// ctx ends; ack every message regardless of handler outcome (no redelivery).
func runJobSubscription(ctx context.Context, cfg *config.Config, topicName string, handler func(context.Context, *pubsub.Message)) error {
	conn, err := utils.GenerateSubPubConn(cfg.Database.PubSubConnString, topicName)
	if err != nil {
		log.Err(err).Str("topic", topicName).Msg("failed to generate pubsub connection string")
		return err
	}
	topic, err := pubsub.OpenTopic(ctx, conn)
	if err != nil {
		return err
	}
	defer func() {
		if err := topic.Shutdown(ctx); err != nil {
			log.Err(err).Str("topic", topicName).Msg("failed to shutdown pubsub topic")
		}
	}()

	sub, err := pubsub.OpenSubscription(ctx, conn)
	if err != nil {
		log.Err(err).Str("topic", topicName).Msg("failed to open pubsub subscription")
		return err
	}
	defer func() {
		if err := sub.Shutdown(ctx); err != nil {
			log.Err(err).Str("topic", topicName).Msg("failed to shutdown pubsub subscription")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := sub.Receive(ctx)
			if err != nil {
				log.Err(err).Str("topic", topicName).Msg("failed to receive message from pubsub topic")
				continue
			}
			if msg == nil {
				continue
			}
			handler(ctx, msg)
			msg.Ack()
		}
	}
}
