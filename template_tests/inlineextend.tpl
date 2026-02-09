{% extends "inheritance/base_inline.tpl" %}

{% block body %}
Hello world!!
{% inlineextend "inheritance/modal.tpl" %}
{% block header %}Modal title{% endblock %}
{% block body %}Lorem ipsum{% endblock %}
{% block footer %}Modal footer{% endblock %}
{% endinlineextend -%}
{% endblock %}
