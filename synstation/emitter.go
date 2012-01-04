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

// our little interface for emitters
type EmitterInt interface {
	AddConnection(c *Connection, dbs *DBS)
	BERT() float64
	Req() float64
	GetE() *Emitter

	GetARB() []bool
	GetFuturARB() []bool
	SetARB(i int)
	UnSetARB(i int)
	IsSetARB(i int) bool
	IsFuturSetARB(i int) bool
	GetFirstRB() int
	GetFirstFutureRB() int
	ReSetARB()
	GetNumARB() int
	CopyFuturARB() // presets the future allocation to the current one
	ClearFuturARB()

	GetPower(i int) float64
	GetMeanPower() float64
	PowerDelta(int, float64)
	SetPowerRB(int, float64)
	SetPower(float64)
	GetPos() geom.Pos

	GetMasterConnec() *Connection
	GetId() int

	GetSpeed() float64

	GetSNRrb(rb int) float64

	GetMeanTR() float64
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

func (e *Emitter) GetMeanTR() float64 {
	return e.meanTR.Get()
}

func (e *Emitter) GetSNRrb(rb int) float64 {
	return e.SNRrb[rb]
}

func (e *Emitter) GetSpeed() float64 {
	return math.Sqrt(e.Speed[0]*e.Speed[0] + e.Speed[1]*e.Speed[1])
}

func (e *EmitterS) GetARB() []bool {
	return e.ARB[:]
}

func (e *Emitter) GetFuturARB() []bool {
	return e.ARBfutur[:]
}

func (e *EmitterS) IsSetARB(i int) bool {
	return e.ARB[i]
}

func (e *Emitter) IsFuturSetARB(i int) bool {
	return e.ARBfutur[i]
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

func (e *Emitter) SetARB(i int) {
	e.ARBfutur[i] = true
}

func (e *Emitter) UnSetARB(i int) {
	e.ARBfutur[i] = false
}

func (e *Emitter) ReSetARB() {
	for i := 1; i < NCh; i++ {
		e.UnSetARB(i)
	}
	e.SetARB(0)

}

func (e *EmitterS) GetNumARB() (n int) {
	for _, v := range e.ARB {
		if v {
			n++
		}
	}
	return
}

func (e *Emitter) GetId() int {
	return e.Id
}

func (e *Emitter) GetMasterConnec() *Connection {
	return e.MasterConnection
}

func (e *Emitter) GetPos() geom.Pos {
	return e.Pos
}

func (e *Emitter) GetE() *Emitter {
	return e
}

func (e *EmitterS) GetPower(i int) float64 {
	return e.Power[i]
}

func (e *EmitterS) GetMeanPower() (mp float64) {
	nrb := 0
	for rb, prb := range e.Power {
		if e.ARB[rb] {
			mp += prb
			nrb++
		}
	}
	mp /= float64(nrb)
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
		for rb := range e.ARB {
			e.SSNRrb[rb] += c.SNRrb[rb]
			e.SSNRb+= math.Exp(-c.SNRrb[rb] / betae)
		}
	}

}

func (e *Emitter) BERT() float64 { return e.BERtotal }
func (e *Emitter) Req() float64  { return e.Requested }

func (M *Emitter) PowerDelta(rb int, delta float64) {
	M.SetPowerRB(rb, M.Power[rb]+delta)
}

func (M *Emitter) SetPowerRB(rb int, P float64) {
	if P > 1.0 {
		P = 1.0
	}
	if P < 0.001 {
		P = 0.001
	}
	M.Power[rb] = P
}

// Sets power on all RBs
func (M *Emitter) SetPower(P float64) {
	if P > 1.0 {
		P = 1.0
	}
	if P < 0.001 {
		P = 0.001
	}
	for i := range M.Power {
		M.Power[i] = P
	}
}

// this function saves Resets temporary variable after saving the Emitter's connection status
//	and selects the master connection 
// 	reset power and channel if all connections losts
// finnaly sents to syncchannel BER level
func (M *Emitter) FetchData() {

	var syncval float64

	M.SNRb, M.BERtotal, M.Diversity, M.MaxBER, M.InstMaxBER = M.SSNRb, M.SBERtotal, M.SDiversity, M.SMaxBER, M.SInstMaxBER
	M.SSNRb, M.SInstMaxBER, M.SBERtotal, M.SDiversity, M.SMaxBER, M.SMinDist = 0, 0, 0, 0, 0, Field*16*Field
	for rb := 0; rb < NCh; rb++ {
		M.SNRrb[rb], M.SSNRrb[rb] = M.SSNRrb[rb], 0
	}

	M.TransferRate = 0
	M.InstSNR = 0

	M.Outage++

	syncval = 1

	//beta:= 1.//1.5/ -(M.BERtotal*2.3026)

	effectSNR := 0.0
	minSNR :=100000000.0
	nARB := 0
	if M.Diversity == 0 {

		M.MasterConnection = nil
		M.SetPower(1)
		M.ReSetARB()
		syncval = 0

	} else {

		M.MasterConnection.Status = 0 // flag the best connection as master

		for rb := 1; rb < NCh; rb++ {

			if M.ARB[rb] {

				switch TRATETECH {
					case OFDM :
						effectSNR += math.Exp(-M.SNRrb[rb] / betae)
					case SCFDM :
						effectSNR += M.SNRrb[rb]
					case NORMAL :
						s := EffectiveBW*math.Log2(1+beta*M.SNRrb[rb])
						s=math.Fmin(s,10000)
						if s> 100{
							effectSNR+=s
						} 							
				}

				if minSNR> M.SNRrb[rb] {minSNR=M.SNRrb[rb]}		

				nARB++

				if M.InstSNR < M.SNRrb[rb] {
					M.InstSNR = M.SNRrb[rb]
				}
			}
		}

		if nARB>0{
		switch TRATETECH {
			case OFDM2 :
				//this hack is to prevent overflow in the exponential /logarithm leading otherwise to +Inf transferrate
				// at high SINR the TR is anyways limited by the RB's lowest SINR
				M.SNRb = -betae*math.Log(M.SNRb/(float64(nARB)*float64(M.Diversity)) )
				if M.SNRb>600 {M.SNRb=minSNR}			
				M.TransferRate = EffectiveBW * float64(M.Diversity) * float64(nARB) * math.Log2(1 + M.SNRb)
				if M.TransferRate < float64(100*nARB) {M.TransferRate=0}			

			case OFDM :
					//this hack is to prevent overflow in the exponential /logarithm leading otherwise to +Inf transferrate
					// at high SINR the TR is anyways limited by the RB's lowest SINR
				M.SNRb = -betae*math.Log(effectSNR/float64(nARB))
				if M.SNRb>600 {M.SNRb=minSNR}			
				M.TransferRate = EffectiveBW * float64(nARB) * math.Log2(1 + M.SNRb)
				if M.TransferRate < float64(100*nARB) {M.TransferRate=0}	
			case SCFDM :
				effectSNR /= float64(nARB)
				M.SNRb=effectSNR
				M.TransferRate = EffectiveBW * float64(nARB) * math.Log2(1 + M.SNRb)			
				if M.TransferRate < float64(100*nARB) {M.TransferRate=0}
			case NORMAL :
				M.SNRb=math.Pow(2,effectSNR/EffectiveBW/float64(nARB))-1					
				M.TransferRate = effectSNR
		}

		if M.TransferRate != 0 {
			syncval = M.BERtotal
		}
		}	

	}

	if M.TransferRate > 100{
		M.Outage=0
	}else{
		M.TransferRate=0	
	}

	M.meanTR.Add(M.TransferRate)

	SyncChannel <- syncval

}

