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

mul/div
{{ 2 * 5 }}
{{ 2 * 5.0 }}
{{ 2 * 0 }}
{{ 2.5 * 5.3 }}
{{ 1/2 }}
{{ 1/2.0 }}
{{ 1/0.000001 }}

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

in/not in
{{ 5 in simple.intmap }}
{{ 2 in simple.intmap }}
{{ 7 in simple.intmap }}
{{ !(5 in simple.intmap) }}
{{ not(7 in simple.intmap) }}
{{ 1 in simple.multiple_item_list }}
{{ 4 in simple.multiple_item_list }}
{{ !(4 in simple.multiple_item_list) }}
{{ "Hello" in simple.misc_list }}
{{ "Hello2" in simple.misc_list }}
{{ 99 in simple.misc_list }}
{{ False in simple.misc_list }}

issue #48 (associativity for infix operators)
{{ 34/3*3 }}
{{ 10 + 24 / 6 / 2 }}
{{ 6 - 4 - 2 }}

issue #64 (uint comparison with int const)
{{ simple.uint }}
{{ simple.uint == 8 }}
{{ simple.uint == 9 }}
{{ simple.uint >= 8 }}
{{ simple.uint <= 8 }}
{{ simple.uint < 8 }}
{{ simple.uint > 8 }}

string concatenation
{{ "a" + "b" }}
{{ 1 + "a" }}
{{ "a" + "1" }}

operator precedence (PEMDAS/BODMAS)
{{ 2 + 3 * 4 }}
{{ 2 * 3 + 4 }}
{{ 10 - 2 * 3 }}
{{ 10 / 2 + 3 }}
{{ 10 + 6 / 2 }}
{{ 2 + 3 * 4 - 5 }}
{{ 2 ^ 3 * 4 }}
{{ 4 * 2 ^ 3 }}
{{ 2 ^ 3 + 4 * 5 }}
{{ 100 / 10 / 2 }}
{{ 2 ^ 2 ^ 3 }}

deeply nested parentheses
{{ ((((1 + 2)))) }}
{{ (((2 * 3) + 4) * 5) }}
{{ ((2 + 3) * (4 + 5)) }}
{{ (((10 - 5) * 2) + ((3 + 2) * 4)) }}
{{ ((((1 + 1) * 2) + 1) * 2) + 1 }}

complex mixed operations
{{ 1 + 2 * 3 - 4 / 2 + 5 % 3 }}
{{ (1 + 2) * (3 - 4) / (2 + 5) % 3 }}
{{ 10 * 2 + 30 / 3 - 5 * 2 + 1 }}
{{ 100 - 50 + 25 - 12 + 6 - 3 + 1 }}
{{ 2 * 3 * 4 * 5 / 10 / 2 }}

power operator edge cases
{{ 2 ^ 0 }}
{{ 0 ^ 5 }}
{{ 1 ^ 100 }}
{{ 10 ^ 2 }}
{{ 2 ^ 10 }}
{{ 3 ^ 3 ^ 0 }}
{{ (3 ^ 3) ^ 0 }}
{{ 2 ^ 3 ^ 2 }}

negative numbers and signs
{{ -5 + 10 }}
{{ 10 + -5 }}
{{ -5 * -3 }}
{{ -10 / -2 }}
{{ -2 ^ 2 }}
{{ (-2) ^ 2 }}
{{ (-2) ^ 3 }}
{{ -(-5) }}
{{ -(-(-5)) }}

modulo operations
{{ 17 % 5 }}
{{ 100 % 7 }}
{{ 15 % 5 }}
{{ 7 % 10 }}
{{ -7 % 3 }}
{{ 10 % 3 % 2 }}

mixed integer and float
{{ 5 + 2.5 }}
{{ 10 - 3.5 }}
{{ 4 * 2.5 }}
{{ 10 / 4.0 }}
{{ 2.5 * 2 + 3.5 }}
{{ (5 + 2.5) * 2 }}
{{ 10.5 / 2 + 1.25 }}

comparison with expressions
{{ 2 + 3 == 5 }}
{{ 2 * 3 == 7 }}
{{ 10 / 2 > 4 }}
{{ 10 / 2 >= 5 }}
{{ 2 ^ 3 < 10 }}
{{ 2 ^ 3 <= 8 }}
{{ 2 + 2 != 5 }}
{{ (2 + 3) * 2 == 10 }}

logical with arithmetic
{{ 2 + 2 == 4 && 3 * 3 == 9 }}
{{ 2 + 2 == 5 || 3 * 3 == 9 }}
{{ !(2 + 2 == 5) }}
{{ not(10 / 2 == 6) }}
{{ (5 > 3) && (10 / 2 == 5) }}
{{ (2 ^ 3 == 8) || (3 ^ 2 == 10) }}

chained comparisons
{{ 1 < 2 == true }}
{{ 5 > 3 == true }}
{{ 10 >= 10 == true }}
{{ 5 <= 5 == true }}

expression with variables
{{ simple.number + 8 }}
{{ simple.number * 2 }}
{{ simple.number / 6 }}
{{ simple.number % 5 }}
{{ simple.number ^ 2 }}
{{ simple.number + simple.uint }}
{{ (simple.number - 2) * 2 }}
{{ simple.float * 2 }}
{{ simple.float + 1.5 }}

large number calculations
{{ 1000000 + 1 }}
{{ 999999 * 2 }}
{{ 1000000 / 1000 }}
{{ 123456789 % 1000 }}

zero edge cases
{{ 0 + 0 }}
{{ 0 * 100 }}
{{ 0 - 0 }}
{{ 100 * 0 }}
{{ 0 ^ 0 }}
{{ 0 + 5 - 5 }}

one as identity
{{ 1 * 100 }}
{{ 100 * 1 }}
{{ 100 / 1 }}
{{ 1 ^ 1000 }}
{{ 1000 ^ 1 }}

subtraction associativity (left-to-right)
{{ 100 - 50 - 25 }}
{{ 100 - (50 - 25) }}
{{ 10 - 5 - 3 - 1 }}

division associativity (left-to-right)
{{ 100 / 10 / 5 }}
{{ 100 / (10 / 5) }}
{{ 1000 / 10 / 10 / 10 }}

combined associativity tests
{{ 2 - 3 + 4 }}
{{ 2 + 3 - 4 }}
{{ 12 / 3 * 2 }}
{{ 12 * 3 / 2 }}
{{ 2 - 3 + 4 - 5 + 6 }}
{{ 24 / 4 * 2 / 3 * 6 }}