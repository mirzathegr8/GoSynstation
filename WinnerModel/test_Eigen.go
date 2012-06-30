package main

import "fmt"
//import "math/rand"
//import "math/cmplx"
//import "math"
import "../compMatrix"



func main() {

/********************************************************/

/********************************************************/
	N := 2
	H := compMatrix.Zeros(N, N)

	//fmt.Println("Complex Matrix:\n", H, "\n")
	
/*	for i := 0; i < N; i++ {
		for j := i; j < N; j++ {
			A := complex(rand.Float64()/math.Sqrt(2), rand.Float64()/math.Sqrt(2))
			H.Set(i, j, A)
			H.Set(j, i, cmplx.Conj(A))
			if i == j {
				H.Set(i, j, complex(cmplx.Abs(A),0))
			}
		}
	}*/
	
	H.Set(0,0,complex(1,0))
	H.Set(1,1,complex(1,0))
	H.Set(0,1,complex(0.83504,-0.00610))
	H.Set(1,0,complex(0.83504,0.00610))
	fmt.Println("Complex Matrix:\n", H, "\n")
	
	U,D,_ := H.Eigen()
		
	fmt.Println("U: \n", U, "\n D :\n", D)
	//fmt.Println("\n Complex Matrix:\n", compMatrix.Product(compMatrix.Product(U,D),U.Transpose()), "\n")

}

