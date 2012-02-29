function [tt,T]= outage(O)   	
 
        [i,j]= find(O==0); % we look for items that precedes a 0 (which gives the TTI period of outage)
	i=i(j>2); %
	j=j(j>2)-1;% we dont want to access data outside the matrix
	T= O(i+size(O,1)*(j-1) ); %get the values
	T=T(T>0); % take only the non null ones, which are the maximum lenth of each period of outage
	e=O(:,size(O,2)); % dont forget the last column for mobiles that have not been connected
	e=e(e>0);
	T=[T; e];
	tt=[];
	for k=1:size(O,2); tt=[tt sum(T>k)]; end % here we take the cumulative density

end
