

func SaveMobBER( BER chan float64, M int){




}


func Draw(syns []synstation.DBS, mobs []synstation.Mob, k int) {

	data := <-sentData

	l := 0
	for i := range syns {
		for e := syns[i].Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*synstation.Connection)
			data.connec[l].Copy(c)
			data.connec[l].B = syns[i].R.Pos
			data.connec[l].Ch = c.GetCh()
			l++
		}
	}
	data.NumConn = l
	for i := range mobs {
		data.mobs[i] = mobs[i].EmitterS
	}

	data.k = k

	data.ackChan <- 1

}

