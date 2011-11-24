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

func ARBScheduler5(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64

	var MasterMobs [NConnec]*Emitter


	// Eval Metric for all connections
	var Nmaster int

	//Get SNR for all master mobiles all rbs

	r:= dbs.Pos

	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)
		E := c.GetE()

		if c.Status == 0 {

			MasterMobs[Nmaster] = E
			for rb := 1; rb < NCh; rb++ {
				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.GetSNRrb(rb)
				}
				Metric[Nmaster][rb] = snrrb
			}

			ICIMfunc(&r,E,Metric[Nmaster][:],dbs.Color)

			Nmaster++
		}
	}

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	//Assign RB for master connections

	//var NumAss [NConnec]int

	//	var Popul [popsize][NCh]int	
	//	var pool [popsize * 11][NCh]int

	var PopulAr [popsize][NCh]int
	var Popul [popsize]allocSet
	for i := range Popul {
		Popul[i].vect = PopulAr[i][:]
	}

	var poolAr [(popsize + 1) * generations][NCh]int
	var pool [(popsize + 1) * 10]allocSet
	for i := range pool {
		pool[i].vect = poolAr[i][:]
	}

	//First assign RB to best Metric
	for i := 0; i < popsize; i++ {
		nbrb := int(float64(NCh) / (float64(Nmaster))) //int(float64(NCh) * float64(r) / float64(numberAmobs))
		a := Rgen.Perm(Nmaster)
		//expand
		pp := Popul[i].vect
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
			createDesc(Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		for i := range Popul {
			copy(pool[popsize*10+i].vect, Popul[i].vect)
		}
		//copy(pool[popsize*10:popsize*11].vect, Popul[:].vect)

		//select the new population
		var metricpool [popsize * 11]float64

		for i := 0; i < popsize*11; i++ {
			var AL [NCh]int
			copy(AL[:], pool[i].vect) //copies 
			AL[0] = -1
			metricpool[i] = Trimm(AL[:], &Metric, MasterMobs[0:Nmaster])
		}

		//find the best 100
		S := initSequence(metricpool[:])
		sort.Sort(S)
		for i := 0; i < len(Popul); i++ {
			Popul[i] = pool[S.index[i]]
		}

	}

	AL := Popul[0].vect

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//Trimm(AL[:],&Metric,MasterMobs[0:Nmaster])	

	//testSCFDMA(AL)

	//copy(dbs.ALsave[:],AL)

	Allocate(AL[:], MasterMobs[0:Nmaster])
	//AllocateOld(AL[:],dbs)

}
