package synstation

import "geom"
import "container/list"
import "rand"

// structure to store evaluation of interference at a location
// this has to be initialized with PhysReceiver.Init() function to init memory
type PhysReceiverSectored struct {
	R []PhysReceiver
}

func (r *PhysReceiverSectored) Init(p geom.Pos, Rgen *rand.Rand) {
	r.R = make([]PhysReceiver, 3)
	for i := 0; i < 3; i++ {
		r.R[i].Init(p, Rgen)
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

func (r *PhysReceiverSectored) SetPos(p geom.Pos) {
	for i := 0; i < 3; i++ {
		r.R[i].SetPos(p)
	}
}
func (r *PhysReceiverSectored) GetPos() geom.Pos {
	return r.R[0].Pos
}


// Evaluates interference for all channels with overlapping effect,
// channel 0 is considered to have no interference as traffic is suppose to only hold minimal signalization 
func (rx *PhysReceiverSectored) MeasurePower(tx EmitterInt) {
	for i := 0; i < 3; i++ {
		rx.R[i].MeasurePower(tx)
	}
}


func (r *PhysReceiverSectored) EvalBestSignalSNR(ch int) (Rc *ChanReceiver, SNR float64) {
	var e [3]float64
	var R [3]*ChanReceiver

	for i := range e {
		R[i], e[i] = r.R[i].EvalBestSignalSNR(ch)
	}
	ir := findMax(e[:])
	return R[ir], e[ir]

	/*Rc = &r.R[0].Channels[ch]
	SNR = 0

	if Rc.Signal != nil {

		if ch == 0 {
			SNR = Rc.PrMax / 1e-15 //WNoise
		} else {
			SNR = Rc.PrMax / (Rc.Pint - Rc.PrMax + WNoise)
		}

	}

	return*/

}


func (r *PhysReceiverSectored) EvalChRSignalSNR(ch int, k int) (Rc *ChanReceiver, SNR float64) {
	var e [3]float64
	var R [3]*ChanReceiver

	for i := range e {
		R[i], e[i] = r.R[i].EvalChRSignalSNR(ch, k)
	}
	ir := findMax(e[:])
	return R[ir], e[ir]

}

func (r *PhysReceiverSectored) EvalSignalConnection(ch int) (*ChanReceiver, float64, float64) {
	var e [3]float64
	var R [3]*ChanReceiver
	var BER [3]float64

	for i := range e {
		R[i], e[i], BER[i] = r.R[i].EvalSignalConnection(ch)
	}
	ir := findMax(e[:])
	return R[ir], e[ir], BER[ir]
}


func (r *PhysReceiverSectored) EvalSignalSNR(ex EmitterInt, ch int) (Rc *ChanReceiver, SNR, Pr, K float64) {
	var R [3]*ChanReceiver
	var s [3]float64
	var p [3]float64
	var k [3]float64

	for i := range s {
		R[i], s[i], p[i], k[i] = r.R[i].EvalSignalSNR(ex, ch)
	}
	ir := findMax(s[:])
	return R[ir], s[ir], p[ir], k[ir]
}


func (r *PhysReceiverSectored) EvalSignalBER(ex EmitterInt, ch int) (Rc *ChanReceiver, BER, SNR, Pr float64) {
	var R [3]*ChanReceiver
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


func (rx *PhysReceiverSectored) DoTracking(Connec *list.List) bool {
	return false
}


func (rx *PhysReceiverSectored) GenFastFading() {
	for i := range rx.R {
		rx.R[i].GenFastFading()
	}
}


func (r *PhysReceiverSectored) GetPr(mi, ch int) (p float64, Rc *ChanReceiver) {

	var b [3]float64

	for i := range b {
		b[i] = r.R[i].Channels[ch].pr[mi]
	}
	ir := findMax(b[:])
	return b[ir], &r.R[ir].Channels[ch]

}


func (r *PhysReceiverSectored) GetK(mi int) (k float64) {

	/*var b [3]float64
	for i := range b {
		b[i] = r.R[i].Channels[ch].pr[mi]
	}
	ir := findMax(b[:])*/
	// TODO, figuree out if AOA of line of sight is a problem here or not
	return r.R[0].kk[mi]

}

