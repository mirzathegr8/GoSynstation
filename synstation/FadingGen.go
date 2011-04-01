package synstation

//import c "cmath"
import "math"
//import "fmt"

const F = 900 * 10e6 //fréquence du canal en Hz

const cel = 3 * 10e8 //vitesse de propagation en m/s


type FilterInt interface {
	nextValue(input float64) (output float64)
}


type PassNull struct{}

func (p *PassNull) nextValue(input float64) (output float64) {
	return 1.0
}

var PNF PassNull

type Filter struct {
	a []float64
	b []float64
	z [3]float64
	//z2 [3]float64
}

func (f *Filter) nextValue(input float64) (output float64) {
	f.z[0] = input * f.a[0]
	for i := 1; i < len(f.a); i++ {
		f.z[0] -= f.a[i] * f.z[i]
	}
	for i := 0; i < len(f.b); i++ {
		output += f.b[i] * f.z[i]
	}
	for i := len(f.z) - 1; i > 0; i-- {
		f.z[i] = f.z[i-1]
	}

	/*for i := len(f.z) - 1; i > 0; i-- {
		f.z[i] = f.z[i-1]
	}

	for i := len(f.z2) - 1; i > 0; i-- {
		f.z2[i] = f.z2[i-1]
	}

	f.z[0] = input

	for i := 0; i < len(f.a); i++ {
		output += f.a[i] * f.z[i]
	}

	for i := 1; i < len(f.b); i++ {
		output -= f.b[i] * f.z2[i]
	}

	f.z2[0] = output*/

	//fmt.Println(f.z, f.z2)

	return
}


func Butter(W float64) (f *Filter) {

	f = new(Filter)
	// Prewarp to the band edges to s plane

	var T = complex128(2) //      # sampling frequency of 2 Hz
	Wt := math.Tan(math.Pi * W / 2.0)

	// Generate splane poles for the prototype butterworth filter
	// source: Kuc

	pole := []complex128{
		(-0.707106781186547 + 0.707106781186548i),
		(-0.707106781186548 - 0.707106781186547i),
	} //C*c.Exp( 1i*pi*(2*[1:n] + n - 1)/(2*n));


	gain := complex128(1.0)

	// splane frequency transform
	//  zero, pole, gain = sftrans(zero, pole, gain, Wt);
	gain = gain * complex(math.Pow(1.0/Wt, -2), 0)
	for i := range pole {
		pole[i] = complex(Wt, 0) * pole[i]
	}

	// zero, pole, gain = bilinear(zero, pole, gain, T);
	gain = gain / ((T - pole[0]*T) / T * (T - pole[1]*T) / T)
	for i := range pole {
		pole[i] = (T + pole[i]*T) / (T - pole[i]*T)
	}

	zero := []complex128{
		(-1),
		(-1),
	}

	//fmt.Println(pole, zero, gain)

	f.b = Sreal(poly(zero))
	for i := range f.b {
		f.b[i] = real(gain) * f.b[i]
	}
	f.a = Sreal(poly(pole))

	for i := range f.z {
		f.z[i] = 1 //init stream to non null
		//f.z2[i] = 1 //init stream to non null
	}

	//adjust gain
	for j := range f.b {
		f.b[j] *= math.Sqrt(1 / W) //scale input to compensate for lowpass and have same output power as input
	}

	return
}

func poly(x []complex128) (y []complex128) {

	n := len(x)
	y = make([]complex128, n+1)
	y2 := make([]complex128, n+1)
	y[0] = complex(1, 0)
	y2[0] = complex(1, 0)
	for j := 0; j < n; j++ {
		for i := 0; i <= j; i++ {
			y2[i+1] = y[i+1] - x[j]*y[i]
		}
		for i := 0; i <= j; i++ {
			y[i+1] = y2[i+1]
		}
	}
	return
}
func Sreal(x []complex128) (y []float64) {
	y = make([]float64, len(x))
	for i := range x {
		y[i] = real(x[i])
	}
	return
}


func Cheby(Rp, W float64) (f *Filter) {

	f = new(Filter)

	var T = complex128(2) //      # sampling frequency of 2 Hz
	Wt := math.Tan(math.Pi * W / 2.0)

	epsilon := math.Sqrt(math.Pow(10, (Rp/10)) - 1)
	v0 := math.Asinh(1/epsilon) / 2
	pole := []complex128{ //exp(1i*pi*[-(n-1):2:(n-1)]/(2*n));
		(0.707106781186548 - 0.707106781186547i),
		(0.707106781186548 + 0.707106781186547i),
	}

	for i := range pole {
		pole[i] = complex(-math.Sinh(v0)*real(pole[i]), math.Cosh(v0)*imag(pole[i]))
	}

	// compensate for amplitude at s=0
	gain := complex(1, 0)
	for i := range pole {
		gain *= -pole[i]
	}

	// if n is even, the ripple starts low, but if n is odd the ripple
	// starts high. We must adjust the s=0 amplitude to compensate.

	gain = gain / complex(math.Pow(10, (Rp/20)), 0)

	// splane frequency transform
	//  [zero, pole, gain] = sftrans(zero, pole, gain, Wt, stop);
	gain = gain * complex(math.Pow(1.0/Wt, -2), 0)
	for i := range pole {
		pole[i] = complex(Wt, 0) * pole[i]
	}

	// Use bilinear transform to convert poles to the z plane
	//    [zero, pole, gain] = bilinear(zero, pole, gain, T);

	gain = gain / ((T - pole[0]*T) / T * (T - pole[1]*T) / T)
	for i := range pole {
		pole[i] = (T + pole[i]*T) / (T - pole[i]*T)
	}

	zero := []complex128{
		(-1),
		(-1),
	}

	f.b = Sreal(poly(zero))
	for i := range f.b {
		f.b[i] = real(gain) * f.b[i]
	}
	f.a = Sreal(poly(pole))

	for i := range f.z {
		f.z[i] = 1 //init stream to non null
		//f.z2[i] = 1 //init stream to non null
	}

	for j := range f.b {
		f.b[j] *= math.Sqrt(2.6 / W) //scale input to compensate for output power

	}

	return

}


func (f *Filter) PassNull() {
	f.a = []float64{0, -1.0}
	f.b = []float64{0, 1.0}
	for i := range f.z {
		f.z[i] = 1 //init stream to non null
		//f.z2[i] = 1 //init stream to non null
	}
}

