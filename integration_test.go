package wikilink_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"go.abhg.dev/goldmark/wikilink"
	"gopkg.in/yaml.v3"

	wikilinkparser "go.abhg.dev/goldmark/wikilink/parser"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	testsdata, err := os.ReadFile("testdata/tests.yaml")
	require.NoError(t, err)

	var tests []struct {
		Desc string `yaml:"desc"`
		Give string `yaml:"give"`
		Want string `yaml:"want"`
	}
	require.NoError(t, yaml.Unmarshal(testsdata, &tests))

	md := goldmark.New(goldmark.WithExtensions(&wikilink.Extender{
		Resolver:       _resolver,
		ParserResolver: _parserResolver,
	}))

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Desc, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, md.Convert([]byte(tt.Give), &buf))
			require.Equal(t, tt.Want, buf.String())
		})
	}
}

var (
	_resolver = resolver{}

	_parserResolver = parserResolver{}

	// Links with this target will return a nil destination.
	_doesNotExistTarget = []byte("Does Not Exist")
)

type resolver struct{}

func (resolver) ResolveWikilink(n *wikilink.Node) ([]byte, error) {
	if bytes.Equal(n.Target, _doesNotExistTarget) {
		return nil, nil
	}

	return wikilink.DefaultResolver.ResolveWikilink(n)
}

type parserResolver struct{}

func (parserResolver) ResolveWikilink(target, fragment []byte) ([]byte, error) {
	if bytes.Equal(target, _doesNotExistTarget) {
		return nil, nil
	}

	return wikilinkparser.DefaultResolver.ResolveWikilink(target, fragment)
}
