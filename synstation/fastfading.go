package synstation

//import "math"
import "rand"
//import "fmt"

//fast fading parameters
type FadingData struct {
	ff_R [M]float64
}

const K = 0

func (f *FadingData) GenerateFading(Rgen *rand.Rand) {

	for i := 0; i < M; i++ {
		a := Rgen.NormFloat64()
		b := Rgen.NormFloat64()
		f.ff_R[i] = ((K+a)*(K+a) + b*b) / 2

		//fmt.Println(f.ff_R[i])
	}
}

func (f *FadingData) GetFastFading(m int) float64 {
	return f.ff_R[m]
}

