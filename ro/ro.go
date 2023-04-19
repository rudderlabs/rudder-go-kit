package ro

func Memoize[R any](f func() R) func() R {
	var result R
	var called bool
	return func() R {
		if called {
			return result
		}
		result = f()
		called = true
		return result
	}
}
