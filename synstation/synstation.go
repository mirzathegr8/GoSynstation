package synstation

import "container/list"
import "container/vector"
import "rand"
import "math"
import "geom"
import "fmt"


// counters to observe connection agents health
var sens_connect, sens_disconnect, sens_lostconnect int
var Hopcount int

func GetConnect() int     { a := sens_connect; sens_connect = 0; return a }
func GetDisConnect() int  { a := sens_disconnect; sens_disconnect = 0; return a }
func GetLostConnect() int { a := sens_lostconnect; sens_lostconnect = 0; return a }
func GetHopCount() int    { a := Hopcount; Hopcount = 0; return a }

// a DBS is a reciever, a list of active connection
// it also is an agent and has a clock and internal random number generator
// RndCh stores channels sequence used when parsing channels for allocation
type DBS struct {
	R      PhysRecieverInt
	Connec *list.List
	Clock  int
	Rgen   *rand.Rand

	RndCh []int
}


func (dbs *DBS) Init() {
	switch SetReceiverType {
	case OMNI, BEAM:
		dbs.R = new(PhysReciever)
	case SECTORED:
		dbs.R = new(PhysRecieverSectored)
	}
	dbs.Connec = list.New()
	dbs.RndCh = make([]int, NCh)
	dbs.R.Init()
	var p geom.Pos
	p.X = Rgen.Float64() * Field
	p.Y = Rgen.Float64() * Field
	dbs.R.SetPos(p)
	dbs.Rgen = rand.New(rand.NewSource(Rgen.Int63()))
}

// Physics : evaluate SNRs at reciever, evaluate BER of connections
func (dbs *DBS) RunPhys() {

	dbs.R.DoTracking(dbs.Connec)

	dbs.R.MeasurePower(nil)

	dbs.Clock = dbs.Rgen.Intn(3)

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.BitErrorRate(dbs.R)

	}

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
	sens_disconnect++
}

func (dbs *DBS) IsInUse(i int) bool { // 

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.GetCh() == i {
			return true
		}
	}
	return false

}

func syncThread() { SyncChannel <- 1 }

func (dbs *DBS) RunAgent() {

	defer syncThread()
	// defer dbs.optimizePowerAllocationSimple()
	defer dbs.optimizePowerAllocation()

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.GetCh() == 0 {
			//check that the signal on the connecting channel is still the same 
			// and that its power is still enough
			Rc, Eval := dbs.R.EvalBestSignalSNR(0)
			if c.GetE() != Rc.Signal {
				dbs.disconnect(e)
				sens_disconnect--
				sens_lostconnect++
			} else {
				if 10*math.Log10(Eval) < SNRThresConnec-2 {
					dbs.disconnect(e)
					sens_disconnect--
					sens_lostconnect++
				}
			}

		}
		// remove any connection that does not satify the threshold
		if c.BER > math.Log10(BERThres) {
			dbs.disconnect(e)
			sens_disconnect--
			sens_lostconnect++
		}

	}

	if dbs.Clock == 0 {
		dbs.channelHopping()
	} else if dbs.Clock == 1 {

		if dbs.Connec.Len() >= NConnec {
			//disconnect
			var disc *list.Element
			var min float64
			min = 0.0

			for e := dbs.Connec.Front(); e != nil; e = e.Next() {
				c := e.Value.(*Connection)
				if c.E.GetCh() != 0 {
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

			{
				Rc, Eval := dbs.R.EvalBestSignalSNR(0)

				if Rc.Signal != nil {
					if !dbs.IsConnected(Rc.Signal) {
						if 10*math.Log10(Eval) > SNRThresConnec {
							dbs.connect(Rc.Signal)
							return // we are done connecting
						}
					}

				}
			}

			// if no unconnected mobiles got connected, find one to provide it with macrodiversity
			var max float64
			max = -10.0
			var Rc *ChanReciever
			for j := NConnec - dbs.Connec.Len(); j > 0; j-- {
				for i := NChRes; i < NCh; i++ {
					if !dbs.IsInUse(i) {
						Rt, r := dbs.R.EvalSignalConnection(i)
						if r > max {
							max = r
							Rc = Rt
						}
					}
				}
				if Rc != nil {
					dbs.connect(Rc.Signal)
				} else {
					break
				}
			}

		}

	}

}

func (dbs *DBS) connect(e EmitterInt) {
	Conn := new(Connection)
	Conn.E = e
	dbs.Connec.PushBack(Conn)
	sens_connect++
}


func (dbs *DBS) channelHopping2() {

	//pour trier les connections actives
	var MobileList vector.Vector

	//pour trier les canaux
	dbs.RandomChan()

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.GetCh() == 0 { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					i := dbs.RndCh[j]
					if !dbs.IsInUse(i) && i != c.E.GetCh() {

						_, ber, snr, _ := dbs.R.EvalSignalBER(c.E, i)

						if ber < math.Log10(BERThres/10) {
							if snr > ratio {
								ratio = snr
								nch = i
								//assign and exit
							}
						}
					}
				}
				if nch != 0 {
					dbs.changeChannel(c, nch)
					return
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
	for k := 0; k < MobileList.Len() && k < 15; k++ {
		co := MobileList.At(k).(ConnecType)
		//ratio := co.EvalRatio(&dbs.R)		

		d := co.GetE().GetPos().Distance(dbs.R.GetPos())

		//if (10*math.Log10(co.GetSNR())< SNRThres){
		//var ir int
		//ir:= NCh-NChRes + (6+int(math.Log10(Pr)))
		//if d<100 {ir=28
		//}else {ir=0}

		//	ir:= NCh-NChRes + int(( -float(d)/1500*float(NCh-NChRes) ))
		//if (ir<0) {ir=0}
		//	if ir> NCh-2 {ir=NCh-2}

		ir := 5
		if d < 300 {
			if !(co.GetE().GetCh() > NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
				for j := NCh - ir; j < NCh; j++ {

					i := dbs.RndCh[j]
					if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
						_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)
						if snr > SNRThresChHop {
							dbs.changeChannel(co, i)
							Hopcount++
							break
						}
					}
				}
			}
		} else {

			if !(co.GetE().GetCh() < NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
				for j := NChRes; j < NCh-ir; j++ {

					i := dbs.RndCh[j]

					if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
						_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)
						if snr > SNRThresChHop {
							dbs.changeChannel(co, i)
							Hopcount++
							break
						}
					}
				}
			}

		}

	}
	//}

	/*
		for k := 0; k < MobileList.Len() && k < 1; k++ {
			co := MobileList.At(k).(ConnecType)
			ratio := co.EvalRatio(&dbs.R)


			if (Pr<8e-9) && co.GetE().GetCh()>NCh-3{

		//push down		
			for j := NChRes; j < NCh; j++ {

				i := dbs.RndCh[j]

				if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
						dbs.changeChannel(co, i)
						Hopcount++
						break
					}
				}
			}} else{

		//push up		
			for j := NCh-2; j < NCh; j++ {

				i := dbs.RndCh[j]			
				if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
						dbs.changeChannel(co, i)
						Hopcount++
						break
					}
				}
			}}


		}
	*/


}


func (dbs *DBS) channelHopping() {

	//pour trier les connections actives
	var MobileList vector.Vector

	//pour trier les canaux
	dbs.RandomChan()

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.GetCh() == 0 { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					i := dbs.RndCh[j]
					if !dbs.IsInUse(i) && i != c.E.GetCh() {
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
					dbs.changeChannel(c, nch)
					return
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

			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {

				_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)

				if snr > ratio {
					ratio = snr
					chHop = i
				}
			}
		}
		if chHop > 0 {
			dbs.changeChannel(co, chHop)
			Hopcount++
		}

	}

}

func (dbs *DBS) changeChannel(co ConnecType, nch int) {
	co.GetE().SetCh(nch)
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

		if c.Status == 0 && c.E.GetCh() != 0 { // if master connection	

			var b, need, delta, alpha float64

			b = M.BERT() / M.Req()
			//b = M.BERT() / M.Req()/meanPtotPd;
			b = b * 1.5
			alpha = 1.0
			need = 2.0*math.Exp(-b)*(b+1.0) - 1.0
			delta = math.Pow(geom.Abs(need), 1) *
				math.Pow(M.GetPower(), 1) *
				geom.Sign(need-M.GetPower()) * alpha *
				math.Pow(geom.Abs(need-M.GetPower()), 1.5)

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
			if c.E.GetCh() != 0 { // if master connection	

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

func (dbs *DBS) IsConnected(tx EmitterInt) bool {

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E == tx {
			return true
		}
	}
	return false

}

