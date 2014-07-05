{% for comment in complex.comments %}[{{ forloop.Counter }} {{ forloop.Counter0 }} {{ forloop.First }} {{ forloop.Last }} {{ forloop.Revcounter }} {{ forloop.Revcounter0 }}] {{ comment.Author.Name }}

{# nested loop #}
{% for char in comment.Text %}{{forloop.Parentloop.Counter0}}.{{forloop.Counter0}}:{{ char|safe }} {% endfor %}

{% endfor %}