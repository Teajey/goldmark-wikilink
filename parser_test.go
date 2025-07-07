package wikilink

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	wikilinkparser "go.abhg.dev/goldmark/wikilink/parser"
)

func TestParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give string

		wantDestination string
		wantLabel       string
		wantEmbed       bool

		remainder string // unconsumed portion of tt.give
	}{
		{
			desc:            "simple",
			give:            "[[foo]] bar",
			wantDestination: "foo.html",
			wantLabel:       "foo",
			remainder:       " bar",
		},
		{
			desc:            "spaces",
			give:            "[[foo bar]]baz",
			wantDestination: "foo bar.html",
			wantLabel:       "foo bar",
			remainder:       "baz",
		},
		{
			desc:            "label",
			give:            "[[foo|bar]]",
			wantDestination: "foo.html",
			wantLabel:       "bar",
		},
		{
			desc:            "label with spaces",
			give:            "[[foo bar|baz qux]] quux",
			wantDestination: "foo bar.html",
			wantLabel:       "baz qux",
			remainder:       " quux",
		},
		{
			desc:            "fragment",
			give:            "[[foo#bar]] baz",
			wantDestination: "foo.html#bar",
			wantLabel:       "foo#bar",
			remainder:       " baz",
		},
		{
			desc:            "fragment with label",
			give:            "[[foo#bar|baz]]",
			wantDestination: "foo.html#bar",
			wantLabel:       "baz",
		},
		{
			desc:            "fragment without target",
			give:            "[[#foo]]",
			wantDestination: "#foo",
			wantLabel:       "#foo",
		},
		{
			desc:            "fragment without target with label",
			give:            "[[#foo|bar]]",
			wantDestination: "#foo",
			wantLabel:       "bar",
		},
		{
			desc:            "label with spaces. embedded",
			give:            "![[foo bar|baz qux]] quux",
			wantDestination: "foo bar.html",
			wantLabel:       "baz qux",
			remainder:       " quux",
			wantEmbed:       true,
		},
		{
			desc:            "fragment without target with label. embedded",
			give:            "![[#foo|bar]]",
			wantDestination: "#foo",
			wantLabel:       "bar",
			wantEmbed:       true,
		},
		{
			desc:            "fragment without target with label. embedded",
			give:            "![[baz#foo|bar]]",
			wantDestination: "baz.html#foo",
			wantLabel:       "bar",
			wantEmbed:       true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			r := text.NewReader([]byte(tt.give))

			p := Parser{
				Resolver: wikilinkparser.DefaultResolver,
			}
			got := p.Parse(nil /* parent */, r, parser.NewContext())
			require.NotNil(t, got, "expected Node, got nil")

			switch n := got.(type) {
			case *ast.Link:
				assert.Equal(t, tt.wantDestination, string(n.Destination), "Destination mismatch")
				assert.False(t, tt.wantEmbed, "embed mismatch")
			case *ast.Image:
				assert.Equal(t, tt.wantDestination, string(n.Destination), "Destination mismatch")
				assert.True(t, tt.wantEmbed, "embed mismatch")
			default:
				assert.Fail(t, "expected *ast.Link, got %T", got)
			}

			if assert.Equal(t, 1, got.ChildCount(), "children mismatch") {
				child := got.FirstChild()
				if label, ok := child.(*ast.Text); assert.True(t, ok, "expected Text, got %T", child) {
					assert.Equal(t, tt.wantLabel, string(r.Value(label.Segment)), "label mismatch")
				}
			}

			_, pos := r.Position()
			assert.Equal(t, tt.remainder, string(r.Value(pos)),
				"remaining text does not match")
		})
	}
}
