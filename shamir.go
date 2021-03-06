package polynomial

import "math/big"

// GenRandomShares generates a polynomial and n points
// The polynomial can be solved with k points
func GenRandomShares(n, k int, q *big.Int) (ps Points, p Poly) {
	if q.ProbablyPrime(100) == false {
		ps = nil
		p = nil
		return
	}
	size := q.BitLen()/8 + 1
	p = make([]*big.Int, k)
	for i := 0; i < k; i++ {
		coeff := RandomBigInt(size)
		coeff.Mod(coeff, q)
		p[i] = coeff
	}
	ps = make([]Point, n)
	for i := 0; i < n; i++ {
		r := RandomBigInt(size)
		r.Mod(r, q)
		var t Point
		t.x = r
		t.y = p.Eval(r, q)
		ps[i] = t
	}
	return
}
