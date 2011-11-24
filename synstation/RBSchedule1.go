
package synstation


import rand "math/rand"
import "math"
//import "fmt"
//import "geom"


func ARBScheduler(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64
	var MasterMobs [NConnec]*Emitter
	Nmaster :=0

	r:= dbs.Pos


	for _, e := 0, dbs.Connec.Front(); e != nil; e = e.Next() {	
	
		c := e.Value.(*Connection)
		if c.Status==0 {
		
			E := c.GetE()
			m_m := E.GetMeanTR()	

			for rb := 0; rb < NCh; rb++ {	

				snrrb := E.GetSNRrb(rb)
	
				m := EffectiveBW * math.Log2(1+beta*snrrb)

				if m > 100 {
					Metric[Nmaster][rb] = math.Log2(m + 1)
					b := (m_m + 1)			
					Metric[Nmaster][rb] /= b				
				}					
			}
			MasterMobs[Nmaster]=E	

			ICIMfunc(&r,E,Metric[Nmaster][:],dbs.Color)
		
			Nmaster++

		}
	}

	if Nmaster==0 {return}


	//Assign RB for master connections

	

	var CAL [NCh]int
	MaxMP:=0.0
	for g :=0 ;g<1;g++{

		MetricBis:=Metric

		var AL [NCh]int
		var NumAss [NConnec]int
		//First assign RB to best Metric	
		MetricPool:=0.0
		RBrange := Rgen.Perm(NCh-NChRes) //add NChRes
		AL[0]=-1

		//for rb := 1; rb < NCh; rb++ {
		for _,rbo:= range RBrange{
			rb:=rbo+NChRes
			AL[rb] = -1
			for i :=range MasterMobs[0:Nmaster] {			
				if MetricBis[i][rb] > 0 {
					if AL[rb] < 0 {
						AL[rb] = i
					} else if MetricBis[i][rb] > MetricBis[AL[rb]][rb] {
						AL[rb] = i
					}
				}
			}
			if AL[rb] >= 0 {
				MetricPool+=MetricBis[AL[rb]][rb]
				NumAss[AL[rb]]++ //this emitter will have one more assigned RB
				for rb2 := 1; rb2 < NCh; rb2++ {
					MetricBis[AL[rb]][rb2] *= float64(NumAss[AL[rb]]) / float64(NumAss[AL[rb]]+1)
				}
			}
		}
		if MetricPool>MaxMP {	
			MaxMP=MetricPool
			CAL=AL
		}
	}
	

	Allocate(CAL[:],MasterMobs[0:Nmaster])
	//AllocateOld(AL[:], dbs )
	
	

}


