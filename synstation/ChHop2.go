package synstation

import "container/vector"
import "math"
import "rand"
import "fmt"


const ChRX= 0

func init(){
fmt.Println("init")
}



// Sorts channels in random order
func (dbs *DBS) RandomChan2() {

	var SortCh vector.IntVector

	for i := NChRes; i < NCh; i+=subsetSize {
		SortCh.Push(i)
		dbs.RndCh[i] = i
	}

	//randomizesd reserved top canals
	/*	for i := 10; i > 1; i-- {
			j := dbs.Rgen.Intn(i) + NCh-10
			SortCh.Swap(NCh-10+i-1, j)
			dbs.RndCh[NCh-10+i-1] =SortCh.Pop()
		}
		dbs.RndCh[NCh-10]=SortCh.Pop()
	*/
	//randomizes other canals
	/*	for i := NCh - 11; i > NChRes; i-- {
			j := dbs.Rgen.Intn(i-NChRes) + NChRes 
			SortCh.Swap(i, j)
			dbs.RndCh[i] =SortCh.Pop()
		}
		dbs.RndCh[NChRes] = SortCh.Pop()
		dbs.RndCh[0] = 0
	*/
	//fmt.Println(dbs.RndCh);

	for i := SortCh.Len()-1; i > 0; i-- {
		j := dbs.Rgen.Intn(i) 
		SortCh.Swap(i, j)
		dbs.RndCh[i] = SortCh.Pop()
	}
	dbs.RndCh[0] = SortCh.Pop()

}

//var plus int

func ChHopping2(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	var MobileList vector.Vector
	var MobileListRX vector.Vector

	//pour trier les canaux
//	dbs.RandomChan2()

	var stop = 0

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			//first copy the future vector

			//c.E.CopyFuturARB()
	
			//if c.E.GetFirstRB()<NChRes {
			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh-subsetSize+1; j+=subsetSize {
					i := j// dbs.RndCh[j]
					if dbs.IsInUse(i)==nil && !dbs.IsInFuturUse(i) {
				
						snr:=0.0;
						for l:=0; l<subsetSize ; l++{
							 snr += c.GetE().GetSNRrb(i+l) 
						}
						snr/=float64(subsetSize)
						if 10*math.Log10(snr) > SNRThresChHop {
	
						if snr > ratio {
							ratio = snr
							nch = i
							//assign and exit
						}
					}
						}
				}
				if nch != 0 {
					c.E.UnSetARB(0)					
					for l:=0 ; l<subsetSize;l++{	
						c.E.SetARB(nch+l)
						Hopcount++
					}
					stop++
					if stop > 1 {
						return
					}
				}
	
					// sort mobile connection for channel hopping
			} else {
				//ratio := c.EvalRatio(dbs.R)
				ratio:= EvalRatio(c.GetE())
				i:=0
				if c.GetE().GetFirstRB()< NChRes+subsetSize*ChRX{
					for i = 0; i < MobileListRX.Len(); i++ {				
						co := MobileListRX.At(i).(ConnecType)
						if ratio < co.EvalRatio(dbs.R) {
							break
						}
					}
					MobileListRX.Insert(i, c)
				}else{
					for i = 0; i < MobileList.Len(); i++ {				
						co := MobileList.At(i).(ConnecType)
						if ratio < EvalRatio(co.GetE()) {
							break
						}
					}	
					MobileList.Insert(i, c)
				}
			}
		}
	}

	// change channel to some mobiles

	fact:=0.8
	var MobileListUSE *vector.Vector
	if MobileListRX.Len()>0{
		MobileListUSE= &MobileListRX
		fact=0
	}else{
		MobileListUSE= &MobileList
	}


	for k := 0; k < MobileListUSE.Len() && k < 2; k++ {
		co := MobileListUSE.At(k).(ConnecType)
		E:=co.GetE();
		ratio:=EvalRatio(E) *fact

		chHop := 0

		for j := NChRes+subsetSize*ChRX; j < NCh-subsetSize+1; j+=subsetSize {

			i := j//dbs.RndCh[j]

			if dbs.IsInUse(i)==nil && !dbs.IsInFuturUse(i) && !E.IsSetARB(i) {

				snr:=0.0;
				for l:=0; l<subsetSize ; l++{
					 snr += E.GetSNRrb(i+l)
				}
				snr/=float64(subsetSize)

				if snr > ratio {
					ratio = snr
					chHop = i
				}
			}
		}
		oldCh:=E.GetFirstRB()
		if chHop > 0 {			
			for l:=0; l<subsetSize ; l++{
				E.UnSetARB(oldCh+l)
				E.SetARB(chHop+l)
				Hopcount++
			}					
		}
	}
}

func EvalRatio(E EmitterInt) float64{

		ratio := 0.0
		oldCh:=E.GetFirstRB()		
		for l:=0; l<subsetSize ; l++{
		 	ratio += E.GetSNRrb(oldCh+l)
		}
		ratio/=float64(subsetSize)
	return ratio
}
