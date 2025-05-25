-- add roles (constants, needs code refactoring for a change)
insert into auth.roles(role_name, role_description) values ('root','allows any action with grant actions too');
insert into auth.roles(role_name, role_description) values ('admin','allows any action but cannot grant');
insert into auth.roles(role_name, role_description) values ('editor','crud operations are allowed');
insert into auth.roles(role_name, role_description) values ('reader','read only operations are allowed');



-----------------------------------------------------------
-- TODO: ADD IN HERE ALL THE RESOURCES ONE SHOULD ACCESS --
-----------------------------------------------------------
-- self group: display user info
call auth.add_resource(ARRAY['reader','editor','admin','root']::text[],'EQUALS','/user/whoami','self');
call auth.add_resource(ARRAY['reader','editor','admin','root']::text[],'EQUALS','/user/password','self');
-- management group: create, delete or manage access for user
call auth.add_resource(ARRAY['admin','root']::text[],'EQUALS','/admin/user/create','management');
call auth.add_resource(ARRAY['admin','root']::text[],'MATCHES','/admin/user/roles/*','management');
call auth.add_resource(ARRAY['root']::text[],'MATCHES','/root/user/delete/*','management');
--------------------------------------------------------
