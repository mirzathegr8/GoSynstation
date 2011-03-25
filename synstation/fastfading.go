package synstation

//import "math"
import "rand"
//import "fmt"

//fast fading parameters
type FadingData struct {
	ff_R [M]float64
}


func (f *FadingData) GenerateFading(Rgen *rand.Rand, p PowerData) {
	K := float64(0)
	for i := 0; i < M; i++ {
		a := Rgen.NormFloat64()
		b := Rgen.NormFloat64()
		f.ff_R[i] = ((K+a)*(K+a) + b*b) / 2
		f.ff_R[i] *= p.pr[i]
	}
}

