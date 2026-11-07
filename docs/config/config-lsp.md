# LSP configuration

The `[lsp]` [configuration file](config.md) section provides settings to
fine-tune the [LSP editors integration](../tips/editors-integration.md).

## Completion

Customize how completion items appear in your editor when auto-completing links
with the `[lsp.completion]` sub-section.

| Setting                     | Type       | Description                                                                           |
| --------------------------- | ---------- | ------------------------------------------------------------------------------------- |
| `note-label`                | `template` | Label displayed in the completion pop-up for each note                                |
| `note-filter`               | `string`   | CLI filtering options used to select which notes are offered for link completion      |
| `note-filter-text`          | `template` | Text used as a source when filtering the completion pop-up with keystrokes            |
| `note-detail`               | `template` | Additional information about a completion item                                        |
| `use-additional-text-edits` | `boolean`  | Indicates whether `additionalTextEdits` will be used to remove the trigger characters |

`note-filter` accepts the same filtering options as `zk list`, for example
`"--tag project --sort modified-"`. Named filters declared in `[filter]` can
also be used.

Each template key accepts a [template](../notes/template.md) with the following
context:

| Variable        | Type   | Description                                                        |
| --------------- | ------ | ------------------------------------------------------------------ |
| `filename`      | string | Filename of the note, including its extension                      |
| `filename-stem` | string | Filename of the note without the file extension                    |
| `path`          | string | File path to the note, relative to the notebook root               |
| `abs-path`      | string | Absolute file path to the note                                     |
| `rel-path`      | string | File path to the note, relative to the current directory           |
| `title`         | string | Note title                                                         |
| `title-or-path` | string | Note title or path if empty                                        |
| `metadata`      | map    | YAML frontmatter metadata, e.g. `metadata.description`<sup>1</sup> |

1. YAML keys are normalized to lower case.

## Diagnostics

Use the `[lsp.diagnostics]` sub-section to configure how LSP diagnostics are
reported to your editors. Each diagnostic setting can be:

- An empty string or `none` to ignore this diagnostic.
- `hint`, `info`, `warning` or `error` to enable and set the severity of the
  diagnostic.

| Setting            | Default   | Description                                                               |
| ------------------ | --------- | ------------------------------------------------------------------------- |
| `wiki-title`       | `"none"`  | Report titles of wiki-links, which is useful if you use IDs for filenames |
| `dead-link`        | `"error"` | Warn for dead links between notes                                         |
| `self-link`        | `"none"`  | Warn when a note links to itself                                          |
| `missing-backlink` | `"none"`  | Warn when another notes link to current note without reciprocal backlinks |

### Missing backlink diagnostic

The `missing-backlink` diagnostic warns when other notes link to the current
note but the current note doesn't link back to them. Unlike other diagnostics,
this one requires a table with two fields:

- `level`: The severity (`hint`, `info`, `warning`, or `error`)
- `position`: Where to display the diagnostic (`top`, `bottom`, or
  `last-section`)

The `position` values are:

- `top`: First line of the note.
- `bottom`: Last line of the note.
- `last-section`: The last section heading (falls back to `"top"`).

Example:

```toml
[lsp.diagnostics]
missing-backlink = { level = "warning", position = "bottom" }
```

## Inlay hints

Use the `[lsp.inlay]` sub-section to show information inline, as virtual text
rendered by your editor. Inlay hints require an editor supporting LSP 3.17.

| Setting               | Type       | Default         | Description                                                   |
| --------------------- | ---------- | --------------- | ------------------------------------------------------------- |
| `wiki-title`          | `boolean`  | `false`         | Show the title of the linked note in wiki links without label |
| `wiki-title-template` | `template` | `"\|{{title}}"` | Text of the wiki-title hint                                   |

The hint is shown before the closing brackets, so `[[202605261823-fq2v]]`
displays as `[[202605261823-fq2v|Für Elise]]`. Links to notes without a title
get no hint.

`wiki-title-template` accepts the same template context as the
[completion](#completion) templates.

Unlike the `wiki-title` diagnostic, inlay hints don't read as warnings in your
editor, but they don't show up in diagnostic lists such as the quickfix list
either. You can enable both.

## Complete example

```toml
[lsp]

[lsp.diagnostics]
# Report titles of wiki-links as hints.
wiki-title = "hint"
# Warn for dead links between notes.
dead-link = "error"
# Warn when a note links to itself.
self-link = "warning"
# Report if backlinks are missing (shown at end of file).
missing-backlink = { level = "hint", position = "bottom" }

[lsp.inlay]
# Show the titles of wiki-links inline.
wiki-title = true
wiki-title-template = "|{{title}}"

[lsp.completion]
# Show the note title in the completion pop-up, or fallback on its path if empty.
note-label = "{{title-or-path}}"
# Only suggest notes matching the given filtering options.
note-filter = "--tag project --sort modified-"
# Filter out the completion pop-up using the note title or its path.
note-filter-text = "{{title}} {{path}}"
# Show the note filename without extension as detail.
note-detail = "{{filename-stem}}"
```
