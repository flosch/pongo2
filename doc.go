// A Django-syntax like template-engine
//
// Introduction on what's new and a migration tutorial on:
// http://www.florian-schlachter.de/post/pongo2/
//
// A tiny example with template strings:
//
//     // Compile the template first (i. e. creating the AST)
//     tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
//     if err != nil {
//         panic(err)
//     }
//     // Now you can render the template with the given
//     // *pongo2.Context how often you want to.
//     out, err := tpl.Execute(&pongo2.Context{"name": "florian"})
//     if err != nil {
//         panic(err)
//     }
//     fmt.Println(out) // Output: Hello Florian!
//
package pongo2
