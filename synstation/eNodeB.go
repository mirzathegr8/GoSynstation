package synstation

import "container/list"
import "math"

import "compMatrix"
import "math/cmplx"
import "fmt"

func init() {

	fmt.Println("initEnodeB.go")
}

// counters to observe connection agents health
var sens_connect, sens_disconnect, sens_lostconnect int
var Hopcount int

func GetConnect() int     { a := sens_connect; sens_connect = 0; return a }
func GetDisConnect() int  { a := sens_disconnect; sens_disconnect = 0; return a }
func GetLostConnect() int { a := sens_lostconnect; sens_lostconnect = 0; return a }
func GetHopCount() int    { a := Hopcount; Hopcount = 0; return a }

// a DBS is a receiver, a list of active connection
// it also is an agent and has a clock and internal random number generator
// RndCh stores channels sequence used when parsing channels for allocation
type DBS struct {
	PhysReceiverBase
	Connec *list.List
	Clock  int

	RndCh []int

	//ConnectionBank list.List

	Color int // This value is used to store some colorisation of the eNodeB, that is for example to use inside schedulers in honeycomb layout, where it will use a subset of RBs for ICIM

	ALsave [NCh]int

	Id int

	RBReuseFactor float64

	scheduler Scheduler

	Masters [M]bool

	NMaxConnec int // numbers of connections per dbs
}

func (dbs *DBS) Init(i int) {

	dbs.Id = i
	dbs.NMaxConnec = NConnec

	/*for i := 0; i < NConnec; i++ {
		dbs.ConnectionBank.PushBack(NewConnection(dbs))
	}*/

	dbs.Connec = list.New()
	dbs.PhysReceiverBase.Init()
	//dbs.RBReuseFactor = 0.5	

	dbs.scheduler = initScheduler()

	SyncChannel <- 1

}

// Physics : evaluate SNRs at receiver, evaluate BER of connections
func (dbs *DBS) RunPhys() {

	dbs.Compute(dbs.Connec)

	dbs.Clock = dbs.Rgen.Intn(EnodeBClock)

	//must be done for all before evaluting gains and interfernce
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.GenerateChannel(dbs)
	}

	/*for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.SetGains(dbs)
	}*/

	dbs.SetReceiverGainsMMSE()

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		c.EvalInterference(dbs)
		c.BitErrorRate(dbs)
	}

	SyncChannel <- 1
}

func (dbs *DBS) FetchData() {
	SyncChannel <- 1
}

func (dbs *DBS) disconnect(e *list.Element) {	
	c := e.Value.(*Connection)
	dbs.Connec.Remove(e)
	c.clear()
	c.PushBack()
	//.ConnectionBank.PushBack(e.Value.(*Connection))
	sens_disconnect++
}

func (dbs *DBS) IsRBFree(rb int) bool { // 
	used := 0
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.ARB[rb] {
			used++
			if used >= mDf {
				return false
			}
		}
	}
	return true

}

func (dbs *DBS) IsInFuturUse(rb int) bool { // 
	used := 0
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.ARBfutur[rb] {
			used++
			if used >= mDf {
				return true
			}
		}
	}
	return false

}

func (dbs *DBS) connect(e *Emitter, m float64) (c *Connection) {
	//Connection instance are now created once and reused for memory consumption purpose
	// so the Garbage Collector needs not to lots of otherwise unessary work
	select {
		case c = <-e.ConnectionBank :
			c.Reset(m, dbs)
			dbs.Connec.PushBack(c)
			sens_connect++
		default:
			c= nil
		
	}
	return
}

func (dbs *DBS) IsConnected(tx *Emitter) bool {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E == tx {
			return true
		}
	}
	return false

}

func (dbs *DBS) GetConnectedMobiles() *[M]bool {
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		dbs.Masters[c.E.Id] = true
	}
	return &dbs.Masters
}

/*func (dbs *DBS) GetCancelation() *[M]bool {

	for m := range dbs.Masters {
		dbs.Masters[m] = false
	}

	switch InterferenceCancel {

	case MASTERCANCELATION:
		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			if c.Status == 0 {
				dbs.Masters[c.E.Id] = true
			}
		}
	case CONNECTEDCANCELATION:
		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			dbs.Masters[c.E.Id] = true
		}

	case NOCANCEL:

	}
	return &dbs.Masters
}*/

func (dbs *DBS) GetCancelationRB(rb int) *[M]bool {

	for m := range dbs.Masters {
		dbs.Masters[m] = false
	}

	for _, m := range dbs.Channels[rb].Signal {
		if m >= 0 {
			dbs.Masters[m] = true
		}
	}

	return &dbs.Masters
}

func (dbs *DBS) IsConnectedMaster(tx *Emitter) bool {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E == tx {
			if c.Status == 0 {
				return true
			} else {
				return false
			}
		}
	}
	return false

}

func (dbs *DBS) RunAgent() {

	dbs.checkLinkViability()

	if dbs.Clock == 0 {
		dbs.connectionAgent()
		dbs.scheduler.Schedule(dbs, dbs.Rgen)
	}
	PowerAllocation(dbs)

	SyncChannel <- 1.0

}

func (dbs *DBS) checkLinkViability() {
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.E.ARB[0] {

			Pr := dbs.GetPr(c.E.Id, 0)

			if 10*math.Log10(Pr/WNoise) < SNRThresConnec-2 {
				dbs.disconnect(e)
				sens_disconnect--
				sens_lostconnect++
			}
		} else if c.GetLogMeanBER() > math.Log10(BERThres) {
			dbs.disconnect(e)
			sens_disconnect--
			sens_lostconnect++
		}

	}

}

func (dbs *DBS) connectionAgent() {

	var conn int
	if dbs.Connec.Len() >= dbs.NMaxConnec-1 {
		//disconnect
		var disc *list.Element
		var min float64
		min = 0.0

		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			if !c.E.ARB[0] {
				r := c.EvalRatioDisconnect()
				if r < min {
					min = r
					disc = e
				}
			}
		}
		if disc != nil {
			dbs.disconnect(disc)
		}

	} else {

		//find and connect

		//First try to connect unconnected mobiles

		Rc := dbs.Channels[0]
		for i := 0; dbs.Connec.Len() < dbs.NMaxConnec && i < SizeES; i++ {

			if Rc.Signal[i] >= 0 {
				//fmt.Println("get a signal ", i, SizeES, dbs.Connec.Len(), dbs.NMaxConnec)
				if dbs.BelongsToNetwork(Rc.Signal[i]) {
					//fmt.Println("part of network")
					Eval := dbs.EvalConnection(i)
					if !dbs.IsConnected(&Mobiles[Rc.Signal[i]].Emitter) {
						//fmt.Println("is not connected")
						if 10*math.Log10(Eval) > SNRThresConnec {
							if dbs.connect(&Mobiles[Rc.Signal[i]].Emitter, 0.5)!=nil{
								conn++
							}
							//fmt.Println("connected")

							//return // we are done connecting
						}
					}
				}

			}

		}

		// if no unconnected mobiles got connected, find one to provide it with macrodiversity

		for j := dbs.NMaxConnec - dbs.Connec.Len(); j > 0; j-- {
		//if dbs.Connec.Len()<dbs.NMaxConnec{
			var max float64
			max = -10.0

			EmitterId := -1
			for rb := NChRes; rb < NCh; rb++ {
				if dbs.IsRBFree(rb) {
					r, EId := dbs.EvalSignalConnection(rb)
					if r > max {
						max = r
						EmitterId = EId
					}
				}
			}
			if EmitterId >= 0 {
				if  dbs.connect(&Mobiles[EmitterId].Emitter, 0.001)!=nil{
					conn++
				} 
			} else {
				break
			}

		}

	}

}

// return  quality indicator for unconnected mobiles
func (dbs *DBS) EvalConnection(k int) float64 {
	return dbs.pr[dbs.Channels[0].Signal[k]] / WNoise
}

// return quality indicator for mobiles connected to other dbs
func (dbs *DBS) EvalSignalConnection(rb int) (EvalMax float64, EmitterId int) {

	Rc := &dbs.Channels[rb]
	EvalMax = -100 //Eval is in [0 inf[, -100 means no signal
	EmitterId = -1
	for S := 0; S < SizeES; S++ {
		if Rc.Signal[S] >= 0 {
			if dbs.BelongsToNetwork(Rc.Signal[S]) && !dbs.IsConnected(&Mobiles[Rc.Signal[S]].Emitter) &&
					Mobiles[Rc.Signal[S]].Emitter.Diversity<MaxMacrodiv	{

				E := &Mobiles[Rc.Signal[S]].Emitter
				BER := dbs.EvalSignalBER(E, rb)
				BER = math.Log10(BER)
				Ptot := E.BERtotal + BER
				Eval := Ptot * math.Log(Ptot/BER)
				if EvalMax < Eval {
					EmitterId = E.Id
					EvalMax = Eval
				}
			}
		}
	}
	return
}

func (dbs *DBS) BelongsToNetwork(m int) bool {
	/*if dbs.Id >= D/2 {
		if m >= M/2 {
			return true
		} else {
			return false
		}
	} else if m < M/2 {
		return true
	}

	return false*/

	/*if dbs.Id%2==1{
			if m%2==1 {return true} else {return false}
	}else if (m+1)%2==1{return true}

	return false*/

	return true

}

func (dbs *DBS) EvalSignalSNR(e *Emitter, rb int) (SNR float64) {
	m := e.Id
	Pr := dbs.Channels[rb].pr[m]
	Pint := dbs.Channels[rb].Pint
	Psig := 0.0
	if e.ARB[rb] {
		Psig = Pr
	}
	if rb == 0 {
		Pint = 0
		Psig = 0
	}
	SNR = Pr / GetNoisePInterference(Pint, Psig)
	return
}

func (dbs *DBS) EvalSignalBER(e *Emitter, rb int) (BER float64) {

	K := dbs.kk[e.Id]
	SNR := dbs.EvalSignalSNR(e, rb)

	sigma := SNR / (K + 1.0)
	musqr := SNR - sigma
	eta := 1.0/sigma + 1.0/L2

	BER = math.Exp(-musqr/sigma) / (sigma * eta) * math.Exp(musqr/(sigma*sigma*eta))

	return
}

func (dbs *DBS) MU_factor_measure() (fact, nARB float64) {

	var reuse [NCh]int
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		//if c.Status == 0 {
		for m := 1; m < NCh; m++ {
			if c.E.ARBfutur[m] {
				reuse[m]++
				if reuse[m] == 1 {
					nARB++
				} //=float64(r)
			}
		}
		//}
	}

	for _, r := range reuse {

		if r > 1 {
			fact += float64(r - 1)
		}
	}

	return

}

func (dbs *DBS) SetReceiverGainsMMSE() {

	//sigma2 is the estimated variance of the noise + interferes far awway and not connected to the enode
	// hence sigma2 is the shadowing+ path loss * emitted power of all interferers  plus Wnoise
	// this is a worst case scenario

	var ConnecList [NConnec]*Connection

	R := compMatrix.Zeros(dbs.NAr, dbs.NAr)


	ConnectedArray := dbs.GetConnectedMobiles()


	Sigma2 := WNoise

	for rb := 0; rb < NCh; rb++ {

		Nc := 0
		Nm := 0
		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			if c.E.ARB[rb] {
				ConnecList[Nm] = c
				Nc+= c.E.NAt
				Nm++
			}
		}

		//out of reach mobiles interferer included in sigma noise

		if Nc > 0 {

			//number of other interferers
			Nc2 := 0
			for m := range Mobiles {
				if !ConnectedArray[m] {
					if Mobiles[m].ARB[rb] {
						Nc2++
					}
				}
			}

			H := compMatrix.Zeros(Nc+Nc2, dbs.NAr)
			Wh := compMatrix.Zeros(dbs.NAr, Nc)

			row:=0
			for m := 0; m < Nm; m++ {
				NAt:=ConnecList[m].E.NAt				
				for nat := 0; nat < NAt; nat++ {	
					//copies col from HHRB to H
					ConnecList[m].HRB.BufferCol(rb*NAt+nat, H.GetRow(row))
					row++
				}
			}

		
			//add vectors of interferers
			nc := Nc
		
			farInt:=0.0

			for m := range Mobiles {
				if !ConnectedArray[m] && Mobiles[m].ARB[rb] {

						Pi := dbs.Channels[rb].pr[m]

						farInt+=Pi

						AoA := dbs.AoA[m]
						//Compute the antenna geometrical phase shift for the signal

						cosAoA_2 := math.Cos(AoA) / 2.0
						sin, cos := math.Sincos(cosAoA_2)
						phase := complex(cos, sin)
						coef := complex(Pi, 0.0)
						H.Set(nc, 0, coef)
						for na := 1; na < dbs.NAr; na++ {
							coef *= phase
							H.Set(nc, na, coef)
						}

						nc++

					}
			}

	//fmt.Println(H)


		//	Sigma2+=farInt
		//Sigma2=farInt	+WNoise	
		//	fmt.Println(farInt)

			compMatrix.HilbertTimes(H, H, R)
			Eye := compMatrix.Eye(dbs.NAr)
			Eye.Scale(complex(Sigma2, 0))
			R.Add(Eye)		
			Ri, err := R.Inverse()

			if err == nil {

				compMatrix.TimesHilbert(Ri, H.GetMatrix(0, 0, Nc, dbs.NAr), Wh)
				W := Wh.Transpose() //Hilbert()
				Wrows := W.Arrays()				

				row:=0
				for m := 0; m < Nm ; m++ {
					for nat:=0;nat<ConnecList[m].E.NAt; nat++{
					/*	P:=0.0
						for _,v:=range Wrows[row]{
							P+=compMatrix.Mag(v)
						}
						P=math.Sqrt(P)
						for na,v:=range Wrows[row]{
							Wrows[row][na]= v/complex(P,0)
						}*/


						ConnecList[m].SetGains(dbs, Wrows[row], rb,nat)
						row++
					}
				}
			} else {
				fmt.Println(err)

			}

		}


	}
}

//func (dbs *DBS) SetReceiverGains() {

//	//sigma2 is the estimated variance of the noise + interferes far awway and not connected to the enode
//	// hence sigma2 is the shadowing+ path loss * emitted power of all interferers plus Wnoise
//	// this is a worst case scenario

//	var ConnecList [NConnec]*Connection

//	R := compMatrix.Zeros(NA, NA)


//	//ConnectedArray := dbs.GetConnectedMobiles()

//	//It := compMatrix.Zeros(NA, NA)
//	//Iv := compMatrix.Zeros(1, NA)
//	II := compMatrix.Zeros(NA, NA)

//	// var zerosV [NA*NA]float64

//	for rb := 0; rb < NCh; rb++ {

//		//out of reach mobiles interferer included in sigma noise
//		Sigma2 := WNoise
//	/*	for m := range Mobiles {
//			if !ConnectedArray[m] && Mobiles[m].ARB[rb] {
//				Pi := dbs.Channels[rb].pr[m]
//				AoA := dbs.AoA[m]
//				//Compute the antenna geometrical phase shift for the signal

//				cosAoA_2 := math.Cos(AoA) / 2.0
//				sin, cos := math.Sincos(cosAoA_2)
//				phase := complex(cos, sin)
//				coef := complex(Pi, 0.0)
//				Iv.Set(0, 0, coef)
//				for na := 1; na < NA; na++ {
//					coef *= phase
//					Iv.Set(0, na, coef)
//				}
//				for i := 0; i < NA; i++ {
//					for j := 0; j < NA; j++ {
//						It.Set(i, j, 0)
//					}
//				}
//				//It.elements

//				compMatrix.HilbertTimes(Iv, Iv, It)
//				II.Add(It)

////				Sigma2+=Pi

//			}
//		}
//*/
//		Sigma2+=1e-12

//		Nc := 0
//		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
//			c := e.Value.(*Connection)
//			if c.E.ARB[rb] {
//				ConnecList[Nc] = c
//				Nc++
//			}

//		}

//		if Nc > 0 {

//			H := compMatrix.Zeros(Nc, NA)
//			Wh := compMatrix.Zeros(NA, Nc)

//			for m := 0; m < Nc; m++ {
//				for na := 0; na < NA; na++ {
//					H.Set(m, na, ConnecList[m].H[rb][na])
//				}
//			}

//			compMatrix.HilbertTimes(H, H, R)

//			Eye := compMatrix.Eye(NA)
//			Eye.Scale(complex(Sigma2, 0))
//			R.Add(Eye)
//			R.Add(II)

//			Ri, err := R.Inverse()


//		/*	Ri,err := R.Inverse()
//			Eye.Scale(complex(Sigma2+1, 0))
//			BC,_ := Ri.TimesDense(II)
//			BCB,_ := BC.TimesDense(Ri)
//			BC.Add(Eye)
//			IBCinv,err3 :=  BC.Inverse()

//			MM,_:=IBCinv.TimesDense(BCB)			

//			Ri.Subtract( MM )*/
//			

//			if err == nil  {
//				compMatrix.TimesHilbert(Ri, H, Wh)
//				W := Wh.Transpose()
//				Wrows := W.Arrays()
//				for m := 0; m < Nc; m++ {
//					
//					P:=0.0
//					for _,v:=range Wrows[m]{
//						P+=Mag(v)
//					}
//					P=math.Sqrt(P)
//					for na,v:=range Wrows[m]{
//						Wrows[m][na]= v/complex(P,0)
//					}

//					ConnecList[m].SetGains(dbs, Wrows[m], rb)
//				}
//			} else {
//				fmt.Println(err)
//			}

//		}

//	}
//}


func (dbs *DBS) SetReceiverGainsMRC() {

	for rb := 0; rb < NCh; rb++ {

		//out of reach mobiles interferer included in sigma noise
		
		var Wh [NA]complex128

		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			NAt:=c.E.NAt
			if c.E.ARB[rb] {
				for nat:=0; nat< c.E.NAt;nat++{

					c.HRB.BufferCol(NAt*rb+nat, Wh[0:dbs.NAr])
					for nar:=0;nar<dbs.NAr;nar++{
						Wh[nar] = cmplx.Conj(Wh[nar])
					}
			
					c.SetGains(dbs, Wh[0:NAt], rb,nat)			
				}
			}

		}
}
}

