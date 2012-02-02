package synstation
const BERThres= 0.15
const SNRThresConnec= 0
var initScheduler=initARBSchedulerSDMA // ChHopping2  //ARBSchedulerSDMA // //ARBSchedulerSDMA //  ChHopping2 //
//var estimateFactor=estimateFactor0
const conservationFactor=10
const DiversityType=MRC
const BeamAngle=  0.4254//  1.1345//
const mDf=4
const NetLayout= HONEYCOMB
const subsetSize= 3
const InterferenceCancel= SIZEESCANCELATION //NOCANCEL ////NOCANCEL //
const SetReceiverType = BEAM

//types of eNode interference cancelation
// either none, or only master signals are canceled from others received signals
// or all mobiles listened too are canceled from every listened signals
// or SizeES best received signals on each RB are canceld out
const (
	NOCANCEL = iota
	MASTERCANCELATION
	CONNECTEDCANCELATION
	SIZEESCANCELATION
)

// number of signal id saved in the list of the ChanReceiver
const SizeES = 4
