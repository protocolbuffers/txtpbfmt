package parser

import (
	"strings"
	"testing"

	"github.com/protocolbuffers/txtpbfmt/ast"
)

func findNode(name string, nodes []*ast.Node) *ast.Node {
	for _, n := range nodes {
		if n.Name == name {
			return n
		}
		if found := findNode(name, n.Children); found != nil {
			return found
		}
	}
	return nil
}

func mkString(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestParsePositions(t *testing.T) {
	for i, tc := range []struct {
		in                 string
		startByte, endByte int
		start, end         ast.Position
	}{{
		in:        `trackme: "foo"` + "\n",
		startByte: 0,
	}, {
		in: mkString(
			"# top",      // 5 bytes + newline
			"trackme: 4", // 10 bytes + newline
		),
		startByte: 0,
	}, {
		in: mkString(
			"# top",        // 5 bytes + newline
			"unrelated: 1", // 12 bytes + newline
			"trackme: 4",
		),
		startByte: 19,
	}, {
		in:        "trackme: {\n}",
		startByte: 0, endByte: 11,
		start: ast.Position{Line: 1, Column: 1},
		end:   ast.Position{Line: 2},
	}, {
		in: mkString(
			"# top",        // 5 bytes + newline
			"unrelated: 1", // 12 bytes + newline
			"",             // this already gets accounted for trackme
			"trackme: 4",
		),
		startByte: 19,
	}, {
		in: mkString(
			"# top",        // 5 bytes + newline
			"unrelated: 1", // 12 bytes + newline
			"# boo",        // this already gets accounted for trackme
			"trackme: 4",
		),
		startByte: 19,
	}, {
		in: mkString(
			"# top",        // 5 bytes + newline
			"unrelated: 1", // 12 bytes + newline
			"",             // this already gets accounted for trackme
			"# boo",
			"trackme: 4",
		),
		startByte: 19,
	}, {
		in: mkString(
			"outer: {", // 8 bytes + newline
			"  foo: 1", // 8 bytes + newline
			"  trackme: 4",
			"}",
		),
		startByte: 18,
	}, {
		in: mkString(
			"outer: {", // 8 bytes + newline
			"  foo: 1", // 8 bytes + newline
			"",         // acounted already for trackme
			"  # multiline desc",
			"  # for trackme",
			"  trackme: 4",
			"}",
		),
		startByte: 18,
	}, {
		in: mkString(
			"trackme: {",   // 10 bytes + newline
			"  content: 1", // 12 bytes + newline
			"}",
		),
		startByte: 0, endByte: 24,
		start: ast.Position{Line: 1, Column: 1},
		end:   ast.Position{Line: 3},
	}, {
		in: mkString(
			"outer: {",          // 8 bytes + newline
			"  a: 1",            // 6 bytes + newline
			"  trackme: {",      // 12 bytes + newline
			"    b: 2",          // 8 bytes + newline
			"    # end comment", // 17 bytes + newline
			"  }",
			"}",
		),
		startByte: 9 + 7, endByte: 9 + 7 + 13 + 9 + 18,
		start: ast.Position{Line: 3, Column: 1},
		end:   ast.Position{Line: 6},
	}} {
		nodes, err := Parse([]byte(tc.in))
		if err != nil {
			t.Fatal(err)
		}
		found := findNode("trackme", nodes)
		if found == nil {
			t.Fatalf("%d. TestParsePositions no node 'trackme' found", i)
		}
		if (uint32(tc.startByte) != found.Start.Byte) ||
			(uint32(tc.endByte) != found.End.Byte) ||
			(tc.start.Line > 0 && found.Start.Line != tc.start.Line) ||
			(tc.start.Column > 0 && found.Start.Column != tc.start.Column) ||
			(tc.end.Line > 0 && found.End.Line != tc.end.Line) {
			t.Errorf("%d. TestParsePositions got = %+v, want = (%d, %d), (%d, %d, %d); input:\n%s",
				i,
				found,
				tc.startByte, tc.endByte,
				tc.start.Line, tc.start.Column, tc.end.Line,
				tc.in)
		}

	}
}
