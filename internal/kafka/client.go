package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	pkcs12 "software.sslmate.com/src/go-pkcs12"

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

	// Keystore — клиентский сертификат + ключ (mTLS, опционально)
	if cfg.TLSKeystoreFile != "" {
		cert, err := loadKeystore(cfg.TLSKeystoreFile, cfg.TLSKeystorePassword)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}

// loadTruststore читает PKCS12 truststore и возвращает пул CA-сертификатов.
// Поддерживает Java-style truststore (только сертификаты, без ключей).
func loadTruststore(path, password string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading truststore %q: %w", path, err)
	}

	certs, err := pkcs12.DecodeTrustStore(data, password)
	if err != nil {
		return nil, fmt.Errorf("decoding truststore %q: %w", path, err)
	}

	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool, nil
}

// loadKeystore читает PKCS12 keystore и возвращает клиентский TLS-сертификат.
func loadKeystore(path, password string) (tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("reading keystore %q: %w", path, err)
	}

	privKey, cert, caCerts, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("decoding keystore %q: %w", path, err)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privKey,
		Leaf:        cert,
	}
	for _, ca := range caCerts {
		tlsCert.Certificate = append(tlsCert.Certificate, ca.Raw)
	}
	return tlsCert, nil
}
