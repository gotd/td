package main

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gotd/td/wiretrace/uiattr"
)

// Alignment modes of diff --align.
const (
	alignPhase  = "phase"
	alignUI     = "ui"
	alignMarker = "marker"
)

// alignItem is one aligned unit: a UI action or marker (--align ui), or the
// segment of the trace that starts at a marker (--align marker).
type alignItem struct {
	token   string
	t       float64
	methods []string
}

type alignSide struct {
	items []alignItem
	// background are the methods of requests outside of the items.
	background []string
	start      float64
	hasUI      bool
}

func requestMethods(names *namer, reqs []*uiattr.Request) []string {
	out := make([]string, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, names.Name(r.Type))
	}
	return out
}

func newAlignSide(res *uiattr.Result, names *namer, mode string) *alignSide {
	s := &alignSide{hasUI: res.HasUI}
	// Offsets are from the first request, action or marker.
	first := true
	at := func(t float64) {
		if first || t < s.start {
			s.start, first = t, false
		}
	}
	for _, r := range res.Requests {
		at(r.T)
	}
	for _, a := range res.Actions {
		at(a.T)
	}
	for _, m := range res.Markers {
		at(m.T)
	}
	if mode == alignMarker {
		s.items = append(s.items, alignItem{token: "(start)", t: s.start})
		for _, m := range res.Markers {
			s.items = append(s.items, alignItem{token: m.Token(), t: m.T})
		}
		for _, r := range res.Requests {
			// The last segment that starts at or before the request.
			i := sort.Search(len(s.items)-1, func(i int) bool { return s.items[i+1].t > r.T })
			s.items[i].methods = append(s.items[i].methods, names.Name(r.Type))
		}
		return s
	}
	for _, a := range res.Actions {
		s.items = append(s.items, alignItem{token: a.Token(), t: a.T, methods: requestMethods(names, a.Requests)})
	}
	for _, m := range res.Markers {
		s.items = append(s.items, alignItem{token: m.Token(), t: m.T})
	}
	sort.SliceStable(s.items, func(i, j int) bool { return s.items[i].t < s.items[j].t })
	s.background = requestMethods(names, res.Background())
	return s
}

type alignOp struct {
	kind byte // '=', '-', '+'
	a, b *alignItem
}

// alignLCS aligns two item sequences by token.
func alignLCS(a, b []alignItem) []alignOp {
	n, m := len(a), len(b)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i].token == b[j].token:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var ops []alignOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i].token == b[j].token:
			ops = append(ops, alignOp{kind: '=', a: &a[i], b: &b[j]})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			ops = append(ops, alignOp{kind: '-', a: &a[i]})
			i++
		default:
			ops = append(ops, alignOp{kind: '+', b: &b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, alignOp{kind: '-', a: &a[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, alignOp{kind: '+', b: &b[j]})
	}
	return ops
}

func countMethods(methods []string) map[string]int {
	out := map[string]int{}
	for _, m := range methods {
		out[m]++
	}
	return out
}

// summarize is "messages.getHistory ×3, upload.getFile" in order of first
// appearance.
func summarize(methods []string) string {
	counts := countMethods(methods)
	var parts []string
	seen := map[string]bool{}
	for _, m := range methods {
		if seen[m] {
			continue
		}
		seen[m] = true
		if counts[m] > 1 {
			parts = append(parts, fmt.Sprintf("%s ×%d", m, counts[m]))
		} else {
			parts = append(parts, m)
		}
	}
	return strings.Join(parts, ", ")
}

// methodsDiff reports the multiset difference of two request sequences
// and, when the multisets are equal, a different order.
func methodsDiff(a, b []string) []string {
	ca, cb := countMethods(a), countMethods(b)
	keys := map[string]bool{}
	for k := range ca {
		keys[k] = true
	}
	for k := range cb {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	var out []string
	for _, k := range sorted {
		if ca[k] != cb[k] {
			out = append(out, fmt.Sprintf("~ %s: A %d, B %d", k, ca[k], cb[k]))
		}
	}
	if len(out) == 0 && strings.Join(a, ",") != strings.Join(b, ",") {
		out = append(out, fmt.Sprintf("order: A %s | B %s", strings.Join(a, ", "), strings.Join(b, ", ")))
	}
	return out
}

func reqCount(it *alignItem) string {
	if len(it.methods) == 0 {
		return "0 req"
	}
	return fmt.Sprintf("%d req: %s", len(it.methods), summarize(it.methods))
}

func printAlign(w io.Writer, mode string, a, b *alignSide, timing bool) {
	what := "actions and markers"
	if mode == alignMarker {
		what = "marker segments"
	}
	_, _ = fmt.Fprintf(w, "== align %s: A %d %s, B %d %s\n", mode, len(a.items), what, len(b.items), what)
	switch {
	case mode == alignUI && !a.hasUI && !b.hasUI:
		_, _ = fmt.Fprintln(w, "  (no UI events in A or B)")
	case mode == alignMarker && len(a.items) == 1 && len(b.items) == 1:
		_, _ = fmt.Fprintln(w, "  (no ui.marker in A or B)")
	}
	for _, op := range alignLCS(a.items, b.items) {
		var line string
		var diffs []string
		switch op.kind {
		case '=':
			line = fmt.Sprintf("  = %s   A %d req, B %d req", op.a.token, len(op.a.methods), len(op.b.methods))
			diffs = methodsDiff(op.a.methods, op.b.methods)
			if timing {
				line += fmt.Sprintf("   A+%.3fs B+%.3fs", op.a.t-a.start, op.b.t-b.start)
			}
		case '-':
			line = fmt.Sprintf("  - %s   A %s", op.a.token, reqCount(op.a))
			if timing {
				line += fmt.Sprintf("   A+%.3fs", op.a.t-a.start)
			}
		case '+':
			line = fmt.Sprintf("  + %s   B %s", op.b.token, reqCount(op.b))
			if timing {
				line += fmt.Sprintf("   B+%.3fs", op.b.t-b.start)
			}
		}
		_, _ = fmt.Fprintln(w, line)
		for _, d := range diffs {
			_, _ = fmt.Fprintf(w, "      %s\n", d)
		}
	}
	if mode == alignUI {
		_, _ = fmt.Fprintf(w, "== background (update-driven and unattributed): A %d req, B %d req\n",
			len(a.background), len(b.background))
		for _, d := range methodsDiff(a.background, b.background) {
			if strings.HasPrefix(d, "order:") {
				// Background order depends on timers; it is noise.
				continue
			}
			_, _ = fmt.Fprintf(w, "      %s\n", d)
		}
	}
}

func collectAlign(ctx context.Context, dir, schemaFlag, mode string, opt *uiOptions, stderr io.Writer) (*alignSide, error) {
	run, err := openRun(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = run.Close() }()
	schema, err := loadSchema(stderr, schemaFlag, run, false)
	if err != nil {
		return nil, err
	}
	var bad badLines
	// Peers are not needed for alignment: no schema, no join.
	res, err := buildUI(ctx, run, nil, opt.attr(), &bad, stderr)
	if err != nil {
		return nil, err
	}
	bad.report(stderr)
	return newAlignSide(res, newNamer(schema), mode), nil
}
