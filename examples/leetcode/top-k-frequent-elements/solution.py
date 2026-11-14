from collections import Counter


class Solution:
    def topKFrequent(self, nums: list[int], k: int) -> list[int]:
        buckets = [[] for _ in range(len(nums) + 1)]
        freqs = Counter(nums)

        for num, freq in freqs.items():
            buckets[freq].append(num)

        res = []
        for bucket in reversed(buckets):
            res.extend(bucket)

            if len(res) >= k:
                return res[:k]

        return res
