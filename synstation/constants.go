package synstation

const Field = 6000 //length in meters

const Duration = 500 // in iterations 

const M = 1000 //numbers of mobiles
const D = 100  // numbers of DBS

const L2 = 2 // modulation factor

var NCh = 70 // number of channels
// 10 0 11 .1 12 .2 19 .5 37 .75
var roverlap = 0.0 // ratio of overlaping of two adjacent channels

const WNoise = 1e-12 // White noise at reciever
var NChRes = 5       //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25   // numbers of connections per dbs

var BERThres = float64(0.8)
var SNRThresConnec = float64(35)

var SNRThresChHop = float64(0)

var MaxSpeed = float64(5)


var SetReceiverType = SECTORED

type ReceiverType int

const (
	OMNI = iota
	BEAM
	SECTORED
)

