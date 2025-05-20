package jsonparser

import (
	"github.com/rudderlabs/rudder-go-kit/config"
)

const (
	// GjsonLib is the implementation using gjson/sjson libraries
	GjsonLib = "gjson"
	// JsonparserLib is the implementation using jsonparser library
	JsonparserLib = "jsonparser"
	// DefaultLib is the default implementation
	DefaultLib = GjsonLib
)

// switcher is a JSONParser implementation that switches between different implementations based on configuration
type switcher struct {
	getterFn func() string
	setterFn func() string
	impls    map[string]JSONParser
}

// GetValue delegates to the configured getter implementation
func (s *switcher) GetValue(data []byte, key string) (interface{}, error) {
	return s.getter().GetValue(data, key)
}

// GetBoolean delegates to the configured getter implementation
func (s *switcher) GetBoolean(data []byte, key string) (bool, error) {
	return s.getter().GetBoolean(data, key)
}

// GetInt delegates to the configured getter implementation
func (s *switcher) GetInt(data []byte, key string) (int64, error) {
	return s.getter().GetInt(data, key)
}

// GetFloat delegates to the configured getter implementation
func (s *switcher) GetFloat(data []byte, key string) (float64, error) {
	return s.getter().GetFloat(data, key)
}

// GetString delegates to the configured getter implementation
func (s *switcher) GetString(data []byte, key string) (string, error) {
	return s.getter().GetString(data, key)
}

// SetValue delegates to the configured setter implementation
func (s *switcher) SetValue(data []byte, key string, value interface{}) ([]byte, error) {
	return s.setter().SetValue(data, key, value)
}

// SetBoolean delegates to the configured setter implementation
func (s *switcher) SetBoolean(data []byte, key string, value bool) ([]byte, error) {
	return s.setter().SetBoolean(data, key, value)
}

// SetInt delegates to the configured setter implementation
func (s *switcher) SetInt(data []byte, key string, value int64) ([]byte, error) {
	return s.setter().SetInt(data, key, value)
}

// SetFloat delegates to the configured setter implementation
func (s *switcher) SetFloat(data []byte, key string, value float64) ([]byte, error) {
	return s.setter().SetFloat(data, key, value)
}

// SetString delegates to the configured setter implementation
func (s *switcher) SetString(data []byte, key, value string) ([]byte, error) {
	return s.setter().SetString(data, key, value)
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

// NewWithConfig returns a new JSONParser implementation based on the configuration
func NewWithConfig(conf *config.Config) JSONParser {
	getter := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Getter", "JSONParser.Library").Load
	setter := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Setter", "JSONParser.Library").Load

	return &switcher{
		impls: map[string]JSONParser{
			GjsonLib:      &gjsonJSONParser{},
			JsonparserLib: &jsonparserJSONParser{},
		},
		getterFn: getter,
		setterFn: setter,
	}
}

// NewGetterWithConfig returns a new JSONGetter implementation based on the configuration
func NewGetterWithConfig(conf *config.Config) JSONGetter {
	getter := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Getter", "JSONParser.Library").Load

	return &switcher{
		impls: map[string]JSONParser{
			GjsonLib:      &gjsonJSONParser{},
			JsonparserLib: &jsonparserJSONParser{},
		},
		getterFn: getter,
		setterFn: func() string { return "" }, // Not used for getter-only
	}
}

// NewSetterWithConfig returns a new JSONSetter implementation based on the configuration
func NewSetterWithConfig(conf *config.Config) JSONSetter {
	setter := conf.GetReloadableStringVar(DefaultLib, "JSONParser.Library.Setter", "JSONParser.Library").Load

	return &switcher{
		impls: map[string]JSONParser{
			GjsonLib:      &gjsonJSONParser{},
			JsonparserLib: &jsonparserJSONParser{},
		},
		getterFn: func() string { return "" }, // Not used for setter-only
		setterFn: setter,
	}
}

// NewWithLibrary returns a new JSONParser implementation based on the specified library
func NewWithLibrary(library string) JSONParser {
	switch library {
	case GjsonLib:
		return &gjsonJSONParser{}
	case JsonparserLib:
		return &jsonparserJSONParser{}
	default:
		return &gjsonJSONParser{}
	}
}

// NewGetterWithLibrary returns a new JSONGetter implementation based on the specified library
func NewGetterWithLibrary(library string) JSONGetter {
	switch library {
	case GjsonLib:
		return &gjsonJSONParser{}
	case JsonparserLib:
		return &jsonparserJSONParser{}
	default:
		return &gjsonJSONParser{}
	}
}

// NewSetterWithLibrary returns a new JSONSetter implementation based on the specified library
func NewSetterWithLibrary(library string) JSONSetter {
	switch library {
	case GjsonLib:
		return &gjsonJSONParser{}
	case JsonparserLib:
		return &jsonparserJSONParser{}
	default:
		return &gjsonJSONParser{}
	}
}
