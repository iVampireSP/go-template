package must

func Do(err error) {
	if err != nil {
		panic(err)
	}
}

func Get[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func Must[T any](value T, err error) T {
	return Get(value, err)
}
