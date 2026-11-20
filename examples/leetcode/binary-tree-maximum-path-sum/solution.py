class Solution:
    def maxPathSum(self, root: TreeNode | None) -> int:
        res = root.val

        def dfs(node: TreeNode | None) -> int:
            nonlocal res

            if node is None:
                return 0

            left = max(0, dfs(node.left))
            right = max(0, dfs(node.right))

            res = max(res, node.val + left + right)
            return node.val + max(left, right)

        dfs(root)
        return res
