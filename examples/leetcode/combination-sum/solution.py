class Solution:
    def combinationSum(self, candidates: list[int], target: int) -> list[list[int]]:
        res = []
        subset = []

        def dfs(total: int, start: int) -> None:
            if total >= target:
                if total == target:
                    res.append(subset.copy())
                return

            for i in range(start, len(candidates)):
                num = candidates[i]
                if total + num > target:
                    continue

                subset.append(num)
                dfs(total + num, i)
                subset.pop()

        dfs(0, 0)
        return res
