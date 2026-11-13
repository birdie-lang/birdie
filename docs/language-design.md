# First language decisions

Proposed syntax, not yet implemented.

Examples: [Two Sum](../examples/leetcode/two-sum.bd),
[Valid Parentheses](../examples/leetcode/valid-parentheses.bd),
[Binary Search](../examples/leetcode/binary-search.bd),
[Group Anagrams](../examples/leetcode/group-anagrams.bd).

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
- `int()`: zero.
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
- Strings are immutable; indexed assignment is a type error.
- String equality and hashing use the Unicode scalar sequence as stored.
- `char`: one Unicode scalar value, written with single quotes: `'a'`.
- Empty or multi-scalar character literals are errors: `''`, `'ab'`.
- A visible symbol can contain multiple Unicode scalar values.
- String iteration yields `char` values.
- String length and indices count Unicode scalar values, not bytes.
- String indexing returns `char`.
- `s.sorted()`: new `str` with characters sorted by Unicode scalar value;
  no locale-dependent ordering or mutation of the original string.

## Lists

- `list[T]`: growable list of elements of type `T`.
- `list[list[str]]`: list of lists of strings.
- `list()`: new empty list.
- List literals: `[]`, `[1, 2, 3]`.
- Use `push` for adding one element; no `append` alias.

### Operations

For `items: list[T]`, `value` must have type `T` and `i` must have type `int`.

| Operation | Returns | Mutates | Behavior |
| --- | --- | --- | --- |
| `items.push(value)` | No value | Yes | Add one element at the end. |
| `items.pop()` | `T` | Yes | Remove and return the last element; runtime error if empty. |
| `items[i]` | `T` | No | Read an element; requires a valid index. |
| `items[i] = value` | No value | Yes | Replace an element; requires a valid index. |
| `items.last` | `T` | No | Read the last element; runtime error if empty. |
| `items.size` | `int` | No | Number of elements. |
| `items.sorted()` | `list[T]` | No | Allocate a new list in ascending order; `T` must have a defined ordering. |

## Tuples

- `(A, B)`: fixed pair with a type for each position.
- `list[(int, int)]`: list of integer pairs.
- Tuple unpacking requires parentheses: `(x, y) = point`.

## Size and indexing

- `.size`: read-only `int`; element count for lists and stacks, entry count for maps,
  Unicode scalar count for strings.
- List and string indices are zero-based: `0 <= i < value.size`.
- Negative and out-of-bounds indices cause a runtime error.

## Maps

- `map[K, V]`: map from keys of type `K` to values of type `V`.
- `map()`: empty map with no default; infer key and value types from usage.
- `{'a': 1, 'b': 2}`: populated map literal; same map type as the constructor.

### Operations

For `m: map[K, V]`, `key` must have type `K`; `value` and `fallback` must have type `V`.

| Operation | Returns | Mutates | Behavior |
| --- | --- | --- | --- |
| `m[key]` | `V` | On a miss with a default | Return the stored value; on a miss, insert the configured default or raise a runtime error if there is none. |
| `m[key] = value` | No value | Yes | Insert or replace the entry; never call the default factory. |
| `key in m` | `bool` | No | Check stored keys; never call the default factory. |
| `m.get(key, fallback)` | `V` | No | Return the stored value or the supplied fallback; never insert or call the default factory. |
| `m.size` | `int` | No | Number of stored entries. |
| `m.values()` | `list[V]` | No | Allocate a snapshot of stored values in map iteration order; never call the default factory. |

`get` requires an explicit fallback, even on maps with configured defaults.
An existing `0`, `false`, or empty collection is returned as stored.

`values()` copies only the outer container. Later map entry changes do not update
the returned list; nested mutable collections remain shared.

### Defaults

- `map(value)`: non-callable immutable default value, evaluated at construction.
- `map(factory)`: callable default factory; must accept no arguments.
- Callable arguments always mean factories. A function-valued default requires a
  factory that returns that function.
- Argument types determine whether the constructor receives a value or a factory.
- The map's value type is inferred from the default value or factory return type.
- On a missing-key lookup, obtain the default, insert it under the key, and return it.
- Call factories only for missing keys; store and return their results without copying.
- Existing values, including `0`, `false`, and empty collections, are returned unchanged.
- Mutable value defaults such as `map([])` or `map(list())` are type errors.

```text
map(0)       # default 0
map(int)     # factory returns 0
map(1)       # default 1
map(list)    # fresh empty list per missing key
map(map)     # fresh map with no default per missing key
map(stack)   # fresh empty stack per missing key
```

Custom factories can construct nested collections. Each missing-key lookup calls
the factory independently; any sharing of returned values is controlled by the factory.

## Stacks

- `stack[T]`: stack of elements of type `T`.
- `stack()`: empty stack; infer the element type from usage.
- Separate type with LIFO access; no indexing, arbitrary insertion, or sorting.
- Use truthiness for emptiness checks: `while s`, `if not s`.

### Operations

For `s: stack[T]`, `value` must have type `T`.

| Operation | Returns | Mutates | Behavior |
| --- | --- | --- | --- |
| `s.push(value)` | No value | Yes | Add one element to the top. |
| `s.pop()` | `T` | Yes | Remove and return the top element; runtime error if empty. |
| `s.peek()` | `T` | No | Read the top element; runtime error if empty. |
| `s.size` | `int` | No | Number of elements. |

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
