class Solution:
    def encode(self, words: list[str]) -> str:
        parts = []
        for word in words:
            parts.extend([str(len(word)), "#", word])
        return "".join(parts)

    def decode(self, data: str) -> list[str]:
        words = []
        cursor = 0

        while cursor < len(data):
            separator = cursor
            while data[separator] != "#":
                separator += 1

            size = int(data[cursor:separator])
            start = separator + 1
            cursor = start + size
            words.append(data[start:cursor])

        return words
