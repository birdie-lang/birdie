# First language decisions

Proposed syntax, not yet implemented. Example: [Two Sum](../examples/leetcode/two-sum.bd).

## Functions

- Declaration: `fn name(param: Type) ReturnType:`
- Explicit parameter types; optional return type.
- Return type follows `)`, without an arrow.
- Blocks use a colon and indentation.

## Typing

- Static typing.
- Infer local variables, return types, and collection types from usage.
- Require annotations when inference is ambiguous.

## Lists and tuples

- `[T]`: growable list of elements of type `T`.
- `(A, B)`: fixed pair with a type for each position.
- `[(int, int)]`: list of integer pairs.
- Tuple unpacking requires parentheses: `(x, y) = point`.

## Maps

- `map()`: empty map; infer key and value types from usage.
- `key in m`: key membership.
- `m[key] = value`: insert or replace.
- `m[key]`: lookup; missing keys cause a runtime error.

## Loops

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
