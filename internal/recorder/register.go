package recorder

func init() {
	RegisterRecorder("File", NewFileRecorderWithOptions)
}
