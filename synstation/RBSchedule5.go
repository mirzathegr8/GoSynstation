package synstation

import "sort"
import rand "math/rand"
import "fmt"

//import "geom"
//import "math"

//const uARBcost = 0.01 //meanMeanCapa / 5 //0.5 // math.Log2(1 + meanMeanCapa)

func init() {

	fmt.Println("init to keep fmt")
}


// memory for the scheduler
type ARBScheduler5 struct {
	Metric [NConnec][NCh]float64
	metricpool [popsize * 11]float64
	index [popsize * 11]int
	S Sequence
	MasterMobs [NConnec]*Emitter

	PopulAr [popsize][NCh]int
	Popul [popsize]allocSet
	
	poolAr [(popsize + 1) * generations][NCh]int
	pool [(popsize + 1) * 10]allocSet
	AL [NCh]int
}

func initARBScheduler5() Scheduler {
	d := new(ARBScheduler5)
	d.S.index = d.index[0 : popsize*11]
	for i := range d.Popul {
		d.Popul[i].vect = d.PopulAr[i][:]
	}
	for i := range d.pool {
		d.pool[i].vect = d.poolAr[i][:]
	}

	return d
}

func (d *ARBScheduler5) Schedule(dbs *DBS, Rgen *rand.Rand) {


	// Eval Metric for all connections
	var Nmaster int

	//Get SNR for all master mobiles all rbs

	r := dbs.Pos

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)
		E := c.E

		if c.Status == 0 {

			d.MasterMobs[Nmaster] = E
			for rb := 1; rb < NCh; rb++ {
				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.SNRrb[rb]
				}
			if E.ARB[rb] {snrrb*=float64(E.GetNumARB())} //since we redivide the power according to the new number of allocated RBs
		
				d.Metric[Nmaster][rb] = snrrb
			}

			ICIMfunc(&r, E, d.Metric[Nmaster][:], dbs.Color)

			Nmaster++
		}
	}

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	

	//First assign RB to best Metric
	for i := 0; i < popsize; i++ {
		nbrb := int(float64(NCh) / (float64(Nmaster))) //int(float64(NCh) * float64(r) / float64(numberAmobs))
		a := Rgen.Perm(Nmaster)
		//expand
		pp := d.Popul[i].vect
		// first dealocate everything
		for j := 0; j < NCh; j++ {
			pp[j] = -1
		}
		//second alocate one chanel
		for j := 0; j < Nmaster; j++ {
			index := int(float64(NCh) / float64(Nmaster) * float64(j))
			for k := 0; k < nbrb; k++ {
				pp[index+k] = a[j]
			}
		}
	}

	for gen := 0; gen < generations; gen++ {

		for j := 0; j < popsize; j++ {
			createDesc(d.Popul[j], d.pool[j*10:(j+1)*10], Rgen)
		}

		for i := range d.Popul {
			copy(d.pool[popsize*10+i].vect, d.Popul[i].vect)
		}
		//copy(pool[popsize*10:popsize*11].vect, Popul[:].vect)

		//select the new population

		for i := 0; i < popsize*11; i++ {
			
			copy(d.AL[:], d.pool[i].vect) //copies 
			d.AL[0] = -1
			d.metricpool[i] = Trimm(d.AL[:], &d.Metric, d.MasterMobs[0:Nmaster])
		}

		//find the best 100
		initSequence(d.metricpool[:], &d.S)
		sort.Sort(d.S)
		for i := 0; i < len(d.Popul); i++ {
			d.Popul[i] = d.pool[d.S.index[i]]
		}

	}

	ALf := d.Popul[0].vect

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//Trimm(AL[:],&Metric,MasterMobs[0:Nmaster])	

	//testSCFDMA(AL)

	//copy(dbs.ALsave[:],AL)

	Allocate(ALf[:], d.MasterMobs[0:Nmaster])
	//AllocateOld(AL[:],dbs)

}
