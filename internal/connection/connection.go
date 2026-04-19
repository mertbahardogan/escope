package connection

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/elastic"
	"sync"
	"time"
)

type Config struct {
	Host     string
	Username string
	Password string
	Secure   bool
}

var (
	once   sync.Once
	client *elasticsearch.Client
	conf   Config
)

func SetConfig(c Config) {
	conf = c
	once = sync.Once{}
	client = nil
}

// CurrentHost returns the host URL for the active connection (empty if unset).
func CurrentHost() string {
	return conf.Host
}

func ClearConfig() {
	conf = Config{}
	once = sync.Once{}
	client = nil
}

func LoadConfigFromFile(alias string) error {
	cfg, err := config.LoadHost(alias)
	if err != nil {
		return err
	}
	SetConfig(Config(cfg))
	return nil
}

func GetSavedConfig(alias string) Config {
	cfg, err := config.LoadHost(alias)
	if err != nil {
		return Config{}
	}
	return Config(cfg)
}

func ListSavedConfigs() ([]string, error) {
	return config.ListHosts()
}

func GetActiveHost() (string, error) {
	return config.GetActiveHost()
}

func GetClient() *elasticsearch.Client {
	if conf.Host == "" {
		aliases, err := ListSavedConfigs()
		if err != nil || len(aliases) == 0 {
			return nil
		}
		_ = LoadConfigFromFile(aliases[0])
	}

	if conf.Host == "" {
		return nil
	}

	once.Do(func() {
		client = elastic.NewClient(conf.Host, conf.Username, conf.Password, conf.Secure)
	})
	return client
}

func TestConnection(cfg Config, timeoutSeconds int) error {
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}

	tempClient := elastic.NewClient(cfg.Host, cfg.Username, cfg.Password, cfg.Secure)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	res, err := tempClient.Ping(tempClient.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("connection failed with status: %s", res.Status())
	}

	return nil
}
