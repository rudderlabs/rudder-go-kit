package maputil

// MapLookup traverses a nested map[string]interface{} using a sequence of keys.
//
// Given a root map and a variadic list of keys, MapLookup attempts to follow the path
// described by the keys through nested maps. If the full path exists, it returns the
// value found at the end of the path. If any key is missing or a value along the path
// is not a map[string]interface{}, it returns nil.
//
// Example usage:
//
//	m := map[string]interface{}{
//	    "a": map[string]interface{}{
//	        "b": map[string]interface{}{
//	            "c": 42,
//	        },
//	    },
//	}
//	v := MapLookup(m, "a", "b", "c") // v == 42
//
// Parameters:
//   - mapToLookup: the root map to search
//   - keys: a variadic list of keys representing the path to traverse
//
// Returns:
//   - The value at the end of the key path if found, or nil otherwise.
func MapLookup(mapToLookup map[string]any, keys ...string) any {
	if len(keys) == 0 {
		return nil
	}
	if val, ok := mapToLookup[keys[0]]; ok {
		if len(keys) == 1 {
			return val
		}
		nextMap, ok := val.(map[string]any)
		if !ok {
			return nil
		}
		return MapLookup(nextMap, keys[1:]...)
	}
	return nil
}
