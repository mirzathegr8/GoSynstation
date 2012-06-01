package main

import "fmt"
import "math/rand"
//import "math/cmplx"
//import "math"
import "../compMatrix"



func main() {

/********************************************************/

/********************************************************/
	N := 2
	M := 3
	A := compMatrix.Zeros(N, M)
	B := compMatrix.Zeros(M, N)

	//fmt.Println("Complex Matrix:\n", H, "\n")
	
	for i := 0; i < N; i++ {
		for j := 0; j < M; j++ {
			A.Set(i, j, complex(rand.Float64(), rand.Float64()))
		}
	}

	for i := 0; i < M; i++ {
		for j := 0; j < N; j++ {
			B.Set(i, j, complex(rand.Float64(), rand.Float64()))
		}
	}
	
	fmt.Println("Complex Matrix:\n", A, "\n")
	fmt.Println("Complex Matrix:\n", B, "\n")

	fmt.Println("Product : ", compMatrix.Product(A, B))

}

