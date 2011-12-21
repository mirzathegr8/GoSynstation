


%vNetLayout = {"RANDOM"; "HONEYCOMB"}
vNetLayout = {"HONEYCOMB"}
%V_BERThres=[0.05:0.05:0.25]
%V_BERThres=[0.5 0.15 0.05];
V_BERThres=[0.15];
%V_SNRThresConnec=[0 5 10 15 20 25 30 35 ];
V_SNRThresConnec=[0];
%ARBSchedul={"ARBScheduler4"} %;"ARBScheduler";"ChHopping2"};
%EstimateF={"estimateFactor0";"estimateFactor1";"estimateFactor1"};
ARBSchedul={"ChHopping2"};
EstimateF={"estimateFactor1"};
SUBSETSIZE=[6,7,8]

%DIVTYP={"SELECTION";"MRC"};
DIVTYP={"MRC"};
%RECEIVERTYP={"OMNI";"BEAM"};

%Array_cF={ [0.2 1 10 20];
%	 [0.2 0.8 2 10];
%	 [0.2 0.8 2 10]};

Array_cF={[1]};

BEAMFORM= [1.1345 0.85088 0.68070 0.56725 0.4254]
MDF=[1 2 3 4 5 6 8]

for sss=SUBSETSIZE
for h=1:size(vNetLayout,1)
NetLayout=vNetLayout{h};
for BER= V_BERThres
for SNR= V_SNRThresConnec
for i=1:size(ARBSchedul,1)
ARB=ARBSchedul{i};
est=EstimateF{i};
for cF= Array_cF{i}
for j=1:size(DIVTYP,1)
divType=DIVTYP{j};

for mdf=MDF
for bf=BEAMFORM


DIR=sprintf("SDMA_%s_%1.2f_%d_%s_%1.1f_%s_%1.1f_%d_sss%d",NetLayout,BER,SNR,ARB,cF,divType,bf,mdf,sss)

if exist([DIR "/out.mat"]) == 0 %here we check for the existance of the output result file, to run the simulation only if we dont already have the results

mkdir(DIR);

 
fid = fopen('synstation/constantV.go', 'w');
fprintf(fid,"package synstation\n");
fprintf(fid,"const BERThres= %1.2f\n", BER);
fprintf(fid,"const SNRThresConnec= %d\n", SNR)
fprintf(fid,"var ARBSchedulFunc=%s\n", ARB)
fprintf(fid,"var estimateFactor=%s\n", est)
fprintf(fid,"const conservationFactor=%1.1f\n", cF)
fprintf(fid,"const DiversityType=%s\n", divType)
fprintf(fid,"const BeamAngle=%1.4f\n", bf)
fprintf(fid,"const mDf=%d\n", mdf)
fprintf(fid,"const NetLayout= %s\n",NetLayout)
fprintf(fid,"const subsetSize= %d\n",sss)
fclose(fid)

shell_cmd make 
cd(DIR)
shell_cmd ../simu
cd ..

endif

endfor
endfor
endfor
endfor
endfor
endfor
endfor
endfor
endfor
