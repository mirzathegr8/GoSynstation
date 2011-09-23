package synstation

import "math"
import "geom"
import "rand"
import "cmath"
//import "fmt"


var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }


type Connection struct {
	E *Emitter

	Pr     float64 // to store power level received	
	SNR, K float64
	BER    float64

	Status int //0 master ,1 slave

	meanPr   MeanData
	meanSNR  MeanData
	meanBER  MeanData
	meanCapa MeanData

	filterAr [NRB]FilterInt //filter ban to use for channel gain FF generator
	ff_R     [NCh]float64   // stores channel gain for every RB
	SNRrb    [NCh]float64   //stores SNR per RB


	IfilterAr [NRB]FilterInt //filter bank for first interferer
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
	c.BER = cc.BER
	c.Status = cc.Status
}

type ConnecType interface {
	BitErrorRate(rx *PhysReceiver, dbs *DBS)
	EvalRatioConnect() float64
	EvalRatioDisconnect() float64
	GetE() *Emitter
	GetPr() float64
	EvalRatio(rx PhysReceiverInt) float64
	GetSNR() float64

	GetInstantSNIR() []float64
}

func (co *Connection) GetE() *Emitter { return co.E }
func (co *Connection) GetPr() float64   { return co.Pr }
func (co *Connection) GetSNR() float64  { return co.SNR }


func (co *Connection) BitErrorRate(rx *PhysReceiver, dbs *DBS) {

	co.evalInstantBER(co.GetE(), rx, dbs)

	co.Status = 1 //let mobile set master state		
	co.E.AddConnection(co,dbs)

}

func (co *Connection) EvalRatio(rx PhysReceiverInt) float64 {
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
				Conn.filterAr[i].nextValue(Conn.filterF.nextValue(complex(Conn.Rgen.NormFloat64(),Conn.Rgen.NormFloat64() ) ) )
			}
			
		}

		if FadingOnPint1==Fading{
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
	
				for i := 0; i < NCh; i++ {
					Conn.IfilterAr[i].nextValue(Conn.filterF.nextValue(
						complex(Conn.Rgen.NormFloat64() , Conn.Rgen.NormFloat64() ) ) )
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
func (c *Connection) evalInstantBER(E *Emitter, rx *PhysReceiver, dbs *DBS) {

	ARB := E.GetARB()

	c.SNR = 0 //reset

	K := rx.GetK(E.GetId())

	//Generate DopplerFading
	//pass some values to decorelate

	for tdma:=0;tdma<NCh-NChRes; tdma+=NRB{
	
	for i := 0; i < 50; i++ {
		c.filterF.nextValue(complex(c.Rgen.NormFloat64(),c.Rgen.NormFloat64()))
	}

	for i := 0; i < NRB; i++ {
		c.initz[i] = c.filterF.nextValue(complex(c.Rgen.NormFloat64(),c.Rgen.NormFloat64()))
	}
	
	for rb := 0; rb < NRB; rb++ {
		a := c.filterAr[rb].nextValue(c.initz[rb]) + complex(K,0)
		c.ff_R[rb+tdma] = (real(a*cmath.Conj(a)) ) / (2 + K*K)
	}


	if FadingOnPint1==Fading{
	// Generate fading values for First Interferer
	//
	for i := 0; i < 50; i++ {
		c.filterF.nextValue(complex(c.Rgen.NormFloat64(),c.Rgen.NormFloat64()))
	}

	for i := 0; i < NRB; i++ {
		c.initz[i] = c.filterF.nextValue(complex(c.Rgen.NormFloat64(),c.Rgen.NormFloat64()))
	}

	for rb := 0; rb < NRB; rb++ {
		a := c.filterAr[rb].nextValue(c.initz[rb]) + complex(K,0)
		c.Iff_R[rb+tdma] = (real(a*cmath.Conj(a)) ) / (2 + K*K)
	}
	}

	}

	var touch bool

	prbase := rx.pr[E.GetId()]

	//RcBase := rx.GetRcBase(E.GetId())

	var capaTTI float64

	for rb:= 1; rb<NCh; rb++ {

		use:=ARB[rb]

		Rc := &rx.Channels[rb]

			NotPint1:=0.0
			if FadingOnPint1==Fading{ //if cancel, c.Iff_r =0 and total multiplied by -1, if fading, remove difference
			if(Rc.Signal[1]>=0){
				NotPint1=Rc.pr[Rc.Signal[1]]*(c.Iff_R[rb]-1)
			}
			}else if FadingOnPint1==Cancel{
				if(Rc.Signal[1]>=0){
					NotPint1=-Rc.pr[Rc.Signal[1]]
				}
			}


		if use {			

			Pr := Rc.pr[E.GetId()]

			c.Pr = Pr // to save data to file

		
			c.SNRrb[rb] = Pr * c.ff_R[rb] / (GetNoisePInterference(Rc.Pint, Pr)  + NotPint1 )


			BER := L1 * math.Exp(-c.SNRrb[rb]/2/L2) / 2.0

			c.meanPr.Add(Pr)
			c.meanSNR.Add(c.SNRrb[rb])
			c.meanBER.Add(BER)
			c.SNR += c.SNRrb[rb]

			//beta:= 1.5/-math.Log(5*c.meanBER.Get())

			capaTTI += EffectiveBW * math.Log2(  c.SNRrb[rb] + 1 )

			touch = true
		} else {

			//In case not using the RB, we eval the SNR on this rb using the base power level
			// that is the power received if not using beam forming
			// TODO, evaluate what would happen if estimating Pr with 10db gain for beam forming
			// thing is however, we can not evaluate interference with different beam shape
			// maybe we can suppose *estimate* interference of Pint*size_beam/360 with +10dB
			//_, Rc := rx.GetPr(E.GetId(), rb)

			/*if !dbs.IsInUse(rb) && SetReceiverType == BEAM { //!dbs.IsInUse(rb) if orientation < 0, not in use
				c.SNRrb[rb] = 9 * prbase * c.ff_R[rb] / (3.333*Rc.Pint + WNoise)
			} else { // else in use

				c.SNRrb[rb] = 9 * prbase * c.ff_R[rb] / (1.5*(Rc.Pint) + WNoise) // else it is in use
				// and we can suppose that the first signal is listened too and of course will not be emitting on this RB 						anymore if it is assigned to the current mobiles
			}*/

			//we substract

			Psig2:=0.0 //value of received power of the first signal it it is used by the dbs
			c2:=dbs.IsInUse(rb) //other signal
			if c2!=nil {Psig2= Rc.pr[c2.GetE().GetId()]}

			c.SNRrb[rb] = prbase * c.ff_R[rb] / ( GetNoisePInterference(Rc.Pint,Psig2) +NotPint1 ) // WNoise // Wnoise check

			//c.SNRrb[rb] = c.meanPr.Get() / ( GetNoisePInterference(Rc.Pint,Psig2) +NotPint1 ) // WNoise // Wnoise check

			c.SNRrb[rb] *= estimateFactor(dbs, E)
		

		}
	}
	

	if E.IsSetARB(0) {
		_, c.BER, c.SNR, c.Pr = rx.EvalSignalBER(E, 0)

		c.meanBER.Add(c.BER)
		c.meanSNR.Add(c.SNR)
		c.meanPr.Add(c.Pr)
		c.SNR=0
		//return
	}else if !touch { // add null to mean BER
		
		c.meanPr.Add(0)
		c.meanSNR.Add(0)
		c.SNR = 0
		c.meanBER.Add(1)


	}

	c.meanCapa.Add(capaTTI)

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

	//if SetReceiverType == BEAM && !dbs.IsInUse(rb) {
	o *= conservationFactor //}

	//c.SNRrb[rb] = 10 * prbase * c.ff_R[rb] / (3*Rc.Pint + WNoise)
	return

}

