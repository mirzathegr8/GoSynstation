package synstation


import "geom"
import "math"
import "fmt"
import "container/list"


func init(){
	fmt.Println(" fmt arb")
}

// This struct stores flat data to be directly output for serialization, i.e. no pointers, no channels
type EmitterS struct {
	geom.Pos
	Power float64 // current emitted power

	BERtotal   float64
	Diversity  int
	Requested  float64
	MaxBER     float64
	SNRb       float64
	InstSNR    float64
	PrMaster   float64
	InstMaxBER float64

	Outage int

	ARB          [NCh]bool //allocated RB
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
	SMinDist float64

	SBERrb [NCh]float64
	SSNRrb [NCh]float64
	SNRrb  [NCh]float64


	MasterConnection *Connection

	Id int

	Speed [2]float64

	meanTR MeanData

	ARBfutur     [NCh]bool //allocated RB
	ARBe         [NCh]*list.Element //allocated RB
}


// our little interface for emitters
type EmitterInt interface {
	AddConnection(c *Connection, dbs *DBS)
	BERT() float64
	Req() float64
	GetE() *Emitter
	GetPower() float64

	GetARB() []bool
	GetFuturARB() []bool
	SetARB(i int)
	UnSetARB(i int)
	IsSetARB(i int) bool
	IsFuturSetARB(i int) bool
	GetFirstRB() int
	ReSetARB()
	GetNumARB() int
	CopyFuturARB() // presets the future allocation to the current one
	ClearFuturARB()

	PowerDelta(float64)
	SetPower(float64)
	GetPos() geom.Pos

	GetMasterConnec() *Connection
	GetId() int

	GetSpeed() float64

	GetSNRrb(rb int) float64

	GetMeanTR() float64
}


func (e *EmitterS) GetDataState() int{
	return e.Data
}

func (e *Emitter) ClearFuturARB(){
	for i:=range e.ARBfutur {
		e.ARBfutur[i]=false
	}
}

func (e *Emitter) CopyFuturARB(){
	copy(e.ARBfutur[:],e.ARB[:])
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


func (e *EmitterS) GetFirstRB() int {
	for i, use := range e.ARB {
		if use {
			return i
		}
	}
	return -1
}

func (e *Emitter) SetARB(i int) {
	e.ARBfutur[i]=true
}

func (e *Emitter) UnSetARB(i int) {
	e.ARBfutur[i]=false
}

func (e *Emitter) ReSetARB() {
	for i := 1; i < NCh; i++ {
		e.UnSetARB(i)
	}
	e.SetARB(0)

}

func (e *EmitterS) GetNumARB() (n int) {
	for _,v := range e.ARB {
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

func (e *Emitter) GetPower() float64 {
	return e.Power
}


// function called by connections to inform BER quality of a link to the emitter
func (e *Emitter) AddConnection(c *Connection, dbs *DBS) {

	lber := c.GetLogMeanBER()
	
	e.SBERtotal += lber
	e.SDiversity++
	c.Status = 1 //we set the status as slave, as master status will be set after all connections data has been recieved
	num_con++

	d:=e.Pos.DistanceSquare(dbs.R.GetPos())
	//if d < e.SMinDist {
	if e.SMaxBER > lber { //evaluate which connection is the best and memorizes which will be masterconnection
	

		e.MasterConnection = c
		e.SMaxBER = lber
		e.SMinDist = d
		e.SInstMaxBER = math.Log10(c.BER + 1e-40)
		e.SNRb = c.SNR
		e.PrMaster = c.Pr

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
		}
	}

}


func (e *Emitter) BERT() float64 { return e.BERtotal }
func (e *Emitter) Req() float64  { return e.Requested }

func (M *Emitter) PowerDelta(delta float64) {
	M.SetPower(M.Power + delta)
}

func (M *Emitter) SetPower(P float64) {
	if P > 1.0 {
		P = 1.0
	}
	if P < 0.001 {
		P = 0.001
	}
	M.Power = P
}


// this function saves Resets temporary variable after saving the Emitter's connection status
//	and selects the master connection 
// 	reset power and channel if all connections losts
// finnaly sents to syncchannel BER level
func (M *Emitter) FetchData() {

	var syncval float64

	M.BERtotal, M.Diversity, M.MaxBER, M.InstMaxBER = M.SBERtotal, M.SDiversity, M.SMaxBER, M.SInstMaxBER
	M.SInstMaxBER, M.SBERtotal, M.SDiversity, M.SMaxBER, M.SMinDist = 0, 0, 0, 0, Field*16*Field

	M.TransferRate = 0
	M.InstSNR =0

	M.Outage++

	syncval = 1

	//beta:= 1.//1.5/ -(M.BERtotal*2.3026)

	if M.Diversity == 0 {

		M.MasterConnection = nil
		M.Power = 1
		M.ReSetARB()
		syncval = 0

	} else {

		for rb := 1; rb < NCh; rb++ {

			if M.IsSetARB(rb) {

				TransferRate := EffectiveBW * math.Log2(1 +  beta * M.SSNRrb[rb])

				if 100 < TransferRate {

					M.Outage = 0
					if TransferRate > 10000 {
						TransferRate = 10000
					}

				} else {

					TransferRate = 0

				}

				M.TransferRate += TransferRate

			}
			M.SNRrb[rb], M.SSNRrb[rb] = M.SSNRrb[rb], 0
		
			if (M.InstSNR< M.SNRrb[rb]) {M.InstSNR= M.SNRrb[rb]}
		}

		M.MasterConnection.Status = 0 // we are master

	//		M.InstSNR/=float64(M.GetNumARB());

		if M.TransferRate!=0 {
		syncval = M.BERtotal
		}

	}

	M.meanTR.Add(M.TransferRate)

	SyncChannel <- syncval

}

