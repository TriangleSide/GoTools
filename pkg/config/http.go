package config

import (
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
)

const (
	HTTPServerBindPortEnvName      envprocessor.EnvName = "HTTP_SERVER_BIND_PORT"
	HTTPServerTLSModeEnvName       envprocessor.EnvName = "HTTP_SERVER_TLS_MODE"
	HTTPServerCertEnvName          envprocessor.EnvName = "HTTP_SERVER_CERT"
	HTTPServerKeyEnvName           envprocessor.EnvName = "HTTP_SERVER_KEY"
	HTTPServerClientCACertsEnvName envprocessor.EnvName = "HTTP_SERVER_CLIENT_CA_CERTS"
)

// HTTPServerTLSMode represents the TLS mode of the HTTP server.
type HTTPServerTLSMode string

const (
	// HTTPServerTLSModeOff represents plain HTTP without TLS.
	HTTPServerTLSModeOff HTTPServerTLSMode = "off"

	// HTTPServerTLSModeTLS represents HTTP over TLS.
	HTTPServerTLSModeTLS HTTPServerTLSMode = "tls"

	// HTTPServerTLSModeMutualTLS represents HTTP over mutual TLS.
	HTTPServerTLSModeMutualTLS HTTPServerTLSMode = "mutual_tls"
)

// HTTPServer holds configuration parameters for an HTTP server.
type HTTPServer struct {
	// HTTPServerBindIP is the IP address the server listens on.
	HTTPServerBindIP string `config_format:"snake" config_default:"::1" validate:"required,ip_addr"`

	// HTTPServerBindPort is the port number the server listens on.
	HTTPServerBindPort uint16 `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerReadTimeoutSeconds is the maximum time (in seconds) to read the request.
	// Zero or negative means no timeout.
	HTTPServerReadTimeoutSeconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// HTTPServerWriteTimeoutSeconds is the maximum time (in seconds) to write the response.
	// Zero or negative means no timeout.
	HTTPServerWriteTimeoutSeconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// HTTPServerIdleTimeoutSeconds sets the max idle time (in seconds) between requests when keep-alives are enabled.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerIdleTimeoutSeconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerHeaderReadTimeoutSeconds is the maximum time (in seconds) to read request headers.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerHeaderReadTimeoutSeconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerTLSMode specifies the TLS mode of the server: off, tls, or mutual_tls.
	HTTPServerTLSMode HTTPServerTLSMode `config_format:"snake" config_default:"tls" validate:"oneof=off tls mutual_tls"`

	// HTTPServerCert is the path to the TLS certificate file.
	HTTPServerCert string `config_format:"snake" config_default:"" validate:"required_if=HTTPServerTLSMode tls HTTPServerTLSMode mutual_tls,omitempty,filepath"`

	// HTTPServerKey is the path to the TLS private key file.
	HTTPServerKey string `config_format:"snake" config_default:"" validate:"required_if=HTTPServerTLSMode tls HTTPServerTLSMode mutual_tls,omitempty,filepath"`

	// HTTPServerClientCACerts is a list of paths to client CA certificate files (used in mutual TLS).
	HTTPServerClientCACerts []string `config_format:"snake" config_default:"[]" validate:"required_if=HTTPServerTLSMode mutual_tls,dive,required,filepath"`

	// HTTPServerMaxHeaderBytes sets the maximum size in bytes of request headers. It doesn't limit the request body size.
	HTTPServerMaxHeaderBytes int `config_format:"snake" config_default:"1048576" validate:"gte=4096,lte=1073741824"`
}
