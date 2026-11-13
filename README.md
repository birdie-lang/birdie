# Birdie

Birdie is an experimental programming language for solving algorithmic problems, written in Go.

It takes inspiration from Python, with static typing, type inference, and built-in data structures commonly used in coding interviews and programming challenges.

The goal is readable code, not necessarily the shortest code.

### Example

_Proposed syntax — subject to change._

```python
fn two_sum(nums: [int], target: int):
    seen = map()

    for i, n in enumerate(nums):
        if target - n in seen:
            return [seen[target - n], i]
        seen[n] = i
```

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
