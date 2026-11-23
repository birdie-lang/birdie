from collections import defaultdict


class TrieNode:
    def __init__(self):
        self.chars = defaultdict(TrieNode)
        self.word = None


class Solution:
    def findWords(self, board: list[list[str]], words: list[str]) -> list[str]:
        root = TrieNode()
        res = []
        m, n = len(board), len(board[0])
        directions = [(-1, 0), (1, 0), (0, -1), (0, 1)]

        for word in words:
            node = root
            for char in word:
                node = node.chars[char]
            node.word = word

        def dfs(r: int, c: int, parent: TrieNode) -> None:
            char = board[r][c]
            node = parent.chars[char]

            if node.word is not None:
                res.append(node.word)
                node.word = None

            board[r][c] = "#"

            for dr, dc in directions:
                nr, nc = r + dr, c + dc

                if (
                    0 <= nr < m
                    and 0 <= nc < n
                    and board[nr][nc] in node.chars
                ):
                    dfs(nr, nc, node)

            board[r][c] = char

            if not node.chars and node.word is None:
                del parent.chars[char]

        for r in range(m):
            for c in range(n):
                if board[r][c] in root.chars:
                    dfs(r, c, root)

        return res
