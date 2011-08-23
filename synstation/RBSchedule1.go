
package synstation


import "rand"
import "math"
//import "fmt"



func ARBScheduler(dbs *DBS, Rgen *rand.Rand) {

	var Metric [NConnec][NCh]float64
	var MasterMobs [NConnec]EmitterInt
	Nmaster :=0

	for _, e := 0, dbs.Connec.Front(); e != nil; e = e.Next() {	
	
		c := e.Value.(*Connection)
		if c.Status==0 {
		
			E := c.GetE()
			m_m := E.GetMeanTR()	

			for rb := 0; rb < NCh; rb++ {

				E.UnSetARB(rb)	

				snrrb := E.GetSNRrb(rb)
	
				m := EffectiveBW * math.Log2(1+beta*snrrb)

				if m > 100 {
					Metric[Nmaster][rb] = math.Log2(m + 1)
					b := (m_m + 1)			
					Metric[Nmaster][rb] /= b				
				}					
			}
			MasterMobs[Nmaster]=E			
			Nmaster++
		}
	}

	if Nmaster==0 {return}


	//Assign RB for master connections

	var AL [NCh]int
	AL[0] = -1 // connect all
	var NumAss [NConnec]int
	//First assign RB to best Metric
	for rb := 1; rb < NCh; rb++ {
		AL[rb] = -1
		for i :=range MasterMobs[0:Nmaster] {			
			if Metric[i][rb] > 0 {
				if AL[rb] < 0 {
					AL[rb] = i
				} else if Metric[i][rb] > Metric[AL[rb]][rb] {
					AL[rb] = i
				}
			}
		}
		if AL[rb] >= 0 {
			NumAss[AL[rb]]++ //this emitter will have one more assigned RB
			for rb2 := 1; rb2 < NCh; rb2++ {
				Metric[AL[rb]][rb2] *= float64(NumAss[AL[rb]]) / float64(NumAss[AL[rb]]+1)
			}
		}
	}	
	
	

	Allocate(AL[:],MasterMobs[0:Nmaster])
	
	//AllocateOld(AL[:],dbs)

	/*for _,E := range MasterMobs[0:Nmaster]{
		E.UnSetARB(0)
	}
	for rb, vAL := range AL {		
		if (vAL>=0) {	
			for i,E := range MasterMobs[0:Nmaster]{
				if i!= vAL{
					E.UnSetARB(rb)
				}else{
					if !MasterMobs[vAL].IsSetARB(rb) {
						Hopcount++
						MasterMobs[vAL].SetARB(rb)
					}

				}
			}
		}	
	}*/

}

