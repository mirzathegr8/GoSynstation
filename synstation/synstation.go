package synstation

import "container/list"
import "container/vector"
import "rand"
import "math"
import "geom"
import "fmt"
//import "sort"

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
	R      PhysReceiverInt
	Connec *list.List
	Clock  int
	Rgen   *rand.Rand

	RndCh []int

	ConnectionBank vector.Vector
}


func (dbs *DBS) Init() {

	for i := 0; i < NConnec; i++ {
		dbs.ConnectionBank.Push(new(Connection))
	}

	switch SetReceiverType {
	case OMNI, BEAM:
		dbs.R = new(PhysReceiver)
	case SECTORED:
		dbs.R = new(PhysReceiverSectored)
	default:
		dbs.R = new(PhysReceiver)
	}
	dbs.Connec = list.New()
	dbs.RndCh = make([]int, NCh)
	var p geom.Pos
	p.X = Rgen.Float64() * Field
	p.Y = Rgen.Float64() * Field
	dbs.Rgen = rand.New(rand.NewSource(Rgen.Int63()))
	dbs.R.Init(p, dbs.Rgen)

	SyncChannel <- 1
}

// Physics : evaluate SNRs at receiver, evaluate BER of connections
func (dbs *DBS) RunPhys() {

	dbs.R.Compute(dbs.Connec)

	dbs.Clock = dbs.Rgen.Intn(EnodeBClock)

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.BitErrorRate(dbs.R.GetPhysReceiver(c.GetE().GetId()), dbs)

	}

	SyncChannel <- 1
}

func (dbs *DBS) FetchData() {
	SyncChannel <- 1
}

// Sorts channels in random order
func (dbs *DBS) RandomChan() {

	var SortCh vector.IntVector

	for i := 0; i < NCh; i++ {
		SortCh.Push(i)
		dbs.RndCh[i] = i
	}

	//randomizesd reserved top canals
	/*	for i := 10; i > 1; i-- {
			j := dbs.Rgen.Intn(i) + NCh-10
			SortCh.Swap(NCh-10+i-1, j)
			dbs.RndCh[NCh-10+i-1] =SortCh.Pop()
		}
		dbs.RndCh[NCh-10]=SortCh.Pop()
	*/
	//randomizes other canals
	/*	for i := NCh - 11; i > NChRes; i-- {
			j := dbs.Rgen.Intn(i-NChRes) + NChRes 
			SortCh.Swap(i, j)
			dbs.RndCh[i] =SortCh.Pop()
		}
		dbs.RndCh[NChRes] = SortCh.Pop()
		dbs.RndCh[0] = 0
	*/
	//fmt.Println(dbs.RndCh);

	for i := NCh - 1; i > NChRes; i-- {
		j := dbs.Rgen.Intn(i-NChRes) + NChRes
		SortCh.Swap(i, j)
		dbs.RndCh[i] = SortCh.Pop()
	}
	dbs.RndCh[NChRes] = SortCh.Pop()
	dbs.RndCh[0] = 0

}


func (dbs *DBS) disconnect(e *list.Element) {
	dbs.Connec.Remove(e)
	dbs.ConnectionBank.Push(e.Value.(*Connection))
	sens_disconnect++
}

func (dbs *DBS) IsInUse(i int) bool { // 

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.IsSetARB(i) {
			return true
		}
	}
	return false

}


func (dbs *DBS) connect(e EmitterInt, m float64) {
	//Connection instance are now created once and reused for memory consumption purpose
	// so the Garbage Collector needs not to lots of otherwise unessary work
	Conn := dbs.ConnectionBank.Pop().(*Connection)
	// these connection instance of course need to be initialized
	Conn.InitConnection(e, m, dbs.Rgen)
	dbs.Connec.PushBack(Conn)
	sens_connect++
}

func (dbs *DBS) IsConnected(tx EmitterInt) bool {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E == tx {
			return true
		}
	}
	return false

}


func (dbs *DBS) RunAgent() {

	dbs.checkLinkViability()

	if dbs.Clock == 0 {

		//if BWallocation == CHHOPPING {
		//	dbs.channelHopping()
		//} else {
		//dbs.ARBScheduler()
		//ARBScheduler2(dbs, dbs.Rgen)

		ARBSchedulFunc(dbs, dbs.Rgen)
		//}

		dbs.connectionAgent()

		if PowerControl == AGENTPC {
			dbs.optimizePowerAllocation()
		}
		//dbs.optimizePowerAllocationSimple()
	}

	SyncChannel <- 1.0

}

func (dbs *DBS) checkLinkViability() {
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.E.IsSetARB(0) {

			Pr, _ := dbs.R.GetPr(c.E.GetId(), 0)

			if 10*math.Log10(Pr/WNoise) < SNRThresConnec-2 {
				dbs.disconnect(e)
				sens_disconnect--
				sens_lostconnect++
			}
		}
		// remove any connection that does not satify the threshold
		if c.GetLogMeanBER() > math.Log10(BERThres) {
			dbs.disconnect(e)
			sens_disconnect--
			sens_lostconnect++
		}

	}

}

func (dbs *DBS) connectionAgent() {

	var conn int
	if dbs.Connec.Len() >= NConnec-1 {
		//disconnect
		var disc *list.Element
		var min float64
		min = 0.0

		for e := dbs.Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			if !c.E.IsSetARB(0) {
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

		for i, j := 0, NConnec-dbs.Connec.Len(); j >= 0 && i < SizeES; j-- {

			//var i=0
			Rc, Eval := dbs.R.EvalChRSignalSNR(0, i)

			if Rc.Signal[i] >= 0 {
				if !dbs.IsConnected(&Mobiles[Rc.Signal[i]]) {
					if 10*math.Log10(Eval) > SNRThresConnec {
						dbs.connect(&Mobiles[Rc.Signal[i]], 0.01)
						conn++
						return // we are done connecting
					}
				}

			}
			i++
		}

		// if no unconnected mobiles got connected, find one to provide it with macrodiversity

		for j := NConnec - dbs.Connec.Len(); j > 0; j-- {
			var max float64
			max = -10.0
			var Rc *ChanReceiver
			Rc = nil
			for i := NChRes; i < NCh; i++ {
				if dbs.IsInUse(i) == false {
					Rt, r, _ := dbs.R.EvalSignalConnection(i)
					if r > max {
						max = r
						Rc = Rt
						//	fmt.Println("attempt connect ",max,bb )
						//BERe = e
					}
				}
			}
			if Rc != nil {
				dbs.connect(&Mobiles[Rc.Signal[0]], 0.01)
				conn++
			} else {
				break
			}

		}

	}

}

//
//func (dbs *DBS) channelHopping2() {
//
//	//pour trier les connections actives
//	var MobileList vector.Vector
//
//	//pour trier les canaux
//	dbs.RandomChan()
//
//	// find a mobile
//	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
//		c := e.Value.(*Connection)
//
//		if c.Status == 0 { // only change if master
//
//			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel
//
//				var ratio float64
//				nch := 0
//
//				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
//				for j := NChRes; j < NCh; j++ {
//					i := dbs.RndCh[j]
//					if !dbs.IsInUse(i)  {
//
//						_, ber, snr, _ := dbs.R.EvalSignalBER(c.E, i)
//						ber = math.Log10(ber)
//
//						if ber < math.Log10(BERThres/10) {
//							if snr > ratio {
//								ratio = snr
//								nch = i
//								//assign and exit
//							}
//						}
//					}
//				}
//				if nch != 0 {
//					dbs.changeChannel(c, nch)
//					return
//				}
//
//				// sort mobile connection for channel hopping
//			} else {
//				ratio := c.EvalRatio(dbs.R)
//				var i int
//				for i = 0; i < MobileList.Len(); i++ {
//					co := MobileList.At(i).(ConnecType)
//					if ratio < co.EvalRatio(dbs.R) {
//						break
//					}
//				}
//				MobileList.Insert(i, c)
//			}
//		}
//	}
//
//	// change channel to some mobiles
//	for k := 0; k < MobileList.Len() && k < 15; k++ {
//		co := MobileList.At(k).(ConnecType)
//		//ratio := co.EvalRatio(&dbs.R)		
//
//		d := co.GetE().GetPos().Distance(dbs.R.GetPos())
//
//		//if (10*math.Log10(co.GetSNR())< SNRThres){
//		//var ir int
//		//ir:= NCh-NChRes + (6+int(math.Log10(Pr)))
//		//if d<100 {ir=28
//		//}else {ir=0}
//
//		//	ir:= NCh-NChRes + int(( -float(d)/1500*float(NCh-NChRes) ))
//		//if (ir<0) {ir=0}
//		//	if ir> NCh-2 {ir=NCh-2}
//
//		ir := 5
//		if d < 300 {
//			if !(co.GetE().GetARB()[0] > NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
//				for j := NCh - ir; j < NCh; j++ {
//
//					i := dbs.RndCh[j]
//					if !dbs.IsInUse(i) && i != co.GetE().GetARB()[0] {
//						_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)
//						if snr > SNRThresChHop {
//							dbs.changeChannel(co, i)
//							Hopcount++
//							break
//						}
//					}
//				}
//			}
//		} else {
//
//			if !(co.GetE().GetARB()[0] < NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
//				for j := NChRes; j < NCh-ir; j++ {
//
//					i := dbs.RndCh[j]
//
//					if !dbs.IsInUse(i) && i != co.GetE().GetARB()[0] {
//						_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)
//						if snr > SNRThresChHop {
//							dbs.changeChannel(co, i)
//							Hopcount++
//							break
//						}
//					}
//				}
//			}
//
//		}
//
//	}
//	//}
//
//	/*
//		for k := 0; k < MobileList.Len() && k < 1; k++ {
//			co := MobileList.At(k).(ConnecType)
//			ratio := co.EvalRatio(&dbs.R)
//
//
//			if (Pr<8e-9) && co.GetE().GetCh()>NCh-3{
//
//		//push down		
//			for j := NChRes; j < NCh; j++ {
//
//				i := dbs.RndCh[j]
//
//				if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
//					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
//					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
//						dbs.changeChannel(co, i)
//						Hopcount++
//						break
//					}
//				}
//			}} else{
//
//		//push up		
//			for j := NCh-2; j < NCh; j++ {
//
//				i := dbs.RndCh[j]			
//				if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
//					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
//					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
//						dbs.changeChannel(co, i)
//						Hopcount++
//						break
//					}
//				}
//			}}
//
//
//		}
//	*/
//
//
//}
//

func ChHopping(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	var MobileList vector.Vector

	//pour trier les canaux
	dbs.RandomChan()

	var stop = 0

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					i := dbs.RndCh[j]
					if !dbs.IsInUse(i) && !c.E.IsSetARB(i) {
						_, snr, _, _ := dbs.R.EvalSignalSNR(c.E, i)
						if 10*math.Log10(snr) > SNRThresChHop {
							if snr > ratio {
								ratio = snr
								nch = i
								//assign and exit
							}
						}
					}
				}
				if nch != 0 {
					dbs.changeChannel(c, 0, nch)
					stop++
					if stop > 5 {
						return
					}
				}

				// sort mobile connection for channel hopping
			} else {
				ratio := c.EvalRatio(dbs.R)
				var i int
				for i = 0; i < MobileList.Len(); i++ {
					co := MobileList.At(i).(ConnecType)
					if ratio < co.EvalRatio(dbs.R) {
						break
					}
				}
				MobileList.Insert(i, c)
			}
		}
	}

	// change channel to some mobiles
	for k := 0; k < MobileList.Len() && k < 2; k++ {
		co := MobileList.At(k).(ConnecType)
		ratio := co.EvalRatio(dbs.R)
		chHop := 0

		for j := NChRes; j < NCh; j++ {

			i := dbs.RndCh[j]

			if !dbs.IsInUse(i) && !co.GetE().IsSetARB(i) {

				_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)

				if snr > ratio {
					ratio = snr
					chHop = i
				}
			}
		}
		if chHop > 0 {
			dbs.changeChannel(co, co.GetE().GetFirstRB(), chHop)
			Hopcount++
		}

	}

}

func (dbs *DBS) changeChannel(co ConnecType, pch, nch int) {
	co.GetE().UnSetARB(pch)
	co.GetE().SetARB(nch)
}

func (dbs *DBS) optimizePowerAllocation() {

	var meanPtot, meanPd, meanPtotPd, meanPePd, meanPr, meanPe float64
	//var meanBER float64

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E
		meanPtotPd += M.BERT() / M.Req()
		meanPePd += c.BER / M.Req()
		meanPtot += M.BERT()
		meanPe += c.BER
		meanPd += M.Req()
		meanPr += c.Pr
	}

	nbconnec := float64(dbs.Connec.Len())

	meanPtot /= nbconnec
	meanPtotPd /= nbconnec
	meanPePd /= nbconnec
	meanPe /= nbconnec
	meanPd /= nbconnec
	meanPr /= nbconnec

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 && !c.E.IsSetARB(0) { // if master connection	

			var b, need, delta, alpha float64

			b = M.BERT() / M.Req() // meanPtotPd
			//b = M.BERT() / M.Req() / meanPtotPd
			b = b * 1.5
			alpha = 1.0
			need = 2.0*math.Exp(-b)*(b+1.0) - 1.0

			//need = .5 - math.Atan(5*(b-1))/math.Pi

			delta = math.Pow(geom.Abs(need), 1) *
				math.Pow(M.GetPower(), 1) *
				geom.Sign(need-M.GetPower()) * alpha *
				math.Pow(geom.Abs(need-M.GetPower()), 1.5)

			if math.IsNaN(delta) {
				fmt.Println("delta NAN", need, M.GetPower(), b, M.BERT())
				delta = -1
			}

			if delta > 0 {
				v := (1.0 - M.GetPower()) / 2.0
				if delta > v {
					delta = v
				}
			} else {
				v := -M.GetPower() / 2.0
				if delta < v {
					delta = v
				}
			}

			if delta > 1 || delta < -1 {
				fmt.Println("Power Error ", delta)
			}

			M.PowerDelta(delta)

			//M.SetPower(1.0/math.Pow(2.0+dbs.R.Pos.Distance(M.GetPos()),4 ) )
		}

	}

}


func (dbs *DBS) optimizePowerAllocationSimple() {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		M := c.E

		if c.Status == 0 {
			if !c.E.IsSetARB(0) { // if master connection and transmitting data

				L := float64(100)
				d := (L - dbs.R.GetPos().Distance(M.GetPos())) / L
				var p float64
				if d > 0 {
					//p=1-math.Pow(d,1)
					p = 1 - d
				} else {
					p = 1
				}

				//p:=0.001*(math.Pow(dbs.R.GetPos().Distance(M.GetPos()),4)/100000)

				M.SetPower(p)

			} else {
				M.SetPower(1)
			}

		}

	}

}


//Minimum Area-Difference to the Envelope


func ARBScheduler(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64

	//var MobilesID [NConnec]int

	var meanMeanCapa float64
	var maxMeanCapa float64

	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)

		m_m := c.GetE().GetMeanTR()
		meanMeanCapa += m_m
		if maxMeanCapa < m_m {
			maxMeanCapa = m_m
		}

	}
	meanMeanCapa /= float64(dbs.Connec.Len())
	meanMeanCapa += 0.1

	// Eval Metric for all connections
	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.GetE()

		if c.Status == 0 {

			m_m := c.GetE().GetMeanTR()

			for rb := 1; rb < NCh; rb++ {

				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.GetSNRrb(rb)
				}

				m := EffectiveBW * math.Log2(1+snrrb)

				//m_m := c.meanCapa.Get()

				//b := math.Exp(10 * (meanMeanCapa - E.GetMeanTR()) / maxMeanCapa)

				if m > 100 && m_m < 100000026000 {
					Metric[i][rb] = math.Log2(m + 1)
					b := (m_m + 1)
					if b > 1 {
						Metric[i][rb] /= b
					}

					//Metric[i][rb] *= b

					//Metric[i][rb] = m / c.GetE().GetMeanTR())
					//Metric[i][rb] = math.Log2(math.Log2(1 + c.ff_R[rb])) //* c.GetE().Req() / c.GetE().BERT()
				} else {
					Metric[i][rb] = 0
				}

			}

		}

	}

	//fmt.Println(Metric)

	//Assign RB for master connections

	var AL [NCh]int

	var NumAss [NConnec]int

	//First assign RB to best Metric
	for rb := 1; rb < NCh; rb++ {
		AL[rb] = -1
		for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
			c := e.Value.(*Connection)
			if c.Status == 0 {
				if Metric[i][rb] > 0.001 {
					if AL[rb] < 0 {
						AL[rb] = i
					} else if Metric[i][rb] > Metric[AL[rb]][rb] {
						AL[rb] = i
					}
				}
			}
		}
		if AL[rb] >= 0 {
			NumAss[AL[rb]]++ //this emitter will have one more assigned RB
			for rb2 := 1; rb2 < NCh; rb2++ {
				Metric[AL[rb]][rb2] *= float64(NumAss[AL[rb]]) / float64(NumAss[AL[rb]]+1)
			}
		}
	}
	//do not allocate RB for which capacity is too low // interference management
	/*for rb := 0; rb < NCh; rb++ {
		if Metric[AL[rb]][rb] < math.Log2(80) {
			AL[rb] = -1
		}

	}*/

	//Allocate RB effectivelly

	AL[0] = -1 // connect all
	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
		c := e.Value.(*Connection)
		E := c.GetE()
		if c.Status == 0 {
			if E.IsSetARB(0) {
				E.UnSetARB(0)
			}

		}
	}
	for rb := 1; rb < NCh; rb++ {
		if AL[rb] >= 0 {
			for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
				c := e.Value.(*Connection)
				E := c.GetE()

				if c.Status == 0 {

					if E.IsSetARB(rb) {
						if AL[rb] != i {
							E.UnSetARB(rb)
						}
					} else {
						if AL[rb] == i {
							E.SetARB(rb)
							Hopcount++

						}
					}
				}
			}
		}
	}

}


type vectorFloat64 struct {
	f []float64
	i []int
}

func (v vectorFloat64) Len() int           { return len(v.f) }
func (v vectorFloat64) Less(i, j int) bool { return v.f[i] < v.f[j] }
func (v vectorFloat64) Swap(i, j int) {
	v.f[i], v.f[j] = v.f[j], v.f[i]
	v.i[i], v.i[j] = v.i[j], v.i[i]
}


func max(v []float64) (a float64, i int) {
	if len(v) > 0 {
		a = v[0]
		i = 0
	}
	for j := 1; j < len(v); j++ {
		if v[j] > a {
			i = j
			a = v[j]
		}
	}
	return
}

func (v vectorFloat64) min() (a float64, i int) {
	if len(v.f) > 0 {
		a = v.f[0]
		i = 0
	}
	for j := 1; j < len(v.f); j++ {
		if v.f[j] < a {
			i = j
			a = v.f[j]
		}
	}
	return

}

func (v vectorFloat64) mean() float64 {
	var a float64
	for _, val := range v.f {
		a += val
	}
	return a
}

