integers and complex expressions
{{ 10-100 }}
{{ -(10-100) }}
{{ -(-(10-100)) }}
{{ -1 * (-(-(10-100))) }}
{{ -1 * (-(-(10-100)) ^ 2) ^ 3 + 3 * (5 - 17) + 1 + 2 }}

floats
{{ 5.5 }}
{{ 5.172841 }}
{{ 5.5 - 1.5 == 4 }}
{{ 5.5 - 1.5 == 4.0 }}

logic expressions
{{ !true }}
{{ !(true || false) }}
{{ true || false }}
{{ true or false }}
{{ false or false }}
{{ false || false }}
{{ true && (true && (true && (true && (1 == 1 || false)))) }}

float comparison
{{ 5.5 <= 5.5 }}
{{ 5.5 < 5.5 }}
{{ 5.5 > 5.5 }}
{{ 5.5 >= 5.5 }}

remainders
{{ (simple.number+7)%7 }}
{{ (simple.number+7)%7 == 0 }}
{{ (simple.number+7)%6 }}