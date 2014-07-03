// A Django-syntax like template-engine
//
// Current caveats
//  - Parallelism: Please make sure you're not sharing the Context-object you're passing to Execute() between several parallel Execute() function calls. You will have to create your own pongo2.Context per Execute() call.
//  - date/time-filter: The date and time filter are taking the Golang specific time- and date-format (not Django's one) currently. Take a look on the format here.
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
