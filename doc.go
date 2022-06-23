// Package pongo2 is a Django-syntax like template-engine
//
// More info about pongo2: https://www.schlachter.tech/pongo2
//
// Complete documentation on the template language:
// https://docs.djangoproject.com/en/dev/topics/templates/
//
// Make sure to read README.md in the repository as well.
//
// A tiny example with template strings:
//
//
//     // Compile the template first (i. e. creating the AST)
//     tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
//     if err != nil {
//         panic(err)
//     }
//     // Now you can render the template with the given
//     // pongo2.Context how often you want to.
//     out, err := tpl.Execute(pongo2.Context{"name": "fred"})
//     if err != nil {
//         panic(err)
//     }
//     fmt.Println(out) // Output: Hello Fred!
//
package pongo2
