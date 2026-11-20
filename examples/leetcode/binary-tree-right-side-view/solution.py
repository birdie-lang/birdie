from collections import deque


class Solution:
    def rightSideView(self, root: TreeNode | None) -> list[int]:
        if root is None:
            return []

        res = []
        pending = deque([root])

        while pending:
            for _ in range(len(pending)):
                node = pending.popleft()

                if node.left is not None:
                    pending.append(node.left)
                if node.right is not None:
                    pending.append(node.right)

            res.append(node.val)

        return res
