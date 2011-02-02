package synstation

import "geom"
import "container/list"

// structure to store evaluation of interference at a location
// this has to be initialized with PhysReciever.Init() function to init memory
type PhysRecieverSectored struct {
	R []PhysReciever
}

func (r *PhysRecieverSectored) Init() {
	r.R = make([]PhysReciever, 3)
	for i := 0; i < 3; i++ {
		r.R[i].Init()
	}

	r.R[0].Orientation[0] = -1
	r.R[1].Orientation[0] = -1
	r.R[2].Orientation[0] = -1

	for i := 1; i < len(r.R[0].Orientation); i++ {
		r.R[0].Orientation[i] = 0
		r.R[1].Orientation[i] = 120
		r.R[2].Orientation[i] = 240
	}
}

func (r *PhysRecieverSectored) SetPos(p geom.Pos) {
	for i := 0; i < 3; i++ {
		r.R[i].SetPos(p)
	}
}
func (r *PhysRecieverSectored) GetPos() *geom.Pos {
	return &r.R[0].Pos
}


// Evaluates interference for all channels with overlapping effect,
// channel 0 is considered to have no interference as traffic is suppose to only hold minimal signalization 
func (rx *PhysRecieverSectored) MeasurePower(tx EmitterInt) {
	for i := 0; i < 3; i++ {
		rx.R[i].MeasurePower(tx)
	}
}


func (r *PhysRecieverSectored) EvalSignalPr(e EmitterInt, ch int) (Pr, K float64) {
	var p [3]float64
	var k [3]float64

	for i := range p {
		p[i], k[i] = r.R[i].EvalSignalPr(e, ch)
	}
	ir := findMax(p[:])
	return p[ir], k[ir]
}

func (r *PhysRecieverSectored) EvalBestSignalSNR(ch int) (Rc *ChanReciever, SNR float64) {
	/*var e [3]float64
	var R [3]*ChanReciever		

	for i:=range e {
		R[i],e[i] = r.R[i].EvalBestSignalSNR(ch)	
	}
	ir := findMax(e[:])
	return R[ir],e[ir]*/

	Rc = &r.R[0].Channels[ch]
	SNR = 0

	if Rc.Signal != nil {

		if ch == 0 {
			SNR = Rc.PrMax / 1e-15 //WNoise
		} else {
			SNR = Rc.PrMax / (Rc.Pint - Rc.PrMax + WNoise)
		}

	}

	return

}

func (r *PhysRecieverSectored) EvalSignalConnection(ch int) (*ChanReciever, float64) {
	var e [3]float64
	var R [3]*ChanReciever

	for i := range e {
		R[i], e[i] = r.R[i].EvalSignalConnection(ch)
	}
	ir := findMax(e[:])
	return R[ir], e[ir]
}


func (r *PhysRecieverSectored) EvalSignalSNR(ex EmitterInt, ch int) (Rc *ChanReciever, SNR, Pr, K float64) {
	var R [3]*ChanReciever
	var s [3]float64
	var p [3]float64
	var k [3]float64

	for i := range s {
		R[i], s[i], p[i], k[i] = r.R[i].EvalSignalSNR(ex, ch)
	}
	ir := findMax(s[:])
	return R[ir], s[ir], p[ir], k[ir]
}


func (r *PhysRecieverSectored) EvalSignalBER(ex EmitterInt, ch int) (Rc *ChanReciever, BER, SNR, Pr float64) {
	var R [3]*ChanReciever
	var b [3]float64
	var s [3]float64
	var p [3]float64

	for i := range b {
		R[i], b[i], s[i], p[i] = r.R[i].EvalSignalBER(ex, ch)
	}
	ir := findMin(b[:])
	return R[ir], b[ir], s[ir], p[ir]
}

func findMax(arr []float64) int {
	max := arr[0]
	r := 0
	for i, v := range arr {
		if v > max {
			max = v
			r = i
		}
	}
	return r
}


func findMin(arr []float64) int {
	min := arr[0]
	r := 0
	for i, v := range arr {
		if v < min {
			min = v
			r = i
		}
	}
	return r
}


func (rx *PhysRecieverSectored) DoTracking(Connec *list.List) bool {
	return false
}

func (rx *PhysRecieverSectored) RicePropagation(E EmitterInt) (fading float64, K float64) {
	return rx.R[0].RicePropagation(E)
}

