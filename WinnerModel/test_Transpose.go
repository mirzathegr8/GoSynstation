package main

import "fmt"
import "math/rand"
//import "math/cmplx"
//import "math"
import "../compMatrix"



func main() {

/********************************************************/

/********************************************************/
	N := 3
	H := compMatrix.Zeros(N, N)

	//fmt.Println("Complex Matrix:\n", H, "\n")
	
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			H.Set(i, j, complex(rand.Float64(), rand.Float64()))
		}
	}
	
	fmt.Println("Complex Matrix:\n", H, "\n")
//	conj := H.Conj()
	fmt.Println("Conjugate Transpose : ", H.Hilbert())

}

