

figure(1);
load TransferRate.mat ;
plot(sort(sum(TransferRate,2)),c);
hold on;

figure(2);
load Outage.mat;
[rr,T]=outage(Outage);
plot(log10(rr+1),c);
hold on;

figure(3);
load Ptxr.mat;
plot(sort(mean(Ptxr,2)),c);
hold on;


figure(4);
load NumARB.mat;
[V,I]=sort(mean(NumARB,2));
plot(V,c);
hold on;

figure(5);
plot(mean(TransferRate,2)(I),c);
hold on


sum(sum(TransferRate))
