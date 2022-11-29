package tools

func MapSlice[T, U any](in []T, fn func(T) U) []U {
	r := make([]U, 0, len(in))
	for _, elm := range in {
		r = append(r, fn(elm))
	}
	return r
}

func MapMap[K, T comparable, V, U any](in map[K]V, fn func(K, V) (T, U)) map[T]U {
	r := make(map[T]U, len(in))

	for k, v := range in {
		t, u := fn(k, v)
		r[t] = u
	}

	return r
}
