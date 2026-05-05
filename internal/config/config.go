package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Kafka       KafkaConfig   `yaml:"kafka"`
	Defaults    Defaults      `yaml:"defaults"`
	RandomWords []string      `yaml:"random_words"`
	Blocks      []BlockConfig `yaml:"blocks"`
}

type KafkaConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Brokers  []string `yaml:"brokers"`
	Topic    string   `yaml:"topic"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	Secure   bool     `yaml:"secure"`

	TLSSkipVerify         bool   `yaml:"tls_skip_verify"`
	TLSKeystoreFile       string `yaml:"tls_keystore_file"`
	TLSKeystorePassword   string `yaml:"tls_keystore_password"`
	TLSTruststoreFile     string `yaml:"tls_truststore_file"`
	TLSTruststorePassword string `yaml:"tls_truststore_password"`

	// HostAliases позволяет задать маппинг hostname→IP для брокеров,
	// которые не резолвятся через стандартный DNS (например, .local-домены).
	HostAliases map[string]string `yaml:"host_aliases"`
}

func (k KafkaConfig) BrokerAddresses() []string {
	if len(k.Brokers) > 0 {
		return k.Brokers
	}
	host := k.Host
	if host == "" {
		host = "localhost"
	}
	port := k.Port
	if port == 0 {
		port = 9092
	}
	return []string{fmt.Sprintf("%s:%d", host, port)}
}

type Defaults struct {
	BatchSize int `yaml:"batch_size"`
	Workers   int `yaml:"workers"`
}

type BlockConfig struct {
	Count     int            `yaml:"count"`
	Key       string         `yaml:"key"`
	Topic     string         `yaml:"topic"`
	BatchSize int            `yaml:"batch_size"`
	Workers   int            `yaml:"workers"`
	Fields    map[string]any `yaml:"fields"`
}

func (b BlockConfig) EffectiveBatchSize(d Defaults) int {
	if b.BatchSize > 0 {
		return b.BatchSize
	}
	if d.BatchSize > 0 {
		return d.BatchSize
	}
	return 1
}

func (b BlockConfig) EffectiveWorkers(d Defaults) int {
	if b.Workers > 0 {
		return b.Workers
	}
	if d.Workers > 0 {
		return d.Workers
	}
	return 1
}

func Load(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *AppConfig) validate() error {
	if len(c.Blocks) == 0 {
		return fmt.Errorf("config: at least one block is required")
	}
	for i, b := range c.Blocks {
		if b.Count <= 0 {
			return fmt.Errorf("config: block[%d]: count must be > 0", i)
		}
		if len(b.Fields) == 0 {
			return fmt.Errorf("config: block[%d]: fields must not be empty", i)
		}
	}
	return nil
}
