package stats

import (
	"testing"

	"github.com/def4alt/logstat/internal/parser"
)

func makeEntry(host, method, path, status, bytes string) parser.LogEntry {
	return parser.LogEntry{
		Host: host, Ident: "-", User: "-",
		Timestamp: "ts",
		Method: method, Path: path, Protocol: "HTTP/1.0",
		Status: status, Bytes: bytes,
	}
}

func TestTotalEntries(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h1", "GET", "/a", "200", "100"),
		makeEntry("h2", "POST", "/b", "404", "50"),
	}

	if got := TotalEntries(entries); got != 2 {
		t.Errorf("TotalEntries = %d, want 2", got)
	}

	if got := TotalEntries(nil); got != 0 {
		t.Errorf("TotalEntries(nil) = %d, want 0", got)
	}
}

func TestTotalBytes(t *testing.T) {
	tests := []struct {
		name string
		entries []parser.LogEntry
		want int
	}{
		{
			name: "sum of positive bytes",
			entries: []parser.LogEntry{
				makeEntry("h1", "GET", "/a", "200", "100"),
				makeEntry("h2", "GET", "/b", "200", "200"),
			},
			want: 300,
		},
		{
			name: "dash bytes treated as zero",
			entries: []parser.LogEntry{
				makeEntry("h1", "GET", "/a", "404", "-"),
				makeEntry("h2", "GET", "/b", "200", "50"),
			},
			want: 50,
		},
		{
			name: "empty bytes treated as zero",
			entries: []parser.LogEntry{
				makeEntry("h1", "GET", "/a", "200", ""),
			},
			want: 0,
		},
		{
			name: "empty entries",
			entries: nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalBytes(tt.entries); got != tt.want {
				t.Errorf("TotalBytes = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUniqueHosts(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h1", "GET", "/a", "200", "10"),
		makeEntry("h1", "GET", "/b", "200", "20"),
		makeEntry("h2", "GET", "/c", "404", "30"),
	}

	if got := UniqueHosts(entries); got != 2 {
		t.Errorf("UniqueHosts = %d, want 2", got)
	}

	if got := UniqueHosts(nil); got != 0 {
		t.Errorf("UniqueHosts(nil) = %d, want 0", got)
	}
}

func TestStatusCodeCounts(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h1", "GET", "/a", "200", "10"),
		makeEntry("h2", "GET", "/b", "404", "20"),
		makeEntry("h3", "POST", "/c", "200", "30"),
		makeEntry("h4", "GET", "/d", "500", "40"),
	}

	counts := StatusCodeCounts(entries)

	if counts["200"] != 2 {
		t.Errorf("counts[200] = %d, want 2", counts["200"])
	}
	if counts["404"] != 1 {
		t.Errorf("counts[404] = %d, want 1", counts["404"])
	}
	if counts["500"] != 1 {
		t.Errorf("counts[500] = %d, want 1", counts["500"])
	}
	if counts["302"] != 0 {
		t.Errorf("counts[302] = %d, want 0", counts["302"])
	}

	if got := StatusCodeCounts(nil); got == nil {
		t.Errorf("StatusCodeCounts(nil) = nil, want initialized map")
	}
}

func TestMethodCounts(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h1", "GET", "/a", "200", "10"),
		makeEntry("h2", "POST", "/b", "200", "20"),
		makeEntry("h3", "GET", "/c", "200", "30"),
	}

	counts := MethodCounts(entries)

	if counts["GET"] != 2 {
		t.Errorf("counts[GET] = %d, want 2", counts["GET"])
	}
	if counts["POST"] != 1 {
		t.Errorf("counts[POST] = %d, want 1", counts["POST"])
	}
	if counts["PUT"] != 0 {
		t.Errorf("counts[PUT] = %d, want 0", counts["PUT"])
	}
}

func TestTopKHosts(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("a", "GET", "/1", "200", "1"),
		makeEntry("b", "GET", "/2", "200", "1"),
		makeEntry("a", "GET", "/3", "200", "1"),
		makeEntry("c", "GET", "/4", "200", "1"),
		makeEntry("b", "GET", "/5", "200", "1"),
		makeEntry("b", "GET", "/6", "200", "1"),
	}

	top2 := TopKHosts(entries, 2)

	if len(top2) != 2 {
		t.Fatalf("len(top2) = %d, want 2", len(top2))
	}
	if top2[0].Key != "b" || top2[0].Value != 3 {
		t.Errorf("top2[0] = %+v, want {b 3}", top2[0])
	}
	if top2[1].Key != "a" || top2[1].Value != 2 {
		t.Errorf("top2[1] = %+v, want {a 2}", top2[1])
	}

	top10 := TopKHosts(entries, 10)
	if len(top10) != 3 {
		t.Errorf("len(top10) = %d, want 3 (only 3 unique hosts)", len(top10))
	}
}

func TestTopKPaths(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h", "GET", "/a", "200", "1"),
		makeEntry("h", "GET", "/b", "200", "1"),
		makeEntry("h", "GET", "/a", "200", "1"),
	}

	top := TopKPaths(entries, 1)

	if len(top) != 1 {
		t.Fatalf("len(top) = %d, want 1", len(top))
	}
	if top[0].Key != "/a" || top[0].Value != 2 {
		t.Errorf("top[0] = %+v, want {/a 2}", top[0])
	}
}

func TestPercentiles(t *testing.T) {
	// 10 entries with bytes 0..9
	entries := make([]parser.LogEntry, 10)
	for i := range entries {
		entries[i] = makeEntry("h", "GET", "/", "200", string(rune('0'+i)))
	}

	// Bytes: "0","1","2","3","4","5","6","7","8","9"
	// Sorted: 0,1,2,3,4,5,6,7,8,9
	// Sorted bytes: [0,1,2,3,4,5,6,7,8,9]
	// P50 (0.5)  = index int(10*0.5)  = 5 → value 5
	// P90 (0.9)  = index int(10*0.9)  = 9 → value 9
	// P95 (0.95) = index int(10*0.95) = 9 → value 9
	// P99 (0.99) = index int(10*0.99) = 9 → value 9

	tests := []struct {
		name string
		fn   func([]parser.LogEntry) int
		want int
	}{
		{"P50", P50Bytes, 5},
		{"P90", P90Bytes, 9},
		{"P95", P95Bytes, 9},
		{"P99", P99Bytes, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(entries); got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestPercentilesEmpty(t *testing.T) {
	tests := []struct {
		name string
		fn   func([]parser.LogEntry) int
	}{
		{"P50", P50Bytes},
		{"P90", P90Bytes},
		{"P95", P95Bytes},
		{"P99", P99Bytes},
	}

	for _, tt := range tests {
		t.Run(tt.name+" empty", func(t *testing.T) {
			if got := tt.fn(nil); got != 0 {
				t.Errorf("%s = %d, want 0", tt.name, got)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	entries := []parser.LogEntry{
		makeEntry("h1", "GET", "/a", "200", "100"),
		makeEntry("h1", "GET", "/b", "404", "0"),
		makeEntry("h2", "POST", "/a", "200", "50"),
	}

	summary := GenerateSummary(entries, 2, 5)

	if summary.TotalEntries != 3 {
		t.Errorf("TotalEntries = %d, want 3", summary.TotalEntries)
	}
	if summary.SkippedEntries != 2 {
		t.Errorf("SkippedEntries = %d, want 2", summary.SkippedEntries)
	}
	if summary.TotalBytes != 150 {
		t.Errorf("TotalBytes = %d, want 150", summary.TotalBytes)
	}
	if summary.UniqueHosts != 2 {
		t.Errorf("UniqueHosts = %d, want 2", summary.UniqueHosts)
	}
	if summary.StatusCodeCounts["200"] != 2 {
		t.Errorf("StatusCodeCounts[200] = %d, want 2", summary.StatusCodeCounts["200"])
	}
	if summary.MethodCounts["GET"] != 2 {
		t.Errorf("MethodCounts[GET] = %d, want 2", summary.MethodCounts["GET"])
	}
	if len(summary.TopKHosts) != 2 {
		t.Errorf("len(TopKHosts) = %d, want 2", len(summary.TopKHosts))
	}
	if len(summary.TopKPaths) != 2 {
		t.Errorf("len(TopKPaths) = %d, want 2", len(summary.TopKPaths))
	}
}
