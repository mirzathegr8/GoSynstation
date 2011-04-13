package synstation

import "rand"
//import "fmt"
import "math"


//  a mobile is an emitter with mobility : speed data, 
// it also is an agent and has an internal clock
type Mob struct {
	Emitter
	//R    PhysReceiver
	Rgen  *rand.Rand
	clock int

	//meanBERInstTot MeanData
}


func (M *Mob) Init(i int) {

	M.Id = i
	M.Rgen = rand.New(rand.NewSource(Rgen.Int63()))

	M.Requested = -15.0

	M.X = M.Rgen.Float64() * Field
	M.Y = M.Rgen.Float64() * Field
	M.Power = 1

	M.SetARB(0) // start trying to connect

	speed := M.Rgen.Float64() * MaxSpeed / 1000
	//speed := float64(0.00)
	angle := M.Rgen.Float64() * 2 * math.Pi
	M.Speed[0] = speed * math.Cos(angle)
	M.Speed[1] = speed * math.Sin(angle)

}

func (M *Mob) getInt() EmitterInt {
	return M
}


func (M *Mob) move() {

	M.X += M.Speed[0]
	M.Y += M.Speed[1]

	if M.X < 0 {
		M.Speed[0] = -M.Speed[0]
		M.X = 0
	} else if M.X > Field {
		M.Speed[0] = -M.Speed[0]
		M.X = Field
	}

	if M.Y < 0 {
		M.Speed[1] = -M.Speed[1]
		M.Y = 0
	} else if M.Y > Field {
		M.Speed[1] = -M.Speed[1]
		M.Y = Field
	}

}


func (M *Mob) RunPhys() {

	M.clock = M.Rgen.Intn(15)
	if M.clock == 1 {
		//	M.R.MeasurePower(M)
	}

	SyncChannel <- 1

}

// this function saves totalBER and inform which dbs connection is master
// reset power and channel if all connections losts
// applies agent functionality (move mobiles)
// finnaly sents to syncchannel BER level
func (M *Mob) FetchData() {

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

			M.TransferRate += 80 * math.Log2(1+M.SSNRrb[rb])

			if 100 < M.TransferRate {

				M.Outage = 0

			} else {
				//M.Rate--
				M.TransferRate = 0

			}

			//M.TransferRate = math.Log10(M.TransferRate)


		}
		M.SSNRrb[rb] = 0
	}

	if M.BERtotal == 0 && !M.IsSetARB(0) {
		M.MasterConnection = nil
		M.Power = 1
		M.ReSetARB()
	} else if M.MasterConnection != nil {
		M.MasterConnection.Status = 0 // we are master
	}

	if M.IsSetARB(0) {
		SyncChannel <- 0.0
	} else {
		SyncChannel <- M.BERtotal
	}

	M.move()
}

