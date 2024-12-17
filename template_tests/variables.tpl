{{ 1 }}
{{ -5 }}
{{ "hallo" }}
{{ true }}
{{ false }}
{{ simple.uint }}
{{ simple.nil }}
{{ simple.str }}
{{ simple.bool_false }}
{{ simple.bool_true }}
{{ simple.uint }}
{{ simple.uint|integer }}
{{ simple.uint|float }}
{{ simple.multiple_item_list.10 }}
{{ simple.multiple_item_list.4 }}
{{ simple.misc_list[simple.uint - 8] }}
{{ simple.intmap[simple.uint - 7] }}
{{ simple.intmap["non-integer-key"] }}
{{ simple.strmap["ab" + "c"] }}
{{ complex.comments.0["Tex" + "t"]|safe }}
{{ complex.comments.0[0] }}
{{ simple.stringer }}
{{ simple.stringerPtr }}