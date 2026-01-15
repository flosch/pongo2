# Changelog

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
