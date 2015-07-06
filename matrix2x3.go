package main

import ()

type Matrix2x3 struct {
	a float32
	b float32
	c float32
	d float32
	e float32
	f float32
}

func (m2 *Matrix2x3) Concat(m1 Matrix2x3) Matrix2x3 {
	return Matrix2x3{m1.a*m2.a + m1.b*m2.c,
		m1.a*m2.b + m1.b*m2.d,
		m1.c*m2.a + m1.d*m2.c,
		m1.c*m2.b + m1.d*m2.d,
		m1.e*m2.a + m1.f*m2.c + m2.e,
		m1.e*m2.b + m1.f*m2.d + m2.f}
}

func (m *Matrix2x3) Translate(x float32, y float32) Matrix2x3 {
	return Matrix2x3{m.a, m.b, m.c, m.d,
		m.e + (x*m.a + y*m.c),
		m.f + (x*m.b + y*m.d)}
}

func (m *Matrix2x3) Scale(x float32, y float32) Matrix2x3 {
	return Matrix2x3{m.a * x,
		m.b * x,
		m.c * y,
		m.d * y,
		m.e,
		m.f}
}

func (m *Matrix2x3) Array() [16]float32 {
	return [...]float32{
		m.a, m.b, 0.0, 0.0,
		m.c, m.d, 0.0, 0.0,
		0.0, 0.0, 1.0, 0.0,
		m.e, m.f, 0.0, 1.0,
	}
}

func IdentityMatrix2x3() Matrix2x3 {
	return Matrix2x3{1, 0, 0, 1, 0, 0}
}
