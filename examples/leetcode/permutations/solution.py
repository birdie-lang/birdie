class Solution:
    def permute(self, nums: list[int]) -> list[list[int]]:
        res = []
        perm = []
        used = [False] * len(nums)

        def dfs() -> None:
            if len(perm) == len(nums):
                res.append(perm.copy())
                return

            for i, num in enumerate(nums):
                if used[i]:
                    continue

                used[i] = True
                perm.append(num)
                dfs()
                perm.pop()
                used[i] = False

        dfs()
        return res
