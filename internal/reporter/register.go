package reporter

func init() {
	RegisterReporter("Json", &JsonReporter{})
	RegisterReporter("Text", NewTextReporterWithOptions)
}
