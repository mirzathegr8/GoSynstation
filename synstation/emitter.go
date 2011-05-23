package synstation


import "geom"
import "math"


// This struct stores flat data to be directly output for serialization, i.e. no pointers, no channels
type EmitterS struct {
	geom.Pos
	Power float64 // current emitted power

	BERtotal   float64
	Diversity  int
	Requested  float64
	MaxBER     float64
	SNRb       float64
	PrMaster   float64
	InstMaxBER float64

	Outage int

	ARB          [NCh]bool //allocated RB
	TransferRate float64
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

	SBERrb [NCh]float64
	SSNRrb [NCh]float64
	SNRrb  [NCh]float64

	MasterConnection *Connection

	touch bool

	Id int

	Speed [2]float64

	meanTR MeanData
}


// our little interface for emitters
type EmitterInt interface {
	AddConnection(c *Connection)
	BERT() float64
	Req() float64
	GetE() *Emitter
	GetPower() float64

	GetARB() []bool
	SetARB(i int)
	UnSetARB(i int)
	IsSetARB(i int) bool
	GetFirstRB() int
	ReSetARB()
	GetNumARB() int

	PowerDelta(float64)
	SetPower(float64)
	GetPos() geom.Pos
	//	isdone() chan int
	GetMasterConnec() *Connection
	GetId() int
	_setCh(i int)
	_unsetCh(i int)
	GetSpeed() float64

	GetSNRrb(rb int) float64

	GetMeanTR() float64
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

func (e *Emitter) _setCh(i int) {
	e.ARB[i] = true
}

func (e *Emitter) _unsetCh(i int) {
	e.ARB[i] = false
}


func (e *EmitterS) GetARB() []bool {
	return e.ARB[:]
}

func (e *EmitterS) IsSetARB(i int) bool {
	return e.ARB[i]
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
	if !e.ARB[i] {
		SystemChan[i].Change <- e
	}
}

func (e *Emitter) UnSetARB(i int) {
	if e.ARB[i] {
		SystemChan[i].Remove <- e
	}
}

func (e *Emitter) ReSetARB() {
	for i := 1; i < NCh; i++ {
		e.UnSetARB(i)
	}
	e.SetARB(0)

}

func (e *EmitterS) GetNumARB() (n int) {
	for i := 0; i < NCh; i++ {
		if e.ARB[i] {
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
func (e *Emitter) AddConnection(c *Connection) {

	lber := c.GetLogMeanBER()
	if lber <= math.Log10(BERThres) {
		e.SBERtotal += lber
		e.SDiversity++
		c.Status = 1 //we set the status as slave, as master status will be set after all connections data has been recieved
		num_con++

		if e.SMaxBER > lber { //evaluate which connection is the best and memorizes which will be masterconnection
			e.MasterConnection = c
			e.SMaxBER = lber
			e.SInstMaxBER = math.Log10(c.BER + 1e-40)
			e.SNRb = c.SNR
			e.PrMaster = c.Pr

			//for test with selection diversity

			if DiversityType == SELECTION {
				for rb := range e.ARB {
					//if use {
					e.SSNRrb[rb] = c.SNRrb[rb]
					//}
				}
			}

		}

		// for maximal RC
		if DiversityType == MRC {
			for rb := range e.ARB {
				//if use {
				e.SSNRrb[rb] += c.SNRrb[rb]
				//}
			}
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

	M.BERtotal, M.Diversity, M.MaxBER, M.InstMaxBER = M.SBERtotal, M.SDiversity, M.SMaxBER, M.SInstMaxBER
	M.SInstMaxBER, M.SBERtotal, M.SDiversity, M.SMaxBER = 0, 0, 0, 0

	M.TransferRate = 0

	M.Outage++
	for rb := 1; rb < NCh; rb++ {

		if M.IsSetARB(rb) {
			/*M.SBERrb[rb] = 0
			pe := L1 * math.Exp(-M.SSNRrb[rb]/2/L2) / 2.0
			for i := 0; i < 10; i++ {
				M.SBERrb[rb] += math.Pow(1-pe, 1024-float64(i)) *
					math.Pow(pe, float64(i)) * factorial[i]
			}
			M.SBERrb[rb] = 1 - M.SBERrb[rb]*/

			//M.meanBERInstTot.Add(M.SSNRrb[rb]) //for now as we use only 1 rb

			/*if M.Diversity > 0 {
				M.TransferRate = L1/2.0*math.Exp(-M.SSNRrb[rb]/2/L2) + 1e-40
			} else {
				M.TransferRate = 1

			}*/

			TransferRate := EffectiveBW * math.Log2(1+M.SSNRrb[rb])

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
	}

	M.meanTR.Add(M.TransferRate)

	if M.BERtotal == 0 && !M.IsSetARB(0) || M.MasterConnection == nil {
		M.MasterConnection = nil
		M.Power = 1
		M.ReSetARB() //That probably could run in conflicts. have to modify unset/setarb to only send one copy of emitter in channel if called multiple times
	} else if M.MasterConnection != nil {
		M.MasterConnection.Status = 0 // we are master
	}

	if M.IsSetARB(0) {

		if M.BERtotal == 0 {
			SyncChannel <- 0.0 //not even listened to
		} else {
			SyncChannel <- 1 //listen to but not connected
		}

	} else {
		SyncChannel <- M.BERtotal
	}

}

