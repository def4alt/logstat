package parser

import (
	"strings"
	"testing"
)

func TestProcessLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    LogEntry
		wantErr bool
	}{
		{
			name: "valid line with protocol",
			line: `h - u [28/Jul/1995:13:32:23 -0400] "GET /path HTTP/1.0" 200 2326`,
			want: LogEntry{
				Host: "h", Ident: "-", User: "u",
				Timestamp: "28/Jul/1995:13:32:23 -0400",
				Method: "GET", Path: "/path", Protocol: "HTTP/1.0",
				Status: "200", Bytes: "2326",
			},
		},
		{
			name: "valid line without protocol",
			line: `h - u [28/Jul/1995:13:32:23 -0400] "GET /path" 200 2326`,
			want: LogEntry{
				Host: "h", Ident: "-", User: "u",
				Timestamp: "28/Jul/1995:13:32:23 -0400",
				Method: "GET", Path: "/path", Protocol: "",
				Status: "200", Bytes: "2326",
			},
		},
		{
			name: "valid line with ident and user as dash",
			line: `199.0.2.27 - - [28/Jul/1995:13:32:23 -0400] "GET /image.gif HTTP/1.0" 200 5866`,
			want: LogEntry{
				Host: "199.0.2.27", Ident: "-", User: "-",
				Timestamp: "28/Jul/1995:13:32:23 -0400",
				Method: "GET", Path: "/image.gif", Protocol: "HTTP/1.0",
				Status: "200", Bytes: "5866",
			},
		},
		{
			name: "valid line with dash for bytes",
			line: `h - u [ts] "GET /path HTTP/1.0" 404 -`,
			want: LogEntry{
				Host: "h", Ident: "-", User: "u",
				Timestamp: "ts",
				Method: "GET", Path: "/path", Protocol: "HTTP/1.0",
				Status: "404", Bytes: "-",
			},
		},
		{
			name:    "malformed timestamp — missing closing bracket",
			line:    `h - u [28/Jul/1995:13:32:23 -0400 "GET /path" 200 2326`,
			wantErr: true,
		},
		{
			name:    "malformed timestamp — no brackets at all",
			line:    `h - u 28/Jul/1995:13:32:23 -0400 "GET /path" 200 2326`,
			wantErr: true,
		},
		{
			name:    "malformed request — only method, no path",
			line:    `h - u [ts] "GET" 200 2326`,
			wantErr: true,
		},
		{
			name:    "malformed request — unbalanced quotes",
			line:    `h - u [ts] "GET /path HTTP/1.0 200 2326`,
			wantErr: true,
		},
		{
			name:    "malformed request — three quotes corrupts data",
			line:    `h - u [ts] "GET /images/" HTTP/1.0" 404 -`,
			wantErr: true,
		},
		{
			name:    "invalid status — non-numeric",
			line:    `h - u [ts] "GET /path" abc 2326`,
			wantErr: true,
		},
		{
			name:    "invalid status — too short",
			line:    `h - u [ts] "GET /path" 20 2326`,
			wantErr: true,
		},
		{
			name:    "invalid status — too long",
			line:    `h - u [ts] "GET /path" 2000 2326`,
			wantErr: true,
		},
		{
			name:    "invalid status — out of range (099)",
			line:    `h - u [ts] "GET /path" 099 2326`,
			wantErr: true,
		},
		{
			name:    "invalid status — out of range (600)",
			line:    `h - u [ts] "GET /path" 600 2326`,
			wantErr: true,
		},
		{
			name:    "missing bytes — only 6 fields",
			line:    `h - u [ts] "GET /path" 200`,
			wantErr: true,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
		{
			name:    "too few fields — only host and ident",
			line:    `h -`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processLine(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("processLine() =\n%+v\nwant:\n%+v", got, tt.want)
			}
		})
	}
}

func TestProcessLog(t *testing.T) {
	t.Run("non-strict skips malformed lines", func(t *testing.T) {
		input := strings.NewReader(
			"h - u [ts] \"GET /ok HTTP/1.0\" 200 2326\n" +
				"bad line\n" +
				"h2 - u [ts] \"GET /ok2 HTTP/1.0\" 404 123\n",
		)

		entries, skipped, err := ProcessLog(input, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skipped != 1 {
			t.Errorf("expected 1 skipped, got %d", skipped)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Host != "h" || entries[1].Host != "h2" {
			t.Errorf("unexpected entries: %+v", entries)
		}
	})

	t.Run("strict aborts on malformed line", func(t *testing.T) {
		input := strings.NewReader(
			"h - u [ts] \"GET /ok HTTP/1.0\" 200 2326\n" +
				"bad line\n" +
				"h2 - u [ts] \"GET /ok2 HTTP/1.0\" 404 123\n",
		)

		entries, skipped, err := ProcessLog(input, true)
		if err == nil {
			t.Fatal("expected error, got none")
		}
		if skipped != 0 {
			t.Errorf("expected 0 skipped before abort, got %d", skipped)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries on abort, got %d", len(entries))
		}
	})

	t.Run("all valid lines pass strict", func(t *testing.T) {
		input := strings.NewReader(
			"h - u [ts] \"GET /a HTTP/1.0\" 200 1\n" +
				"h2 - u [ts] \"GET /b HTTP/1.0\" 404 0\n",
		)

		entries, skipped, err := ProcessLog(input, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if skipped != 0 {
			t.Errorf("expected 0 skipped, got %d", skipped)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})
}
