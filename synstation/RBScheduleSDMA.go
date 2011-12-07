package synstation

import "sort"
import rand "rand"
import "math"
import "fmt"



func init() {

	fmt.Println("init to keep fmt")
}

/*

split the mobiles in mDf groups
run mDf genetic modification in paralel
add cross modification between groups

evaluate metric combined
continue

*/

type SDMA_alloc struct {
	vect *[mDf][NCh]int
	subSize [mDf]int
}

func ARBSchedulerSDMA(dbs *DBS, Rgen *rand.Rand) {

	var SNRrbAll [NConnec][NCh]float64
	var MasterMobs [NConnec]*Emitter
	var MasterConnec [NConnec]*Connection
	var Pr [NConnec]float64

	// Eval Metric for all connections
	var Nmaster int

	//Get SNR for all master mobiles all rbs	

	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.GetE()

		if c.Status == 0 {

			MasterConnec[Nmaster] = c
			MasterMobs[Nmaster] = E

			for rb := 0; rb < NCh; rb++ {
				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.SNRrb[rb]
				}
				SNRrbAll[Nmaster][rb] = snrrb
				Pr[Nmaster] = dbs.pr[E.Id] * math.Fmax(1,float64(E.GetNumARB())) //total power before splitted over RBs
			}
			Nmaster++

		}
	}

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	var PopulAr [popsize][mDf][NCh]int
	var Popul [popsize]SDMA_alloc 
	for i := range Popul {
		Popul[i].vect = &PopulAr[i]
	}

	var poolAr [(popsize + 1) * generations][mDf][NCh]int
	var pool [(popsize + 1) * 10]SDMA_alloc 
	for i := range pool {
		pool[i].vect = &poolAr[i]
	}

	// create the initial random partitions on mDf levels
	for i := 0; i < popsize; i++ {

		//first, create subgroups
		var subGroups [mDf][NConnec]int
		a := Rgen.Perm(Nmaster)

		var k,l int
		for _, v := range a {			
			subGroups[k][l] = v
			k++
			if k>=mDf {k=0; l++}
			
		}
		

		for s := 0; s < mDf; s++ {
			
			NSubMaster := Nmaster/mDf + int(math.Ceil( float64((Nmaster - i)%mDf)/float64(mDf)))
			
			nbrb := int(float64(NCh) / (float64(NSubMaster)))
			a := Rgen.Perm(NSubMaster)
			//expand
			pp := &Popul[i].vect[s]
			Popul[i].subSize[s]=NSubMaster
			// first dealocate everything
			for j := 0; j < NCh; j++ {
				pp[j] = -1
			}
			//second alocate contiguous chanel subsets
			for j := 0; j < NSubMaster; j++ {
				index := int(float64(NCh) / float64(NSubMaster) * float64(j))
				for k := 0; k < nbrb; k++ {
					pp[index+k] = subGroups[s][a[j]]
				}
			}

		}
	}



	// TODO

	for gen := 0; gen < generations; gen++ {

		for j := 0; j < popsize; j++ {
			createDescSDMA(Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		for i := range Popul {
			for s := 0; s < mDf; s++ {
				//copy(pool[popsize*10+i].vect[s], Popul[i].vect[s])
				pool[popsize*10+i].vect[s]=Popul[i].vect[s]
			}
		}

		//select the new population
		var metricpool [popsize * 11]float64

		for i := 0; i < popsize*11; i++ {
			for s := 0; s < mDf; s++ {
				pool[i].vect[s][0] = -1
			}
			metricpool[i] = MetricSDMA(pool[i].vect,&SNRrbAll, &Pr, MasterConnec[0:Nmaster])
		}

		//find the best 100
		S := initSequence(metricpool[:])
		sort.Sort(S)
		for i := 0; i < len(Popul); i++ {
			Popul[i] = pool[S.index[i]]
		}

	}

	AL := Popul[0].vect

	//fmt.Println(Nmaster, AL)

	AllocateSDMA(AL, MasterMobs[0:Nmaster])

}

/*
*	AL [mDf][NCh] allows mDf allocation of on RB to mobiles for
*	
*
 */
func MetricSDMA(AL *[mDf][NCh]int, SNRrbAll *[NConnec][NCh]float64, Pr *[NConnec]float64, MasterConn []*Connection) (metric float64){

	var SNRres [NConnec][NCh]float64
	var NumARB [NConnec]int
	var Corr [NConnec][NConnec]float64

	// rememeber the number of allocated RB with the porposed allocation
	for _, ALsub := range AL {
		for _, v := range ALsub {
			if v > -1 {
				NumARB[v]++
			}
		}
	}

	// remember the intra-interference ratios
	for m1, C1 := range MasterConn {
		for m2, C2 := range MasterConn {

			//a := C1.gainM[C1.E.Id]
			b := C1.gainM[C2.E.Id]

			Corr[m1][m2] = b /// a

		}
	}

	// inverse the estimated SINR per rb to process for new allocation
	for v1,C1 := range MasterConn {
		for rb := range C1.Channels[:] {
			SNRres[v1][rb] = C1.Channels[rb].Pint - C1.Channels[rb].pr[C1.E.Id] 
			//alternative : 1.0 / SNRrbAll[v1][rb]
		}
	}

	//substract power received from local master mobiles on the same rb
	for v1, C1 := range MasterConn {

		for rb := range C1.Channels[:] {

			for v2, C2 := range MasterConn {
				if C2.E.ARB[rb] && v2!=v1 {
					SNRres[v1][rb] -= C1.Channels[rb].pr[C2.E.Id]

				}
			}
		}

	}

	// add the possible intra-interference genereated by the new allocation
	for i, ALsub := range AL {
		for rb, v1 := range ALsub {
			if v1 > -1 {
				for j := 0; j < mDf; j++ {
					v2 := AL[j][rb]
					if v2 > -1 {
						if i != j {
		SNRres[v1][rb] += Pr[v2] / (float64(NumARB[v2])) * Corr[v1][v2]
						}

					}
				}
			}
		}
	}

	var metricM [NConnec]float64
	// taking the new evaluated SINR, evaluate the new metric
	for rb := range SNRres {
		for m := range MasterConn {
			a := Pr[m] / float64(NumARB[m]) / (SNRres[m][rb] + WNoise)
			SNRres[m][rb] = a
		}
	}

	// rememeber the number of allocated RB with the porposed allocation
	for m, ALsub := range AL {
		for rb, v := range ALsub {
			if v > -1 {
				metricM[v]+= EffectiveBW * math.Log2(1+SNRres[m][rb])
			}
		}
	}

	//fmt.Println(Pr)	
	//fmt.Println(metricM)			

	for v,C:=range MasterConn{
		m_m := C.E.GetMeanTR()	

		metric += math.Log2(1 + metricM[v])/math.Log2(1+m_m+0.0001)
	}

	return 
}

func AllocateSDMA(ALv *[mDf][NCh]int, MasterMobs []*Emitter) {

	for _, M := range MasterMobs {
		M.ClearFuturARB()
	}
	for s := 0; s < mDf; s++ {
		AL := ALv[s][:]
		for rb, vAL := range AL {
			if vAL >= 0 {
				if !MasterMobs[vAL].IsSetARB(rb) {
					Hopcount++
				}
				MasterMobs[vAL].SetARB(rb)
			}
		}
	}

}

func createDescSDMA(ppO SDMA_alloc, pool []SDMA_alloc, Rgen *rand.Rand) {

	nCh := len(ppO.vect)

	for s := 0; s < mDf; s++ {

		for i := 0; i < 10; i++ {

			//copy

			//copy(pool[i].vect[s], ppO.vect[s]) // copies value since arraytype
			pool[i].vect[s]= ppO.vect[s]
			
			pp := pool[i].vect[s][:] //short name to consider that array descendant
			//modify
			if pool[i].subSize[s]>0{			
			//take a random location
			v := -1
			rnd := 0

			for r := 0; v == -1 && r < 20; r++ {
				rnd = Rgen.Intn(nCh)
				v = pp[rnd]
			}

			if v >= 0 {

				a := Rgen.Float64()

				if a < 0.25 { //either increase size to the right
					for ; rnd < nCh && pp[rnd] == v; rnd++ {
					}
					if rnd < nCh { //copy prev or next
						a := Rgen.Float64()
						if a < .5 {
							pp[rnd] = pp[rnd-1]
						} else {
							pp[rnd-1] = pp[rnd]
						}
					}
				} else if a < 0.50 { // or to the left						

					for ; rnd > 0 && pp[rnd] == v; rnd-- {
					}
					if rnd > 0 {
						a := Rgen.Float64()
						if a < .5 {
							pp[rnd] = pp[rnd+1]
						} else {
							pp[rnd+1] = pp[rnd]
						}

					}
				} else if a < .75 { // or swap two variables
					v1 := v
					v2 := -1
					rnd1 := rnd
					rnd2 := 0
					for r := 0; v2 == -1 && r < 20; r++ {
						rnd2 = Rgen.Intn(nCh)
						v2 = pp[rnd]
					}
					if v2 >= 0 {
						if v1 != v2 { //swap values

							for ; rnd1 > 0 && pp[rnd1] == v1; rnd1-- {
							}
							for ; rnd2 > 0 && pp[rnd2] == v2; rnd2-- {
							}

							for ; rnd1 < nCh && pp[rnd1] == v1; rnd1++ {
								pp[rnd1] = v2
							}
							for ; rnd2 < nCh && pp[rnd2] == v2; rnd2++ {
								pp[rnd2] = v1
							}
						}
					}
				} else { // puncture
					for ; rnd < nCh && pp[rnd] == v; rnd++ {
						pp[rnd] = -1
					}

				}
			}
			}else{
				for k:=range pp{
					pp[k]=-1
				}
			}
		}

	}

	// additional modification : swap between 2 sdma branch


	b1 := Rgen.Intn(mDf)
	b2 := Rgen.Intn(mDf)

	for i := 0; i < 10; i++ {

		//copy		
	
		if pool[i].subSize[b1] >0 && pool[i].subSize[b2]>0 {

		pp1 := pool[i].vect[b1] //short name to consider that array descendant
		pp2 := pool[i].vect[b2] //short name to consider that array descendant
		//modify

		

		//take a random location
		v1 := -1
		rnd1 := 0

		for r := 0; v1 == -1 && r < 20; r++ {
			rnd1 = Rgen.Intn(nCh)
			v1 = pp1[rnd1]
		}

		v2 := -1
		rnd2 := 0

		for r := 0; v2 == -1 && r < 20; r++ {
			rnd2 = Rgen.Intn(nCh)
			v2 = pp2[rnd2]
		}

		if v1 >= 0 && v2 >= 0 {

			for ; rnd1 > 0 && pp1[rnd1] == v1; rnd1-- {
			}
			for ; rnd2 > 0 && pp2[rnd2] == v2; rnd2-- {
			}

			for ; rnd1 < nCh && pp1[rnd1] == v1; rnd1++ {
				pp1[rnd1] = v2
			}
			for ; rnd2 < nCh && pp2[rnd2] == v2; rnd2++ {
				pp2[rnd2] = v1
			}

		}
		}
	}

}

//   Reformatted by   lerouxp    Tue Dec 6 11:01:19 EST 2011

