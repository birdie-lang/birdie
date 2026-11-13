# First language decisions

Proposed syntax, not yet implemented.

Examples: [Two Sum](../examples/leetcode/two-sum.bd),
[Valid Parentheses](../examples/leetcode/valid-parentheses.bd),
[Binary Search](../examples/leetcode/binary-search.bd).

## Functions

- Declaration: `fn name(param: Type) ReturnType:`
- Explicit parameter types; optional return type.
- Return type follows `)`, without an arrow.
- Blocks use a colon and indentation.

## Typing

- Static typing.
- Infer local variables, return types, and collection types from usage.
- Require annotations when inference is ambiguous.

## Variables

- `=` introduces a variable when the name is not already defined; otherwise it reassigns it.
- Reassignments must be compatible with the variable's inferred type.
- Local variables require an initial value; no declaration-only form.
- Every read requires an assignment on every execution path reaching it.
- Assignment inside a block updates an existing enclosing local variable.

## Integers and division

- `int`: signed, arbitrary precision; no fixed bit width or wrapping overflow.
- Value size is limited by available resources; operation costs depend on operand size.
- `//`: floor division; integer operands produce an integer.
- Floor rounds toward negative infinity: `7 // 2` is `3`; `-7 // 2` is `-4`.
- `/`: reserved for fractional division.
- Division by zero causes a runtime error.

## Booleans and truthiness

- `bool`: `true` or `false`.
- Conditions and `not` accept booleans, lists, maps, and stacks.
- Empty collections are false; nonempty collections are true.
- `not` returns a boolean: `not opened` tests whether the stack is empty.
- Numbers require explicit comparisons, such as `count != 0`.
- `or` evaluates left to right and short-circuits when its left operand is true.

## Strings and characters

- `str`: string, written with double quotes: `"text"`, `"a"`.
- `char`: one Unicode scalar value, written with single quotes: `'a'`.
- Empty or multi-scalar character literals are errors: `''`, `'ab'`.
- A visible symbol can contain multiple Unicode scalar values.
- String iteration yields `char` values.
- String length and indices count Unicode scalar values, not bytes.
- String indexing returns `char`.

## Lists and tuples

- `[T]`: growable list of elements of type `T`.
- `(A, B)`: fixed pair with a type for each position.
- `[(int, int)]`: list of integer pairs.
- Tuple unpacking requires parentheses: `(x, y) = point`.

## Size and indexing

- `.size`: read-only `int`; element count for lists and stacks, entry count for maps,
  Unicode scalar count for strings.
- List and string indices are zero-based: `0 <= i < value.size`.
- Negative and out-of-bounds indices cause a runtime error.
- `list.last`: final element; an empty list causes a runtime error.

## Maps

- `map()`: empty map; infer key and value types from usage.
- `{'a': 1, 'b': 2}`: populated map literal; same map type as the constructor.
- `key in m`: key membership.
- `m[key] = value`: insert or replace.
- `m[key]`: lookup; missing keys cause a runtime error.

## Stacks

- `stack()`: empty stack; infer the element type from usage.
- Separate type with LIFO access; no indexing, arbitrary insertion, or sorting.
- `s.push(value)`: add to the top.
- `s.pop()`: remove and return the top; an empty stack causes a runtime error.
- Use truthiness for emptiness checks: `while s`, `if not s`.
- Lists can also support `push` and `pop` at the end.

## Loops

- `while condition:` tests the condition before each iteration and stops when false.

```text
for num in nums:                  # value
for i, num in nums:               # zero-based index and value
for (x, y) in points:             # unpack a tuple
for i, (x, y) in points:          # index and unpacked tuple
```

Map iteration proposal: `for key, value in counts:`.

## Open questions

- Single-binding map iteration.
- Map iteration order.
- Counter + key + value syntax.
- Conversion and concatenation between `char` and `str`.
- Visibility outside a block of variables first introduced inside it.
- Floating-point arithmetic.
- Remainder semantics and explicit truncating division.
