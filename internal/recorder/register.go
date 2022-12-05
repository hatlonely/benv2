package recorder

func init() {
	RegisterRecorder("File", NewFileRecorderWithOptions)

	RegisterAnalyst("File", NewFileAnalystWithOptions)
}
