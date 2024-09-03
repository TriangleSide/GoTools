package metrics_server

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
	udplistener "github.com/TriangleSide/GoBase/pkg/network/udp/listener"
)

// Config is configured by the caller with the Options functions.
type Config struct {
	configProvider      func() (*config.MetricsServer, error)
	udpListenerProvider func(localHost string, localPort uint16) (udp.Conn, error)
	encryptorProvider   func(key string) (symmetric.Encryptor, error)
	unmarshallingFunc   func(data []byte, v any) error
	errorHandler        func(err error)
}

// Options is used to configure the metrics server.
type Options func(*Config) error

// WithEncryptorProvider overwrites the default encryptor provider.
func WithEncryptorProvider(provider func(key string) (symmetric.Encryptor, error)) Options {
	return func(serverConfig *Config) error {
		serverConfig.encryptorProvider = provider
		return nil
	}
}

// WithUDPListenerProvider overwrites the default UDP listener provider.
func WithUDPListenerProvider(provider func(localHost string, localPort uint16) (udp.Conn, error)) Options {
	return func(serverConfig *Config) error {
		serverConfig.udpListenerProvider = provider
		return nil
	}
}

// WithConfigProvider overwrites the default config provider.
func WithConfigProvider(provider func() (*config.MetricsServer, error)) Options {
	return func(serverConfig *Config) error {
		serverConfig.configProvider = provider
		return nil
	}
}

// WithUnmarshallingFunc overwrites the default unmarshalling function.
func WithUnmarshallingFunc(unmarshallingFunc func(data []byte, v any) error) Options {
	return func(serverConfig *Config) error {
		serverConfig.unmarshallingFunc = unmarshallingFunc
		return nil
	}
}

// WithErrorHandler overwrites the default error handling function.
func WithErrorHandler(errorHandler func(err error)) Options {
	return func(serverConfig *Config) error {
		serverConfig.errorHandler = errorHandler
		return nil
	}
}

// Server represents a metrics server.
type Server struct {
	cfg           *config.MetricsServer
	unmarshalFunc func(data []byte, v any) error
	errorHandler  func(err error)
	enc           symmetric.Encryptor
	conn          udp.Conn
	metrics       chan *metrics.Metric
	wg            sync.WaitGroup
	shutdown      *atomic.Bool
	ran           *atomic.Bool
}

// New creates a new Server instance from a configuration parsed from the environment variables.
func New(opts ...Options) (*Server, error) {
	serverConfig := &Config{
		configProvider: func() (*config.MetricsServer, error) {
			return envprocessor.ProcessAndValidate[config.MetricsServer]()
		},
		udpListenerProvider: udplistener.New,
		encryptorProvider: func(key string) (symmetric.Encryptor, error) {
			return symmetric.New(key)
		},
		unmarshallingFunc: json.Unmarshal,
		errorHandler:      func(error) {},
	}

	for _, opt := range opts {
		if err := opt(serverConfig); err != nil {
			return nil, fmt.Errorf("failed to configure metrics server (%s)", err.Error())
		}
	}

	cfg, err := serverConfig.configProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics server configuration (%s)", err.Error())
	}

	conn, err := serverConfig.udpListenerProvider(cfg.MetricsBindIP, cfg.MetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics connection (%s)", err.Error())
	}

	err = conn.SetReadBuffer(int(cfg.MetricsOSBufferSize))
	if err != nil {
		return nil, fmt.Errorf("failed to set the read buffer size to %d (%s)", cfg.MetricsOSBufferSize, err.Error())
	}

	encryptor, err := serverConfig.encryptorProvider(cfg.MetricsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create the encryptor (%s)", err.Error())
	}

	shutdownFlag := &atomic.Bool{}
	shutdownFlag.Store(false)

	ran := &atomic.Bool{}
	ran.Store(false)

	return &Server{
		cfg:           cfg,
		unmarshalFunc: serverConfig.unmarshallingFunc,
		errorHandler:  serverConfig.errorHandler,
		enc:           encryptor,
		conn:          conn,
		metrics:       make(chan *metrics.Metric, cfg.MetricsQueueSize),
		wg:            sync.WaitGroup{},
		shutdown:      shutdownFlag,
		ran:           ran,
	}, nil
}

// Run starts the metrics server and launches some processing go routines.
// Received metrics are sent over the metrics channel returned by this function.
func (server *Server) Run() chan *metrics.Metric {
	if server.ran.Swap(true) {
		panic("metrics server can only be run once per instance")
	}

	for threadIndex := 0; threadIndex < server.cfg.MetricsReadThreads; threadIndex++ {
		server.wg.Add(1)

		go func() {
			defer func() {
				server.wg.Done()
				if err := server.Shutdown(); err != nil {
					server.errorHandler(err)
				}
			}()

			readBuffer := make([]byte, server.cfg.MetricsReadBufferSize)

			for !server.shutdown.Load() {
				amountRead, err := server.conn.Read(readBuffer)
				if err != nil {
					if !server.shutdown.Load() {
						server.errorHandler(err)
					}
					break
				}

				if amountRead >= len(readBuffer) {
					server.errorHandler(errors.New("data received from the socket is too large"))
					continue
				}

				jsonBytes, err := server.enc.Decrypt(readBuffer[:amountRead])
				if err != nil {
					server.errorHandler(fmt.Errorf("failed to decrypt the data from the socket (%s)", err.Error()))
					continue
				}

				var metricsList []*metrics.Metric
				err = server.unmarshalFunc(jsonBytes, &metricsList)
				if err != nil {
					server.errorHandler(fmt.Errorf("failed to unmarshall the data from the socket (%s)", err.Error()))
					continue
				}

				for _, metric := range metricsList {
					select {
					case server.metrics <- metric:
					default:
						server.errorHandler(errors.New("metric queue is full"))
					}
				}
			}
		}()
	}

	return server.metrics
}

// Running returns whether the server has any go routines running.
func (server *Server) Running() bool {
	return !server.shutdown.Load()
}

// Shutdown stops the metrics server, and waits for the go routines to finish.
// This function can be called concurrently, but only the first will perform the shutdown action.
func (server *Server) Shutdown() error {
	if !server.shutdown.Swap(true) {
		err := server.conn.Close()
		server.wg.Wait()
		close(server.metrics)
		return err
	}
	return nil
}
