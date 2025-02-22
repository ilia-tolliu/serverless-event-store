package eserror

func Ignore(f func() error) {
	_ = f()
}
