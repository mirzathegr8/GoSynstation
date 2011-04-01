package synstation


import "geom"
import "math"


// This struct stores flat data to be directly output for serialization, i.e. no pointers, no channels
type EmitterS struct {
	geom.Pos
	Power      float64 // current emitted power
	Ch         int     // current channel used
	BERtotal   float64
	Diversity  int
	Requested  float64
	MaxBER     float64
	SNRb       float64
	PrMaster   float64
	InstMaxBER float64

	Outage int
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

	MasterConnection *Connection

	touch bool

	Id int
}


// our little interface for emitters
type EmitterInt interface {
	AddConnection(c *Connection)
	BERT() float64
	Req() float64
	GetE() *Emitter
	GetPower() float64
	GetCh() int
	SetCh(c int)
	PowerDelta(float64)
	SetPower(float64)
	GetPos() geom.Pos
	//	isdone() chan int
	GetMasterConnec() *Connection
	GetId() int
	_setCh(i int)
}

func (e *Emitter) _setCh(i int) {
	e.Ch = i
	e.touch = false
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

func (e *Emitter) GetCh() int {
	return e.Ch
}

// channel used by channels change thread to inform emitter that channel hop has been applied
/*func (e *Emitter) isdone() chan int {
	return e.done
}*/

// function called by connections to inform BER quality of a link to the emitter
func (e *Emitter) AddConnection(c *Connection) {
	lber := c.GetLogMeanBER()
	if lber < math.Log10(BERThres) {
		e.SBERtotal += lber
		e.SDiversity++
		c.Status = 1 //we set the status as slave, as master status will be set after all connections data has been recieved
		num_con++
	}

	if e.SMaxBER > lber { //evaluate which connection is the best and memorizes which will be masterconnection
		e.MasterConnection = c
		e.SMaxBER = lber
		e.SInstMaxBER = math.Log10(c.BER)
		e.SNRb = c.SNR
		e.PrMaster = c.Pr
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
	if P < 0.01 {
		P = 0.01
	}
	M.Power = P
}

// only function that should be called to hop channel
// synchronizes lists
func (M *Emitter) SetCh(nch int) {

	//	M.nch = nch

	//if M.touch == false {

	SystemChan[nch].Change <- M
	//	M.touch = true
	//}
}

