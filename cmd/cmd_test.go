package cmd

import "testing"

func TestRootCommand_HasSubcommands(t *testing.T) {
	commands := rootCmd.Commands()
	expected := map[string]bool{
		"run":      false,
		"dotfiles": false,
		"install":  false,
		"repos":    false,
		"config":   false,
		"status":   false,
		"upgrade":  false,
	}

	for _, cmd := range commands {
		if _, ok := expected[cmd.Name()]; ok {
			expected[cmd.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestRootCommand_Flags(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"config", "c"},
		{"dry-run", "n"},
		{"verbose", "v"},
	}

	for _, f := range flags {
		flag := rootCmd.PersistentFlags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

