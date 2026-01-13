package config

import (
	"github.com/mertbahardogan/escope/internal/constants"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type ConnectionConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Secure   bool   `yaml:"secure"`
}

type AppConfig struct {
	ConnectionTimeout int    `yaml:"connection_timeout"`
	InstalledVersion  string `yaml:"installed_version,omitempty"`
}

type HostConfig struct {
	Config     AppConfig                   `yaml:"config"`
	Hosts      map[string]ConnectionConfig `yaml:"hosts"`
	ActiveHost string                      `yaml:"active_host,omitempty"`
}

func configFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, constants.ConfigFilePath)
}

func Save(cfg HostConfig) error {
	f, err := os.Create(configFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	return enc.Encode(cfg)
}

func Load() (HostConfig, error) {
	var cfg HostConfig
	f, err := os.Open(configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return HostConfig{
				Config: AppConfig{ConnectionTimeout: constants.DefaultConfigTimeout},
				Hosts:  make(map[string]ConnectionConfig),
			}, nil
		}
		return cfg, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]ConnectionConfig)
	}

	if cfg.Config.ConnectionTimeout == 0 {
		cfg.Config.ConnectionTimeout = constants.DefaultConfigTimeout2
	}

	return cfg, nil
}

func SaveHost(alias string, connCfg ConnectionConfig) error {
	hostCfg, err := Load()
	if err != nil {
		hostCfg = HostConfig{
			Hosts: make(map[string]ConnectionConfig),
		}
	}
	if len(hostCfg.Hosts) == 0 {
		hostCfg.ActiveHost = alias
	}
	hostCfg.Hosts[alias] = connCfg
	return Save(hostCfg)
}

func LoadHost(alias string) (ConnectionConfig, error) {
	hostCfg, err := Load()
	if err != nil {
		return ConnectionConfig{}, err
	}

	connCfg, exists := hostCfg.Hosts[alias]
	if !exists {
		return ConnectionConfig{}, os.ErrNotExist
	}

	return connCfg, nil
}

func ListHosts() ([]string, error) {
	hostCfg, err := Load()
	if err != nil {
		return nil, err
	}

	var aliases []string
	for alias := range hostCfg.Hosts {
		aliases = append(aliases, alias)
	}

	return aliases, nil
}

func DeleteHost(alias string) error {
	hostCfg, err := Load()
	if err != nil {
		return err
	}

	if _, exists := hostCfg.Hosts[alias]; !exists {
		return os.ErrNotExist
	}

	delete(hostCfg.Hosts, alias)
	if hostCfg.ActiveHost == alias {
		hostCfg.ActiveHost = constants.EmptyString
	}

	return Save(hostCfg)
}

func SetActiveHost(alias string) error {
	_, err := LoadHost(alias)
	if err != nil {
		return err
	}
	hostCfg, err := Load()
	if err != nil {
		return err
	}

	hostCfg.ActiveHost = alias
	return Save(hostCfg)
}

func GetActiveHost() (string, error) {
	hostCfg, err := Load()
	if err != nil {
		return constants.EmptyString, err
	}
	return hostCfg.ActiveHost, nil
}

func ClearActiveHost() error {
	hostCfg, err := Load()
	if err != nil {
		return err
	}

	hostCfg.ActiveHost = constants.EmptyString
	return Save(hostCfg)
}

func GetAppConfig() (AppConfig, error) {
	hostCfg, err := Load()
	if err != nil {
		return AppConfig{}, err
	}
	return hostCfg.Config, nil
}

func GetConnectionTimeout() (int, error) {
	appCfg, err := GetAppConfig()
	if err != nil {
		return constants.DefaultTimeout, err
	}
	return appCfg.ConnectionTimeout, nil
}

func SetConnectionTimeout(timeout int) error {
	hostCfg, err := Load()
	if err != nil {
		return err
	}

	hostCfg.Config.ConnectionTimeout = timeout
	return Save(hostCfg)
}

func GetInstalledVersion() (string, error) {
	appCfg, err := GetAppConfig()
	if err != nil {
		return "", err
	}
	return appCfg.InstalledVersion, nil
}

func SetInstalledVersion(version string) error {
	hostCfg, err := Load()
	if err != nil {
		return err
	}

	hostCfg.Config.InstalledVersion = version
	return Save(hostCfg)
}
