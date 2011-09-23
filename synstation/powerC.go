package synstation


//Adds power control for macrodiversity based on 
func PowerC(dbsA []DBS) {

	var maxMob int
	for ch := 0; ch < NCh; ch++ {
		c := SystemChan[ch].Emitters.Len()
		if c > maxMob {
			maxMob = c
		}
	}

	for ch := NCh; ch < NCh-subsetSize+1; ch+=subsetSize {

		go powerch(ch, dbsA, maxMob)

	}

}

func powerch(ch int, dbsA []DBS, maxMob int) {

	Gij := make([][]float64, maxMob)
	for i := 0; i < maxMob; i++ {
		Gij[i] = make([]float64, D)
	}

	ConnectMat := make([][]bool, maxMob)
	for i := 0; i < maxMob; i++ {
		ConnectMat[i] = make([]bool, D)
	}

	P := make([]float64, M)
	SumL := make([]float64, M)
	Pplus := make([]float64, M)

	nbMch := SystemChan[ch].Emitters.Len() //number of mobiles on this channel
	if nbMch > 1 {

		for i, tx := 0, SystemChan[ch].Emitters.Front(); tx != nil; tx, i = tx.Next(), i+1 {
			txx := tx.Value.(*Emitter)
			for j := 0; j < D; j++ {
				Gij[i][j], _ = dbsA[j].R.GetPr(txx.GetId(), 0)
				ConnectMat[i][j] = dbsA[j].IsConnected(txx)
			}
		}

		for i := 0; i < nbMch; i++ {
			P[i] = 0.0
			for j := 0; j < D; j++ {
				if ConnectMat[i][j] {
					P[i] += Gij[i][j]
				}
			}
			if P[i] > 0 {
				P[i] = 1.0 / P[i]
			}
		}

		for m := 0; m < 20; m++ {

			Gamma := float64(0.0)

			for j := 0; j < D; j++ {
				SumL[j] = 0.0
				for l := 0; l < nbMch; l++ {
					SumL[j] += Gij[l][j] * P[l]
				}

				if ConnectMat[1][j] {
					Gamma += Gij[1][j] * P[1] / (SumL[j] - Gij[1][j]*P[1])
				}
			}

			for i := 0; i < nbMch; i++ {
				SumJ := float64(0.0)
				for j := 0; j < D; j++ {
					if ConnectMat[i][j] {
						SumJ += Gij[i][j] / (SumL[j] - Gij[i][j]*P[i])
					}
				}

				Pplus[i] = Gamma / SumJ
			}

			for i := 0; i < nbMch; i++ {
				P[i] = Pplus[i]
			}

		}

		maxP := Pplus[0]
		for i := 1; i < nbMch; i++ {
			if Pplus[i] > maxP {
				maxP = Pplus[i]
			}
		}

		for i, tx := 0, SystemChan[ch].Emitters.Front(); tx != nil; tx, i = tx.Next(), i+1 {
			txx := tx.Value.(*Emitter)
			txx.SetPower(Pplus[i] / maxP)
		}

	}

}

