package synstation

import "container/list"
//import "fmt"


type channel struct {
	i        int        //id of channel
	Emitters *list.List //list of mobiles in channels

	coIntC []coIntChan //set of interfering channels and factors

	added, removed int //counter
}


func (c *channel) Init(i int) {
	c.i = i
	c.Emitters = list.New()
}

// Channels used in the simulation
var SystemChan []*channel

// structure that specifies a channel and how much it overlaps the channel to which this instance belongs to
type coIntChan struct { // co-interfering channels
	factor float64
	c      int
}



// counters to know how many mobiles where added / removed from RB
var countp, countm int

// counters to observe how many mobiles or added or removed from a channel
// usefull for channel 0 to know how many mobiles where disconnected
func (ch *channel) GetAdded() int   { a := ch.added; ch.added = 0; return a }
func (ch *channel) GetRemoved() int { a := ch.removed; ch.removed = 0; return a }


// by default the futurARB vector is kept the same
// it is the ARB scheduler that has to clear it
func (ch *channel) HopChans() {

	for m :=range Mobiles{
		M:= &Mobiles[m]
		if M.ARBfutur[ch.i]==false && M.ARB[ch.i]==true{
			ch.Emitters.Remove(M.ARBe[ch.i])
			M.ARBe[ch.i]=nil
			M.ARB[ch.i]=false;
			ch.removed++
		} else if M.ARBfutur[ch.i]==true{
	
			if M.ARB[ch.i]==false{
				M.ARBe[ch.i]= ch.Emitters.PushBack(M.GetE())
				M.ARB[ch.i]=true				
				ch.added++		
			}
		}
	}

	SyncChannel <- 1
}


func ChannelHop() {
		
	countp = 0
	countm = 0

	for i := range SystemChan {	
		go SystemChan[i].HopChans()
	}
	for i := 0; i < NCh; i++ {
		_ = <-SyncChannel
	}	

}

