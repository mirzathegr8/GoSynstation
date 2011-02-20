package synstation

import "math"
import "geom"
//import "fmt"


var num_con int

func GetDiversity() int { a := num_con; num_con = 0; return a }

type Connection struct {
	E EmitterInt

	Pr     float64 // to store power level received	
	SNR, K float64
	BER    float64

	Status int //0 master ,1 slave

	meanPr  MeanData
	meanSNR MeanData
	meanBER MeanData
}


type ConnectionS struct {
	A, B   geom.Pos
	Status int
	BER    float64
	Ch     int
}

func (c *ConnectionS) Copy(cc *Connection) {
	c.A = *cc.E.GetPos()
	c.BER = cc.BER
	c.Status = cc.Status
}

type ConnecType interface {
	BitErrorRate(rx PhysReceiverInt)
	EvalRatioConnect() float64
	EvalRatioDisconnect() float64
	GetE() EmitterInt
	GetPr() float64
	EvalRatio(rx PhysReceiverInt) float64
	GetCh() int
	GetSNR() float64
}

func (co *Connection) GetCh() int       { return co.E.GetCh() }
func (co *Connection) GetE() EmitterInt { return co.E }
func (co *Connection) GetPr() float64   { return co.Pr }
func (co *Connection) GetSNR() float64  { return co.SNR }


func (co *Connection) BitErrorRate(rx PhysReceiverInt) {

	_, co.BER, co.SNR, co.Pr = rx.EvalInstantBER(co.GetE())

	co.meanPr.Add(co.Pr)
	co.meanBER.Add(co.BER)
	co.meanSNR.Add(co.SNR)

	co.Status = 1 //let mobile set master state		
	co.E.AddConnection(co)

}

func (co *Connection) EvalRatio(rx PhysReceiverInt) float64 {
	//_,SNR,_,_ := rx.EvalSignalSNR(co.E,co.E.GetCh())
	//return co.SNR
	return co.meanSNR.Get()
}

func (co *Connection) EvalRatioConnect() float64 {
	return co.E.BERT()
}

func (co *Connection) EvalRatioDisconnect() float64 {
	Ptot := co.E.BERT()
	return Ptot * math.Log(Ptot/co.meanBER.Get())
}

func CreateConnection(E EmitterInt, v float64) *Connection {
	Conn := new(Connection)
	Conn.E = E
	Conn.meanBER.Clear(v)
	Conn.Status = 1
	return Conn
}

