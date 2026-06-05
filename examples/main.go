// examples — interactive menu launcher for all tlog examples.
//
// Run from the module root:
//
//	go run ./examples/
//
// Select an example by number, watch its output, then return here.
// Press q (or Ctrl-C) to quit.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// entry describes one example program.
type entry struct {
	dir  string // path relative to module root
	name string // short display name
	desc string // one-line description
}

var examples = []entry{
	{"examples/basic", "basic", "Minimal logger: levels and fields"},
	{"examples/structured", "structured", "All typed field constructors"},
	{"examples/multi-output", "multi-output", "Console + file fan-out"},
	{"examples/custom-theme", "custom-theme", "Built-in themes and style overrides"},
	{"examples/json-mode", "json-mode", "Production JSON logging pattern"},
	{"examples/context", "context", "Context propagation through call chains"},
	{"examples/global", "global", "Package-level global logger"},
	{"examples/webserver", "webserver", "HTTP access log + structured middleware"},
}

// ── styles ───────────────────────────────────────────────────────────────────

var (
	styleBrand = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	styleSubtitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	styleHeaderCell = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true)

	styleDivider = lipgloss.NewStyle().
			Foreground(lipgloss.Color("237"))

	styleIndex = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")).
			Width(3)

	styleName = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Width(16)

	styleDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	styleQuit = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	stylePrompt = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	styleRunHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	styleDone = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	styleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	styleWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// ── rendering ────────────────────────────────────────────────────────────────

func clearScreen() { fmt.Print("\033[H\033[2J") }

func printMenu() {
	clearScreen()
	fmt.Println()
	fmt.Println("  " + styleBrand.Render("tlog") + styleSubtitle.Render(" — example programs"))
	fmt.Println("  " + styleSubtitle.Render("Select an example to run. Press q to quit."))
	fmt.Println()

	header := styleHeaderCell.Render(fmt.Sprintf("  %-3s  %-16s  %s", "#", "example", "description"))
	fmt.Println(header)
	fmt.Println("  " + styleDivider.Render(strings.Repeat("─", 54)))

	for i, e := range examples {
		idx := styleIndex.Render(fmt.Sprintf("%d", i+1))
		name := styleName.Render(e.name)
		desc := styleDesc.Render(e.desc)
		fmt.Printf("  %s  %s  %s\n", idx, name, desc)
	}

	fmt.Println()
	fmt.Println("  " + styleQuit.Render("q") + styleSubtitle.Render("    quit"))
	fmt.Println()
}

// ── execution ─────────────────────────────────────────────────────────────────

func run(e entry) {
	fmt.Println()
	fmt.Println(styleRunHeader.Render("── running: " + e.name + " ──"))
	fmt.Println()

	// #nosec G204 -- we're not passing user input to the shell, just a static command
	cmd := exec.CommandContext(
		context.Background(),
		"go", "run", "./"+filepath.Clean(e.dir)+"/",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintln(os.Stderr, styleError.Render("error: "+err.Error()))
		}
	}

	fmt.Println()
	fmt.Print(styleDone.Render("── done. Press Enter to return to the menu..."))
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// ── main loop ─────────────────────────────────────────────────────────────────

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		printMenu()
		fmt.Print("  " + stylePrompt.Render(">") + " ")

		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		choice := strings.TrimSpace(line)

		switch choice {
		case "q", "Q", "quit", "exit":
			clearScreen()
			fmt.Println("  " + styleSubtitle.Render("Bye!"))
			fmt.Println()
			return
		case "":
			// redraw
		default:
			idx := 0
			if n, _ := fmt.Sscanf(choice, "%d", &idx); n == 1 && idx >= 1 && idx <= len(examples) {
				run(examples[idx-1])
			} else {
				fmt.Println()
				fmt.Println("  " + styleWarn.Render(fmt.Sprintf("unknown choice %q — enter a number 1-%d or q", choice, len(examples))))
				fmt.Print("  Press Enter to continue...")
				_, _ = reader.ReadString('\n')
			}
		}
	}
}
