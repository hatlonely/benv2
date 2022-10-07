package driver

type Driver interface {
	Do(v interface{}) (interface{}, error)
}

