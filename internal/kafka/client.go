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

	if cfg.Secure || cfg.TLSPfxFile != "" {
		tlsCfg, err := buildTLS(cfg)
		if err != nil {
			return nil, err
		}
		sc.Net.TLS.Enable = true
		sc.Net.TLS.Config = tlsCfg
	}

	if cfg.User != "" {
		sc.Net.SASL.Enable = true
		sc.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		sc.Net.SASL.User = cfg.User
		sc.Net.SASL.Password = cfg.Password
		sc.Net.SASL.Handshake = true
	}

	return sc, nil
}

func buildTLS(cfg config.KafkaConfig) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.TLSSkipVerify, //nolint:gosec
	}

	if cfg.TLSPfxFile == "" {
		return tlsCfg, nil
	}

	data, err := os.ReadFile(cfg.TLSPfxFile)
	if err != nil {
		return nil, fmt.Errorf("reading PFX file %q: %w", cfg.TLSPfxFile, err)
	}

	pemBlocks, err := pkcs12.ToPEM(data, cfg.TLSPfxPassword)
	if err != nil {
		return nil, fmt.Errorf("decoding PFX: %w", err)
	}

	var certPEM, keyPEM []byte
	for _, b := range pemBlocks {
		switch b.Type {
		case "CERTIFICATE":
			certPEM = append(certPEM, pem.EncodeToMemory(b)...)
		case "PRIVATE KEY":
			keyPEM = append(keyPEM, pem.EncodeToMemory(b)...)
		}
	}

	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, fmt.Errorf("PFX file %q: missing certificate or private key", cfg.TLSPfxFile)
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("building TLS certificate: %w", err)
	}
	tlsCfg.Certificates = []tls.Certificate{tlsCert}

	if len(tlsCert.Certificate) > 0 {
		leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
		if err == nil {
			pool := x509.NewCertPool()
			pool.AddCert(leaf)
			tlsCfg.RootCAs = pool
		}
	}

	return tlsCfg, nil
}
