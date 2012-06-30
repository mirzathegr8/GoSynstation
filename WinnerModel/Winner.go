package main

import "fmt"
import "math/rand"
import "math/cmplx"
import "math"
import "../compMatrix"

/*func IIDChannel(Nr, Nt int) [][]complex128 {
	H := make([][]complex128, Nr) 
   	for i,_ := range H { 
		H[i] = make([]complex128, Nt) 
   	}
	for i := 0; i < Nr; i++ {
		for j := 0; j < Nt; j++ {
			H[i][j] = complex(rand.NormFloat64()/math.Sqrt(2), rand.NormFloat64()/math.Sqrt(2))
		}
	}
	return H 
}

func MatrixInitialization(Nr, Nt int) [][]complex128 {
	H := make([][]complex128, Nr) 
   	for i,_ := range H { 
		H[i] = make([]complex128, Nt) 
   	}
	for i := 0; i < Nr; i++ {
		for j := 0; j < Nt; j++ {
			H[i][j] = complex(0, 0)
		}
	}
	return H 
}*/

func PAS(Nc int, phi []float64, pdB []float64, phi0 []float64, sigma []float64, delta_theta []float64) []float64 {

	var func_temp1 = make([]float64, len(phi))
	var func_temp2 = make([]float64, len(phi))
	var sum_temp = make([]float64, len(phi))
	var Q = make([]float64, Nc)
	var PowerAngularSpectrum = make([]float64, len(phi))
	
	Q = CalQ(Nc, pdB, sigma, delta_theta)
	for i := 0; i < Nc; i++ {
		for ii, v := range phi {
			func_temp1[ii] = mystep(v - (phi0[i] - delta_theta[i])) - mystep(v - (phi0[i] + delta_theta[i]))
		}
		func_temp2 = LapPAS(phi, phi0[i], (sigma[i]/math.Sqrt(2.0)))
		
		for ii := 0; ii < len(func_temp2); ii++ {
			temp3 := func_temp2[ii] * func_temp1[ii]
			sum_temp[ii] = sum_temp[ii] + (Q[i] * temp3)
		}
	}
	
	for i,v := range sum_temp {
		PowerAngularSpectrum[i] = v
	}
	
	return PowerAngularSpectrum
		
}

func mystep(n float64) float64 {
	
	var y float64
	if n >= 0 {
		y = 1.0
	} else {
		y = 0.0
	}
	return y;
}

func LapPAS(phi []float64, mu float64, b float64) []float64 {
	
	var f = make([]float64, len(phi))

	for i,v := range phi {
		f[i] = 	math.Exp(-math.Abs(v - mu)/b)/(2*b)
	}	
	return f
}

func CalQ(Nc int, powerdb []float64, sigma []float64, delta_theta []float64) []float64 {

// change power in dB to linear
	
	var powerlin = make([]float64, len(powerdb))
	var Q = make([]float64, Nc)
	for i,v := range powerdb {
		powerlin[i] = math.Pow(10, (v/10))		// dB to linear scale conversion
	}

// calculate Q
	if Nc == 1 {
//		Q = 1/(1 - math.Exp(-math.Sqrt(2) * delta_theta/sigma))
	} else {
		if Nc > 1 {
			//Q = zeros(1, Nc);
			denom := 1 - math.Exp(-math.Sqrt(2) * delta_theta[0]/sigma[0])
			for i := 1; i < Nc; i++ {
				denom = denom + (sigma[i] * powerlin[i])/(sigma[0] * powerlin[0]) * (1 - math.Exp(-math.Sqrt(2) * delta_theta[i]/sigma[i]))
			}
			Q[0] = 1/denom
			for i := 1; i < Nc; i++ {
				Q[i] = Q[0] * (sigma[i] * powerlin[i])/(sigma[0] * powerlin[0])
			}
		}
	}


	return Q
}

func MakeComplexMatrix(m, n int) [][]complex128 {
	mat := make([][]complex128, m) 
	for i,_ := range mat { 
		mat[i] = make([]complex128, n) 
	}
	return mat 
}

type ChParameters struct {
	
	power, mean_Angle, AS, TruncRange []float64
}

/*func MakeMatrix(m, n int) [][]ChParameters { 
	mat := make([][]ChParameters, m) 
	for i,_ := range mat { 
		mat[i] = make([]ChParameters, n) 
   	} 
   	return mat 
}*/


func main() {

	spacing := 0.5			// Antenna spacing
	Nt := 2				// Number of transmit antennas
	Nr := 2				// Number of recieve antennas
	var SNRdB = [...]float64{-5,-4,-3,-2,-1,0,1,2,3,4,5}	// Input SNR in dB
	var SNR [11]float64		// Input SNR
	for i,v := range SNRdB {
		SNR[i] = math.Pow(10, (v/10))		// dB to linear scale conversion
	}

	iter := 1000			// Number of iterations
	Nc := 3				// Number of clusters

/*	range_a := -180.0
	range_b := 180.0
	range_c := 0.0
	range_d := 360.0
*/
	var Ptx, Prx ChParameters

	Ptx.power = make([]float64, Nc)		// Power of path in dB at transmitter
	Ptx.mean_Angle = make([]float64, Nc)	// Mean AoD
	Ptx.AS = make([]float64, Nc)		// Angle spread at the transmitter
	Ptx.TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter

	var power_model = [...]float64{0.0, -15.8, -13.5}
	var mean_angle_model = [...]float64{0.0, -107.0, -100.0}
//	var mean_angle_model = [...]float64{0, -107, -100}
//	var mean_angle_model = [...]float64{0, -107, -100}

	for k := 0; k < Nc; k++ {
//		Ptx.power[k] = range_a + (range_b - range_a) * rand.Float64()
		Ptx.power[k] = power_model[k]
//		Ptx.mean_Angle[k] = range_a + (range_b - range_a) * rand.Float64()
		Ptx.mean_Angle[k] = mean_angle_model[k]
//		Ptx.AS[k] = range_c + (range_d - range_c) * rand.Float64()
		Ptx.AS[k] = 5.0
		Ptx.TruncRange[k] = 90.0
	}

	Prx.power = make([]float64, Nc)		// Power of path in dB at receiver
	Prx.mean_Angle = make([]float64, Nc)	// Mean AoD
	Prx.AS = make([]float64, Nc)		// Angle spread at the receiver
	Prx.TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter
	
	var mean_angle_model1 = [...]float64{0.0, -110.0, 102.0}
	for k := 0; k < Nc; k++ {
//		Prx.power[k] = range_a + (range_b - range_a) * rand.Float64()
		Ptx.power[k] = power_model[k]
//		Prx.mean_Angle[k] = range_a + (range_b - range_a) * rand.Float64()
		Prx.mean_Angle[k] = mean_angle_model1[k]
//		Prx.AS[k] = range_c + (range_d - range_c) * rand.Float64()
		Prx.AS[k] = 5.0
		Prx.TruncRange[k] = 90.0
	}


/********************************************************/

/********************************************************/
/* This piece of code is just implementing the MATLAB command 
	phi = -pi:0.01:pi 
*/
	j := -math.Pi
	step := 0.01
	var phi = make([]float64, 629)
	for i,_ := range phi {
		phi[i] = j
		j = j + step
	}
/********************************************************/

/********************************************************/
	var pdB = make([]float64, Nc)
	var phi0 = make([]float64, Nc)
	var sigma = make([]float64, Nc)
	var delta_theta = make([]float64, Nc)
	
	var pdf float64
	var Rxx_temp = make([]float64, len(phi))
	var Rxy_temp = make([]float64, len(phi))
	var PowerAngularSpectrum = make([]float64, len(phi))	
	var Rxx, Rxy float64
	var rho complex128
	Rtx := MakeComplexMatrix(Nt, Nt)		// Transmit correlation matrix
	Rrx := MakeComplexMatrix(Nr, Nr)		// Recieve correlation matrix
	var func1 = make([]float64, len(phi))
	var func2 = make([]float64, len(phi)) 

/********************************************************/

/********************************************************/
/* This piece of code is used to generate transmit correlation matrix Rtx. Also, it normalize it at the end. PAS is a function which generates the power angular spectrum (sum of truncated Laplace distribution for each cluster).
*/
	for i,v := range Ptx.power {
		pdB[i] = v
	}
	for i,v := range Ptx.mean_Angle {
		phi0[i] = v * math.Pi/180
	}
	for i,v := range Ptx.AS {
		sigma[i] = v * math.Pi/180
	}
	for i,v := range Ptx.TruncRange {
		delta_theta[i] = v * math.Pi/180
	}

	PowerAngularSpectrum = PAS(Nc, phi, pdB, phi0, sigma, delta_theta)
//	fmt.Println("Sum of Truncated Laplace pdfs at Transmitter:", PowerAngularSpectrum, "\n")

	pdf = 0
	for _,v := range PowerAngularSpectrum {
		pdf = pdf + v
	}
//	fmt.Println("Truncated Laplace pdf sum at transmitter:", pdf * step, "\n")

	for i := 0; i < Nt; i++ {
		for j := i; j < Nt; j++ {

			if i == j {
				Rtx[i][j] = 1
			} else {
				D := (float64(j) - float64(i)) * 2 * math.Pi * spacing	// D is the wave number x element spacing

				for i1,v := range phi {
					func1[i1] = math.Cos(D * math.Sin(v))
					func2[i1] = math.Sin(D * math.Sin(v))
				}

				for ii := 0; ii < len(phi); ii++ {
					Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
					Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
				}

				for _,v := range Rxx_temp {
					Rxx = Rxx + v
				}
	
				for _,v := range Rxy_temp {
					Rxy = Rxy + v
				}

			        rho = complex(Rxx * step, Rxy * step)

				Rtx[i][j] = rho
				Rtx[j][i] = cmplx.Conj(rho)
			}
		}
	}


/********************************************************/

/********************************************************/
/* This piece of code is used to generate recieve correlation matrix Rrx. Also, it normalize it at the end. PAS is a function which generates the power angular spectrum (sum of truncated Laplace distribution for each cluster).
*/

	for i,v := range Prx.power {
		pdB[i] = v
	}	
	for i,v := range Prx.mean_Angle {
		phi0[i] = v * math.Pi/180
	}
	for i,v := range Prx.AS {
		sigma[i] = v * math.Pi/180
	}
	for i,v := range Prx.TruncRange {
		delta_theta[i] = v * math.Pi/180
	}

	PowerAngularSpectrum = PAS(Nc, phi, pdB, phi0, sigma, delta_theta)
//	fmt.Println("Sum of Truncated Laplace pdfs at receiver:", PowerAngularSpectrum, "\n")

	pdf = 0	
	for _,v := range PowerAngularSpectrum {
		pdf = pdf + v
	}
//	fmt.Println("Truncated Laplace pdf sum at receiver:", pdf * step, "\n")

	for i := 0; i < Nr; i++ {
		for j := i; j < Nr; j++ {

			if i == j {
				Rrx[i][j] = 1
			} else {
				D := (float64(j) - float64(i)) * 2 * math.Pi * spacing	// D is the wave number x element spacing

				for i1,v := range phi {
					func1[i1] = math.Cos(D * math.Sin(v))
					func2[i1] = math.Sin(D * math.Sin(v))
				}

				for ii := 0; ii < len(phi); ii++ {
					Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
					Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
				}

				for _,v := range Rxx_temp {
					Rxx = Rxx + v
				}

				for _,v := range Rxy_temp {
					Rxy = Rxy + v
				}

			        rho = complex(Rxx * step, Rxy * step)

				Rrx[i][j] = rho
				Rrx[j][i] = cmplx.Conj(rho)
			}
		}
	}


/********************************************************/

/********************************************************/
	

	Hiid := compMatrix.Zeros(Nr,Nt)
	H := compMatrix.Zeros(Nr,Nt)
	Rtx_act := compMatrix.Zeros(Nt,Nt)
	Rrx_act := compMatrix.Zeros(Nr,Nr)
	Rtx_sqrt := compMatrix.Zeros(Nt,Nt)
	Rrx_sqrt := compMatrix.Zeros(Nr,Nr)

	for i := 0; i < Nt; i++ {
		for j := 0; j < Nt; j++ {
			Rtx_act.Set(i, j, Rtx[i][j])
		}
	}
	fmt.Println("Transmitter Correlation Matrix: \n", Rtx_act, "\n")

	for i := 0; i < Nr; i++ {
		for j := 0; j < Nr; j++ {
			Rrx_act.Set(i, j, Rrx[i][j])
		}
	}
	fmt.Println("Receiver Correlation Matrix: \n", Rrx_act, "\n")

	U,D,_ := Rrx_act.Eigen()
	Dsqrt := compMatrix.Zeros(Nr,Nr)
	for i := 0; i < Nr; i++ {
		Dsqrt.Set(i, i, complex( math.Sqrt( cmplx.Abs( D.Get(i,i) ) ), 0) )
	}
	Rrx_sqrt = compMatrix.Product(U, compMatrix.Product(Dsqrt, U.Hilbert()))
	fmt.Println("Receiver Correlation Matrix Square Root: \n", Rrx_sqrt, "\n")
	fmt.Println("Receiver Correlation Matrix Square Root Check: \n", compMatrix.Product(Rrx_sqrt, Rrx_sqrt), "\n")

	U,D,_ = Rtx_act.Eigen()
	Dsqrt = compMatrix.Zeros(Nr,Nr)
	for i := 0; i < Nr; i++ {
		Dsqrt.Set(i, i, complex( math.Sqrt( cmplx.Abs( D.Get(i,i) ) ), 0) )
	}
	Rtx_sqrt = compMatrix.Product(U, compMatrix.Product(Dsqrt, U.Hilbert()))
	fmt.Println("Transmitter Correlation Matrix Square Root: \n", Rtx_sqrt, "\n")
	fmt.Println("Transmitter Correlation Matrix Square Root Check: \n", compMatrix.Product(Rtx_sqrt, Rtx_sqrt), "\n")
	norm := MakeComplexMatrix(Nr,Nt)
	var link00 = make([]float64, iter)
	var link10 = make([]float64, iter)

	for it := 0; it < iter; it++ {
		for i := 0; i < Nr; i++ {
			for j := 0; j < Nt; j++ {
				Hiid.Set(i, j, complex(rand.NormFloat64()/math.Sqrt(2), rand.NormFloat64()/math.Sqrt(2)))
			}
		}
		H = compMatrix.Product(Rrx_sqrt, compMatrix.Product(Hiid, Rtx_sqrt.Transpose()))

		
		link00[it] = math.Log(cmplx.Abs(H.Get(0,0)))
		link10[it] = math.Log(cmplx.Abs(H.Get(1,0)))
		for i := 0; i < Nr; i++ {
			for j := 0; j < Nt; j++ {
				norm[i][j] = norm[i][j] + complex(math.Pow(cmplx.Abs(H.Get(i,j)), 2),0)
			}
		}


	}

	for i := 0; i < Nr; i++ {
		for j := 0; j < Nt; j++ {
			norm[i][j] = norm[i][j]/complex(float64(iter),0)
		}
	}

	MeanMatrix := compMatrix.Zeros(Nr,Nt)
	for i := 0; i < Nr; i++ {
		for j := 0; j < Nt; j++ {
			MeanMatrix.Set(i, j, norm[i][j])
		}
	}

	

	
	fmt.Println("IID Channel Matrix:\n", Hiid, "\n")
	fmt.Println("Correlated Channel Matrix:\n", H, "\n")

	fmt.Println("MeanMatrix:\n", MeanMatrix, "\n")
//	fmt.Println("Semilogy Plot:\n", link00, "\n")
//	fmt.Println("Semilogy Plot:\n", link10, "\n")
}
