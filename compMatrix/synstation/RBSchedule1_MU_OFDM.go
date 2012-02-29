package synstation

import rand "rand"
import "math"
//import "fmt"
//import "geom"

type ARBSchedulerMU_OFDM struct {
	Metric      [NConnec][NCh]float64
	MasterMobs  [NConnec]*Emitter
	MasterMobsB [NConnec]*Emitter

	AL     [NCh]int
	NumAss [NConnec]int
	CAL    [NCh]int

	means [mDf]float64
	Max   [mDf]float64
	MaxId [mDf]int
	NumE  [mDf]int

	angles     [NConnec]float64
	Centers    [mDf]float64 //stores angles
	anglesDiff [mDf][NConnec]float64
	subs       [NConnec]int
}

func initARBSchedulerMU_OFDM() Scheduler {
	d := new(ARBSchedulerMU_OFDM)
	return d
}

func (d *ARBSchedulerMU_OFDM) Schedule(dbs *DBS, Rgen *rand.Rand) {

	Nmaster := 0

	r := dbs.Pos

	for _, e := 0, dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)
		if c.Status == 0 {

			E := c.E
			m_m := E.meanTR.Get()

			for rb := 0; rb < NCh; rb++ {

				snrrb := E.SNRrb[rb]

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

	for _, M := range d.MasterMobs[0:Nmaster] {
		M.ClearFuturARB()
	}

	//Assign RB for master connections

	//create mDf clusters


	mDfT := mDf
	//clustering
	if Nmaster > mDf {
		// initialisation
		for m := 0; m < Nmaster; m++ {
			d.angles[m] = dbs.AoA[d.MasterMobs[m].Id]
		}
		for i := range d.Centers {
			d.Centers[i] = float64(i) / mDf * PI2
		}

		//iterations
		for k := 0; k < 10; k++ {

			//compute angle diferences
			for m := 0; m < Nmaster; m++ {
				for i := range d.Centers {
					d.anglesDiff[i][m] = math.Fmin(math.Fabs(d.angles[m]-d.Centers[i]),
						math.Fabs(-d.angles[m]+d.Centers[i]))

				}
			}

			// group mobiles to nearest center
			for m := 0; m < Nmaster; m++ {
				d.subs[m] = 0
				for i := 1; i < mDf; i++ {
					if d.anglesDiff[i][m] < d.anglesDiff[d.subs[m]][m] {
						d.subs[m] = i
					}
				}
			}
			// re-evaluate centers without their extrems

			for i := range d.means {
				d.means[i] = 0
				d.Max[i] = 0
				d.MaxId[i] = 0
				d.NumE[i] = 0
			}
			for m := 0; m < Nmaster; m++ {
				d.means[d.subs[m]] += d.angles[m]
				d.NumE[d.subs[m]]++
				if d.Max[d.subs[m]] < d.anglesDiff[d.subs[m]][m] {
					d.Max[d.subs[m]] = d.anglesDiff[d.subs[m]][m]
					d.MaxId[d.subs[m]] = m
				}
			}
			// reposition centers
			for i := range d.means {
				if d.NumE[i] > 1 {
					d.Centers[i] = (d.means[i] - d.angles[d.MaxId[i]]) / float64(d.NumE[i]-1)
				} else if d.NumE[i] == 1 {
					d.Centers[i] = d.means[i]
				} else { // NumE==0 
					d.Centers[i] = d.Centers[(i+1)%mDf] + Rgen.NormFloat64()/10
				}
			}

		}
	} else {
		for m := 0; m < Nmaster; m++ {
			d.subs[m] = m
		}
		mDfT = Nmaster
	}

	//fmt.Println(d.NumE)

	//for all subgroups
	NmasterB := 0
	for i := 0; i < mDfT; i++ {

		for m := 0; m < Nmaster; m++ {
			if d.subs[m] == i {
				d.MasterMobsB[NmasterB] = d.MasterMobs[m]
				NmasterB++
			}
		}
		if NmasterB > 0 {

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

				for _, rbo := range RBrange {
					rb := rbo + NChRes
					d.AL[rb] = -1
					for i := range d.MasterMobsB[0:NmasterB] {
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
							if d.NumAss[d.AL[rb]] > 15 {
								MetricBis[d.AL[rb]][rb2] = 0
							} else {
								MetricBis[d.AL[rb]][rb2] *= float64(d.NumAss[d.AL[rb]]) / float64(d.NumAss[d.AL[rb]]+1)
							}
						}
					}
				}
				if MetricPool > MaxMP {
					MaxMP = MetricPool
					d.CAL = d.AL
				}
			}

			//allocate subsets
			for rb, vAL := range d.CAL {
				if vAL >= 0 {
					if !d.MasterMobsB[vAL].ARB[rb] {
						Hopcount++
					}
					d.MasterMobsB[vAL].ARBfutur[rb] = true
				}
			}
			//Allocate(d.CAL[:], d.MasterMobsB[0:NmasterB])
		}
	}

}


