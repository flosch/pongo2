add
{{ 5|add:2 }}
{{ 5|add:simple.number }}
{{ 5|add:nothing }}
{{ 5|add:"test" }}
{{ "hello "|add:"flosch" }}
{{ "hello "|add:simple.name }}

cut
{{ 15|cut:"5" }}
{{ "Hello world"|cut: " " }}

default
{{ simple.nothing|default:"n/a" }}
{{ nothing|default:simple.number }}
{{ simple.number|default:"n/a" }}
{{ 5|default:"n/a" }}

default_if_none
{{ simple.nothing|default_if_none:"n/a" }}
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
{{ 85|divisibleby:simple.number }}
{{ 84|divisibleby:simple.number }}

striptags
{{ "<strong><i>Hello!</i></strong>"|striptags|safe }}

removetags
{{ "<strong><i>Hello!</i></strong>"|removetags:"i"|safe }}

yesno
{{ simple.bool_true|yesno }}
{{ simple.bool_false|yesno }}
{{ simple.nil|yesno }}
{{ simple.nothing|yesno }}
{{ simple.bool_true|yesno:"ja,nein,vielleicht" }}
{{ simple.bool_false|yesno:"ja,nein,vielleicht" }}
{{ simple.nothing|yesno:"ja,nein" }}

pluralize
customer{{ 0|pluralize }}
customer{{ 1|pluralize }}
customer{{ 2|pluralize }}
cherr{{ 0|pluralize:"y,ies" }}
cherr{{ 1|pluralize:"y,ies" }}
cherr{{ 2|pluralize:"y,ies" }}
walrus{{ 0|pluralize:"es" }}
walrus{{ 1|pluralize:"es" }}
walrus{{ simple.number|pluralize:"es" }}

first
{{ "Test"|first }}
{{ complex.comments|first }}
{{ 5|first }}
{{ true|first }}
{{ nothing|first }}

last
{{ "Test"|last }}
{{ complex.comments|last }}
{{ 5|last }}
{{ true|last }}
{{ nothing|last }}

urlencode
{{ "http://www.example.org/foo?a=b&c=d"|urlencode }}

linebreaksbr
{{ simple.newline_text|linebreaksbr }}
{{ ""|linebreaksbr }}
{{ "hallo"|linebreaksbr }}