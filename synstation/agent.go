package synstation


type Agent interface {
	RunPhys()
	FetchData()
	RunAgent()
}


func syncThread() { SyncChannel <- 1 }


func Sync(k int) {
	for n := 0; n < k; n++ {
		<-SyncChannel
	}
}

func GoRunPhys() {


	for i := range Synstations {
		go func(i int) {
			Synstations[i].RunPhys()
		}(i)

	}
	Sync(D)
}


func GoFetchData() {

	for i := range Mobiles {
		go func(i int) {
			Mobiles[i].FetchData()
		}(i)
	}
}


func GoRunAgent() {


	for _, a := range Agents {
		go func(a Agent) {
			a.RunAgent()
		}(a)
	}
	Sync(len(Agents))

}


//go tutorial

/*	A := new(DBS)
	var B Agent
	B = A	// B i not a pointer to DBS
	B->MyAgentmehod() // not go
	B.MyDBSmethod()	  // this is not OK
	B.MyAgentmethod() // this OK
	C := B.(*DBS) // C is a *DBS
	
	C.MyDBSmethod() // this is ok
*/

