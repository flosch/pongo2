{% with val = "{{ simple.name }}" %}{% exec %}{{val}}{% endexec %}{% endwith %}

{% exec %}{% macro m1() %}some text{% endmacro %}{% endexec %}{{m1()}}

{% exec %}{% autoescape off %}{% macro m2(arg) export %}some text with {{arg}}{% endmacro %}{% endautoescape %}{% endexec %}{{m2("arg value")}}

{% with val = "{% macro m3() export %}m3 text{% endmacro %}{{m3()}}" %}{% exec %}{{val}}{% endexec %}{% endwith %}

{% with val = '{% import "template_tests/exec.helper" m4 %}' %}{% exec %}{% autoescape off %}{{val}}{% endautoescape %}{% endexec %}{{m4("arg value")}}{% endwith %}

{% exec %}{% macro m1() export %}some text{% endmacro %}{% endexec %}{{m1()}}

{% with chars = "abc"|make_list val = "{% for i in chars %}{{i}}{% endfor %}" %}{% exec %}{{val}}{% endexec %}{% endwith %}