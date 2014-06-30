add
{{ 5|add:2 }}
{{ 5|add:number }}
{{ 5|add:nothing }}
{{ 5|add:"test" }}
{{ "hello "|add:"flosch" }}
{{ "hello "|add:name }}

cut
{{ 15|cut:"5" }}
{{ "Hello world"|cut: " " }}

default
{{ nothing|default:"n/a" }}
{{ number|default:"n/a" }}
{{ 5|default:"n/a" }}

default_if_none
{{ nothing|default_if_none:"n/a" }}
{{ ""|default_if_none:"n/a" }}
{{ nil|default_if_none:"n/a" }}

safe
{{ "<script>" }}
{{ "<script>"|safe }}

escape
{{ "<script>"|safe|escape }}

divisibleby
{{ 21|divisibleby:3 }}
{{ 21|divisibleby:"3" }}
{{ 21|float|divisibleby:"3" }}
{{ 22|divisibleby:"3" }}

striptags
{{ "<strong><i>Hello!</i></strong>"|striptags|safe }}

removetags
{{ "<strong><i>Hello!</i></strong>"|removetags:"i"|safe }}