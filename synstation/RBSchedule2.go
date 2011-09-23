package synstation

import "sort"
import "rand"
import "math"
import "fmt"

func init() {

	fmt.Println("init to keep fmt")
}


type allocSet struct {
	vect []int
} 

func createDesc(ppO allocSet, pool []allocSet, Rgen *rand.Rand) {
	//fmt.Print(" in")

	nCh:= len(ppO.vect)

	for i := 0; i < 10; i++ {

		//copy

		copy(pool[i].vect,ppO.vect) // copies value since arraytype

		pp := pool[i].vect //short name to consider that array descendant
		//modify

		//take a random location
		v := -1
		rnd := 0

		for r := 0; v == -1 && r < 20; r++ {
			rnd = Rgen.Intn(nCh)
			v = pp[rnd]
		}

		if v >= 0 {

			//fmt.Println("rnd  ", rnd)
			a := Rgen.Float64()
			//	fmt.Println("a", a)

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

						for ; rnd1 > 0 && pp[rnd] == v1; rnd1-- {
						}
						for ; rnd2 > 0 && pp[rnd] == v2; rnd2-- {
						}

						for ; rnd1 < nCh && pp[rnd1] == v1; rnd1++ {
							pp[rnd1] = v2
						}
						for ; rnd2 < nCh && pp[rnd2] == v2; rnd1++ {
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

	}

}


func ARBScheduler2(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64

	//var MobilesID [NConnec]int

	// Eval Metric for all connections
	var Nmaster int

	for j, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.GetE()

		if c.Status == 0 {
			for i:=0; i<NCh; i++{E.UnSetARB(i)}

			Nmaster++

			for rb := 1; rb < NCh; rb++ {

				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.GetSNRrb(rb)
				}

				m := EffectiveBW * math.Log2(1+beta*snrrb)
				m_m := c.GetE().GetMeanTR()

				if m > 100 && m_m < 100000026000 {
					Metric[j][rb] = math.Log2(m + 1)
					b := (m_m + 1)
					if b > 1 {
						Metric[j][rb] /= b
					}

				} else {
					Metric[j][rb] = 0
				}

			}

			j++

		}

	}

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	//Assign RB for master connections

	//var NumAss [NConnec]int

	var PopulAr [popsize][NCh]int
	var Popul [popsize]allocSet
	for i:=range Popul{ Popul[i].vect=PopulAr[i][:]}

	r := 6.0 / 6.0 // fraction of RB to allocate

	//fmt.Print("Nmaster ", Nmaster, " ", int(float64(NCh)*float64(r)/float64(Nmaster)))

	var poolAr [(popsize+1)*generations][NCh]int
	var pool [(popsize+1)*10]allocSet
	for i:=range pool{ pool[i].vect=poolAr[i][:]}
	

	nbrb := int(float64(NCh) * float64(r) / float64(Nmaster))

	//First assign RB to best Metric
	for i := 0; i < generations; i++ {

		a := Rgen.Perm(Nmaster)

		//expand
		pp := Popul[i].vect

		// first dealocate everything
		for j := 0; j < NCh; j++ {
			pp[j] = -1
		}
		//second alocate one chanel
		for j := 0; j < Nmaster; j++ {
			index := int(float64(NCh)/float64(Nmaster)*float64(j)) + Rgen.Intn(int(float64(NCh)*(1-r)/float64(Nmaster)))
			pp[index] = a[j]

			for k := 0; k < nbrb; k++ {
				pp[index+k] = a[j]
			}
		}

		//fmt.Println(pp)
	}

	for gen := 0; gen < 50; gen++ {

		//fmt.Println(gen)

		for j := 0; j < 100; j++ {
			createDesc(Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		//fmt.Println("desc created")


		for i:=range Popul{
			copy(pool[popsize*10+i].vect, Popul[i].vect)	
		}
		//copy(pool[1000:1100], Popul[:])

		//select the new population
		var metricpool [1100]float64

		for i := 0; i < 1100; i++ {
			metricT := float64(0.0)
			na := 0
			for j := 1; j < NCh; j++ {
				if pool[i].vect[j] >= 0 {
					metricT += Metric[pool[i].vect[j]][j]
				} else {
					na++
				}
			}
			if na < int(float64(NCh)*(1-r)*0.85) {
				metricT = 0
			} //eliminate entries with too many ARB
			metricpool[i] = metricT
		}
		//find the best 100

		S := initSequence(metricpool[:])

		sort.Sort(S)

		for i := range Popul {
			//fmt.Print(" ", S.value[S.index[i]])
			copy(Popul[i].vect, pool[S.index[i]].vect)
		}
		//fmt.Println(S.value[S.index[0]])
	}

	AL := Popul[0].vect

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//fmt.Println(AL)

	const CAPAthres = 1.5
	var max float64
	// for i := 0; i < len(AL); i++ {
	//
	// /* if AL[i] >= 0 {
	// fmt.Print(Metric[AL[i]][i], " ")
	// } else {
	// fmt.Print("-1 ")
	// }*/
	//
	// if AL[i] >= 0 && max < Metric[AL[i]][i] {
	//
	// max = Metric[AL[i]][i]
	// }
	// }
	//fmt.Println()
	AL[0] = -1
	//if AL[NCh-1] != -1 && Metric[AL[NCh-1]][NCh-1] < max/CAPAthres {
	// AL[NCh-1] = -1
	//}
	for i := 1; i < len(AL); i++ {

		switch AL[i] {
		case -1: // just go on

		default:

			//find max

			v := AL[i]
			max = 0
			for j := i; j < len(AL) && AL[j] == v; j++ {
				if Metric[AL[j]][j] > max {
					max = Metric[AL[j]][j]
				}
			}

			//loop
			j := i
			for j = i; j < len(AL) && AL[j] == v; j++ {
				if AL[j-1] != v && Metric[AL[j]][j] < max/CAPAthres || Metric[AL[j]][j] < 100 {
					AL[j] = -1
				} else if (j == NCh-1 || AL[j+1] != v) && Metric[AL[j]][j] < max/CAPAthres || Metric[AL[j]][j] < 100 {
					AL[j] = -1
				}
			}
			i = j

		}

	}
	// fmt.Println(AL)

	//Allocate RB effectivelly

	AL[0] = -1 // connect all
	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
		c := e.Value.(*Connection)
		E := c.GetE()
		if c.Status == 0 {
			if E.IsSetARB(0) {
				E.UnSetARB(0)
			}

		}
	}
	for rb := 1; rb < NCh; rb++ {
		if AL[rb] >= 0 {
			for k, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
				c := e.Value.(*Connection)
				E := c.GetE()

				if c.Status == 0 {

					if E.IsSetARB(rb) {
						if AL[rb] != k {
							E.UnSetARB(rb)
						}
					} else {
						if AL[rb] >= 0 {

							if AL[rb] == k {
								E.SetARB(rb)
								Hopcount++

							}
						}
					}
					k++
				}
			}
		}
	}

	//fmt.Println("done")

}


func ARBScheduler3(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64

	var MasterMobs [NConnec]*Emitter
	var MasterConnec [NConnec]*Connection

	//var MobilesID [NConnec]int

	// Eval Metric for all connections
	var Nmaster int

	var meanMeanCapa float64
	var maxMeanCapa float64

	//NConnected := dbs.Connec.Len()

	for j, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.GetE()

		m_m := E.GetMeanTR()

		meanMeanCapa += m_m
		if maxMeanCapa < m_m {
			maxMeanCapa = m_m
		}

		if c.Status == 0 {

			MasterConnec[Nmaster] = c
			MasterMobs[Nmaster] = E

			for rb := 1; rb < NCh; rb++ {

				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.GetSNRrb(rb)
				}

				Metric[j][rb] = snrrb

			}
			Nmaster++
			j++

		}

	}

	//fmt.Println(Metric[0:Nmaster])

	meanMeanCapa /= float64(dbs.Connec.Len())
	meanMeanCapa += 0.1

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	//Assign RB for master connections

	//var NumAss [NConnec]int

//	var Popul [popsize][NCh]int
	var PopulAr [popsize][NCh]int
	var Popul [popsize]allocSet
	for i:=range Popul{ Popul[i].vect=PopulAr[i][:]}

	r := 6.0 / 6.0 // fraction of RB to allocate

	//fmt.Print("Nmaster ", Nmaster, " ", int(float64(NCh)*float64(r)/float64(Nmaster)))

//	var pool [popsize * 11][NCh]int
	var poolAr [(popsize+1)*10][NCh]int
	var pool [(popsize+1)*10]allocSet
	for i:=range pool{ pool[i].vect=poolAr[i][:]}

	//First assign RB to best Metric
	for i := 0; i < popsize; i++ {

		numberAmobs := Nmaster                                               //Rgen.Intn(Nmaster) + 1
		nbrb := int(float64(NCh) * float64(r) / (float64(Nmaster) )) //int(float64(NCh) * float64(r) / float64(numberAmobs))

		a := Rgen.Perm(Nmaster)

		//expand
		pp := Popul[i].vect

		// first dealocate everything
		for j := 0; j < NCh; j++ {
			pp[j] = -1
		}
		//second alocate one chanel
		for j := 0; j < numberAmobs; j++ {
			index := int(float64(NCh)/float64(numberAmobs)*float64(j)) + Rgen.Intn(int(float64(NCh)*(1-r)/float64(numberAmobs)))
			pp[index] = a[j]

			for k := 0; k < nbrb; k++ {
				pp[index+k] = a[j]
			}
		}

		//fmt.Println(pp)
	}

	for gen := 0; gen < generations; gen++ {

		//fmt.Println(gen)

		for j := 0; j < popsize; j++ {
			createDesc(Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		//fmt.Println("desc created")

		for i:=range Popul{
			copy(pool[popsize*10+i].vect, Popul[i].vect)	
		}


		//select the new population
		var metricpool [popsize * 11]float64

		uARBcost := 0.00 //meanMeanCapa / 5 //0.5 // math.Log2(1 + meanMeanCapa)

		for i := 0; i < popsize*11; i++ {
			metricT := float64(0.0)
			//na := 0

			//eval what would be the AL after trimming
			var max float64
			var AL [NCh]int
			copy(AL[:],pool[i].vect) //copies 
			AL[0] = -1

			for i := 1; i < len(AL); i++ {

				switch AL[i] {
				case -1: // just go on

					metricT += uARBcost

				default:

					//find max
					jmin := i
					v := AL[i]
					max = 0
					nARBm := 0
					j := 0
					for j = i; j < len(AL) && AL[j] == v; j++ {
						nARBm++
						if Metric[AL[j]][j] > max {
							max = Metric[AL[j]][j]
						}
					}
					//loop to trim
					// here we only consider the Original metric 

					jmax := i +nARBm - 1 //save first
					jmaxo := jmax
					for j = i; j <= jmax; j++ {

						/*var snrrb float64
						if DiversityType == SELECTION {
							snrrb = MasterConnec[v].SNRrb[j]
						} else {
							snrrb = MasterMobs[v].GetSNRrb(j)
						}*/

						m := EffectiveBW * math.Log2(1+beta*Metric[v][j]/float64(nARBm))

						if AL[j-1] != v && (Metric[v][j] < max/CAPAthres || m < 100) {
							AL[j] = -1
							nARBm--
							jmin++

							metricT += uARBcost
						} else {
							break
						}
					}

					for j = jmax; j >= jmin; j-- {

						/*var snrrb float64
						if DiversityType == SELECTION {
							snrrb = MasterConnec[v].SNRrb[j]
						} else {
							snrrb = MasterMobs[v].GetSNRrb(j)
						}*/

						m := EffectiveBW * math.Log2(1+beta*Metric[v][j]/float64(nARBm))

						if (j == NCh-1 || AL[j+1] != v) && (Metric[v][j] < max/CAPAthres || m < 100) {
							AL[j] = -1
							nARBm--
							metricT += uARBcost
							jmax--
						} else {
							break
						}
					}

					//this range should only have allocated RB to a unique mobile

					var m float64
					for rb := jmin; rb <= jmax; rb++ {

						m += EffectiveBW * math.Log2(1+beta*Metric[v][rb]/float64(nARBm))

					}

					m_m := MasterConnec[v].GetE().GetMeanTR()

					/*a := (m - m_m*0.8)
					if a > 0 {
						metricT += a * (meanMeanCapa*1.2 - m_m)
					}*/

					//metricT += math.Log2(1+m) * math.Exp(m_m-meanMeanCapa) //* (1 + (meanMeanCapa-MasterMobs[v].GetMeanTR())/meanMeanCapa)

					metricT += math.Log2(1 + m/(m_m+0.0001)) // fair metric, low latency// math.Log2(1+math.Exp((meanMeanCapa-m_m)/maxMeanCapa))

					i = jmaxo
				}

			}

			metricpool[i] = metricT
		}
		//find the best 100


		S := initSequence(metricpool[:])

		sort.Sort(S)

		shift := 0 //len(metricpool)/4 - len(Popul)/2
		if shift < 0 {
			shift = 0
		}

		for i := 0; i < len(Popul); i++ {
			//fmt.Print(" ", S.value[S.index[i]])
			Popul[i] = pool[S.index[i+shift]]
			//fmt.Println(Popul[i])
		}
		//fmt.Println(S.value[S.index[0]])
	}

	AL := Popul[0].vect

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//fmt.Println("AL       ", AL)

	var max float64

	AL[0] = -1

	for i := 1; i < len(AL); i++ {

		switch AL[i] {
		case -1: // just go on

		default:

			//find max
			jmin := i
			v := AL[i]
			max = 0
			nARBm := 0
			j := 0
			for j = i; j < len(AL) && AL[j] == v; j++ {
				nARBm++
				if Metric[AL[j]][j] > max {
					max = Metric[AL[j]][j]
				}
			}
			//loop to trim
			// here we only consider the Original metric 

			jmax := j - 1 //save first
			jmaxo := j - 1
			for j = i; j <= jmax; j++ {

				m := EffectiveBW * math.Log2(1+Metric[AL[j]][j]/float64(nARBm))

				if AL[j-1] != v && (Metric[AL[j]][j] < max/CAPAthres || m < 100) {
					AL[j] = -1
					nARBm--
					jmin++
				} else {
					break
				}
			}

			for j = jmax; j >= jmin; j-- {

				m := EffectiveBW * math.Log2(1+Metric[AL[j]][j]/float64(nARBm))

				if (j == NCh-1 || AL[j+1] != v) && (Metric[AL[j]][j] < max/CAPAthres || m < 100) {
					AL[j] = -1
					nARBm--
					jmax--
				} else {
					break
				}
			}
			i = jmaxo

		}

	}
	/*fmt.Println(AL)
	for h := 0; h < NCh; h++ {
		if AL[h] >= 0 {
			fmt.Printf("%2.1f ", math.Log10(Metric[AL[h]][h]))
		} else {
			fmt.Print("-1 ")
		}
	}
	fmt.Println()*/		
 
	
/*
	for i,v:= range AL[0:len(AL)-1]{
		
	if  v!=sv && sv!=-1{ 
		
		for _,vv:=range AL[i:len(AL)]{
			if sv==vv {fmt.Println("NOT SC-FDMA " ,AL)}	
			goto br	
		}
		sv=v	
	}else if sv==-1 && v!= -1{
		sv=v	
	}


	}
	br:*/


	//Allocate RB effectivelly	

	/*AL[0] = -1 // connect all
	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
		c := e.Value.(*Connection)
		E := c.GetE()
		if c.Status == 0 {
			if E.IsSetARB(0) {
				E.UnSetARB(0)
			}

		}
	}
	for rb := 1; rb < NCh; rb++ {
		if AL[rb] >= 0 {
			for k, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
				c := e.Value.(*Connection)
				E := c.GetE()

				if c.Status == 0 {

					if E.IsSetARB(rb) {
						if AL[rb] != k {
							E.UnSetARB(rb)
						}
					} else {
						if AL[rb] >= 0 {

							if AL[rb] == k {
								E.SetARB(rb)
								Hopcount++

							}
						}
					}
					k++
				}
			}
		}
	}*/

	//fmt.Println("done")
 	AL[0] = -1 // connect all	
	for rb, vAL := range AL {		
		for k, E := range MasterMobs[0:Nmaster] {
			if E.IsSetARB(rb) {
				if vAL != k {
					E.UnSetARB(rb)
				}
			} else {
				if vAL == k {
					E.SetARB(rb)
					Hopcount++
				}
			}										
		}
	}


	//test for SCFDMA on past assignment
	/*for i:=range AL {AL[i]=-1}
	for i,E:=range MasterMobs[0:Nmaster]{
		for rb,v:= range E.GetARB(){
			if v {AL[rb]=i}
		}
	}
	
	testSCFDMA(AL[:])*/


}


type Sequence struct {
	index []int
	value []float64
}

func initSequence(value []float64) (s Sequence) {
	s.index = make([]int, len(value))
	for i := range value {
		s.index[i] = i
	}
	s.value = value
	return
}

// Methods required by sort.Interface.
func (s Sequence) Len() int {
	return len(s.value)
}
func (s Sequence) Less(i, j int) bool {
	return s.value[s.index[i]] > s.value[s.index[j]] // sort bigest to smalest
}
func (s Sequence) Swap(i, j int) {
	s.index[i], s.index[j] = s.index[j], s.index[i]
}



func AllocateOld(AL []int, dbs *DBS) {

	AL[0] = -1 // connect all
	for i, e := 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {
		c := e.Value.(*Connection)
		E := c.GetE()
		if c.Status == 0 {
			if E.IsSetARB(0) {
				E.UnSetARB(0)
			}

		}
	}
	for rb := 1; rb < NCh; rb++ {
		if AL[rb] >= 0 {
			for k, e := 0, dbs.Connec.Front(); e != nil; e = e.Next() {
				c := e.Value.(*Connection)
				E := c.GetE()

				if c.Status == 0 {

					if E.IsSetARB(rb) {
						if AL[rb] != k {
							E.UnSetARB(rb)
						}
					} else {
						if AL[rb] >= 0 {

							if AL[rb] == k {
								E.SetARB(rb)
								Hopcount++

							}
						}
					}
					k++
				}
			}
		}
	}

}
