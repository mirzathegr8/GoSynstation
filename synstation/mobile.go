package synstation

import rand "math/rand"
import "math"

//  a mobile is an emitter with mobility : speed data, 
// it also is an agent and has an internal clock
type Mob struct {
	Emitter //Emitter is a structure defined in emitter.go file.
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
	M.SetPower(1.0)

	M.ReSetARB() // start trying to connect

	speed := M.Rgen.Float64() * MaxSpeed / 1000
	//speed := float64(0.00)
	angle := M.Rgen.Float64() * 2 * math.Pi
	M.Speed[0] = speed * math.Cos(angle)
	M.Speed[1] = speed * math.Sin(angle)

	M.Data = 4e6

	M.Emitter.Init(M.Rgen)

	SyncChannel <- 1
}

// 	applies agent functionality (move mobiles)
func (M *Mob) RunAgent() {

	//Move the mobile
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

	if Tti >= 0 {
		M.Data -= int(M.TransferRate)
		if M.Data < 0 {
			M.Data += 200
		}
	}
	//	if M.Data<2 { M.Power=0}

	SyncChannel <- 1.0
}

func (M *Mob) RunPhys() {

	//M.clock = M.Rgen.Intn(15)
	//if M.clock == 1 {
	//	M.R.MeasurePower(M)
	//}

	SyncChannel <- 1

}
