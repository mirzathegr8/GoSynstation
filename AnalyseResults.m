
source outage.m

vNetLayout = {"RANDOM"}
V_BERThres=[0.05:0.05:0.25];
V_SNRThresConnec=[0 5 10 15 20 25 30 35 ];
ARBSchedul={"ARBScheduler3";"ARBScheduler";"ChHopping2"};
EstimateF={"estimateFactor0";"estimateFactor1";"estimateFactor1"};

DIVTYP={"SELECTION";"MRC"};
RECEIVERTYP={"OMNI";"BEAM"};

Array_cF={ [ 1   5  6  10 11  15 16   20 21   26   31 ];
	 [0.4:0.2:2];
	 [0.4:0.2:2]};


CapacityMat=[];
OutageMat=[];
m=0;
%default values
BER=0.15;
SNR=15;
i=2;
ARB=ARBSchedul{i};
est=EstimateF{i};
cF=0.8;%Array_cF{i}(3);
j=2;
divType=DIVTYP{j};
k=2;
RecType=RECEIVERTYP{k};


%for BER= V_BERThres
for SNR= V_SNRThresConnec
%for i=1:size(ARBSchedul,1)
%for cF= Array_cF{i}
%for j=1:size(DIVTYP,1)
%for k=1:size(RECEIVERTYP,1)

DIR=sprintf("RANDOM_%1.2f_%d_%s_%1.1f_%s_%s",BER,SNR,ARB,cF,divType,RecType);

cd(DIR);
m=m+1;
load TransferRate.mat ;
CapacityMat(m,:)= sort(sum(TransferRate,2)/size(TransferRate,2)*1000/1e6);

load Outage.mat;
[rr,T]=outage(Outage);
OutageMat(m,:)= log10(rr/size(TransferRate,1)*1000+1);

cd ..

%endfor
%endfor
%endfor
%endfor
%endfor
endfor

figure(1)
mesh(OutageMat(:,1:20:1000))
ylabel("data")
xlabel("TTI")
figure(2)
mesh(CapacityMat(:,1:20:1000))
ylabel("data")
xlabel("mobiles")
