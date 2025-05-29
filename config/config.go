// Package config uses the exact same precedence order as Viper, where
// each item takes precedence over the item below it:
//
//   - explicit call to Set (case insensitive)
//   - flag (case insensitive)
//   - env (case sensitive - see notes below)
//   - config (case insensitive)
//   - key/value store (case insensitive)
//   - default (case insensitive)
//
// Environment variable resolution is performed based on the following rules:
//   - If the key contains only uppercase characters, numbers and underscores, the environment variable is looked up in its entirety, e.g. SOME_VARIABLE -> SOME_VARIABLE
//   - In all other cases, the environment variable is transformed before being looked up as following:
//     1. camelCase is converted to snake_case, e.g. someVariable -> some_variable
//     2. dots (.) are replaced with underscores (_), e.g. some.variable -> some_variable
//     3. the resulting string is uppercased and prefixed with ${PREFIX}_ (default RSERVER_), e.g. some_variable -> RSERVER_SOME_VARIABLE
//
// Order of keys:
//
//		When registering a variable with multiple keys, the order of the keys is important as it determines the
//		hierarchical order of the keys.
//		The first key is the most important one, and the last key is the least important one.
//		Example:
//		config.GetReloadableDurationVar(90, time.Second,
//			"JobsDB.Router.CommandRequestTimeout",
//			"JobsDB.CommandRequestTimeout",
//		)
//
//		In the above example "JobsDB.Router.CommandRequestTimeout" is checked first. If it doesn't exist then
//	    JobsDB.CommandRequestTimeout is checked.
//
//	    WARNING: for this reason, registering with the same keys but in a different order is going to return two
//	    different variables.
package config

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const DefaultEnvPrefix = "RSERVER"

// Default is the singleton config instance
var Default *Config

func init() {
	Default = New()
}

// Reset resets the default, singleton config instance.
// Shall only be used by tests, until we move to a proper DI framework
func Reset() {
	Default = New()
}

type Opt func(*Config)

// WithEnvPrefix sets the environment variable prefix (default: RSERVER)
func WithEnvPrefix(prefix string) Opt {
	return func(c *Config) {
		c.envPrefix = prefix
	}
}

// New creates a new config instance
func New(opts ...Opt) *Config {
	c := &Config{
		envPrefix:             DefaultEnvPrefix,
		reloadableVars:        make(map[string]any),
		reloadableVarsMisuses: make(map[string]string),
		nonReloadableKeys:     make(map[string]string),
		notifier:              &notifier{},
	}
	c.RegisterObserver(&printObserver{})
	for _, opt := range opts {
		opt(c)
	}
	c.load()
	return c
}

// Config is the entry point for accessing configuration
type Config struct {
	vLock sync.RWMutex // protects reading and writing to the config (viper is not thread-safe)
	v     *viper.Viper

	hotReloadableConfigLock sync.RWMutex              // protects map holding hot reloadable config keys
	hotReloadableConfig     map[string][]*configValue // key -> comma-separated list of config keys, e.g. jobsdb.host, value -> list of configValue pointers (different types)
	reloadableVars          map[string]any            // key -> type with comma-separated list of config keys, e.g. string:jobsdb.host, value -> Reloadable[T] pointer
	reloadableVarsMisuses   map[string]string         // key -> type with comma-separated list of config keys, e.g. string:jobsdb.host, value -> original default value as a string, e.g. "localhost:5432"

	// nonHotReloadableConfigLock sync.RWMutex // protects map holding hot reloadable config keys
	// nonHotReloadableConfig     map[string][]*configValue
	nonReloadableKeysLock sync.RWMutex      // protects nonReloadableKeys
	nonReloadableKeys     map[string]string // key -> non-reloadable config key in lowercase, e.g. jobsdb.host, value -> original key, e.g. JobsDB.Host
	currentSettings       map[string]any    // current config settings. Keys are always stored flattened and in lower case, e.g. jobsdb.host

	envsLock  sync.RWMutex // protects the envs map below
	envs      map[string]string
	envPrefix string // prefix for environment variables

	reloadableVarsLock sync.RWMutex // used to protect both the reloadableVars and reloadableVarsMisuses maps
	configPath         string
	configPathErr      error
	godotEnvErr        error
	notifier           *notifier // for notifying subscribers of config changes
}

// GetBool gets bool value from config
func GetBool(key string, defaultValue bool) (value bool) {
	return Default.GetBool(key, defaultValue)
}

// GetBool gets bool value from config
func (c *Config) GetBool(key string, defaultValue bool) (value bool) {
	return c.GetBoolVar(defaultValue, key)
}

// getBoolInternal gets bool value from config
func (c *Config) getBoolInternal(key string, defaultValue bool) (value bool) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetBool(key)
}

// GetInt gets int value from config
func GetInt(key string, defaultValue int) (value int) {
	return Default.GetInt(key, defaultValue)
}

// GetInt gets int value from config
func (c *Config) GetInt(key string, defaultValue int) (value int) {
	return c.GetIntVar(defaultValue, 1, key)
}

// getIntInternal gets int value from config
func (c *Config) getIntInternal(key string, defaultValue int) (value int) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetInt(key)
}

// GetStringMap gets string map value from config
func GetStringMap(key string, defaultValue map[string]any) (value map[string]any) {
	return Default.GetStringMap(key, defaultValue)
}

// GetStringMap gets string map value from config
func (c *Config) GetStringMap(key string, defaultValue map[string]any) (value map[string]any) {
	return c.GetStringMapVar(defaultValue, key)
}

// getStringMapInternal gets string map value from config
func (c *Config) getStringMapInternal(key string, defaultValue map[string]any) (value map[string]any) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetStringMap(key)
}

// MustGetInt gets int value from config or panics if the config doesn't exist
func MustGetInt(key string) (value int) {
	return Default.MustGetInt(key)
}

// MustGetInt gets int value from config or panics if the config doesn't exist
func (c *Config) MustGetInt(key string) (value int) {
	c.registerNonReloadableConfigKeys(key)
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		panic(fmt.Errorf("config key %s not found", key))
	}
	return c.v.GetInt(key)
}

// GetInt64 gets int64 value from config
func GetInt64(key string, defaultValue int64) (value int64) {
	return Default.GetInt64(key, defaultValue)
}

// GetInt64 gets int64 value from config
func (c *Config) GetInt64(key string, defaultValue int64) (value int64) {
	return c.GetInt64Var(defaultValue, 1, key)
}

// getInt64Internal gets int64 value from config
func (c *Config) getInt64Internal(key string, defaultValue int64) (value int64) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetInt64(key)
}

// GetFloat64 gets float64 value from config
func GetFloat64(key string, defaultValue float64) (value float64) {
	return Default.GetFloat64(key, defaultValue)
}

// GetFloat64 gets float64 value from config
func (c *Config) GetFloat64(key string, defaultValue float64) (value float64) {
	return c.GetFloat64Var(defaultValue, key)
}

// getFloat64Internal gets float64 value from config
func (c *Config) getFloat64Internal(key string, defaultValue float64) (value float64) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetFloat64(key)
}

// GetString gets string value from config
func GetString(key, defaultValue string) (value string) {
	return Default.GetString(key, defaultValue)
}

// GetString gets string value from config
func (c *Config) GetString(key, defaultValue string) (value string) {
	return c.GetStringVar(defaultValue, key)
}

// getStringInternal gets string value from config
func (c *Config) getStringInternal(key, defaultValue string) (value string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetString(key)
}

// MustGetString gets string value from config or panics if the config doesn't exist
func MustGetString(key string) (value string) {
	return Default.MustGetString(key)
}

// MustGetString gets string value from config or panics if the config doesn't exist
func (c *Config) MustGetString(key string) (value string) {
	c.registerNonReloadableConfigKeys(key)
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		panic(fmt.Errorf("config key %s not found", key))
	}
	return c.v.GetString(key)
}

// GetStringSlice gets string slice value from config
func GetStringSlice(key string, defaultValue []string) (value []string) {
	return Default.GetStringSlice(key, defaultValue)
}

// GetStringSlice gets string slice value from config
func (c *Config) GetStringSlice(key string, defaultValue []string) (value []string) {
	return c.GetStringSliceVar(defaultValue, key)
}

// getStringSliceInternal gets string slice value from config
func (c *Config) getStringSliceInternal(key string, defaultValue []string) (value []string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetStringSlice(key)
}

// GetDuration gets duration value from config
func GetDuration(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	return Default.GetDuration(key, defaultValueInTimescaleUnits, timeScale)
}

// GetDuration gets duration value from config
func (c *Config) GetDuration(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	return c.GetDurationVar(defaultValueInTimescaleUnits, timeScale, key)
}

// getDurationInternal gets duration value from config
func (c *Config) getDurationInternal(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return time.Duration(defaultValueInTimescaleUnits) * timeScale
	} else {
		v := c.v.GetString(key)
		parseDuration, err := time.ParseDuration(v)
		if err == nil {
			return parseDuration
		} else {
			_, err = strconv.ParseFloat(v, 64)
			if err == nil {
				return c.v.GetDuration(key) * timeScale
			} else {
				return time.Duration(defaultValueInTimescaleUnits) * timeScale
			}
		}
	}
}

// IsSet checks if config is set for a key
func IsSet(key string) bool {
	return Default.IsSet(key)
}

// IsSet checks if config is set for a key
func (c *Config) IsSet(key string) bool {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	return c.isSetInternal(key)
}

// isSetInternal checks if config is set for a key. Caller needs to hold a read lock on vLock.
func (c *Config) isSetInternal(key string) bool {
	c.bindEnv(key)
	return c.v.IsSet(key)
}

// Override Config by application or command line

// Set override existing config
func Set(key string, value any) {
	Default.Set(key, value)
}

// Set override existing config
func (c *Config) Set(key string, value any) {
	c.vLock.Lock()
	c.v.Set(key, value)
	c.vLock.Unlock()
	c.onConfigChange()
}

func getReloadableMapKeys[T configTypes](defaultValue T, orderedKeys ...string) (string, string) {
	k := fmt.Sprintf("%T:%s", defaultValue, strings.Join(orderedKeys, ",")) // key is a combination of type and ordered keys
	return k, fmt.Sprintf("%v", defaultValue)                               // dvKey is the string representation of the default value
}

func getOrCreatePointer[T configTypes](
	reloadableVars map[string]any, reloadableVarsMisuses map[string]string, // this function MUST receive maps that are already initialized
	lock *sync.RWMutex, defaultValue T, orderedKeys ...string,
) (ptr *Reloadable[T], exists bool) {
	key, dvKey := getReloadableMapKeys(defaultValue, orderedKeys...)

	lock.Lock()
	defer lock.Unlock()

	defer func() {
		if _, ok := reloadableVarsMisuses[key]; !ok {
			reloadableVarsMisuses[key] = dvKey
		}
		if reloadableVarsMisuses[key] != dvKey {
			panic(fmt.Errorf(
				"detected misuse of config variable registered with different default values for %q: %+v - %+v",
				key, reloadableVarsMisuses[key], dvKey,
			))
		}
	}()

	if p, ok := reloadableVars[key]; ok {
		return p.(*Reloadable[T]), true
	}

	p := &Reloadable[T]{}
	reloadableVars[key] = p
	return p, false
}

// bindEnv handles rudder server's unique snake case replacement by registering
// the environment variables to viper, that would otherwise be ignored.
// Viper uppercases keys before sending them to its EnvKeyReplacer, thus
// the replacer cannot detect camelCase keys.
func (c *Config) bindEnv(key string) {
	envVar := ConfigKeyToEnv(c.envPrefix, key)
	// bind once
	c.envsLock.RLock()
	if _, ok := c.envs[key]; !ok {
		c.envsLock.RUnlock()
		c.envsLock.Lock() // don't really care about race here, setting the same value
		c.envs[strings.ToUpper(key)] = envVar
		c.envsLock.Unlock()
	} else {
		c.envsLock.RUnlock()
	}
}

type envReplacer struct {
	c *Config
}

func (r *envReplacer) Replace(s string) string {
	r.c.envsLock.RLock()
	defer r.c.envsLock.RUnlock()
	if v, ok := r.c.envs[s]; ok {
		return v
	}
	return s // bound environment variables
}

// Fallback environment variables supported (historically) by rudder-server
func bindLegacyEnv(v *viper.Viper) {
	_ = v.BindEnv("DB.host", "JOBS_DB_HOST")
	_ = v.BindEnv("DB.user", "JOBS_DB_USER")
	_ = v.BindEnv("DB.name", "JOBS_DB_DB_NAME")
	_ = v.BindEnv("DB.port", "JOBS_DB_PORT")
	_ = v.BindEnv("DB.password", "JOBS_DB_PASSWORD")
	_ = v.BindEnv("DB.sslMode", "JOBS_DB_SSL_MODE")
	_ = v.BindEnv("SharedDB.dsn", "SHARED_DB_DSN")
}

// registerNonReloadableConfigKeys tracks all non-reloadable config keys in their lowercase form
func (c *Config) registerNonReloadableConfigKeys(keys ...string) {
	c.nonReloadableKeysLock.Lock()
	defer c.nonReloadableKeysLock.Unlock()
	for _, key := range keys {
		c.nonReloadableKeys[strings.ToLower(key)] = key
	}
}

// RegisterObserver registers an observer for configuration changes
func (c *Config) RegisterObserver(observer Observer) {
	c.notifier.Register(observer)
}

// UnregisterObserver unregisters an observer
func (c *Config) UnregisterObserver(observer Observer) {
	c.notifier.Unregister(observer)
}

// OnReloadableConfigChanges registers a function to be called whenever a reloadable config changes happens
func (c *Config) OnReloadableConfigChange(fn func(key string, oldValue, newValue any)) {
	c.notifier.Register(ReloadableConfigChangesFunc(fn))
}

// OnNonReloadableConfigChanges registers a function to be called whenever a non-reloadable config change happens
func (c *Config) OnNonReloadableConfigChange(fn func(key string, op KeyOperation)) {
	c.notifier.Register(NonReloadableConfigChangesFunc(fn))
}
