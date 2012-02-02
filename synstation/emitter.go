package synstation

import "geom"
import "math"
import "fmt"
import "container/list"

func init() {
	fmt.Println(" fmt arb")
}

// This struct stores flat data to be directly output for serialization, i.e. no pointers, no channels
type EmitterS struct {
	geom.Pos
	Power [NCh]float64 // current emitted power

	BERtotal   float64
	Diversity  int
	Requested  float64
	MaxBER     float64
	SNRb       float64 //the sum of mean SNRs of all connections
	InstSNR    float64
	PrMaster   float64
	InstMaxBER float64

	Outage int

	ARB   [NCh]bool //allocated RB
	SNRrb [NCh]float64

	TransferRate float64

	Data int //quantity of data to send in bits

	IdB int // saves the id of the master BS

	MasterMultiPath [NCh]float64

}

// EmitterS with additional registers for BER and diversity evaluation, 
// link to master connection and channel to synchronize (radio) channel hopping
type Emitter struct {
	EmitterS

	//data used during calculation runtime
	SBERtotal   float64
	SMaxBER     float64
	SDiversity  int
	SInstMaxBER float64
	SMinDist    float64
	SSNRb       float64

	SBERrb [NCh]float64
	SSNRrb [NCh]float64

	MCSjoint float64

	MasterConnection *Connection

	Id int

	Speed [2]float64

	meanTR MeanData

	ARBfutur [NCh]bool          //allocated RB
	ARBe     [NCh]*list.Element //allocated RB
}


func (e *EmitterS) GetDataState() int {
	return e.Data
}

func (e *Emitter) ClearFuturARB() {
	for i := range e.ARBfutur {
		e.ARBfutur[i] = false
	}
}

func (e *Emitter) CopyFuturARB() {
	copy(e.ARBfutur[:], e.ARB[:])
}



func (e *Emitter) GetSpeed() float64 {
	return math.Sqrt(e.Speed[0]*e.Speed[0] + e.Speed[1]*e.Speed[1])
}



func (e *Emitter) GetFirstFutureRB() int {
	for i, use := range e.ARBfutur {
		if use {
			return i
		}
	}
	return -1
}

func (e *EmitterS) GetFirstRB() int {
	for i, use := range e.ARB {
		if use {
			return i
		}
	}
	return -1
}


func (e *Emitter) ReSetARB() {
	for i := 1; i < NCh; i++ {
		e.ARBfutur[i]=false
	}
	e.ARBfutur[0]=true

}

func (e *EmitterS) GetNumARB() (n int) {
	for _, v := range e.ARB {
		if v {
			n++
		}
	}
	return
}

func (e *EmitterS) GetMeanPower() (mp float64) {
	nrb := 0
	for rb, prb := range e.Power {
		if e.ARB[rb] {
			mp += prb
			nrb++
		}
	}
	if nrb>0 { mp /= float64(nrb)} 
	return
}

const betae = 1.0

// function called by connections to inform BER quality of a link to the emitter
func (e *Emitter) AddConnection(c *Connection, dbs *DBS) {

	lber := c.GetLogMeanBER()

	e.SBERtotal += lber
	e.SDiversity++
	c.Status = 1 //we set the status as slave, as master status will be set after all connections data has been recieved
	num_con++

	d := e.Pos.DistanceSquare(dbs.GetPos())

	//if d < e.SMinDist {
	if e.SMaxBER > lber { //evaluate which connection is the best and memorizes which will be masterconnection


		e.MasterConnection = c
		e.SMaxBER = lber
		e.SMinDist = d
		e.SInstMaxBER = math.Log10(c.meanBER.Get() + 1e-40)
		
		e.PrMaster = c.meanPr.Get()

		//for test with selection diversity
		if DiversityType == SELECTION {
			for rb := range e.ARB {
				e.SSNRrb[rb] = c.SNRrb[rb]
			}
		}

	}

	// for maximal RC
	if DiversityType == MRC {
		for rb ,use := range e.ARB {
			e.SSNRrb[rb] += c.SNRrb[rb]
			if use {
			e.SSNRb += math.Exp(-c.SNRrb[rb] / betae)
			}
		}
	}

}

func (e *Emitter) PowerDelta(rb int, delta float64) {
	e.SetPowerRB(rb, e.Power[rb]+delta)
}

func (e *Emitter) SetPowerRB(rb int, P float64) {
	if P > 1.0 {
		P = 1.0
	} else 	if P < 0.001 {
		P = 0.001
	} 
	
	e.Power[rb] = P
}

// Sets power on all RBs
func (e *Emitter) SetPower(P float64) {
	if P > 1.0 {
		P = 1.0
	}
	if P < 0.001 {
		P = 0.001
	}
	
	for i := range e.Power {
		e.Power[i] = P
	}
}

// this function saves Resets temporary variable after saving the Emitter's connection status
//	and selects the master connection 
// 	reset power and channel if all connections losts
// finnaly sents to syncchannel BER level
func (e *Emitter) FetchData() {

	var syncval float64

	e.SNRb, e.BERtotal, e.Diversity, e.MaxBER, e.InstMaxBER = e.SSNRb, e.SBERtotal, e.SDiversity, e.SMaxBER, e.SInstMaxBER
	e.SSNRb, e.SInstMaxBER, e.SBERtotal, e.SDiversity, e.SMaxBER, e.SMinDist = 0, 0, 0, 0, 0, Field*16*Field
	for rb := 0; rb < NCh; rb++ {
		e.SNRrb[rb], e.SSNRrb[rb] = e.SSNRrb[rb], 0
	}

	e.TransferRate = 0
	e.InstSNR = 0

	e.Outage++

	syncval = 1

	//beta:= 1.//1.5/ -(e.BERtotal*2.3026)

	effectSNR := 0.0
	minSNR :=100000000.0
	nARB := 0
	if e.Diversity == 0 {

		e.MasterConnection = nil
		e.IdB=-1
		e.SetPower(1)
		e.ReSetARB()
		syncval = 0

	} else {

		e.MasterConnection.Status = 0 // flag the best connection as master
		e.IdB=e.MasterConnection.IdB

		copy(e.MasterMultiPath[:],e.MasterConnection.MultiPathMAgain[:])

		for rb := 1; rb < NCh; rb++ {

			if e.ARB[rb] {

				switch TRATETECH {
					case OFDM :
						effectSNR += math.Exp(-e.SNRrb[rb] / betae)
					case SCFDM :
						effectSNR += e.SNRrb[rb]
					case NORMAL :
						s := EffectiveBW*math.Log2(1+beta*e.SNRrb[rb])
						s=math.Fmin(s,10000)
						if s> 100{
							effectSNR+=s
						} 							
				}

				if minSNR> e.SNRrb[rb] {minSNR=e.SNRrb[rb]}		

				nARB++

				if e.InstSNR < e.SNRrb[rb] {
					e.InstSNR = e.SNRrb[rb]
				}
			}
		}

		if nARB>0{
		switch TRATETECH {
			case OFDM2 :
				//this hack is to prevent overflow in the exponential /logarithm leading otherwise to +Inf transferrate
				// at high SINR the TR is anyways limited by the RB's lowest SINR
				e.SNRb = -betae*math.Log(e.SNRb/(float64(nARB)*float64(e.Diversity)) )
				if e.SNRb>600 {e.SNRb=minSNR}	
				e.TransferRate = EffectiveBW * float64(e.Diversity) * float64(nARB) * math.Log2(1 + e.SNRb)
				if e.TransferRate < float64(100*nARB) {e.TransferRate=0}			

			case OFDM :
				//this hack is to prevent overflow in the exponential /logarithm leading otherwise to +Inf transferrate
				// at high SINR the TR is anyways limited by the RB's lowest SINR
				e.SNRb = -betae*math.Log(effectSNR/float64(nARB))
				if e.SNRb>600 {e.SNRb=minSNR}			
				e.TransferRate = EffectiveBW * float64(nARB) * math.Log2(1 + e.SNRb)
				if e.TransferRate < float64(100*nARB) {e.TransferRate=0}	
			case SCFDM :
				effectSNR /= float64(nARB)
				e.SNRb=effectSNR
				e.TransferRate = EffectiveBW * float64(nARB) * math.Log2(1 + e.SNRb)			
				if e.TransferRate < float64(100*nARB) {e.TransferRate=0}
			case NORMAL :
				e.SNRb=math.Pow(2,effectSNR/EffectiveBW/float64(nARB))-1					
				e.TransferRate = effectSNR
		}

		if e.TransferRate != 0 {
			syncval = e.BERtotal
		}
		}	

	}

	if e.TransferRate > 100{
		e.Outage=0
	}else{
		e.TransferRate=0	
	}

	e.meanTR.Add(e.TransferRate)

	SyncChannel <- syncval

}

