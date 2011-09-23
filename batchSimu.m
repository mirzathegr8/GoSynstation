


vNetLayout = {"RANDOM"; "HONEYCOMB"}
%V_BERThres=[0.05:0.05:0.25]
V_BERThres=[0.5 0.15 0.05];
%V_SNRThresConnec=[0 5 10 15 20 25 30 35 ];
V_SNRThresConnec=[0];
ARBSchedul={"ARBScheduler4"} %;"ARBScheduler";"ChHopping2"};
EstimateF={"estimateFactor0";"estimateFactor1";"estimateFactor1"};

DIVTYP={"SELECTION";"MRC"};
RECEIVERTYP={"OMNI";"BEAM"};

Array_cF={ [0.2 1 10 20];
	 [0.2 0.8 2 10];
	 [0.2 0.8 2 10]};

%Array_cF={ [0.2 1:5:35];
%	 [0.1:0.2:2];
%	 [0.1:0.2:2]};

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
for k=1:size(RECEIVERTYP,1)
RecType=RECEIVERTYP{k};

DIR=sprintf("%s_%1.2f_%d_%s_%1.1f_%s_%s",NetLayout,BER,SNR,ARB,cF,divType,RecType)

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
fprintf(fid,"const SetReceiverType=%s\n", RecType)
fprintf(fid,"const NetLayout= %s\n",NetLayout)
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

