package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/emrebaba-ui/pokedexcli/internal"
)

func TestGetCommands(t *testing.T) {
	want := map[string]string{
		"help": "Displays a help message",
		"exit": "Exit the Pokedex",
		"map":  "Displays the location names",
		"mapb": "Displays the previous location names",
		"explore": "Displays the names of all the Pokémon",
		"catch": "Adds Pokemon into the User's Pokedex",
		"inspect": "Checks if a Pokémon is in the Pokedex",
		"pokedex": "Lists all the Pokémon in the Pokedex",
	}

	got := getCommands()
	if len(got) != len(want) {
		t.Fatalf("getCommands() returned %d commands, want %d", len(got), len(want))
	}

	for name, description := range want {
		command, ok := got[name]
		if !ok {
			t.Errorf("getCommands() is missing command %q", name)
			continue
		}
		if command.name != name {
			t.Errorf("command %q has name %q", name, command.name)
		}
		if command.description != description {
			t.Errorf("command %q description = %q, want %q", name, command.description, description)
		}
		if command.callback == nil {
			t.Errorf("command %q has no callback", name)
		}
	}
}

func TestCommandHelp(t *testing.T) {
	output := captureStdout(t, func() {
		if err := commandHelp(&config{commands: getCommands()}); err != nil {
			t.Errorf("commandHelp() error = %v, want nil", err)
		}
	})

	for _, want := range []string{
		"Welcome to the Pokedex!",
		"Usage:",
		"help: Displays a help message",
		"exit: Exit the Pokedex",
		"map: Displays the location names",
		"mapb: Displays the previous location names",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("commandHelp() output %q does not contain %q", output, want)
		}
	}
}

func TestMapCommands(t *testing.T) {
	// Run the same behavior checks against both commands with different URLs.
	cache := internal.NewCache(time.Hour)
	names := make([]string, 20)
	for i := range names {
		names[i] = "location-" + string(rune('a'+i))
	}
	nextPage := "https://example.com/next"
	previousPage := "https://example.com/previous"

	tests := []struct {
		name        string
		command     func(*config, ...string) error
		requestPath string
		configure   func(string) *config
	}{
		{
			name:        "map",
			command:     commandMap,
			requestPath: "/next",
			configure: func(serverURL string) *config {
				return &config{next: serverURL + "/next", previous: "old-previous"}
			},
		},
		{
			name:        "mapb",
			command:     commandMapb,
			requestPath: "/previous",
			configure: func(serverURL string) *config {
				return &config{next: "old-next", previous: serverURL + "/previous"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make([]nameUrl, len(names))
			for i, name := range names {
				results[i] = nameUrl{Name: name, Url: "https://example.com/" + name}
			}

			// A local test server makes the response predictable and avoids calling PokeAPI.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("request method = %q, want %q", r.Method, http.MethodGet)
				}
				if r.URL.Path != tt.requestPath {
					t.Errorf("request path = %q, want %q", r.URL.Path, tt.requestPath)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(location{
					Count:    len(results),
					Next:     &nextPage,
					Previous: &previousPage,
					Results:  results,
				}); err != nil {
					t.Errorf("encoding test response: %v", err)
				}
			}))
			defer server.Close()

			// The command's URL points to the local server; its response supplies the next links.
			cfg := tt.configure(server.URL)
			cfg.client = &http.Client{Timeout: 3 * time.Second}
			cfg.cache = cache
			output := captureStdout(t, func() {
				if err := tt.command(cfg); err != nil {
					t.Errorf("%s command returned unexpected error: %v", tt.name, err)
				}
			})

			if cfg.next != nextPage {
				t.Errorf("config next = %q, want %q", cfg.next, nextPage)
			}
			if cfg.previous != previousPage {
				t.Errorf("config previous = %q, want %q", cfg.previous, previousPage)
			}

			wantOutput := strings.Join(names, "\n") + "\n"
			if output != wantOutput {
				t.Errorf("command output = %q, want %q", output, wantOutput)
			}
		})
	}
}

func TestMapCommandsKeepPaginationWhenResponseOmitsLinks(t *testing.T) {
	// Missing links should leave the existing pagination values unchanged.
	cache := internal.NewCache(time.Hour)
	tests := []struct {
		name    string
		command func(*config, ...string) error
		setURL  func(*config, string)
	}{
		{
			name:    "map",
			command: commandMap,
			setURL: func(cfg *config, url string) {
				cfg.next = url
			},
		},
		{
			name:    "mapb",
			command: commandMapb,
			setURL: func(cfg *config, url string) {
				cfg.previous = url
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make([]nameUrl, 20)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if err := json.NewEncoder(w).Encode(location{Results: results}); err != nil {
					t.Errorf("encoding test response: %v", err)
				}
			}))
			defer server.Close()

			cfg := newCommandTestConfig(cache)
			cfg.next = "keep-next"
			cfg.previous = "keep-previous"
			tt.setURL(cfg, server.URL)
			wantNext, wantPrevious := cfg.next, cfg.previous
			captureStdout(t, func() {
				if err := tt.command(cfg); err != nil {
					t.Errorf("%s command returned unexpected error: %v", tt.name, err)
				}
			})

			if cfg.next != wantNext {
				t.Errorf("config next = %q, want %q", cfg.next, wantNext)
			}
			if cfg.previous != wantPrevious {
				t.Errorf("config previous = %q, want %q", cfg.previous, wantPrevious)
			}
		})
	}
}

func TestMapCommandsReturnErrors(t *testing.T) {
	// These cases check failures before a valid page can be displayed.
	cache := internal.NewCache(time.Hour)
	tests := []struct {
		name    string
		command func(*config, ...string) error
		setURL  func(*config, string)
	}{
		{
			name:    "map",
			command: commandMap,
			setURL: func(cfg *config, url string) {
				cfg.next = url
			},
		},
		{
			name:    "mapb",
			command: commandMapb,
			setURL: func(cfg *config, url string) {
				cfg.previous = url
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/invalid URL", func(t *testing.T) {
			cfg := newCommandTestConfig(cache)
			tt.setURL(cfg, "://invalid")
			if err := tt.command(cfg); err == nil {
				t.Fatal("command returned nil error for an invalid URL")
			}
		})

		t.Run(tt.name+"/request failure", func(t *testing.T) {
			// Closing the server before the request simulates an unreachable API.
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			url := server.URL
			server.Close()

			cfg := newCommandTestConfig(cache)
			tt.setURL(cfg, url)
			err := tt.command(cfg)
			if err == nil {
				t.Fatal("command returned nil error when the server was unavailable")
			}
			if !strings.Contains(err.Error(), "getting response") {
				t.Errorf("error = %q, want a response error", err)
			}
		})

		t.Run(tt.name+"/invalid JSON", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, "not JSON")
			}))
			defer server.Close()

			cfg := newCommandTestConfig(cache)
			cfg.next = "keep-next"
			cfg.previous = "keep-previous"
			tt.setURL(cfg, server.URL)
			wantNext, wantPrevious := cfg.next, cfg.previous
			err := tt.command(cfg)
			if err == nil {
				t.Fatal("command returned nil error for an invalid JSON response")
			}
			if !strings.Contains(err.Error(), "decoding response") {
				t.Errorf("error = %q, want a decoding error", err)
			}
			if cfg.next != wantNext || cfg.previous != wantPrevious {
				t.Errorf("pagination changed after invalid response: next=%q previous=%q, want next=%q previous=%q", cfg.next, cfg.previous, wantNext, wantPrevious)
			}
		})
	}
}

func TestMapCommandsUseCachedResponse(t *testing.T) {
	tests := []struct {
		name    string
		command func(*config, ...string) error
		setURL  func(*config, string)
	}{
		{
			name:    "map",
			command: commandMap,
			setURL: func(cfg *config, url string) {
				cfg.next = url
			},
		},
		{
			name:    "mapb",
			command: commandMapb,
			setURL: func(cfg *config, url string) {
				cfg.previous = url
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := make([]string, 20)
			results := make([]nameUrl, len(names))
			for i := range names {
				names[i] = "cached-location"
				results[i] = nameUrl{Name: names[i]}
			}

			nextPage := "next-page"
			previousPage := "previous-page"
			page := location{
				Next:     &nextPage,
				Previous: &previousPage,
				Results:  results,
			}

			// The server is a fallback: a cache hit should display the saved page without requesting it.
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests++
				_ = json.NewEncoder(w).Encode(page)
			}))
			defer server.Close()

			cache := internal.NewCache(time.Hour)
			cfg := newCommandTestConfig(cache)
			tt.setURL(cfg, server.URL)
			cachedPage, err := json.Marshal(page)
			if err != nil {
				t.Fatalf("encoding cached page: %v", err)
			}
			cache.Add(server.URL, cachedPage)

			output := captureStdout(t, func() {
				if err := tt.command(cfg); err != nil {
					t.Errorf("%s command returned unexpected error: %v", tt.name, err)
				}
			})

			if requests != 0 {
				t.Errorf("command made %d HTTP requests for a cached page, want 0", requests)
			}
			if cfg.next != nextPage || cfg.previous != previousPage {
				t.Errorf("pagination = (%q, %q), want (%q, %q)", cfg.next, cfg.previous, nextPage, previousPage)
			}
			wantOutput := strings.Join(names, "\n") + "\n"
			if output != wantOutput {
				t.Errorf("command output = %q, want %q", output, wantOutput)
			}
		})
	}
}

func TestCommandExit(t *testing.T) {
	const childEnv = "POKEDEX_COMMAND_EXIT_TEST"
	if os.Getenv(childEnv) == "1" {
		// commandExit calls os.Exit, so run it in a subprocess to protect the test process.
		_ = commandExit(nil)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestCommandExit$")
	cmd.Env = append(os.Environ(), childEnv+"=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("exit command subprocess failed: %v\noutput: %s", err, output)
	}
	if got, want := string(output), "Closing the Pokedex... Goodbye!\n"; got != want {
		t.Errorf("commandExit() output = %q, want %q", got, want)
	}
}

func newCommandTestConfig(cache *internal.Cache) *config {
	return &config{
		client: &http.Client{Timeout: 3 * time.Second},
		cache:  cache,
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()

	// Temporarily redirect process stdout so command output can be asserted.
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
		_ = writer.Close()
		_ = reader.Close()
	}()

	run()
	if err := writer.Close(); err != nil {
		t.Fatalf("closing stdout pipe: %v", err)
	}
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(output)
}
