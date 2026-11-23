from collections import defaultdict


class TrieNode:
    def __init__(self):
        self.chars = defaultdict(TrieNode)
        self.isEnd = False


class Trie:
    def __init__(self):
        self.root = TrieNode()

    def insert(self, word: str) -> None:
        node = self.root
        for char in word:
            node = node.chars[char]
        node.isEnd = True

    def search(self, word: str) -> bool:
        node = self.findNode(word)
        return node is not None and node.isEnd

    def startsWith(self, prefix: str) -> bool:
        return self.findNode(prefix) is not None

    def findNode(self, word: str) -> TrieNode | None:
        node = self.root
        for char in word:
            node = node.chars.get(char)
            if node is None:
                return None
        return node
