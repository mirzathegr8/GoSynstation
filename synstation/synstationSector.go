package synstation

import "container/list"
import "container/vector"
import "rand"
import "math"
import "geom"
//import "fmt"


// counters to observe connection agents health
var sens_connect, sens_disconnect int
func GetConnect() int    { a := sens_connect; sens_connect = 0; return a }
func GetDisConnect() int { a := sens_disconnect; sens_disconnect = 0; return a }

// a DBS is a reciever, a list of active connection
// it also is an agent and has a clock and internal random number generator
// RndCh stores channels sequence used when parsing channels for allocation
type DBS struct {
	R      PhysReciever	
	Connec *list.List
	clock  int
	Rgen   *rand.Rand

	RndCh []int
}


func (dbs *DBS) Init() {

	dbs.Connec = list.New()
	dbs.RndCh= make([]int,NCh)
	dbs.R.Init()
	Pos p;	
	p.X = Rgen.Float64() * Field
	p.Y = Rgen.Float64() * Field
	dbs.R.SetPos(p)
	dbs.Rgen = rand.New(rand.NewSource(Rgen.Int63()))
}

// Physics : evaluate SNRs at reciever, evaluate BER of connections
func (dbs *DBS) RunPhys() {

	//for i:=0;i<len(dbs.R.Orientation);i++ {dbs.R.Orientation[i]=-1}
	/*for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.GetCh()>0{		
			p:= c.GetE().GetPos().Minus(dbs.R.Pos)
			theta := math.Atan2(p.Y,p.X) * 180/ math.Pi
			if theta<0 {theta=theta+360}		
			dbs.R.Orientation[c.GetCh()]=theta + (dbs.Rgen.Float64()*30-15)

		}
		//fmt.Println(theta)
	}*/

	dbs.R.MeasurePower(nil)
	

	/*for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
if (c.GetCh()>=NChRes){
		fmt.Print ("  ",10*math.Log10( dbs.R.Channels[c.GetCh()].PrMax / (dbs.R.Channels[c.GetCh()].Pint - dbs.R.Channels[c.GetCh()].PrMax  +WNoise ) ))}
	}
		fmt.Println()
*/

	dbs.clock = dbs.Rgen.Intn(3)

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.BitErrorRate(&dbs.R)
		//if (c.GetCh()>NChRes){	fmt.Print("  ",c.BER,"  ",10*math.Log10( dbs.R.Channels[c.GetCh()].PrMax / (dbs.R.Channels[c.GetCh()].Pint - dbs.R.Channels[c.GetCh()].PrMax  +WNoise ) ))}
	}
//	fmt.Println()
	SyncChannel <- 1
}


// Sorts channels in random order
func (dbs *DBS) RandomChan() {

	var SortCh vector.IntVector

	for i := 0; i < NCh; i++ {
		SortCh.Push(i)
	dbs.RndCh[i]=i
	}
/*
//randomizesd reserved top canals
	for i := 10; i > 1; i-- {
		j := dbs.Rgen.Intn(i) + NCh-10
		SortCh.Swap(NCh-10+i-1, j)
		dbs.RndCh[NCh-10+i-1] =SortCh.Pop()
	}
	dbs.RndCh[NCh-10]=SortCh.Pop()

//randomizes other canals
	for i := NCh - 11; i > NChRes; i-- {
		j := dbs.Rgen.Intn(i-NChRes) + NChRes 
		SortCh.Swap(i, j)
		dbs.RndCh[i] =SortCh.Pop()
	}
	dbs.RndCh[NChRes] = SortCh.Pop()
	dbs.RndCh[0] = 0
*/
	//fmt.Println(dbs.RndCh);

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

func syncThread(){ SyncChannel<- 1}

func (dbs *DBS) RunAgent() {

	 defer syncThread()
	 defer dbs.optimizePowerAllocation()
		

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.BER > math.Log10(BERThres) {
			dbs.disconnect(e)
			sens_disconnect--
		}

	}

	if dbs.clock == 0 {
		dbs.channelHopping()
	} else if dbs.clock == 1 {

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
			if dbs.R.Channels[0].Signal != nil {
				Signal := dbs.R.Channels[0].Signal
				if 10*math.Log10(dbs.R.Channels[0].Eval_Signal()) > 35.0 {
					dbs.connect(Signal)
					return // we are done connecting
				}

			}

		// if no unconnected mobiles got connected, find one to provide it with macrodiversity
			var max float64
			max = -10.0 
			var Rc *ChanReciever

			for j:=NConnec-dbs.Connec.Len();j>0;j--{
			for i := range dbs.R.Channels {

				if !dbs.IsInUse(i) {			
					//r := dbs.R.Channels[i].Eval_Signal()
					r := dbs.R.Eval_Signal(i)
					if r > max {
						max = r
						Rc = &dbs.R.Channels[i]
					}
				}
			}
			if Rc != nil {
				dbs.connect(Rc.Signal)
			} else {break;}
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


var Hopcount int



func (dbs *DBS) channelHopping2(){

//pour trier les connections actives
	var MobileList vector.Vector

	//pour trier les canaux
	dbs.RandomChan()

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.GetCh() == 0 { //if the mobile is waiting to be assigned a proper channel

				Pr := c.Pr
				var ratio float64
				nch := 0
				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					i := dbs.RndCh[j]
					Rnew := &dbs.R.Channels[i]
					ev:=dbs.R.Eval_Signal(i)
					if !dbs.IsInUse(i) && i != c.E.GetCh()  && 
						ev<math.Log10(BERThres/10){
						b := Pr / (Rnew.Pint + WNoise)
						if b > ratio {
							ratio = b
							nch = i
							//assign and exit
						}
					}
				}
				if nch != 0 {
					dbs.changeChannel(c, nch)
					return
				}

			// sort mobile connection for channel hopping
			} else {
				ratio := c.EvalRatio(&dbs.R)
				var i int
				for i = 0; i < MobileList.Len(); i++ {
					co := MobileList.At(i).(ConnecType)
					if ratio < co.EvalRatio(&dbs.R) {
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
		Pr := co.GetPr()
		
		d:= co.GetE().GetPos().Distance(&dbs.R.Pos)
		
	//if (10*math.Log10(co.GetSNR())< SNRThres){
		//var ir int
		//ir:= NCh-NChRes + (6+int(math.Log10(Pr)))
		//if d<100 {ir=28
		//}else {ir=0}
				
	//	ir:= NCh-NChRes + int(( -float(d)/1500*float(NCh-NChRes) ))
		//if (ir<0) {ir=0}
	//	if ir> NCh-2 {ir=NCh-2}
	
		ir:=5
		if (d<300){
		if !( co.GetE().GetCh() > NCh-ir) || (co.GetSNR()<SNRThres-3){
		for j := NCh-ir; j < NCh; j++ {
			
			i := dbs.RndCh[j]
			Rnew := &dbs.R.Channels[i]
			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
				if Pr/(Rnew.Pint+WNoise) > SNRThres  {
					dbs.changeChannel(co, i)
					Hopcount++
					break
				}
			}
		}}}else {

		if !( co.GetE().GetCh() < NCh-ir) || (co.GetSNR()<SNRThres-3){
		for j := NChRes; j < NCh-ir; j++ {
			
			i := dbs.RndCh[j]
			Rnew := &dbs.R.Channels[i]
			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
				if Pr/(Rnew.Pint+WNoise) > SNRThres  {
					dbs.changeChannel(co, i)
					Hopcount++
					break
				}
			}}}


		}

	}
		//}


	/*
	for k := 0; k < MobileList.Len() && k < 1; k++ {
		co := MobileList.At(k).(ConnecType)
		ratio := co.EvalRatio(&dbs.R)
		Pr := co.GetPr()
				
		if (Pr<8e-9) && co.GetE().GetCh()>NCh-3{

	//push down

		for j := NChRes; j < NCh; j++ {
			
			i := dbs.RndCh[j]
			Rnew := &dbs.R.Channels[i]
			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
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
			Rnew := &dbs.R.Channels[i]
			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
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

				Pr := c.Pr
				var ratio float64
				nch := 0
				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					i := dbs.RndCh[j]
					Rnew := &dbs.R.Channels[i]
					ev:=dbs.R.Eval_Signal(i)
					if !dbs.IsInUse(i) && i != c.E.GetCh()  && 
						ev<math.Log10(BERThres/10){
						b := Pr / (Rnew.Pint + WNoise)
						if b > ratio {
							ratio = b
							nch = i
							//assign and exit
						}
					}
				}
				if nch != 0 {
					dbs.changeChannel(c, nch)
					return
				}

				// sort mobile connection for channel hopping
			} else {
				ratio := c.EvalRatio(&dbs.R)
				var i int
				for i = 0; i < MobileList.Len(); i++ {
					co := MobileList.At(i).(ConnecType)
					if ratio < co.EvalRatio(&dbs.R) {
						break
					}
				}
				MobileList.Insert(i, c)
			}
		}
	}

	// change channel to some mobiles
	for k := 0; k < MobileList.Len() && k < 1; k++ {
		co := MobileList.At(k).(ConnecType)
		ratio := co.EvalRatio(&dbs.R)
		Pr := co.GetPr()
				

		for j := NChRes; j < NCh; j++ {
			
			i := dbs.RndCh[j]
			Rnew := &dbs.R.Channels[i]
			if !dbs.IsInUse(i) && i != co.GetE().GetCh() {
				if Pr/(Rnew.Pint+WNoise) > ratio {
					dbs.changeChannel(co, i)
					Hopcount++
					break
				}
			}
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
			b = b * .8
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

			M.PowerDelta(delta)

			//M.SetPower(1.0/math.Pow(2.0+dbs.R.Pos.Distance(M.GetPos()),4 ) )
		}

	}

}

