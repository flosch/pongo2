{% macro existing_function() %} exsisting {% endmacro %}

{%allowmissingval%} No tag {{not_defined_function()}}no error{%endallowmissingval%}

{%allowmissingval%} Finally calling an{{existing_function()}}value.{%endallowmissingval%}

{%allowmissingval%} This will return an empty{{not_defined_variable}}string.{%endallowmissingval%}