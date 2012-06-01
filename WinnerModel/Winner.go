package main

import "fmt"
import "math/rand"
import "math/cmplx"
import "math"
import "../compMatrix"

func IIDChannel(Nr, Nt int) [][]complex128 {
	H := make([][]complex128, Nr) 
   	for i,_ := range H { 
		H[i] = make([]complex128, Nt) 
   	}
	for i := 0; i < Nr; i++ {
		for j := 0; j < Nt; j++ {
			H[i][j] = complex(rand.Float64(), rand.Float64())
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
}

func PAS(Nc int, phi []float64, phi0 []float64, sigma []float64, delta_theta []float64) []float64 {

	var func_temp1 = make([]float64, len(phi))
	var func_temp21 = make([]float64, len(phi))
	var func_temp2 = make([]float64, len(phi))
	var sum_temp = make([]float64, len(phi))
	var PowerAngularSpectrum = make([]float64, len(phi))
	var temp, temp1 float64;
	for i := 0; i < Nc; i++ {
		for ii, v := range phi {
			func_temp1[ii] = mystep(v - (phi0[i] - delta_theta[i])) - mystep(v - (phi0[i] + delta_theta[i]))
		}
		func_temp21 = LapPAS(phi, phi0[i], (sigma[i]/math.Sqrt(2.0)))
		for _,v := range func_temp21 {
			temp = temp + v
		}
		for ii,v := range func_temp21 {
			func_temp2[ii] = v / temp
		}
		
		for ii := 0; ii < len(func_temp2); ii++ {
			temp3 := func_temp2[ii] * func_temp1[ii]
			sum_temp[ii] = sum_temp[ii] + temp3
		}
	}
	for _,v := range sum_temp {
		temp1 = temp1 + v
	}
	
	for i,v := range sum_temp {
		PowerAngularSpectrum[i] = v / temp1
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

func MakeComplexMatrix(m, n int) [][]complex128 {
	mat := make([][]complex128, m) 
	for i,_ := range mat { 
		mat[i] = make([]complex128, n) 
	} 
	return mat 
}

type ChParameters struct {
	
	mean_Angle, AS, TruncRange []float64
}

func MakeMatrix(m, n int) [][]ChParameters { 
	mat := make([][]ChParameters, m) 
	for i,_ := range mat { 
		mat[i] = make([]ChParameters, n) 
   	} 
   	return mat 
}


func main() {

	spacing := 0.5			// Antenna spacing
	Nt := 3				// Number of transmit antennas
	Nr := 3				// Number of recieve antennas
	var SNRdB = [...]float64{-5,-4,-3,-2,-1,0,1,2,3,4,5}	// Input SNR in dB
	var SNR [11]float64		// Input SNR
	for i,v := range SNRdB {
		SNR[i] = math.Pow(10, (v/10));		// dB to linear scale conversion
	}

	iter := 1000			// Number of iterations
	Nc := 3				// Number of clusters

	range_a := -180.0
	range_b := 180.0
	range_c := 0.0
	range_d := 360.0

//	Ptx := MakeMatrix(Nt,Nt)

	var Ptx, Prx ChParameters

/*	for i := 0; i < Nt; i++ {
		for j := 0; j < Nt; j++ {
			Ptx[i][j].mean_Angle = make([]float64, Nc)	// Mean AoD
			Ptx[i][j].AS = make([]float64, Nc)		// Angle spread at the transmitter
			Ptx[i][j].TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter
		}
	}*/

	Ptx.mean_Angle = make([]float64, Nc)	// Mean AoD
	Ptx.AS = make([]float64, Nc)		// Angle spread at the transmitter
	Ptx.TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter

	for k := 0; k < Nc; k++ {
		Ptx.mean_Angle[k] = range_a + (range_b - range_a) * rand.Float64()
		Ptx.AS[k] = range_c + (range_d - range_c) * rand.Float64()
		Ptx.TruncRange[k] = 90.0
	}
/*	for k := 0; k < Nc; k++ {
		temp_angle := range_a + (range_b - range_a) * rand.Float64()
		temp_AS := range_c + (range_d - range_c) * rand.Float64()
		for i := 0; i < Nt; i++ {
			for j := 0; j < Nt; j++ {
				Ptx[i][j].mean_Angle[k] = temp_angle
				Ptx[i][j].AS[k] = temp_AS
				Ptx[i][j].TruncRange[k] = 90.0
			}			
		}
	}*/
	
	Prx.mean_Angle = make([]float64, Nc)	// Mean AoD
	Prx.AS = make([]float64, Nc)		// Angle spread at the transmitter
	Prx.TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter

	for k := 0; k < Nc; k++ {
		Prx.mean_Angle[k] = range_a + (range_b - range_a) * rand.Float64()
		Prx.AS[k] = range_c + (range_d - range_c) * rand.Float64()
		Prx.TruncRange[k] = 90.0
	}

/*	Prx := MakeMatrix(Nr,Nr)

	for i := 0; i < Nr; i++ {
		for j := 0; j < Nr; j++ {
			Prx[i][j].mean_Angle = make([]float64, Nc)	// Mean AoA
			Prx[i][j].AS = make([]float64, Nc)		// Angle spread at the reciever
			Prx[i][j].TruncRange = make([]float64, Nc)	// Laplace distribution truncated parameter
		}
	}
	for k := 0; k < Nc; k++ {
		temp_angle := range_a + (range_b - range_a) * rand.Float64()
		temp_AS := range_c + (range_d - range_c) * rand.Float64()
		for i := 0; i < Nr; i++ {
			for j := 0; j < Nr; j++ {
				Prx[i][j].mean_Angle[k] = temp_angle
				Prx[i][j].AS[k] = temp_AS
				Prx[i][j].TruncRange[k] = 90.0
			}
		}
	}*/

//fmt.Println("Transmit Paramenters: \n", Ptx, "\n")
//fmt.Println("Recieve Paramenters: \n", Prx, "\n")

/*	var mean_AoD = [3]float64 {0,0,0}		// Mean AoD
	var AS_Tx = [3]float64 {1,1,1}			// Angle spread at the transmitter
	var TruncRange_Tx = [3]float64 {90,90,90}	// Laplace distribution truncated parameter

	var mean_AoA = [3]float64 {0,0,0}		// Mean AoA
	var AS_Rx = [3]float64 {1,1,1}			// Angle spread at the reciever
	var TruncRange_Rx = [3]float64 {90,90,90}	// Laplace distribution truncated parameter
*/
		

/********************************************************/

/********************************************************/
/* This piece of code is just implementing the MATLAB command 
	phi = -pi:0.01:pi 
*/
	j := -math.Pi
	var phi = make([]float64, 628)
	for i,_ := range phi {
		phi[i] = j
		j = j + 0.01
	}
/********************************************************/

/********************************************************/
	var phi0 = make([]float64, Nc)
	var sigma = make([]float64, Nc)
	var delta_theta = make([]float64, Nc)
	
/*	var phi0_Rx = make([]float64, Nc)
	var sigma_Rx = make([]float64, Nc)
	var delta_theta_Rx = make([]float64, Nc)
*/
	var pdf float64
	var Rxx_temp = make([]float64, len(phi))
	var Rxy_temp = make([]float64, len(phi))
	var PowerAngularSpectrum = make([]float64, len(phi))	
	var Rxx, Rxy float64
	var rho complex128
	Rtx := MakeComplexMatrix(Nt, Nt)		// Transmit correlation matrix
	Rrx := MakeComplexMatrix(Nr, Nr)		// Recieve correlation matrix
/********************************************************/

/********************************************************/
/* This piece of code is implementing the following MATLAB commands
	func1 = cos(D*sin(phi))
	func2 = sin(D*sin(phi))
*/	
	D := 2 * math.Pi * spacing 			// D is the wave number x element spacing
	var func1 = make([]float64, len(phi))
	var func2 = make([]float64, len(phi)) 
	for i,v := range phi {
		func1[i] = math.Cos(D * math.Sin(v));
		func2[i] = math.Sin(D * math.Sin(v));
	}
/********************************************************/

/********************************************************/
/* This piece of code is used to generate transmit correlation matrix Rtx. Also, it normalize it at the end. PAS is a function which generates the power angular spectrum (sum of truncated Laplace distribution for each cluster).
*/
	for i,v := range Ptx.mean_Angle {
		phi0[i] = v * math.Pi/180;
	}
	for i,v := range Ptx.AS {
		sigma[i] = v * math.Pi/180;
	}
	for i,v := range Ptx.TruncRange {
		delta_theta[i] = v * math.Pi/180;
	}

	PowerAngularSpectrum = PAS(Nc, phi, phi0, sigma, delta_theta);
	pdf = 0	
	for _,v := range PowerAngularSpectrum {
		pdf = pdf + v
	}

	for ii := 0; ii < len(phi); ii++ {
		Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
		Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
	}
//	Rxx = 0
	for _,v := range Rxx_temp {
		Rxx = Rxx + v
	}
//	Rxy = 0
	for _,v := range Rxy_temp {
		Rxy = Rxy + v
	}

        rho = complex(Rxx, Rxy)

	for i := 0; i < Nt; i++ {
		for j := i; j < Nt; j++ {

/*			for i,v := range Ptx[i][j].mean_Angle {
				phi0_Tx[i] = v * math.Pi/180;
			}
			for i,v := range Ptx[i][j].AS {
				sigma_Tx[i] = v * math.Pi/180;
			}
			for i,v := range Ptx[i][j].TruncRange {
				delta_theta_Tx[i] = v * math.Pi/180;
			}*/
          
			if i == j {
				Rtx[i][j] = 1
			} else {
				Rtx[i][j] = rho
				Rtx[j][i] = cmplx.Conj(rho)
			}
		}
	}
	var temp1 float64
	for i := 0; i < Nt; i++ {
		temp1 = 0
		for j := 0; j < Nt; j++ {
			temp1 = temp1 + cmplx.Abs(Rtx[i][j])
		}
		for j := 0; j < Nt; j++ {
			Rtx[i][j] = Rtx[i][j] / complex(temp1,0)
		}
	}

//	fmt.Println("Transmitter Correlation Matrix: \n", Rtx, "\n")


/********************************************************/

/********************************************************/
/* This piece of code is used to generate recieve correlation matrix Rrx. Also, it normalize it at the end. PAS is a function which generates the power angular spectrum (sum of truncated Laplace distribution for each cluster).
*/
	for i,v := range Prx.mean_Angle {
		phi0[i] = v * math.Pi/180;
	}
	for i,v := range Prx.AS {
		sigma[i] = v * math.Pi/180;
	}
	for i,v := range Prx.TruncRange {
		delta_theta[i] = v * math.Pi/180;
	}

	PowerAngularSpectrum = PAS(Nc, phi, phi0, sigma, delta_theta);
	pdf = 0	
	for _,v := range PowerAngularSpectrum {
		pdf = pdf + v
	}

	for ii := 0; ii < len(phi); ii++ {
		Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
		Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
	}
//	Rxx = 0
	for _,v := range Rxx_temp {
		Rxx = Rxx + v
	}
//	Rxy = 0
	for _,v := range Rxy_temp {
		Rxy = Rxy + v
	}

        rho = complex(Rxx, Rxy)
 
	for i := 0; i < Nr; i++ {
		for j := i; j < Nr; j++ {
/*			for i,v := range Prx[i][j].mean_Angle {
				phi0_Rx[i] = v * math.Pi/180;
			}
			for i,v := range Prx[i][j].AS {
				sigma_Rx[i] = v * math.Pi/180;
			}
			for i,v := range Prx[i][j].TruncRange {
				delta_theta_Rx[i] = v * math.Pi/180;
			}*/

			if i == j {
				Rrx[i][j] = 1
			} else {
				Rrx[i][j] = rho
				Rrx[j][i] = cmplx.Conj(rho)
			}
		}
	}

	for i := 0; i < Nr; i++ {
		temp1 = 0
		for j := 0; j < Nr; j++ {
			temp1 = temp1 + cmplx.Abs(Rrx[i][j])
		}
		for j := 0; j < Nr; j++ {
			Rrx[i][j] = Rrx[i][j] / complex(temp1,0)
		}
	}
//	fmt.Println("Receiver Correlation Matrix: \n", Rrx, "\n")
/********************************************************/

/********************************************************/
	
//	var Hiid [][]complex128
	var temp2 = make([]complex128, len(SNR))
	var temp3 = make([]complex128, len(SNR))
	Hiid := compMatrix.Zeros(Nr, Nt)
	H := compMatrix.Zeros(Nr, Nt)
	Rtx_act := compMatrix.Zeros(Nt,Nt)
	Rrx_act := compMatrix.Zeros(Nr,Nr)
	R := compMatrix.Zeros(Nr,Nr)
	Riid := compMatrix.Zeros(Nr,Nr)
	Eye := compMatrix.Eye(Nr)

	for j := 0; j < Nt; j++ {
		for k := 0; k < Nt; k++ {
			Rtx[j][k] = cmplx.Sqrt(Rtx[j][k])
		}
	}
	for j := 0; j < Nr; j++ {
		for k := 0; k < Nr; k++ {
			Rrx[j][k] = cmplx.Sqrt(Rrx[j][k])
		}
	}
	for i := 0; i < Nt; i++ {
		for j := 0; j < Nt; j++ {
			Rtx_act.Set(i, j, Rtx[i][j])
		}
	}

	for i := 0; i < Nr; i++ {
		for j := 0; j < Nr; j++ {
			Rrx_act.Set(i, j, Rrx[i][j])
		}
	}
	
	for it := 0; it < iter; it++ {
//		Hiid = IIDChannel(Nr, Nt)
		
		for i := 0; i < Nr; i++ {
			for j := 0; j < Nt; j++ {
				Hiid.Set(i, j, complex(rand.Float64(), rand.Float64()))
			}
		}
		H = compMatrix.Product(Rrx_act, compMatrix.Product(Hiid, Rtx_act.Transpose()))


		compMatrix.TimesHilbert(H,H,R)
		compMatrix.TimesHilbert(Hiid, Hiid, Riid)

//		fmt.Println("Covariance Matrix:", R, "\n")
//		fmt.Println("Covariance Matrix IID:", Riid, "\n")

		for i,v := range SNR {
			R.Scale(complex(1/v,0))
			Riid.Scale(complex(1/v,0))
			R.Add(Eye)
			Riid.Add(Eye)
//			fmt.Println("Covariance Matrix:", R, "\n")
//			fmt.Println("Covariance Matrix IID:", Riid, "\n")
			temp2[i] += cmplx.Log(R.Det()) / cmplx.Log(complex(2,0))
			temp3[i] += cmplx.Log(Riid.Det()) / cmplx.Log(complex(2,0))
		}
//		fmt.Println("Cap:", temp2, "\n")
//		fmt.Println("Cap IID:", temp3, "\n")
	}
//	fmt.Println("Determinant : ", temp3)
	var Cap_H = make([]float64, len(SNR))
	var Cap_Hiid = make([]float64, len(SNR))
	for i,_ := range SNR {
		Cap_H[i] = cmplx.Abs(temp2[i]/complex(float64(iter),0))
		Cap_Hiid[i] = cmplx.Abs(temp3[i]/complex(float64(iter),0))
	}

	fmt.Println("Truncated Laplace pdf sum:", pdf, "\n")
	fmt.Println("Transmitter Correlation Matrix Sqrtroot transpose: \n", Rtx_act, "\n")
	fmt.Println("Receiver Correlation Matrix Sqrtroot: \n", Rrx_act, "\n")
	fmt.Println("IID Channel Matrix:\n", Hiid, "\n")
	fmt.Println("Correlated Channel Matrix:\n", H, "\n")

	fmt.Println("Capacity Hiid:\n", Cap_Hiid, "\n")
	fmt.Println("Capacity H:\n", Cap_H, "\n")

}

