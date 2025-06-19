-- best way to get fired
call auth.upsert_user_auth('root','root');


-- grant root on any resource
insert into auth.grants(user_id, role_id, feature_name) 
select distinct USR.user_id, RCO.role_id, RES.feature_name  
from  auth.users USR 
cross join auth.roles RCO 
cross join auth.resources RES 
where USR.user_login = 'root';


-- ensure root has all accesses
select * from auth.get_grants_for_user('root');