Start '{% with what_am_i=simple.name %}I'm {{what_am_i}}{% endwith %}' End
Start '{% with what_am_i=simple.name %}I'm {{what_am_i}}11{% endwith %}' End
Start '{% with number=7 what_am_i="guest" %}I'm {{what_am_i}}{{number}}{% endwith %}' End
Start '{% include "with.helper" with what_am_i=simple.name number=10 %}' End