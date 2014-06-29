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