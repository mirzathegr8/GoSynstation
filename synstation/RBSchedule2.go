package synstation

import "sort"
import "rand"
import "math"
//import "fmt"

const popsize = 100

func createDesc(ppO *[NCh]int, pool [][NCh]int, Rgen *rand.Rand) {
	//fmt.Print(" in")
	for i := 0; i < 10; i++ {

		//copy

		pool[i] = *ppO // copies value since arraytype
		//fmt.Println(pool[i])
		//fmt.Println(ppO)
		pp := pool[i][:] //short name to consider that array descendant
		//modify

		//take a random location
		v := -1
		rnd := 0

		for r := 0; v == -1 && r < 20; r++ {
			rnd = Rgen.Intn(NCh)
			v = pp[rnd]
		}

		if v >= 0 {

			//fmt.Println("rnd  ", rnd)
			a := Rgen.Float64()
			//	fmt.Println("a", a)

			if a < 0.333333 { //either increase size to the right
				for ; rnd < NCh && pp[rnd] == v; rnd++ {
				}
				if rnd < NCh { //copy prev or next
					a := Rgen.Float64()
					if a < .5 {
						pp[rnd] = pp[rnd-1]
					} else {
						pp[rnd-1] = pp[rnd]
					}
				}
			} else if a < 0.66666 { // or to the left						

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
			} else { // or swap two variables
				v1 := v
				v2 := -1
				rnd1 := rnd
				rnd2 := 0
				for r := 0; v2 == -1 && r < 20; r++ {
					rnd2 = Rgen.Intn(NCh)
					v2 = pp[rnd]
				}
				if v2 >= 0 {
					if v1 != v2 { //swap values

						for ; rnd1 > 0 && pp[rnd] == v1; rnd1-- {
						}
						for ; rnd2 > 0 && pp[rnd] == v2; rnd2-- {
						}

						for ; rnd1 < NCh && pp[rnd1] == v1; rnd1++ {
							pp[rnd1] = v2
						}
						for ; rnd2 < NCh && pp[rnd2] == v2; rnd1++ {
							pp[rnd2] = v1
						}
					}
				}
			}
		}
		//fmt.Println("Print pool")
		//fmt.Println(pool[i])

		//	fmt.Println(ppO)

	}
	//fmt.Print(" out")
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

			Nmaster++

			for rb := 1; rb < NCh; rb++ {

				var snrrb float64
				if DiversityType == SELECTION {
					snrrb = c.SNRrb[rb]
				} else {
					snrrb = E.GetSNRrb(rb)
				}

				m := EffectiveBW * math.Log2(1+snrrb)
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

	var Popul [popsize][NCh]int

	r := 6.0 / 6.0 // fraction of RB to allocate

	//fmt.Print("Nmaster ", Nmaster, " ", int(float64(NCh)*float64(r)/float64(Nmaster)))

	var pool [1100][NCh]int

	nbrb := int(float64(NCh) * float64(r) / float64(Nmaster))

	//First assign RB to best Metric
	for i := 0; i < 100; i++ {

		a := Rgen.Perm(Nmaster)

		//expand
		pp := &Popul[i]

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
			createDesc(&Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		//fmt.Println("desc created")

		copy(pool[1000:1100], Popul[:])

		//select the new population
		var metricpool [1100]float64

		for i := 0; i < 1100; i++ {
			metricT := float64(0.0)
			na := 0
			for j := 1; j < NCh; j++ {
				if pool[i][j] >= 0 {
					metricT += Metric[pool[i][j]][j]
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
			Popul[i] = pool[S.index[i]]
		}
		//fmt.Println(S.value[S.index[0]])
	}

	AL := Popul[0][:]

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//fmt.Println(AL)

	const CAPAthres = 1.5
	var max float64
	//		for i := 0; i < len(AL); i++ {
	//	
	//			/*	if AL[i] >= 0 {
	//					fmt.Print(Metric[AL[i]][i], " ")
	//				} else {
	//					fmt.Print("-1 ")
	//				}*/
	//	
	//			if AL[i] >= 0 && max < Metric[AL[i]][i] {
	//	
	//				max = Metric[AL[i]][i]
	//			}
	//		}
	//fmt.Println()
	AL[0] = -1
	//if AL[NCh-1] != -1 && Metric[AL[NCh-1]][NCh-1] < max/CAPAthres {
	//	AL[NCh-1] = -1
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
	//	fmt.Println(AL)

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

