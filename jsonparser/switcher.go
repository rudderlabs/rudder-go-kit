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
	getterFn     func() string
	setterFn     func() string
	deleterFn    func() string
	softGetterFn func() string
	impls        map[string]JSONParser
}

// GetValue delegates to the configured Getter implementation
func (s *switcher) GetValue(data []byte, path ...string) ([]byte, error) {
	return s.getter().GetValue(data, path...)
}

// GetBoolean delegates to the configured Getter implementation
func (s *switcher) GetBoolean(data []byte, path ...string) (bool, error) {
	return s.getter().GetBoolean(data, path...)
}

// GetInt delegates to the configured Getter implementation
func (s *switcher) GetInt(data []byte, path ...string) (int64, error) {
	return s.getter().GetInt(data, path...)
}

// GetFloat delegates to the configured Getter implementation
func (s *switcher) GetFloat(data []byte, path ...string) (float64, error) {
	return s.getter().GetFloat(data, path...)
}

// GetString delegates to the configured Getter implementation
func (s *switcher) GetString(data []byte, path ...string) (string, error) {
	return s.getter().GetString(data, path...)
}

// SetValue delegates to the configured Setter implementation
func (s *switcher) SetValue(data []byte, value interface{}, path ...string) ([]byte, error) {
	return s.setter().SetValue(data, value, path...)
}

// SetBoolean delegates to the configured Setter implementation
func (s *switcher) SetBoolean(data []byte, value bool, path ...string) ([]byte, error) {
	return s.setter().SetBoolean(data, value, path...)
}

// SetInt delegates to the configured Setter implementation
func (s *switcher) SetInt(data []byte, value int64, path ...string) ([]byte, error) {
	return s.setter().SetInt(data, value, path...)
}

// SetFloat delegates to the configured Setter implementation
func (s *switcher) SetFloat(data []byte, value float64, path ...string) ([]byte, error) {
	return s.setter().SetFloat(data, value, path...)
}

// SetString delegates to the configured Setter implementation
func (s *switcher) SetString(data []byte, value string, path ...string) ([]byte, error) {
	return s.setter().SetString(data, value, path...)
}

// DeleteKey delegates to the configured Setter implementation
func (s *switcher) DeleteKey(data []byte, path ...string) ([]byte, error) {
	return s.deleter().DeleteKey(data, path...)
}

func (s *switcher) GetValueOrEmpty(data []byte, path ...string) []byte {
	return s.safeGetter().GetValueOrEmpty(data, path...)
}

func (s *switcher) GetBooleanOrFalse(data []byte, path ...string) bool {
	return s.safeGetter().GetBooleanOrFalse(data, path...)
}

func (s *switcher) GetIntOrZero(data []byte, path ...string) int64 {
	return s.safeGetter().GetIntOrZero(data, path...)
}

func (s *switcher) GetFloatOrZero(data []byte, path ...string) float64 {
	return s.safeGetter().GetFloatOrZero(data, path...)
}

func (s *switcher) GetStringOrEmpty(data []byte, path ...string) string {
	return s.safeGetter().GetStringOrEmpty(data, path...)
}

// Getter returns the configured Getter implementation
func (s *switcher) getter() JSONParser {
	if impl, ok := s.impls[s.getterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// Setter returns the configured Setter implementation
func (s *switcher) setter() JSONParser {
	if impl, ok := s.impls[s.setterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// Deleter returns the configured Deleter implementation
func (s *switcher) deleter() JSONParser {
	if impl, ok := s.impls[s.deleterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// SoftGetter returns the configured SoftGetter implementation
func (s *switcher) safeGetter() JSONParser {
	if impl, ok := s.impls[s.softGetterFn()]; ok {
		return impl
	}
	return s.impls[DefaultLib]
}

// NewWithConfig returns a new JSONParser implementation based on the configuration
func NewWithConfig(conf *config.Config) JSONParser {
	getterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Getter", "JSONParser.Library").Load
	setterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Setter", "JSONParser.Library").Load
	deleterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Deleter", "JSONParser.Library").Load
	softGetterFn := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.SoftGetter", "JSONParser.Library").Load
	return &switcher{
		impls: map[string]JSONParser{
			TidwallLib: &tidwallJSONParser{},
			GrafanaLib: &grafanaJSONParser{},
		},
		getterFn:     getterFn,
		setterFn:     setterFn,
		deleterFn:    deleterFn,
		softGetterFn: softGetterFn,
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
