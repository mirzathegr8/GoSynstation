package synstation

import "math"
import "geom"
import "container/list"
import "fmt"

// Structure to hold interference level, and multilevel interference with overlaping channels calculation
// as well as the best signal and its recieved power
type ChanReciever struct {
	Pint     float64
	Pint1lvl float64    // to store total received power level including interference;
	Signal   EmitterInt // to store received power_levels of emitters in channel without co-	
	PrMax    float64
}

func (chR ChanReciever) String() string { return fmt.Sprintf("{%f %f}", chR.Pint*1e15, chR.PrMax*1e15) }


type PhysRecieverInt interface {
	EvalSignalConnection(ch int) (*ChanReciever, float64)
	EvalBestSignalSNR(ch int) (Rc *ChanReciever, eval float64)
	EvalSignalPr(e EmitterInt, ch int) (Pr, K float64)

	EvalSignalSNR(e EmitterInt, ch int) (Rc *ChanReciever, SNR, Pr, K float64)
	EvalSignalBER(e EmitterInt, ch int) (Rc *ChanReciever, BER, SNR, Pr float64)

	MeasurePower(tx EmitterInt)
	SetPos(p geom.Pos)
	GetPos() *geom.Pos

	DoTracking(Connec *list.List) bool

	RicePropagation(E EmitterInt) (fading float64, K float64)
}


// structure to store evaluation of interference at a location
// this has to be initialized with PhysReciever.Init() function to init memory
type PhysReciever struct {
	geom.Pos
	Channels    []ChanReciever
	Orientation []float64 //angle of orientation for beamforming for each channel -1 indicates no beamforming
}

func (r *PhysReciever) Init() {
	r.Channels = make([]ChanReciever, NCh)
	r.Orientation = make([]float64, NCh)
	for i := 0; i < len(r.Orientation); i++ {
		r.Orientation[i] = -1
	}
}

func (r *PhysReciever) SetPos(p geom.Pos) {
	r.Pos = p
}
func (r *PhysReciever) GetPos() *geom.Pos {
	return &r.Pos
}


// function used to evaluate a potential connection of the for the best recieved signal on a channel
func (r *PhysReciever) EvalSignalConnection(ch int) (Rc *ChanReciever, Eval float64) {

	Rc = &r.Channels[ch]
	Eval = -100 //Eval is in [0 inf[, -100 means no signal

	if Rc.Signal != nil {
		_, BER, _, _ := r.EvalSignalBER(Rc.Signal, ch)
		Ptot := Rc.Signal.BERT() + BER
		Eval = Ptot * math.Log(Ptot/BER)
	}

	return

}

func (r *PhysReciever) EvalBestSignalSNR(ch int) (Rc *ChanReciever, Eval float64) {

	Rc = &r.Channels[ch]
	Eval = 0

	if Rc.Signal != nil {

		if ch == 0 {
			Eval = Rc.PrMax / 1e-15 //WNoise
		} else {
			Eval = Rc.PrMax / (Rc.Pint - Rc.PrMax + WNoise)
		}

	}

	return
}


func (r *PhysReciever) EvalSignalPr(e EmitterInt, ch int) (Pr, K float64) {
	var fading float64

	gain := r.GainBeam(e, ch)

	fading, K = r.RicePropagation(e)

	return e.GetPower() * fading * gain, K
}

func (r *PhysReciever) EvalSignalSNR(e EmitterInt, ch int) (Rc *ChanReciever, SNR float64, Pr float64, K float64) {

	Rc = &r.Channels[ch]
	SNR = 0
	Pr, K = r.EvalSignalPr(e, ch)

	switch {
	case ch == 0: //this channel is the obsever channel to follow mobiles while they are not assigned a channel
		SNR = Pr / 1e-15 //WNoise			
	case ch == e.GetCh(): // same channel so substract Pr from Pint
		SNR = Pr / (Rc.Pint - Pr + WNoise)
	default: // different channel so Pr is not in the sum Pint
		SNR = Pr / (Rc.Pint + WNoise)
	}

	return

}

func (r *PhysReciever) EvalSignalBER(e EmitterInt, ch int) (Rc *ChanReciever, BER float64, SNR float64, Pr float64) {
	var K float64
	Rc, SNR, Pr, K = r.EvalSignalSNR(e, ch)

	sigma := SNR / (K + 1.0)
	musqr := SNR - sigma
	eta := 1.0/sigma + 1.0/L2

	BER = math.Exp(-musqr/sigma) / (sigma * eta) * math.Exp(musqr/(sigma*sigma*eta))
	BER = math.Log10(BER)

	return Rc, BER, SNR, Pr
}


// first level interference calculation for all channels. internal function
func (rx *PhysReciever) measurePowerFromChannel(em EmitterInt) {

	for i := 0; i < NCh; i++ {
		rx.Channels[i].Pint1lvl = 0
		rx.Channels[i].Signal = nil
		rx.Channels[i].Pint = 0
		rx.Channels[i].PrMax = 0

		for tx := SystemChan[i].Emitters.Front(); tx != nil; tx = tx.Next() {
			txx := tx.Value.(EmitterInt)

			if em != txx {
				P, _ := rx.EvalSignalPr(txx, i)

				if P > rx.Channels[i].PrMax {
					rx.Channels[i].Signal = txx
					rx.Channels[i].PrMax = P

				}

				rx.Channels[i].Pint1lvl += P
			}
		}

	}

}

// Evaluates interference for all channels with overlapping effect,
// channel 0 is considered to have no interference as traffic is suppose to only hold minimal signalization 
func (rx *PhysReciever) MeasurePower(tx EmitterInt) {

	rx.measurePowerFromChannel(tx)

	for i := 0; i < NCh; i++ {
		rx.Channels[i].Pint = rx.Channels[i].Pint1lvl
		for _, coc := range SystemChan[i].coIntC {
			rx.Channels[i].Pint += coc.factor * rx.Channels[coc.c].Pint1lvl
		}
	}

	for i := 0; i < NChRes; i++ {
		rx.Channels[i].Pint = rx.Channels[i].PrMax + WNoise //just a bit more noise 
	}

}


func (rx *PhysReciever) GainBeam(tx EmitterInt, ch int) float64 {

	if ch > 0 && rx.Orientation[ch] >= 0 {

		p := tx.GetPos().Minus(rx.Pos)
		theta := math.Atan2(p.Y, p.X) * 180 / math.Pi
		if theta < 0 {
			theta += 360
		}
		theta -= rx.Orientation[ch]
		theta = math.Remainder(theta, 360)
		//		t1:=o-theta
		//		t2:=theta-o
		//		if t1>t2 {theta=t1} else {theta=t2}			
		if theta > 180 {
			theta += 360
		}

		if theta < -180 || theta > 180 {
			fmt.Println("ThetaError")
		}

		g := 12 * (theta / 65) * (theta / 65)
		if g > 20 {
			g = 20
		}

		g = math.Pow(10, (-g+10)/10)

		return g
	}

	return 1.0

}


func (rx *PhysReciever) RicePropagation(E EmitterInt) (fading float64, K float64) {
	d := rx.Distance(E.GetPos())
	a := d + 2
	a = a * a
	a = a * a
	fading = (1.0 + rx.SlowFading(E.GetPos())) / a
	K = 1 / (d + 1)
	//K=0;
	return
}

func (rx *PhysReciever) SlowFading(E *geom.Pos) (c float64) {
	//	a:=math.Sin(E.X/math.Pi/300.0) 
	//	b:=math.Cos(E.Y/math.Pi/320.0)
	//	return a*a * b*b  +.02;
	return 0.0
}

func (rx *PhysReciever) DoTracking(Connec *list.List) bool {

	if SetTracking {
		for i := 0; i < len(rx.Orientation); i++ {
			rx.Orientation[i] = -1
		}
		for e := Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			if c.GetCh() > 0 {
				p := c.GetE().GetPos().Minus(rx.Pos)
				theta := math.Atan2(p.Y, p.X) * 180 / math.Pi
				if theta < 0 {
					theta = theta + 360
				}
				rx.Orientation[c.GetCh()] = theta //+ (dbs.Rgen.Float64()*30-15)
			}
		}
	}

	return SetTracking
}

