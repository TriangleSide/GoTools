package metrics

import (
	"encoding/json"
	"fmt"

	"intelligence/pkg/crypto/symmetric"
)

// MarshalAndEncrypt marshals the metrics slice to JSON, and encrypts it using the provided Encryptor.
func MarshalAndEncrypt(metrics []*Metric, enc *symmetric.Encryptor) ([]byte, error) {
	jsonBytes, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metrics (%s)", err.Error())
	}
	cypher, err := enc.Encrypt(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt (%s)", err.Error())
	}
	return cypher, nil
}

// DecryptAndUnmarshal decrypts the data using the provided Encryptor, and unmarshals it into a Metric slice.
func DecryptAndUnmarshal(data []byte, enc *symmetric.Encryptor) ([]*Metric, error) {
	jsonBytes, err := enc.Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt (%s)", err.Error())
	}
	var metrics []*Metric
	err = json.Unmarshal(jsonBytes, &metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal (%s)", err.Error())
	}
	return metrics, nil
}
