package serverconf

import (
	"errors"
	"path/filepath"
	"strconv"
)

var globalKeys = map[string]bool{
	"LogPath": true,
}

type cfg interface {
	// GlobalMap returns a copy of the top-level (global) configuration map.
	GlobalMap() map[string]string

	// SubMap returns a copy of the server-specific (if existing) configuration map.
	SubMap(sub int64) map[string]string
}

type ConfigFile struct {
	cfg
}

func NewConfigFile(path string) (*ConfigFile, error) {
	var f cfg
	var err error
	switch filepath.Ext(path) {
	case ".ini":
		f, err = newinicfg(path)
	case ".json":
		f, err = newjsoncfg(path)
	default:
		return nil, errors.New("unknown config file format")
	}
	if err != nil {
		return nil, err
	}
	return &ConfigFile{f}, nil
}

// GlobalConfig returns a new *serverconf.Config representing the top-level
// (global) configuration.
func (c *ConfigFile) GlobalConfig() *Config {
	return New(c.GlobalMap())
}

// ServerConfig returns a new *serverconf.Config representing the union
// between the global configuration and the server-specific overrides.
// Optionally a base map m which to merge into may be passed. This map
// is consumed and cannot be reused.
func (c *ConfigFile) ServerConfig(id int64, m map[string]string) *Config {
	if m == nil {
		m = c.GlobalMap()

		// Strip the global keys so they don't get repeated in the freeze.
		for k := range globalKeys {
			delete(m, k)
		}
	} else {
		// Merge the global config into the base map
		for k, v := range c.GlobalMap() {
			if _, ok := globalKeys[k]; ok {
				// Ignore the global keys so they don't get repeated in the freeze.
				continue
			}
			if v != "" {
				m[k] = v
			} else {
				// Allow unset of base values through empty keys.
				delete(m, k)
			}
		}
	}

	// Some server specific values from the global config must be offset.
	if v, ok := m["Port"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			m["Port"] = strconv.FormatInt(i+id-1, 10)
		}
	}
	if v, ok := m["WebPort"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			m["WebPort"] = strconv.FormatInt(i+id-1, 10)
		}
	}

	// Merge the server-specific override (if one exists).
	for k, v := range c.SubMap(id) {
		if v != "" {
			m[k] = v
		} else {
			// Allow unset of global values through empty keys.
			delete(m, k)
		}
	}
	return New(m)
}
