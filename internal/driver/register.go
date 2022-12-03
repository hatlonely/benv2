package driver

func init() {
	RegisterDriver("Shell", NewWrapDriverWithMethodName(NewShellDriverWithOptions, "Do"))
	RegisterDriver("Http", NewWrapDriverWithMethodName(NewHttpDriverWithOptions, "Do"))
}
