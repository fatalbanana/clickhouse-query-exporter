package envcfg

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

var (
	Cfg EnvConfig
)

type (
	EnvConfig struct {
		BindAddr string `split_words:"true"`
		BindPort int    `default:"37825" split_words:"true"`
		Debug    bool
		DSNMap   StringStringMap `split_words:"true" required:"true"`
		QueryMap StringStringMap `split_words:"true" required:"true"`
	}

	StringStringMap map[string]string
)

func (s *StringStringMap) Decode(value string) error {
	return json.Unmarshal([]byte(value), s)
}

func init() {
	err := envconfig.Process("clickhousequeryexporter", &Cfg)
	if err != nil {
		panic(err)
	}
}
