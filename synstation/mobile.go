package synstation

import "rand"


//  a mobile is an emitter with mobility : speed data, 
// it also is an agent and has an internal clock
type Mob struct {
	Emitter
	Speed [2]float64
	//R    PhysReceiver
	Rgen  *rand.Rand
	clock int
}


func (M *Mob) Init() {

	M.done = make(chan int)

	M.Rgen = rand.New(rand.NewSource(Rgen.Int63()))

	M.Requested = -3.0

	M.X = M.Rgen.Float64() * Field
	M.Y = M.Rgen.Float64() * Field
	M.Power = 1

	M.Ch = 1 // not connected // that's a trick to initialize SystemChannel emitters list
	M.SetCh(0)

	M.Speed[0] = (M.Rgen.Float64()*2 - 1) * MaxSpeed
	M.Speed[1] = (M.Rgen.Float64()*2 - 1) * MaxSpeed

	//M.R.X = M.X
	//M.R.Y = M.Y
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

	M.BERtotal, M.Diversity, M.MaxBER = M.SBERtotal, M.SDiversity, M.SMaxBER
	M.SBERtotal, M.SDiversity, M.SMaxBER = 0.0, 0, 0.0

	M.move()

	if M.BERtotal == 0 && M.Ch != 0 {
		M.MasterConnection = nil
		M.Power = 1
		M.SetCh(0)
	} else if M.MasterConnection != nil {
		M.MasterConnection.Status = 0 // we are master
	}

	if M.Ch == 0 {
		SyncChannel <- 0.0
	} else {
		SyncChannel <- M.BERtotal
	}
}

