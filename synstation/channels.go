package synstation

import "container/list"


import "math"


type channel struct {
	i        int		//id of channel
	Emitters *list.List	//list of mobiles in channels
	coIntC   []coIntChan	//set of interfering channels and factors

	Change chan EmitterInt	// channel to inform change of emitter

	added, removed int	//counter
}

// Channels used in the simulation
var SystemChan []*channel

// function called automatically on pacakge load, initializes system channels
func init() {

	// create channels
	
	SystemChan = make([]*channel, NCh)
	for i := range SystemChan {
		c := new(channel)
		c.i = i
		c.Emitters = list.New()
		c.Change = make(chan EmitterInt, 100)
		go c.changeChan()
		SystemChan[i] = c
	}


	// evaluate overlaping factors of channels	

	overlapN:= int(math.Floor (1.0 /(1.0-float64(roverlap))));
	if overlapN<1 {overlapN=1}
	if roverlap>0.0{ 


	for i:=NChRes ; i<NCh ; i++{

		SystemChan[i].coIntC= make([]coIntChan, overlapN*2)

		for k:=1; k<=overlapN; k++{

			fac:= 1.0- float(k)*(1.0- roverlap)

			SystemChan[i].coIntC[overlapN-k].c=i-k
			SystemChan[i].coIntC[overlapN-k].factor= float64(fac)

			if (i+k<NCh){
				SystemChan[i].coIntC[overlapN+k-1].c=i+k
				SystemChan[i].coIntC[overlapN+k-1].factor= float64(fac)
			}

		}

	}}


}

// structure that specifies a channel and how much it overlaps the channel to which this instance belongs to
type coIntChan struct { // co-interfering channels
	factor float64
	c      int
}


// functions to deal with synchronized lists

func (ch *channel) addToChan(er EmitterInt) {
	ch.Emitters.PushBack(er)
}

func (ch *channel) remove(er EmitterInt) {

	for e := ch.Emitters.Front(); e != nil; e = e.Next() {
		if e.Value.(EmitterInt) == er {
			ch.Emitters.Remove(e)		
			return
		}
	}
	
}

func (ch *channel) changeChan() {

	for tx := range ch.Change {

		if tx.GetCh() != ch.i {
			ch.remove(tx)
			SystemChan[tx.GetCh()].Change <- tx
			ch.removed++
		} else {
			ch.addToChan(tx)
			tx.isdone() <- 1
			ch.added++
		}

	}

}


// counters to observe how many mobiles or added or removed from a channel
// usefull for channel 0 to know how many mobiles where disconnected
func (ch *channel) GetAdded() int   { a := ch.added; ch.added = 0; return a }
func (ch *channel) GetRemoved() int { a := ch.removed; ch.removed = 0; return a }


