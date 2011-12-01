package synstation

import "math"
import "geom"
import rand "math/rand"
//import cmath "math/cmplx"
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

	//filterAr FilterBank   //filter ban to use for channel gain FF generator
	filterAr [NCh]FilterInt //filter ban to use for channel gain FF generator
	ff_R     [NCh]float64   // stores channel gain for every RB
	SNRrb    [NCh]float64   //stores SNR per RB

	//IfilterAr FilterBank   //filter bank for first interferer
	IfilterAr [NCh]FilterInt //filter bank for first interferer
	Iff_R     [NCh]float64   //channel FF gain  for first interer if fading is generated

	filterF FilterInt //Coherence Frequency filter	

	Rgen *rand.Rand

	initz [NCh]complex128 //generation of random number per RB	

	paralel_sync chan int

	gainM [M]float64
}

// for output to save
func (co *Connection) GetInstantSNIR() []float64 {
	return co.ff_R[:]
}

type ConnectionS struct {
	A, B   geom.Pos
	Status int
	BER    float64
	Ch     int
}

func (co *ConnectionS) Copy(cc *Connection) {
	co.A = cc.E.GetPos()
	co.BER = cc.meanBER.Get()
	co.Status = cc.Status
}

func (co *Connection) GetE() *Emitter     { return co.E }
func (co *Connection) GetMeanPr() float64 { return co.meanPr.Get() }

func (co *Connection) Interference(dbs *DBS) {
	//co.evalInstantSINR(&dbs.PhysReceiverBase)

	Orientation := dbs.AoA[co.E.GetId()]

	thres := dbs.pr[co.E.Id] / 10000

	for m := range Mobiles {
	//	go func(m int) {
			co.gainM[m] = 0.0
			if dbs.pr[m] > thres {

				theta := dbs.AoA[m] - Orientation

				if theta > math.Pi {
					theta -= PI2
				} else if theta < -math.Pi {
					theta += PI2
				}
				theta = math.Abs(theta)
				if theta < BeamAngle/20 {
					co.gainM[m] = 10
				} else {
					theta /= BeamAngle
					g := 0.34641 * theta
					if g > 0.44721 {
						g = 0.44721
					}

					/*tmp:= (math.Float64bits(10) >>32 ) -1072632447;
					tmp2:= math.Float64bits(-g* (math.Float64frombits(tmp)) + 1072632447 )
					gain=20*math.Float64frombits( tmp2<<32)*/
					co.gainM[m] = 20 * math.Pow(10, -g)

				}
			}
	//		co.paralel_sync <- 1
	//	}(m)
	}
	//for _ = range Mobiles {	<-co.paralel_sync}

	for rb := range co.Channels {
		chR := &co.Channels[rb]
		//go func(chR *ChanReceiver, rb int, co *Connection) {
			chR.Clear()
			for m := range Mobiles {
				//if Mobiles[m].ARB[rb] {
					// Evaluate Beam Gain	
					f := dbs.Channels[rb].pr[m] * co.gainM[m]
					chR.pr[m] = f
				//	chR.Pint += f
				/*	if f > chR.Pmax {
						chR.Pmax = f
						chR.Signal[0] = m
					}*/
				//}
			}
			//co.paralel_sync<-1
		//}(chR, rb,co)


	}
	//for _ = range co.Channels {  <- co.paralel_sync}
	
	co.SumInterference()

}

func (co *Connection) Fading(dbs *DBS) {
	//co.evalInstantBER(dbs)

	//Generate DopplerFading
	//pass some values to decorelate

	//refactored not functional
	/*for i := 0; i < 50; i++ {
		co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
	}
	co.filterF.nextValues(co.initz[0:50])

	for i := 0; i < NCh; i++ {
		co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
	}
	co.filterF.nextValues(co.initz[0:NCh])

	/*	if FadingOnPint1 == Fading {
		// Generate fading values for First Interferer
		//
		for i := 0; i < 50; i++ {
			co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		for i := 0; i < NCh; i++ {
			co.initz[i] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		for rb := 0; rb < NCh; rb++ {
			a := co.filterAr[rb].nextValue(co.initz[rb]) + complex(K, 0)
			co.Iff_R[rb] = (real(a * cmath.Conj(a))) / (2 + K*K)
		}
	}*/

	for i := 0; i < 50*corrF; i++ {
		co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
	}

	for i := 0; i < NCh; i++ {
		co.initz[i] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
	}

	K := dbs.GetK(co.E.Id)

	for rb := range co.ff_R { //0; rb < NCh; rb++ {
		//go func(rb int) {
			a := co.filterAr[rb].nextValue(co.initz[rb]) //+ complex(K, 0)
			ar := real(a) + K
			ai := imag(a)
			co.ff_R[rb] = (ar*ar + ai*ai) / (2 + K*K) //(real(a * cmath.Conj(a))) / (2 + K*K)
			//co.paralel_sync <- 1
		//}(rb)
	}
	//for _ = range co.ff_R {<-co.paralel_sync}

	/*if FadingOnPint1 == Fading {
		// Generate fading values for First Interferer
		//
		for i := 0; i < 50; i++ {
			co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		for i := 0; i < NCh; i++ {
			co.initz[i] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		for rb := 0; rb < NCh; rb++ {
			a := co.filterAr[rb].nextValue(co.initz[rb]) + complex(K, 0)
			co.Iff_R[rb] = (real(a * cmath.Conj(a))) / (2 + K*K)
		}
	}*/

}

func (co *Connection) Fading2(dbs *DBS) {

	/*K := dbs.GetK(co.E.Id)

	co.filterAr.nextValues(&co.initz) 	

	var ar,ai float64

	for rb, v := range co.initz {
		ar= real(v) + K
		ai = imag(v) 
		co.ff_R[rb] = (ar*ar+ai*ai) / (2 + K*K)
	}*/
}

func (co *Connection) SNR(dbs *DBS) {
	var touch bool

	for rb, use := range co.E.ARB {

		Rc := &co.Channels[rb]

		NotPint1 := 0.0
		/*	if FadingOnPint1 == Fading { //if cancel, co.Iff_r =0 and total multiplied by -1, 
				//if fading, remove difference
				if Rc.Signal[1] >= 0 {
					NotPint1 = Rc.pr[Rc.Signal[1]] * (co.Iff_R[rb] - 1)
				}
			} else if FadingOnPint1 == Cancel {
				if Rc.Signal[1] >= 0 {
					NotPint1 = -Rc.pr[Rc.Signal[1]]
				}
			}*/

		if use {

			Pr := Rc.pr[co.E.Id]

			co.SNRrb[rb] = Pr * co.ff_R[rb] / (GetNoisePInterference(Rc.Pint, Pr) + NotPint1)

			BER := L1 * math.Exp(-co.SNRrb[rb]/2/L2) / 2.0

			co.meanPr.Add(Pr)
			co.meanSNR.Add(co.SNRrb[rb])
			co.meanBER.Add(BER)

			touch = true
		} else {
			co.SNRrb[rb] = dbs.pr[co.E.Id] * co.ff_R[rb] /
				(GetNoisePInterference(Rc.Pint, 0) + NotPint1)

			co.SNRrb[rb] *= estimateFactor(dbs, co.E)
		}
	}

	if !touch { // add null to mean BER
		co.meanPr.Add(0)
		co.meanSNR.Add(0)
		co.meanBER.Add(1)
	}

}

func (co *Connection) BitErrorRate(dbs *DBS) {

	co.Interference(dbs)
	co.Fading(dbs)
	//co.Fading2(dbs)
	co.SNR(dbs)
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

func (co *Connection) InitConnection(E *Emitter, v float64, Rgen *rand.Rand) {

	co.E = E
	co.meanBER.Clear(v)
	co.Status = 1

	co.Rgen = Rgen

	Speed := E.GetSpeed()
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000 for RB TTI

	if DopplerF < 0.002 { // the frequency is so low, a simple antena diversity will compensate for 	
		DopplerF = 0.002
	}

	A := Butter(DopplerF)
	B := Cheby(10, DopplerF)
	C := MultFilter(A, B)

	co.filterF = CoherenceFilter.Copy()

	for i := 0; i < NCh; i++ {
		co.filterAr[i] = C.Copy()
	}
	//	co.filterAr.Build(C)

	// initalize filters : sent some values to prevent having empty values z^-n in filters 
	// 1.5/DopplerF gives a good number of itterations to decorelate initial null variables
	// and set the channel to a steady state		

	//refactored not functional
	/*for l := 0; l < int(2.5/DopplerF); l++ {


		// for speed optimization, decorelation samples or not used, it makes little difference 
		for i := 0; i < 50; i++ {
			co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
		}
		co.filterF.nextValues(co.initz[0:50])

		for i := 0; i < NCh; i++ {
			co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
		}
		co.filterF.nextValues(co.initz[0:NCh])		

		co.filterAr.nextValues(&co.initz)


	}

	if FadingOnPint1 == Fading {
			co.IfilterAr.Build(C)	


		// initalize filters : sent some values to prevent having empty values z^-n in filters 
		// 1.5/DopplerF gives a good number of itterations to decorelate initial null variables
		// and set the channel to a steady state		

		for l := 0; l < int(2.5/DopplerF); l++ {

			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
			}
			co.filterF.nextValues(co.initz[0:50])

			for i := 0; i < NCh; i++ {
				co.initz[i] = complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())
			}
			co.filterF.nextValues(co.initz[0:NCh])		

			co.IfilterAr.nextValues(&co.initz)

		}
	}*/

	for l := 0; l < int(2.5/DopplerF); l++ {

		// for speed optimization, decorelation samples or not used, it makes little difference 
		for i := 0; i < 50; i++ {
			co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}
		for i := 0; i < NCh; i++ {
			co.initz[i] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		for rb := range co.ff_R { //0; rb < NCh; rb++ {
		//	go func(rb int) {
				co.filterAr[rb].nextValue(co.initz[rb])
		//		co.paralel_sync <- 1
		//	}(rb)
		}
		//for _ = range co.ff_R {<-co.paralel_sync}

		/*for i := 0; i < NCh; i++ {
			co.filterAr[i].nextValue(co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())))
		}*/

	}

	/*if FadingOnPint1 == Fading {
		for i := 0; i < NCh; i++ {
			co.IfilterAr[i] = C.Copy()
		}

		// initalize filters : sent some values to prevent having empty values z^-n in filters 
		// 1.5/DopplerF gives a good number of itterations to decorelate initial null variables
		// and set the channel to a steady state		

		for l := 0; l < int(2.5/DopplerF); l++ {

			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
			}

			for i := 0; i < NCh; i++ {
				co.IfilterAr[i].nextValue(co.filterF.nextValue(
					complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())))
			}

		}
	}*/

}

func NewConnection() (Conn *Connection) {

	Conn = new(Connection)
	Conn.paralel_sync = make(chan int, 10000)
	return

}

/*func CreateConnection(E *Emitter, v float64, Rgen *rand.Rand) *Connection {
	Conn := new(Connection)
	Conn.InitConnection(E, v, Rgen)
	return Conn
}*/

func (co *Connection) GetLogMeanBER() float64 {
	return math.Log10(co.meanBER.Get() + 1e-10) //prevent saturation
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

//   Reformatted by   lerouxp    Tue Nov 29 12:47:29 EST 2011

//   Reformatted by   lerouxp    Thu Dec 1 09:57:20 EST 2011

