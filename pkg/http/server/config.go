package server

import (
	"github.com/TriangleSide/GoTools/pkg/validation"
)

const (
	// httpServerTLSFilePath validates optional file paths required when TLS is enabled.
	httpServerTLSFilePath = "http_server_tls_file_path"

	// httpServerMutualTLSFilePaths validates required file paths for mutual TLS client CA certs.
	httpServerMutualTLSFilePaths = "http_server_mutual_tls_file_paths"
)

// init registers validation aliases used by the Config struct.
func init() {
	validation.MustRegisterAlias(
		httpServerTLSFilePath,
		"required_if=HTTPServerTLSMode TLS,required_if=HTTPServerTLSMode MUTUAL_TLS,omitempty,filepath",
	)
	validation.MustRegisterAlias(
		httpServerMutualTLSFilePaths,
		"required_if=HTTPServerTLSMode MUTUAL_TLS,dive,required,filepath",
	)
}

// TLSMode represents the TLS mode of the HTTP server.
type TLSMode string

const (
	// TLSModeOff represents plain HTTP without TLS.
	TLSModeOff TLSMode = "OFF"

	// TLSModeTLS represents HTTP over TLS.
	TLSModeTLS TLSMode = "TLS"

	// TLSModeMutualTLS represents HTTP over mutual TLS.
	TLSModeMutualTLS TLSMode = "MUTUAL_TLS"
)

// Config holds configuration parameters for an HTTP server.
type Config struct {
	// HTTPServerBindIP is the IP address the server listens on.
	HTTPServerBindIP string `config:"ENV" config_default:"::1" validate:"required,ip_addr"`

	// HTTPServerBindPort is the port number the server listens on.
	HTTPServerBindPort uint16 `config:"ENV" config_default:"0" validate:"gte=0"`

	// HTTPServerReadTimeoutMillis is the maximum time (in milliseconds) to read the request.
	// Zero or negative means no timeout.
	HTTPServerReadTimeoutMillis int `config:"ENV" config_default:"120000" validate:"gte=0"`

	// HTTPServerWriteTimeoutMillis is the maximum time (in milliseconds) to write the response.
	// Zero or negative means no timeout.
	HTTPServerWriteTimeoutMillis int `config:"ENV" config_default:"120000" validate:"gte=0"`

	// HTTPServerIdleTimeoutMillis sets the max idle time (in milliseconds) between requests when
	// keep-alives are enabled. If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerIdleTimeoutMillis int `config:"ENV" config_default:"0" validate:"gte=0"`

	// HTTPServerHeaderReadTimeoutMillis is the maximum time (in milliseconds) to read request headers.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerHeaderReadTimeoutMillis int `config:"ENV" config_default:"0" validate:"gte=0"`

	// HTTPServerTLSMode specifies the TLS mode of the server: OFF, TLS, or MUTUAL_TLS.
	HTTPServerTLSMode TLSMode `config:"ENV" config_default:"TLS" validate:"oneof=OFF TLS MUTUAL_TLS"`

	// HTTPServerCert is the path to the TLS certificate file.
	HTTPServerCert string `config:"ENV" config_default:"" validate:"http_server_tls_file_path"`

	// HTTPServerKey is the path to the TLS private key file.
	HTTPServerKey string `config:"ENV" config_default:"" validate:"http_server_tls_file_path"`

	// HTTPServerClientCACerts is a list of paths to client CA certificate files (used in mutual TLS).
	HTTPServerClientCACerts []string `config:"ENV" config_default:"[]" validate:"http_server_mutual_tls_file_paths"`

	// HTTPServerMaxHeaderBytes sets the maximum size in bytes of request headers.
	// It doesn't limit the request body size.
	HTTPServerMaxHeaderBytes int `config:"ENV" config_default:"1048576" validate:"gte=4096,lte=1073741824"`

	// HTTPServerKeepAlive controls whether HTTP keep-alives are enabled.
	// By default, keep-alives are always enabled.
	HTTPServerKeepAlive bool `config:"ENV" config_default:"true"`
}
