class Solution:
    def isBalanced(self, root: TreeNode | None) -> bool:
        return height(root) is not None


def height(node: TreeNode | None) -> int | None:
    if node is None:
        return 0

    left = height(node.left)
    if left is None:
        return None

    right = height(node.right)
    if right is None:
        return None

    if abs(right - left) > 1:
        return None

    return 1 + max(left, right)
