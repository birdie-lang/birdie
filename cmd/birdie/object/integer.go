package object

import (
	"cmp"
	"math"
	"math/big"
	"math/bits"
	"strconv"
)

// Integer is immutable and uses large only for values outside the int64 range.
type Integer struct {
	small int64
	large *big.Int
}

// IntegerFromBig takes a snapshot of value.
func IntegerFromBig(value *big.Int) *Integer {
	integer := wrapBigInteger(value)
	if integer.large != nil {
		integer.large = new(big.Int).Set(value)
	}
	return integer
}

// wrapBigInteger uses value directly when it needs the large representation.
func wrapBigInteger(value *big.Int) *Integer {
	if value.IsInt64() {
		return &Integer{small: value.Int64()}
	}
	return &Integer{large: value}
}

func (x *Integer) Type() ObjectType { return INTEGER_OBJ }

func (x *Integer) Inspect() string {
	if x.large != nil {
		return x.large.String()
	}
	return strconv.FormatInt(x.small, 10)
}

func (x *Integer) AsInt64() (int64, bool) {
	return x.small, x.large == nil
}

func (x *Integer) IsZero() bool {
	return x.large == nil && x.small == 0
}

// asBig returns an operand for read-only use by arithmetic operations.
func (x *Integer) asBig() *big.Int {
	if x.large != nil {
		return x.large
	}
	return big.NewInt(x.small)
}

func (x *Integer) Add(y *Integer) *Integer {
	if x.large == nil && y.large == nil {
		a, b := x.small, y.small
		sum := a + b
		overflow := (b > 0 && sum < a) || (b < 0 && sum > a)
		if !overflow {
			return &Integer{small: sum}
		}
	}
	return wrapBigInteger(new(big.Int).Add(x.asBig(), y.asBig()))
}

func (x *Integer) Sub(y *Integer) *Integer {
	if x.large == nil && y.large == nil {
		a, b := x.small, y.small
		difference := a - b
		overflow := (b > 0 && difference > a) || (b < 0 && difference < a)
		if !overflow {
			return &Integer{small: difference}
		}
	}
	return wrapBigInteger(new(big.Int).Sub(x.asBig(), y.asBig()))
}

func (x *Integer) Mul(y *Integer) *Integer {
	if x.large == nil && y.large == nil {
		a, b := x.small, y.small
		high, magnitude := bits.Mul64(absUint64(a), absUint64(b))
		negative := (a < 0) != (b < 0)
		if high == 0 {
			if negative && magnitude <= 1<<63 {
				return &Integer{small: int64(-magnitude)}
			}
			if !negative && magnitude <= math.MaxInt64 {
				return &Integer{small: int64(magnitude)}
			}
		}
	}
	return wrapBigInteger(new(big.Int).Mul(x.asBig(), y.asBig()))
}

func absUint64(value int64) uint64 {
	if value < 0 {
		// The magnitude of MinInt64 fits in uint64.
		return -uint64(value)
	}
	return uint64(value)
}

func (x *Integer) Neg() *Integer {
	if x.large == nil && x.small != math.MinInt64 {
		return &Integer{small: -x.small}
	}
	return wrapBigInteger(new(big.Int).Neg(x.asBig()))
}

// Quo rounds toward zero and panics if y is zero.
func (x *Integer) Quo(y *Integer) *Integer {
	return x.divide(y, big.Trunc)
}

// FloorDiv rounds toward negative infinity and panics if y is zero.
func (x *Integer) FloorDiv(y *Integer) *Integer {
	return x.divide(y, big.Floor)
}

func (x *Integer) divide(y *Integer, mode big.RoundingMode) *Integer {
	if x.large == nil && y.large == nil {
		a, b := x.small, y.small
		if a != math.MinInt64 || b != -1 {
			quotient := a / b
			if mode == big.Floor && (a < 0) != (b < 0) && a%b != 0 {
				quotient--
			}
			return &Integer{small: quotient}
		}
	}
	quotient, _ := new(big.Int).Divide(x.asBig(), y.asBig(), nil, mode)
	return wrapBigInteger(quotient)
}

// Cmp returns the ordering of x and y as -1, 0, or 1.
func (x *Integer) Cmp(y *Integer) int {
	switch {
	case x.large == nil && y.large == nil:
		return cmp.Compare(x.small, y.small)
	case x.large != nil && y.large != nil:
		return x.large.Cmp(y.large)
	case x.large != nil:
		return x.large.Sign()
	default:
		return -y.large.Sign()
	}
}
