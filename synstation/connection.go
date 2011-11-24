package synstation

import "math"
import "geom"
import rand "math/rand"
import cmath "math/cmplx"
//import "fmt"


var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }

type Connection struct {
	E *Emitter
	
	Status int //0 master ,1 slave

	RBsReceiver   

	meanPr   MeanData
	meanSNR  MeanData
	meanBER  MeanData
	meanCapa MeanData

	filterAr [NCh]FilterInt //filter ban to use for channel gain FF generator
	ff_R     [NCh]float64   // stores channel gain for every RB
	SNRrb    [NCh]float64   //stores SNR per RB


	IfilterAr [NCh]FilterInt //filter bank for first interferer
	Iff_R     [NCh]float64   //channel FF gain  for first interer if fading is generated

	filterF FilterInt //Coherence Frequency filter	

	Rgen *rand.Rand

	initz [NRB]complex128 //generation of random number per RB	

}

// for output to save
func (c *Connection) GetInstantSNIR() []float64 {
	return c.ff_R[:]
}

type ConnectionS struct {
	A, B   geom.Pos
	Status int
	BER    float64
	Ch     int
}

func (c *ConnectionS) Copy(cc *Connection) {
	c.A = cc.E.GetPos()
	c.BER = cc.meanBER.Get()
	c.Status = cc.Status
}


func (co *Connection) GetE() *Emitter     { return co.E }
func (co *Connection) GetMeanPr() float64 { return co.meanPr.Get() }


func (co *Connection) BitErrorRate(dbs *DBS) {

	co.evalInstantSINR(&dbs.PhysReceiverBase)

	co.evalInstantBER(dbs)

	co.Status = 1 //let mobile set master state		
	co.E.AddConnection(co, dbs)

}

func (co *Connection) EvalRatio() float64 {
	return co.meanSNR.Get()
}

func (co *Connection) EvalRatioConnect() float64 {
	return co.E.BERT()
}

func (co *Connection) EvalRatioDisconnect() float64 {
	Ptot := co.E.BERT()
	return Ptot * math.Log(Ptot/co.GetLogMeanBER())
}

func (Conn *Connection) InitConnection(E *Emitter, v float64, Rgen *rand.Rand) {

	Conn.E = E
	Conn.meanBER.Clear(v)
	Conn.Status = 1

	Conn.Rgen = Rgen

	Speed := E.GetSpeed()
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000 for RB TTI

	if DopplerF < 0.002 { // the frequency is so low, a simple antena diversity will compensate for 	
		for i := 0; i < NRB; i++ {
			Conn.filterAr[i] = &PNF

			Conn.IfilterAr[i] = &PNF

		}
		Conn.filterF = &PNF
	} else {
		A := Butter(DopplerF)
		B := Cheby(10, DopplerF)
		C := MultFilter(A, B)

		Conn.filterF = CoherenceFilter.Copy()

		for i := 0; i < NRB; i++ {
			Conn.filterAr[i] = C.Copy()
		}

		// initalize filters : sent some values to prevent having empty values z^-n in filters 
		// 1.5/DopplerF gives a good number of itterations to decorelate initial null variables
		// and set the channel to a steady state		

		for l := 0; l < int(2.5/DopplerF); l++ {

			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				Conn.filterF.nextValue(complex(Conn.Rgen.NormFloat64(), Conn.Rgen.NormFloat64()))
			}

			for i := 0; i < NRB; i++ {
				Conn.filterAr[i].nextValue(Conn.filterF.nextValue(complex(Conn.Rgen.NormFloat64(), Conn.Rgen.NormFloat64())))
			}

		}

		if FadingOnPint1 == Fading {
			for i := 0; i < NRB; i++ {
				Conn.IfilterAr[i] = C.Copy()
			}

			// initalize filters : sent some values to prevent having empty values z^-n in filters 
			// 1.5/DopplerF gives a good number of itterations to decorelate initial null variables
			// and set the channel to a steady state		

			for l := 0; l < int(2.5/DopplerF); l++ {

				// for speed optimization, decorelation samples or not used, it makes little difference 
				for i := 0; i < 50; i++ {
					Conn.filterF.nextValue(complex(Conn.Rgen.NormFloat64(), Conn.Rgen.NormFloat64()))
				}

				for i := 0; i < NRB; i++ {
					Conn.IfilterAr[i].nextValue(Conn.filterF.nextValue(
						complex(Conn.Rgen.NormFloat64(), Conn.Rgen.NormFloat64())))
				}

			}
		}

	}

}

func CreateConnection(E *Emitter, v float64, Rgen *rand.Rand) *Connection {
	Conn := new(Connection)
	Conn.InitConnection(E, v, Rgen)
	return Conn
}

func (co *Connection) GetLogMeanBER() float64 {
	return math.Log10(co.meanBER.Get() + 1e-10) //prevent saturation
}

//This function is only called once per iteration, so it is where the FF value is generated
func (c *Connection) evalInstantBER(dbs *DBS) {

	ARB := c.E.GetARB()

	rx:= dbs.PhysReceiverBase
	K := rx.GetK(c.E.GetId())

	//Generate DopplerFading
	//pass some values to decorelate

		for i := 0; i < 50; i++ {
			c.filterF.nextValue(complex(c.Rgen.NormFloat64(), c.Rgen.NormFloat64()))
		}

		for i := 0; i < NCh; i++ {
			c.initz[i] = c.filterF.nextValue(complex(c.Rgen.NormFloat64(), c.Rgen.NormFloat64()))
		}

		for rb := 0; rb < NCh; rb++ {
			a := c.filterAr[rb].nextValue(c.initz[rb]) + complex(K, 0)
			c.ff_R[rb] = (real(a * cmath.Conj(a))) / (2 + K*K)
		}

		if FadingOnPint1 == Fading {
			// Generate fading values for First Interferer
			//
			for i := 0; i < 50; i++ {
				c.filterF.nextValue(complex(c.Rgen.NormFloat64(), c.Rgen.NormFloat64()))
			}

			for i := 0; i < NCh; i++ {
				c.initz[i] = c.filterF.nextValue(complex(c.Rgen.NormFloat64(), c.Rgen.NormFloat64()))
			}

			for rb := 0; rb < NCh; rb++ {
				a := c.filterAr[rb].nextValue(c.initz[rb]) + complex(K, 0)
				c.Iff_R[rb] = (real(a * cmath.Conj(a))) / (2 + K*K)
			}
		}

	

	var touch bool	

	for rb , use:= range ARB {

		Rc := &c.Channels[rb]

		NotPint1 := 0.0
		if FadingOnPint1 == Fading { //if cancel, c.Iff_r =0 and total multiplied by -1, 
						//if fading, remove difference
			if Rc.Signal[1] >= 0 {
				NotPint1 = Rc.pr[Rc.Signal[1]] * (c.Iff_R[rb] - 1)
			}
		} else if FadingOnPint1 == Cancel {
			if Rc.Signal[1] >= 0 {
				NotPint1 = -Rc.pr[Rc.Signal[1]]
			}
		}

		Pr := Rc.pr[c.E.GetId()]

		if use {
			
			c.SNRrb[rb] = Pr * c.ff_R[rb] / (GetNoisePInterference(Rc.Pint, Pr) + NotPint1)

			BER := L1 * math.Exp(-c.SNRrb[rb]/2/L2) / 2.0

			c.meanPr.Add(Pr)
			c.meanSNR.Add(c.SNRrb[rb])
			c.meanBER.Add(BER)			

			touch = true
		} else {
			c.SNRrb[rb] = Pr * c.ff_R[rb] / 
				(GetNoisePInterference(Rc.Pint, 0) + NotPint1) 				

			c.SNRrb[rb] *= estimateFactor(dbs, c.E)
		}
	}

	if !touch { // add null to mean BER
		c.meanPr.Add(0)
		c.meanSNR.Add(0)
		c.meanBER.Add(1)
	}

	return
}

func estimateFactor0(dbs *DBS, E *Emitter) float64 {

	return conservationFactor

}

func estimateFactor1(dbe *DBS, E *Emitter) (o float64) {
	o = 1
	div := E.GetNumARB()
	if div > 1 {
		o = 1 / float64(div)
	}

	o *= conservationFactor 

	return

}


func (co *Connection) evalInstantSINR(rx *PhysReceiverBase) {

	Orientation := rx.AoA[co.E.GetId()]

	for m,E  := range Mobiles{
		
		gain := float64(1)
			
				theta := rx.AoA[m] - Orientation

				if theta > math.Pi {
					theta -= PI2
				} else if theta < -math.Pi {
					theta += PI2
				}

				if math.Abs(theta) < BeamAngle/20 {
					gain = 10
				} else {
					theta /= BeamAngle
					g := 12 * theta * theta
					if g > 20 {
						g = 20
					}
					gain = math.Pow(10, (-g+10)/10)
				}

		for rb, use := range E.ARB { 
			if use {
				// Evaluate Beam Gain		
				co.Channels[rb].pr[m] = rx.Channels[rb].pr[m] * gain 
			}
		}

	}

	co.RBsReceiver.SumInterference()
}

