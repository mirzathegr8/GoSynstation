// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compMatrix

import "runtime"

import "math"
import "math/cmplx"

func Mag(x complex128) float64{

	return real(x)*real(x) + imag(x)*imag(x)

}

func MagMax(x, y complex128) complex128 {
	if Mag(x) > Mag(y) {
		return x
	}
	return y
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MagMin(x, y complex128) complex128 {
	if Mag(x) < Mag(y) {
		return x
	}
	return y
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func sum(a []complex128) (s complex128) {
	for _, v := range a {
		s += v
	}
	return
}

func product(a []complex128) complex128 {
	p := complex128(1)
	for _, v := range a {
		p *= v
	}
	return p
}

type box interface{}

func countBoxes(start, cap int) chan box {
	ints := make(chan box)
	go func() {
		for i := start; i < cap; i++ {
			ints <- i
		}
		close(ints)
	}()
	return ints
}

func parFor(inputs <-chan box, foo func(i box)) (wait func()) {
	n := runtime.GOMAXPROCS(0)
	block := make(chan bool, n)
	for j := 0; j < n; j++ {
		go func() {
			for {
				i, ok := <-inputs
				if !ok {
					break
				}
				foo(i)
			}
			block <- true
		}()
	}
	wait = func() {
		for i := 0; i < n; i++ {
			<-block
		}
	}
	return
}


func compHypot(a,b complex128) float64{
	return math.Sqrt(real(a*cmplx.Conj(a)+b*cmplx.Conj(b)))
}
