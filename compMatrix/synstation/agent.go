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

	for i := range Agents {
		go func(i int) {
			Agents[i].RunAgent()
		}(i)
	}
	Sync(len(Agents))

}


