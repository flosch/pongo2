Start '{% include "includes.helper" %}' End
Start '{% include "includes.helper" with what_am_i=simple.name only %}' End
Start '{% include "includes.helper" with what_am_i=simple.name %}' End
Start '{% include simple.included_file|lower with number=7 what_am_i="guest" %}' End