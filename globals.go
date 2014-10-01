package pongo2

var globals Context

func init() {
	globals = make(Context)
}

func RegisterGlobal(name string, value interface{}) {
	if name == "pongo2" {
		panic("Global variable with name pongo2 is not allowed.")
	}
	globals[name] = value
}
