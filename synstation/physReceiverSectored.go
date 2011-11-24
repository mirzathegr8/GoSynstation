package synstation

import "geom"
import "container/list"
import rand "math/rand"

// structure to store evaluation of interference at a location
// this has to be initialized with PhysReceiver.Init() function to init memory
type PhysReceiverSectored struct {
	PhysReceiverSDMA
}

func (r *PhysReceiverSectored) Init(p geom.Pos, Rgen *rand.Rand) {
	r.R = make([]PhysReceiver, 3)
	for i := 0; i < 3; i++ {
		r.R[i].Init(p, Rgen)
	}

	for i := 0; i < len(r.R[0].Orientation); i++ {
		r.R[0].Orientation[i] = 0
		r.R[1].Orientation[i] = PI2/3.0
		r.R[2].Orientation[i] = PI2*2/3.0
	}
}

func (rx *PhysReceiverSDMA) SDMA(Connec *list.List){

}

