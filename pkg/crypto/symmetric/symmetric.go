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

package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

// Config is the configuration for the encryptor.
type Config struct {
	blockCypherProvider func(key []byte) (cipher.Block, error)
	randomDataFunc      func(buffer []byte) error
}

// Option is optional configuration of the encryptor.
type Option func(*Config) error

// WithBlockCypherProvider overwrites the provider for the block cipher.
func WithBlockCypherProvider(provider func(key []byte) (cipher.Block, error)) Option {
	return func(c *Config) error {
		c.blockCypherProvider = provider
		return nil
	}
}

// WithRandomDataFunc overwrites the random data function.
func WithRandomDataFunc(randomDataFunc func(buffer []byte) error) Option {
	return func(c *Config) error {
		c.randomDataFunc = randomDataFunc
		return nil
	}
}

// Encryptor does symmetric encryption and decryption.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// aesEncryptor holds the data needed to do AES symmetric encryption.
type aesEncryptor struct {
	aesBlock       cipher.Block
	randomDataFunc func(buffer []byte) error
}

// New allocates and configures an Encryptor.
func New(key string, opts ...Option) (Encryptor, error) {
	config := &Config{
		blockCypherProvider: aes.NewCipher,
		randomDataFunc: func(buffer []byte) error {
			_, err := io.ReadFull(rand.Reader, buffer)
			return err
		},
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failure while configuring the encryptor (%s)", err.Error())
		}
	}

	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	hash := sha256.Sum256([]byte(key))

	block, err := config.blockCypherProvider(hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create the block cipher (%s)", err.Error())
	}

	return &aesEncryptor{
		aesBlock:       block,
		randomDataFunc: config.randomDataFunc,
	}, nil
}

// Encrypt takes a slice of data and returns an encrypted version of that data using the AES algorithm.
// It returns a cypher-text slice of data and an error if any occurs during the encryption process.
func (encryptor *aesEncryptor) Encrypt(data []byte) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(data))

	iv := ciphertext[:aes.BlockSize]
	if err := encryptor.randomDataFunc(iv); err != nil {
		return nil, fmt.Errorf("failed to generate initialization vector (%s)", err.Error())
	}

	cfb := cipher.NewCFBEncrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// Decrypt performs symmetric decryption on a slice of data using the AES algorithm.
// It returns a plain-text slice of data and an error if any occurs during the decryption process.
func (encryptor *aesEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("cipher-text of len %d is shorter then the minimum length of %d", len(encryptedData), aes.BlockSize)
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(encryptedData, encryptedData)

	return encryptedData, nil
}
