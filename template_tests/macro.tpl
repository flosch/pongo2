Begin
{% macro greetings(to, from=simple.name, name2="guest") %}
Greetings to {{ to }} from {{ from }}. Howdy, {% if name2 == "guest" %}anonymous guest{% else %}{{ name2 }}{% endif %}!
{% endmacro %}
{{ greetings() }}
{{ greetings(10) }}
{{ greetings("john") }}
{{ greetings("john", "michelle") }}
{{ greetings("john", "michelle", "johann") }}
{{ greetings("john", "michelle", "johann", "foobar") }}
End