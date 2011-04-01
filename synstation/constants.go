package synstation


const Field = 12000 //length in meters

const Duration = 2000 // in iterations 

const M = 4000 //numbers of mobiles
const D = 400  // numbers of DBS

const L2 = 2 // modulation factor
const L1 = 1

var NCh = 35 // number of channels
// 10 0 11 .1 12 .2 19 .5 37 .75
var roverlap = float64(0.0) // ratio of overlaping of two adjacent channels

const WNoise = 1e-11 // White noise at reciever
var NChRes = 5       //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25   // numbers of connections per dbs

var BERThres = float64(0.8)
var SNRThresConnec = float64(35)

var SNRThresChHop = float64(0)

var MaxSpeed = float64(15)


var SetReceiverType = BEAM

//type ReceiverType int

const (
	OMNI = iota
	BEAM
	SECTORED
)

var SetShadowMap = SHADOWMAP

//type ReceiverType int

const (
	NOSHADOW = iota
	SHADOWMAP
)


const shadow_deviance = 10
const corr_res = 50
const shadow_sampling = 15
const mval = 0.1
const maplength = 1500
const mapsize = 600

var FastFading = MONTECARLO

const (
	MEANEVAL = iota
	MONTECARLO
)

