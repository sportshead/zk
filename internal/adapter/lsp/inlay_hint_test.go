package lsp

import (
	"encoding/json"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/adapter/fs"
	"github.com/zk-org/zk/internal/adapter/handlebars"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/adapter/sqlite"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func setupInlayHintTest(t *testing.T, inlay core.LSPInlayConfig) (*Server, *core.Notebook, *document) {
	fixture := getNotebookFixture("inlay")
	fs, err := fs.NewFileStorage(fixture.Path, &util.NullLogger)
	if err != nil {
		t.Fatal(err)
	}
	fixture.FS = fs

	db, err := sqlite.OpenInMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	config := core.NewDefaultConfig()
	config.LSP.Inlay = inlay
	index := sqlite.NewNoteIndex(fixture.Path, db, &util.NullLogger, config.Note.Extension)
	parser := markdown.NewParser(markdown.ParserOpts{}, &util.NullLogger)
	templateLoader := handlebars.NewLoader(handlebars.LoaderOpts{Styler: core.NullStyler})

	notebook := core.NewNotebook(fixture.Path, config, core.NotebookPorts{
		NoteIndex:         index,
		NoteContentParser: parser,
		FS:                fs,
		TemplateLoaderFactory: func(lang string) (core.TemplateLoader, error) {
			return templateLoader, nil
		},
		Logger: &util.NullLogger,
	})
	if _, err := notebook.Index(core.NoteIndexOpts{Force: true}); err != nil {
		t.Fatal(err)
	}

	docStore := newDocumentStore(fs, &util.NullLogger)
	server := &Server{
		documents:      docStore,
		templateLoader: templateLoader,
		logger:         &util.NullLogger,
	}

	param, err := fixture.MakeDidOpenParam("source.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := docStore.DidOpen(param, nil)
	if err != nil {
		t.Fatal(err)
	}
	return server, notebook, doc
}

var wholeDocument = protocol.Range{
	Start: protocol.Position{Line: 0, Character: 0},
	End:   protocol.Position{Line: 100, Character: 0},
}

func TestServer_buildInlayHints(t *testing.T) {
	server, notebook, doc := setupInlayHintTest(t, core.LSPInlayConfig{
		WikiTitle:         true,
		WikiTitleTemplate: "|{{title}}",
	})

	hints, err := server.buildInlayHints(notebook, doc, wholeDocument)
	assert.Nil(t, err)
	// Labeled, dead, untitled and Markdown links get no hint.
	assert.Equal(t, hints, []inlayHint{
		{Position: protocol.Position{Line: 2, Character: 7}, Label: "|Alpha", Tooltip: "alpha.md"},
		{Position: protocol.Position{Line: 3, Character: 33}, Label: "|Alpha", Tooltip: "alpha.md"},
	})
}

func TestServer_buildInlayHints_Range(t *testing.T) {
	server, notebook, doc := setupInlayHintTest(t, core.LSPInlayConfig{
		WikiTitle:         true,
		WikiTitleTemplate: "|{{title}}",
	})

	hints, err := server.buildInlayHints(notebook, doc, protocol.Range{
		Start: protocol.Position{Line: 3, Character: 0},
		End:   protocol.Position{Line: 4, Character: 0},
	})
	assert.Nil(t, err)
	assert.Equal(t, len(hints), 1)
	assert.Equal(t, hints[0].Position.Line, protocol.UInteger(3))
}

func TestServer_buildInlayHints_Template(t *testing.T) {
	server, notebook, doc := setupInlayHintTest(t, core.LSPInlayConfig{
		WikiTitle:         true,
		WikiTitleTemplate: " ({{path}})",
	})

	hints, err := server.buildInlayHints(notebook, doc, wholeDocument)
	assert.Nil(t, err)
	assert.Equal(t, hints[0].Label, " (alpha.md)")
}

func TestServer_buildInlayHints_Disabled(t *testing.T) {
	server, notebook, doc := setupInlayHintTest(t, core.LSPInlayConfig{
		WikiTitle:         false,
		WikiTitleTemplate: "|{{title}}",
	})

	hints, err := server.buildInlayHints(notebook, doc, wholeDocument)
	assert.Nil(t, err)
	assert.Equal(t, hints, []inlayHint{})
}

func TestInitializeResult_AdvertisesInlayHints(t *testing.T) {
	data, err := json.Marshal(initializeResult{
		Capabilities: serverCapabilities{
			ServerCapabilities: protocol.ServerCapabilities{HoverProvider: true},
			InlayHintProvider:  true,
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, string(data), `{"capabilities":{"hoverProvider":true,"inlayHintProvider":true}}`)
}
