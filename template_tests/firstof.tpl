{%allowmissingval%}{% firstof doesnotexist 42 %}{%endallowmissingval%}
{%allowmissingval%}{% firstof doesnotexist "<script>alert('xss');</script>" %}{%endallowmissingval%}
{%allowmissingval%}{% firstof doesnotexist "<script>alert('xss');</script>"|safe %}{%endallowmissingval%}
{%allowmissingval%}{% firstof doesnotexist simple.uint 42 %}{%endallowmissingval%}
{%allowmissingval%}{% firstof doesnotexist "test" simple.number 42 %}{%endallowmissingval%}
{% firstof %}
{% firstof "test" "test2" %}