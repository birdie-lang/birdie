from collections import defaultdict


class TrieNode:
    def __init__(self):
        self.chars = defaultdict(TrieNode)
        self.isEnd = False


class WordDictionary:
    def __init__(self):
        self.root = TrieNode()

    def addWord(self, word: str) -> None:
        node = self.root
        for char in word:
            node = node.chars[char]
        node.isEnd = True

    def search(self, word: str) -> bool:
        return matches(self.root, word, 0)


def matches(node: TrieNode, word: str, start: int) -> bool:
    for i in range(start, len(word)):
        char = word[i]

        if char == ".":
            return any(
                matches(child, word, i + 1)
                for child in node.chars.values()
            )

        child = node.chars.get(char)
        if child is None:
            return False
        node = child

    return node.isEnd
