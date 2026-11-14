class Solution:
    def isValidSudoku(self, board: list[list[str]]) -> bool:
        for i in range(9):
            if not validGroup(board[i]):
                return False

            column = [board[r][i] for r in range(9)]
            if not validGroup(column):
                return False

        for sr in range(0, 9, 3):
            for sc in range(0, 9, 3):
                square = [
                    board[r][c]
                    for r in range(sr, sr + 3)
                    for c in range(sc, sc + 3)
                ]

                if not validGroup(square):
                    return False

        return True


def validGroup(cells: list[str]) -> bool:
    seen = set()

    for cell in cells:
        if cell == '.':
            continue

        if cell in seen:
            return False

        seen.add(cell)

    return True
