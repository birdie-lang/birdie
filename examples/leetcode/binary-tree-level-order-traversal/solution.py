from collections import deque


class Solution:
    def levelOrder(self, root: TreeNode | None) -> list[list[int]]:
        if root is None:
            return []

        res = []
        pending = deque([root])

        while pending:
            level_res = []

            for _ in range(len(pending)):
                node = pending.popleft()
                level_res.append(node.val)

                if node.left is not None:
                    pending.append(node.left)

                if node.right is not None:
                    pending.append(node.right)

            res.append(level_res)

        return res
