package synstation

import "math"
import "geom"
import "rand"
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

	filterAr [NCh]FilterInt //stores received power
	filterBr [NCh]FilterInt //stores received power
	ff_R     [NCh]float64   //stores received power with FF
	SNRrb    [NCh]float64   //stores received power with FF

	filterF FilterInt //stores received power

	Rgen *rand.Rand

	initz [2][NCh]float64 //generation of random number per RB	

	//BERrb   [NCh]float64 //stores results BER
}

// for output to save
func (c *Connection) GetInstantSNIR() []float64 {
	return c.ff_R[:]
}


type ConnectionS struct {
	A, B   geom.Pos
	Status int
	BER    float64
	Ch     int
}

func (c *ConnectionS) Copy(cc *Connection) {
	c.A = cc.E.GetPos()
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
	GetSNR() float64

	GetInstantSNIR() []float64

	//EvalInstantBER(E EmitterInt, rx PhysReceiverInt) (Rate, BER, SNR, Pr float64)
}

func (co *Connection) GetE() EmitterInt { return co.E }
func (co *Connection) GetPr() float64   { return co.Pr }
func (co *Connection) GetSNR() float64  { return co.SNR }


func (co *Connection) BitErrorRate(rx PhysReceiverInt) {

	co.evalInstantBER(co.GetE(), rx)

	co.Status = 1 //let mobile set master state		
	co.E.AddConnection(co)

}

func (co *Connection) EvalRatio(rx PhysReceiverInt) float64 {
	return co.meanSNR.Get()
}

func (co *Connection) EvalRatioConnect() float64 {
	return co.E.BERT()
}

func (co *Connection) EvalRatioDisconnect() float64 {
	Ptot := co.E.BERT()
	return Ptot * math.Log(Ptot/co.GetLogMeanBER())
}


func (Conn *Connection) InitConnection(E EmitterInt, v float64, Rgen *rand.Rand) {

	Conn.E = E
	Conn.meanBER.Clear(v)
	Conn.Status = 1

	Conn.Rgen = Rgen

	Speed := E.GetSpeed()
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000 for RB TTI

	if DopplerF < 0.002 { // the frequency is so low, a simple antena diversity will compensate for 	
		for i := 0; i < NCh; i++ {
			Conn.filterAr[i] = &PNF
			Conn.filterBr[i] = &PNF

		}
		Conn.filterF = &PNF
	} else {
		A := Butter(DopplerF)
		B := Cheby(10, DopplerF)
		C := MultFilter(A, B)

		Conn.filterF = CoherenceFilter.Copy()

		for i := 0; i < NCh; i++ {
			Conn.filterAr[i] = C.Copy()
			Conn.filterBr[i] = C.Copy()
		}

		// initalize filters : sent some values to prevent having empty values z^-n in filters 
		// 2.5/DopplerF gives a good number of itterations to decorelate initial null variables
		// and set the channel to a steady state
		for l := 0; l < int(2.5/DopplerF); l++ {
			for i := 0; i < 50; i++ {
				Conn.filterF.nextValue(Conn.Rgen.NormFloat64())
			}
			for i := 0; i < NCh; i++ {
				Conn.initz[0][i] = Conn.filterF.nextValue(Conn.Rgen.NormFloat64())
			}
			for i := 0; i < 50; i++ {
				Conn.filterF.nextValue(Conn.Rgen.NormFloat64())
			}
			for i := 0; i < NCh; i++ {
				Conn.initz[1][i] = Conn.filterF.nextValue(Conn.Rgen.NormFloat64())
			}

			for i := 0; i < NCh; i++ {
				Conn.filterAr[i].nextValue(Conn.initz[0][i])
				Conn.filterBr[i].nextValue(Conn.initz[1][i])
			}

		}

	}

}


func CreateConnection(E EmitterInt, v float64, Rgen *rand.Rand) *Connection {
	Conn := new(Connection)
	Conn.InitConnection(E, v, Rgen)
	return Conn
}

func (co *Connection) GetLogMeanBER() float64 {
	return math.Log10(co.meanBER.Get() + 1e-40) //prevent saturation
}


//This function is only called once per iteration, so it is where the FF value is generated
func (c *Connection) evalInstantBER(E EmitterInt, rx PhysReceiverInt) {

	ARB := E.GetARB()

	if E.IsSetARB(0) {
		_, c.BER, c.SNR, c.Pr = rx.EvalSignalBER(E, 0)

		c.meanBER.Add(c.BER)
		c.meanSNR.Add(c.SNR)
		c.meanPr.Add(c.Pr)
		return
	}

	//Generate DopplerFading
	//pass some values to decorelate
	for i := 0; i < 50; i++ {
		c.filterF.nextValue(c.Rgen.NormFloat64())
	}

	for i := 0; i < NCh; i++ {
		c.initz[0][i] = c.filterF.nextValue(c.Rgen.NormFloat64())
	}
	//pass some values to decorelate
	for i := 0; i < 50; i++ {
		c.filterF.nextValue(c.Rgen.NormFloat64())
	}

	for i := 0; i < NCh; i++ {
		c.initz[1][i] = c.filterF.nextValue(c.Rgen.NormFloat64())
	}

	K := rx.GetK(E.GetId())

	for rb := 0; rb < NCh; rb++ {
		a := c.filterAr[rb].nextValue(c.initz[0][rb]) + K
		b := c.filterBr[rb].nextValue(c.initz[1][rb])
		c.ff_R[rb] = (a*a + b*b) / 2
	}

	for rb, use := range ARB {

		if use {

			Pr, Rc := rx.GetPr(E.GetId(), rb)

			c.Pr = Pr // to save data to file

			c.SNRrb[rb] = Pr * c.ff_R[rb] / (Rc.Pint - Pr + WNoise)
			BER := L1 * math.Exp(-c.SNRrb[rb]/2/L2) / 2.0

			c.meanBER.Add(BER)
			c.meanPr.Add(Pr)
			c.meanSNR.Add(c.SNR)

			c.SNR = c.SNRrb[rb]
		}
	}

	return
}

