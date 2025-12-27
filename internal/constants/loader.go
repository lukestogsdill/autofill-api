package constants

import (
	"encoding/json"
	"fmt"
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

	// First unmarshal to interface{} to handle mixed types
	var rawConstants map[string]interface{}
	if err := json.Unmarshal(data, &rawConstants); err != nil {
		return nil, err
	}

	// Convert all values to strings (booleans become "yes"/"no")
	constants := make(map[string]string)
	for key, value := range rawConstants {
		constants[key] = interfaceToString(value)
	}

	constantsCache = constants
	return constants, nil
}

// interfaceToString converts interface{} values to strings
// Booleans are converted to "yes"/"no" for form compatibility
func interfaceToString(value interface{}) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "yes"
		}
		return "no"
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
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

	// First unmarshal to interface{} to handle mixed types
	var rawConstants map[string]interface{}
	if err := json.Unmarshal(data, &rawConstants); err != nil {
		return err
	}

	// Convert all values to strings (booleans become "yes"/"no")
	constants := make(map[string]string)
	for key, value := range rawConstants {
		constants[key] = interfaceToString(value)
	}

	constantsCache = constants
	return nil
}
