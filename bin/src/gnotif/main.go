package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//go:embed overlay.js
var overlayJS []byte

const usage = `notify — show an overlay notification with text-to-speech

Usage:
  notify [options] <message>

Options:
  --color <color>     Background color: red, blue, yellow, green, purple, orange (default: blue)
  --duration <secs>   How long to show the notification in seconds (default: 4)
  --no-say            Skip text-to-speech
  --help              Show this help

Examples:
  notify "Build complete"
  notify --color green "Tests passed"
  notify --color red --duration 6 "Deployment failed"
  notify --no-say "Silent notification"
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	color := "blue"
	duration := 4.0
	say := true
	var message string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			fmt.Print(usage)
			os.Exit(0)
		case "--color", "-c":
			if i+1 >= len(args) {
				fatal("--color requires a value")
			}
			i++
			color = args[i]
		case "--duration", "-d":
			if i+1 >= len(args) {
				fatal("--duration requires a value")
			}
			i++
			d, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				fatal("invalid duration: %s", args[i])
			}
			duration = d
		case "--no-say":
			say = false
		default:
			if strings.HasPrefix(args[i], "--") {
				fatal("unknown flag: %s\n\n%s", args[i], usage)
			}
			// Remaining args joined as message
			message = strings.Join(args[i:], " ")
			i = len(args) // break
		}
	}

	if message == "" {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	validColors := map[string]bool{
		"red": true, "blue": true, "yellow": true,
		"green": true, "purple": true, "orange": true,
	}
	if !validColors[color] {
		fatal("invalid color %q — valid: red, blue, yellow, green, purple, orange", color)
	}

	// Write overlay.js to a temp file
	tmpDir, err := os.MkdirTemp("", "notify-*")
	if err != nil {
		fatal("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	jsPath := filepath.Join(tmpDir, "overlay.js")
	if err := os.WriteFile(jsPath, overlayJS, 0644); err != nil {
		fatal("failed to write overlay script: %v", err)
	}

	var wg sync.WaitGroup

	// Run overlay
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("osascript", "-l", "JavaScript", jsPath,
			message, color, "0", fmt.Sprintf("%.1f", duration))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "overlay error: %v\n", err)
		}
	}()

	// Run say
	if say {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command("say", message)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "say error: %v\n", err)
			}
		}()
	}

	wg.Wait()
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
