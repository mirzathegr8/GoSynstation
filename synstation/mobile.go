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
}


func (M *Mob) Init(i int) {

	M.Id = i
	M.Rgen = rand.New(rand.NewSource(Rgen.Int63()))

	M.Requested = -7.0

	M.X = M.Rgen.Float64() * Field
	M.Y = M.Rgen.Float64() * Field
	M.Power = 1

	M.ARB.Resize(1, 1)
	M.ARB.Set(0, 0) //Ch = 0

	SystemChan[0].Change <- &M.Emitter

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

	if M.SBERtotal == 0 {
		M.Outage++
	} else {
		M.Outage = 0
	}

	M.BERtotal, M.Diversity, M.MaxBER, M.InstMaxBER = M.SBERtotal, M.SDiversity, M.SMaxBER, M.SInstMaxBER
	M.SInstMaxBER, M.SBERtotal, M.SDiversity, M.SMaxBER = 0, 0, 0, 0

	M.move()

	if M.BERtotal == 0 && M.ARB[0] != 0 {
		M.MasterConnection = nil
		M.Power = 1
		M.SetCh(0)
	} else if M.MasterConnection != nil {
		M.MasterConnection.Status = 0 // we are master
	}

	if M.ARB[0] == 0 {
		SyncChannel <- 0.0
	} else {
		SyncChannel <- M.BERtotal
	}
}

