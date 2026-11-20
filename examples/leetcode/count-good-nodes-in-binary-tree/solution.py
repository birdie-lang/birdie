class Solution:
    def goodNodes(self, root: TreeNode) -> int:
        return dfs(root, root.val)


def dfs(node: TreeNode | None, max_val: int) -> int:
    if node is None:
        return 0

    count = 0
    if node.val >= max_val:
        count = 1
        max_val = node.val

    return count + dfs(node.left, max_val) + dfs(node.right, max_val)
