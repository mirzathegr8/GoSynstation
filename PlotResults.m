

figure(1+fign);
load TransferRate.mat ;
plot(sort(sum(TransferRate,2)/size(TransferRate,2)*1000/1e6),c);
%ylabel ("capacity")
xlabel("mobiles")
hold on;

figure(2+fign);
load Outage.mat;
[rr,T]=outage(Outage);
plot(log10(rr/size(TransferRate,1)*1000+1),c);
%ylabel ("number of mobiles in outage")
xlabel("TTI")
hold on;

figure(3+fign);

load SNR.mat
load NumARB.mat
%plot(10*log10( sort(mean( SNR./(NumARB+0.000001) .* (NumARB>0)  ,2 )      )   ),c)
plot(10*log10( sort(mean( SNR ,2 )      )   ),c)
hold on;

figure(4+fign);

load InstSNR.mat
load NumARB.mat
%plot(10*log10( sort(mean( SNR./(NumARB+0.000001) .* (NumARB>0)  ,2 )      )   ),c)
plot(10*log10( sort(mean( InstSNR ,2 )      )   ),c)
hold on;

figure(6+fign);
load Ptxr.mat;
plot(sort(mean(Ptxr,2)),c);
hold on;


figure(5+fign);
load NumARB.mat;
[V,I]=sort(mean(NumARB,2));
plot(V,c);
hold on;

#figure(5);
#plot(mean(TransferRate,2)(I),c);
#hold on


#'Total TransferRate'
#sum(sum(TransferRate))
#'Fairness square error'
#tt=sum(TransferRate,2);
#tm=(tt-mean(tt));
#sum(tm(tm<0).^2)
#'Fairness square error2'
#tm=(tt-median(tt));
#sum(tm(tm<0).^2)
