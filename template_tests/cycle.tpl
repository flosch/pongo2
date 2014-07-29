{% for item in simple.multiple_item_list %}
    '{% cycle "item1" simple.name simple.number %}'
{% endfor %}
{% for item in simple.multiple_item_list %}
    {% cycle "item1" simple.name simple.number as cycleitem %}
    May I present the cycle item: '{{ cycleitem }}'
{% endfor %}