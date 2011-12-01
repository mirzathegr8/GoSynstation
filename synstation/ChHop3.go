package synstation

//import "container/vector"
import "math"
import rand "rand"
import "fmt"
import "sort"
//import "geom"


func init() {
	fmt.Println("init")
}

//var plus int


type AssignedRB struct {
	E      int     // index of MobileList of the emitting mobile
	rb     int     // its RB
	snr    float64 //its snr
	metric float64 //its metric
}


func ChHopping3(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	var MobilesList [NConnec]*Emitter

	var MobilesRate [NConnec]float64
	var AssignedRBs [NConnec * 50]AssignedRB //should not assigne more than 50 RBs to one mobile
	var numAsRB int
	var numMaster int

	numDeAll := 1 // number of RBs to de-allocate
	numAll := 1   // number of RBs to allocate
	NumTotalARB := 0
	for rb := NChRes; rb < NCh; rb++ {
		if !dbs.IsRBFree(rb) {
			NumTotalARB++
		}
	}
	if float64(NumTotalARB) > float64(NRB)*dbs.RBReuseFactor {
		numDeAll+= ( NumTotalARB - int(float64(NRB)*dbs.RBReuseFactor) )/2 +1
		numAll--
	} else {
		numDeAll--
		numAll+=  (int(float64(NRB)*dbs.RBReuseFactor) - NumTotalARB )/2 +2
	}

	// else add another one

	// find RBs to deallocate
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {

		c := e.Value.(*Connection)
		M := c.GetE()

		if c.Status == 0 { // only change if master

			MobilesList[numMaster] = M
			//Trmm = M.GetMeanTR()

			for rb := 1; rb < NCh; rb++ {

				if M.IsSetARB(rb) {
					AssignedRBs[numAsRB].E = numMaster
					AssignedRBs[numAsRB].rb = rb
					snr := M.GetSNRrb(rb)
					AssignedRBs[numAsRB].snr = snr
					capa := 80 * math.Log2(1+snr)
					if capa < 100 {
						capa = 0
					}
					AssignedRBs[numAsRB].metric = capa
					MobilesRate[numMaster] += capa
					numAsRB++
				}
			}
			numMaster++
		}
	}

	eMobilesList := MobilesList[0:numMaster]
	eAssignedRBs := AssignedRBs[0:numAsRB]

	//make sur noone emits on rb 0
	for i := range eMobilesList {
		if MobilesList[i].IsSetARB(0) {
			MobilesList[i].UnSetARB(0)
		}

	}

	// remove any RB with too low capa	
	//fmt.Println("assignedrbs len ", numAsRB)

	for i := range eAssignedRBs {
		ARB := &eAssignedRBs[i]
		if ARB.metric < 100 { // remove allocation
			eMobilesList[ARB.E].UnSetARB(ARB.rb)
			ARB.E = -1
			ARB.metric = 0
		} else {
			ARB.metric = math.Log2(1+ARB.metric) / math.Log2(2+MobilesRate[ARB.E]) / math.Log2(2+eMobilesList[ARB.E].GetMeanTR())
		}
	}

	// Sort mobiles
	var RBmetrics [NConnec * 100]float64
	for i := range eAssignedRBs {
		RBmetrics[i] = eAssignedRBs[i].metric
	}
	S := initSequence(RBmetrics[0:numAsRB])
	sort.Sort(S)
	//fmt.Print("disconnect"); S.PrintOrder(); fmt.Println();

	// dealocate a number of RBs
	for ir, irm := 0, 0; ir < numAsRB && irm < numDeAll; ir++ {
		ARB := &eAssignedRBs[S.index[ir]]
		if ARB.E >= 0 {
			eMobilesList[ARB.E].UnSetARB(ARB.rb)
			irm++
		}
	}

	// find potential RBs to allocate

	var NonAssignedRBs [100 * NConnec]AssignedRB
	numNAsRB := 0
	for rb := NChRes; rb < NCh; rb++ {

		if !dbs.IsInFuturUse(rb) {

			for j, M := range eMobilesList {
				NonAssignedRBs[numNAsRB].E = j
				NonAssignedRBs[numNAsRB].rb = rb
				snr := M.GetSNRrb(rb)
				NonAssignedRBs[numNAsRB].snr = snr
				capa := 80 * math.Log2(1+snr)
				if capa < 100 {
					capa = 0
				}
				NonAssignedRBs[numNAsRB].metric = capa
				numNAsRB++
			}

		}
	}

	eNonAssignedRBs := NonAssignedRBs[0:numNAsRB]

	//eval metric
	for i := range eNonAssignedRBs {
		ARB := &eNonAssignedRBs[i]
		ARB.metric =  math.Log2(1+ARB.metric) / math.Log2(2+MobilesRate[ARB.E]+ARB.metric) /
							 math.Log2(2+eMobilesList[ARB.E].GetMeanTR())
	}

	//sort

	for i := range eNonAssignedRBs {
		RBmetrics[i] =  eNonAssignedRBs[i].metric //negative to sort the other way arround
	}
	//fmt.Println(RBmetrics[0:numNAsRB])
	S = initSequence(RBmetrics[0:numNAsRB])
	sort.Sort(S)
	//fmt.Print("connect"); S.PrintOrder(); fmt.Println();

	for ir := 0; ir < numNAsRB && ir < numAll; ir++ {
		ARB := &eNonAssignedRBs[S.index[ir]]
		//if ARB.metric <= 0.0 {break} // snr insufficient
		eMobilesList[ARB.E].SetARB(ARB.rb)
		Hopcount++
	}

}

func EvalRatio2(E EmitterInt) float64 {

	ratio := 0.0
	numarb := 0
	for rb, use := range E.GetARB() {
		if use {
			ratio += E.GetSNRrb(rb)
			numarb++
		}
	}
	ratio /= float64(numarb)
	return ratio
}

func FindFreeChannels(dbs *DBS, E *Emitter, ratio float64) []int {

	var nch [20]int //maximum 20RBs allocated

	var SNRs [NCh]float64

	for i := range SNRs {
		SNRs[i] = E.GetSNRrb(i)
	}

	r := dbs.Pos

	ICIMfunc(&r, E, SNRs[:], dbs.Color)

	S := initSequence(SNRs[:])
	sort.Sort(S)

	NumARB := int(float64(NCh) / float64(dbs.Connec.Len()) * dbs.RBReuseFactor)

	return nch[0:NumARB]

}

//   Reformatted by   lerouxp    Mon Oct 3 16:12:22 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 3 16:50:49 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 3 16:57:42 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 3 17:17:10 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 3 17:42:18 CEST 2011

//   Reformatted by   lerouxp    Mon Oct 3 18:06:29 CEST 2011

