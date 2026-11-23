"""Invoke a reference method or method sequence; assertions remain in Go."""

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


def index_tree(root):
    """Node selectors name existing nodes by their unique integer values."""
    nodes = {}
    pending = [root]
    while pending:
        node = pending.pop()
        if node is None:
            continue
        if type(node.val) is not int or node.val in nodes:
            raise ValueError("node adapters require unique integer values")
        nodes[node.val] = node
        pending.extend((node.left, node.right))
    return nodes


def main():
    if sys.version_info < (3, 10):
        raise RuntimeError("reference examples require Python 3.10 or newer")

    request = json.load(sys.stdin)
    args = request["args"]
    adapters = request.get("argAdapters", [])
    for index, adapter in enumerate(adapters):
        if adapter == "tree":
            args[index] = decode_tree(args[index])

    nodes = {}
    if "node" in adapters or request.get("resultAdapter") == "node":
        if not adapters or adapters[0] != "tree":
            raise ValueError("node adapters require a tree as the first argument")
        nodes = index_tree(args[0])
    for index, adapter in enumerate(adapters):
        if adapter == "node":
            value = args[index]
            if type(value) is not int or value not in nodes:
                raise ValueError(f"node selector not found: {value!r}")
            args[index] = nodes[value]

    # Solution output cannot corrupt the JSON response protocol.
    with contextlib.redirect_stdout(sys.stderr):
        namespace = runpy.run_path(sys.argv[1], init_globals={"TreeNode": TreeNode})
        calls = request.get("calls")
        if calls is not None:
            instance = namespace[request["entry"]](*args)
            result = [getattr(instance, call["method"])(*call["args"]) for call in calls]
        else:
            solution_type = namespace.get("Solution") or namespace.get("Codec")
            if solution_type is None:
                raise ValueError("reference requires a Solution or Codec class")
            result = getattr(solution_type(), request["entry"])(*args)
        if request.get("resultAdapter") == "tree":
            result = encode_tree(result)
        elif request.get("resultAdapter") == "node":
            if not isinstance(result, TreeNode) or nodes.get(result.val) is not result:
                raise ValueError("node result must refer to an original input node")
            result = result.val

    json.dump(result, sys.stdout, ensure_ascii=True, allow_nan=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
