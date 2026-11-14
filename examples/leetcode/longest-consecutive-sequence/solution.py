class Solution:
    def longestConsecutive(self, nums: list[int]) -> int:
        longest_seq = 0
        nums_set = set(nums)

        for num in nums_set:
            if num - 1 in nums_set:
                continue

            curr_seq = 1
            while num + curr_seq in nums_set:
                curr_seq += 1

            longest_seq = max(longest_seq, curr_seq)

        return longest_seq
