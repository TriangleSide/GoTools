package server_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/server"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestNew_MutualTLSModeWithMissingClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerClientCACerts = []string{}
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "no client CAs provided")
	assert.Nil(t, srv)
}

func TestNew_MutualTLSModeWithNonexistentClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerClientCACerts = []string{"does_not_exist.pem"}
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "could not read client CA certificate")
	assert.Nil(t, srv)
}

func TestNew_MutualTLSModeWithInvalidClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidCertPath := filepath.Join(fixture.TempDir, "invalid_ca.pem")
	assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0600))
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		cfg.HTTPServerClientCACerts = []string{invalidCertPath}
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "failed to load client CA certificates")
	assert.Nil(t, srv)
}

func TestRun_MutualTLSModeWithoutClientCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
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
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddress, nil)
	assert.NoError(t, err)
	assert.NotNil(t, request)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "tls: certificate required")
	assert.Nil(t, response)
}

func TestRun_MutualTLSModeWithValidClientCert_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.ClientCertificateKeyPair},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_MutualTLSModeWithInvalidClientCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.InvalidClientCert},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddress, nil)
	assert.NoError(t, err)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "tls: certificate required")
	assert.Nil(t, response)
}

func TestRun_MutualTLSModeWithMultipleClientCAs_AcceptsClientsFromAnyCA(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	secondCACertPath := filepath.Join(fixture.TempDir, "second_ca_cert.pem")
	secondCAPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	secondCACertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			Organization: []string{"Second Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	secondCACertBytes, err := x509.CreateCertificate(
		rand.Reader, &secondCACertTemplate, &secondCACertTemplate,
		&secondCAPrivateKey.PublicKey, secondCAPrivateKey)
	assert.NoError(t, err)
	secondCACertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: secondCACertBytes})
	assert.NoError(t, os.WriteFile(secondCACertPath, secondCACertPEM, 0600))
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		cfg.HTTPServerClientCACerts = append(cfg.HTTPServerClientCACerts, secondCACertPath)
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.ClientCertificateKeyPair},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}
