package synstation

//import "container/vector"
import "math"
import rand "rand"
import "fmt"
//import "geom"

const ChRX = 0

func init() {
	fmt.Println("init")
}

// Sorts channels in random order
func (dbs *DBS) RandomChan() {

	SortCh := make([]int, NCh)
	SortCh = SortCh[0:0]

	for i := NChRes; i < NCh; i += subsetSize {
		SortCh = append(SortCh, i)
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

	for i := len(SortCh) - 1; i > 0; i-- {
		j := dbs.Rgen.Intn(i)
		SortCh[i], SortCh[j] = SortCh[j], SortCh[i]
		dbs.RndCh[i], SortCh = SortCh[len(SortCh)-1], SortCh[:len(SortCh)-1]
	}
	dbs.RndCh[0], SortCh = SortCh[len(SortCh)-1], SortCh[:len(SortCh)-1]

}

//var plus int

func ChHopping2(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	//var MobileList vector.Vector
	MobileList := make([]*Connection, NConnec)
	MobileList = MobileList[0:0]
	//var MobileListRX vector.Vector
	MobileListRX := make([]*Connection, NConnec)
	MobileListRX = MobileListRX[0:0]

	//pour trier les canaux
	//	dbs.RandomChan()

	var stop = 0

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			//first copy the future vector

			//c.E.CopyFuturARB()

			//if c.E.GetFirstRB()<NChRes {
			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel				

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 

				nch := FindFreeChan(dbs, c.E, math.Pow10(SNRThresChHop/10.0))

				if nch != 0 {				
					c.E.UnSetARB(0)
					for l := 0; l < subsetSize; l++ {
						c.E.SetARB(nch + l)
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
				ratio := EvalRatio(c.GetE())
				i := 0
				if c.GetE().GetFirstRB() < NChRes+subsetSize*ChRX {
					for i = 0; i < len(MobileListRX); i++ {
						co := MobileListRX[i]
						if ratio < co.EvalRatio() {
							break
						}
					}
					MobileListRX = append(MobileListRX[:i], append([]*Connection{c}, MobileListRX[i:]...)...)
					//MobileListRX.Insert(i, c)
				} else {
					for i = 0; i < len(MobileList); i++ {
						co := MobileList[i]
						if ratio < EvalRatio(co.GetE()) {
							break
						}
					}
					MobileList = append(MobileList[:i], append([]*Connection{c}, MobileList[i:]...)...)
					//MobileList.Insert(i, c)
				}
			}
		}
	}

	// change channel to some mobiles

	fact := 0.8
	var MobileListUSE []*Connection
	if len(MobileListRX) > 0 {
		MobileListUSE = MobileListRX
		fact = 0
	} else {
		MobileListUSE = MobileList
	}

	for k := 0; k < len(MobileListUSE) && k < 1; k++ {
		co := MobileListUSE[k]
		E := co.GetE()
		ratio := EvalRatio(E) * fact

		chHop := FindFreeChan(dbs, E, ratio)
		oldCh := E.GetFirstRB()

		if chHop > 0 {
			for l := 0; l < subsetSize; l++ {
				E.UnSetARB(oldCh + l)
				E.SetARB(chHop + l)
				Hopcount++
			}
		}
	}
}

func EvalRatio(E EmitterInt) float64 {

	ratio := 0.0
	oldCh := E.GetFirstRB()
	for l := 0; l < subsetSize; l++ {
		ratio += E.GetSNRrb(oldCh + l)
	}
	ratio /= float64(subsetSize)
	return ratio
}

func FindFreeChan(dbs *DBS, E *Emitter, ratio float64) (nch int) {	

	var SNRs [NCh]float64

	copy(SNRs[:],E.SNRrb[:])
	r := dbs.Pos
	ICIMfunc(&r, E, SNRs[:], dbs.Color)

	for j := 1; j < NCh-subsetSize+1; j += subsetSize {
		rb := j // dbs.RndCh[j]
		if !dbs.IsInFuturUse(rb) {			
			snr := 0.0
			for l := 0; l < subsetSize; l++ {
				snr += SNRs[rb+l]
			}
			snr /= float64(subsetSize)			
			if snr > ratio {
				ratio = snr
				nch = rb
				//assign and exit
			}
		}
	}

	return 

}

//   Reformatted by   lerouxp    Mon Oct 3 09:49:03 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 31 15:43:58 CET 2011

