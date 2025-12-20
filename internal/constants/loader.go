package constants

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	constantsCache map[string]string
	cacheMutex     sync.RWMutex
)

// LoadConstants loads constants from constants.json file
func LoadConstants() (map[string]string, error) {
	cacheMutex.RLock()
	if constantsCache != nil {
		defer cacheMutex.RUnlock()
		return constantsCache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if constantsCache != nil {
		return constantsCache, nil
	}

	data, err := os.ReadFile("constants.json")
	if err != nil {
		return nil, err
	}

	var constants map[string]string
	if err := json.Unmarshal(data, &constants); err != nil {
		return nil, err
	}

	constantsCache = constants
	return constants, nil
}

// GetConstant retrieves a constant value by key
func GetConstant(key string) (string, bool) {
	constants, err := LoadConstants()
	if err != nil {
		return "", false
	}

	value, exists := constants[key]
	return value, exists
}

// ReloadConstants forces a reload of constants from disk
func ReloadConstants() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	data, err := os.ReadFile("constants.json")
	if err != nil {
		return err
	}

	var constants map[string]string
	if err := json.Unmarshal(data, &constants); err != nil {
		return err
	}

	constantsCache = constants
	return nil
}
