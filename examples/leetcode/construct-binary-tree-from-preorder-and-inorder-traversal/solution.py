class Solution:
    def buildTree(self, preorder: list[int], inorder: list[int]) -> TreeNode | None:
        if not preorder or not inorder:
            return None

        value = preorder[0]
        mid = inorder.index(value)

        return TreeNode(
            value,
            self.buildTree(preorder[1:mid + 1], inorder[:mid]),
            self.buildTree(preorder[mid + 1:], inorder[mid + 1:]),
        )
