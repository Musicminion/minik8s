package main

import (
	"os"

	"github.com/jedib0t/go-pretty/table"
)

func main() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"First Name", "Last Name", "Age"})
	t.AppendRows([]table.Row{
		{"John", "Doe", "30"},
		{"Jane", "Smith", "25"},
		{"Bob", "Johnson", "45"},
	})
	t.Render()
}
