package synstation

import "container/list"
import "math"

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

	ConnectionBank list.List

	Color int // This value is used to store some colorisation of the eNodeB, that is for example to use inside schedulers in honeycomb layout, where it will use a subset of RBs for ICIM

	ALsave [NCh]int

	Id int	

	RBReuseFactor float64
	
}

var idtmp int

func (dbs *DBS) Init() {

	for i := 0; i < NConnec; i++ {
		dbs.ConnectionBank.PushBack(new(Connection))
	}

	dbs.Id=idtmp
	idtmp++
	
	dbs.Connec = list.New()	
	dbs.PhysReceiverBase.Init()
	//dbs.RBReuseFactor = 0.5	

	SyncChannel <- 1
	
}


// Physics : evaluate SNRs at receiver, evaluate BER of connections
func (dbs *DBS) RunPhys() {

	dbs.Compute(dbs.Connec)

	dbs.Clock = dbs.Rgen.Intn(EnodeBClock)

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		c.BitErrorRate(dbs)
	}		

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


func (dbs *DBS) IsRBFree(rb int) bool { // 
	used:=0
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.IsSetARB(rb) {
			used++
			if used>=mDf{
				return false
			}
		}
	}
	return true

}

func (dbs *DBS) IsInFuturUse(i int) bool { // 
	used:=0
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)
		if c.E.IsFuturSetARB(i) {
			used++
			if used>= mDf{	
				return true
			}
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
	}
	PowerAllocation(dbs)

	SyncChannel <- 1.0

}

func (dbs *DBS) checkLinkViability() {
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.E.IsSetARB(0) {

			Pr := dbs.GetPr(c.E.GetId(),0)

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
			Eval := dbs.EvalConnection(i)
			Rc:=dbs.Channels[0]
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
			for rb := NChRes; rb < NCh; rb++ {
				if dbs.IsRBFree(rb)  {
					r := dbs.EvalSignalConnection(rb)
					if r > max {
						max = r
						Rc = &dbs.Channels[rb]				
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

// return  quality indicator for unconnected mobiles
func (dbs *DBS) EvalConnection(k int) float64{
 	return dbs.pr[dbs.Channels[0].Signal[k]]
}
// return quality indicator for mobiles connected to other dbs
func (dbs *DBS) EvalSignalConnection(rb int) (Eval float64) {

	Rc := &dbs.Channels[rb]
	Eval = -100 //Eval is in [0 inf[, -100 means no signal

	if Rc.Signal[0] >= 0 {
		E := &Mobiles[Rc.Signal[0]].Emitter
		BER := dbs.EvalSignalBER(E, rb)
		BER = math.Log10(BER)
		Ptot := E.BERT() + BER
		Eval = Ptot * math.Log(Ptot/BER)
	}

	return
}


func (dbs *DBS) EvalSignalSNR(e *Emitter, rb int) (SNR float64) {
	m := e.GetId()
	Pr:= dbs.Channels[rb].pr[m]
 	Pint := dbs.Channels[rb].Pint
	Psig:=0.0
	if e.IsSetARB(rb) {Psig= Pr}
	if rb==0 {Pint=0; Psig= 0}
	SNR = Pr / GetNoisePInterference(Pint,Psig) 
	return
}

func (dbs *DBS) EvalSignalBER(e *Emitter, rb int) (BER float64) {

	K := dbs.kk[e.GetId()]
	SNR:=dbs.EvalSignalSNR(e,rb)
 		
	sigma := SNR / (K + 1.0)
	musqr := SNR - sigma
	eta := 1.0/sigma + 1.0/L2

	BER = math.Exp(-musqr/sigma) / (sigma * eta) * math.Exp(musqr/(sigma*sigma*eta))

	return
}
