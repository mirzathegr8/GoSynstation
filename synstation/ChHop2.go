package synstation

import "container/vector"
import "math"
import "rand"
//import "fmt"







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


func ChHopping2(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	var MobileList vector.Vector

	//pour trier les canaux
	dbs.RandomChan2()

	var stop = 0

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := 0; j < NCh-NChRes; j+=subsetSize {
					i := dbs.RndCh[j]
					if !dbs.IsInUse(i) && !c.E.IsSetARB(i) {

						/*snr:=0.0
						for l:=0; l<subsetSize ; l++{
						_, a,_,_ :=dbs.R.EvalSignalSNR(c.E, i+l)
						snr+=a
						}
						snr/=float64(subsetSize)
					*/
				snr:=0.0;
				for l:=0; l<subsetSize ; l++{
					 snr += c.GetE().GetSNRrb(i+l) //dbs.R.EvalSignalSNR(co.GetE(), i+l)
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
					dbs.changeChannel(c, 0, nch)
					for l:=1 ; l<subsetSize;l++{	
						c.GetE().SetARB(nch+l)
					}
					stop++
					if stop > 1 {
						return
					}
				}

				// sort mobile connection for channel hopping
			} else {
				ratio := c.EvalRatio(dbs.R)
				var i int
				for i = 0; i < MobileList.Len(); i++ {
					co := MobileList.At(i).(ConnecType)
					if ratio < co.EvalRatio(dbs.R) {
						break
					}
				}
				MobileList.Insert(i, c)
			}
		}
	}

	// change channel to some mobiles
	for k := 0; k < MobileList.Len() && k < 2; k++ {
		co := MobileList.At(k).(ConnecType)
		ratio := co.EvalRatio(dbs.R)
		chHop := 0

		for j := 0; j < NCh-NChRes; j+=subsetSize {

			i := dbs.RndCh[j]

			if !dbs.IsInUse(i) && !co.GetE().IsSetARB(i) {

				snr:=0.0;
				for l:=0; l<subsetSize ; l++{
					 snr += co.GetE().GetSNRrb(i+l) //dbs.R.EvalSignalSNR(co.GetE(), i+l)
				}
				snr/=float64(subsetSize)
//				_, snr, _, _ := dbs.R.EvalSignalSNR(co.GetE(), i)

				if snr > ratio {
					ratio = snr
					chHop = i
				}
			}
		}
		if chHop > 0 {
			oldCh:=co.GetE().GetFirstRB()
			for l:=0; l<subsetSize ; l++{
			dbs.changeChannel(co, oldCh+l, chHop+l)
			}
			Hopcount++
		}

	}

}
