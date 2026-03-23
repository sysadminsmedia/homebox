// Package config provides the configuration for the application.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/conf/v3"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

type Config struct {
	conf.Version
	Mode       string         `yaml:"mode"       conf:"default:development"` // development or production
	Web        WebConfig      `yaml:"web"`
	Storage    Storage        `yaml:"storage"`
	Database   Database       `yaml:"database"`
	Log        LoggerConf     `yaml:"logger"`
	Mailer     MailerConf     `yaml:"mailer"`
	Demo       bool           `yaml:"demo"`
	Debug      DebugConf      `yaml:"debug"`
	Options    Options        `yaml:"options"`
	OIDC       OIDCConf       `yaml:"oidc"`
	LabelMaker LabelMakerConf `yaml:"labelmaker"`
	Thumbnail  Thumbnail      `yaml:"thumbnail"`
	Barcode    BarcodeAPIConf `yaml:"barcode"`
	Auth       AuthConfig     `yaml:"auth"`
	Notifier   NotifierConf   `yaml:"notifier"`
}

type Options struct {
	AllowRegistration    bool   `yaml:"disable_registration"    conf:"default:true"`
	AutoIncrementAssetID bool   `yaml:"auto_increment_asset_id" conf:"default:true"`
	CurrencyConfig       string `yaml:"currencies"`
	GithubReleaseCheck   bool   `yaml:"check_github_release"    conf:"default:true"`
	AllowAnalytics       bool   `yaml:"allow_analytics"         conf:"default:false"`
	AllowLocalLogin      bool   `yaml:"allow_local_login"       conf:"default:true"`
	TrustProxy           bool   `yaml:"trust_proxy"             conf:"default:false"`
	Hostname             string `yaml:"hostname"`
}

type Thumbnail struct {
	Enabled bool `yaml:"enabled" conf:"default:true"`
	Width   int  `yaml:"width"   conf:"default:500"`
	Height  int  `yaml:"height"  conf:"default:500"`
}

type DebugConf struct {
	Enabled bool   `yaml:"enabled" conf:"default:false"`
	Port    string `yaml:"port"    conf:"default:4000"`
}

type WebConfig struct {
	Port          string        `yaml:"port"            conf:"default:7745"`
	Host          string        `yaml:"host"`
	MaxUploadSize int64         `yaml:"max_file_upload" conf:"default:10"`
	ReadTimeout   time.Duration `yaml:"read_timeout"    conf:"default:10s"`
	WriteTimeout  time.Duration `yaml:"write_timeout"   conf:"default:10s"`
	IdleTimeout   time.Duration `yaml:"idle_timeout"    conf:"default:30s"`
}

type LabelMakerConf struct {
	Width                 int64          `yaml:"width"                 conf:"default:526"`
	Height                int64          `yaml:"height"                conf:"default:200"`
	Padding               int64          `yaml:"padding"               conf:"default:32"`
	Margin                int64          `yaml:"margin"                conf:"default:32"`
	FontSize              float64        `yaml:"font_size"             conf:"default:32.0"`
	PrintCommand          *string        `yaml:"string"`
	AdditionalInformation *string        `yaml:"string"`
	DynamicLength         bool           `yaml:"bool"                  conf:"default:true"`
	LabelServiceUrl       *string        `yaml:"label_service_url"`
	LabelServiceTimeout   *time.Duration `yaml:"label_service_timeout"`
	RegularFontPath       *string        `yaml:"regular_font_path"`
	BoldFontPath          *string        `yaml:"bold_font_path"`
}

type OIDCConf struct {
	Enabled            bool          `yaml:"enabled"              conf:"default:false"`
	IssuerURL          string        `yaml:"issuer_url"`
	ClientID           string        `yaml:"client_id"`
	ClientSecret       string        `yaml:"client_secret"`
	Scope              string        `yaml:"scope"                conf:"default:openid profile email"`
	AllowedGroups      string        `yaml:"allowed_groups"`
	AutoRedirect       bool          `yaml:"auto_redirect"        conf:"default:false"`
	VerifyEmail        bool          `yaml:"verify_email"         conf:"default:false"`
	GroupClaim         string        `yaml:"group_claim"          conf:"default:groups"`
	EmailClaim         string        `yaml:"email_claim"          conf:"default:email"`
	NameClaim          string        `yaml:"name_claim"           conf:"default:name"`
	EmailVerifiedClaim string        `yaml:"email_verified_claim" conf:"default:email_verified"`
	ButtonText         string        `yaml:"button_text"          conf:"default:Sign in with OIDC"`
	StateExpiry        time.Duration `yaml:"state_expiry"         conf:"default:10m"`
	RequestTimeout     time.Duration `yaml:"request_timeout"      conf:"default:30s"`
}

type BarcodeAPIConf struct {
	TokenBarcodespider string `yaml:"token_barcodespider"`
}

type AuthConfig struct {
	RateLimit AuthRateLimit `yaml:"rate_limit"`
}

type AuthRateLimit struct {
	Enabled     bool          `yaml:"enabled"      conf:"default:true"`
	Window      time.Duration `yaml:"window"       conf:"default:1m"`
	MaxAttempts int           `yaml:"max_attempts" conf:"default:5"`
	BaseBackoff time.Duration `yaml:"base_backoff" conf:"default:10s"`
	MaxBackoff  time.Duration `yaml:"max_backoff"  conf:"default:5m"`
}

// New parses the CLI/Config file and returns a Config struct. If the file argument is an empty string, the
// file is not read. If the file is not empty, the file is read and the Config struct is returned.
func New(buildstr string, description string) (*Config, error) {
	var cfg Config
	const prefix = "HBOX"

	cfg.Version = conf.Version{
		Build: buildstr,
		Desc:  description,
	}

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return &cfg, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Print prints the configuration to stdout as a json indented string
// This is useful for debugging. If the marshaller errors out, it will panic.
func (c *Config) Print() {
	res, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(res))
}
