package synstation

import "container/list"
//import "fmt"

type channel struct {
	i        int         //id of channel
	Emitters *list.List  //list of mobiles in channels
	coIntC   []coIntChan //set of interfering channels and factors

	Change chan EmitterInt // channel to inform change of emitter
//	Remove chan EmitterInt // channel to inform change of emitter

	added, removed int //counter

	//done chan int
}


func (c *channel) Init(i int) {
	c.i = i
	c.Emitters = list.New()
	c.Change = make(chan EmitterInt, M+10)
//	c.Remove = make(chan EmitterInt, M+10)

	//c.done = make(chan int)

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
			//	fmt.Println("remove ")
			return
		}
	}

}

var countp, countm int

func (ch *channel) ChangeChan() {

	for true {

		tx := <-ch.Change
		if tx != nil {
			tx._setCh(ch.i)
			//ch.addToChan(tx)
			//ch.added++

		} else {
			break
		}
		if ch.i != 0 {
			countp++
		}

	}
	SyncChannel <- 1

}

//func (ch *channel) RemoveChan() {
//
//	/*for true {
//		tx := <-ch.Remove
//		if tx != nil {
//			if ch.i == 0 {
//				//	fmt.Println(" remove 0 ", tx.GetId())
//			}
//			ch.remove(tx)
//			ch.removed++
//		} else {
//			break
//		}
//		if ch.i != 0 {
//			countm++
//		}
//	}*/
//
//	SyncChannel <- 1
//
//}
//

func ChannelHop() {

	countp = 0
	countm = 0

//	for i := range SystemChan {
//		SystemChan[i].Remove <- nil
//		go SystemChan[i].RemoveChan()
//
//	}
//	for i := 0; i < NCh; i++ {
//		_ = <-SyncChannel
//	}

	for i := range SystemChan {
		SystemChan[i].Change <- nil
		go SystemChan[i].ChangeChan()

	}
	for i := 0; i < NCh; i++ {
		_ = <-SyncChannel
	}

	//fmt.Println(" Count ", countp, countm)

}
// counters to observe how many mobiles or added or removed from a channel
// usefull for channel 0 to know how many mobiles where disconnected
func (ch *channel) GetAdded() int   { a := ch.added; ch.added = 0; return a }
func (ch *channel) GetRemoved() int { a := ch.removed; ch.removed = 0; return a }

