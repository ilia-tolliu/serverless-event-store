package internal

func IgnoreError(f func() error) {
	_ = f()
}
