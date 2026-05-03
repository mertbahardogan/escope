package connection

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/spf13/cobra"
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

func CurrentHost() string {
	return conf.Host
}

func SessionHostURL() (string, bool) {
	if h := strings.TrimSpace(CurrentHost()); h != "" {
		return h, true
	}
	alias, err := config.GetActiveHost()
	if err != nil || strings.TrimSpace(alias) == "" {
		return "", false
	}
	cfg, err := config.LoadHost(alias)
	if err != nil {
		return "", false
	}
	h := strings.TrimSpace(cfg.Host)
	if h == "" {
		return "", false
	}
	return h, true
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

func NewRecordSamplerClient() *elasticsearch.Client {
	if conf.Host == "" {
		return nil
	}
	return elastic.NewRecordSamplerClient(conf.Host, conf.Username, conf.Password, conf.Secure)
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

func ApplyPersistentConnection(cmd *cobra.Command) error {
	r := cmd.Root()
	aliasFlag, err := r.PersistentFlags().GetString("alias")
	if err != nil {
		return err
	}
	hostFlag, err := r.PersistentFlags().GetString("host")
	if err != nil {
		return err
	}
	usernameFlag, err := r.PersistentFlags().GetString("username")
	if err != nil {
		return err
	}
	passwordFlag, err := r.PersistentFlags().GetString("password")
	if err != nil {
		return err
	}
	secureFlag, err := r.PersistentFlags().GetBool("secure")
	if err != nil {
		return err
	}

	if aliasFlag != constants.EmptyString {
		saved := GetSavedConfig(aliasFlag)
		if saved.Host == constants.EmptyString {
			fmt.Printf("Error: Host alias '%s' not found. Available aliases:\n", aliasFlag)
			aliases, err := ListSavedConfigs()
			if err == nil && len(aliases) > 0 {
				for _, a := range aliases {
					fmt.Printf("  - %s\n", a)
				}
			} else {
				fmt.Println("No hosts configured. Use 'escope config --help' to set up hosts.")
			}
			return fmt.Errorf("host alias '%s' not found", aliasFlag)
		}
		SetConfig(saved)
		return nil
	}

	if hostFlag != constants.EmptyString {
		SetConfig(Config{
			Host:     hostFlag,
			Username: usernameFlag,
			Password: passwordFlag,
			Secure:   secureFlag,
		})
		return nil
	}

	aliases, err := ListSavedConfigs()
	if err != nil || len(aliases) == 0 {
		fmt.Println(constants.ErrNoConfigurationFound)
		fmt.Println(constants.MsgPleaseSetConfiguration)
		fmt.Println(constants.MsgConfigSetExample)
		fmt.Println("")
		fmt.Println(constants.MsgExampleHeader)
		fmt.Println(constants.MsgConfigSetLocalhost)
		fmt.Println(constants.MsgConfigSetSecure)
		fmt.Println("")
		fmt.Println(constants.MsgUseFlagsDirectly)
		fmt.Println(constants.MsgUseFlagsExample)
		return fmt.Errorf("no configuration found")
	}

	activeHost, err := GetActiveHost()
	if err != nil || activeHost == constants.EmptyString {
		fmt.Println("Error: No active host set.")
		fmt.Println("Available hosts:")
		for _, a := range aliases {
			fmt.Printf("  - %s\n", a)
		}
		fmt.Println("")
		fmt.Println("Use 'escope config switch <alias>' to set an active host.")
		return fmt.Errorf("no active host set")
	}

	saved := GetSavedConfig(activeHost)
	if saved.Host == constants.EmptyString {
		fmt.Printf("Error: Active host '%s' not found. Available hosts:\n", activeHost)
		for _, a := range aliases {
			fmt.Printf("  - %s\n", a)
		}
		fmt.Println("")
		fmt.Println("Use 'escope config switch <alias>' to set an active host.")
		return fmt.Errorf("active host '%s' not found", activeHost)
	}
	SetConfig(saved)
	return nil
}
