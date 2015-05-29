package polynomial

// 공개된 다항식 라이브러리는 대부분 실수를 계수로 이용하는 것이다.
// 이 라이브러리는 큰 정수(Big Integer)를 계수로 사용하는 다항식 연산을 위한 것이다.
// 따라서 부동소수점 계수나 분수로 표현된 계수는 연산하지 못한다.
// 큰 정수는 보통 암호 관련 기술에 사용되며 대부분 Cyclic group에서 동작하도록 설정되어
// modulo 연산을 함께 수행해주어야 한다.
// 본 라이브러리는 다항식의 덧셈, 뺄셈, 곱셈, 나눗셈(나머지), 최대공약다향식을 구한다.
import (
	"fmt"
	"math/big"
	"math/rand"
	"time"
)

// 다항식 계수가 sparsely 배치될 수도 있어 효율적인 자료구조를 설정할 수 있지만
// 1) 구현 편의성
// 2) 곱셈이나 나눗셈을 하면 점점 dense해지는 것
// 을 이유로 배열 형태로 계수를 저장한다.
// 다항식은 계수를 역순으로 저장한다. y = 3x^3 + 2x + 1이라면
// [1 2 0 3] 형식으로 저장한다.
type Poly []*big.Int

// Golang에서 큰 정수를 만드는 것은 다소 귀찮은 작업이므로,
// 편의를 위해 정수를 나열해주면 다항식을 생성해주는 헬퍼이다.
func NewPolyInts(coeffs ...int) (p Poly) {
	if coeffs == nil {
		return Poly{big.NewInt(0)}
	}
	p = make([]*big.Int, len(coeffs))
	for i := 0; i < len(coeffs); i++ {
		p[i] = big.NewInt(int64(coeffs[i]))
	}
	p.trim()
	return
}

// 주어진 차수(degree)의 임의 다항식을 만든다.
// 계수의 크기는 [0, 2^bits)의 임의 숫자.
func RandomPoly(degree, bits int64) (p Poly) {
	p = make(Poly, degree+1)
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	exp := big.NewInt(2)
	exp.Exp(exp, big.NewInt(bits), nil)
	for i := 0; i <= p.GetDegree(); i++ {
		p[i] = new(big.Int)
		p[i].Rand(rr, exp)
	}
	p.trim()
	return
}

// trim()은 다항식의 최고차 항의 계수가 0이 되지 않도록 조정한다.
// 최고차 항의 계수가 0인 다항식은 있을 수 없기 때문에 항상 내부적으로만 사용된다.
// 덧셈, 뺄셈 등을 수행하다보면 최고차 항이 소거되는 경우가 발생하고
// 이 때 계수 0이 남아 있어 degree가 잘못 계산되는 것을 방지하기 위해 사용한다.
func (p *Poly) trim() {
	var last int = 0
	for i := p.GetDegree(); i > 0; i-- { // i > 0 인 이유는 상수항은 제거되지 않기 때문.
		if (*p)[i].Sign() != 0 { // 역으로 검색하면서 0이 아닌 최고차 항을 찾는다.
			last = i
			break
		}
	}
	*p = (*p)[:(last + 1)]
}

// isZero() 함수는 현재 다항식 P = 0 인지 점검하는 함수.
func (p *Poly) isZero() bool {
	if p.GetDegree() == 0 && (*p)[0].Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

// 최고 차수를 반환하는 함수. 3차 방정식이면 3이 반환된다.
func (p Poly) GetDegree() int {
	return len(p) - 1
}

func (p Poly) String() (s string) {
	s = "["
	for i := len(p) - 1; i >= 0; i-- {
		switch p[i].Sign() {
		case -1:
			if i == len(p)-1 {
				s += "-"
			} else {
				s += " - "
			}
			if i == 0 || p[i].Int64() != -1 {
				s += p[i].String()[1:]
			}
		case 0:
			continue
		case 1:
			if i < len(p)-1 {
				s += " + "
			}
			if i == 0 || p[i].Int64() != 1 {
				s += p[i].String()
			}
		}
		if i > 0 {
			s += "x"
			if i > 1 {
				s += "^" + fmt.Sprintf("%d", i)
			}
		}
	}
	if s == "[" {
		s += "0"
	}
	s += "]"
	return
}

// Compare 함수는 두 개의 다항식을 비교한다.
// 현 다항식을 복사할 필요는 없으므로 포인터로 받으며,
// 비교 대상 다항식도 효율성을 위해 포인터로 받는다.
// 두 디항식이 동일하면 0,
// 인자로 넘겨준 다항식이 더 크면 1,
// 그렇지 않으면 -1을 반환한다.
// 알고리즘은 간단하다.
// 차수가 크면 무조건 큰 다항식이며, 차수가 같을 시에는 계수값을 비교한다.
func (p *Poly) Compare(q *Poly) int {
	switch {
	case p.GetDegree() > q.GetDegree():
		return 1
	case p.GetDegree() < q.GetDegree():
		return -1
	}
	for i := 0; i <= p.GetDegree(); i++ {
		switch (*p)[i].Cmp((*q)[i]) {
		case 1:
			return 1
		case -1:
			return -1
		}
	}
	return 0
}

// Add()는 두 다항식을 더하는 함수다.
// 추가 인자로는 modulo를 줄 수 있으며,
// modulo 연산을 하고 싶지 않을 경우에는 nil을 주면 된다.
func (p Poly) Add(q Poly, m *big.Int) Poly {
	if p.Compare(&q) < 0 {
		return q.Add(p, m)
	}
	var r Poly = make([]*big.Int, len(p))
	for i := 0; i < len(q); i++ {
		a := new(big.Int)
		a.Add(p[i], q[i])
		r[i] = a
	}
	for i := len(q); i < len(p); i++ {
		a := new(big.Int)
		a.Set(p[i])
		r[i] = a
	}
	if m != nil {
		for i := 0; i < len(p); i++ {
			r[i].Mod(r[i], m)
		}
	}
	r.trim()
	return r
}

// 주어진 다항식의 계수에 모두 -1를 곱한다.
func (p *Poly) NegSelf() {
	for i := 0; i < len(*p); i++ {
		(*p)[i].Neg((*p)[i])
	}
}

// 주어진 다항식예 계수에 모두 마이너스를 취한 다항식을 새로 만들어서 반환한다.
func (p *Poly) Neg() Poly {
	var q Poly = make([]*big.Int, len(*p))
	for i := 0; i < len(*p); i++ {
		b := new(big.Int)
		b.Neg((*p)[i])
		q[i] = b
	}
	return q
}

// Clone()은 주어진 다항식을 deep copy하여 새로운 다항식을 만들어주는 함수.
// 인자로 주어지는 adjust 정수값은 복사를 하면서 차수 변경을 할 때 이용한다.
// adjust는 음수값을 가질 수 없으며 이 경우에는 다항식 0를 반환한다.
// adjust값만큼 차수가 높아진 상태로 반환된다.
// 예를 들어, x + 1 다항식에 adjust값을 2를 주면 x^3 + x^2가 반환된다.
// 동일한 다항식을 복사하고 싶으면 adjust에 0을 넘겨주면 된다.
func (p Poly) Clone(adjust int) Poly {
	var q Poly = make([]*big.Int, len(p)+adjust)
	if adjust < 0 {
		return NewPolyInts(0)
	}
	for i := 0; i < adjust; i++ {
		q[i] = big.NewInt(0)
	}
	for i := adjust; i < len(p)+adjust; i++ {
		b := new(big.Int)
		b.Set(p[i-adjust])
		q[i] = b
	}
	return q
}

// sanitize() 함수는 주어진 modulo 값을 이용하여,
// 현재 다항식의 계수 전체에 modulo 연산을 적용한다.
func (p *Poly) sanitize(m *big.Int) {
	if m == nil {
		return
	}
	for i := 0; i <= (*p).GetDegree(); i++ {
		(*p)[i].Mod((*p)[i], m)
	}
}

// 두 다항식을 빼는 함수. 미리 만들어둔 Add 함수를 활용하기 위해 A + (-B)로 계산한다.
func (p Poly) Sub(q Poly, m *big.Int) Poly {
	r := q.Neg()
	return p.Add(r, m)
}

// 두 다항식을 곱하는 함수.
func (p Poly) Mul(q Poly, m *big.Int) Poly {
	if m != nil {
		p.sanitize(m)
		q.sanitize(m)
	}
	var r Poly = make([]*big.Int, p.GetDegree()+q.GetDegree()+1)
	for i := 0; i < len(r); i++ {
		r[i] = big.NewInt(0)
	}
	for i := 0; i < len(p); i++ {
		for j := 0; j < len(q); j++ {
			a := new(big.Int)
			a.Mul(p[i], q[j])
			a.Add(a, r[i+j])
			if m != nil {
				a.Mod(a, m)
			}
			r[i+j] = a
		}
	}
	r.trim()
	return r
}

//	현 다항식을 주어진 다항식으로 나누고 몫과 나머지를 반환하는 함수.
//	modulo값을 줄 수 있고, 원하지 않을 경우 nil을 주면 된다.
//	계수를 정수만 사용할 수 있어서 계수가 정확히 정수로 나눠지지 않을 경우는
//	나눗셈을 수행할 수 없다. 따라서 이 경우에는 아래의 설명과 같이 나눗셈을 수행하지 않는다.
func (p Poly) Div(q Poly, m *big.Int) (quo, rem Poly) {
	if m != nil {
		p.sanitize(m)
		q.sanitize(m)
	}
	if p.GetDegree() < q.GetDegree() || q.isZero() {
		quo = NewPolyInts(0)
		rem = p.Clone(0)
		return
	}
	quo = make([]*big.Int, p.GetDegree()-q.GetDegree()+1)
	rem = p.Clone(0)
	for i := 0; i < len(quo); i++ {
		quo[i] = big.NewInt(0)
	}
	t := p.Clone(0)
	qd := q.GetDegree()
	for {
		td := t.GetDegree()
		rd := td - qd
		if rd < 0 || t.isZero() {
			rem = t
			break
		}
		r := new(big.Int)
		if m != nil {
			r.ModInverse(q[qd], m)
			r.Mul(r, t[td])
			r.Mod(r, m)
		} else {
			r.Div(t[td], q[qd])
		}
		// r의 값이 0이 된다는 것은 (modulo 연산을 하지 않을 때) 최고차 항이 배수 관계
		// 아닌 경우다. 이 경우에는 결과가 실수(분수)로 나오게 되는데, 본 다항식 라이브러리
		// 암호화를 위한 BigInt 다항식 계산을 위한 것으로 실수 결과가 필요 없다.
		// 따라서 처리하지 않고 몫은 0, 나머지는 나누려했던 값으로 반환한다.
		if r.Cmp(big.NewInt(0)) == 0 {
			quo = NewPolyInts(0)
			rem = p.Clone(0)
			return
		}
		u := q.Clone(rd)
		for i := rd; i < len(u); i++ { // rd 밑으로는 어차피 모두 0므로 곱셈을 할 필요 없음
			u[i].Mul(u[i], r)
			if m != nil {
				u[i].Mod(u[i], m)
			}
		}
		t = t.Sub(u, m)
		t.trim()
		quo[rd] = r
	}
	quo.trim()
	rem.trim()
	return
}

// 유클리드 알고리즘을 이용하여 최대공약 다항식을 계산하는 함수.
// 다항식 나눗셈, 나머지 연산이 구현되어 있으므로 그것을 활용
func (p Poly) Gcd(q Poly, m *big.Int) Poly {
	// fmt.Println("p:", p, ", q:", q)
	if p.Compare(&q) < 0 {
		return q.Gcd(p, m)
	}
	if q.isZero() {
		// fmt.Println("Found:", p)
		return p
	} else {
		_, rem := p.Div(q, m)
		// fmt.Println("rem:", rem)
		return q.Gcd(rem, m)
	}
}

// Eval()은 주어진 함수 p(x)에 x값을 넣었을 때 어떤 값이 나오는지 계산하는 함수.
// modulo값 m을 줄 수 있다.
func (p Poly) Eval(x *big.Int, m *big.Int) (y *big.Int) {
	y = big.NewInt(0)
	accx := big.NewInt(1)
	xd := new(big.Int)
	for i := 0; i <= p.GetDegree(); i++ {
		xd.Mul(accx, p[i])
		y.Add(y, xd)
		accx.Mul(accx, x)
		if m != nil {
			y.Mod(y, m)
			accx.Mod(accx, m)
		}
	}
	return y
}
