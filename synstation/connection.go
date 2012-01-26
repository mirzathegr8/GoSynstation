package synstation

import "math"
import "geom"
import rand "rand"
//import cmath "math/cmplx"
//import "fmt"


var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }

type Connection struct {
	E *Emitter
	IdB int //Id of the base station

	Status int //0 master ,1 slave

	RBsReceiver

	meanPr   MeanData
	meanSNR  MeanData
	meanBER  MeanData
	meanCapa MeanData

	//filterAr FilterBank   //filter ban to use for channel gain FF generator
	filterAr [NCh]*Filter //filter ban to use for channel gain FF generator
	ff_R     [NCh]float64 // stores channel gain for every RB
	SNRrb    [NCh]float64 //stores SNR per RB

	//IfilterAr FilterBank   //filter bank for first interferer
	IfilterAr [NCh]*Filter //filter bank for first interferer
	Iff_R     [NCh]float64 //channel FF gain  for first interer if fading is generated

	filterF *Filter //Coherence Frequency filter	

	Rgen *rand.Rand

	initz [NCh]complex128 //generation of random number per RB	

	//ComplexRand chan complex128

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
	co.A = cc.E.Pos
	co.BER = cc.meanBER.Get()
	co.Status = cc.Status
}

func (co *Connection) GetMeanPr() float64 { return co.meanPr.Get() }

func (co *Connection) BitErrorRate(dbs *DBS) {

	Orientation := dbs.AoA[co.E.Id]	

	//var sector int
	if SetReceiverType==SECTORED{
		if Orientation< PI/3 || Orientation > PI2*5/6{
			Orientation=0
		} else if Orientation<PI{
			Orientation=PI2/3
		} else {
			Orientation=4*PI/3
		}

	}


	for m := range Mobiles {
		switch SetReceiverType{

		case BEAM:
			fallthrough
		case SECTORED:
		co.gainM[m] = 0.0		

		theta := dbs.AoA[m] - Orientation

		if theta > math.Pi {
			theta -= PI2
		} else if theta < -math.Pi {
			theta += PI2
		}
		//theta = math.Fabs(theta)

		if theta < BeamAngle/20 && theta > BeamAngle/20 {
			co.gainM[m] = 10
		} else {
			theta /= BeamAngle
			g := 12 * theta * theta
			if g > 20 {
				g = 20
			}
			co.gainM[m] = math.Pow(10, (-g+10)/10)
		}

		case OMNI: 
			co.gainM[m]=1.0
	
		}
	
	}

	MasterConnectedArray := dbs.GetCancelation()

	if co.Status == 0 {
		for rb := range co.Channels {
			chR := &co.Channels[rb]
			chR.Clear()

			if InterferenceCancel== SIZEESCANCELATION {MasterConnectedArray = dbs.GetCancelationRB(rb)}

			for m, g := range co.gainM { //Mobiles {
				if Mobiles[m].ARB[rb] && !(MasterConnectedArray[m] && m!= co.E.Id){
					// Evaluate Beam Gain	
					f := dbs.Channels[rb].pr[m] * g //co.gainM[m]
					chR.pr[m] = f
					chR.Pint += f
					if f > chR.Pmax {
						chR.Pmax = f
						chR.Signal[0] = m
					}
				}
			}
		}
	} else { //we do not evaluate interference on unsused RB if we are not master, as this information is only usefull to scheduling algos
		// however, if at the fetchdata call on mobiles it changes master connection, then the sinr information will be used from this calculation and will be wrong : need to add a delay before the new master dbs considers the link master and allocates rb to that emitter
		for rb, use := range co.E.ARB {
			if use {
				chR := &co.Channels[rb]
				chR.Clear()

				if InterferenceCancel== SIZEESCANCELATION {MasterConnectedArray = dbs.GetCancelationRB(rb)}

				for m, g := range co.gainM { //Mobiles {
					//if g>0 {//dbs.pr[m] > thres {
					if Mobiles[m].ARB[rb]  && !(MasterConnectedArray[m] && m!= co.E.Id) {
						// Evaluate Beam Gain	
						f := dbs.Channels[rb].pr[m] * g //co.gainM[m]
						chR.pr[m] = f
						chR.Pint += f
						if f > chR.Pmax {
							chR.Pmax = f
							chR.Signal[0] = m
						}
					}
					//}
				}
			}
		}

	}

	


	for i := 0; i < 50*corrF; i++ {
		co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))	
	}

	for rb := range co.filterAr {
		co.initz[rb] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))		
	}
	
	for rb := range co.ff_R { 
		co.initz[rb] = co.filterAr[rb].nextValue(co.initz[rb])
	}
	K:=dbs.kk[co.E.Id]
	for rb := range co.ff_R {	
		ar := real(co.initz[rb]) + K
		ai := imag(co.initz[rb])
		co.ff_R[rb] = (ar*ar + ai*ai) / (2 + K*K) //(real(a * cmath.Conj(a))) / (2 + K*K)
	}




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
				(GetNoisePInterference(Rc.Pint, 0) + NotPint1) * conservationFactor

			//co.SNRrb[rb] *= estimateFactor(dbs, co.E)
		}
	}

	if !touch { // add null to mean BER
		co.meanPr.Add(0)
		co.meanSNR.Add(0)
		co.meanBER.Add(1)
	}



	
	co.Status = 1 //let mobile set master state		
	co.E.AddConnection(co, dbs)

}

func (co *Connection) EvalRatio() float64 {
	return co.meanSNR.Get()
}

func (co *Connection) EvalRatioConnect() float64 {
	return co.E.BERtotal
}

func (co *Connection) EvalRatioDisconnect() float64 {
	Ptot := co.E.BERtotal
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
	

	/*
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
			co.filterAr[i].nextValue(co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())))
		}

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

func (co *Connection) clear(){
	// free some memory . perhaps need to rethink this and have a filterbank
	for rb :=range co.filterAr {
		co.filterAr[rb]=nil
		co.IfilterAr[rb]=nil
	}

}

func NewConnection(i int) (Conn *Connection) {
	Conn = new(Connection)
	Conn.IdB=i
	return
}


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

