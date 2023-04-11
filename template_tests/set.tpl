{% set new_var = "hello" %}{{ new_var }}
{% block content %}{% set new_var = "world" %}{{ new_var }}{% endblock %}
{{ new_var }}{% for item in simple.misc_list %}
{% set new_var = item %}{{ new_var }}{% endfor %}
{{ new_var }}
{% set car=someUndefinedVar %}{{ car.Drive }}No Panic

{% set new_var = ["hello", "val2"] %}

{% for var in new_var %}{{ var }}
{% endfor %}

{% set int_variables = [1, 2, 3] %}{% for var in int_variables %}{{ var }}
{% endfor %}

{% set mixed_variables = [1, "two", 3.5, "4"] %}{% for var in mixed_variables %}{{ var }}
{% endfor %}

{% set empty_set = [] %}{% for var in empty_set %}{{ var }}
{% endfor %}

{% set item = "item1" %}
{% set ref_set = [item, "item2", some_random_variable] %}{% for var in ref_set %}{{ var }}
{% endfor %}

{% set nil_item_set = [nil] %}{% for var in nil_item_set %}-{{ var }}-{# printing an additional dash here to show the loop over the array with a nil item  #}
{% endfor %}

{% set nil_set = nil %}{% for var in nil_set %}-{{ var }} {# printing an additional dash here to show that the nil array won't be looped  #}
{% endfor %}
