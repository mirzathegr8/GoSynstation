package synstation


const Field = 6000 //length in meters

const Duration = 2000 // in iterations 

const M = 500 //numbers of mobiles
const D = 80  // numbers of DBS

const L2 = 2 // modulation factor
const L1 = 1

const NCh = 100 // number of channels

// 10 0 11 .1 12 .2 19 .5 37 .75
const roverlap = float64(0.0) // ratio of overlaping of two adjacent channels

const WNoise = 7.45e-17 // White noise at reciever
const NChRes = 5        //numbers of reserved channels, not used yet, but chan 0 must be reserved
const NConnec = 25      // numbers of connections per dbs

const BERThres = float64(0.3)
const SNRThresConnec = float64(35)

const SNRThresChHop = float64(0)

const MaxSpeed = float64(15)

const EnodeBClock = 2


const SetReceiverType = BEAM

//type ReceiverType int

const (
	OMNI = iota
	BEAM
	SECTORED
)

const SetShadowMap = NOSHADOW

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

const FastFading = MONTECARLO

const (
	MEANEVAL = iota
	MONTECARLO
)

const DiversityType = SELECTION

const (
	SELECTION = iota
	MRC
)

const PowerControl = NOPC

const (
	NOPC = iota
	AGENTPC
)

const BWallocation = ARBSCHEDUL

const (
	CHHOPPING = iota
	ARBSCHEDUL
)

