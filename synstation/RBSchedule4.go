package synstation

import "sort"
import rand "math/rand"
import "math"
import "fmt"

type Scheduler interface {
	Schedule(dbs *DBS, Rgen *rand.Rand)
}

func init() {

	fmt.Println("init to keep fmt")
}

// memory for the scheduler
type ARBScheduler4 struct {
	metricpool [popsize * 11]float64 // keeps metric for each allocation tested in
	// 					curent generation and its descendants
	index        [popsize * 11]int     // keeps index for sorting
	S            Sequence              // Sequence for the sort
	Metric       [NConnec][NCh]float64 // data to evaluate metric for each RB and each mobiles
	MasterMobs   [NConnec]*Emitter     // list of master connection mobiles
	MasterConnec [NConnec]*Connection  // list of the respective connections to these master connected mobiles

	PopulAr [popsize][NCh]int //current generation of popsize allocation
	Popul   [popsize]allocSet // just a set of reference to the previous pool, 
	//				in order to pass these data to subroutines
	poolAr [popsize * 11][NCh]int // the pool to hold all descandants plus the current generation 
	pool   [popsize * 11]allocSet // just refs to the previous pool
	AL     [NCh]int               // a temporary allocation vector
}

func initARBScheduler4() Scheduler {
	d := new(ARBScheduler4)
	d.S.index = d.index[0 : popsize*11]
	return d
}

func (d *ARBScheduler4) Schedule(dbs *DBS, Rgen *rand.Rand) {

	var Nmaster int
	var meanMeanCapa float64
	var maxMeanCapa float64
	var minPower float64

	// we first collect SNRs for the metric evaluation later
	for j, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.E
		m_m := E.meanTR.Get()
		meanMeanCapa += m_m

		if maxMeanCapa < m_m {
			maxMeanCapa = m_m
		}

		if minPower > math.Log10(c.meanPr.Get()) {
			minPower = c.meanPr.Get()
		}

		if c.Status == 0 {

			d.MasterConnec[Nmaster] = c
			d.MasterMobs[Nmaster] = E

			//for i:=0; i<NCh; i++{E.UnSetARB(i)}

			for rb := 1; rb < NCh; rb++ {
				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.SNRrb[rb]
				}
				d.Metric[j][rb] = snrrb
			}
			Nmaster++
			j++
		}
	}

	meanMeanCapa /= float64(dbs.Connec.Len())
	meanMeanCapa += 0.1

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	// create the refs to the pools		
	for i := range d.Popul {
		d.Popul[i].vect = d.PopulAr[i][:]
	}
	for i := range d.pool {
		d.pool[i].vect = d.poolAr[i][:]
	}

	//Create #popsize initial allocations
	for i := 0; i < popsize; i++ {
		nbrb := int(float64(NCh) / (float64(Nmaster))) // subset size
		//int(float64(NCh) * float64(r) / float64(numberAmobs))
		a := Rgen.Perm(Nmaster)    // ordering of master connected mobiles
		pp := d.Popul[i].vect      // vector allocation we work on		
		for j := 0; j < NCh; j++ { // first dealocate everything
			pp[j] = -1
		}
		//second alocate subsets of rbs to each master mobiles, in random order
		for j := 0; j < Nmaster; j++ {
			index := int(float64(NCh) / float64(Nmaster) * float64(j))
			for k := 0; k < nbrb; k++ {
				pp[index+k] = a[j]
			}
		}
	}

	//Iterate over generations
	for gen := 0; gen < generations; gen++ {

		for j := 0; j < popsize; j++ { // create 10 descendants for each
			createDesc(d.Popul[j], d.pool[j*10:(j+1)*10], Rgen)
		}

		for i := range d.Popul { //copy the current generation inside the main pool with its descendants
			copy(d.pool[popsize*10+i].vect, d.Popul[i].vect)
		}

		for i := 0; i < popsize*11; i++ { // evaluate the metric for each potential allocation
			copy(d.AL[:], d.pool[i].vect) // we copy since Trimm() can modify the allocation with trimming
			d.AL[0] = -1
			d.metricpool[i] = Trimm(d.AL[:], &d.Metric, d.MasterMobs[0:Nmaster])
		}

		//find the best 100
		initSequence(d.metricpool[:], &d.S)
		sort.Sort(d.S)
		for i := 0; i < len(d.Popul); i++ { // and copy the best 100 as survivors to next generation
			d.Popul[i] = d.pool[d.S.index[i]]
		}

	}

	Trimm(d.Popul[0].vect, &d.Metric, d.MasterMobs[0:Nmaster])
	Allocate(d.Popul[0].vect, d.MasterMobs[0:Nmaster]) // allocate the best 

}

func Trimm(AL []int, Metric *[NConnec][NCh]float64, MasterMobs []*Emitter) (metricT float64) {

	AL[0] = -1

	for i := 1; i < len(AL); {

		v := AL[i]
		switch v {
		case -1:
			metricT += uARBcost
			i++
			// just go on
		default:
			Mvect := Metric[v][0:NCh]

			//find max
			jmin := i
			var max float64
			nARBm := 0

			for j, vAL := range AL[jmin:len(AL)] {
				if vAL != v {
					break
				}
				nARBm++
				if Mvect[j] > max {
					max = Mvect[j]
				}

			}
			//loop to trim
			// here we only consider the Original metric 

			nARBmOrg := nARBm
			jmax := jmin + nARBm //save first

			for _, Mval := range Mvect[jmin:jmax] { //j:=jmin; j<jmax ;j++ { //

				//Mval:=Mvect[j]
				m := EffectiveBW * math.Log2(1+Mval/float64(nARBm))

				if Mval < max/CAPAthres || m < 100 {
					AL[jmin] = -1
					nARBm--
					metricT += uARBcost
					jmin++
				} else {
					break
				}
			}

			for j := jmax - 1; j >= jmin; j-- {

				Mval := Mvect[j]
				m := EffectiveBW * math.Log2(1+Mval/float64(nARBm))

				if Mval < max/CAPAthres || m < 100 {
					AL[j] = -1
					nARBm--
					metricT += uARBcost
					jmax--
				} else {
					break
				}
			}

			m := float64(0.0)
			for _, Mval := range Mvect[jmin:jmax] {
				m += EffectiveBW * math.Log2(1+beta*Mval/float64(nARBm))
			}

			m_m := MasterMobs[v].meanTR.Get()
			metricT += math.Log2(1+m) / math.Log2(1+m_m+0.0001)

			i += nARBmOrg
		}
	}

	return
}

func Allocate(AL []int, MasterMobs []*Emitter) {

	AL[0] = -1 // connect all
	for _, M := range MasterMobs {
		M.ClearFuturARB()
	}
	for rb, vAL := range AL {
		if vAL >= 0 {
			if !MasterMobs[vAL].ARB[rb] {
				Hopcount++
			}
			MasterMobs[vAL].ARBfutur[rb] = true
		}
	}

}

func testSCFDMA(AL []int) int {

	sv := -1

	for i := 0; i < NCh; i++ {
		if sv == -1 {
			sv = AL[i]
		}
		if sv != AL[i] {
			for j := i; j < NCh; j++ {
				if AL[j] == sv {
					return -1
				}
			}
			sv = AL[i]
		}
	}

	return 0

}
