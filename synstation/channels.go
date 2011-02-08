package synstation

import "container/list"


type channel struct {
	i        int         //id of channel
	Emitters *list.List  //list of mobiles in channels
	coIntC   []coIntChan //set of interfering channels and factors

	Change chan EmitterInt // channel to inform change of emitter

	added, removed int //counter
}

// Channels used in the simulation
var SystemChan []*channel

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

