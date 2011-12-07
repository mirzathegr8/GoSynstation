package synstation

const BERThres= 0.15
const SNRThresConnec= 0
var ARBSchedulFunc=ARBSchedulerSDMA //ChHopping2//
var estimateFactor=estimateFactor0
const conservationFactor=0.2
const DiversityType=MRC
const SetReceiverType=BEAM
const NetLayout= HONEYCOMB

const TransferRateTechnique=    MCSJOINT //NormalTR  //EFFECTIVESINR//, MCSJOINT //NormalTR   , MCSJOINT

const ICIMtype= ICIMb

// ICIM Theta is for the moment deprecated,
const ICIMTheta=false
// ICIMdistRatio sets the ratio of distance (emiter-enode/neighboring enode) for reusing all RB in the cell's center
const ICIMdistRatio=0.3


// These two parameters allows arranging eNodeBs to colocate 3 into one and allow RB reuse
// a hack is being used to prevent the 3 colocated enodes to connect more than once to one emitter
// this is a work in progress and the architecture needs to be rethought TODO
//const OneAgentPerBEAM =false // controls for beamforming or sectorisation the ability to reuse RB accross different beam
const mDf =1 // multiplies the number of eNodes for beamforming or SECTORED2 which creates one enode per sector  

// The enodebclock, that sets the interval before reactivation
const EnodeBClock = 6

//var ICIMfunc=ICIMSplitEdgeCenter2
var ICIMfunc=ICIMNone
var PowerAllocation =      optimizePowerAllocationAgent // optimizePowerAllocationAgentRB//  optimizePowerAllocationAgent // optimizePowerNone // 
const PowerAgentFact = 0.8//0.8//0.2 
const PowerAgentAlpha = 1 //0.8//0.2 

var subsetSize=4

const uARBcost = 0.10000 //meanMeanCapa / 5 //0.5 // math.Log2(1 + meanMeanCapa)
