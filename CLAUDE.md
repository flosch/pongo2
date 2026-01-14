# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests
go test ./...

# Run a single test
go test -run TestName ./...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkFilterEscape -benchmem ./...

# Compare benchmarks with benchstat (required for commit messages)
go test -bench=. -benchmem -count=10 > old.txt
# make changes
go test -bench=. -benchmem -count=10 > new.txt
benchstat old.txt new.txt

# Run fuzz tests (lexer, template, value, filter, expression)
go test -fuzz=FuzzLexer -fuzztime=30s

# Linting
go vet ./...
golangci-lint run
```

## Development Guidelines

### Code Style

- Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).
- Use `go fmt`, `go test`, `go vet` and `golangci-lint run` before committing.
- Functions should rarely exceed 50 lines; decompose into smaller helpers if needed.
- Functions should have at most 4 parameters; group related parameters into a struct if more are needed.
- Interfaces should be small (ideally 1 method, max 3) and defined by the consumer.

### Testing

- Issues/bugs should be reproduced upfront by writing a test (usually in `pongo2_issues_test.go`) using `fstest`.
- Template tests (such as filters and tags) usually go into `template_tests/` and one of its test templates.
- Use Go's built-in test utilities: `testing/fstest`, `testing/iotest`, `testing/quick`.
- Tests using NewSet should use the DummyLoader instead of using a local file system.

### Benchmarking

- Run at least 10 iterations (ideally 20) for statistically significant results.
- Interleave before/after runs to evenly distribute noise.
- Include benchstat output in commit messages for performance-related changes.
- Do not rerun benchmarks seeking significant changes (avoids multiple testing bias).

## Architecture Overview

pongo2 is a Django-syntax template engine for Go. The execution flow is:

```
Template String → Lexer → Tokens → Parser → AST (INode tree) → Execute → Output
```

### Core Components

**Lexer** (`lexer.go`): State-machine tokenizer that converts template strings into tokens. Handles `{{ }}` variables, `{% %}` tags, `{# #}` comments, and whitespace trimming (`{{-`, `-}}`).

**Parser** (`parser.go`, `parser_document.go`, `parser_expression.go`): Converts tokens into an AST. Expression parser handles operator precedence: logical → relational → additive → multiplicative → power → factor.

**Value** (`value.go`): Reflection-based wrapper for dynamic Go values. Provides type checking (`IsString()`, `IsInteger()`, etc.) and operations (`EqualValueTo()`, `Contains()`).

**Variable Resolution** (`variable.go`): Handles property access (`user.name`), subscripts (`items[0]`), and function calls (`func(arg)`).

**Template Execution** (`template.go`, `context.go`): `ExecutionContext` holds Public (user data), Private (engine data like `forloop`), and Shared contexts during rendering.

### Extension Points

**Filters** (`filters.go`, `filters_builtin.go`): Functions that transform values. Register with `RegisterFilter("name", func(in, param *Value) (*Value, error))`.

**Tags** (`tags.go`, `tags_*.go`): Template control structures. Register with `RegisterTag("name", TagParser)`. Each tag implements `INodeTag` with an `Execute` method.

**Template Loaders** (`template_loader.go`): Implement `TemplateLoader` interface (`Abs`, `Get`) for custom template sources.

**TemplateSet** (`template_sets.go`): Groups templates with shared globals, sandboxing (ban tags/filters), and caching. Use `NewSet()` for isolated template groups.

### Template Inheritance

Templates support inheritance via `{% extends %}` and `{% block %}`. The `Template` struct maintains `parent`/`child` references and a `blocks` map for override resolution.

### Error Handling

Use `*Error` type (`error.go`) for returning errors from tags/filters. Set `Sender` to identify origin (e.g., `filter:myfilter` or `tag:mytag`). The error includes template location info for debugging.

## Testing Patterns

**Template tests** use `.tpl` files in `template_tests/` with corresponding `.tpl.out` expected output files. The test harness in `pongo2_template_test.go` automatically runs all template pairs.

**Error tests** use `.err` files with `.err.out` containing regex patterns to match expected errors:

- `*-compilation.err` / `*-compilation.err.out`: Parser/compilation errors
- `*-execution.err` / `*-execution.err.out`: Runtime execution errors

**Template options** can be configured per-test via `.tpl.options` files (e.g., `TrimBlocks=true`, `LStripBlocks=true`).

**Block tests** in `template_tests/block_render/` test `ExecuteBlocks()` for rendering specific named blocks.
