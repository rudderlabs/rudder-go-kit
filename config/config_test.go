package config

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_Getters_Existing_and_Default(t *testing.T) {
	tc := New()
	tc.Set("string", "string")
	require.Equal(t, "string", tc.GetString("string", "default"), "it should return the key value")
	require.Equal(t, "default", tc.GetString("other", "default"), "it should return the default value")

	tc.Set("bool", false)
	require.Equal(t, false, tc.GetBool("bool", true), "it should return the key value")
	require.Equal(t, true, tc.GetBool("other", true), "it should return the default value")

	tc.Set("int", 0)
	require.Equal(t, 0, tc.GetInt("int", 1), "it should return the key value")
	require.Equal(t, 1, tc.GetInt("other", 1), "it should return the default value")
	require.EqualValues(t, 0, tc.GetInt64("int", 1), "it should return the key value")
	require.EqualValues(t, 1, tc.GetInt64("other", 1), "it should return the default value")

	tc.Set("float", 0.0)
	require.EqualValues(t, 0, tc.GetFloat64("float", 1), "it should return the key value")
	require.EqualValues(t, 1, tc.GetFloat64("other", 1), "it should return the default value")

	tc.Set("stringslice", []string{"string", "string"})
	require.Equal(t, []string{"string", "string"}, tc.GetStringSlice("stringslice", []string{"default"}), "it should return the key value")
	require.Equal(t, []string{"default"}, tc.GetStringSlice("other", []string{"default"}), "it should return the default value")

	tc.Set("duration", "2ms")
	require.Equal(t, 2*time.Millisecond, tc.GetDuration("duration", 1, time.Second), "it should return the key value")
	require.Equal(t, time.Second, tc.GetDuration("other", 1, time.Second), "it should return the default value")

	tc.Set("duration", "2")
	require.Equal(t, 2*time.Second, tc.GetDuration("duration", 1, time.Second), "it should return the key value")
	require.Equal(t, time.Second, tc.GetDuration("other", 1, time.Second), "it should return the default value")

	tc.Set("stringmap", map[string]interface{}{"string": "any"})
	require.Equal(t, map[string]interface{}{"string": "any"}, tc.GetStringMap("stringmap", map[string]interface{}{"default": "value"}), "it should return the key value")
	require.Equal(t, map[string]interface{}{"default": "value"}, tc.GetStringMap("other", map[string]interface{}{"default": "value"}), "it should return the default value")
}

func Test_MustGet(t *testing.T) {
	tc := New()
	tc.Set("string", "string")
	require.Equal(t, "string", tc.MustGetString("string"), "it should return the key value")
	require.Panics(t, func() { tc.MustGetString("other") })

	tc.Set("int", 0)
	require.Equal(t, 0, tc.MustGetInt("int"), "it should return the key value")
	require.Panics(t, func() { tc.MustGetInt("other") })
}

func Test_Register_Existing_and_Default(t *testing.T) {
	tc := New()
	tc.Set("string", "string")
	var stringValue string
	var otherStringValue string
	tc.RegisterStringConfigVariable("default", &stringValue, false, "string")
	require.Equal(t, "string", stringValue, "it should return the key value")
	tc.RegisterStringConfigVariable("default", &otherStringValue, false, "other")
	require.Equal(t, "default", otherStringValue, "it should return the default value")

	tc.Set("bool", false)
	var boolValue bool
	var otherBoolValue bool
	tc.RegisterBoolConfigVariable(true, &boolValue, false, "bool")
	require.Equal(t, false, boolValue, "it should return the key value")
	tc.RegisterBoolConfigVariable(true, &otherBoolValue, false, "other")
	require.Equal(t, true, otherBoolValue, "it should return the default value")

	tc.Set("int", 0)
	var intValue int
	var otherIntValue int
	var int64Value int64
	var otherInt64Value int64
	tc.RegisterIntConfigVariable(1, &intValue, false, 1, "int")
	require.Equal(t, 0, intValue, "it should return the key value")
	tc.RegisterIntConfigVariable(1, &otherIntValue, false, 1, "other")
	require.Equal(t, 1, otherIntValue, "it should return the default value")
	tc.RegisterInt64ConfigVariable(1, &int64Value, false, 1, "int")
	require.EqualValues(t, 0, int64Value, "it should return the key value")
	tc.RegisterInt64ConfigVariable(1, &otherInt64Value, false, 1, "other")
	require.EqualValues(t, 1, otherInt64Value, "it should return the default value")

	tc.Set("float", 0.0)
	var floatValue float64
	var otherFloatValue float64
	tc.RegisterFloat64ConfigVariable(1, &floatValue, false, "float")
	require.EqualValues(t, 0, floatValue, "it should return the key value")
	tc.RegisterFloat64ConfigVariable(1, &otherFloatValue, false, "other")
	require.EqualValues(t, 1, otherFloatValue, "it should return the default value")

	tc.Set("stringslice", []string{"string", "string"})
	var stringSliceValue []string
	var otherStringSliceValue []string
	tc.RegisterStringSliceConfigVariable([]string{"default"}, &stringSliceValue, false, "stringslice")
	require.Equal(t, []string{"string", "string"}, stringSliceValue, "it should return the key value")
	tc.RegisterStringSliceConfigVariable([]string{"default"}, &otherStringSliceValue, false, "other")
	require.Equal(t, []string{"default"}, otherStringSliceValue, "it should return the default value")

	tc.Set("duration", "2ms")
	var durationValue time.Duration
	var otherDurationValue time.Duration
	tc.RegisterDurationConfigVariable(1, &durationValue, false, time.Second, "duration")
	require.Equal(t, 2*time.Millisecond, durationValue, "it should return the key value")
	tc.RegisterDurationConfigVariable(1, &otherDurationValue, false, time.Second, "other")
	require.Equal(t, time.Second, otherDurationValue, "it should return the default value")

	tc.Set("stringmap", map[string]interface{}{"string": "any"})
	var stringMapValue map[string]interface{}
	var otherStringMapValue map[string]interface{}
	tc.RegisterStringMapConfigVariable(map[string]interface{}{"default": "value"}, &stringMapValue, false, "stringmap")
	require.Equal(t, map[string]interface{}{"string": "any"}, stringMapValue, "it should return the key value")
	tc.RegisterStringMapConfigVariable(map[string]interface{}{"default": "value"}, &otherStringMapValue, false, "other")
	require.Equal(t, map[string]interface{}{"default": "value"}, otherStringMapValue, "it should return the default value")
}

func TestStatic_checkAndHotReloadConfig(t *testing.T) {
	configMap := make(map[string][]*configValue)

	var var1 string
	var var2 string
	configVar1 := newConfigValue(&var1, 1, "var1", []string{"keyVar"})
	configVar2 := newConfigValue(&var2, 1, "var2", []string{"keyVar"})

	configMap["keyVar"] = []*configValue{configVar1, configVar2}
	t.Setenv("RSERVER_KEY_VAR", "value_changed")

	Default.checkAndHotReloadConfig(configMap)

	varptr1 := configVar1.value.(*string)
	varptr2 := configVar2.value.(*string)
	require.Equal(t, *varptr1, "value_changed")
	require.Equal(t, *varptr2, "value_changed")
}

func TestCheckAndHotReloadConfig(t *testing.T) {
	var (
		stringValue            string
		stringConfigValue      = newConfigValue(&stringValue, nil, "default", []string{"string"})
		boolValue              bool
		boolConfigValue        = newConfigValue(&boolValue, nil, false, []string{"bool"})
		intValue               int
		intConfigValue         = newConfigValue(&intValue, 1, 0, []string{"int"})
		int64Value             int64
		int64ConfigValue       = newConfigValue(&int64Value, int64(1), int64(0), []string{"int64"})
		float64Value           float64
		float64ConfigValue     = newConfigValue(&float64Value, 1.0, 0.0, []string{"float64"})
		stringSliceValue       []string
		stringSliceConfigValue = newConfigValue(&stringSliceValue, nil, []string{"default"}, []string{"stringslice"})
		durationValue          time.Duration
		durationConfigValue    = newConfigValue(&durationValue, time.Second, int64(1), []string{"duration"})
		stringMapValue         map[string]interface{}
		stringMapConfigValue   = newConfigValue(&stringMapValue, nil, map[string]interface{}{"default": "value"}, []string{"stringmap"})
	)

	t.Run("with envs", func(t *testing.T) {
		t.Setenv("RSERVER_INT", "1")
		t.Setenv("RSERVER_INT64", "1")
		t.Setenv("RSERVER_STRING", "string")
		t.Setenv("RSERVER_DURATION", "2s")
		t.Setenv("RSERVER_BOOL", "true")
		t.Setenv("RSERVER_FLOAT64", "1.0")
		t.Setenv("RSERVER_STRINGSLICE", "string string")
		t.Setenv("RSERVER_STRINGMAP", "{\"string\":\"any\"}")

		Default.checkAndHotReloadConfig(map[string][]*configValue{
			"string":      {stringConfigValue},
			"bool":        {boolConfigValue},
			"int":         {intConfigValue},
			"int64":       {int64ConfigValue},
			"float64":     {float64ConfigValue},
			"stringslice": {stringSliceConfigValue},
			"duration":    {durationConfigValue},
			"stringmap":   {stringMapConfigValue},
		})

		require.Equal(t, *stringConfigValue.value.(*string), "string")
		require.Equal(t, *boolConfigValue.value.(*bool), true)
		require.Equal(t, *intConfigValue.value.(*int), 1)
		require.Equal(t, *int64ConfigValue.value.(*int64), int64(1))
		require.Equal(t, *float64ConfigValue.value.(*float64), 1.0)
		require.Equal(t, *durationConfigValue.value.(*time.Duration), 2*time.Second)
		require.Equal(t, *stringSliceConfigValue.value.(*[]string), []string{"string", "string"})
		require.Equal(t, *stringMapConfigValue.value.(*map[string]any), map[string]any{"string": "any"})
	})

	t.Run("without envs", func(t *testing.T) {
		Default.checkAndHotReloadConfig(map[string][]*configValue{
			"string":      {stringConfigValue},
			"bool":        {boolConfigValue},
			"int":         {intConfigValue},
			"int64":       {int64ConfigValue},
			"float64":     {float64ConfigValue},
			"stringslice": {stringSliceConfigValue},
			"duration":    {durationConfigValue},
			"stringmap":   {stringMapConfigValue},
		})

		require.Equal(t, *stringConfigValue.value.(*string), "default")
		require.Equal(t, *boolConfigValue.value.(*bool), false)
		require.Equal(t, *intConfigValue.value.(*int), 0)
		require.Equal(t, *int64ConfigValue.value.(*int64), int64(0))
		require.Equal(t, *float64ConfigValue.value.(*float64), 0.0)
		require.Equal(t, *durationConfigValue.value.(*time.Duration), 1*time.Second)
		require.Equal(t, *stringSliceConfigValue.value.(*[]string), []string{"default"})
		require.Equal(t, *stringMapConfigValue.value.(*map[string]any), map[string]any{"default": "value"})
	})
}

func TestNewReloadableAPI(t *testing.T) {
	t.Run("non reloadable", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			c := New()
			v := c.GetIntVar(5, 1, t.Name())
			require.Equal(t, 5, v)
		})
		t.Run("int64", func(t *testing.T) {
			c := New()
			v := c.GetInt64Var(5, 1, t.Name())
			require.EqualValues(t, 5, v)
		})
		t.Run("bool", func(t *testing.T) {
			c := New()
			v := c.GetBoolVar(true, t.Name())
			require.True(t, v)
		})
		t.Run("float64", func(t *testing.T) {
			c := New()
			v := c.GetFloat64Var(0.123, t.Name())
			require.EqualValues(t, 0.123, v)
		})
		t.Run("string", func(t *testing.T) {
			c := New()
			v := c.GetStringVar("foo", t.Name())
			require.Equal(t, "foo", v)
		})
		t.Run("duration", func(t *testing.T) {
			c := New()
			v := c.GetDurationVar(123, time.Second, t.Name())
			require.Equal(t, 123*time.Second, v)
		})
		t.Run("[]string", func(t *testing.T) {
			c := New()
			v := c.GetStringSliceVar([]string{"a", "b"}, t.Name())
			require.NotNil(t, v)
			require.Equal(t, []string{"a", "b"}, v)

			c.Set(t.Name(), []string{"c", "d"})
			require.Equal(t, []string{"a", "b"}, v, "variable is not reloadable")
		})
		t.Run("map[string]interface{}", func(t *testing.T) {
			c := New()
			v := c.GetStringMapVar(map[string]interface{}{"a": 1, "b": 2}, t.Name())
			require.NotNil(t, v)
			require.Equal(t, map[string]interface{}{"a": 1, "b": 2}, v)

			c.Set(t.Name(), map[string]interface{}{"c": 3, "d": 4})
			require.Equal(t, map[string]interface{}{"a": 1, "b": 2}, v, "variable is not reloadable")
		})
	})
	t.Run("reloadable", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			c := New()
			v := c.GetReloadableIntVar(5, 1, t.Name())
			require.Equal(t, 5, v.Load())

			c.Set(t.Name(), 10)
			require.Equal(t, 10, v.Load())

			c.Set(t.Name(), 10)
			require.Equal(t, 10, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"int:TestNewReloadableAPI/reloadable/int:5 - "+
					"int:TestNewReloadableAPI/reloadable/int:10\n",
				func() {
					// changing just the valueScale also changes the default value
					_ = c.GetReloadableIntVar(5, 2, t.Name())
				},
			)
		})
		t.Run("int64", func(t *testing.T) {
			c := New()
			v := c.GetReloadableInt64Var(5, 1, t.Name())
			require.EqualValues(t, 5, v.Load())

			c.Set(t.Name(), 10)
			require.EqualValues(t, 10, v.Load())

			c.Set(t.Name(), 10)
			require.EqualValues(t, 10, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"int64:TestNewReloadableAPI/reloadable/int64:5 - "+
					"int64:TestNewReloadableAPI/reloadable/int64:10\n",
				func() {
					// changing just the valueScale also changes the default value
					_ = c.GetReloadableInt64Var(5, 2, t.Name())
				},
			)
		})
		t.Run("bool", func(t *testing.T) {
			c := New()
			v := c.GetReloadableBoolVar(true, t.Name())
			require.True(t, v.Load())

			c.Set(t.Name(), false)
			require.False(t, v.Load())

			c.Set(t.Name(), false)
			require.False(t, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"bool:TestNewReloadableAPI/reloadable/bool:true - "+
					"bool:TestNewReloadableAPI/reloadable/bool:false\n",
				func() {
					_ = c.GetReloadableBoolVar(false, t.Name())
				},
			)
		})
		t.Run("float64", func(t *testing.T) {
			c := New()
			v := c.GetReloadableFloat64Var(0.123, t.Name())
			require.EqualValues(t, 0.123, v.Load())

			c.Set(t.Name(), 4.567)
			require.EqualValues(t, 4.567, v.Load())

			c.Set(t.Name(), 4.567)
			require.EqualValues(t, 4.567, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"float64:TestNewReloadableAPI/reloadable/float64:0.123 - "+
					"float64:TestNewReloadableAPI/reloadable/float64:0.1234\n",
				func() {
					_ = c.GetReloadableFloat64Var(0.1234, t.Name())
				},
			)
		})
		t.Run("string", func(t *testing.T) {
			c := New()
			v := c.GetReloadableStringVar("foo", t.Name())
			require.Equal(t, "foo", v.Load())

			c.Set(t.Name(), "bar")
			require.EqualValues(t, "bar", v.Load())

			c.Set(t.Name(), "bar")
			require.EqualValues(t, "bar", v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"string:TestNewReloadableAPI/reloadable/string:foo - "+
					"string:TestNewReloadableAPI/reloadable/string:qux\n",
				func() {
					_ = c.GetReloadableStringVar("qux", t.Name())
				},
			)
		})
		t.Run("duration", func(t *testing.T) {
			c := New()
			v := c.GetReloadableDurationVar(123, time.Nanosecond, t.Name())
			require.Equal(t, 123*time.Nanosecond, v.Load())

			c.Set(t.Name(), 456*time.Millisecond)
			require.Equal(t, 456*time.Millisecond, v.Load())

			c.Set(t.Name(), 456*time.Millisecond)
			require.Equal(t, 456*time.Millisecond, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"time.Duration:TestNewReloadableAPI/reloadable/duration:123ns - "+
					"time.Duration:TestNewReloadableAPI/reloadable/duration:2m3s\n",
				func() {
					_ = c.GetReloadableDurationVar(123, time.Second, t.Name())
				},
			)
		})
		t.Run("[]string", func(t *testing.T) {
			c := New()
			v := c.GetReloadableStringSliceVar([]string{"a", "b"}, t.Name())
			require.Equal(t, []string{"a", "b"}, v.Load())

			c.Set(t.Name(), []string{"c", "d"})
			require.Equal(t, []string{"c", "d"}, v.Load())

			c.Set(t.Name(), []string{"c", "d"})
			require.Equal(t, []string{"c", "d"}, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"[]string:TestNewReloadableAPI/reloadable/[]string:[a b] - "+
					"[]string:TestNewReloadableAPI/reloadable/[]string:[a b c]\n",
				func() {
					_ = c.GetReloadableStringSliceVar([]string{"a", "b", "c"}, t.Name())
				},
			)
		})
		t.Run("map[string]interface{}", func(t *testing.T) {
			c := New()
			v := c.GetReloadableStringMapVar(map[string]interface{}{"a": 1, "b": 2}, t.Name())
			require.Equal(t, map[string]interface{}{"a": 1, "b": 2}, v.Load())

			c.Set(t.Name(), map[string]interface{}{"c": 3, "d": 4})
			require.Equal(t, map[string]interface{}{"c": 3, "d": 4}, v.Load())

			c.Set(t.Name(), map[string]interface{}{"c": 3, "d": 4})
			require.Equal(t, map[string]interface{}{"c": 3, "d": 4}, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"Detected misuse of config variable registered with different default values "+
					"map[string]interface {}:TestNewReloadableAPI/reloadable/map[string]interface{}:map[a:1 b:2] - "+
					"map[string]interface {}:TestNewReloadableAPI/reloadable/map[string]interface{}:map[a:2 b:1]\n",
				func() {
					_ = c.GetReloadableStringMapVar(map[string]interface{}{"a": 2, "b": 1}, t.Name())
				},
			)
		})
	})
}

func TestGetOrCreatePointer(t *testing.T) {
	var (
		m   = make(map[string]any)
		dvs = make(map[string]string)
		rwm sync.RWMutex
	)
	p1, exists := getOrCreatePointer(m, dvs, &rwm, 123, "foo", "bar")
	require.NotNil(t, p1)
	require.False(t, exists)

	p2, exists := getOrCreatePointer(m, dvs, &rwm, 123, "foo", "bar")
	require.True(t, p1 == p2)
	require.True(t, exists)

	p3, exists := getOrCreatePointer(m, dvs, &rwm, 123, "bar", "foo")
	require.True(t, p1 != p3)
	require.False(t, exists)

	p4, exists := getOrCreatePointer(m, dvs, &rwm, 123, "bar", "foo", "qux")
	require.True(t, p1 != p4)
	require.False(t, exists)

	require.PanicsWithError(t,
		"Detected misuse of config variable registered with different default values "+
			"int:bar,foo,qux:123 - int:bar,foo,qux:456\n",
		func() {
			getOrCreatePointer(m, dvs, &rwm, 456, "bar", "foo", "qux")
		},
	)
}

func TestReloadable(t *testing.T) {
	t.Run("scalar", func(t *testing.T) {
		var v Reloadable[int]
		v.store(123)
		require.Equal(t, 123, v.Load())
	})
	t.Run("nullable", func(t *testing.T) {
		var v Reloadable[[]string]
		require.Nil(t, v.Load())

		v.store(nil)
		require.Nil(t, v.Load())

		v.store([]string{"foo", "bar"})
		require.Equal(t, v.Load(), []string{"foo", "bar"})

		v.store(nil)
		require.Nil(t, v.Load())
	})
}

func TestConfigKeyToEnv(t *testing.T) {
	expected := "RSERVER_KEY_VAR1_VAR2"
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "Key.Var1.Var2"))
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "key.var1.var2"))
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "KeyVar1Var2"))
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "RSERVER_KEY_VAR1_VAR2"))
	require.Equal(t, "KEY_VAR1_VAR2", ConfigKeyToEnv(DefaultEnvPrefix, "KEY_VAR1_VAR2"))
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "Key_Var1.Var2"))
	require.Equal(t, expected, ConfigKeyToEnv(DefaultEnvPrefix, "key_Var1_Var2"))
	require.Equal(t, "RSERVER_KEY_VAR_1_VAR2", ConfigKeyToEnv(DefaultEnvPrefix, "Key_Var.1.Var2"))
}

func TestGetEnvThroughViper(t *testing.T) {
	expectedValue := "VALUE"

	t.Run("detects dots", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR2", expectedValue)
		tc := New()
		require.Equal(t, expectedValue, tc.GetString("Key.Var1.Var2", ""))
	})

	t.Run("detects camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR2", expectedValue)
		tc := New()
		require.Equal(t, expectedValue, tc.GetString("KeyVar1Var2", ""))
	})

	t.Run("detects dots with camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR_VAR", expectedValue)
		tc := New()
		require.Equal(t, expectedValue, tc.GetString("Key.Var1VarVar", ""))
	})

	t.Run("with env prefix", func(t *testing.T) {
		envPrefix := "ANOTHER_PREFIX"

		tc := New(WithEnvPrefix(envPrefix))
		t.Setenv(envPrefix+"_KEY_VAR1_VAR_VAR", expectedValue)

		require.Equal(t, expectedValue, tc.GetString("Key.Var1VarVar", ""))
	})

	t.Run("detects uppercase env variables", func(t *testing.T) {
		t.Setenv("SOMEENVVARIABLE", expectedValue)
		tc := New()
		require.Equal(t, expectedValue, tc.GetString("SOMEENVVARIABLE", ""))

		t.Setenv("SOME_ENV_VARIABLE", expectedValue)
		require.Equal(t, expectedValue, tc.GetString("SOME_ENV_VARIABLE", ""))

		t.Setenv("SOME_ENV_VARIABLE12", expectedValue)
		require.Equal(t, expectedValue, tc.GetString("SOME_ENV_VARIABLE12", ""))
	})

	t.Run("doesn't use viper's default env var matcher (uppercase)", func(t *testing.T) {
		t.Setenv("KEYVAR1VARVAR", expectedValue)
		tc := New()
		require.Equal(t, "", tc.GetString("KeyVar1VarVar", ""))
	})

	t.Run("can retrieve legacy env", func(t *testing.T) {
		t.Setenv("JOBS_DB_HOST", expectedValue)
		tc := New()
		require.Equal(t, expectedValue, tc.GetString("DB.host", ""))
	})
}

func TestRegisterEnvThroughViper(t *testing.T) {
	expectedValue := "VALUE"

	t.Run("detects dots", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR2", expectedValue)
		tc := New()
		var v string
		tc.RegisterStringConfigVariable("", &v, true, "Key.Var1.Var2")
		require.Equal(t, expectedValue, v)
	})

	t.Run("detects camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR_VAR", expectedValue)
		tc := New()
		var v string
		tc.RegisterStringConfigVariable("", &v, true, "KeyVarVar")
		require.Equal(t, expectedValue, v)
	})

	t.Run("detects dots with camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR_VAR", expectedValue)
		tc := New()
		var v string
		tc.RegisterStringConfigVariable("", &v, true, "Key.Var1VarVar")
		require.Equal(t, expectedValue, v)
	})
}

func Test_Set_CaseInsensitive(t *testing.T) {
	tc := New()
	tc.Set("sTrIng.One", "string")
	require.Equal(t, "string", tc.GetString("String.one", "default"), "it should return the key value")
}

func Test_Misc(t *testing.T) {
	t.Setenv("KUBE_NAMESPACE", "value")
	require.Equal(t, "value", GetKubeNamespace())

	t.Setenv("KUBE_NAMESPACE", "")
	require.Equal(t, "none", GetNamespaceIdentifier())

	t.Setenv("WORKSPACE_TOKEN", "value1")
	t.Setenv("CONFIG_BACKEND_TOKEN", "value2")
	require.Equal(t, "value1", GetWorkspaceToken())

	t.Setenv("WORKSPACE_TOKEN", "")
	t.Setenv("CONFIG_BACKEND_TOKEN", "value2")
	require.Equal(t, "value2", GetWorkspaceToken())

	t.Setenv("RELEASE_NAME", "value")
	require.Equal(t, "value", GetReleaseName())
}

func TestConfigLocking(t *testing.T) {
	const (
		timeout   = 2 * time.Second
		configKey = "test"
	)
	c := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	doWithTimeout := func(f func(), timeout time.Duration) error {
		var err error
		var closed bool
		var closedMutex sync.Mutex
		wait := make(chan struct{})
		t := time.NewTimer(timeout)

		go func() {
			<-t.C
			err = fmt.Errorf("timeout after %s", timeout)
			closedMutex.Lock()
			if !closed {
				closed = true
				close(wait)
			}
			closedMutex.Unlock()
		}()

		go func() {
			f()
			t.Stop()
			closedMutex.Lock()
			if !closed {
				closed = true
				close(wait)
			}
			closedMutex.Unlock()
		}()
		<-wait
		return err
	}

	startOperation := func(name string, op func()) {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					if err := doWithTimeout(op, timeout); err != nil {
						return fmt.Errorf("%s: %w", name, err)
					}
				}
			}
		})
	}

	startOperation("set the config value", func() { c.Set(configKey, "value1") })
	startOperation("try to read the config value using GetString", func() { _ = c.GetString(configKey, "") })
	startOperation("try to read the config value using GetStringVar", func() { _ = c.GetStringVar(configKey, "") })
	startOperation("try to read the config value using GetReloadableStringVar", func() {
		r := c.GetReloadableStringVar("", configKey)
		_ = r.Load()
	})

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			cancel()
			return nil
		}
	})

	require.NoError(t, g.Wait())
}

func TestConfigLoad(t *testing.T) {
	// create a temporary file:
	f, err := os.CreateTemp("", "*config.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	t.Setenv("CONFIG_PATH", f.Name())
	c := New()
	require.NoError(t, err)

	t.Run("successfully loaded config file", func(t *testing.T) {
		configFile, err := c.ConfigFileUsed()
		require.NoError(t, err)
		require.Equal(t, f.Name(), configFile)
	})

	err = os.Remove(f.Name())
	require.NoError(t, err)

	t.Run("attempt to load non-existent config file", func(t *testing.T) {
		c := New()
		configFile, err := c.ConfigFileUsed()
		require.Error(t, err)
		require.Equal(t, f.Name(), configFile)
	})

	t.Run("dot env error", func(t *testing.T) {
		c := New()
		err := c.DotEnvLoaded()
		require.Error(t, err)
	})

	t.Run("dot env found", func(t *testing.T) {
		c := New()
		f, err := os.Create(".env")
		require.NoError(t, err)
		defer os.Remove(f.Name())

		err = c.DotEnvLoaded()
		require.Error(t, err)
	})
}

// Benchmark for the original ConfigKeyToEnv function
func BenchmarkConfigKeyToEnv(b *testing.B) {
	envPrefix := "MYAPP"
	configKey := "myConfig.KeyName"
	for i := 0; i < b.N; i++ {
		_ = ConfigKeyToEnv(envPrefix, configKey)
	}
	b.ReportAllocs()
}
