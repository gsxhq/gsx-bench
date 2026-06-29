// Package data holds the shared input models for the render benchmarks so the
// gsx and templ renderers consume identical values.
package data

import "fmt"

// Person is the model for the Document scenario (a faithful port of templ's own
// benchmark template).
type Person struct {
	Name  string
	Email string
}

// Row is the model for the list/table/composition scenarios — enough fields to
// exercise text, attribute, and conditional paths per element.
type Row struct {
	ID     int
	Name   string
	Email  string
	Role   string
	Active bool
}

// Href is the canonical link for a row — a shared helper so the gsx and templ
// page templates compute an identical dynamic URL attribute.
func (r Row) Href() string { return fmt.Sprintf("/users/%d", r.ID) }

// Rows builds n deterministic rows (no time/random, so results are stable).
func Rows(n int) []Row {
	rows := make([]Row, n)
	for i := range rows {
		rows[i] = Row{
			ID:     i + 1,
			Name:   fmt.Sprintf("User %d", i+1),
			Email:  fmt.Sprintf("user%d@example.com", i+1),
			Role:   []string{"admin", "editor", "viewer"}[i%3],
			Active: i%2 == 0,
		}
	}
	return rows
}
