from math import inf


class Solution:
    def isValidBST(self, root: TreeNode | None) -> bool:
        return dfs(root, -inf, inf)


def dfs(node: TreeNode | None, min_val: int | float, max_val: int | float) -> bool:
    if node is None:
        return True

    if not min_val < node.val < max_val:
        return False

    return (
        dfs(node.right, node.val, max_val)
        and dfs(node.left, min_val, node.val)
    )
