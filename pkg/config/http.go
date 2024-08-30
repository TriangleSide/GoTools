package config

import (
	"intelligence/pkg/config/envprocessor"
)

const (
	HTTPServerCertEnvName     envprocessor.EnvName = "HTTP_SERVER_CERT"
	HTTPServerKeyEnvName      envprocessor.EnvName = "HTTP_SERVER_KEY"
	HTTPServerBindPortEnvName envprocessor.EnvName = "HTTP_SERVER_BIND_PORT"
)

// HTTPServer contains the parameters needed to configure an HTTP server.
type HTTPServer struct {
	// HTTPServerBindIP specifies which network interface a server uses to handle incoming connections,
	// controlling access based on network location.
	HTTPServerBindIP string `config_format:"snake" config_default:"::1" validate:"required,ip_addr"`

	// HTTPServerBindPort designates a specific port number for an application to listen on,
	// uniquely identifying the service and managing incoming data traffic.
	HTTPServerBindPort uint16 `config_format:"snake" config_default:"36963" validate:"gt=0"`

	// HTTPServerReadTimeout is the maximum duration for reading the entire request,
	// including the body. A zero or negative value means there will be no timeout.
	HTTPServerReadTimeoutSeconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// HTTPServerWriteTimeout is the maximum duration before timing out writes of the response.
	// A zero or negative value means there will be no timeout.
	HTTPServerWriteTimeoutSeconds int `config_format:"snake" config_default:"120" validate:"gte=0"`

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, there is no timeout.
	HTTPServerIdleTimeoutSeconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// ReadHeaderTimeout is the amount of time allowed to read request headers. The connection's
	// read deadline is reset after reading the headers and the Handler can decide what is considered
	// too slow for the body. If ReadHeaderTimeout is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	HTTPServerHeaderReadTimeoutSeconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HTTPServerCert is the file path to the server's TLS certificate.
	// This certificate contains the public key part of the key pair used by the server
	// to establish its identity with clients during the TLS handshake. The certificate
	// must be issued by a trusted certificate authority (CA) or be a self-signed certificate
	// that clients trust. The certificate includes the server's public key along with
	// the identity of the server (like hostname), and it is encoded in PEM format.
	HTTPServerCert string `config_format:"snake" validate:"required,filepath"`

	// HTTPServerKey is the file path to the server's private key.
	// This key is the private part of the key pair associated with the server's certificate.
	// It is crucial for decrypting the information encrypted with the server's public key
	// by clients during the TLS handshake. The server's private key must be kept secure
	// and confidential because unauthorized access to this key compromises the entire
	// security of the TLS encryption. This key is typically also encoded in PEM format.
	HTTPServerKey string `config_format:"snake" validate:"required,filepath"`

	// HTTPServerMaxHeaderBytes controls the maximum number of bytes the server will read parsing
	// the request header's keys and values, including the request line. It does not limit the
	// size of the request body.
	HTTPServerMaxHeaderBytes int `config_format:"snake" config_default:"1048576" validate:"gte=4096,lte=1073741824"`
}
