package lsp

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/core"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// Inlay hints were introduced in LSP 3.17, which glsp doesn't support yet.
// The types and handler below implement the subset zk needs.
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_inlayHint

const methodTextDocumentInlayHint = "textDocument/inlayHint"

type inlayHintParams struct {
	TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	Range        protocol.Range                  `json:"range"`
}

type inlayHint struct {
	Position protocol.Position `json:"position"`
	Label    string            `json:"label"`
	Tooltip  string            `json:"tooltip,omitempty"`
}

// serverCapabilities extends the LSP 3.16 capabilities with the 3.17 ones zk
// supports.
type serverCapabilities struct {
	protocol.ServerCapabilities
	InlayHintProvider bool `json:"inlayHintProvider,omitempty"`
}

type initializeResult struct {
	Capabilities serverCapabilities                   `json:"capabilities"`
	ServerInfo   *protocol.InitializeResultServerInfo `json:"serverInfo,omitempty"`
}

// extendedHandler dispatches the requests unknown to glsp's 3.16 handler.
type extendedHandler struct {
	*protocol.Handler
	inlayHint func(context *glsp.Context, params *inlayHintParams) ([]inlayHint, error)
}

func (h *extendedHandler) Handle(context *glsp.Context) (r any, validMethod bool, validParams bool, err error) {
	if context.Method != methodTextDocumentInlayHint {
		return h.Handler.Handle(context)
	}
	if !h.IsInitialized() {
		return nil, true, true, errors.New("server not initialized")
	}

	var params inlayHintParams
	if err := json.Unmarshal(context.Params, &params); err != nil {
		return nil, true, false, err
	}
	r, err = h.inlayHint(context, &params)
	return r, true, true, err
}

// buildInlayHints returns the inlay hints of the given document within the
// range rng.
func (s *Server) buildInlayHints(notebook *core.Notebook, doc *document, rng protocol.Range) ([]inlayHint, error) {
	hints := []inlayHint{}

	config := notebook.Config.LSP.Inlay
	if !config.WikiTitle {
		return hints, nil
	}

	template, err := s.templateLoader.LoadTemplate(config.WikiTitleTemplate)
	if err != nil {
		return nil, err
	}

	links, err := doc.DocumentLinks()
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if !link.IsWikiLink || link.HasTitle || strutil.IsURL(link.Href) {
			continue
		}
		if !rangeContains(rng, link.LabelEnd) {
			continue
		}

		target, err := s.noteForLink(link, notebook)
		if err != nil {
			s.logger.Err(err)
			continue
		}
		if target == nil || target.Title == "" {
			continue
		}

		context, err := newCompletionItemRenderContext(target.MinimalNote, notebook.Path, filepath.Dir(doc.Path))
		if err != nil {
			return nil, err
		}
		label, err := template.Render(context)
		if err != nil {
			return nil, err
		}
		if label == "" {
			continue
		}

		hints = append(hints, inlayHint{
			Position: link.LabelEnd,
			Label:    label,
			Tooltip:  target.Path,
		})
	}

	return hints, nil
}

// rangeContains returns whether pos is within rng, bounds included. Unlike
// positionInRange, it accepts ranges extending past the end of the document.
func rangeContains(rng protocol.Range, pos protocol.Position) bool {
	return !positionBefore(pos, rng.Start) && !positionBefore(rng.End, pos)
}

// positionBefore returns whether a is strictly before b.
func positionBefore(a, b protocol.Position) bool {
	return a.Line < b.Line || (a.Line == b.Line && a.Character < b.Character)
}
