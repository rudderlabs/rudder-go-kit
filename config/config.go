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
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
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
		envPrefix:                 DefaultEnvPrefix,
		hotReloadableConfig:       make(map[string]*configValue),
		hotReloadableVars:         make(map[string]any),
		hotReloadableVarsDefaults: make(map[string]string),
		nonReloadableConfig:       make(map[string]*configValue),
		nonReloadableKeys:         make(map[string]string),
		envs:                      make(map[string]string),

		notifier: &notifier{},
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

	hotReloadableConfigLock   sync.RWMutex            // protects hot reloadable maps
	hotReloadableConfig       map[string]*configValue // key -> <data-type>:<comma-separated list of config keys>, e.g. string:jobsdb.host, value -> configValue pointer
	hotReloadableVars         map[string]any          // key -> <data-type>:<comma-separated list of config keys>, e.g. string:jobsdb.host, value -> Reloadable[T] pointer
	hotReloadableVarsDefaults map[string]string       // key -> <data-type>:<comma-separated list of config keys>, e.g. string:jobsdb.host, value -> original default value as a string, e.g. "localhost:5432"

	enableNonReloadableAdvancedDetection bool                    // if true, non-reloadable config key changes are detected with a more advanced mechanism
	nonReloadableConfigLock              sync.RWMutex            // protects non hot reloadable maps
	nonReloadableConfig                  map[string]*configValue // key -> <data-type>:<comma-separated list of config keys><default-value>, e.g. string:jobsdb.host:localhost, value -> configValue pointer
	nonReloadableKeys                    map[string]string       // key -> <lowercase-config-key>, e.g. jobsdb.host, value -> original key, e.g. JobsDB.Host
	currentSettings                      map[string]any          // current config settings. Keys are always stored flattened and in lower case, e.g. jobsdb.host

	envsLock  sync.RWMutex // protects the envs map below
	envs      map[string]string
	envPrefix string // prefix for environment variables

	configPath    string
	configPathErr error
	godotEnvErr   error
	notifier      *notifier // for notifying subscribers of config changes
}

func (c *Config) load() {
	c.godotEnvErr = godotenv.Load()
	c.enableNonReloadableAdvancedDetection = getEnv("CONFIG_ADVANCED_DETECTION", "false") == "true"
	configPath := getEnv("CONFIG_PATH", "./config/config.yaml")

	v := viper.NewWithOptions(viper.EnvKeyReplacer(&envReplacer{c: c}))
	v.AutomaticEnv()
	bindLegacyEnv(v)

	v.SetConfigFile(configPath)

	// Find and read the config file
	// If config.yaml is not found or error with parsing. Use the default config values instead
	c.configPathErr = v.ReadInConfig()
	c.configPath = v.ConfigFileUsed()
	c.v = v

	c.currentSettings = c.getCurrentSettings()
	c.v.OnConfigChange(func(_ fsnotify.Event) {
		c.onConfigChange()
	})
	c.v.WatchConfig()
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

// ConfigFileUsed returns the file used to load the config.
// If we failed to load the config file, it also returns an error.
func (c *Config) ConfigFileUsed() (string, error) {
	return c.configPath, c.configPathErr
}

// DotEnvLoaded returns an error if there was an error loading the .env file.
// It returns nil otherwise.
func (c *Config) DotEnvLoaded() error {
	return c.godotEnvErr
}

// getMapKey returns the map key (<type>:<comma-separated-keys>) along with the string representation of the default value
func getMapKey[T configTypes](defaultValue T, orderedKeys ...string) (string, string) {
	mapKey := getTypeName(defaultValue) + ":" + strings.Join(orderedKeys, ",") // key is a combination of type and ordered keys
	defaultValueStr := getStringValue(defaultValue)
	return mapKey, defaultValueStr
}

func getOrCreatePointer[T configTypes](
	reloadableVars map[string]any, reloadableVarsMisuses map[string]string, reloadableConfigVals map[string]*configValue, // this function MUST receive maps that are already initialized
	lock *sync.RWMutex, defaultValue T, cv *configValue, orderedKeys ...string,
) (ptr *Reloadable[T], exists bool) {
	key, dv := getMapKey(defaultValue, orderedKeys...)
	lock.Lock()
	defer lock.Unlock()
	defer func() {
		if _, ok := reloadableVarsMisuses[key]; !ok {
			reloadableVarsMisuses[key] = dv
		}
		if reloadableVarsMisuses[key] != dv {
			panic(fmt.Errorf(
				"detected misuse of config variable registered with different default values for %q: %+v - %+v",
				key, reloadableVarsMisuses[key], dv,
			))
		}
	}()
	if p, ok := reloadableVars[key]; ok {
		return p.(*Reloadable[T]), true
	}
	p := &Reloadable[T]{}
	reloadableVars[key] = p
	cv.value = p
	reloadableConfigVals[key] = cv
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
func registerNonReloadableConfigKeys[T configTypes](c *Config, dv T, cv *configValue) {
	c.nonReloadableConfigLock.Lock()
	defer c.nonReloadableConfigLock.Unlock()

	if !c.enableNonReloadableAdvancedDetection {
		for _, key := range cv.keys {
			c.nonReloadableKeys[strings.ToLower(key)] = key // store the original key in lowercase
		}
		return
	}
	key, dvKey := getMapKey(dv, cv.keys...)
	// final key should be a combination of type, ordered keys & default value
	k := key + ":" + dvKey // TODO: consider ignoring default value for non-reloadable keys
	if _, exists := c.nonReloadableConfig[k]; !exists {
		c.nonReloadableConfig[k] = cv
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

// OnReloadableConfigChange registers a function to be called whenever a reloadable config change happens
func (c *Config) OnReloadableConfigChange(fn func(key string, oldValue, newValue any)) {
	c.notifier.Register(ReloadableConfigChangesFunc(fn))
}

// OnNonReloadableConfigChange registers a function to be called whenever a non-reloadable config change happens
func (c *Config) OnNonReloadableConfigChange(fn func(key string)) {
	c.notifier.Register(NonReloadableConfigChangesFunc(fn))
}
