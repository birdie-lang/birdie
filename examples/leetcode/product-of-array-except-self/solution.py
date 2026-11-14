class Solution:
    def productExceptSelf(self, nums: list[int]) -> list[int]:
        res = [1] * len(nums)

        prod = 1
        for i, num in enumerate(nums):
            res[i] = prod
            prod *= num

        prod = 1
        for i in reversed(range(len(nums))):
            res[i] *= prod
            prod *= nums[i]

        return res
