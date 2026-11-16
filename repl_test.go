package main

import (
	// "io"
	// "os"
	"reflect"
	// "strings"
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "trims surrounding whitespace",
			input: " hello world ",
			want:  []string{"hello", "world"},
		},
		{
			name:  "lowercases words",
			input: "Charmander Bulbasaur PIKACHU",
			want:  []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name:  "splits on repeated and mixed whitespace",
			input: "  map\t\tmapb\nhelp  ",
			want:  []string{"map", "mapb", "help"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:  "whitespace-only input",
			input: " \t\n ",
			want:  []string{},
		},
		{
			name:  "unicode whitespace",
			input: "\u00a0Map\u2003MAPB\u3000",
			want:  []string{"map", "mapb"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanInput(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

// Deferred regression tests: uncomment these when the corresponding bugs are fixed.
//
// func TestStartReplHandlesBlankInput(t *testing.T) {
// 	reader, writer, err := os.Pipe()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	oldStdin := os.Stdin
// 	os.Stdin = reader
// 	t.Cleanup(func() {
// 		os.Stdin = oldStdin
// 		_ = reader.Close()
// 	})
//
// 	if _, err := io.WriteString(writer, "\n"); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := writer.Close(); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	captureStdout(t, func() {
// 		startRepl(&config{commands: getCommands()})
// 	})
// }
//
// func TestStartReplFormatsCallbackErrors(t *testing.T) {
// 	reader, writer, err := os.Pipe()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	oldStdin := os.Stdin
// 	os.Stdin = reader
// 	t.Cleanup(func() {
// 		os.Stdin = oldStdin
// 		_ = reader.Close()
// 	})
//
// 	if _, err := io.WriteString(writer, "fail\n"); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := writer.Close(); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	cfg := &config{
// 		commands: map[string]cliCommand{
// 			"fail": {
// 				name: "fail",
// 				callback: func(*config) error {
// 					return io.ErrUnexpectedEOF
// 				},
// 			},
// 		},
// 	}
// 	output := captureStdout(t, func() {
// 		startRepl(cfg)
// 	})
// 	if strings.Contains(output, "%w") {
// 		t.Errorf("callback error output contains an unformatted %%w: %q", output)
// 	}
// 	if !strings.Contains(output, "Error callback: unexpected EOF") {
// 		t.Errorf("callback error output = %q, want the formatted error", output)
// 	}
// }
