package synstation


const meanN = 100

type MeanData struct {
	sum   float64
	meanD [meanN]float64
	p     int
}

func (m *MeanData) Add(a float64) {
	m.sum += a
	m.sum -= m.meanD[m.p]
	m.meanD[m.p] = a
	m.p++
	if m.p >= meanN {
		m.p = 0
	}
}
func (m *MeanData) Get() float64 {
	return m.sum / meanN
}
/*
func (m *MeanData) GetLast() float64 {
	if m.p == 0 {
		return m.meanD[meanN-1]
	}
	return m.meanD[m.p-1]
}*/


func (m *MeanData) Clear(a float64) {
	m.p = 0
	for i := range m.meanD {
		m.meanD[i] = a
	}
	m.sum = a * meanN
}

