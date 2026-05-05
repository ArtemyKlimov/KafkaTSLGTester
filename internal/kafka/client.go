package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	"golang.org/x/crypto/pkcs12"

	"kafkatsgltest/internal/config"
)

// NewSaramaConfig builds a sarama.Config from the application KafkaConfig.
func NewSaramaConfig(cfg config.KafkaConfig) (*sarama.Config, error) {
	sc := sarama.NewConfig()
	sc.Version = sarama.V2_8_0_0
	sc.Producer.Return.Successes = false
	sc.Producer.Return.Errors = true
	sc.Producer.RequiredAcks = sarama.WaitForLocal
	sc.Producer.Compression = sarama.CompressionNone

	if cfg.Secure {
		tlsCfg, err := buildTLS(cfg)
		if err != nil {
			return nil, err
		}
		sc.Net.TLS.Enable = true
		sc.Net.TLS.Config = tlsCfg

		// SASL/PLAIN — только поверх TLS и только если заданы оба поля
		if cfg.User != "" && cfg.Password != "" {
			sc.Net.SASL.Enable = true
			sc.Net.SASL.Mechanism = sarama.SASLTypePlaintext
			sc.Net.SASL.User = cfg.User
			sc.Net.SASL.Password = cfg.Password
			sc.Net.SASL.Handshake = true
		}
	}

	return sc, nil
}

func buildTLS(cfg config.KafkaConfig) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.TLSSkipVerify, //nolint:gosec
	}

	// Truststore — CA-сертификаты для проверки сертификата брокера
	if cfg.TLSTruststoreFile != "" {
		pool, err := loadTruststore(cfg.TLSTruststoreFile, cfg.TLSTruststorePassword)
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = pool
	}

	// Keystore — клиентский сертификат + ключ (mTLS)
	if cfg.TLSKeystoreFile != "" {
		cert, err := loadKeystore(cfg.TLSKeystoreFile, cfg.TLSKeystorePassword)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}

// loadTruststore читает PKCS12-файл и возвращает пул CA-сертификатов.
func loadTruststore(path, password string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading truststore %q: %w", path, err)
	}
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return nil, fmt.Errorf("decoding truststore %q: %w", path, err)
	}
	pool := x509.NewCertPool()
	for _, b := range blocks {
		if b.Type == "CERTIFICATE" {
			pool.AppendCertsFromPEM(pem.EncodeToMemory(b))
		}
	}
	return pool, nil
}

// loadKeystore читает PKCS12-файл и возвращает клиентский TLS-сертификат.
func loadKeystore(path, password string) (tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("reading keystore %q: %w", path, err)
	}
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("decoding keystore %q: %w", path, err)
	}
	var certPEM, keyPEM []byte
	for _, b := range blocks {
		switch b.Type {
		case "CERTIFICATE":
			certPEM = append(certPEM, pem.EncodeToMemory(b)...)
		case "PRIVATE KEY":
			keyPEM = append(keyPEM, pem.EncodeToMemory(b)...)
		}
	}
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return tls.Certificate{}, fmt.Errorf("keystore %q: missing certificate or private key", path)
	}
	return tls.X509KeyPair(certPEM, keyPEM)
}
