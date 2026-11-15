class Solution:
    def isSubtree(self, root: TreeNode | None, subRoot: TreeNode | None) -> bool:
        if not root:
            return False

        return (
            sameTree(root, subRoot) or
            self.isSubtree(root.left, subRoot) or
            self.isSubtree(root.right, subRoot)
        )


def sameTree(p: TreeNode | None, q: TreeNode | None) -> bool:
    if not p and not q:
        return True

    if not p or not q or p.val != q.val:
        return False

    return sameTree(p.left, q.left) and sameTree(p.right, q.right)
