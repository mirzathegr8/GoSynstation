package synstation

// #include "math.h"
//import "C"

import "math"
import "geom"
import "container/list"
import "fmt"
import "rand"
//import "sselib"


// Structure to hold interference level, and multilevel interference with overlaping channels calculation
// as well as the best signal and its recieved power
type ChanReceiver struct {
	Pint     float64
	Pint1lvl float64    // to store total received power level including interference;
	Signal   EmitterInt // to store received power_levels of emitters in channel without co-	
	PrMax    float64
}

func (chR ChanReceiver) String() string { return fmt.Sprintf("{%f %f}", chR.Pint*1e15, chR.PrMax*1e15) }


type PhysReceiverInt interface {
	Init(r *rand.Rand)
	EvalSignalConnection(ch int) (*ChanReceiver, float64)
	EvalBestSignalSNR(ch int) (Rc *ChanReceiver, eval float64)
	EvalSignalPr(e EmitterInt, ch int) (Pr, K float64)

	EvalSignalSNR(e EmitterInt, ch int) (Rc *ChanReceiver, SNR, Pr, K float64)
	EvalSignalBER(e EmitterInt, ch int) (Rc *ChanReceiver, BER, SNR, Pr float64)

	MeasurePower(tx EmitterInt)
	SetPos(p geom.Pos)
	GetPos() *geom.Pos

	DoTracking(Connec *list.List) bool

	RicePropagation(E EmitterInt) (fading float64, K float64)
}


// structure to store evaluation of interference at a location
// this has to be initialized with PhysReceiver.Init() function to init memory
type PhysReceiver struct {
	geom.Pos
	Channels    []ChanReceiver
	Orientation []float64 //angle of orientation for beamforming for each channel -1 indicates no beamforming
	shadow      shadowMapInt
}

func (r *PhysReceiver) Init(Rgen *rand.Rand) {
	r.Channels = make([]ChanReceiver, NCh)
	r.Orientation = make([]float64, NCh)
	for i := 0; i < len(r.Orientation); i++ {
		r.Orientation[i] = -1
	}
	switch SetShadowMap {
	case NOSHADOW:
		r.shadow = new(noshadow)
	case SHADOWMAP:
		r.shadow = new(shadowMap)
	}
	r.shadow.Init(corr_res, Rgen)
}

func (r *PhysReceiver) SetPos(p geom.Pos) {
	r.Pos = p
}
func (r *PhysReceiver) GetPos() *geom.Pos {
	return &r.Pos
}


// function used to evaluate a potential connection of the for the best recieved signal on a channel
func (r *PhysReceiver) EvalSignalConnection(ch int) (Rc *ChanReceiver, Eval float64) {

	Rc = &r.Channels[ch]
	Eval = -100 //Eval is in [0 inf[, -100 means no signal

	if Rc.Signal != nil {
		_, BER, _, _ := r.EvalSignalBER(Rc.Signal, ch)
		Ptot := Rc.Signal.BERT() + BER
		Eval = Ptot * math.Log(Ptot/BER)
	}

	return

}

func (r *PhysReceiver) EvalBestSignalSNR(ch int) (Rc *ChanReceiver, Eval float64) {

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


func (r *PhysReceiver) EvalSignalPr(e EmitterInt, ch int) (Pr, K float64) {
	var fading float64

	gain := r.GainBeam(e, ch)

	fading, K = r.RicePropagation(e)

	return e.GetPower() * fading * gain, K
}

func (r *PhysReceiver) EvalSignalSNR(e EmitterInt, ch int) (Rc *ChanReceiver, SNR float64, Pr float64, K float64) {

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

func (r *PhysReceiver) EvalSignalBER(e EmitterInt, ch int) (Rc *ChanReceiver, BER float64, SNR float64, Pr float64) {
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
func (rx *PhysReceiver) measurePowerFromChannel(em EmitterInt) {

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
func (rx *PhysReceiver) MeasurePower(tx EmitterInt) {

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


func (rx *PhysReceiver) GainBeam(tx EmitterInt, ch int) float64 {

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


func (rx *PhysReceiver) RicePropagation(E EmitterInt) (fading float64, K float64) {
	d := rx.Distance(E.GetPos())
	a := d + 2
	a = a * a
	a = a * a
	fading = (1.0 * rx.SlowFading(E.GetPos())) / a
	K = 1 / (d + 1)
	//K=0;
	return
}

func (rx *PhysReceiver) SlowFading(E *geom.Pos) (c float64) {
	return rx.shadow.evalShadowFading(E.Minus(rx.Pos))
}

func (rx *PhysReceiver) DoTracking(Connec *list.List) bool {

	if SetReceiverType == BEAM {
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
		return true
	}

	return false
}


type shadowMapInt interface {
	Init(corr float64, Rgen *rand.Rand)
	evalShadowFading(d geom.Pos) (val float64)
}


type noshadow struct{}

func (s *noshadow) Init(f float64, rgen *rand.Rand) {
}
func (s *noshadow) evalShadowFading(d geom.Pos) (val float64) {
	return 1.0
}

type shadowMap struct {
	xcos []float64
	ycos []float64

	xsin []float64
	ysin []float64

	power float64

	smap [][]float32
}
/*
type float32x4 struct {
	A, B, C, D float32
}

func (f float32x4) mult(f float32x4) float32 {

}*/


var mapres = mapsize / float64(maplength)

func (s *shadowMap) Init(corr_dist float64, Rgen2 *rand.Rand) {

	nval := int(Field / corr_dist / shadow_sampling)

	//fmt.Println(" shadowMap down Sampling ", shadow_sampling, " ", nval)

	s.xcos = make([]float64, nval)
	s.ycos = make([]float64, nval)
	s.xsin = make([]float64, nval)
	s.ysin = make([]float64, nval)

	for i := 0; i < nval; i++ {
		s.xcos[i] = Rgen2.NormFloat64()
		//s.xcos[i] *= s.xcos[i]
		s.ycos[i] = Rgen2.NormFloat64()

		s.xsin[i] = Rgen2.Float64() * 2 * math.Pi
		s.ysin[i] = Rgen2.Float64() * 2 * math.Pi

		if s.xcos[i] < mval {
			s.xcos[i] = 0
		}
		if s.ycos[i] < mval {
			s.ycos[i] = 0
		}
		s.power += s.xcos[i] * s.xcos[i]
		s.power += s.ycos[i] * s.ycos[i]

	}

	s.power = math.Sqrt(s.power) / shadow_deviance

	for i := 0; i < nval; i++ {
		s.xcos[i] /= s.power
		s.ycos[i] /= s.power
	}

	s.smap = make([][]float32, mapsize)
	for i := 0; i < mapsize; i++ {
		s.smap[i] = make([]float32, mapsize)
		x := (float64(i) - mapsize/2) / mapres
		for j := 0; j < mapsize; j++ {
			d := geom.Pos{x, (float64(j) - mapsize/2) / mapres}
			s.smap[i][j] = float32(s.evalShadowFadingDirect(d))
			//lets not have -Inf here			
			if s.smap[i][j] < 0.0000001 {
				s.smap[i][j] = 0.0000001
			}
		}
	}

}

func (s *shadowMap) interpolFading(d geom.Pos) (val float64) {

	x := int(d.X*mapres + mapsize/2)
	y := int(d.Y*mapres + mapsize/2)

	if x < 0 || x >= mapsize || y < 0 || y >= mapsize {
		return 1.0
	}

	return float64(s.smap[x][y])

}

var facr = float64(2.0 * math.Pi / Field)

func (s *shadowMap) evalShadowFading(d geom.Pos) (val float64) {
	return s.interpolFading(d)
	//return s.evalShadowFadingDirect(d)
}

func (s *shadowMap) evalShadowFadingDirect(d geom.Pos) float64 {

	posx := float64(d.X) * facr * shadow_sampling
	posy := float64(d.Y) * facr * shadow_sampling
	var rx, ry, val float64
	for i := 0; i < len(s.xcos); i++ {
		rx += posx
		ry += posy
		val += s.xcos[i]*(math.Cos((rx + s.xsin[i]))) + s.ycos[i]*(math.Cos((ry + s.ysin[i])))
	}

	return math.Pow(10, val/10)
}

