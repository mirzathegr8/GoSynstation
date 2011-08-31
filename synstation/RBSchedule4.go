package synstation

import "sort"
import "rand"
import "math"
import "fmt"



const uARBcost = 0.01 //meanMeanCapa / 5 //0.5 // math.Log2(1 + meanMeanCapa)

func init() {

	fmt.Println("init to keep fmt")
}


func ARBScheduler4(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64

	var MasterMobs [NConnec]*Emitter
	var MasterConnec [NConnec]*Connection

	// Eval Metric for all connections
	var Nmaster int

	var meanMeanCapa float64
	var maxMeanCapa float64

	//NConnected := dbs.Connec.Len()


	//Get SNR for all master mobiles all rbs

	var minPower float64

	for j, i, e := 0, 0, dbs.Connec.Front(); e != nil; e, i = e.Next(), i+1 {

		c := e.Value.(*Connection)
		E := c.GetE()
		m_m := E.GetMeanTR()
		meanMeanCapa += m_m

		if maxMeanCapa < m_m {
			maxMeanCapa = m_m
		}

		if minPower>  math.Log10( c.meanPr.Get()  ){
			minPower=c.meanPr.Get()
		}

		if c.Status == 0 {
			
			MasterConnec[Nmaster] = c
			MasterMobs[Nmaster] = E

			for i:=0; i<NCh; i++{E.UnSetARB(i)}

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

	

	meanMeanCapa /= float64(dbs.Connec.Len())
	meanMeanCapa += 0.1

	if Nmaster == 0 { // in this case nothing to assign
		return
	}

	//Assign RB for master connections

	//var NumAss [NConnec]int

	var Popul [popsize][NCh]int	
	var pool [popsize * 11][NCh]int

	//First assign RB to best Metric
	for i := 0; i < popsize; i++ {
		nbrb := int(float64(NCh) / (float64(Nmaster) ) ) //int(float64(NCh) * float64(r) / float64(numberAmobs))
		a := Rgen.Perm(Nmaster)
		//expand
		pp := &Popul[i]
		// first dealocate everything
		for j := 0; j < NCh; j++ {
			pp[j] = -1
		}
		//second alocate one chanel
		for j := 0; j < Nmaster; j++ {
			index := int(float64(NCh)/float64(Nmaster)*float64(j)) 			
			for k := 0; k < nbrb; k++ {
				pp[index+k] = a[j]
			}
		}	
	}

	

	for gen := 0; gen < generations; gen++ {
		
		for j := 0; j < popsize; j++ {
			createDesc(&Popul[j], pool[j*10:(j+1)*10], Rgen)
		}

		copy(pool[popsize*10:popsize*11], Popul[:])

		//select the new population
		var metricpool [popsize * 11]float64
	
		for i := 0; i < popsize*11; i++ {					
			var AL [NCh]int
			AL = pool[i] //copies 
			AL[0] = -1
			metricpool[i] = Trimm(AL[:],&Metric, MasterMobs[0:Nmaster])
		}
		
		//find the best 100
		S := initSequence(metricpool[:])
		sort.Sort(S)		
		for i := 0; i < len(Popul); i++ {
			Popul[i] = pool[S.index[i]]	
		}
		
	}

	AL := Popul[0][:]

	

	//Trimm we delete endings of allocation sequence if the capacity of these RB is not that good
	//this is to prevent spendin too much energy for little gain, and also to give a chance to minimize interference

	//Trimm(AL[:],&Metric,MasterMobs[0:Nmaster])	

	//testSCFDMA(AL)

	//copy(dbs.ALsave[:],AL)

	Allocate(AL[:],MasterMobs[0:Nmaster])
	//AllocateOld(AL[:],dbs)
	
}


func Trimm(AL []int, Metric *[NConnec][NCh]float64, MasterMobs []*Emitter) (metricT float64) {

	AL[0] = -1	

	for i:=1 ; i<len(AL);  {

		v:=AL[i]
		switch v {
		case -1:		
			metricT += uARBcost	
			i++
			 // just go on
		default:	
			Mvect:=Metric[v][0:NCh]
	
			//find max
			jmin := i			
			var max float64
			nARBm := 0

			for j,vAL := range AL[jmin:len(AL)] {
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

			nARBmOrg:=nARBm
			jmax := jmin + nARBm //save first
		
		
		/*	for _,Mval :=range Mvect[jmin:jmax]{ //j:=jmin; j<jmax ;j++ { //

				//Mval:=Mvect[j]
				m := EffectiveBW * math.Log2(1+Mval/float64(nARBm))
				
			 	if  Mval < max/CAPAthres || m < 100 {
					AL[jmin] = -1
					nARBm--
					metricT += uARBcost
					jmin++
				} else {
					break
				}			
			}

			for j := jmax-1; j >= jmin; j-- {

				Mval:=Mvect[j]
				m := EffectiveBW * math.Log2(1+Mval/float64(nARBm))

				if Mval < max/CAPAthres || m < 100 {
					AL[j] = -1
					nARBm--
					metricT += uARBcost
					jmax--				
				} else {
					break
				}
			}*/


			m:=float64(0.0)
			for _,Mval := range Mvect[jmin:jmax]{
				m += EffectiveBW * math.Log2(1+beta*Mval/float64(nARBm))
			}

			m_m := MasterMobs[v].GetMeanTR()				
			metricT += math.Log2(1 + m/(m_m+0.0001)) 

			i += nARBmOrg
		}
	}

	return
}


func Allocate(AL []int,  MasterMobs []*Emitter){


	AL[0] = -1 // connect all
	for _,M :=range MasterMobs{		
		M.ClearFuturARB()
	}
	for rb, vAL := range AL {		
		if vAL>=0{	
			if !MasterMobs[vAL].IsSetARB(rb) {Hopcount++}	
			MasterMobs[vAL].SetARB(rb)
		}
	}
	
}


func testSCFDMA(AL []int) int{

	sv:=-1
		
	for i:=0 ; i<NCh;i++{
		if sv==-1{
			sv=AL[i]
		}		
		if sv!=AL[i]{
			for j:=i;j<NCh;j++{
				if AL[j]==sv {	return -1}			
			}
			sv=AL[i]
		}
	}

	return 0


}
