// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compMatrix

import "math"
import "math/cmplx"

/*
Swap two rows in this matrix.
*/
func (A *SparseMatrix) SwapRows(r1, r2 int) {
	js := map[int]bool{}
	for index := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i == r1 || i == r2 {
			js[j] = true
		}
	}
	for j := range js {
		tmp := A.Get(r1, j)
		A.Set(r1, j, A.Get(r2, j))
		A.Set(r2, j, tmp)
	}
}

/*
Scale a row by a scalar.
*/
func (A *SparseMatrix) ScaleRow(r int, f complex128) {
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i == r {
			A.Set(i, j, value*f)
		}
	}
}

/*
Add a multiple of row rs to row rd.
*/
func (A *SparseMatrix) ScaleAddRow(rd, rs int, f complex128) {
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i == rs {
			A.Set(rd, j, A.Get(rd, j)+value*f)
		}
	}
}

func (A *SparseMatrix) Symmetric() bool {
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i != j && value != A.Get(j, i) {
			return false
		}
	}
	return true
}

func (A *SparseMatrix) Transpose() *SparseMatrix {
	B := ZerosSparse(A.cols, A.rows)
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		B.Set(j, i, value)
	}
	return B
}

func (A *SparseMatrix) Det() complex128 {
	//TODO: obviously this is a horrible way to do it
	return A.DenseMatrix().Det()
}

func (A *SparseMatrix) Trace() (res complex128) {
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index)
		if i == j {
			res += value
		}
	}
	return
}

func (A *SparseMatrix) OneNorm() (res float64) {
	for _, value := range A.elements {
		res += cmplx.Abs(value)
	}
	return
}

func (A *SparseMatrix) TwoNorm() complex128 {
	var sum complex128 = 0
	for _, value := range A.elements {
		sum += value * value
	}
	return cmplx.Sqrt(sum)
}

func (A *SparseMatrix) InfinityNorm() (res float64) {
	for _, value := range A.elements {
		res = math.Max(res, cmplx.Abs(value))
	}
	return
}