// Package data holds the shared input model for the render benchmarks so the
// gsx and templ renderers consume an identical value.
package data

type Person struct {
	Name  string
	Email string
}
