package commands

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPrimeNamesOnlyExistingCommands(t *testing.T) {
	oldConfigPath := configPath
	configPath = "set" // skips the upward search for .beans.yml
	defer func() { configPath = oldConfigPath }()

	out := string(captureStdout(t, func() {
		if err := primeCmd.RunE(primeCmd, nil); err != nil {
			t.Fatal(err)
		}
	}))

	root := &cobra.Command{Use: "beans"}
	RegisterCoreCommands(root)
	known := map[string]bool{}
	for _, c := range root.Commands() {
		known[c.Name()] = true
		for _, a := range c.Aliases {
			known[a] = true
		}
	}

	for _, m := range regexp.MustCompile("`beans ([a-z]+)").FindAllStringSubmatch(out, -1) {
		if !known[m[1]] {
			t.Errorf("prompt names `beans %s`, which is not a command", m[1])
		}
	}
	for _, m := range regexp.MustCompile(`(?m)^beans ([a-z]+)`).FindAllStringSubmatch(out, -1) {
		if !known[m[1]] {
			t.Errorf("prompt example runs `beans %s`, which is not a command", m[1])
		}
	}
	if !strings.Contains(out, "beans next --json") {
		t.Error("prompt does not point agents at `beans next --json`")
	}
}
