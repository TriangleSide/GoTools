// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package config

import (
	"time"
)

// Server contains the parameters needed to configure an HTTP server.
type Server struct {
	// ServerBindIP specifies which network interface a server uses to handle incoming connections,
	// controlling access based on network location.
	ServerBindIP string `config_format:"snake" config_default:"::1" validate:"required,ip_addr"`

	// ServerBindPort designates a specific port number for an application to listen on,
	// uniquely identifying the service and managing incoming data traffic.
	ServerBindPort uint16 `config_format:"snake" config_default:"36963" validate:"required,gt=0"`

	// ServerReadTimeout is the maximum duration for reading the entire request,
	// including the body. A zero or negative value means there will be no timeout.
	ServerReadTimeout time.Duration `config_format:"snake" config_default:"120s" validate:"required,gte=0"`

	// ServerWriteTimeout is the maximum duration before timing out writes of the response.
	// A zero or negative value means there will be no timeout.
	ServerWriteTimeout time.Duration `config_format:"snake" config_default:"120s" validate:"required,gte=0"`

	// ServerCert is the file path to the server's TLS certificate.
	// This certificate contains the public key part of the key pair used by the server
	// to establish its identity with clients during the TLS handshake. The certificate
	// must be issued by a trusted certificate authority (CA) or be a self-signed certificate
	// that clients trust. The certificate includes the server's public key along with
	// the identity of the server (like hostname), and it is encoded in PEM format.
	ServerCert string `config_format:"snake" validate:"required,filepath"`

	// ServerKey is the file path to the server's private key.
	// This key is the private part of the key pair associated with the server's certificate.
	// It is crucial for decrypting the information encrypted with the server's public key
	// by clients during the TLS handshake. The server's private key must be kept secure
	// and confidential because unauthorized access to this key compromises the entire
	// security of the TLS encryption. This key is typically also encoded in PEM format.
	ServerKey string `config_format:"snake" config_default:"120s" validate:"required,filepath"`

	// ServerMaxHeaderBytes controls the maximum number of bytes the server will read parsing
	// the request header's keys and values, including the request line. It does not limit the
	// size of the request body.
	ServerMaxHeaderBytes int `config_format:"snake" config_default:"1048576" validate:"required,gte=4096,lte=1073741824"`
}
