package synstation

//import "container/vector"
import "math"
import rand "rand"
import "fmt"
//import "geom"

const ChRX = 0



func init() {
	fmt.Println("init to keep fmt")
}

// memory for the scheduler
type ChHopping2 struct {
	MobileList []*Connection 	
	MobileListRX []*Connection
	SNR [NCh] float64
}

func initChHopping2() Scheduler {
	d := new(ChHopping2)	
	d.MobileList = make([]*Connection, NConnec)
	d.MobileListRX = make([]*Connection, NConnec)
	return d
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

func (d *ChHopping2) Schedule(dbs *DBS, Rgen *rand.Rand) {

	
	//pour trier les canaux
	//	dbs.RandomChan()

	var stop = 0
	var nMLRX=0
	var nML=0


	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			//first copy the future vector

			//c.E.CopyFuturARB()

			//if c.E.GetFirstRB()<NChRes {
			if c.E.ARB[0] { //if the mobile is waiting to be assigned a proper channel				

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 

				nch := FindFreeChan(dbs, c.E, math.Pow10(SNRThresChHop/10.0),&d.SNR)

				if nch != 0 {				
					c.E.ARBfutur[0]=false
					for l := 0; l < subsetSize; l++ {
						c.E.ARBfutur[nch + l]=true
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
				ratio := EvalRatio(c.E)
				i := 0
				if c.E.GetFirstRB() < NChRes+subsetSize*ChRX {
					for i = 0; i < nMLRX; i++ {
						co := d.MobileListRX[i]
						if ratio < co.EvalRatio() {
							break
						}
					}
					nMLRX++
					for j:= nMLRX; j>i; j--{
						d.MobileListRX[j]=d.MobileListRX[j-1]
					}
					d.MobileListRX[i] = c
					
				} else {
					for i = 0; i < nML; i++ {
						co := d.MobileList[i]
						if ratio < EvalRatio(co.E) {
							break
						}
					}
					nML++
					for j:= nML; j>i; j--{
						d.MobileList[j]=d.MobileList[j-1]
					}
					d.MobileList[i] = c

//					d.MobileList = append(d.MobileList[:i], append([]*Connection{c}, d.MobileList[i:]...)...)
					
				}
			}
		}
	}

	// change channel to some mobiles
	
	fact := 0.8
	var MobileListUSE []*Connection
	if nMLRX > 0 {
		MobileListUSE = d.MobileListRX[0:nMLRX]
		fact = 0
	} else {
		MobileListUSE = d.MobileList[0:nML]
	}

	for k := 0; k < len(MobileListUSE) && k < 1; k++ {
		co := MobileListUSE[k]
		E := co.E
		ratio := EvalRatio(E) * fact

		chHop := FindFreeChan(dbs, E, ratio,&d.SNR)
		oldCh := E.GetFirstRB()

		if chHop > 0 {
			for l := 0; l < subsetSize; l++ {
				E.ARBfutur[oldCh + l]=false
				E.ARBfutur[chHop + l]=true
				Hopcount++
			}
		}
	}
}

func EvalRatio(E *Emitter) float64 {

	ratio := 0.0
	oldCh := E.GetFirstRB()
	for l := 0; l < subsetSize; l++ {
		ratio += E.SNRrb[oldCh + l]
	}
	ratio /= float64(subsetSize)
	return ratio
}

func FindFreeChan(dbs *DBS, E *Emitter, ratio float64, SNRs *[NCh]float64) (nch int) {	



	copy(SNRs[:],E.SNRrb[:])

	ICIMfunc(&dbs.Pos, E, SNRs[:], dbs.Color)

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

