package synstation

// #include "math.h"
//import "C"

import "math"
import "geom"
import "container/list"
import "fmt"
import "rand"
//import "sselib"

const SizeES = 10

// Structure to hold interference level, and multilevel interference with overlaping channels calculation
// as well as the best signal and its recieved power
type ChanReceiver struct {
	Pint     float64 // to store total received power level including interference;
	Pint1lvl float64 // to store received power_levels of emitters in channel without co-interference 


	Signal [SizeES]int

	pr [M]float64 //stores received power

}

func (chR *ChanReceiver) Clear() {
	for i := range chR.Signal {
		chR.Signal[i] = -1
	}
	chR.Pint = 0
	chR.Pint1lvl = 0
}

func (chR *ChanReceiver) Push(S int, P float64, R *PhysReceiver) {
	var i = 0
	var j = 0
	for i = 0; i < SizeES; i++ {
		if chR.Signal[i] < 0 {
			chR.Signal[i] = S
			return
		}
		if chR.pr[chR.Signal[i]] < P {
			break
		}
	}
	if i < SizeES {
		for j = i + 1; j < SizeES-1; j++ {
			if chR.Signal[j] < 0 {
				break
			}
			chR.Signal[j] = chR.Signal[j-1]
		}
		chR.Signal[i] = S
	}

}

func (chR ChanReceiver) String() string { return fmt.Sprintf("{%f }", chR.Pint*1e15) }


type PhysReceiverInt interface {
	Init(p geom.Pos, r *rand.Rand)
	EvalSignalConnection(rb int) (*ChanReceiver, float64, float64)
	EvalBestSignalSNR(rb int) (Rc *ChanReceiver, eval float64)
	EvalSignalBER(e EmitterInt, rb int) (Rc *ChanReceiver, BER, SNR, Pr float64)
	EvalSignalSNR(e EmitterInt, rb int) (Rc *ChanReceiver, SNR, Pr, K float64)

	MeasurePower(tx EmitterInt)

	SetPos(p geom.Pos)
	GetPos() geom.Pos

	DoTracking(Connec *list.List) bool

	GenFastFading()

	EvalChRSignalSNR(rb int, k int) (Rc *ChanReceiver, eval float64)

	GetPr(i, rb int) (p float64, Rc *ChanReceiver)
	GetK(i int) (p float64)
}


// structure to store evaluation of interference at a location
// this has to be initialized with PhysReceiver.Init() function to init memory
type PhysReceiver struct {
	geom.Pos
	Channels    []ChanReceiver
	Orientation []float64 //angle of orientation for beamforming for each channel -1 indicates no beamforming
	shadow      shadowMapInt
	Rgen        *rand.Rand

	kk [M]float64 //stores received power
}

func (r *PhysReceiver) Init(p geom.Pos, Rgen *rand.Rand) {
	r.Pos = p
	r.Rgen = Rgen
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
func (r *PhysReceiver) GetPos() geom.Pos {
	return r.Pos
}


// function used to evaluate a potential connection of the for the best recieved signal on a channel
func (r *PhysReceiver) EvalSignalConnection(rb int) (Rc *ChanReceiver, Eval, BER float64) {

	Rc = &r.Channels[rb]
	Eval = -100 //Eval is in [0 inf[, -100 means no signal

	if Rc.Signal[0] >= 0 {
		E := &Mobiles[Rc.Signal[0]].Emitter
		_, BER, _, _ = r.EvalSignalBER(E, rb)
		BER = math.Log10(BER)
		Ptot := E.BERT() + BER
		Eval = Ptot * math.Log(Ptot/BER)
	}

	return

}

func (r *PhysReceiver) EvalBestSignalSNR(rb int) (Rc *ChanReceiver, Eval float64) {

	Rc = &r.Channels[rb]
	Eval = 0

	if Rc.Signal[0] >= 0 {

		PMax := Rc.pr[Rc.Signal[0]]
		if rb == 0 {
			Eval = PMax / WNoise
		} else {
			Eval = PMax / (Rc.Pint - PMax + WNoise)
		}

	}

	return
}


func (r *PhysReceiver) EvalChRSignalSNR(rb int, k int) (Rc *ChanReceiver, Eval float64) {

	Rc = &r.Channels[rb]
	Eval = 0

	if Rc.Signal[k] >= 0 {

		PMax := Rc.pr[Rc.Signal[k]]
		if rb == 0 {
			Eval = PMax / WNoise
		} else {
			Eval = PMax / (Rc.Pint - PMax + WNoise)
		}

	}

	return
}


func (r *PhysReceiver) EvalSignalSNR(e EmitterInt, rb int) (Rc *ChanReceiver, SNR float64, Pr float64, K float64) {

	Rc = &r.Channels[rb]
	SNR = 0
	K = r.kk[e.GetId()]

	if e.IsSetARB(rb) {
		Pr = Rc.pr[e.GetId()]

	} else { //we supose we will get the same power out of the other channel (fading aside)
		RcO := &r.Channels[e.GetFirstRB()]
		Pr = RcO.pr[e.GetId()]
	}
	switch {
	case rb == 0: //this channel is the obsever channel to follow Mobiles while they are not assigned a channel
		SNR = Pr / 1e-15 //WNoise			
	case e.IsSetARB(rb): // same channel so substract Pr from Pint
		SNR = Pr / (Rc.Pint - Pr + WNoise)
	default: // different channel so Pr is not in the sum Pint
		SNR = Pr / (Rc.Pint + WNoise)
	}

	return

}

func (r *PhysReceiver) EvalSignalBER(e EmitterInt, rb int) (Rc *ChanReceiver, BER float64, SNR float64, Pr float64) {

	var K float64
	Rc, SNR, Pr, K = r.EvalSignalSNR(e, rb)

	sigma := SNR / (K + 1.0)
	musqr := SNR - sigma
	eta := 1.0/sigma + 1.0/L2

	BER = math.Exp(-musqr/sigma) / (sigma * eta) * math.Exp(musqr/(sigma*sigma*eta))

	return Rc, BER, SNR, Pr

	return
}


// first level interference calculation for all channels. internal function
func (rx *PhysReceiver) measurePowerFromChannel(em EmitterInt) {

	for i := 0; i < NCh; i++ {
		rx.Channels[i].Clear()
	}

	/*var Ns = 0
	if em != nil {
		Ns = em.GetId()
	} else {
		Ns = -1
	}*/

	/*for i := 0; i < Ns; i++ {
		for rb, use := range Mobiles[i].ARB {
			if use {
				chR := &rx.Channels[rb]
				chR.Pint1lvl += chR.pr[i]
				chR.Push(i, chR.pr[i], rx)
			}
		}
	}*/

	/*for i := Ns + 1; i < M; i++ {
		for rb, use := range Mobiles[i].ARB {
			if use {
				chR := &rx.Channels[rb]
				chR.Push(i, chR.pr[i], rx)
				chR.Pint1lvl += chR.pr[i]
			}
		}
	}*/

	for i := 0; i < NCh; i++ {

		chR := &rx.Channels[i]

		for e := SystemChan[i].Emitters.Front(); e != nil; e = e.Next() {
			c := e.Value.(EmitterInt)
			m := c.GetId()
			chR.Pint1lvl += chR.pr[m]
			chR.Push(m, chR.pr[m], rx)

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
		//	rx.Channels[i].meanPint.Add(rx.Channels[i].Pint)
	}

}


func (rx *PhysReceiver) SlowFading(E geom.Pos) (c float64) {
	return rx.shadow.evalShadowFading(E.Minus(rx.Pos))
}

func (rx *PhysReceiver) DoTracking(Connec *list.List) bool {

	if SetReceiverType == BEAM {
		for i := 0; i < len(rx.Orientation); i++ {
			rx.Orientation[i] = -1
		}
		for e := Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			for rb, use := range c.GetE().GetARB() {
				if use {
					p := c.GetE().GetPos().Minus(rx.Pos)
					theta := math.Atan2(p.Y, p.X)
					if theta < 0 {
						theta = theta + PI2
					}
					rx.Orientation[rb] = theta //+ (dbs.Rgen.Float64()*30-15)
				}
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


var mapres = mapsize / float64(maplength)

func (s *shadowMap) Init(corr_dist float64, Rgen2 *rand.Rand) {

	nval := int(Field / corr_dist / shadow_sampling)

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


const PI2 = 2 * math.Pi

func (rx *PhysReceiver) GenFastFading() {

	for i := 0; i < M; i++ {

		E := &Mobiles[i]

		// inline minus
		p := geom.Pos{E.X - rx.X, E.Y - rx.Y}
		//Calculate Distance, Fading parameter K, and Fading
		//d := rx.DistanceSquare(Mobiles[i].Pos)


		d := (p.X*p.X + p.Y*p.Y)
		d += 2
		K := 1 / d
		d *= d
		fading := rx.shadow.evalShadowFading(p) / d * E.Power

		rx.kk[i] = K

		for rb, use := range E.ARB { // eval power received over each assigned RB
			// Watch out, here we only eval the powers for these RB and we do not set to 0 the other RB
			// such that  in interference evaluation we must only add eval powers (pr[i]) of RB included in E.ARB vector


			if use {

				// Evaluate Beam Gain


				gain := float64(1)
				if rx.Orientation[rb] >= 0 && rb > 0 {

					theta := math.Atan2(p.Y, p.X)

					if theta < 0 {
						theta += PI2
					}
					theta -= rx.Orientation[rb]

					if theta > math.Pi {
						theta -= PI2
					} else if theta < -math.Pi {
						theta += PI2
					}

					if theta < 0.05 && theta > -0.05 {
						gain = 10
					} else {
						theta /= 1.1345
						g := 12 * theta * theta
						if g > 20 {
							g = 20
						}
						gain = math.Pow(10, (-g+10)/10)
					}

				}

				pr := fading * gain

				rx.Channels[rb].pr[i] = pr

			}
		}
	}

}

func (rx PhysReceiver) GetK(i int) (k float64) {
	k = rx.kk[i]
	return
}

func (rx PhysReceiver) GetPr(i, rb int) (p float64, Rc *ChanReceiver) {
	Rc = &rx.Channels[rb]
	p = Rc.pr[i]
	return
}

