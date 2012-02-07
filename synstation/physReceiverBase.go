package synstation

import "math"
import "geom"
import "container/list"
import rand "math/rand"

// structure to store evaluation of interference at a location
// this has to be initialized with PhysReceiver.Init() function to init memory
type PhysReceiverBase struct {
	geom.Pos
	shadow shadowMapInt
	Rgen   *rand.Rand

	kk [M]float64 //stores k rice factor
	pr [M]float64 //stores base level received power with shadowing and distance,
	//no powerlevel nor fast fading (dependant on RB) 
	AoA [M]float64

	RBsReceiver
}

// Initialise the receiver:
func (r *PhysReceiverBase) Init() {

	r.X = Rgen.Float64() * Field
	r.Y = Rgen.Float64() * Field
	r.Rgen = rand.New(rand.NewSource(Rgen.Int63()))

	switch SetShadowMap {
	case NOSHADOW:
		r.shadow = new(noshadow)
	case SHADOWMAP:
		r.shadow = new(shadowMap)
	}
	r.shadow.Init(corr_res, r.Rgen)
}

// Getters/Setters
func (r *PhysReceiverBase) SetPos(p geom.Pos) {
	r.Pos = p
}
func (r *PhysReceiverBase) GetPos() geom.Pos {
	return r.Pos
}

func (rx *PhysReceiverBase) Compute(Connec *list.List) {

	//*********************************
	//Evaluate recevied power
	for m := range Mobiles {
		E := &Mobiles[m]
		// inline minus
		p := geom.Pos{E.X - rx.X, E.Y - rx.Y}
		//Calculate Distance, Fading parameter K, and Fading
		//d := rx.DistanceSquare(Mobiles[i].Pos)

		d := (p.X*p.X + p.Y*p.Y)
		d += 2
		K := 1 / d
		d *= d

		rx.kk[m] = K
		rx.pr[m] = rx.shadow.evalShadowFading(p) / d

		theta := math.Atan2(p.Y, p.X)
		if theta < 0 {
			theta += PI2
		}
		rx.AoA[m] = theta

		prRB := rx.pr[m] / math.Max(1.0, float64(E.GetNumARB()))
		for rb, use := range Mobiles[m].ARB {
			if use {
				rx.Channels[rb].pr[m] = prRB * E.Power[rb]
			}
		}
	}
	rx.SumInterference()
}

//Returns K value and base level received power (used for estimating potential on other channels)
func (rx PhysReceiverBase) GetK(i int) (k float64) {
	k = rx.kk[i]
	return
}

func (rx PhysReceiverBase) GetPrBase(i int) (p float64) {
	p = rx.pr[i]
	return
}

func (rx PhysReceiverBase) GetPr(m, rb int) float64 {
	return rx.Channels[rb].pr[m]
}
