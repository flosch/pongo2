pongo2
======

**STILL A PREVIEW AND IN BETA - Not ready for production-use yet.**

pongo2 is the successor of [pongo](https://github.com/flosch/pongo), a Django-syntax like templating-language.

New in pongo2
-------------

 * Entirely rewritten from the ground-up.
 * Easy API to create new filters and tags (including parsing arguments); take a look on an example and the differences between pongo1 and pongo2: [old](https://github.com/flosch/pongo/blob/master/filters.go#L65) and [new](https://github.com/flosch/pongo2/blob/master/filters_builtin.go#L72).
 * Advanced C-like expressions.
 * Function calls from whereever you like.

What's missing
--------------

 * Several filters/tags (see `filters_builtin.go` and `tags.go` for a list of missing filters/tags). I try to implement the missing ones over time.
 * Tests
 * Documentation
 * Examples

Documentation
-------------

For a documentation on how the templating language works you can [head over to the Django documentation](https://docs.djangoproject.com/en/dev/topics/templates/). pongo2 aims to be fully compatible with it.

I still have to improve the pongo2-specific documentation over time. It will be available through [godoc](https://godoc.org/github.com/flosch/pongo2).