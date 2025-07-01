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
	LabelMaker LabelMakerConf `yaml:"labelmaker"`
	Thumbnail  Thumbnail      `yaml:"thumbnail"`
	Barcode    BarcodeAPIConf `yaml:"barcode"`
}

type Options struct {
	AllowRegistration    bool   `yaml:"disable_registration"    conf:"default:true"`
	AutoIncrementAssetID bool   `yaml:"auto_increment_asset_id" conf:"default:true"`
	CurrencyConfig       string `yaml:"currencies"`
	GithubReleaseCheck   bool   `yaml:"check_github_release"    conf:"default:true"`
	AllowAnalytics       bool   `yaml:"allow_analytics"         conf:"default:false"`
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
	Width                 int64   `yaml:"width"     conf:"default:526"`
	Height                int64   `yaml:"height"    conf:"default:200"`
	Padding               int64   `yaml:"padding"   conf:"default:32"`
	Margin                int64   `yaml:"margin"    conf:"default:32"`
	FontSize              float64 `yaml:"font_size" conf:"default:32.0"`
	PrintCommand          *string `yaml:"string"`
	AdditionalInformation *string `yaml:"string"`
	DynamicLength         bool    `yaml:"bool"      conf:"default:true"`
}

type BarcodeAPIConf struct {
	TokenBarcodespider string `yaml:"token_barcodespider"`
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
