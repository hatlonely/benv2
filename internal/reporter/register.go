package reporter

func init() {
	RegisterReporter("Json", &JsonReporter{})
}
