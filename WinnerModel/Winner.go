package main

import "fmt"
import "math/rand"
import "math/cmplx"
import "math"

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

type Tx struct {
	
	mean_AoD, AS_Tx, TruncRange_Tx []float64
}

type Rx struct {
	
	mean_AoA, AS_Rx, TruncRange_Rx []float64
}

func PAS(Nc int, phi []float64, phi0 []float64, sigma []float64, delta_theta []float64) []float64 {

	var func_temp11 = make([]float64, len(phi))
	var func_temp12 = make([]float64, len(phi))
	var func_temp1 = make([]float64, len(phi))
	var func_temp21 = make([]float64, len(phi))
	var func_temp2 = make([]float64, len(phi))
	var func_temp3 = make([]float64, len(phi))
	var sum_temp = make([]float64, len(phi))
	var PowerAngularSpectrum = make([]float64, len(phi))
	var temp, temp1 float64;
	for i := 0; i < Nc; i++ {
		for ii, v := range phi {
			func_temp11[ii] = mystep(v - (phi0[i] - delta_theta[i])) 
			func_temp12[ii] = mystep(v - (phi0[i] + delta_theta[i]))
			func_temp1[ii] = func_temp11[ii] - func_temp12[ii]
		}
		func_temp21 = LapPAS(phi, phi0[i], (sigma[i]/math.Sqrt(2.0)))
		for _,v := range func_temp21 {
			temp = temp + v
		}
		for ii,v := range func_temp21 {
			func_temp2[ii] = v / temp
		}
		
		for ii := 0; ii < len(func_temp2); ii++ {
			func_temp3[ii] = func_temp2[ii] * func_temp1[ii]
			sum_temp[ii] = sum_temp[ii] + func_temp3[ii]
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

func MakeTxMatrix(m, n int) [][]Tx { 
	mat := make([][]Tx, m) 
	for i,_ := range mat { 
		mat[i] = make([]Tx, n) 
   	} 
   	return mat 
}

func MakeRxMatrix(m, n int) [][]Rx { 
   	mat := make([][]Rx, m) 
   	for i,_ := range mat { 
      		mat[i] = make([]Rx, n) 
   	} 
   	return mat 
}

func main() {

	c := 3.0e+8			// Speed of light
	fc := 1.9e+9			// Carrier frequency
	spacing := 0.5			// Antenna spacing
	Nt := 3				// Number of transmit antennas
	Nr := 3				// Number of recieve antennas
	var SNRdB = [6]float64{-20,-15,-10,-5,0,5}	// Input SNR in dB
	var SNR [6]float64		// Input SNR
	for i,v := range SNRdB {
		SNR[i] = math.Pow(10, (v/10));		// dB to linear scale conversion
	}

//	iter := 1000			// Number of iterations
	Nc := 3				// Number of clusters

	range_a := -180
	range_b := 180
	range_c := 0
	range_d := 360

	Ttx := MakeTxMatrix(Nt,Nt)
	for i := 0; i < Nt; i++ {
		for j := 0; j < Nt; j++ {
			Ttx[i][j].mean_AoD = make([]float64, Nc)	// Mean AoD
			Ttx[i][j].AS_Tx = make([]float64, Nc)		// Angle spread at the transmitter
			Ttx[i][j].TruncRange_Tx = make([]float64, Nc)	// Laplace distribution truncated parameter
			for k := 0; k < Nc; k++ {
				Ttx[i][j].mean_AoD[k] = float64(range_a + (range_b - range_a)) * rand.Float64()
				Ttx[i][j].AS_Tx[k] = float64(range_c + (range_d - range_c)) * rand.Float64()
				Ttx[i][j].TruncRange_Tx[k] = 90.0
			}			
		}
	}

	Trx := MakeRxMatrix(Nr,Nr)
	for i := 0; i < Nr; i++ {
		for j := 0; j < Nr; j++ {
			Trx[i][j].mean_AoA = make([]float64, Nc)	// Mean AoA
			Trx[i][j].AS_Rx = make([]float64, Nc)		// Angle spread at the reciever
			Trx[i][j].TruncRange_Rx = make([]float64, Nc)	// Laplace distribution truncated parameter
			for k := 0; k < Nc; k++ {
				Trx[i][j].mean_AoA[k] = float64(range_a + (range_b - range_a)) * rand.Float64()
				Trx[i][j].AS_Rx[k] = float64(range_c + (range_d - range_c)) * rand.Float64()
				Trx[i][j].TruncRange_Rx[k] = 90.0
				}
		}
	}

/*	var mean_AoD = [3]float64 {0,0,0}		// Mean AoD
	var AS_Tx = [3]float64 {1,1,1}			// Angle spread at the transmitter
	var TruncRange_Tx = [3]float64 {90,90,90}	// Laplace distribution truncated parameter

	var mean_AoA = [3]float64 {0,0,0}		// Mean AoA
	var AS_Rx = [3]float64 {1,1,1}			// Angle spread at the reciever
	var TruncRange_Rx = [3]float64 {90,90,90}	// Laplace distribution truncated parameter
*/
	D := 2 * math.Pi * fc * spacing / c		// D is the wave number x element spacing
	

/********************************************************/

/********************************************************/
/* This piece of code is just implementing the MATLAB command 
	phi = -pi:0.01:pi 
*/
	var temp int
	
	for i,j := 0,-math.Pi; j <= math.Pi; i++ {
		temp = i
		j = j + 0.01
	}
	
	j := -math.Pi
	var phi = make([]float64, temp)
	for i,_ := range phi {
		
		phi[i] = j
		j = j + 0.01
	}
/********************************************************/

/********************************************************/
	var phi0_Tx = make([]float64, Nc)
	var sigma_Tx = make([]float64, Nc)
	var delta_theta_Tx = make([]float64, Nc)
	
	var phi0_Rx = make([]float64, Nc)
	var sigma_Rx = make([]float64, Nc)
	var delta_theta_Rx = make([]float64, Nc)

	var pdf float64
	var Rxx_temp = make([]float64, len(phi))
	var Rxy_temp = make([]float64, len(phi))
	var PowerAngularSpectrum = make([]float64, len(phi))	
	var Rxx, Rxy float64
	var rho complex128
	Rtx := MakeComplexMatrix(Nt, Nt)
	Rrx := MakeComplexMatrix(Nr, Nr)
/********************************************************/

/********************************************************/
/* This piece of code is implementing the following MATLAB commands
	func1 = cos(D*sin(phi))
	func2 = sin(D*sin(phi))
*/	
	var func1 = make([]float64, len(phi))
	var func2 = make([]float64, len(phi)) 
	for i,v := range phi {
		func1[i] = math.Cos(D * math.Sin(v));
		func2[i] = math.Sin(D * math.Sin(v));
	}
/********************************************************/

/********************************************************/

	for i := 0; i < Nt; i++ {
		for j := i; j < Nt; j++ {

			for i,v := range Ttx[i][j].mean_AoD {
				phi0_Tx[i] = v * math.Pi/180;
			}
			for i,v := range Ttx[i][j].AS_Tx {
				sigma_Tx[i] = v * math.Pi/180;
			}
			for i,v := range Ttx[i][j].TruncRange_Tx {
				delta_theta_Tx[i] = v * math.Pi/180;
			}

			PowerAngularSpectrum = PAS(Nc, phi, phi0_Tx, sigma_Tx, delta_theta_Tx);
			pdf = 0	
			for _,v := range PowerAngularSpectrum {
				pdf = pdf + v
			}
			
			for ii := 0; ii < len(phi); ii++ {
				Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
				Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
			}
			Rxx = 0
			for _,v := range Rxx_temp {
				Rxx = Rxx + v
			}
			Rxy = 0
			for _,v := range Rxy_temp {
				Rxy = Rxy + v
			}

		        rho = complex(Rxx, Rxy)
            
			if i == j {
				Rtx[i][j] = 1
			} else {
				Rtx[i][j] = rho
				Rtx[j][i] = cmplx.Conj(rho)
			}
		}
	}
	var temp1 complex128
	for i := 0; i < Nt; i++ {
		temp1 = 0
		for j := 0; j < Nt; j++ {
			temp1 = temp1 + Rtx[i][j]
		}
		for j := 0; j < Nt; j++ {
			Rtx[i][j] = Rtx[i][j] / complex(cmplx.Abs(temp1),0)
		}
	}
/********************************************************/

/********************************************************/

	for i := 0; i < Nr; i++ {
		for j := i; j < Nr; j++ {
			for i,v := range Trx[i][j].mean_AoA {
				phi0_Rx[i] = v * math.Pi/180;
			}
			for i,v := range Trx[i][j].AS_Rx {
				sigma_Rx[i] = v * math.Pi/180;
			}
			for i,v := range Trx[i][j].TruncRange_Rx {
				delta_theta_Rx[i] = v * math.Pi/180;
			}

			PowerAngularSpectrum = PAS(Nc, phi, phi0_Rx, sigma_Rx, delta_theta_Rx);
			pdf = 0	
			for _,v := range PowerAngularSpectrum {
				pdf = pdf + v
			}
			
			for ii := 0; ii < len(phi); ii++ {
				Rxx_temp[ii] = func1[ii] * PowerAngularSpectrum[ii]
				Rxy_temp[ii] = func2[ii] * PowerAngularSpectrum[ii]
			}
			Rxx = 0
			for _,v := range Rxx_temp {
				Rxx = Rxx + v
			}
			Rxy = 0
			for _,v := range Rxy_temp {
				Rxy = Rxy + v
			}

		        rho = complex(Rxx, Rxy)
            
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
			temp1 = temp1 + Rrx[i][j]
		}
		for j := 0; j < Nr; j++ {
			Rrx[i][j] = Rrx[i][j] / complex(cmplx.Abs(temp1),0)
		}
	}

/********************************************************/

/********************************************************/
	
/*	var Hiid, H [][]complex128
	for i := 0; i < iter; i++ {
		Hiid = IIDChannel(Nr, Nt)
		

	}
*/
	fmt.Println("Truncated Laplace pdf sum:", pdf)
	fmt.Println("Transmitter Correlation Matrix:", Rtx)

	fmt.Println("Receiver Correlation Matrix:", Rrx)


}

