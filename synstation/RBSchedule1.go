package synstation

import rand "math/rand"
import "math"

//import "fmt"
//import "geom"

type ARBScheduler1 struct {
	Metric     [NConnec][NCh]float64
	MasterMobs [NConnec]*Emitter

	AL     [NCh]int
	NumAss [NConnec]int
	CAL    [NCh]int
}

func initARBScheduler1() Scheduler {
	d := new(ARBScheduler1)
	return d
}

func (d *ARBScheduler1) Schedule(dbs *DBS, Rgen *rand.Rand) {

	Nmaster := 0

	r := dbs.Pos

	for _, e := 0, dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)
		if c.Status == 0 {

			E := c.E
			m_m := E.meanTR.Get()

			for rb := 0; rb < NCh; rb++ {

				snrrb := E.SNRrb[rb]

				if E.ARB[rb] {snrrb*=float64(E.GetNumARB())}

				m := EffectiveBW * math.Log2(1+beta*snrrb)

				if m > 100 {
					d.Metric[Nmaster][rb] = math.Log2(m + 1)
					b := (m_m + 1)
					d.Metric[Nmaster][rb] /= b
				}
			}
			d.MasterMobs[Nmaster] = E

			ICIMfunc(&r, E, d.Metric[Nmaster][:], dbs.Color)

			Nmaster++

		}
	}

	if Nmaster == 0 {
		return
	}

	//Assign RB for master connections

	MaxMP := 0.0
	for g := 0; g < 1; g++ {

		MetricBis := d.Metric

		//First assign RB to best Metric	
		MetricPool := 0.0
		RBrange := Rgen.Perm(NCh - NChRes) //add NChRes

		for j := range d.AL {
			d.AL[j] = -1
		}
		for j := range d.NumAss {
			d.NumAss[j] = 0
		}

		//for rb := 1; rb < NCh; rb++ {
		for _, rbo := range RBrange {
			rb := rbo + NChRes
			d.AL[rb] = -1
			for i := range d.MasterMobs[0:Nmaster] {
				if MetricBis[i][rb] > 0 {
					if d.AL[rb] < 0 {
						d.AL[rb] = i
					} else if MetricBis[i][rb] > MetricBis[d.AL[rb]][rb] {
						d.AL[rb] = i
					}
				}
			}
			if d.AL[rb] >= 0 {
				MetricPool += MetricBis[d.AL[rb]][rb]
				d.NumAss[d.AL[rb]]++ //this emitter will have one more assigned RB
				for rb2 := 1; rb2 < NCh; rb2++ {
					MetricBis[d.AL[rb]][rb2] *= float64(d.NumAss[d.AL[rb]]) / float64(d.NumAss[d.AL[rb]]+1)
				}
			}
		}
		if MetricPool > MaxMP {
			MaxMP = MetricPool
			d.CAL = d.AL
		}
	}

	Allocate(d.CAL[:], d.MasterMobs[0:Nmaster])
	//AllocateOld(AL[:], dbs )

}
