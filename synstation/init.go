package synstation

import "math"
import "fmt"
import "rand"
import "container/list"

var SyncChannel chan float64

func init() {

	SyncChannel = make(chan float64, 100000)
	fmt.Println(" SyncChannel created")

	Rgen = rand.New(rand.NewSource(123813541954235))
	Rgen2 = rand.New(rand.NewSource(12384235))
	fmt.Println("init done")

	// function called automatically on pacakge load, initializes system channels

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

	overlapN := int(math.Floor(1.0 / (1.0 - float64(roverlap))))
	if overlapN < 1 {
		overlapN = 1
	}
	if roverlap > 0.0 {

		for i := NChRes; i < NCh; i++ {

			SystemChan[i].coIntC = make([]coIntChan, overlapN*2)

			for k := 1; k <= overlapN; k++ {

				fac := 1.0 - float64(k)*(1.0-roverlap)

				SystemChan[i].coIntC[overlapN-k].c = i - k
				SystemChan[i].coIntC[overlapN-k].factor = float64(fac)

				if i+k < NCh {
					SystemChan[i].coIntC[overlapN+k-1].c = i + k
					SystemChan[i].coIntC[overlapN+k-1].factor = float64(fac)
				}

			}

		}
	}

}


// what a random variable doing in constant declaration file???
// ah, ok the random generator is constant,... :p
var Rgen *rand.Rand  //on used for position
var Rgen2 *rand.Rand //one used to init shadow maps
//different rndvar are used to ensure repeatability of position with or without shadow maps

