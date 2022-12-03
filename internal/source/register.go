package source

func init() {
	RegisterSource("Dict", NewDictSourceWithOptions)
	RegisterSource("File", NewFileSourceWithOptions)
}
