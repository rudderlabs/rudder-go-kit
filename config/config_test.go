package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
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

	tc.Set("stringmap", map[string]any{"string": "any"})
	require.Equal(t, map[string]any{"string": "any"}, tc.GetStringMap("stringmap", map[string]any{"default": "value"}), "it should return the key value")
	require.Equal(t, map[string]any{"default": "value"}, tc.GetStringMap("other", map[string]any{"default": "value"}), "it should return the default value")
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
	stringValue := tc.GetStringVar("default", "string")
	require.Equal(t, "string", stringValue, "it should return the key value")
	otherStringValue := tc.GetStringVar("default", "other")
	require.Equal(t, "default", otherStringValue, "it should return the default value")

	tc.Set("bool", false)
	boolValue := tc.GetBoolVar(true, "bool")
	require.Equal(t, false, boolValue, "it should return the key value")
	otherBoolValue := tc.GetBoolVar(true, "other")
	require.Equal(t, true, otherBoolValue, "it should return the default value")

	tc.Set("int", 0)
	intValue := tc.GetIntVar(1, 1, "int")
	require.Equal(t, 0, intValue, "it should return the key value")
	otherIntValue := tc.GetIntVar(1, 1, "other")
	require.Equal(t, 1, otherIntValue, "it should return the default value")
	int64Value := tc.GetInt64Var(1, 1, "int")
	require.EqualValues(t, 0, int64Value, "it should return the key value")
	otherInt64Value := tc.GetInt64Var(1, 1, "other")
	require.EqualValues(t, 1, otherInt64Value, "it should return the default value")

	tc.Set("float", 0.0)
	floatValue := tc.GetFloat64Var(1, "float")
	require.EqualValues(t, 0, floatValue, "it should return the key value")
	otherFloatValue := tc.GetFloat64Var(1, "other")
	require.EqualValues(t, 1, otherFloatValue, "it should return the default value")

	tc.Set("stringslice", []string{"string", "string"})
	stringSliceValue := tc.GetStringSliceVar([]string{"default"}, "stringslice")
	require.Equal(t, []string{"string", "string"}, stringSliceValue, "it should return the key value")
	otherStringSliceValue := tc.GetStringSliceVar([]string{"default"}, "other")
	require.Equal(t, []string{"default"}, otherStringSliceValue, "it should return the default value")

	tc.Set("duration", "2ms")
	durationValue := tc.GetDurationVar(1, time.Second, "duration")
	require.Equal(t, 2*time.Millisecond, durationValue, "it should return the key value")
	otherDurationValue := tc.GetDurationVar(1, time.Second, "other")
	require.Equal(t, time.Second, otherDurationValue, "it should return the default value")

	tc.Set("stringmap", map[string]any{"string": "any"})
	stringMapValue := tc.GetStringMapVar(map[string]any{"default": "value"}, "stringmap")
	require.Equal(t, map[string]any{"string": "any"}, stringMapValue, "it should return the key value")
	otherStringMapValue := tc.GetStringMapVar(map[string]any{"default": "value"}, "other")
	require.Equal(t, map[string]any{"default": "value"}, otherStringMapValue, "it should return the default value")
}

func TestStatic_checkAndHotReloadConfig(t *testing.T) {
	configMap := make(map[string][]*configValue)

	var var1 Reloadable[string]
	var var2 Reloadable[string]
	configVar1 := newConfigValue(&var1, 1, "var1", []string{"keyVar"})
	configVar2 := newConfigValue(&var2, 1, "var2", []string{"keyVar"})

	configMap["keyVar"] = []*configValue{configVar1, configVar2}
	t.Setenv("RSERVER_KEY_VAR", "value_changed")

	Default.checkAndHotReloadConfig(configMap)

	varptr1 := configVar1.value.(*Reloadable[string])
	varptr2 := configVar2.value.(*Reloadable[string])
	require.Equal(t, varptr1.Load(), "value_changed")
	require.Equal(t, varptr2.Load(), "value_changed")
}

func TestCheckAndHotReloadConfig(t *testing.T) {
	var (
		stringValue            Reloadable[string]
		stringConfigValue      = newConfigValue(&stringValue, nil, "default", []string{"string"})
		boolValue              Reloadable[bool]
		boolConfigValue        = newConfigValue(&boolValue, nil, false, []string{"bool"})
		intValue               Reloadable[int]
		intConfigValue         = newConfigValue(&intValue, 1, 0, []string{"int"})
		int64Value             Reloadable[int64]
		int64ConfigValue       = newConfigValue(&int64Value, int64(1), int64(0), []string{"int64"})
		float64Value           Reloadable[float64]
		float64ConfigValue     = newConfigValue(&float64Value, 1.0, 0.0, []string{"float64"})
		stringSliceValue       Reloadable[[]string]
		stringSliceConfigValue = newConfigValue(&stringSliceValue, nil, []string{"default"}, []string{"stringslice"})
		durationValue          Reloadable[time.Duration]
		durationConfigValue    = newConfigValue(&durationValue, time.Second, int64(1), []string{"duration"})
		stringMapValue         Reloadable[map[string]any]
		stringMapConfigValue   = newConfigValue(&stringMapValue, nil, map[string]any{"default": "value"}, []string{"stringmap"})
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

		require.Equal(t, stringConfigValue.value.(*Reloadable[string]).Load(), "string")
		require.Equal(t, boolConfigValue.value.(*Reloadable[bool]).Load(), true)
		require.Equal(t, intConfigValue.value.(*Reloadable[int]).Load(), 1)
		require.Equal(t, int64ConfigValue.value.(*Reloadable[int64]).Load(), int64(1))
		require.Equal(t, float64ConfigValue.value.(*Reloadable[float64]).Load(), 1.0)
		require.Equal(t, durationConfigValue.value.(*Reloadable[time.Duration]).Load(), 2*time.Second)
		require.Equal(t, stringSliceConfigValue.value.(*Reloadable[[]string]).Load(), []string{"string", "string"})
		require.Equal(t, stringMapConfigValue.value.(*Reloadable[map[string]any]).Load(), map[string]any{"string": "any"})
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

		require.Equal(t, stringConfigValue.value.(*Reloadable[string]).Load(), "default")
		require.Equal(t, boolConfigValue.value.(*Reloadable[bool]).Load(), false)
		require.Equal(t, intConfigValue.value.(*Reloadable[int]).Load(), 0)
		require.Equal(t, int64ConfigValue.value.(*Reloadable[int64]).Load(), int64(0))
		require.Equal(t, float64ConfigValue.value.(*Reloadable[float64]).Load(), 0.0)
		require.Equal(t, durationConfigValue.value.(*Reloadable[time.Duration]).Load(), 1*time.Second)
		require.Equal(t, stringSliceConfigValue.value.(*Reloadable[[]string]).Load(), []string{"default"})
		require.Equal(t, stringMapConfigValue.value.(*Reloadable[map[string]any]).Load(), map[string]any{"default": "value"})
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
		t.Run("map[string]any", func(t *testing.T) {
			c := New()
			v := c.GetStringMapVar(map[string]any{"a": 1, "b": 2}, t.Name())
			require.NotNil(t, v)
			require.Equal(t, map[string]any{"a": 1, "b": 2}, v)

			c.Set(t.Name(), map[string]any{"c": 3, "d": 4})
			require.Equal(t, map[string]any{"a": 1, "b": 2}, v, "variable is not reloadable")
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
				"detected misuse of config variable registered with different default values for \"int:TestNewReloadableAPI/reloadable/int\": "+
					"5 - 10",
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
				"detected misuse of config variable registered with different default values for \"int64:TestNewReloadableAPI/reloadable/int64\": "+
					"5 - 10",
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
				"detected misuse of config variable registered with different default values for \"bool:TestNewReloadableAPI/reloadable/bool\": "+
					"true - false",
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
				"detected misuse of config variable registered with different default values for \"float64:TestNewReloadableAPI/reloadable/float64\": "+
					"0.123 - 0.1234",
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
				"detected misuse of config variable registered with different default values for \"string:TestNewReloadableAPI/reloadable/string\": "+
					"foo - qux",
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
				"detected misuse of config variable registered with different default values for \"time.Duration:TestNewReloadableAPI/reloadable/duration\": "+
					"123ns - 2m3s",
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
				"detected misuse of config variable registered with different default values for \"[]string:TestNewReloadableAPI/reloadable/[]string\": "+
					"[a b] - [a b c]",
				func() {
					_ = c.GetReloadableStringSliceVar([]string{"a", "b", "c"}, t.Name())
				},
			)
		})
		t.Run("map[string]any", func(t *testing.T) {
			c := New()
			v := c.GetReloadableStringMapVar(map[string]any{"a": 1, "b": 2}, t.Name())
			require.Equal(t, map[string]any{"a": 1, "b": 2}, v.Load())

			c.Set(t.Name(), map[string]any{"c": 3, "d": 4})
			require.Equal(t, map[string]any{"c": 3, "d": 4}, v.Load())

			c.Set(t.Name(), map[string]any{"c": 3, "d": 4})
			require.Equal(t, map[string]any{"c": 3, "d": 4}, v.Load(), "value should not change")

			require.PanicsWithError(t,
				"detected misuse of config variable registered with different default values for \"map[string]interface {}:TestNewReloadableAPI/reloadable/map[string]any\": "+
					"map[a:1 b:2] - map[a:2 b:1]",
				func() {
					_ = c.GetReloadableStringMapVar(map[string]any{"a": 2, "b": 1}, t.Name())
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
		"detected misuse of config variable registered with different default values for \"int:bar,foo,qux\": "+
			"123 - 456",
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
		v := tc.GetStringVar("", "Key.Var1.Var2")
		require.Equal(t, expectedValue, v)
	})

	t.Run("detects camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR_VAR", expectedValue)
		tc := New()
		v := tc.GetStringVar("", "KeyVarVar")
		require.Equal(t, expectedValue, v)
	})

	t.Run("detects dots with camelcase", func(t *testing.T) {
		t.Setenv("RSERVER_KEY_VAR1_VAR_VAR", expectedValue)
		tc := New()
		v := tc.GetStringVar("", "Key.Var1VarVar")
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

func TestConfigChangesObserver(t *testing.T) {
	// Helper function to set up the config with a testKey and an observer
	setupConfig := func(t *testing.T, emptyConfig bool) (*Config, *testObserver, string) {
		// Create a temporary file for config
		f, err := os.CreateTemp(t.TempDir(), "*config.yaml")
		require.NoError(t, err)

		// Write initial config
		initialConfig := `
stringKey: initialValue
intKey: 1
int64Key: 1
float64Key: 1.0
boolKey: true
durationKey: 1s
stringSliceKey: 1
stringMapKey: '{"key": "value"}'
`
		if emptyConfig {
			initialConfig = ""
		}
		_, err = f.WriteString(initialConfig)
		require.NoError(t, err)
		err = f.Close()
		require.NoError(t, err)

		// Set environment variable to point to our config file
		t.Setenv("CONFIG_PATH", f.Name())

		observer := &testObserver{nonReloadableOperations: make(map[string][]KeyOperation), reloadableChanges: make(map[string][]reloadableConfigChange)}
		c := New()
		c.RegisterObserver(observer)
		return c, observer, f.Name()
	}

	t.Run("reloadable config changes", func(t *testing.T) {
		// reloadable config change detection is fine-grained, i.e. works on a per configuration var basis
		t.Run("configuration file changes and reloadable configs change", func(t *testing.T) {
			c, observer, filename := setupConfig(t, false)
			// Get value from config
			stringValue := c.GetReloadableStringVar("default", "stringKey")
			require.Equal(t, "initialValue", stringValue.Load())
			intValue := c.GetReloadableIntVar(2, 1, "intKey")
			require.Equal(t, 1, intValue.Load())
			int64Value := c.GetReloadableInt64Var(0, 1, "int64Key")
			require.EqualValues(t, 1, int64Value.Load())
			boolVar := c.GetReloadableBoolVar(false, "boolKey")
			require.True(t, boolVar.Load(), "it should return the key value")
			durVar := c.GetReloadableDurationVar(2, time.Second, "durationKey")
			require.Equal(t, 1*time.Second, durVar.Load(), "it should return the key value")
			stringSliceVar := c.GetReloadableStringSliceVar([]string{"default"}, "stringSliceKey")
			require.Equal(t, []string{"1"}, stringSliceVar.Load(), "it should return the key value")
			floatValue := c.GetReloadableFloat64Var(2, "float64Key")
			require.Equal(t, 1.0, floatValue.Load())
			stringMapKey := c.GetReloadableStringMapVar(map[string]any{"default": "value"}, "stringMapKey")
			require.Equal(t, map[string]any{"key": "value"}, stringMapKey.Load(), "it should return the key value")

			// change the config file
			err := os.WriteFile(filename, []byte(`
stringKey: otherValue
intKey: 2
int64Key: 2
float64Key: 2.0
boolKey: false
durationKey: 2s
stringSliceKey: 2
stringMapKey: '{"key": "value2"}'
`), 0o644)
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
				return stringValue.Load() == "otherValue" &&
					intValue.Load() == 2 &&
					int64Value.Load() == 2 &&
					boolVar.Load() == false &&
					durVar.Load() == 2*time.Second &&
					stringSliceVar.Load()[0] == "2" &&
					floatValue.Load() == 2.0 &&
					stringMapKey.Load()["key"] == "value2"
			}, 2*time.Second, 1*time.Millisecond)

			stringChanes := observer.getReloadableChanges("stringKey")
			require.Len(t, stringChanes, 1)
			require.Equal(t, "initialValue", stringChanes[0].oldValue)
			require.Equal(t, "otherValue", stringChanes[0].newValue)
			intChanges := observer.getReloadableChanges("intKey")
			require.Len(t, intChanges, 1)
			require.Equal(t, 1, intChanges[0].oldValue)
			require.Equal(t, 2, intChanges[0].newValue)
			int64Changes := observer.getReloadableChanges("int64Key")
			require.Len(t, int64Changes, 1)
			require.EqualValues(t, 1, int64Changes[0].oldValue)
			require.EqualValues(t, 2, int64Changes[0].newValue)
			boolChanges := observer.getReloadableChanges("boolKey")
			require.Len(t, boolChanges, 1)
			require.Equal(t, true, boolChanges[0].oldValue)
			require.Equal(t, false, boolChanges[0].newValue)
			durationChanges := observer.getReloadableChanges("durationKey")
			require.Len(t, durationChanges, 1)
			require.Equal(t, 1*time.Second, durationChanges[0].oldValue)
			require.Equal(t, 2*time.Second, durationChanges[0].newValue)
			stringSliceChanges := observer.getReloadableChanges("stringSliceKey")
			require.Len(t, stringSliceChanges, 1)
			require.Equal(t, []string{"1"}, stringSliceChanges[0].oldValue)
			require.Equal(t, []string{"2"}, stringSliceChanges[0].newValue)
			floatChanges := observer.getReloadableChanges("float64Key")
			require.Len(t, floatChanges, 1)
			require.Equal(t, 1.0, floatChanges[0].oldValue)
			require.Equal(t, 2.0, floatChanges[0].newValue)
			stringMapChanges := observer.getReloadableChanges("stringMapKey")
			require.Len(t, stringMapChanges, 1)
			require.Equal(t, map[string]any{"key": "value"}, stringMapChanges[0].oldValue)
			require.Equal(t, map[string]any{"key": "value2"}, stringMapChanges[0].newValue)
		})

		t.Run("configuration file changes but reloadable configs remain the same", func(t *testing.T) {
			c, observer, filename := setupConfig(t, false)
			// using initial values as default, so that after we delete the keys from the config file, it will not change
			stringValue := c.GetReloadableStringVar("initialValue", "stringKey")
			require.Equal(t, "initialValue", stringValue.Load())
			intValue := c.GetReloadableIntVar(1, 1, "intKey")
			require.Equal(t, 1, intValue.Load())
			int64Value := c.GetReloadableInt64Var(1, 1, "int64Key")
			require.EqualValues(t, 1, int64Value.Load())
			boolVar := c.GetReloadableBoolVar(true, "boolKey")
			require.True(t, boolVar.Load(), "it should return the key value")
			durVar := c.GetReloadableDurationVar(1, time.Second, "durationKey")
			require.Equal(t, 1*time.Second, durVar.Load(), "it should return the key value")
			stringSliceVar := c.GetReloadableStringSliceVar([]string{"1"}, "stringSliceKey")
			require.Equal(t, []string{"1"}, stringSliceVar.Load(), "it should return the key value")
			floatValue := c.GetReloadableFloat64Var(1.0, "float64Key")
			require.Equal(t, 1.0, floatValue.Load())
			stringMapKey := c.GetReloadableStringMapVar(map[string]any{"key": "value"}, "stringMapKey")
			require.Equal(t, map[string]any{"key": "value"}, stringMapKey.Load(), "it should return the key value")

			// add one more to know when the config file changed has been detected
			newKey := c.GetReloadableStringVar("default", "newKey")
			require.Equal(t, "default", newKey.Load(), "it should return the key value")

			// change the config file
			err := os.WriteFile(filename, []byte(`
newKey: value
`), 0o644)
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
				return newKey.Load() == "value"
			}, 2*time.Second, 1*time.Millisecond)

			// Verify that no events for other keys have been fired
			stringChanes := observer.getReloadableChanges("stringKey")
			require.Len(t, stringChanes, 0)
			intChanges := observer.getReloadableChanges("intKey")
			require.Len(t, intChanges, 0)
			int64Changes := observer.getReloadableChanges("int64Key")
			require.Len(t, int64Changes, 0)
			boolChanges := observer.getReloadableChanges("boolKey")
			require.Len(t, boolChanges, 0)
			durationChanges := observer.getReloadableChanges("durationKey")
			require.Len(t, durationChanges, 0)
			stringSliceChanges := observer.getReloadableChanges("stringSliceKey")
			require.Len(t, stringSliceChanges, 0)
			floatChanges := observer.getReloadableChanges("float64Key")
			require.Len(t, floatChanges, 0)
			stringMapChanges := observer.getReloadableChanges("stringMapKey")
			require.Len(t, stringMapChanges, 0)
		})
	})
	t.Run("non-reloadable config changes", func(t *testing.T) {
		// non-reloadable config change detection is coarse-grained, i.e. just detects changes in config file of keys which have been used by the application by non-reloadable configuration requests

		t.Run("configuration file changes and modification events are sent for non-reloadable keys", func(t *testing.T) {
			c, observer, filename := setupConfig(t, false)

			// Get non-reloadable values from config
			stringValue := c.GetString("stringKey", "default")
			require.Equal(t, "initialValue", stringValue)
			intValue := c.GetInt("intKey", 0)
			require.Equal(t, 1, intValue)
			int64Value := c.GetInt64("int64Key", 0)
			require.EqualValues(t, 1, int64Value)
			float64Value := c.GetFloat64("float64Key", 0.0)
			require.Equal(t, 1.0, float64Value)
			boolValue := c.GetBool("boolKey", false)
			require.True(t, boolValue)
			durationValue := c.GetDuration("durationKey", 0, time.Second)
			require.Equal(t, 1*time.Second, durationValue)
			stringSliceValue := c.GetStringSlice("stringSliceKey", []string{"default"})
			require.Equal(t, []string{"1"}, stringSliceValue)
			stringMapValue := c.GetStringMap("stringMapKey", map[string]any{"default": "value"})
			require.Equal(t, map[string]any{"key": "value"}, stringMapValue)

			// Modify the config file
			err := os.WriteFile(filename, []byte(`
stringKey: otherValue
intKey: 2
int64Key: 2
float64Key: 2.0
boolKey: false
durationKey: 2s
stringSliceKey: 2
stringMapKey: '{"key": "value2"}'
`), 0o644)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
				// Verify that the values have changed
				return c.GetString("stringKey", "default") == "otherValue" &&
					c.GetInt("intKey", 0) == 2 &&
					c.GetInt64("int64Key", 0) == 2 &&
					c.GetFloat64("float64Key", 0.0) == 2.0 &&
					c.GetBool("boolKey", false) == false &&
					c.GetDuration("durationKey", 0, time.Second) == 2*time.Second &&
					reflect.DeepEqual(c.GetStringSlice("stringSliceKey", []string{"default"}), []string{"2"}) &&
					reflect.DeepEqual(c.GetStringMap("stringMapKey", map[string]any{"default": "value"}), map[string]any{"key": "value2"})
			}, 2*time.Second, 1*time.Millisecond)

			// Te observer should be notified of config file changes
			// for keys that have been previously accessed
			stringOperations := observer.getNonReloadableOperations("stringKey")
			require.Len(t, stringOperations, 1)
			require.Equal(t, KeyOperationModified, stringOperations[0])

			intOperations := observer.getNonReloadableOperations("intKey")
			require.Len(t, intOperations, 1)
			require.Equal(t, KeyOperationModified, intOperations[0])

			int64Operations := observer.getNonReloadableOperations("int64Key")
			require.Len(t, int64Operations, 1)
			require.Equal(t, KeyOperationModified, int64Operations[0])

			floatOperations := observer.getNonReloadableOperations("float64Key")
			require.Len(t, floatOperations, 1)
			require.Equal(t, KeyOperationModified, floatOperations[0])

			boolOperations := observer.getNonReloadableOperations("boolKey")
			require.Len(t, boolOperations, 1)
			require.Equal(t, KeyOperationModified, boolOperations[0])

			durationOperations := observer.getNonReloadableOperations("durationKey")
			require.Len(t, durationOperations, 1)
			require.Equal(t, KeyOperationModified, durationOperations[0])

			sliceOperations := observer.getNonReloadableOperations("stringSliceKey")
			require.Len(t, sliceOperations, 1)
			require.Equal(t, KeyOperationModified, sliceOperations[0])

			mapOperations := observer.getNonReloadableOperations("stringMapKey")
			require.Len(t, mapOperations, 1)
			require.Equal(t, KeyOperationModified, mapOperations[0])
		})

		t.Run("configuration file gets cleared and removal events are sent for non-reloadable keys", func(t *testing.T) {
			c, observer, filename := setupConfig(t, false)

			// Get non-reloadable values from config
			stringValue := c.GetString("stringKey", "default")
			require.Equal(t, "initialValue", stringValue)
			intValue := c.GetInt("intKey", 0)
			require.Equal(t, 1, intValue)
			int64Value := c.GetInt64("int64Key", 0)
			require.EqualValues(t, 1, int64Value)
			float64Value := c.GetFloat64("float64Key", 0.0)
			require.Equal(t, 1.0, float64Value)
			boolValue := c.GetBool("boolKey", false)
			require.True(t, boolValue)
			durationValue := c.GetDuration("durationKey", 0, time.Second)
			require.Equal(t, 1*time.Second, durationValue)
			stringSliceValue := c.GetStringSlice("stringSliceKey", []string{"default"})
			require.Equal(t, []string{"1"}, stringSliceValue)
			stringMapValue := c.GetStringMap("stringMapKey", map[string]any{"default": "value"})
			require.Equal(t, map[string]any{"key": "value"}, stringMapValue)

			// Modify the config file
			err := os.WriteFile(filename, []byte(`
`), 0o644)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
				// Verify that the values have changed
				return c.GetString("stringKey", "default") == "default" &&
					c.GetInt("intKey", 0) == 0 &&
					c.GetInt64("int64Key", 0) == 0 &&
					c.GetFloat64("float64Key", 0.0) == 0.0 &&
					c.GetBool("boolKey", false) == false &&
					c.GetDuration("durationKey", 0, time.Second) == 0*time.Second &&
					reflect.DeepEqual(c.GetStringSlice("stringSliceKey", []string{"default"}), []string{"default"}) &&
					reflect.DeepEqual(c.GetStringMap("stringMapKey", map[string]any{"default": "value"}), map[string]any{"default": "value"})
			}, 2*time.Second, 1*time.Millisecond)

			// Te observer should be notified of config file changes
			// for keys that have been previously accessed
			stringOperations := observer.getNonReloadableOperations("stringKey")
			require.Len(t, stringOperations, 1)
			require.Equal(t, KeyOperationRemoved, stringOperations[0])

			intOperations := observer.getNonReloadableOperations("intKey")
			require.Len(t, intOperations, 1)
			require.Equal(t, KeyOperationRemoved, intOperations[0])

			int64Operations := observer.getNonReloadableOperations("int64Key")
			require.Len(t, int64Operations, 1)
			require.Equal(t, KeyOperationRemoved, int64Operations[0])

			floatOperations := observer.getNonReloadableOperations("float64Key")
			require.Len(t, floatOperations, 1)
			require.Equal(t, KeyOperationRemoved, floatOperations[0])

			boolOperations := observer.getNonReloadableOperations("boolKey")
			require.Len(t, boolOperations, 1)
			require.Equal(t, KeyOperationRemoved, boolOperations[0])

			durationOperations := observer.getNonReloadableOperations("durationKey")
			require.Len(t, durationOperations, 1)
			require.Equal(t, KeyOperationRemoved, durationOperations[0])

			sliceOperations := observer.getNonReloadableOperations("stringSliceKey")
			require.Len(t, sliceOperations, 1)
			require.Equal(t, KeyOperationRemoved, sliceOperations[0])

			mapOperations := observer.getNonReloadableOperations("stringMapKey")
			require.Len(t, mapOperations, 1)
			require.Equal(t, KeyOperationRemoved, mapOperations[0])
		})

		t.Run("configuration file keys get introduced and addition events are sent for non-reloadable keys", func(t *testing.T) {
			c, observer, filename := setupConfig(t, true)

			// Get non-reloadable values from config
			stringValue := c.GetString("stringKey", "default")
			require.Equal(t, "default", stringValue)
			intValue := c.GetInt("intKey", 0)
			require.Equal(t, 0, intValue)
			int64Value := c.GetInt64("int64Key", 0)
			require.EqualValues(t, 0, int64Value)
			float64Value := c.GetFloat64("float64Key", 0.0)
			require.Equal(t, 0.0, float64Value)
			boolValue := c.GetBool("boolKey", false)
			require.False(t, boolValue)
			durationValue := c.GetDuration("durationKey", 0, time.Second)
			require.Equal(t, 0*time.Second, durationValue)
			stringSliceValue := c.GetStringSlice("stringSliceKey", []string{"default"})
			require.Equal(t, []string{"default"}, stringSliceValue)
			stringMapValue := c.GetStringMap("stringMapKey", map[string]any{"default": "value"})
			require.Equal(t, map[string]any{"default": "value"}, stringMapValue)

			// Modify the config file
			err := os.WriteFile(filename, []byte(`
stringKey: initialValue
intKey: 1
int64Key: 1
float64Key: 1.0
boolKey: true
durationKey: 1s
stringSliceKey: 1
stringMapKey: '{"key": "value"}'
`), 0o644)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
				// Verify that the values have changed
				return c.GetString("stringKey", "default") == "initialValue" &&
					c.GetInt("intKey", 0) == 1 &&
					c.GetInt64("int64Key", 0) == 1 &&
					c.GetFloat64("float64Key", 0.0) == 1.0 &&
					c.GetBool("boolKey", false) == true &&
					c.GetDuration("durationKey", 0, time.Second) == 1*time.Second &&
					reflect.DeepEqual(c.GetStringSlice("stringSliceKey", []string{"default"}), []string{"1"}) &&
					reflect.DeepEqual(c.GetStringMap("stringMapKey", map[string]any{"default": "value"}), map[string]any{"key": "value"})
			}, 2*time.Second, 1*time.Millisecond)

			// Te observer should be notified of config file changes
			// for keys that have been previously accessed
			stringOperations := observer.getNonReloadableOperations("stringKey")
			require.Len(t, stringOperations, 1)
			require.Equal(t, KeyOperationAdded, stringOperations[0])

			intOperations := observer.getNonReloadableOperations("intKey")
			require.Len(t, intOperations, 1)
			require.Equal(t, KeyOperationAdded, intOperations[0])

			int64Operations := observer.getNonReloadableOperations("int64Key")
			require.Len(t, int64Operations, 1)
			require.Equal(t, KeyOperationAdded, int64Operations[0])

			floatOperations := observer.getNonReloadableOperations("float64Key")
			require.Len(t, floatOperations, 1)
			require.Equal(t, KeyOperationAdded, floatOperations[0])

			boolOperations := observer.getNonReloadableOperations("boolKey")
			require.Len(t, boolOperations, 1)
			require.Equal(t, KeyOperationAdded, boolOperations[0])

			durationOperations := observer.getNonReloadableOperations("durationKey")
			require.Len(t, durationOperations, 1)
			require.Equal(t, KeyOperationAdded, durationOperations[0])

			sliceOperations := observer.getNonReloadableOperations("stringSliceKey")
			require.Len(t, sliceOperations, 1)
			require.Equal(t, KeyOperationAdded, sliceOperations[0])

			mapOperations := observer.getNonReloadableOperations("stringMapKey")
			require.Len(t, mapOperations, 1)
			require.Equal(t, KeyOperationAdded, mapOperations[0])
		})

		t.Run("setting a non-reloadable value", func(t *testing.T) {
			c, observer, _ := setupConfig(t, false)

			t.Run("key exists", func(t *testing.T) {
				// Get non-reloadable values from config
				stringValue := c.GetString("stringKey", "default")
				require.Equal(t, "initialValue", stringValue)

				// Set a new value for the key
				c.Set("stringKey", "newValue")

				// Verify that the value has changed
				require.Equal(t, "newValue", c.GetString("stringKey", "default"))

				// Verify that the observer was notified of the modification
				stringOperations := observer.getNonReloadableOperations("stringKey")
				require.Len(t, stringOperations, 1)
				require.Equal(t, KeyOperationModified, stringOperations[0])
			})

			t.Run("key does not exist", func(t *testing.T) {
				// Try to get a non-reloadable value for a key that does not exist
				stringValue := c.GetString("nonExistentKey", "default")
				require.Equal(t, "default", stringValue)

				// Set a new value for the key
				c.Set("nonExistentKey", "newValue")

				// Verify that the value has changed
				require.Equal(t, "newValue", c.GetString("nonExistentKey", "default"))

				// Verify that the observer was notified of the addition
				stringOperations := observer.getNonReloadableOperations("nonExistentKey")
				require.Len(t, stringOperations, 1)
				require.Equal(t, KeyOperationAdded, stringOperations[0])
			})
		})
	})

	t.Run("register and unregister observer", func(t *testing.T) {
		c, observer, filename := setupConfig(t, false)

		// First, get a value to ensure the key is tracked
		stringValue := c.GetString("stringKey", "default")
		require.Equal(t, "initialValue", stringValue)

		// Modify the config file to trigger a notification
		err := os.WriteFile(filename, []byte(`
stringKey: changedValue
`), 0o644)
		require.NoError(t, err)

		// Wait for the observer to be notified
		require.Eventually(t, func() bool {
			time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
			ops := observer.getNonReloadableOperations("stringKey")
			return len(ops) == 1 && ops[0] == KeyOperationModified
		}, 2*time.Second, 1*time.Millisecond)

		// Unregister the observer
		c.UnregisterObserver(observer)

		// Clear the observer's record
		observer.mu.Lock()
		observer.nonReloadableOperations = make(map[string][]KeyOperation)
		observer.mu.Unlock()

		// Modify the config file again
		err = os.WriteFile(filename, []byte(`
stringKey: anotherValue
`), 0o644)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			time.Sleep(100 * time.Millisecond) // give some time for the observer to detect changes (viper is not thread-safe)
			return c.GetString("stringKey", "default") == "anotherValue"
		}, 2*time.Second, 1*time.Millisecond)

		// Verify the observer was not notified after unregistering
		ops := observer.getNonReloadableOperations("stringKey")
		require.Empty(t, ops, "Observer should not be notified after unregistering")
	})

	t.Run("observer functions", func(t *testing.T) {
		t.Run("OnNonReloadableConfigChange", func(t *testing.T) {
			c, _, filename := setupConfig(t, false)
			var called atomic.Bool
			c.OnNonReloadableConfigChange(func(key string, op KeyOperation) {
				called.Store(true)
			})
			stringValue := c.GetStringVar("", "stringKey") // Ensure the key is tracked
			require.Equal(t, "initialValue", stringValue)
			// Modify the config file again
			err := os.WriteFile(filename, []byte(`
stringKey: anotherValue
`), 0o644)
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				return called.Load()
			}, 2*time.Second, 1*time.Millisecond)
		})

		t.Run("OnReloadableConfigChange", func(t *testing.T) {
			c, _, filename := setupConfig(t, false)
			var called atomic.Bool
			c.OnReloadableConfigChange(func(key string, oldValue, newValue any) {
				called.Store(true)
			})
			stringValue := c.GetReloadableStringVar("", "stringKey") // Ensure the key is tracked
			require.Equal(t, "initialValue", stringValue.Load())
			// Modify the config file again
			err := os.WriteFile(filename, []byte(`
stringKey: anotherValue
`), 0o644)
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				return called.Load()
			}, 2*time.Second, 1*time.Millisecond)
		})
	})
}

// Helper test observer implementation for testing
type testObserver struct {
	mu                      sync.Mutex
	nonReloadableOperations map[string][]KeyOperation
	reloadableChanges       map[string][]reloadableConfigChange
}

type reloadableConfigChange struct {
	oldValue any
	newValue any
}

func (o *testObserver) OnNonReloadableConfigChange(key string, op KeyOperation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.nonReloadableOperations[key] = append(o.nonReloadableOperations[key], op)
}

func (o *testObserver) OnReloadableConfigChange(key string, oldValue, newValue any) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.reloadableChanges[key] = append(o.reloadableChanges[key], reloadableConfigChange{oldValue: oldValue, newValue: newValue})
}

func (o *testObserver) getNonReloadableOperations(key string) []KeyOperation {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.nonReloadableOperations[key]
}

func (o *testObserver) getReloadableChanges(key string) []reloadableConfigChange {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.reloadableChanges[key]
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
