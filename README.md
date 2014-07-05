# pongo2

[![GoDoc](https://godoc.org/github.com/flosch/pongo2?status.png)](https://godoc.org/github.com/flosch/pongo2)
[![Build Status](https://travis-ci.org/flosch/pongo2.svg?branch=master)](https://travis-ci.org/flosch/pongo2)
[![Coverage Status](https://coveralls.io/repos/flosch/pongo2/badge.png?branch=master)](https://coveralls.io/r/flosch/pongo2?branch=master)
[![GitTip](http://img.shields.io/badge/gittip-support%20pongo-brightgreen.svg)](https://www.gittip.com/flosch/)

pongo2 is the successor of [pongo](https://github.com/flosch/pongo), a Django-syntax like templating-language.

[See my blog post announcement about pongo2 and for a migration- and a "how to write tags/filters"-tutorial.](http://www.florian-schlachter.de/post/pongo2/)

Please use the [issue tracker](https://github.com/flosch/pongo2/issues) if you're encountering any problems with pongo2 or if you need help with implementing tags or filters (create a ticket!).

pongo2 is **still in beta** and under heavy development.

## New in pongo2

 * Entirely rewritten from the ground-up.
 * Easy API to create new filters and tags (including parsing arguments); take a look on an example and the differences between pongo1 and pongo2: [old](https://github.com/flosch/pongo/blob/master/filters.go#L65) and [new](https://github.com/flosch/pongo2/blob/master/filters_builtin.go#L72).
 * Advanced C-like expressions.
 * Complex function calls within expressions.

## What's missing

 * Several filters/tags (see `filters_builtin.go` and `tags.go` for a list of missing filters/tags). I try to implement the missing ones over time.
 * Tests
 * Examples
 * Documentation

## How you can help

 * Write filters / tags (see [tutorial](http://www.florian-schlachter.de/post/pongo2/)) by forking pongo2 and sending pull requests
 * Write tests (use the following command to see what tests are missing: `go test -v -cover -covermode=count -coverprofile=cover.out && go tool cover -html=cover.out`)

# Documentation

For a documentation on how the templating language works you can [head over to the Django documentation](https://docs.djangoproject.com/en/dev/topics/templates/). pongo2 aims to be fully compatible with it.

You can access pongo2's documentation on [godoc](https://godoc.org/github.com/flosch/pongo2).

## Caveats

### General 

 * **Parallelism**: Please make sure you're not sharing the Context-object you're passing to `Execute()` between several parallel `Execute()` function calls. You will have to create your own `pongo2.Context` per `Execute()` call.

### Filters

 * **date** / **time**: The `date` and `time` filter are taking the Golang specific time- and date-format (not Django's one) currently. [Take a look on the format here](http://golang.org/pkg/time/#Time.Format).

# Examples

## A tiny example (template string)

	// Compile the template first (i. e. creating the AST)
	tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
	if err != nil {
		panic(err)
	}
	// Now you can render the template with the given 
	// pongo2.Context how often you want to.
	out, err := tpl.Execute(pongo2.Context{"name": "florian"})
	if err != nil {
		panic(err)
	}
	fmt.Println(out) // Output: Hello Florian!

## Example server-usage (template file)

	package main
	
	import (
		"github.com/flosch/pongo2"
		"net/http"
	)
	
	// Pre-compiling the templates at application startup using the
	// little Must()-helper function (Must() will panic if FromFile()
	// or FromString() will return with an error - that's it).
	// It's faster to pre-compile it anywhere at startup and only
	// execute the template later.
	var tplExample = pongo2.Must(pongo2.FromFile("example.html"))
	
	func examplePage(w http.ResponseWriter, r *http.Request) {
		// Execute the template per HTTP request
		err := tplExample.ExecuteRW(w, pongo2.Context{"query": r.FormValue("query")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	
	func main() {
		http.HandleFunc("/", examplePage)
		http.ListenAndServe(":8080", nil)
	}

# Benchmark

The benchmarks have been run on the my machine (`Intel(R) Core(TM) i7-2600 CPU @ 3.40GHz`) using the command:

    go test -bench . -cpu 1,2,4,8

All benchmarks are compiling (depends on the benchmark) and executing the `template_tests/complex.tpl` template.

The results are:

	BenchmarkExecuteComplex                    50000             66720 ns/op
	BenchmarkExecuteComplex-2                  50000             67013 ns/op
	BenchmarkExecuteComplex-4                  50000             67807 ns/op
	BenchmarkExecuteComplex-8                  50000             68147 ns/op
	BenchmarkCompileAndExecuteComplex          10000            153411 ns/op
	BenchmarkCompileAndExecuteComplex-2        10000            145334 ns/op
	BenchmarkCompileAndExecuteComplex-4        10000            156475 ns/op
	BenchmarkCompileAndExecuteComplex-8        10000            162995 ns/op
	BenchmarkParallelExecuteComplex            50000             65041 ns/op
	BenchmarkParallelExecuteComplex-2          50000             35034 ns/op
	BenchmarkParallelExecuteComplex-4         100000             25046 ns/op
	BenchmarkParallelExecuteComplex-8         100000             22447 ns/op
