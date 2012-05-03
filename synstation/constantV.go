package synstation
const BERThres= 0.15
const SNRThresConnec= 00
var initScheduler= initChHopping2//  initARBSchedulerSDMA //initARBScheduler5 // initChHopping2  // initARBScheduler4 // initARBScheduler1//
//var estimateFactor=estimateFactor0
const conservationFactor=1
const DiversityType=SELECTION
//const BeamAngle=  1.4345//  1.1345//
const mDf=1
const NetLayout=  HONEYCOMB 
const subsetSize= 5

// The enodebclock, that sets the interval before reactivation
const EnodeBClock = 1

//const InterferenceCancel= SIZEESCANCELATION //NOCANCEL ////NOCANCEL //
//const SetReceiverType = BEAM

//types of eNode interference cancelation
// either none, or only master signals are canceled from others received signals
// or all mobiles listened too are canceled from every listened signals
// or SizeES best received signals on each RB are canceld out
/*const (
	NOCANCEL = iota
	MASTERCANCELATION
	CONNECTEDCANCELATION
	SIZEESCANCELATION
)*/

// number of signal id saved in the list of the ChanReceiver
const SizeES = 5

const NArMax =5
const NAtMax =5
