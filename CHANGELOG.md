# Changelog

## v7.0.0-alpha.2

This release brings pongo2 significantly closer to Django template behavior.

### Backwards-Incompatible Fixes

- **`timesince`/`timeuntil`**: Rewritten to use calendar-based month/year arithmetic and Django's adjacency rule (only adjacent time units shown). Future dates return "0&nbsp;minutes" for `timesince`, past dates for `timeuntil`.
- **`linebreaks`**: Rewritten paragraph algorithm to match Django (normalize newlines, strip trailing whitespace, correct `<p>`/`<br />` wrapping).
- **`title`**: Rewritten to match Django titlecase behavior (apostrophes don't break words, non-letter characters act as word boundaries).
- **`filesizeformat`**: Uses singular "byte", non-breaking spaces, and supports negative values.
- **`unordered_list`**: Tab indentation and newlines to match Django format.
- **`linenumbers`**: Zero-padded line numbers and normalized newlines.
- **`center`**: Corrected padding direction to match Django/Python `str.center()`.
- **`wordwrap`**: Wraps at character width instead of word count; normalize `\r\n` and `\r` to `\n` before wrapping.
- **`truncatewords`/`truncatewords_html`**: Use unicode ellipsis (\u2026) instead of three dots.

### Bug Fixes

- **`widthratio`**: Handle division by zero (return "0") and use banker's rounding (`math.RoundToEven`).
- **`truncatechars_html`**: Use rune count instead of byte length for multi-byte character support.
- **`slugify`**: Apply NFKD normalization for accented characters; preserve underscores.
- **`dictsort`**: Use numeric comparison for numeric fields instead of string comparison.
- **`get_digit`**: Handle non-digit characters correctly (return original value for non-integer strings).
- **`pluralize`**: Use float comparison to avoid truncating decimals (e.g., 1.5 is now plural).
- **`urlizetrunc`**: Use rune-based truncation for multi-byte URLs.
- **`yesno`**: Map `nil` to "no" value when only 2 custom args provided.
- **`linebreaksbr`**: Normalize `\r\n` and `\r` before processing.
- **`linebreaks`**: Normalize `\r\n` and `\r` before processing.
- **`json_script`**: Apply full HTML escaping to element ID (not just double quotes).
- **`urlencode`**: Document Django difference (spaces as `+` via `url.QueryEscape` vs Django's `%20`; `/` encoded vs Django preserving it).
- **`iriencode`**: Document Django difference (spaces as `+` via `url.QueryEscape` vs Django's `%20`). Switch from `bytes.Buffer` to `strings.Builder`.
- **`spaceless`**: Strip leading/trailing whitespace to match Django.
- **`lorem`**: Use double newline between paragraphs to match Django.
- **`cycle`**: Apply autoescape to output values like Django.
- **`filters`**: Mark HTML-producing filters (`escape`, `escapejs`, `linebreaks`, `linebreaksbr`, `urlize`, `urlizetrunc`) as safe to prevent double-escaping with autoescape.
- **`include`**: Use start token for lazy parse errors (fixes panic); support `{% include "file" only %}` without `with` keyword.
- **`macro`**: Store default values consistently with positional arguments.
- **`template`**: Use `sync.Once` for TrimBlocks/LStripBlocks token trimming (fixes race condition).
- **`template_sets`**: Close `io.ReadCloser` in `FromFile` (fixes resource leak with file-based loaders).
- **`template_sets`**: Detect recursive includes at parse time (falls back to lazy evaluation instead of infinite recursion).
- **`value`**: Add bounds checking to `Index()` and `Slice()` public API; return nil for unsupported types.
- **`value`**: Dereference pointers in `Contains()` and `GetItem()` type checks.
- **`value`**: Use numeric comparison for mixed int/float sorting.
- Fix operator precedence for `not X in Y` expressions.
- Fix float and int comparison by numeric value.
- Fix `ifchanged` to render else block for content-based comparison.
- Fix rune indexing for multi-byte UTF-8 strings in variable resolution.
- Fix format verb in `ljust`/`rjust` error messages.
- Fix autoescape state restoration on error.
- Fix lexer column tracking by character count instead of byte width.
- Fix `Negate()` return value and `Index()` log message.
- Fix `widthratio` to use `math.Round` for correct ratio calculation.
- Fix `ifnotequal` error message to say "ifnotequal" instead of "ifequal".
- Fix `cycle` and `ifchanged` tags to use per-execution state instead of shared AST node state (fixes race conditions).

### Refactoring

- Extract `normalizeNewlines` helper to reduce duplication across newline-normalizing filters.

### Testing

- Add comprehensive wordwrap tests covering word boundaries, mid-word lengths, newline handling, and Unicode.

### Documentation

- Fix inaccurate filter examples in code comments and docs (ellipsis characters, float rendering format, boolean capitalization, `floatformat` default, `json_script` spacing, `urlizetrunc` output).
- Fix API signatures in extension guides: `*Error` → `error` return types, `TagExists` → `BuiltinTagExists`, `INodeTag.Execute` return type, TokenType enum ordering.
- Update version to 7.0.0-alpha.2, move built-in filters out of addons list, update godoc links to `/v7`.

## v7.0.0-alpha.1

### Features

- New filters: `json_script`, `safeseq`, `escapeseq`, `dictsort`, `unordered_list`, `slugify`, `filesizeformat`, `timesince`, `timeuntil`.
- **[Backwards-Incompatible]** Improved `escapejs` filter to match Django behavior.
- **[Backwards-Incompatible]** Improved `striptags` and `removetags` filters with better security (recursive stripping, iteration limits).
- **[Backwards-Incompatible]** Use proper ellipsis character (…) in `truncatechars`, `truncatechars_html`, and `urlizetrunc` filters.
- **[Backwards-Incompatible]** Empty `{% filter %}` tag now returns a parse error (Django compatibility).
- **[Backwards-Incompatible]** Empty `{% firstof %}` tag now returns a parse error (Django compatibility).
- Support for negative number literals in arguments.
- Support for escape sequences in string literals (`\n`, `\t`, etc.).
- Support for sorted and reversed iteration over strings.
- Inline variable definitions (`{% set foo = "bar" %}`).
- Expand `urlize` filter to support more TLDs.

### Bug Fixes

- **[Backwards-Incompatible]** Fix `and`/`or` operators to return actual values instead of booleans (#362).
- Fix panic in `cycle` tag with no arguments.
- Fix integer overflow when converting `uint64` to `int`.
- Fix panic on uncomparable types in comparisons.
- Fix nil subscript panic.
- Fix `ifchanged` tag to only evaluate else block if it exists.
- Fix `in` operator type compatibility for maps.
- Fix `Contains` method to support all map key types (float64, bool, etc.).
- Fix array parser panic introduced by subscript feature.
- Support virtual filesystems in `ssi` tag plaintext mode and `Error.RawLine()`.
- Prevent memory exhaustion by limiting `lorem` tag generation.
- Prevent infinite loop in unclosed parameterized tag situations.
- Prevent stack overflow by limiting macro call depth.
- Prevent DoS via huge parameters in certain filters.
- Prevent panic from integer divide by zero.
- Fix string indexing to return character instead of byte (Django compatibility).
- Fix `NewSet` to validate that loaders are not nil.

### Performance

- Optimize lexer with keyword map and pre-compiled string replacer.
- Use pre-compiled `strings.Replacer` for HTML escaping filters.
- Cache `getResolvedValue()` result in Value methods.
- Optimize context valid identifier check (#340).
- Speed optimizations for `join` filter on long strings.

### Deprecations

- `ifequal` and `ifnotequal` tags now emit deprecation warnings (use `{% if %}` instead).
- `ssi` tag is deprecated.

### Breaking Changes

- **[Backwards-Incompatible]** Remove `HttpFilesystemLoader`. Use `FSLoader` with `NewFSLoader()` instead, which supports Go's `fs.FS` interface including `os.DirFS()` and `embed.FS`.
- **[Backwards-Incompatible]** Remove incomplete `SandboxedFilesystemLoader`. For sandboxing, use `BanTag()` to restrict `include`/`import`/`ssi`/`extends` tags.

### Other

- Go 1.25 is now the minimum required Go version.
- Added comprehensive fuzz testing infrastructure.
- Comprehensive documentation overhaul with new guides for getting started, template syntax, security, and custom extensions.

Thanks to all contributors.

## v6.0.0

- Go 1.18 is now the minimum required Go version.
- Improved block performance (#293).
- Support for variable subscript syntax (for maps/arrays/slices), such as `mymap["foo"]` or
  `myarray[0]` (#281).
- Backwards-incompatible change: `block.Super` won't be escaped anymore by default (#301).
- `nil` is now supported in function calls (#277).

Thanks to all contributors.
