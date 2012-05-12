package synstation

import "math"
import "geom"
import rand "math/rand"
import . "compMatrix"

import "math/cmplx"
//import "fmt"

// TODO what difference for SNR estimates on used/unsued rbs. should be equal for gensearch

var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }

const NP = 2  // numbers of simulated paths
const NA = 8 //numbers of antennas at receiver


//var PathGain = [5]float64{1, .5, 0.25, 0.05, 0.01} //0.5, 0.125} // relative powers of each path
var PathGain = [5]float64{1, 1, 1, 1, 1} //0.5, 0.125} // relative powers of each path

const NPdiv= NP*NA

func init() {

	sum := 0.0
	for np := 0; np < NP; np++ {
		sum += PathGain[np]
	}
	for np := 0; np < NP; np++ {
		PathGain[np] = math.Sqrt(PathGain[np] / sum)
	}	

}

type Connection struct {
	E   *Emitter
	IdB int //Id of the base station

	Status int //0 master ,1 slave

	//RBsReceiver

	meanPr   MeanData
	meanSNR  MeanData
	meanBER  MeanData
	meanCapa MeanData

	//Variables for generating fast fading
	filterF *Filter //Coherence Frequency filter	
	Rgen *rand.Rand
	initz [NPdiv][NCh]complex128 //generation of random number per RB	
	//filterAr FilterBank   //filter ban to use for channel gain FF generator
	filterAr [NPdiv][NCh]*Filter    //filter ban to use for channel gain FF generator
	ff_R     [NPdiv][NCh]complex128 // stores channel gain and phase for every RB every path

	//Variables for storing received powers 
	MultiPathMAgain   []float64 		// NAt*NCh vector length
	InterferencePowerExtra []float64
	InterferencePowerIntra []float64
	InterferersP [NConnec][]float64
	
	InterferersResidual []float64 //residual interference ifusing the channel

	InstEqSNR float64

	SNRrb []float64 //stores SNR per RB per NAt

	
	// Variables for MIMO Channel 
	//ComplexRand chan complex128
	
	WhRB *DenseMatrix

	sRr *DenseMatrix
	sRt *DenseMatrix

	HRB *DenseMatrix 

	WhHRB *DenseMatrix
//	CorrI *DenseMatrix

	HhH *DenseMatrix

	pathAoA   [NP]float64
	pathAoD   [NP]float64
	pathGains [NP]float64 //amplitutes ,  delay is already in filter fading

	// Data for calculus
	NoisePower []float64	

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

func (co *Connection) GenerateChannel(dbs *DBS) {

	NAt:=co.E.NAt
	NAr:=dbs.NAr

	//co.pathAoA[0] = 0.2* dbs.AoA[co.E.Id] + .8*co.pathAoA[0] + (.05*co.Rgen.Float64() - .025)
	//co.pathGains[0] = PathGain[0]

	/*for np := 1; np < NP; np++ {
		co.pathAoA[np] = co.pathAoA[np] + (.2*co.Rgen.Float64() - .1)
		if co.pathAoA[np] < 0 {
			co.pathAoA[np] += PI2
		}
		if co.pathAoA[np] > PI2 {
			co.pathAoA[np] -= PI2
		}
		//co.pathGains[np] = .8*co.pathGains[np] + .2*co.Rgen.Float64()*3.0334*PathGain[np] //the fact 3.033 is to compensate the power loss in the filter. the fixed pathgain is to set a relative gain for each path, such that their mean power summs to one

	}*/


	// Generate Fading values for all RBs and paths
	for np := 0; np < NPdiv; np++ {
		//first decorelate freq filter
		for i := 0; i < 50*corrF; i++ {
			co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}
		// generate NCh samples in frequencies
		for rb := range co.initz[np] {
			co.initz[np][rb] = co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
		}
		// output values for each path on each rb, multiplied by gain
		for rb := range co.initz[np] {
			co.initz[np][rb] = co.filterAr[np][rb].nextValue(co.initz[np][rb]) 
		}
	}


	//first path with line of sight
	K := dbs.kk[co.E.Id]
	NormFac:=math.Sqrt(2+K*K)
	for rb := range co.ff_R[0] {
		co.ff_R[0][rb] = (co.initz[0][rb] + complex(K, 0)) / complex(NormFac, 0) //(real(a * cmath.Conj(a))) / (2 + K*K)
	}

	for np := 1; np < NPdiv; np++ {
		for rb := range co.ff_R[np] {
			co.ff_R[np][rb] = co.initz[np][rb] / complex(1.4142, 0)

		}
	}

	//signals phase at each antenna (for each path)

	if Tti%5 ==0 {
	for np := 0; np < NP; np++ {
		cosAoA_2 := math.Cos(co.pathAoA[np]) / 2.0
		sin, cos := math.Sincos(cosAoA_2)
		phase := complex(cos, sin)
		Val := complex(1,0)
		co.sRr.Set(0,np, Val )
		for na := 1; na < dbs.NAr; na++ {
			Val = Val*phase
			 co.sRr.Set(na,np,Val) 
		}
	}

	for np := 0; np < NP; np++ {
		cosAoA_2 := math.Cos(co.pathAoD[np]) / 2.0
		sin, cos := math.Sincos(cosAoA_2)
		phase := complex(cos, sin)
		Val := complex(co.pathGains[np],0)
		co.sRt.Set(np,0, Val )
		for na := 1; na < co.E.NAt; na++ {
			Val = Val*phase
			 co.sRt.Set(np,na,Val) 
		}
	}

	}

	//Channel gains computation including power path loss and FF 

	PrEst :=  math.Sqrt(dbs.pr[co.E.Id]) //total power for SINR on unsued RB, to be divided by numARB in scheduler for metric estimation
	Pc:=0.0 //powercont
	numARB:=0.0
	for rb,use:=range co.E.ARB{
		if use{ Pc+=co.E.Power[rb]; numARB++}
	}
	Pc/=float64(numARB)
	PrEst*=Pc // to account for current power control TODO rethink about how to handle that

//	Pt := complex(1/math.Sqrt(float64(NAt)),0) // normalizing power for transmit power

	//fmt.Println(co.sRt)

	for rb := 0; rb < NCh; rb++ {
		
		var p float64		
		if co.E.ARB[rb] {
			p=math.Sqrt(dbs.Channels[rb].pr[co.E.Id]) //emitted power on rb + shadowing + path loss
		}else{
			p=PrEst
		}
		
		for nar:= 0 ; nar< NAr; nar++{
			for nat := 0; nat < NAt; nat++ {
				
				var Val complex128
				for np := 0; np < NP; np++ {			
					Val += co.sRr.Get(nar,np)*co.ff_R[np+nar*NP][rb]*co.sRt.Get(np,nat)	
				}
				Val*=complex(p*co.E.PowerNt[nat],0)
				co.HRB.Set(nar, nat+ NAt*rb, Val )
			}
		}
	}

	
	//SetGains for all rb including unsuded ones
/*	AoA := dbs.AoA[co.E.Id]

	cosAoA_2 := math.Cos(AoA) / 2
	sin, cos := math.Sincos(cosAoA_2)
	phase := complex(cos, -sin)

	//gain direction
	var defaultGain [NA]complex128
	defaultGain[0]=complex(1/dbs.SqrtNAr,0)
	for nar := 1; nar < NAr; nar++ {
		defaultGain[nar]=phase*defaultGain[nar-1]
		
			
	}

	for nat:=0 ;nat<NAt*NCh;nat++{
		co.WhRB.FillRow(nat,defaultGain[0:NAr])		
	}
*/

	for nat:=0 ;nat<NAt*NCh;nat++{
	for nar:=0 ;nar<NAr;nar++{ 
		co.WhRB.Set(nat,nar,cmplx.Conj(co.HRB.Get(nar,nat)))		
	}}


}



func (co *Connection) EvalInterference(dbs *DBS) {

	NAt:=co.E.NAt

	ConnectedArray := dbs.GetConnectedMobiles()

	for i := range co.InterferencePowerIntra {
		co.InterferencePowerIntra[i] = 0
		co.InterferencePowerExtra[i] = 0
	}
	

	for e,i := dbs.Connec.Front(),0; e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		nbRB := float64(c.E.GetNumARB())
		
		if c.E.Id == co.E.Id {

			co.WhRB.BlockTimes(c.HRB,co.WhHRB,NCh) //multiply block matrix

			co.WhHRB.BlockDiagMag(co.InterferersP[i]) //this to save for scheduler 
		
			co.WhHRB.SumNotDiagMag(co.InterferersResidual) // for mmse without iterative canceling
		
			copy(co.MultiPathMAgain, co.InterferersP[i])

		} else{

			co.WhRB.BlockTimesSumMag(c.HRB,co.InterferersP[i],NCh)

		}


		for rb, use := range c.E.ARB {	
			if use {

				for nat:=rb*NAt; nat<(rb+1)*NAt;nat++{

					if c.E.Id != co.E.Id {
						co.InterferencePowerIntra[nat] += co.InterferersP[i][nat]
					} 
		
					co.InterferersP[i][nat]*=nbRB // to normalize value without numARB included 					
					
				}
			}

		}

		i++
	}

	for m := range Mobiles {
		if !ConnectedArray[m] {
			for rb, use := range Mobiles[m].ARB {				
				if use {
					for nat:=0;nat<NAt;nat++{
						g := Mag(co.Gain(dbs.AoA[m],rb,nat))
						co.InterferencePowerExtra[rb*NAt+nat] += dbs.Channels[rb].pr[m] * g
					}
				}
			}
		}
			
	}

}

func (co *Connection) SetGains(dbs *DBS, gains []complex128, rb int, nat int) {

	co.WhRB.FillRow(rb*co.E.NAt+nat, gains)
}


func (co *Connection) Gain(AoA float64, rb int, nat int) complex128 {

	var Val complex128
	cosAoA_2 := math.Cos(AoA) / 2.0
	sin, cos := math.Sincos(cosAoA_2)
	phase := complex(cos, sin)
	delta := complex(1.0, 0.0)
	for _, wa := range co.WhRB.GetRow(rb*co.E.NAt+nat){
		Val += wa * delta
		delta *= phase
	}

	return Val
}



func (co *Connection) GetGain(AoA float64, rb int, nat int) float64 { //evals the gain of that mobile on this connection
	return Mag(co.Gain(AoA,rb,nat))
}



func (co *Connection) BitErrorRate(dbs *DBS) {

	NAt:=co.E.NAt
	var touch bool

	co.WhRB.SumRowMag(co.NoisePower);

//	fmt.Println(co.NoisePower)
//	fmt.Println(co.InterferersResidual)
	for nat, Pr :=  range co.MultiPathMAgain {
		co.SNRrb[nat] = Pr / (
co.InterferencePowerExtra[nat]+ 
co.InterferencePowerIntra[nat] +  
//co.InterferersResidual[nat] + 
WNoise*co.NoisePower[nat])
	}

	var nrbnt int

	for rb, use := range co.E.ARB{

		if use {
			for nat:=rb*NAt; nat<(rb+1)*NAt;nat++{
				if co.SNRrb[nat]>0.001 {

				BER:= L1 * math.Exp(- co.SNRrb[nat] /2/L2) / 2.0
				co.meanPr.Add(co.MultiPathMAgain[nat])
				co.meanSNR.Add(co.SNRrb[nat])
				co.meanBER.Add(BER)

				co.InstEqSNR += co.SNRrb[nat]
				nrbnt++
				touch = true
				}
			}

		} else {
			for nat:=rb*NAt; nat<(rb+1)*NAt;nat++{
				co.SNRrb[nat] *= conservationFactor
			}
		}
		
	}

	if !touch { // add null to mean BER
		co.meanPr.Add(0)
		co.meanSNR.Add(0)
		co.meanBER.Add(1)
	}else{
		co.InstEqSNR/=float64(nrbnt)
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

func (co *Connection) Reset(v float64, dbs *DBS){
	co.IdB = dbs.Id
	co.meanBER.Clear(v)
	co.Status = 1
	co.Rgen = dbs.Rgen

	co.pathAoA[0] = dbs.AoA[co.E.Id]
	co.pathGains[0] = PathGain[0]
	//divF := 5.0
	for np := 1; np < NP; np++ {
		co.pathAoA[np] = PI2 * co.Rgen.Float64()
		co.pathGains[np] = PathGain[np] //co.Rgen.ExpFloat64() / divF
		//	divF *= 5.0
	}

	//init sRt sRr in case they are not updated immediately
	//signals phase at each antenna (for each path)
	for np := 0; np < NP; np++ {
		cosAoA_2 := math.Cos(co.pathAoA[np]) / 2.0
		sin, cos := math.Sincos(cosAoA_2)
		phase := complex(cos, sin)
		Val := complex(1,0)
		co.sRr.Set(0,np, Val )
		for na := 1; na < dbs.NAr; na++ {
			Val = Val*phase
			 co.sRr.Set(na,np,Val) 
		}
	}

	for np := 0; np < NP; np++ {
		cosAoA_2 := math.Cos(co.pathAoD[np]) / 2.0
		sin, cos := math.Sincos(cosAoA_2)
		phase := complex(cos, sin)
		Val := complex(1,0)
		co.sRt.Set(np,0, Val )
		for na := 1; na < co.E.NAt; na++ {
			Val = Val*phase
			 co.sRt.Set(np,na,Val) 
		}
	}

	// Add some decorelation to the filters
	/*Speed := co.E.GetSpeed()
	DopplerF := Speed * F / cel //
	for np := 0; np < NPdiv; np++ {
		for l := 0; l < int(2.5/DopplerF); l++ {

			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64()))
			}
			
			for i := 0; i < NCh; i++ {
				co.filterAr[np][i].nextValue(co.filterF.nextValue(complex(co.Rgen.NormFloat64(), co.Rgen.NormFloat64())))
			}
		}
	}*/


}

func (co *Connection) InitConnection(E *Emitter, R *rand.Rand) {

	co.E = E

	Speed := E.GetSpeed()
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000 for RB TTI

	if DopplerF < 0.002 { // the frequency is so low, a simple antena diversity will compensate for 	
		DopplerF = 0.002
	}

	A := Butter(DopplerF)
	B := Cheby(10, DopplerF)
	C := MultFilter(A, B)

	co.filterF = CoherenceFilter.CopyNew()

	//generate MIMO

	NAt:=co.E.NAt
	NAr:=NArMAX

	
	//co.HhHRB = Zeros(NAt,NAt*NCh)
	co.HRB = Zeros(NAr,NAt*NCh)
	

	co.sRt= Zeros(NP,NAt)
	
	co.sRr= Zeros(NAr,NP)

	co.WhRB = Zeros(NAt*NCh,NAr)

	co.WhHRB = Zeros(NAt*NCh,NAt)

	for np:=0;np<NP;np++{
		co.pathAoD[np]=R.Float64()*PI2
	}


	co.MultiPathMAgain  = make([]float64, NAt*NCh) 		// NAt*NCh vector length
	co.InterferencePowerExtra = make([]float64, NAt*NCh)
	co.InterferencePowerIntra = make([]float64, NAt*NCh)
	for n:=0;n<NConnec;n++{
		co.InterferersP[n] = make([]float64, NAt*NCh)
	}
	co.InterferersResidual = make([]float64, NAt*NCh)
	co.SNRrb = make([]float64, NAt*NCh)

	co.NoisePower = make([]float64, NAt*NCh)


	
	
	for np := 0; np < NPdiv; np++ {
		for i := 0; i < NCh; i++ {
			co.filterAr[np][i] = C.CopyNew()
		}

		for l := 0; l < int(2.5/DopplerF); l++ {
			// for speed optimization, decorelation samples or not used, it makes little difference 
			for i := 0; i < 50; i++ {
				co.filterF.nextValue(complex(R.NormFloat64(), R.NormFloat64()))
			}
			for i := 0; i < NCh; i++ {
				co.filterAr[np][i].nextValue(co.filterF.nextValue(complex(R.NormFloat64(), R.NormFloat64())))
			}
		}
	}


}

func (co *Connection) clear() {
	// free some memory . perhaps need to rethink this and have a filterbank
	/*for np := 0; np < NPdiv; np++ {
		for rb := range co.filterAr {
			co.filterAr[np][rb] = nil
		}
	}*/

}

func NewConnection() (Conn *Connection) {
	Conn = new(Connection)
	return
}

func (co *Connection) GetLogMeanBER() float64 {
	return math.Log10(co.meanBER.Get() + 1e-10) //prevent saturation
}

func (co *Connection) PushBack()  {
	co.E.ConnectionBank <- co
}


