# Birdie

Birdie is an experimental programming language for solving algorithmic problems, written in Go.

It takes inspiration from Python, with static typing, type inference, and built-in data structures commonly used in coding interviews and programming challenges.

The goal is readable code, not necessarily the shortest code.

### Example

_Proposed syntax — subject to change._

```python
fn twoSum(nums: list[int], target: int) list[int]:
    seen = map()

    for i, num in nums:
        diff = target - num

        if diff in seen:
            return [seen[diff], i]

        seen[num] = i

    return []
```

Read the [philosophy](https://birdie.jakmaz.com/docs/philosophy/) for the design principles, the [guide](https://birdie.jakmaz.com/docs/) for an introduction, or the [draft language specification](https://birdie.jakmaz.com/docs/specification/) for syntax and behavior.
Documentation is maintained in the [website repository](https://github.com/birdie-lang/website).

### Design

- Python-like syntax with minimal boilerplate
- Static typing with type inference
- First-class stacks, heaps, queues, and other collections
- No imports for standard functionality
- Deterministic operation counting for comparing algorithms
- Built-in testing and a REPL
- Domain specific tooling

### Status

Early development. Birdie is not yet ready for general use.

The first implementation is an AST interpreter in Go, with a bytecode VM planned for later.
