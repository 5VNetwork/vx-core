//go:build android

package x_android

// StringList is an interface that represents a list of strings
type StringList interface {
	Get(index int) string
	Len() int
}

// stringList is an implementation of StringList using a slice
type stringList struct {
	strings []string
}

// Get returns the string at the specified index
func (l *stringList) Get(index int) string {
	if index < 0 || index >= len(l.strings) {
		return ""
	}
	return l.strings[index]
}

// Len returns the number of elements in the list
func (l *stringList) Len() int {
	return len(l.strings)
}
