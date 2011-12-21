
source outage.m


BEAMFORM= [1.1345 0.85088 0.68070 0.56725 0.4254]
MDF=[1 2 3 4 5 6 8]
SSS=[5 6 7 8]

%for BER= V_BERThres
%for SNR= V_SNRThresConnec



NetLayout="HONEYCOMB"
BER=0.15
SNR=0
divType="MRC"
ARB="ChHopping2"
cF=1

for i=1:length(BEAMFORM)
for j= 1:length(MDF)
for k= 1:length(SSS)

bf=BEAMFORM(i);
mdf=MDF(j);
sss=SSS(k)
DIR=sprintf("SDMA_%s_%1.2f_%d_%s_%1.1f_%s_%1.1f_%d_sss%d",NetLayout,BER,SNR,ARB,cF,divType,bf,mdf,sss);

cd(DIR);

load TransferRate.mat ;
capa=sort(sum(TransferRate,2)/size(TransferRate,2)*1000/1e6);
CapacityMat(i,j,k)= sum(capa)/1000;
perCtile(i,j,k)= capa(50);
mean10min10max(i,j,k)= mean(capa(1:100))/mean(capa(901:1000));

load Outage.mat;
[rr,T]=outage(Outage);
OutageMat(i,j,k)= log10(rr/size(TransferRate,1)*1000+1)(200); %outage  at 200TTI
CoverageMat(i,j,k)= log10(rr/size(TransferRate,1)*1000+1)(1000); %outage  at 200TTI

load Ptxr.mat;
meanPower(i,j,k)=mean(mean(Ptxr));

cd ..

%endfor
%endfor
%endfor
endfor
endfor
endfor

%figure(1)
%mesh(OutageMat(:,1:20:1000))
%ylabel("data")
%xlabel("TTI")
%figure(2)
%mesh(CapacityMat(:,1:20:1000))
%ylabel("data")
%xlabel("mobiles")ivType=D
