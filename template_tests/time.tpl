{% if simple.time1 == simple.time1 %}equal{% else %}not equal{% endif %}
{% if simple.time1 > simple.time2 %}greater{% else %}not greater{% endif %}
{% if simple.time2 < simple.time1 %}lower{% else %}not lower{% endif %}
{% if simple.time1 >= simple.time2 %}greater or equal (greater){% else %}not greater or equal (greater){% endif %}
{% if simple.time1 >= simple.time1 %}greater or equal (equal){% else %}not greater or equal (equal){% endif %}
{% if simple.time2 <= simple.time1 %}lower or equal (lower){% else %}not lower or equal (lower){% endif %}
{% if simple.time2 <= simple.time2 %}lower or equal (equal){% else %}not lower or equal (equal){% endif %}
