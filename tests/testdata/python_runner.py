"""Invoke one reference method using JSON transport; assertions remain in Go."""

import contextlib
import json
import runpy
import sys
from collections import deque


class TreeNode:
    def __init__(self, val=0, left=None, right=None):
        self.val = val
        self.left = left
        self.right = right


def decode_tree(values):
    if not values or values[0] is None:
        return None

    root = TreeNode(values[0])
    parents = deque([root])
    index = 1
    while index < len(values):
        parent = parents.popleft()
        for field in ("left", "right"):
            if index == len(values):
                break
            value = values[index]
            index += 1
            if value is not None:
                child = TreeNode(value)
                setattr(parent, field, child)
                parents.append(child)
    return root


def encode_tree(root):
    values = []
    pending = deque([root])
    seen = set()
    while pending:
        node = pending.popleft()
        if node is None:
            values.append(None)
            continue
        if not isinstance(node, TreeNode) or id(node) in seen:
            raise ValueError("tree result contains an invalid, shared, or cyclic node")
        seen.add(id(node))
        values.append(node.val)
        pending.extend((node.left, node.right))
    while values and values[-1] is None:
        values.pop()
    return values


def main():
    if sys.version_info < (3, 10):
        raise RuntimeError("reference examples require Python 3.10 or newer")

    request = json.load(sys.stdin)
    args = request["args"]
    for index, adapter in enumerate(request.get("argAdapters", [])):
        if adapter == "tree":
            args[index] = decode_tree(args[index])

    # Solution output cannot corrupt the JSON response protocol.
    with contextlib.redirect_stdout(sys.stderr):
        namespace = runpy.run_path(sys.argv[1], init_globals={"TreeNode": TreeNode})
        result = getattr(namespace["Solution"](), request["entry"])(*args)
        if request.get("resultAdapter") == "tree":
            result = encode_tree(result)

    json.dump(result, sys.stdout, ensure_ascii=True, allow_nan=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
