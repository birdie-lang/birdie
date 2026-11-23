package object

import (
	"math/big"
	"testing"
)

func TestIntegerSnapshotsBigInput(t *testing.T) {
	input := new(big.Int).SetUint64(1 << 63)
	value := IntegerFromBig(input)
	input.SetInt64(1)
	if got := value.Inspect(); got != "9223372036854775808" {
		t.Fatalf("integer changed with its source: got %s", got)
	}
}
