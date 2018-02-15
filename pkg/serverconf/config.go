// Copyright (c) 2011 The Grumble Authors
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package serverconf

import (
	"strconv"
	"sync"
	"path/filepath"
)

var defaultCfg = map[string]string{
	"MaxBandwidth":          "72000",
	"MaxUsers":              "1000",
	"MaxUsersPerChannel":    "0",
	"MaxTextMessageLength":  "5000",
	"MaxImageMessageLength": "131072",
	"AllowHTML":             "true",
	"DefaultChannel":        "0",
	"RememberChannel":       "true",
	"WelcomeText":           "Welcome to this server running <b>Grumble</b>.",
	"SendVersion":           "true",
	"LogPath":               "grumble.log",
	"CertPath":              "cert.pem",
	"KeyPath":               "key.pem",
}

type Config struct {
	cfgMap map[string]string
	mutex  sync.RWMutex
}

// Create a new Config using cfgMap as the intial internal config map.
// If cfgMap is nil, ConfigWithMap will create a new config map.
func New(cfgMap map[string]string) *Config {
	if cfgMap == nil {
		cfgMap = make(map[string]string)
	}
	return &Config{cfgMap: cfgMap}
}

// Get a copy of the Config's internal config map
func (cfg *Config) GetAll() (all map[string]string) {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()

	all = make(map[string]string)
	for k, v := range cfg.cfgMap {
		all[k] = v
	}
	return
}

// Set a new value for a config key
func (cfg *Config) Set(key string, value string) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	cfg.cfgMap[key] = value
}

// Reset the value of a config key
func (cfg *Config) Reset(key string) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	delete(cfg.cfgMap, key)
}

// Get the value of a specific config key encoded as a string
func (cfg *Config) StringValue(key string) (value string) {
	cfg.mutex.RLock()
	defer cfg.mutex.RUnlock()

	value, exists := cfg.cfgMap[key]
	if exists {
		return value
	}

	value, exists = defaultCfg[key]
	if exists {
		return value
	}

	return ""
}

// Get the value of a specific config key as an int
func (cfg *Config) IntValue(key string) (intval int) {
	str := cfg.StringValue(key)
	intval, _ = strconv.Atoi(str)
	return
}

// Get the value of a specific config key as a uint32
func (cfg *Config) Uint32Value(key string) (uint32val uint32) {
	str := cfg.StringValue(key)
	uintval, _ := strconv.ParseUint(str, 10, 0)
	return uint32(uintval)
}

// Get the value of a specific config key as a bool
func (cfg *Config) BoolValue(key string) (boolval bool) {
	str := cfg.StringValue(key)
	boolval, _ = strconv.ParseBool(str)
	return
}

// Get the value of a specific config key as a path,
// joined with the path in rel if not absolute.
func (cfg *Config) PathValue(key string, rel string) (path string) {
	str := cfg.StringValue(key)
	if filepath.IsAbs(str) {
		return filepath.Clean(str)
	}
	return filepath.Join(rel, str)
}
