package server

// TLSMode represents the TLS mode of the HTTP server.
type TLSMode string

const (
	// TLSModeOff represents plain HTTP without TLS.
	TLSModeOff TLSMode = "off"

	// TLSModeTLS represents HTTP over TLS.
	TLSModeTLS TLSMode = "tls"

	// TLSModeMutualTLS represents HTTP over mutual TLS.
	TLSModeMutualTLS TLSMode = "mutual_tls"
)

// Config holds configuration parameters for an HTTP server.
// HttpServer is the prefix for all fields because it avoids conflicts with other environment variables.
type Config struct {
	// HTTPServerBindIP is the IP address the server listens on.
	HTTPServerBindIP string `config_format:"snake" config_default:"::1" validate:"required,ip_addr"`

	// HTTPServerBindPort is the port number the server listens on.
	HTTPServerBindPort uint16 `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerReadTimeoutMilliseconds is the maximum time (in seconds) to read the request.
	// Zero or negative means no timeout.
	HTTPServerReadTimeoutMilliseconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// HTTPServerWriteTimeoutMilliseconds is the maximum time (in seconds) to write the response.
	// Zero or negative means no timeout.
	HTTPServerWriteTimeoutMilliseconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// HTTPServerIdleTimeoutMilliseconds sets the max idle time (in seconds) between requests when keep-alives are enabled.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerIdleTimeoutMilliseconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerHeaderReadTimeoutMilliseconds is the maximum time (in seconds) to read request headers.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HTTPServerHeaderReadTimeoutMilliseconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerTLSMode specifies the TLS mode of the server: off, tls, or mutual_tls.
	HTTPServerTLSMode TLSMode `config_format:"snake" config_default:"tls" validate:"oneof=off tls mutual_tls"`

	// HTTPServerCert is the path to the TLS certificate file.
	HTTPServerCert string `config_format:"snake" config_default:"" validate:"required_if=HTTPServerTLSMode tls,required_if=HTTPServerTLSMode mutual_tls,omitempty,filepath"`

	// HTTPServerKey is the path to the TLS private key file.
	HTTPServerKey string `config_format:"snake" config_default:"" validate:"required_if=HTTPServerTLSMode tls,required_if=HTTPServerTLSMode mutual_tls,omitempty,filepath"`

	// HTTPServerClientCACerts is a list of paths to client CA certificate files (used in mutual TLS).
	HTTPServerClientCACerts []string `config_format:"snake" config_default:"[]" validate:"required_if=HTTPServerTLSMode mutual_tls,dive,required,filepath"`

	// HTTPServerMaxHeaderBytes sets the maximum size in bytes of request headers. It doesn't limit the request body size.
	HTTPServerMaxHeaderBytes int `config_format:"snake" config_default:"1048576" validate:"gte=4096,lte=1073741824"`
}
