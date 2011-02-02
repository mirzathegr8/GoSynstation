package synstation

//the dataset structure to draw, with its surface/context to draw to
type Trace struct {
	K       int
	Mobs    []EmitterS
	Connec  []ConnectionS
	NumConn int
}

func CreateTrace(mobs []Mob, syns []DBS, k int) (t *Trace) {
	t = new(Trace)
	t.Init(len(mobs), len(syns)*NConnec)
	t.copyTrace(mobs, syns, k)
	return
}

// function to initialise a dataset, allocate memory, create surface/context and communication channel
func (t *Trace) Init(MobileNumber int, ConnectionNumber int) {
	t.Mobs = make([]EmitterS, MobileNumber)
	t.Connec = make([]ConnectionS, ConnectionNumber)
}

// function used to copy a trace of the data at some point int time
func (t *Trace) copyTrace(mobs []Mob, syns []DBS, k int) {

	l := 0
	for i := range syns {
		for e := syns[i].Connec.Front(); e != nil; e = e.Next() {
			c := e.Value.(*Connection)
			t.Connec[l].Copy(c)
			t.Connec[l].B = *syns[i].R.GetPos()
			t.Connec[l].Ch = c.GetCh()
			l++
		}
	}
	t.NumConn = l
	for i := range mobs {
		t.Mobs[i] = mobs[i].EmitterS
	}

	t.K = k

}

