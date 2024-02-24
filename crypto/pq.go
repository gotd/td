package crypto

import (
	"crypto/rand"
	"io"
	"math/big"
)

// This piece of software can be found in multiple projects, referenced as "splitPQ":
//	* https://github.com/sdidyk/mtproto/blob/cf2cb57ade6932a7e0854f2e3246492a2028d369/math.go#L44
//	* https://github.com/cjongseok/mtproto/blob/dab4e1a6538c990d8a4450d700559197c344a63b/crypto.go#L48
//	* https://gist.github.com/asaelko/6f27c8ccd2edfab4d1efadf80a5221a4
//	* https://github.com/xelaj/mtproto/blob/13ca2c1b5bfaa249ec925921760be0dc05735915/math.go#L39
//
// Also there is StackOverflow question about this part of telegram auth:
//	* https://stackoverflow.com/questions/31953836/decompose-a-number-into-2-prime-co-factors
//
// General information can be found in telegram auth docs:
//	* https://core.telegram.org/mtproto/auth_key#dh-exchange-initiation
//	* https://core.telegram.org/mtproto/samples-auth_key#4-encrypted-data-generation
//
// Looks like this is some kind of integer factorization algorithm implementation
// which is used as Proof of Work by Telegram Server for some reason.
//
// TODO(ernado): Try https://en.wikipedia.org/wiki/Pollard%27s_rho_algorithm

// DecomposePQ decomposes pq into prime factors such that p < q.
func DecomposePQ(pq *big.Int, randSource io.Reader) (p, q *big.Int, err error) { // nolint:gocognit
	var (
		value0  = big.NewInt(0)
		value1  = big.NewInt(1)
		value15 = big.NewInt(15)
		value17 = big.NewInt(17)
		rndMax  = big.NewInt(0).SetBit(big.NewInt(0), 64, 1)

		y        = big.NewInt(0)
		whatNext = big.NewInt(0)

		a = big.NewInt(0)
		b = big.NewInt(0)
		c = big.NewInt(0)

		b2 = big.NewInt(0)

		z = big.NewInt(0)
	)

	what := big.NewInt(0).Set(pq)
	g := big.NewInt(0)
	i := 0
	for !(g.Cmp(value1) == 1 && g.Cmp(what) == -1) {
		v, err := rand.Int(randSource, rndMax)
		if err != nil {
			return nil, nil, err
		}
		v = v.And(v, value15)
		v = v.Add(v, value17)
		v = v.Mod(v, what)

		x, err := rand.Int(randSource, rndMax)
		if err != nil {
			return nil, nil, err
		}
		whatNext.Sub(what, value1)
		x = x.Mod(x, whatNext)
		x = x.Add(x, value1)

		y.Set(x)
		lim := 1 << (uint(i) + 18)
		j := 1
		flag := true

		for j < lim && flag {
			a.Set(x)
			b.Set(x)
			c.Set(v)

			for b.Cmp(value0) == 1 {
				b2.SetInt64(0)
				if b2.And(b, value1).Cmp(value0) == 1 {
					c.Add(c, a)
					if c.Cmp(what) >= 0 {
						c.Sub(c, what)
					}
				}
				a.Add(a, a)
				if a.Cmp(what) >= 0 {
					a.Sub(a, what)
				}
				b.Rsh(b, 1)
			}
			x.Set(c)

			z.SetInt64(0)
			if x.Cmp(y) == -1 {
				z.Add(what, x)
				z.Sub(z, y)
			} else {
				z.Sub(x, y)
			}
			g.GCD(nil, nil, z, what)

			if (j & (j - 1)) == 0 {
				y.Set(x)
			}
			j++

			if g.Cmp(value1) != 0 {
				flag = false
			}
		}
		i++
	}

	p = big.NewInt(0).Set(g)
	q = big.NewInt(0).Div(what, g)

	if p.Cmp(q) == 1 {
		p, q = q, p
	}

	return p, q, nil
}
