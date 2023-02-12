{% import "c.tpl" c %}
{% macro b() export %}b{{ c() }}{% endmacro %}
