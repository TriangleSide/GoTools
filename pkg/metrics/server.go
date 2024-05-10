package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"intelligence/pkg/config"
	"intelligence/pkg/crypto/symmetric"
	"intelligence/pkg/logger"
	netutils "intelligence/pkg/utils/net"
)

// Server represents a metrics server.
type Server struct {
	cfg      *config.MetricsServer
	enc      *symmetric.Encryptor
	conn     *net.UDPConn
	metrics  chan *Metric
	wg       sync.WaitGroup
	shutdown *atomic.Bool
	running  *atomic.Int32
	ran      *atomic.Bool
}

// NewServer creates a new Server instance from a configuration parsed from the environment variables.
func NewServer() (*Server, error) {
	cfg, err := config.ProcessAndValidate[config.MetricsServer]()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics server configuration (%s)", err.Error())
	}

	serverAddr, err := netutils.FormatNetworkAddress(cfg.MetricsBindIP, cfg.MetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to format the metrics server address (%s)", err.Error())
	}

	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the metrics server address (%s)", err.Error())
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial the metrics server address (%s)", err.Error())
	}

	readBufferSize := int(math.Max(float64(cfg.MetricsOSBufferSize), float64(1024*1204*1)))
	err = conn.SetReadBuffer(readBufferSize)
	if err != nil {
		return nil, fmt.Errorf("failed to set the read buffer size to %d (%s)", readBufferSize, err.Error())
	}

	encryptor, err := symmetric.New(cfg.MetricsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create the encryptor (%s)", err.Error())
	}

	shutdownFlag := &atomic.Bool{}
	shutdownFlag.Store(false)

	running := &atomic.Int32{}
	running.Store(0)

	ran := &atomic.Bool{}
	ran.Store(false)

	return &Server{
		cfg:      cfg,
		enc:      encryptor,
		conn:     conn,
		metrics:  make(chan *Metric, cfg.MetricsQueue),
		wg:       sync.WaitGroup{},
		shutdown: shutdownFlag,
		running:  running,
		ran:      ran,
	}, nil
}

// Run starts the metrics server and launches some processing go routines.
// Received metrics are sent over the metrics channel returned by this function.
func (server *Server) Run(ctx context.Context) chan *Metric {
	if server.ran.Swap(true) {
		panic("metrics server can only be run once per instance")
	}

	logEntry := logger.LogEntry(ctx)

	for threadIndex := 0; threadIndex < server.cfg.MetricsReadThreads; threadIndex++ {
		server.wg.Add(1)
		server.running.Add(1)

		go func() {
			// Defers are executed in the LIFO order. Ensure running is at 0 before finishing the work group.
			defer func() { server.wg.Done() }()
			defer func() { server.running.Add(-1) }()

			readBuffer := make([]byte, server.cfg.MetricsReadBufferSize)

			for !server.shutdown.Load() {
				amountRead, err := server.conn.Read(readBuffer)
				if err != nil {
					var opErr *net.OpError
					isClosedConnError := errors.As(err, &opErr) && strings.Contains(opErr.Err.Error(), "use of closed network connection")
					if !isClosedConnError {
						logEntry.WithError(err).Errorf("Error while reading data from the socket.")
					}
					break
				}

				if amountRead == len(readBuffer) {
					logEntry.Errorf("Data received from the socket is too large.")
					continue
				}

				metrics, err := DecryptAndUnmarshal(readBuffer[:amountRead], server.enc)
				if err != nil {
					logEntry.WithError(err).Error("Failed to decrypt and unmarshall the data from the socket.")
					continue
				}

				for _, metric := range metrics {
					select {
					case server.metrics <- metric:
					default:
						logEntry.WithField("metric", metric).Warn("Queue is full.")
					}
				}
			}

			if err := server.Shutdown(); err != nil {
				logEntry.WithError(err).Error("Shutdown error.")
			}
		}()
	}

	return server.metrics
}

// Running returns whether the server has any go routines running.
func (server *Server) Running() bool {
	return server.running.Load() != 0
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
