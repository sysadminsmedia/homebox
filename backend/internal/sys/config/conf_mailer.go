package config

import "encoding/json"

type MailerConf struct {
	Host     string `conf:""`
	Port     int    `conf:""`
	Username string `conf:""`
	Password string `conf:""`
	From     string `conf:""`
}

func (m MailerConf) MarshalJSON() ([]byte, error) {
	type alias MailerConf
	a := alias(m)
	if a.Password != "" {
		a.Password = redactedValue
	}
	return json.Marshal(a)
}

// Ready is a simple check to ensure that the configuration is not empty.
// or with it's default state.
func (mc *MailerConf) Ready() bool {
	return mc.Host != "" && mc.Port != 0 && mc.Username != "" && mc.Password != "" && mc.From != ""
}
