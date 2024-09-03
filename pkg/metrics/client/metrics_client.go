package metrics_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
	"github.com/TriangleSide/GoBase/pkg/crypto/symmetric"
	"github.com/TriangleSide/GoBase/pkg/metrics"
	"github.com/TriangleSide/GoBase/pkg/network/udp"
	udpclient "github.com/TriangleSide/GoBase/pkg/network/udp/client"
)

// Config is configured by the caller with the Option functions.
type Config struct {
	configProvider    func() (*config.MetricsClient, error)
	udpClientProvider func(remoteHost string, remotePort uint16) (udp.Conn, error)
	encryptorProvider func(key string) (symmetric.Encryptor, error)
	marshallingFunc   func(payload any) ([]byte, error)
}

// Option is used to configure the metrics client.
type Option func(serverConfig *Config) error

// WithUDPClientProvider overwrites the UDP client provider.
func WithUDPClientProvider(provider func(remoteHost string, remotePort uint16) (udp.Conn, error)) Option {
	return func(serverConfig *Config) error {
		serverConfig.udpClientProvider = provider
		return nil
	}
}

// WithEncryptorProvider overwrites the encryptor provider.
func WithEncryptorProvider(provider func(key string) (symmetric.Encryptor, error)) Option {
	return func(serverConfig *Config) error {
		serverConfig.encryptorProvider = provider
		return nil
	}
}

// WithMarshallingFunc overwrites the default marshalling func.
func WithMarshallingFunc(marshallingFunc func(payload any) ([]byte, error)) Option {
	return func(serverConfig *Config) error {
		serverConfig.marshallingFunc = marshallingFunc
		return nil
	}
}

// Client represents a client for sending metrics to the server.
type Client struct {
	enc          symmetric.Encryptor
	marshallFunc func(payload any) ([]byte, error)
	cfg          *config.MetricsClient
	conn         udp.Conn
	shutdown     *atomic.Bool
	wg           sync.WaitGroup
}

// New creates a new Client instance from a configuration parsed from the environment variables.
func New(opts ...Option) (*Client, error) {
	clientConfig := &Config{
		configProvider: func() (*config.MetricsClient, error) {
			return envprocessor.ProcessAndValidate[config.MetricsClient]()
		},
		udpClientProvider: func(remoteHost string, remotePort uint16) (udp.Conn, error) {
			return udpclient.New(remoteHost, remotePort)
		},
		encryptorProvider: func(key string) (symmetric.Encryptor, error) {
			return symmetric.New(key)
		},
		marshallingFunc: json.Marshal,
	}

	for _, opt := range opts {
		if err := opt(clientConfig); err != nil {
			return nil, fmt.Errorf("failed to configure metrics client (%s)", err.Error())
		}
	}

	cfg, err := clientConfig.configProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics client configuration (%s)", err.Error())
	}

	conn, err := clientConfig.udpClientProvider(cfg.MetricsHost, cfg.MetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create the metrics client (%s)", err.Error())
	}

	encryptor, err := clientConfig.encryptorProvider(cfg.MetricsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create the encryptor (%s)", err.Error())
	}

	shutdownFlag := &atomic.Bool{}
	shutdownFlag.Store(false)

	return &Client{
		cfg:          cfg,
		marshallFunc: clientConfig.marshallingFunc,
		conn:         conn,
		enc:          encryptor,
		shutdown:     shutdownFlag,
		wg:           sync.WaitGroup{},
	}, nil
}

// Send sends a list of metrics to the server.
// The metrics may fail to make it to the server because the connection is UDP.
// According to https://pkg.go.dev/net#Conn, the UDP conn is thread safe.
func (client *Client) Send(metricsToSend []*metrics.Metric) error {
	client.wg.Add(1)
	defer func() { client.wg.Done() }()

	if client.shutdown.Load() {
		return errors.New("metrics client is closed")
	}

	marshalled, err := client.marshallFunc(metricsToSend)
	if err != nil {
		return fmt.Errorf("failed to marshal the metrics (%s)", err.Error())
	}

	cypher, err := client.enc.Encrypt(marshalled)
	if err != nil {
		return fmt.Errorf("failed to encrypt the metrics (%s)", err.Error())
	}

	n, err := client.conn.Write(cypher)
	if err != nil {
		return fmt.Errorf("failed to send the metrics to the server (%s)", err.Error())
	}
	if n != len(cypher) {
		return fmt.Errorf("only sent %d/%d bytes to the metrics server", n, len(cypher))
	}

	return nil
}

// Close closes the connection to the metrics server.
// Once this is called, the Send function no longer sends metrics to the server.
func (client *Client) Close() error {
	if !client.shutdown.Swap(true) {
		client.wg.Wait()
		return client.conn.Close()
	}
	return nil
}
