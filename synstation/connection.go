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

	filterF FilterInt //stores received power

	Rgen *rand.Rand

	initz [2][NCh]float64 //generation of random number per RB	
}

// for output to save
func (c *Connection) GetffR() []float64 {
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
	GetCh() int
	GetSNR() float64

	EvalInstantBER(E EmitterInt, rx PhysReceiverInt) (Rc *ChanReceiver, BER, SNR, Pr float64)
}

func (co *Connection) GetCh() int       { return co.E.GetCh() }
func (co *Connection) GetE() EmitterInt { return co.E }
func (co *Connection) GetPr() float64   { return co.Pr }
func (co *Connection) GetSNR() float64  { return co.SNR }


func (co *Connection) BitErrorRate(rx PhysReceiverInt) {

	_, co.BER, co.SNR, co.Pr = co.EvalInstantBER(co.GetE(), rx)

	co.meanPr.Add(co.Pr)
	co.meanBER.Add(co.BER)
	co.meanSNR.Add(co.SNR)

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
	DopplerF := Speed * F / cel // 1000 samples per seconds speed already divided by 1000

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
		for l := 0; l < 50; l++ {
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
func (c *Connection) EvalInstantBER(E EmitterInt, rx PhysReceiverInt) (Rc *ChanReceiver, BER, SNR, Pr float64) {

	if E.GetCh() == 0 {
		Rc, BER, SNR, Pr = rx.EvalSignalBER(E, 0)

	} else {
		var K float64
		Pr, K, Rc = rx.GetPrK(E.GetId(), E.GetCh())

		//c.filterF.InitRandom(c.Rgen) // = c.Rgen.NormFloat64()
		//pass some values to decorelate
		for i := 0; i < 50; i++ {
			c.filterF.nextValue(c.Rgen.NormFloat64())
		}

		for i := 0; i < NCh; i++ {
			c.initz[0][i] = c.filterF.nextValue(c.Rgen.NormFloat64())
		}
		//c.filterF.InitRandom(c.Rgen) // = c.Rgen.NormFloat64()
		//pass some values to decorelate
		for i := 0; i < 50; i++ {
			c.filterF.nextValue(c.Rgen.NormFloat64())
		}

		for i := 0; i < NCh; i++ {
			c.initz[1][i] = c.filterF.nextValue(c.Rgen.NormFloat64())
		}

		for i := 0; i < NCh; i++ {
			a := c.filterAr[i].nextValue(c.initz[0][i]) + K
			b := c.filterBr[i].nextValue(c.initz[1][i])
			c.ff_R[i] = math.Sqrt(a*a + b*b)
		}

		//fmt.Println(c.initz[0])

		SNR = 0

		//at this moment the 0.0789 is to normalise the c.ff_R ratio to have a Rayleigh of sigma=1

		SNR = c.ff_R[E.GetCh()] * 0.0789 * Pr / (Rc.Pint - Pr + WNoise)

		Pr = Pr * c.ff_R[E.GetCh()] * 0.0789

		if SNR > 4000 {
			SNR = 4000
		}

		BER = L1 * math.Exp(-SNR/2/L2) / 2.0
	}
	return
}

