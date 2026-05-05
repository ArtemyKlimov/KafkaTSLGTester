package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/IBM/sarama"
	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"kafkatsgltest/internal/config"
)

// NewSaramaConfig builds a sarama.Config from the application KafkaConfig.
func NewSaramaConfig(cfg config.KafkaConfig) (*sarama.Config, error) {
	sc := sarama.NewConfig()
	sc.Version = parseKafkaVersion(cfg.KafkaVersion)
	sc.Producer.Return.Successes = false
	sc.Producer.Return.Errors = true
	sc.Producer.RequiredAcks = sarama.WaitForLocal
	sc.Producer.Compression = sarama.CompressionNone

	// Кастомный диалер: заменяет hostname на IP из host_aliases.
	// Решает проблему с .local-доменами, не резолвящимися через стандартный DNS.
	if len(cfg.HostAliases) > 0 {
		sc.Net.Proxy.Enable = true
		sc.Net.Proxy.Dialer = &aliasDialer{aliases: cfg.HostAliases}
	}

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

		// Принудительно отправляем сертификат при любом запросе сервера,
		// даже если CA не совпадает с AcceptableCAs (Go по умолчанию может пропустить отправку).
		tlsCfg.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			slog.Debug("TLS: server requested client certificate",
				"acceptable_CAs", len(info.AcceptableCAs),
				"our_cert_CN", cert.Leaf.Subject.CommonName)
			return &tlsCfg.Certificates[0], nil
		}
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

	slog.Debug("loaded keystore certificate",
		"subject", cert.Subject.String(),
		"issuer", cert.Issuer.String(),
		"valid_until", cert.NotAfter.Format("2006-01-02"),
		"has_private_key", privKey != nil,
		"ca_chain_len", len(caCerts))

	if privKey == nil {
		return tls.Certificate{}, fmt.Errorf("keystore %q: private key is missing", path)
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

// aliasDialer реализует proxy.Dialer, подменяя hostname на IP из host_aliases
// перед установкой TCP-соединения.
type aliasDialer struct {
	aliases map[string]string
}

func (d *aliasDialer) Dial(network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if ip, ok := d.aliases[host]; ok {
		addr = net.JoinHostPort(ip, port)
	}
	return (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).Dial(network, addr)
}

// parseKafkaVersion разбирает строку версии из конфига.
// Если версия не указана или не распознана, возвращает V2_1_0_0 —
// это последняя версия без flexible-encoding в ApiVersionsRequest,
// что исключает лишний round-trip при подключении к большинству брокеров.
func parseKafkaVersion(v string) sarama.KafkaVersion {
	if v == "" {
		return sarama.V2_1_0_0
	}
	version, err := sarama.ParseKafkaVersion(v)
	if err != nil {
		slog.Warn("unknown kafka_version in config, using 2.1.0", "value", v)
		return sarama.V2_1_0_0
	}
	return version
}
