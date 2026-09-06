# [pongo](https://en.wikipedia.org/wiki/Pongo_%28genus%29)2

[![PkgGoDev](https://pkg.go.dev/badge/github.com/flosch/pongo2/v7)](https://pkg.go.dev/github.com/flosch/pongo2/v7)
[![Build Status](https://schlachterinf.semaphoreci.com/badges/pongo2/branches/master.svg?style=shields&key=4ecde757-d6dd-4db2-8998-c84f55253d8f)](https://schlachterinf.semaphoreci.com/projects/pongo2)
[![Run on Repl.it](https://repl.it/badge/github/flosch/pongo2)](https://repl.it/github/flosch/pongo2)

pongo2 is a Django-syntax like templating-language ([official website](https://pongo2.dev/)).

Install/update using `go get` (no dependencies required by pongo2):

```sh
go get -u github.com/flosch/pongo2/v7
```

Please use the [issue tracker](https://github.com/flosch/pongo2/issues) if you're encountering any problems with pongo2 or if you need help with implementing tags or filters ([create a ticket!](https://github.com/flosch/pongo2/issues/new)).

:question: [**Looking for a Go developer/consultant?** I'm available for hire. :man_technologist:](mailto:florian@schlachter.tech )

## First impression of a template

```django
<html>
  <head>
    <title>Our admins and users</title>
  </head>
  {# This is a short example to give you a quick overview of pongo2's syntax. #}
  {% macro user_details(user, is_admin=false) %}
  <div class="user_item">
    <!-- Let's indicate a user's good karma -->
    <h2 {% if (user.karma>= 40) || (user.karma > calc_avg_karma(userlist)+5) %} class="karma-good"{% endif %}>

      <!-- This will call user.String() automatically if available: -->
      {{ user }}
    </h2>

    <!-- Will print a human-readable time duration like "3 weeks ago" -->
    <p>This user registered {{ user.register_date|naturaltime }}.</p>

    <!-- Let's allow the users to write down their biography using markdown;
             we will only show the first 15 words as a preview -->
    <p>The user's biography:</p>
    <p>
      {{ user.biography|markdown|truncatewords_html:15 }}
      <a href="/user/{{ user.id }}/">read more</a>
    </p>

    {% if is_admin %}
    <p>This user is an admin!</p>
    {% endif %}
  </div>
  {% endmacro %}

  <body>
    <!-- Make use of the macro defined above to avoid repetitive HTML code
         since we want to use the same code for admins AND members -->

    <h1>Our admins</h1>
    {% for admin in adminlist %} {{ user_details(admin, true) }} {% endfor %}

    <h1>Our members</h1>
    {% for user in userlist %} {{ user_details(user) }} {% endfor %}
  </body>
</html>
```

## Documentation

- [Getting Started](docs/getting-started.md) - Installation, basic usage, and first steps
- [Template Syntax](docs/template-syntax.md) - Variables, expressions, operators, and comments
- [Filters Reference](docs/filters.md) - Complete list of all built-in filters
- [Tags Reference](docs/tags.md) - Complete list of all built-in tags
- [Macros](docs/macros.md) - Reusable template fragments with arguments
- [Template Sets](docs/template-sets.md) - Loaders, caching, globals, and sandbox features
- [Security and Sandboxing](docs/security-sandboxing.md) - Autoescape, sandboxing, and security best practices
- [Custom Extensions](docs/custom-extensions.md) - Creating your own filters and tags
- [Changelog](CHANGELOG.md) - Version history and release notes

## Features

- Syntax- and feature-set-compatible with [Django templates](https://docs.djangoproject.com/en/stable/topics/templates/)
- [Advanced C-like expressions](https://github.com/flosch/pongo2/blob/master/template_tests/expressions.tpl).
- [Complex function calls within expressions](https://github.com/flosch/pongo2/blob/master/template_tests/function_calls_wrapper.tpl).
- [Easy API to create new filters and tags](https://pkg.go.dev/github.com/flosch/pongo2/v7#RegisterFilter) ([including parsing arguments](https://pkg.go.dev/github.com/flosch/pongo2/v7#Parser))
- Additional features:
  - Macros including importing macros from other files (see [template_tests/macro.tpl](https://github.com/flosch/pongo2/blob/master/template_tests/macro.tpl))
  - [Template sandboxing](https://pkg.go.dev/github.com/flosch/pongo2/v7#TemplateSet) ([directory patterns](http://golang.org/pkg/path/filepath/#Match), banned tags/filters)

## Caveats

### Filters

- **date** / **time**: The `date` and `time` filter are taking the Golang specific time- and date-format (not Django's one) currently. [Take a look on the format here](http://golang.org/pkg/time/#Time.Format).
- **stringformat**: `stringformat` does **not** take Python's string format syntax as a parameter, instead it takes Go's. Essentially `{{ 3.14|stringformat:"pi is %.2f" }}` is `fmt.Sprintf("pi is %.2f", 3.14)`.
- **escape** / **force_escape**: Unlike Django's behaviour, the `escape`-filter is applied immediately. Therefore there is no need for a `force_escape`-filter yet.

### Tags

- **for**: All the `forloop` fields (like `forloop.counter`) are written with a capital letter at the beginning. For example, the `counter` can be accessed by `forloop.Counter` and the parentloop by `forloop.Parentloop`.
- **now**: takes Go's time format (see **date** and **time**-filter).

### Misc

- **not in-operator**: You can check whether a map/struct/string contains a key/field/substring by using the in-operator (or the negation of it):
  `{% if key in map %}Key is in map{% else %}Key not in map{% endif %}` or `{% if !(key in map) %}Key is NOT in map{% else %}Key is in map{% endif %}`.

## Add-ons, libraries and helpers

### Official

- [pongo2-addons](https://github.com/flosch/pongo2-addons) - Official additional filters/tags for pongo2 (for example a **markdown**-filter). They are in their own repository because they're relying on 3rd-party-libraries.

### 3rd-party

- [beego-pongo2](https://github.com/oal/beego-pongo2) - A tiny little helper for using Pongo2 with [Beego](https://github.com/astaxie/beego).
- [beego-pongo2.v2](https://github.com/ipfans/beego-pongo2.v2) - Same as `beego-pongo2`, but for pongo2 v2.
- [macaron-pongo2](https://github.com/macaron-contrib/pongo2) - pongo2 support for [Macaron](https://github.com/Unknwon/macaron), a modular web framework.
- [ginpongo2](https://github.com/ngerakines/ginpongo2) - middleware for [gin](github.com/gin-gonic/gin) to use pongo2 templates
- [Build'n support for Iris' template engine](https://github.com/kataras/iris)
- [pongo2gin](https://gitlab.com/go-box/pongo2gin) - alternative renderer for [gin](github.com/gin-gonic/gin) to use pongo2 templates
- [pongo2-trans](https://github.com/digitalcrab/pongo2trans) - `trans`-tag implementation for internationalization
- [tpongo2](https://github.com/tango-contrib/tpongo2) - pongo2 support for [Tango](https://github.com/lunny/tango), a micro-kernel & pluggable web framework.
- [p2cli](https://github.com/wrouesnel/p2cli) - command line templating utility based on pongo2
- [pongorenderer](https://github.com/siredwin/pongorenderer) - minimal pongo2 renderer for [Echo](https://github.com/labstack/echo) web framework
- [pongo2gcloud](https://github.com/dieselburner/pongo2gcloud) - Google Cloud Storage loader for pongo2 template files

Please add your project to this list and send me a pull request when you've developed something nice for pongo2.

## Who's using pongo2

[I'm compiling a list of pongo2 users](https://github.com/flosch/pongo2/issues/241). Add your project or company!

## API-usage examples

Please see the documentation for a full list of provided API methods.

### A tiny example (template string)

```go
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
```

## Example server-usage (template file)

```go
package main

import (
    "github.com/flosch/pongo2/v7"
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
    err := tplExample.ExecuteWriter(pongo2.Context{"query": r.FormValue("query")}, w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/", examplePage)
    http.ListenAndServe(":8080", nil)
}
```


## 🌐 Web Resources & Interactive Index
- [CATEGORY CARTOON76](https://themindplay.github.io/category-cartoon76.html)
- [YOUR DREAM ROOM](https://frskillcrafts.pages.dev/your-dream-room.html)
- [MINETAP MERGE CLICKER](https://studyquests.github.io/minetap-merge-clicker.html)
- [ZUMBA STORY](https://studyplaying.github.io/zumba-story.html)
- [DIGIT SHOOTER](https://studyplayings.web.app/digit-shooter.html)
- [FLOAT FOR BRAINROTS](https://studyplaying.github.io/float-for-brainrots.html)
- [SQUID ESCAPE BUT BLOCKWORLD](https://quizverses-9d2f2.web.app/squid-escape-but-blockworld.html)
- [CATEGORY MAKEUP51](https://theskillquest.pages.dev/category-makeup51.html)
- [PARADISE JOURNEY MATCH3](https://quizverses.github.io/paradise-journey-match3.html)
- [INDEX13](https://quizverses.github.io/index13.html)
- [IMPOSTER COLORING BOOK](https://thelearnquesters.pages.dev/imposter-coloring-book.html)
- [CATEGORY GUN238](https://theskillquest.pages.dev/category-gun238.html)
- [TRIANGLE WAY](https://thequizzone.pages.dev/triangle-way.html)
- [CATEGORY FREE DRESS UP GAMES](https://quizverses.pages.dev/category-free-dress-up-games.html)
- [POPCAT CLICKER](https://studyplayings.web.app/popcat-clicker.html)
- [SAILOR CHIC VS PIRATE CHARM](https://studyplayings.web.app/sailor-chic-vs-pirate-charm.html)
- [AVATAR WORLD SECRETS](https://quizverses.github.io/avatar-world-secrets.html)
- [OBBY CARDS THE LEGEND HUNT](https://quizverses-9d2f2.web.app/obby-cards-the-legend-hunt.html)
- [RED LIGHT GREEN LIGHT](https://studyplayings.web.app/red-light-green-light.html)
- [PUSH TO GO](https://studyplaying.github.io/push-to-go.html)
- [CATEGORY BUBBLE SHOOTER GAMES](https://theskillquest.pages.dev/category-bubble-shooter-games.html)
- [CATEGORY PROXY](https://learnquester.github.io/category-proxy.html)
- [CATEGORY BUSINESS137](https://thequizzone.pages.dev/category-business137.html)
- [PIPE CONNECT](https://themindzone.pages.dev/pipe-connect.html)
- [VOXIOM IO](https://theskillquest.pages.dev/voxiom-io.html)
- [KAWAII REALM ADVENTURE](https://quizverses.github.io/kawaii-realm-adventure.html)
- [CATEGORY RESTAURANT64](https://learnquester.github.io/category-restaurant64.html)
- [CATEGORY STORY45](https://studyplayings.pages.dev/category-story45.html)
- [DRAW TO SMASH ZOMBIE](https://studyplaying.github.io/draw-to-smash-zombie.html)
- [BUILD A GO KART](https://themindzone.pages.dev/build-a-go-kart.html)
- [CATEGORY BUBBLE SHOOTER](https://theskillquest.pages.dev/category-bubble-shooter.html)
- [LABUBU DOLL MUKBANG ASMR UNBLOCKED](https://thequizzone.pages.dev/labubu-doll-mukbang-asmr-unblocked.html)
- [THE SORTING MART](https://thelearnquester.web.app/the-sorting-mart.html)
- [CATEGORY BOXING](https://iskillquest.pages.dev/category-boxing.html)
- [HAPPY FRUIT LINK](https://theskillquest.pages.dev/happy-fruit-link.html)
- [INTERIOR DESIGNER DECOR LIFE](https://themindzone.pages.dev/interior-designer-decor-life.html)
- [CAT LIFE SIMULATOR](https://studyplaying.github.io/cat-life-simulator.html)
- [CATEGORY TOWER DEFENSE118](https://quizverses.github.io/category-tower-defense118.html)
- [MANSION STORY MATCH](https://themindzone.pages.dev/mansion-story-match.html)
- [CATEGORY ROBOT49](https://theskillquest.pages.dev/category-robot49.html)
- [CATEGORY STICKMAN175](https://thequizzone.pages.dev/category-stickman175.html)
- [MINIGIANTS IO](https://themindzone.pages.dev/minigiants-io.html)
- [COLOR BUMP DANCER](https://quizverses.github.io/color-bump-dancer.html)
- [SKIBRONX RUNNER](https://theskillquest.pages.dev/skibronx-runner.html)
- [INDEX16](https://iskillquest.pages.dev/index16.html)
- [MR RECKLESS CAR CHASE SIMULATOR](https://themindzone.pages.dev/mr-reckless-car-chase-simulator.html)
- [CATEGORY 3 PLAYER26](https://iskillquest.pages.dev/category-3-player26.html)
- [CATEGORY CASUAL 10](https://quizverses.github.io/category-casual-10.html)
- [CATEGORY BRAIN261](https://thequizzone.pages.dev/category-brain261.html)
- [INDEX34](https://theskillquest.pages.dev/index34.html)
- [ONLINE PORTAL](https://brainquests-fb2c5.web.app/)
- [DRAW TO FLY](https://thequizzone.pages.dev/draw-to-fly.html)
- [CATEGORY RAGDOLL57](https://studyplayings.web.app/category-ragdoll57.html)
- [CATEGORY FIGHTING124](https://themindzone.pages.dev/category-fighting124.html)
- [HAPPY FARM THE CROP](https://studyplaying.github.io/happy-farm-the-crop.html)
- [CATEGORY THINKY](https://quizverses-9d2f2.web.app/category-thinky.html)
- [TREASURE HUNT PUZZLE](https://thequizzone.pages.dev/treasure-hunt-puzzle.html)
- [CUT THE GRASS 3D](https://thequizzone.pages.dev/cut-the-grass-3d.html)
- [CATEGORY DRESS UP](https://studyplayings.pages.dev/category-dress-up.html)
- [AOD ART OF DEFENSE](https://thequizzone.pages.dev/aod-art-of-defense.html)
- [VEGAMIX2 WILD WEST](https://theskillquest.pages.dev/vegamix2-wild-west.html)
- [SKY MAZE CHALLENGE](https://thequizzone.pages.dev/sky-maze-challenge.html)
- [EPIC MINE](https://thelearnquester.web.app/epic-mine.html)
- [INDEX8](https://studyplayings.web.app/index8.html)
- [GET TO THE CHOPPER](https://thelearnquesters.pages.dev/get-to-the-chopper.html)
- [INDEX9](https://iskillquest.pages.dev/index9.html)
- [DIGITAL CIRCUS RUN](https://quizverses-9d2f2.web.app/digital-circus-run.html)
- [LABUBU COLORING ADVENTURE](https://quizverses.github.io/labubu-coloring-adventure.html)
- [PAPERLY PAPER PLANE ADVENTURE](https://thequizzone.pages.dev/paperly-paper-plane-adventure.html)
- [SAVE LITTLE RED HOOD](https://themindzone.pages.dev/save-little-red-hood.html)
- [CATEGORY DRAWING34](https://studyplayings.pages.dev/category-drawing34.html)
- [XMAS PRESENTS MAHJONG](https://studyplaying.github.io/xmas-presents-mahjong.html)
- [BOARD KINGS BOARD DICE](https://studyplayings.web.app/board-kings-board-dice.html)
- [KINGDOM WARS TD](https://quizverses-9d2f2.web.app/kingdom-wars-td.html)
- [CATEGORY MOBILE2 095](https://theskillquest.pages.dev/category-mobile2-095.html)
- [LUCKY BRAINROT BLOCKS ONLINE](https://thelearnquesters.pages.dev/lucky-brainrot-blocks-online.html)
- [CATEGORY DRESS UP97](https://iskillquest.pages.dev/category-dress-up97.html)
- [RAMP CAR JUMPING](https://quizverses.pages.dev/ramp-car-jumping.html)
- [CATEGORY BATTLE523](https://iskillquest.pages.dev/category-battle523.html)
- [ONLINE PORTAL](https://quizverses.github.io/)
- [INDEX36](https://theskillquest.pages.dev/index36.html)
- [CRAZYSTEVEIO](https://quizverses-9d2f2.web.app/crazysteveio.html)
- [CATEGORY AGILITY](https://thequizzone.pages.dev/category-agility.html)
- [PIZZA PUZZLE](https://studyplayings.pages.dev/pizza-puzzle.html)
- [GUN WAR Z1](https://studyplayings.pages.dev/gun-war-z1.html)
- [SERIOUS HEAD](https://themindzone.pages.dev/serious-head.html)
- [CATEGORY LANSCHOOL](https://thequizzone.pages.dev/category-lanschool.html)
- [CYBER HIGHWAY ESCAPE](https://theskillquest.pages.dev/cyber-highway-escape.html)
- [SLIPPERY DRIFT RACING](https://thequizzone.pages.dev/slippery-drift-racing.html)
- [MINERS FURY](https://quizverses.github.io/miners-fury.html)
- [CATEGORY ESCAPE](https://quizverses.pages.dev/category-escape.html)
- [CATEGORY MINECRAFT 3](https://theskillquest.pages.dev/category-minecraft-3.html)
- [AUTHENTIC FOOTBALL](https://theskillquest.pages.dev/authentic-football.html)
- [STARRY STYLE DORAMA OF DREAM](https://themindzone.pages.dev/starry-style-dorama-of-dream.html)
- [SNAKE IO](https://themindzone.pages.dev/snake-io.html)
- [POPCORN STACK](https://studyplayings.pages.dev/popcorn-stack.html)
- [PUZZLE LINES AND KNOTS 1](https://quizverses-9d2f2.web.app/puzzle-lines-and-knots-1.html)
- [RAGDOLL BOB PUZZLE](https://studyplayings.pages.dev/ragdoll-bob-puzzle.html)
- [CITY DRIFT RACING](https://studyplayings.pages.dev/city-drift-racing.html)
- [CATEGORY POOL 3](https://thequizzone.pages.dev/category-pool-3.html)
- [SPRUNKI](https://themindzone.pages.dev/sprunki.html)
- [FIND 6 DIFFERENCES SPOT THE HIDDEN CHANGES](https://studyplaying.github.io/find-6-differences-spot-the-hidden-changes.html)
- [CATEGORY MOBILE2 095](https://themindzone.pages.dev/category-mobile2-095.html)
- [SMASHDOLL](https://studyplaying.github.io/smashdoll.html)
- [STELLAR FUSION](https://thelearnquesters.pages.dev/stellar-fusion.html)
- [REACH 2048](https://thequizzone.pages.dev/reach-2048.html)
- [BACKROOMS AMONG IMPOSTOR ROLLING GIANT](https://studyplaying.github.io/backrooms-among-impostor-rolling-giant.html)
- [LABUBU ADVENTURE](https://themindzone.pages.dev/labubu-adventure.html)
- [COLOR SORT IMPOSTOR EDITION](https://studyplaying.github.io/color-sort-impostor-edition.html)
- [SHADOWMAN RUNNER](https://studyplaying.github.io/shadowman-runner.html)
- [INDEX18](https://quizverses.pages.dev/index18.html)
- [NAIL QUEEN](https://themindzone.pages.dev/nail-queen.html)
- [CATEGORY QUIZ40](https://theskillquest.pages.dev/category-quiz40.html)
- [DRIVERZ ED](https://studyplaying.github.io/driverz-ed.html)
- [CATEGORY BATTLE ROYALE](https://quizverses.github.io/category-battle-royale.html)
- [PRINCESS ROYAL WEDDING](https://quizverses.github.io/princess-royal-wedding.html)
- [CATEGORY CASUAL](https://iskillquest.pages.dev/category-casual.html)
- [CATEGORY MAHJONG GAMES](https://quizverses-9d2f2.web.app/category-mahjong-games.html)
- [DEFENDERS OF THE REALM AN EPIC WAR](https://studyplaying.github.io/defenders-of-the-realm-an-epic-war.html)
- [MAGNET TRUCK](https://quizverses.github.io/magnet-truck.html)
- [CATEGORY 204828](https://quizverses.github.io/category-204828.html)
- [CATEGORY COLLECT565](https://quizverses.pages.dev/category-collect565.html)
- [MERGE ARCHER DEFENSE](https://themindzone.pages.dev/merge-archer-defense.html)
- [CATEGORY UNBLOCKED](https://quizverses.pages.dev/category-unblocked.html)
- [MOJICON FRUIT CONNECT](https://themindzone.pages.dev/mojicon-fruit-connect.html)
- [TWO CARTS DOWNHILL](https://thelearnquester.web.app/two-carts-downhill.html)
- [CATEGORY DRESS UP 3](https://iskillquest.pages.dev/category-dress-up-3.html)
- [CATEGORY BUBBLE SHOOTER](https://thequizzone.pages.dev/category-bubble-shooter.html)
- [EXCAVATOR SIMULATOR 3D](https://thelearnquesters.pages.dev/excavator-simulator-3d.html)
- [DRIFT IO](https://thelearnquesters.pages.dev/drift-io.html)
