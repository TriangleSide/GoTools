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

package metrics

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"intelligence/pkg/config"
	"intelligence/pkg/crypto/symmetric"
	netutils "intelligence/pkg/utils/net"
)

// Client represents a client for sending metrics to the server.
type Client struct {
	enc      *symmetric.Encryptor
	cfg      *config.MetricsClient
	conn     *net.UDPConn
	shutdown *atomic.Bool
	wg       sync.WaitGroup
}

// NewClient creates a new Client instance from a configuration parsed from the environment variables.
func NewClient() (*Client, error) {
	cfg, err := config.ProcessAndValidate[config.MetricsClient]()
	if err != nil {
		return nil, fmt.Errorf("failed to get the metrics client configuration (%s)", err.Error())
	}

	serverAddr, err := netutils.FormatNetworkAddress(cfg.MetricsHost, cfg.MetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to format the metrics server address (%s)", err.Error())
	}

	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the metrics server address (%s)", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial the metrics server (%s)", err.Error())
	}

	encryptor, err := symmetric.New(cfg.MetricsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create the encryptor (%s)", err.Error())
	}

	shutdownFlag := &atomic.Bool{}
	shutdownFlag.Store(false)

	return &Client{
		cfg:      cfg,
		conn:     conn,
		enc:      encryptor,
		shutdown: shutdownFlag,
		wg:       sync.WaitGroup{},
	}, nil
}

// Send sends a list of metrics to the server.
// The metrics may fail to make it to the server because the connection is UDP.
// According to https://pkg.go.dev/net#Conn, the UDP conn is thread safe.
func (client *Client) Send(metrics []*Metric) error {
	client.wg.Add(1)
	defer func() { client.wg.Done() }()

	if client.shutdown.Load() {
		return errors.New("metrics client is closed")
	}

	encryptedBytes, err := MarshalAndEncrypt(metrics, client.enc)
	if err != nil {
		return fmt.Errorf("failed to marshal and encrypt the metrics (%s)", err.Error())
	}

	n, err := client.conn.Write(encryptedBytes)
	if err != nil {
		return fmt.Errorf("failed to send the metrics to the server (%s)", err.Error())
	}
	if n != len(encryptedBytes) {
		return fmt.Errorf("only sent %d/%d bytes to the server", n, len(encryptedBytes))
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
