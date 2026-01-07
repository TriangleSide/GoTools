package server_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/server"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type tlsTestFixture struct {
	TempDir                  string
	ClientCACertPath         string
	ClientCACertPaths        []string
	ServerPrivateKeyPath     string
	ServerCertificatePath    string
	ClientCertificateKeyPair tls.Certificate
	InvalidClientCert        tls.Certificate
	CACertPool               *x509.CertPool
}

type caCredentials struct {
	privateKey *rsa.PrivateKey
	certPEM    []byte
	template   *x509.Certificate
}

func generateCACertificate(t *testing.T) *caCredentials {
	t.Helper()
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	caCertTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caCertBytes, err := x509.CreateCertificate(
		rand.Reader, caCertTemplate, caCertTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	assert.NoError(t, err)
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

	return &caCredentials{
		privateKey: caPrivateKey,
		certPEM:    caCertPEM,
		template:   caCertTemplate,
	}
}

func generateServerCertificate(
	t *testing.T, tempDir string, caCredentials *caCredentials,
) (string, string, []byte) {
	t.Helper()
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	serverPrivateKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)})

	serverCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Server Tests Inc."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	serverCertBytes, err := x509.CreateCertificate(
		rand.Reader, &serverCertTemplate, caCredentials.template, &serverPrivateKey.PublicKey, caCredentials.privateKey)
	assert.NoError(t, err)
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})

	serverPrivateKeyPath := filepath.Join(tempDir, "server_key.pem")
	assert.NoError(t, os.WriteFile(serverPrivateKeyPath, serverPrivateKeyPEM, 0600))

	serverCertificatePath := filepath.Join(tempDir, "server_cert.pem")
	assert.NoError(t, os.WriteFile(serverCertificatePath, serverCertPEM, 0600))

	return serverPrivateKeyPath, serverCertificatePath, serverCertPEM
}

func generateClientCertificate(t *testing.T, tempDir string, caCredential *caCredentials) (tls.Certificate, []byte) {
	t.Helper()
	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	clientPrivateKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)})

	clientCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"Client Tests Inc."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	clientCertBytes, err := x509.CreateCertificate(
		rand.Reader, &clientCertTemplate, caCredential.template, &clientPrivateKey.PublicKey, caCredential.privateKey)
	assert.NoError(t, err)
	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes})

	clientPrivateKeyPath := filepath.Join(tempDir, "client_key.pem")
	assert.NoError(t, os.WriteFile(clientPrivateKeyPath, clientPrivateKeyPEM, 0600))

	clientCertificatePath := filepath.Join(tempDir, "client_cert.pem")
	assert.NoError(t, os.WriteFile(clientCertificatePath, clientCertPEM, 0600))

	clientCertificateKeyPair, err := tls.LoadX509KeyPair(clientCertificatePath, clientPrivateKeyPath)
	assert.NoError(t, err)

	return clientCertificateKeyPair, clientCertPEM
}

func generateSelfSignedClientCertificate(t *testing.T) tls.Certificate {
	t.Helper()
	invalidClientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	invalidClientCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject: pkix.Name{
			Organization: []string{"Invalid Client"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	invalidClientCertBytes, err := x509.CreateCertificate(
		rand.Reader, &invalidClientCertTemplate, &invalidClientCertTemplate,
		&invalidClientPrivateKey.PublicKey, invalidClientPrivateKey)
	assert.NoError(t, err)
	invalidClientCertPEM := pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: invalidClientCertBytes})
	invalidClientKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(invalidClientPrivateKey)})

	invalidClientCert, err := tls.X509KeyPair(invalidClientCertPEM, invalidClientKeyPEM)
	assert.NoError(t, err)

	return invalidClientCert
}

func setupTLSTestFixture(t *testing.T) *tlsTestFixture {
	t.Helper()
	tempDir := t.TempDir()

	caCredential := generateCACertificate(t)

	clientCACertPath := filepath.Join(tempDir, "ca_cert.pem")
	assert.NoError(t, os.WriteFile(clientCACertPath, caCredential.certPEM, 0600))

	serverKeyPath, serverCertPath, serverCertPEM := generateServerCertificate(t, tempDir, caCredential)
	clientCertKeyPair, clientCertPEM := generateClientCertificate(t, tempDir, caCredential)
	invalidClientCert := generateSelfSignedClientCertificate(t)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(clientCertPEM)
	caCertPool.AppendCertsFromPEM(serverCertPEM)

	return &tlsTestFixture{
		TempDir:                  tempDir,
		ClientCACertPath:         clientCACertPath,
		ClientCACertPaths:        []string{clientCACertPath},
		ServerPrivateKeyPath:     serverKeyPath,
		ServerCertificatePath:    serverCertPath,
		ClientCertificateKeyPair: clientCertKeyPair,
		InvalidClientCert:        invalidClientCert,
		CACertPool:               caCertPool,
	}
}

func certPathsConfigProvider(t *testing.T, fixture *tlsTestFixture) *server.Config {
	t.Helper()
	cfg := getDefaultConfig(t)
	cfg.HTTPServerKey = fixture.ServerPrivateKeyPath
	cfg.HTTPServerCert = fixture.ServerCertificatePath
	cfg.HTTPServerClientCACerts = fixture.ClientCACertPaths
	cfg.HTTPServerKeepAlive = false
	return cfg
}

func TestNew_TLSModeWithMissingKeys_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := getDefaultConfig(t)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		cfg.HTTPServerKey = ""
		cfg.HTTPServerCert = ""
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "failed to load the server certificates")
	assert.Nil(t, srv)
}

func TestNew_TLSModeWithMissingCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerCert = ""
			cfg.HTTPServerTLSMode = mode
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_TLSModeWithMissingKey_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerKey = ""
			cfg.HTTPServerTLSMode = mode
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_TLSModeWithInvalidCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidCertPath := filepath.Join(fixture.TempDir, "invalid_cert.pem")
	assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0600))
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerTLSMode = mode
			cfg.HTTPServerCert = invalidCertPath
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_TLSModeWithInvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidKeyPath := filepath.Join(fixture.TempDir, "invalid_key.pem")
	assert.NoError(t, os.WriteFile(invalidKeyPath, []byte("invalid data"), 0600))
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerTLSMode = mode
			cfg.HTTPServerKey = invalidKeyPath
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestRun_TLSModeWithUntrustedClient_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddr, nil)
	assert.NoError(t, err)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "certificate")
	assert.Nil(t, response)
}

func TestRun_TLSModeWithProperClient_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_TLSModeWithClientCA_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	caCertPEM, err := os.ReadFile(fixture.ClientCACertPath)
	assert.NoError(t, err)
	caCertPool := x509.NewCertPool()
	assert.True(t, caCertPool.AppendCertsFromPEM(caCertPEM))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}
