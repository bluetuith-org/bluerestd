package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

// infoSpinner stops the spinner with an info message.
func infoSpinner(s ...any) *pterm.SpinnerPrinter {
	sp := pterm.DefaultSpinner.WithDelay(1 * time.Second)
	spinner, _ := sp.Start(s...)

	return spinner
}

// errorSpinner stops the spinner with an error message.
func errorSpinner(spinner *pterm.SpinnerPrinter, err error) error {
	spinner.Fail(err)

	return err
}

// clearSpinner clears the spinner's text.
func clearSpinner(spinner *pterm.SpinnerPrinter) {
	updateSpinner(spinner, "%s", strings.Repeat(" ", len(spinner.Text)))
}

// updateSpinner updates the spinner's text.
func updateSpinner(spinner *pterm.SpinnerPrinter, fmt string, s ...any) {
	style := pterm.FgDefault.Sprintf(fmt, s...)
	spinner.UpdateText(style)
}

// printWarn prints a warn message.
func printWarn(fmt string, s ...any) {
	style := pterm.Bold.ToStyle().Add(*pterm.FgYellow.ToStyle())
	pterm.Warning.WithMessageStyle(&style).Printfln(fmt, s...)
}

// printInfo prints a info message.
func printInfo(fmt string, s ...any) {
	pterm.Info.WithMessageStyle(pterm.NewStyle(pterm.FgBlue, pterm.Bold)).Printfln(fmt, s...)
}

// printNote prints a note message.
func printNote(fmt string, s ...any) {
	prefix := pterm.Info.Prefix

	pterm.Info.Prefix = pterm.Prefix{Text: "NOTE", Style: pterm.NewStyle(pterm.BgGray, pterm.Bold)}
	pterm.Info.WithMessageStyle(pterm.NewStyle(pterm.FgGray, pterm.Bold)).Printfln(fmt, s...)
	pterm.Info.Prefix = prefix
}

// newline prints a new line.
func newline() {
	fmt.Println()
}
