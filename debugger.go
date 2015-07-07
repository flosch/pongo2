package pongo2

type Debugger interface {
	Debug() bool
	SetDebug(bool)
}
