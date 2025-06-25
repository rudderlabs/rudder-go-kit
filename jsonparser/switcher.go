package jsonparser

import (
	"github.com/rudderlabs/rudder-go-kit/config"
)

const (
	// TidwallLib is the implementation using gjson/sjson libraries
	TidwallLib = "tidwall"
	// GrafanaLib is the implementation using jsonparser library
	GrafanaLib = "grafana"
	// DefaultLib is the default implementation
	DefaultLib = GrafanaLib
)

// switcher is a JSONParser implementation that switches between different implementations based on configuration
type switcher struct {
	getterFn  func() string
	setterFn  func() string
	deleterFn func() string
	impls     map[string]JSONParser
}

// GetValue delegates to the configured getter implementation
func (s *switcher) GetValue(data []byte, keys ...string) ([]byte, error) {
	return s.getter().GetValue(data, keys...)
}

// GetBoolean delegates to the configured getter implementation
func (s *switcher) GetBoolean(data []byte, keys ...string) (bool, error) {
	return s.getter().GetBoolean(data, keys...)
}

// GetInt delegates to the configured getter implementation
func (s *switcher) GetInt(data []byte, keys ...string) (int64, error) {
	return s.getter().GetInt(data, keys...)
}

// GetFloat delegates to the configured getter implementation
func (s *switcher) GetFloat(data []byte, keys ...string) (float64, error) {
	return s.getter().GetFloat(data, keys...)
}

// GetString delegates to the configured getter implementation
func (s *switcher) GetString(data []byte, keys ...string) (string, error) {
	return s.getter().GetString(data, keys...)
}

// SetValue delegates to the configured setter implementation
func (s *switcher) SetValue(data []byte, value interface{}, keys ...string) ([]byte, error) {
	return s.setter().SetValue(data, value, keys...)
}

// SetBoolean delegates to the configured setter implementation
func (s *switcher) SetBoolean(data []byte, value bool, keys ...string) ([]byte, error) {
	return s.setter().SetBoolean(data, value, keys...)
}

// SetInt delegates to the configured setter implementation
func (s *switcher) SetInt(data []byte, value int64, keys ...string) ([]byte, error) {
	return s.setter().SetInt(data, value, keys...)
}

// SetFloat delegates to the configured setter implementation
func (s *switcher) SetFloat(data []byte, value float64, keys ...string) ([]byte, error) {
	return s.setter().SetFloat(data, value, keys...)
}

// SetString delegates to the configured setter implementation
func (s *switcher) SetString(data []byte, value string, keys ...string) ([]byte, error) {
	return s.setter().SetString(data, value, keys...)
}

// DeleteKey delegates to the configured setter implementation
func (s *switcher) DeleteKey(data []byte, keys ...string) ([]byte, error) {
	return s.deleter().DeleteKey(data, keys...)
}

// getter returns the configured getter implementation
func (s *switcher) getter() JSONParser {
	if impl, ok := s.impls[s.getterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// setter returns the configured setter implementation
func (s *switcher) setter() JSONParser {
	if impl, ok := s.impls[s.setterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// deleter returns the configured deleter implementation
func (s *switcher) deleter() JSONParser {
	if impl, ok := s.impls[s.deleterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// NewWithConfig returns a new JSONParser implementation based on the configuration
func NewWithConfig(conf *config.Config) JSONParser {
	getterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Getter", "JSONParser.Library").Load
	setterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Setter", "JSONParser.Library").Load
	deleterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Deleter", "JSONParser.Library").Load
	return &switcher{
		impls: map[string]JSONParser{
			TidwallLib: &tidwallJSONParser{},
			GrafanaLib: &grafanaJSONParser{},
		},
		getterFn:  getterFn,
		setterFn:  setterFn,
		deleterFn: deleterFn,
	}
}

// NewWithLibrary returns a new JSONParser implementation based on the specified library
func NewWithLibrary(library string) JSONParser {
	switch library {
	case TidwallLib:
		return &tidwallJSONParser{}
	case GrafanaLib:
		return &grafanaJSONParser{}
	default:
		return &tidwallJSONParser{}
	}
}
