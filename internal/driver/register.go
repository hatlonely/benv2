package driver

func init() {
	RegisterDriver("Shell", NewWrapDriverWithMethodName(NewShellDriverWithOptions, "Do"))
}
