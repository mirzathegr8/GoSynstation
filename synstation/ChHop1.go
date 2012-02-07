package synstation

import rand "math/rand"

//import "container/vector"
import "math"

//
//func (dbs *DBS) channelHopping2() {
//
//	//pour trier les connections actives
//	var MobileList vector.Vector
//
//	//pour trier les canaux
//	dbs.RandomChan()
//
//	// find a mobile
//	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
//		c := e.Value.(*Connection)
//
//		if c.Status == 0 { // only change if master
//
//			if c.E.IsSetARB(0) { //if the mobile is waiting to be assigned a proper channel
//
//				var ratio float64
//				nch := 0
//
//				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
//				for j := NChRes; j < NCh; j++ {
//					i := dbs.RndCh[j]
//					if !dbs.IsInUse(i)  {
//
//						_, ber, snr, _ := dbs.R.EvalSignalBER(c.E, i)
//						ber = math.Log10(ber)
//
//						if ber < math.Log10(BERThres/10) {
//							if snr > ratio {
//								ratio = snr
//								nch = i
//								//assign and exit
//							}
//						}
//					}
//				}
//				if nch != 0 {
//					dbs.changeChannel(c, nch)
//					return
//				}
//
//				// sort mobile connection for channel hopping
//			} else {
//				ratio := c.EvalRatio(dbs.R)
//				var i int
//				for i = 0; i < MobileList.Len(); i++ {
//					co := MobileList.At(i).(ConnecType)
//					if ratio < co.EvalRatio(dbs.R) {
//						break
//					}
//				}
//				MobileList.Insert(i, c)
//			}
//		}
//	}
//
//	// change channel to some mobiles
//	for k := 0; k < MobileList.Len() && k < 15; k++ {
//		co := MobileList.At(k).(ConnecType)
//		//ratio := co.EvalRatio(&dbs.R)		
//
//		d := co.E.GetPos().Distance(dbs.R.GetPos())
//
//		//if (10*math.Log10(co.GetSNR())< SNRThres){
//		//var ir int
//		//ir:= NCh-NChRes + (6+int(math.Log10(Pr)))
//		//if d<100 {ir=28
//		//}else {ir=0}
//
//		//	ir:= NCh-NChRes + int(( -float(d)/1500*float(NCh-NChRes) ))
//		//if (ir<0) {ir=0}
//		//	if ir> NCh-2 {ir=NCh-2}
//
//		ir := 5
//		if d < 300 {
//			if !(co.E.GetARB()[0] > NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
//				for j := NCh - ir; j < NCh; j++ {
//
//					i := dbs.RndCh[j]
//					if !dbs.IsInUse(i) && i != co.E.GetARB()[0] {
//						_, snr, _, _ := dbs.R.EvalSignalSNR(co.E, i)
//						if snr > SNRThresChHop {
//							dbs.changeChannel(co, i)
//							Hopcount++
//							break
//						}
//					}
//				}
//			}
//		} else {
//
//			if !(co.E.GetARB()[0] < NCh-ir) || (co.GetSNR() < SNRThresChHop-3) {
//				for j := NChRes; j < NCh-ir; j++ {
//
//					i := dbs.RndCh[j]
//
//					if !dbs.IsInUse(i) && i != co.E.GetARB()[0] {
//						_, snr, _, _ := dbs.R.EvalSignalSNR(co.E, i)
//						if snr > SNRThresChHop {
//							dbs.changeChannel(co, i)
//							Hopcount++
//							break
//						}
//					}
//				}
//			}
//
//		}
//
//	}
//	//}
//
//	/*
//		for k := 0; k < MobileList.Len() && k < 1; k++ {
//			co := MobileList.At(k).(ConnecType)
//			ratio := co.EvalRatio(&dbs.R)
//
//
//			if (Pr<8e-9) && co.E.GetCh()>NCh-3{
//
//		//push down		
//			for j := NChRes; j < NCh; j++ {
//
//				i := dbs.RndCh[j]
//
//				if !dbs.IsInUse(i) && i != co.E.GetCh() {
//					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
//					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
//						dbs.changeChannel(co, i)
//						Hopcount++
//						break
//					}
//				}
//			}} else{
//
//		//push up		
//			for j := NCh-2; j < NCh; j++ {
//
//				i := dbs.RndCh[j]			
//				if !dbs.IsInUse(i) && i != co.E.GetCh() {
//					Rnew, ev, Pr:=dbs.R.EvalSignalBER(co.E,i)
//					if Pr/(Rnew.Pint+WNoise) > ratio/2 {
//						dbs.changeChannel(co, i)
//						Hopcount++
//						break
//					}
//				}
//			}}
//
//
//		}
//	*/
//
//
//}
//

func ChHopping(dbs *DBS, Rgen *rand.Rand) {

	//pour trier les connections actives
	//var MobileList vector.Vector
	MobileList := make([]*Connection, NConnec)
	MobileList = MobileList[0:0]

	//pour trier les canaux
	dbs.RandomChan()

	var stop = 0

	// find a mobile
	for e := dbs.Connec.Front(); e != nil; e = e.Next() {
		c := e.Value.(*Connection)

		if c.Status == 0 { // only change if master

			if c.E.ARB[0] { //if the mobile is waiting to be assigned a proper channel

				var ratio float64
				nch := 0

				//Parse channels in some order  given by dbs.RndCh to find a suitable channel 
				for j := NChRes; j < NCh; j++ {
					rb := dbs.RndCh[j]
					if dbs.IsRBFree(rb) && !c.E.ARB[rb] {
						snr := dbs.EvalSignalSNR(c.E, rb)
						if 10*math.Log10(snr) > SNRThresChHop {
							if snr > ratio {
								ratio = snr
								nch = rb
								//assign and exit
							}
						}
					}
				}
				if nch != 0 {
					dbs.changeChannel(c, 0, nch)
					stop++
					if stop > 5 {
						return
					}
				}

				// sort mobile connection for channel hopping
			} else {
				ratio := c.EvalRatio()
				var i int
				for i = 0; i < len(MobileList); i++ {
					co := MobileList[i]
					if ratio < co.EvalRatio() {
						break
					}
				}
				MobileList = append(MobileList[:i], append([]*Connection{c}, MobileList[i:]...)...)
				//MobileList.Insert(i, c)
			}
		}
	}

	// change channel to some mobiles
	for k := 0; k < len(MobileList) && k < 2; k++ {
		co := MobileList[k]
		ratio := co.EvalRatio()
		chHop := 0

		for j := NChRes; j < NCh; j++ {

			rb := dbs.RndCh[j]

			if dbs.IsRBFree(rb) && !co.E.ARB[rb] {

				snr := dbs.EvalSignalSNR(co.E, rb)

				if snr > ratio {
					ratio = snr
					chHop = rb
				}
			}
		}
		if chHop > 0 {
			dbs.changeChannel(co, co.E.GetFirstRB(), chHop)
			Hopcount++
		}

	}

}

func (dbs *DBS) changeChannel(co *Connection, pch, nch int) {
	co.E.ARBfutur[pch] = false
	co.E.ARBfutur[nch] = true
}

//   Reformatted by   lerouxp    Mon Oct 3 09:49:11 CEST 2011
