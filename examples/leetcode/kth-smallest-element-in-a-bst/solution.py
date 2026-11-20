class Solution:
    def kthSmallest(self, root: TreeNode | None, k: int) -> int:
        def dfs(node: TreeNode | None) -> int | None:
            nonlocal k

            if node is None:
                return None

            left = dfs(node.left)
            if left is not None:
                return left

            k -= 1
            if k == 0:
                return node.val

            return dfs(node.right)

        return dfs(root)
