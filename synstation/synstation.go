package synstation

import "container/list"
import "rand"
import "math"
import "geom"
//import "fmt"
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

	ConnectionBank list.List

	Color int // This value is used to store some colorisation of the eNodeB, that is for example to use inside schedulers in honeycomb layout, where it will use a subset of RBs for ICIM

	ALsave [NCh]int

	Id int

	DBS2,DBS3 *DBS

	RBReuseFactor float64
	
}

var idtmp int

func (dbs *DBS) Init() {

	for i := 0; i < NConnec; i++ {
		dbs.ConnectionBank.PushBack(new(Connection))
	}

	
	dbs.Id=idtmp
	idtmp++

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

	dbs.RBReuseFactor = 0.5	

	SyncChannel <- 1
	
}

// Physics : evaluate SNRs at receiver, evaluate BER of connections
func (dbs *DBS) RunPhys() {

/*	var AL [NCh]int	
	for i:= range AL {AL[i]=-1}
	i:=0*/


	dbs.R.Compute(dbs.Connec,dbs.DBS2,dbs.DBS3)

	dbs.Clock = dbs.Rgen.Intn(EnodeBClock)



	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

	/*	if (c.Status==0){		
		for rb,v:= range c.GetE().GetARB(){
			if v {
//				if rb!=0 && AL[rb]!=-1 {fmt.Println("Double Assign")}
				AL[rb]=i

				}
		}
		i++
		}*/
		c.BitErrorRate(dbs.R.GetPhysReceiver(c.GetE().GetId()), dbs)
	}
		/*	AL[0]=-1
			if i>0 && testSCFDMA(AL[:])==-1{
				//for k:=range AL{  AL[k]=AL[k]-dbs.ALsave[k]}	
					fmt.Println(AL)
			}*/
	
	

	SyncChannel <- 1
}

func (dbs *DBS) FetchData() {
	SyncChannel <- 1
}


func (dbs *DBS) disconnect(e *list.Element) {
	dbs.Connec.Remove(e)
	dbs.ConnectionBank.PushBack(e.Value.(*Connection))
	sens_disconnect++
}

func (dbs *DBS) IsInUse(i int) *Connection { // 

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.IsSetARB(i) {
			return c
		}
	}
	return nil

}



func (dbs *DBS) IsInFuturUse(i int) bool { // 

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.IsFuturSetARB(i) {
			return true
		}
	}
	return false

}



func (dbs *DBS) connect(e *Emitter, m float64) {
	//Connection instance are now created once and reused for memory consumption purpose
	// so the Garbage Collector needs not to lots of otherwise unessary work
	Conn := dbs.ConnectionBank.Back().Value.(*Connection)
	dbs.ConnectionBank.Remove(dbs.ConnectionBank.Back())
	// these connection instance of course need to be initialized
	Conn.InitConnection(e, m, dbs.Rgen)
	dbs.Connec.PushBack(Conn)
	sens_connect++
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


func (dbs *DBS) RunAgent() {

	dbs.checkLinkViability()

	if dbs.Clock == 0 {		

		dbs.connectionAgent()
		ARBSchedulFunc(dbs, dbs.Rgen)
		
		
		
		//dbs.optimizePowerAllocationSimple()
	}
	PowerAllocation(dbs)

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
		} else if c.GetLogMeanBER() > math.Log10(BERThres) {	
		//	fmt.Println("disconnect")
			dbs.disconnect(e)
			sens_disconnect--
			sens_lostconnect++
		}

		/* for collocated eNodeB*/
		if  (dbs.DBS2!=nil && dbs.DBS2.IsConnected(c.GetE())) || (dbs.DBS3!=nil && dbs.DBS3.IsConnected(c.GetE())){
			dbs.disconnect(e)
			sens_disconnect--

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
				if !dbs.IsConnected(&Mobiles[Rc.Signal[i]].Emitter) {
					if 10*math.Log10(Eval) > SNRThresConnec {
						dbs.connect(&Mobiles[Rc.Signal[i]].Emitter, 0.001)
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
				if dbs.IsInUse(i) == nil {
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
				dbs.connect(&Mobiles[Rc.Signal[0]].Emitter, 0.001)
				conn++
			} else {
				break
			}

		}

	}

}


// deprecated, not used, will be removed in future
//Minimum Area-Difference to the Envelope

/*

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
*/
