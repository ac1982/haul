# README translations

[README.md](../README.md) is the English source. The repository includes 15
translations beside it, for 16 reading options in total:

`ar`, `arz`, `bn`, `de`, `es`, `fr`, `hi`, `id`, `ja`, `pcm`, `pt`, `ru`, `ur`,
`zh-CN`, and `zh-TW`.

## Updating a translation

- Update the corresponding sections in every `README.<locale>.md` when the
  English instructions change. The translated pages cover the full guide.
- Translate prose, headings, example captions, and accessibility text. Keep
  commands, flags, URLs, file paths, template variables, JSON keys and values,
  and actual terminal output unchanged. Explain commands in adjacent prose.
- Preserve all installation requirements, supported link forms, option groups,
  exit codes, configuration examples, and site-specific behavior.
- Keep the language selectors synchronized. Show each language's own name,
  highlight the current version, and link to the other 15 files. Use `ja` for
  Japanese and `id` for Indonesian.
- Keep explicit navigation anchors stable when translating headings: give every
  section an `<a id>` with the English heading's slug (`install`,
  `for-ai-agents-and-scripts`, `site-notes`, …) and keep the older ids
  (`binary`, `examples`, `agents`, `notes`) beside them. Keep a blank line
  around Markdown inside HTML containers and disclosure panels.
- The command reference is a two-column table: the commands as inline code,
  their meaning translated. It replaces the English `text` block, whose right
  column is prose.
- The language selector appears once, at the top, as on the English page.
- Arabic (`ar`), Egyptian Arabic (`arz`), and Urdu (`ur`) use an outer
  `dir="rtl"` container, `dir="ltr"` containers for fenced examples, and
  `<bdi dir="ltr">` around inline code. Preserve these direction boundaries.
- Translations share [brand.svg](assets/brand.svg), which contains the haul
  wordmark and icons without explanatory text. The tagline and workflow are
  localized as selectable text below it. The English page retains its existing
  [hero](assets/hero.svg) and [workflow](assets/workflow.svg) artwork.

## Review before committing

Compare executable examples and parsed JSON with the English source; only
shell comments may differ. Check option and template-variable coverage, all
local links, language selectors, and section anchors. Render the Markdown and
inspect tables, disclosure panels, narrow layouts, and right-to-left text with
left-to-right commands. Include any referenced artwork in the commit.
