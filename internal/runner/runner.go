// Package runner is responsible for running the command and rendering the output
package runner

import (
	"encoding/json"
	"io"
	"os"
	"strconv"

	"github.com/psampaz/go-mod-outdated/internal/mod"

	"github.com/olekukonko/tablewriter"
)

// OsExit is use here in order to simplify testing
var OsExit = os.Exit

type Style string

const (
	StyleDefault  Style = "default"
	StyleMarkdown Style = "markdown"
)

// Run converts the the json output of go list -u -m -json all to table format
func Run(in io.Reader, out io.Writer, update, direct, exitWithNonZero bool, style Style) error {
	var modules []mod.Module

	dec := json.NewDecoder(in)

	for {
		var v mod.Module
		err := dec.Decode(&v)

		if err != nil {
			if err == io.EOF {
				filteredModules := mod.FilterModules(modules, update, direct)
				if len(filteredModules) > 0 {
					renderTable(out, filteredModules, style)
				}

				if hasOutdated(filteredModules) && exitWithNonZero {
					OsExit(1)
				}

				return nil
			}

			return err
		}

		modules = append(modules, v)
	}
}

func hasOutdated(filteredModules []mod.Module) bool {
	for m := range filteredModules {
		if filteredModules[m].HasUpdate() {
			return true
		}
	}

	return false
}

func renderTable(writer io.Writer, modules []mod.Module, style Style) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Module", "Version", "New Version", "Direct", "Valid Timestamps"})

	// Render table as markdown
	if style == StyleMarkdown {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	}

	for k := range modules {
		table.Append([]string{
			modules[k].Path,
			modules[k].CurrentVersion(),
			modules[k].NewVersion(),
			strconv.FormatBool(!modules[k].Indirect),
			strconv.FormatBool(!modules[k].InvalidTimestamp()),
		})
	}

	table.Render()
}
