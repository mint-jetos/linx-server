package template

import (
	"bytes"
	"encoding/json"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/expiry"
	"gabe565.com/linx-server/internal/util"
)

type Config struct {
	SiteName        string           `json:"site_name"`
	SitePath        string           `json:"site_path"`
	MaxSize         int64            `json:"max_size"`
	ForceRandom     bool             `json:"force_random"`
	Auth            bool             `json:"auth"`
	ExpirationTimes []ExpirationTime `json:"expiration_times"`
	CustomPages     []string         `json:"custom_pages,omitzero"`
	Turnstile       Turnstile        `json:"turnstile,omitzero"`
}

type Turnstile struct {
	Enabled bool   `json:"enabled"`
	SiteKey string `json:"site_key"`
}

type ExpirationTime struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewConfig() Config {
	expirationTimes := expiry.ListExpirationTimes()
	conf := Config{
		SiteName:        config.Default.SiteName,
		SitePath:        config.Default.SiteURL.Path,
		ForceRandom:     config.Default.ForceRandomFilename,
		MaxSize:         int64(config.Default.MaxSize),
		Auth:            config.Default.Auth.Basic || config.Default.Auth.File != "",
		ExpirationTimes: make([]ExpirationTime, 0, len(expirationTimes)),
		CustomPages:     config.CustomPages,
		Turnstile: Turnstile{
			Enabled: config.Default.Turnstile.Enabled,
			SiteKey: config.Default.Turnstile.SiteKey,
		},
	}
	for _, t := range expirationTimes {
		conf.ExpirationTimes = append(conf.ExpirationTimes, ExpirationTime{
			Name:  t.Human,
			Value: t.Duration.String(),
		})
	}
	return conf
}

func ConfigBytes() ([]byte, error) {
	var buf bytes.Buffer

	if config.Default.ViteURL == "" {
		f, err := assets.Static().Open(manifest["src/fouc.ts"].File)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = f.Close()
		}()

		if _, err := buf.ReadFrom(f); err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(NewConfig())
	if err != nil {
		return nil, err
	}
	buf.WriteString("window.obfuscatedConfig='")
	buf.WriteString(util.Obfuscate(b))
	buf.WriteString("';")
	return buf.Bytes(), nil
}
