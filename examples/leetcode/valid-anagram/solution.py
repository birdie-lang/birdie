class Solution:
    def isAnagram(self, s: str, t: str) -> bool:
        if len(s) != len(t):
            return False

        freqs = [0] * 26
        a = ord('a')

        for c1, c2 in zip(s, t):
            freqs[ord(c1) - a] += 1
            freqs[ord(c2) - a] -= 1

        return all(v == 0 for v in freqs)
