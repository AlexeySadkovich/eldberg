package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadHolderPrivateKey(filename string) (string, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return "", fmt.Errorf("failed to open private key file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read privvate key file: %w", err)
	}

	return string(data), nil
}

func StoreHolderPrivateKey(filename, privateKey string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open private key file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(privateKey)
	if err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

func ReadPeers(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read peers file: %w", err)
	}

	var peers map[string]string
	if err := json.Unmarshal(data, &peers); err != nil {
		return nil, fmt.Errorf("invalid json data: %w", err)
	}

	return peers, nil
}

func IsDirectoryExists(dirname string) bool {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateDirectory(dirname string) error {
	if err := os.Mkdir(dirname, 0o777); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}
