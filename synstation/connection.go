package synstation

import "math"
import "geom"
import rand "rand"
import "cmath"
//import cmath "math/cmplx"
//import "fmt"

// TODO make sum multi paht power mean to equal 1
// TODO what difference for SNR estimates on used/unsued rbs. should be equal for gensearch

var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }

const NP = 3 // numbers of simulated paths
const NA = 10 //numbers of antennas at receiver

type Connection struct {
	E   *Emitter
	IdB int //Id of the base station

	Status int //0 master ,1 slave

	RBsReceiver

	meanPr   MeanData
	meanSNR  MeanData
	meanBER  MeanData
	meanCapa MeanData

	//filterAr FilterBank   //filter ban to use for channel gain FF generator
	filterAr [NP][NCh]*Filter    //filter ban to use for channel gain FF generator
	ff_R     [NP][NCh]complex128 // stores channel gain and phase for every RB every path

	MultiPathMAgain   [NCh]float64
	InterferencePower [NCh]float64
	//CorrelationMatrix [NA][NA]float64

	SNRrb [NCh]float64 //stores SNR per RB


	filterF *Filter //Coherence Frequency filter	

	Rgen *rand.Rand

	initz [NP][NCh]complex128 //generation of random number per RB	

	//ComplexRand chan complex128

	antennaGains [NA]complex128
	pathAoA      [NP]float64
	pathGains    [NP]float64 //amplitutes ,  delay is already in filter fading

	//gainM [M]float64
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

var PathGain = [3]float64 {0.79057, 0.5, 0.35355} // to normalize sum power of paths

func (co *Connection) EvalVectPath(dbs *DBS) {

	co.pathAoA[0] = 0.2* dbs.AoA[co.E.Id] + .8*co.pathAoA[0] + (.05*co.Rgen.Float64() - .025)
	co.pathGains[0] = PathGain[0]

	for np := 1; np < NP; np++ {
		co.pathAoA[np] = co.pathAoA[np] + (.2*co.Rgen.Float64() - .1)
		if co.pathAoA[np] < 0 {
			co.pathAoA[np] += PI2
		}
		if co.pathAoA[np] > PI2 {
			co.pathAoA[np] -= PI2
		}
		co.pathGains[np] = .8*co.pathGains[np] + .2*co.Rgen.Float64()*3.0334*PathGain[np] //the fact 3.033 is to compensate the power loss in the filter. the fixed pathgain is to set a relative gain for each path, such that their mean power summs to one

	}

	for np := 0; np < NP; np++ {
		//first decorelate freq filter
		for i := 0; i < 50*corrF; i++ {
			co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		// generate NCh samples in frequencies
		for rb := range co.initz[np] {
			co.initz[np][rb] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}

		// output values for each path on each rb
		for rb := range co.initz[np] {
			co.initz[np][rb] = co.filterAr[np][rb].nextValue(co.initz[np][rb])
		}
	}

	//first path with line of sight
	K := dbs.kk[co.E.Id]
	for rb := range co.ff_R[0] {
		co.ff_R[0][rb] = (co.initz[0][rb] + complex(K, 0)) / complex(2+K*K, 0) //(real(a * cmath.Conj(a))) / (2 + K*K)
	}

	for np := 1; np < NP; np++ {
		for rb := range co.ff_R[np] {
			co.ff_R[np][rb] = co.initz[np][rb] / complex(2, 0)

		}
	}

	//antenna total gains
	for rb := range co.ff_R[0] {
		var val complex128
		for np := 0; np < NP; np++ {
			val += co.Gain(complex(co.pathGains[np], 0)*co.ff_R[np][rb], co.pathAoA[np], dbs)
		}
		co.MultiPathMAgain[rb] = real(val)*real(val) + imag(val)*imag(val)
	}

}

func (co *Connection) EvalInterference(dbs *DBS) {

	//add approximation for non connected mobiles (mean interference) without fading

	ConnectedArray := dbs.GetConnectedMobiles()

	for rb := range co.Channels {

		co.InterferencePower[rb] = 0

		for m := range Mobiles { //Mobiles {
			if Mobiles[m].ARB[rb] && !ConnectedArray[m] {
				// Evaluate Beam Gain	
				gain := co.Gain(complex(1, 0), dbs.AoA[m], dbs)
				co.InterferencePower[rb] += dbs.Channels[rb].pr[m] * (real(gain)*real(gain) + imag(gain)*imag(gain))

			}
		}
	}
	// sum miltupaths for all connections with appropriate gainM

	sumPower := &co.initz[0]
	for rb := 0 ; rb<NCh ;rb++ { sumPower[rb]=0}
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.Id != co.E.Id {
			//others paths for these gain combination
			var IP complex128
			for rb, use := range c.E.ARB {
				if use {
					for np := 0; np < NP; np++ {
						IP += co.Gain(c.ff_R[np][rb]*complex(c.pathGains[np], 0), c.pathAoA[np], dbs)
					}
					sumPower[rb]+= IP* complex(dbs.Channels[rb].pr[c.E.Id],0)
				}				
			}
		}
	}
	for rb :=0 ; rb<NCh;rb++ { 
		co.InterferencePower[rb] += (real(sumPower[rb])*real(sumPower[rb]) + 
						imag(sumPower[rb])*imag(sumPower[rb])) 
	}

	//fmt.Println(" ", co.InterferencePower)
}

func (co *Connection) SetGains(dbs *DBS) {

	AoA := dbs.AoA[co.E.Id]

	delta := math.Cos(AoA) / 2

	for na := 0; na < NA; na++ {
		co.antennaGains[na] = cmath.Exp(complex(0, -float64(na)*delta))
	}

}

func (co *Connection) Gain(Signal complex128, AoA float64, dbs *DBS) complex128 {

	var Val complex128
	delta := math.Cos(AoA) / 2
	for na := 0; na < NA; na++ {

		Val +=  Signal *co.antennaGains[na] * cmath.Exp(complex(0, float64(na)*delta))

	}

	return Val
}


func (co *Connection) GetGain( AoA float64) float64  { //evals the gain of that mobile on this connection

	var Val complex128
	delta := math.Cos(AoA) / 2
	for na := 0; na < NA; na++ {

		Val +=  co.antennaGains[na] * cmath.Exp(complex(0, float64(na)*delta))

	}

	return real(Val)*real(Val)+imag(Val)*imag(Val)

}

/*
func (co *Connection) Gain(AoA float64, dbs *DBS) complex128{




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
}*/

func (co *Connection) BitErrorRate(dbs *DBS) {

	var touch bool

	for rb, use := range co.E.ARB {

		if use {

			Pr := co.MultiPathMAgain[rb] * dbs.Channels[rb].pr[co.E.Id]
			co.SNRrb[rb] = Pr / (co.InterferencePower[rb] + WNoise)
			//GetNoisePInterference(co.InterferencePower[rb], co.MultiPathMAgain[rb])

			BER := L1 * math.Exp(-co.SNRrb[rb]/2/L2) / 2.0

			co.meanPr.Add(Pr)
			co.meanSNR.Add(co.SNRrb[rb])
			co.meanBER.Add(BER)

			touch = true
		} else {
			co.SNRrb[rb] = co.MultiPathMAgain[rb] * dbs.pr[co.E.Id] / (co.InterferencePower[rb] + WNoise) * conservationFactor
			//GetNoisePInterference(co.InterferencePower[rb], 0) 


		}
	}

	if !touch { // add null to mean BER
		co.meanPr.Add(0)
		co.meanSNR.Add(0)
		co.meanBER.Add(1)
	}

	//fmt.Println(co.SNRrb)

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

func (co *Connection) InitConnection(E *Emitter, v float64, dbs *DBS) {

	co.E = E
	co.meanBER.Clear(v)
	co.Status = 1

	co.Rgen = dbs.Rgen

	Speed := E.GetSpeed()
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000 for RB TTI

	if DopplerF < 0.002 { // the frequency is so low, a simple antena diversity will compensate for 	
		DopplerF = 0.002
	}

	A := Butter(DopplerF)
	B := Cheby(10, DopplerF)
	C := MultFilter(A, B)

	co.filterF = CoherenceFilter.Copy()

	for np := 0; np < NP; np++ {
		for i := 0; i < NCh; i++ {
			co.filterAr[np][i] = C.Copy()
		}

		for l := 0; l < int(2.5/DopplerF); l++ {

			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
			}

			for i := 0; i < NCh; i++ {
				co.filterAr[np][i].nextValue(co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())))
			}

		}
	}

	co.pathAoA[0] = dbs.AoA[co.E.Id]
	co.pathGains[0] = 1
	divF := 5.0
	for np := 1; np < NP; np++ {
		co.pathAoA[np] = PI2 * co.Rgen.Float64()
		co.pathGains[np] = co.Rgen.ExpFloat64() / divF
		divF *= 5.0

	}

}

func (co *Connection) clear() {
	// free some memory . perhaps need to rethink this and have a filterbank
	for np := 0; np < NP; np++ {
		for rb := range co.filterAr {

			co.filterAr[np][rb] = nil
			//co.IfilterAr[rb]=nil
		}
	}

}

func NewConnection(i int) (Conn *Connection) {
	Conn = new(Connection)
	Conn.IdB = i
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

//   Reformatted by   lerouxp    Tue Jan 31 23:41:07 CET 2012

