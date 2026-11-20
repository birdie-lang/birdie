from collections.abc import Iterator


class Codec:
    def serialize(self, root: TreeNode | None) -> str:
        parts = []

        def write(node: TreeNode | None) -> None:
            if node is None:
                parts.append("|")
                return

            parts.append(str(node.val))
            write(node.left)
            write(node.right)

        write(root)
        return ",".join(parts)

    def deserialize(self, data: str) -> TreeNode | None:
        return self.readTree(iter(data.split(",")))

    def readTree(self, tokens: Iterator[str]) -> TreeNode | None:
        value = next(tokens)

        if value == "|":
            return None

        node = TreeNode(int(value))
        node.left = self.readTree(tokens)
        node.right = self.readTree(tokens)
        return node
