{{ simple.func_add(simple.func_add(5, 15), simple.number) + 17 }}
{{ simple.func_add_iface(simple.func_add_iface(5, 15), simple.number) + 17 }}
{{ simple.func_variadic("hello") }}
{{ simple.func_variadic("hello, %s", simple.name) }}
{{ simple.func_variadic("%d + %d %s %d", 5, simple.number, "is", 49) }}
{{ simple.func_variadic_sum_int() }}
{{ simple.func_variadic_sum_int(1) }}
{{ simple.func_variadic_sum_int(1, 19, 185) }}
{{ simple.func_variadic_sum_int2() }}
{{ simple.func_variadic_sum_int2(2) }}
{{ simple.func_variadic_sum_int2(1, 7, 100) }}
eqnil: {{ simple.func_ensure_nil(nil) }}
neqnil: {{ simple.func_ensure_nil(1) }}
v1: {{ simple.func_ensure_nil_variadic(nil) }}
v2: {{ simple.func_ensure_nil_variadic() }}
v3: {{ simple.func_ensure_nil_variadic(nil, 1, nil, "test") }}