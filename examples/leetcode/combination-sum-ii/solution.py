class Solution:
    def combinationSum2(self, candidates: list[int], target: int) -> list[list[int]]:
        candidates.sort()
        res = []
        subset = []

        def dfs(total: int, start: int) -> None:
            if total == target:
                res.append(subset.copy())
                return

            for i in range(start, len(candidates)):
                num = candidates[i]
                if total + num > target:
                    return
                if i > start and num == candidates[i - 1]:
                    continue

                subset.append(num)
                dfs(total + num, i + 1)
                subset.pop()

        dfs(0, 0)
        return res
