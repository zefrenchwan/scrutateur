create extension if not exists pgcrypto;

-- auth for users and roles
create schema auth;

-- useful functions 
create or replace function auth.array_intersection(p_a text[], p_b text[]) returns text[] language plpgsql as $$
declare 
    l_result text[];
    l_value text;
begin 
    select ARRAY[]::text[] into l_result;
    foreach l_value in array p_a loop 
        if l_value = ANY(p_b) and not (l_value = ANY(l_result)) then 
            l_result = array_prepend(l_value, l_result);
        end if;
    end loop;

    return l_result;
end;$$;

-- end of useful functions 

-- auth.users contain user information
create table auth.users (
    user_id serial primary key,
    user_login text unique not null,
    user_hash_password bytea not null
);

-- auth.upsert_user_auth upserts an user with that auth
create or replace procedure auth.upsert_user_auth(p_login text, p_password text) language plpgsql as $$
begin 
    insert into auth.users(user_login, user_hash_password) values (p_login,sha256(p_password::bytea)) 
    on conflict (user_login) do update set user_hash_password = sha256(p_password::bytea);
end;$$;

-- best way to get fired
call auth.upsert_user_auth('root','root');

-- auth.validate_auth returns true if user login matches password, false otherwise (no login, or wrong password)
create or replace function auth.validate_auth(p_user text, p_password text) returns bool language plpgsql as $$
declare 
    l_counter int;
    l_hash_password bytea;
begin 
    select sha256(p_password::bytea) into l_hash_password;

    select count(*) into l_counter 
    from auth.users 
    where user_login = p_user 
    and user_hash_password = l_hash_password;

    return l_counter = 1;
end; $$;

-- auth.roles define role name and a description
create table auth.roles (
    role_id serial primary key, 
    role_name text unique not null, 
    role_description text not null
);

insert into auth.roles(role_name, role_description) values ('root','allows any action with grant actions too');
insert into auth.roles(role_name, role_description) values ('admin','allows any action but cannot grant');
insert into auth.roles(role_name, role_description) values ('editor','crud operations are allowed');
insert into auth.roles(role_name, role_description) values ('reader','read only operations are allowed');


-- auth.resources are the resources of the application. 
-- They form a group of resources to that one may grant ALL resources in a group at once (simpler, faster)
create table auth.resources (
    resource_id serial primary key, 
    operator text not null default 'EQUALS' check(operator = ANY('{EQUALS,STARTS_WITH,CONTAINS,MATCHES}'::text[])),
    template_url text not null,
    group_name text not null
);

-- TODO: ADD IN HERE ALL THE RESOURCES ONE SHOULD ADD --
insert into auth.resources(operator, template_url, group_name) values ('EQUALS','/user/whoami', 'self');

--------------------------------------------------------



-- given a resource, get the role it needs to access it
create table auth.authorizations (
    resource_id int not null references auth.resources(resource_id),
    role_id int not null references auth.roles(role_id)
);

-- auth.add_resource adds a resource in a group and expicits roles to access it
create or replace procedure auth.add_resource(p_roles text[], p_operator text, p_template text, p_group text) language plpgsql as $$
declare 
    l_role text; 
    l_role_id int = -1;
    l_resource_id int;
begin 

    insert into auth.resources(operator,template_url,group_name) values (p_operator, p_template, p_group) returning resource_id into l_resource_id;

    foreach l_role in array p_roles loop 
        select role_id into l_role_id from auth.roles where role_name = l_role;
        if l_role_id is null or l_role_id = -1 then 
            raise exception 'no matching role for %', l_role;
        end if;

        insert into auth.authorizations(resource_id, role_id) values (l_resource_id, l_role_id);
    end loop;
end;$$;

-- any role may access /user/whoami
call auth.add_resource(ARRAY['reader','editor','admin','root']::text[],'EQUALS','/user/whoami','self');


-- auth.v_resources_authorizations displays, for a resource, its operator, template, group and needed roles
-- For instance, EQUALS url group {'reader','editor'}
create view auth.v_resources_authorizations as
with resources_agg_auth as (
    select RES.resource_id, array_agg(ROL.role_name) as needed_roles 
    from auth.resources RES 
    join auth.authorizations AUT on AUT.resource_id = RES.resource_id 
    join auth.roles ROL on ROL.role_id = AUT.role_id
    group by RES.resource_id
) 
select RES.operator, RES.template_url, RES.group_name, RAA.needed_roles
from auth.resources RES
join resources_agg_auth RAA on RAA.resource_id = RES.resource_id;


-- grant a role for a user on a group of resources
create table auth.grants (
    user_id int not null references auth.users(user_id),
    role_id int not null references auth.roles(role_id),
    group_name text not null
);

-- auth.v_granted_resources gets login of user, resource operator, template and then roles the user has on this resource 
create view auth.v_granted_resources as
with granted_roles as (
    select USR.user_id, GRA.group_name, array_agg(ROL.role_name::text) as user_roles
    from auth.users USR 
    join auth.grants GRA on GRA.user_id = USR.user_id 
    join auth.roles ROL on ROL.role_id = GRA.role_id  
    group by USR.user_id, GRA.group_name
), resources_auths as (
    select AUT.resource_id, RES.group_name, array_agg(distinct ROL.role_name::text) as expected_roles
    from auth.authorizations AUT 
    join auth.resources RES on RES.resource_id = AUT.resource_id
    join auth.roles ROL on ROL.role_id = AUT.role_id 
    group by AUT.resource_id, RES.group_name
)
select distinct USR.user_login, RES.operator, RES.template_url, auth.array_intersection(GRO.user_roles, RAU.expected_roles) as roles
from auth.users USR 
join granted_roles GRO on GRO.user_id = USR.user_id 
join resources_auths RAU on RAU.group_name = GRO.group_name 
join auth.resources RES on RES.resource_id = RAU.resource_id
where GRO.user_roles && RAU.expected_roles;

create or replace procedure auth.grant_group_access_to_user(p_user text, p_roles text[], p_group text) language plpgsql as $$
declare 
    l_user_id int = -1;
    l_role_id int = -1;
    l_role text;
begin 

    select user_id into l_user_id  from auth.users where user_login = p_user;
    if l_user_id is null or l_user_id < 0 then 
        raise exception 'no user found with login %', p_user;
    end if;

    delete from auth.grants where user_id = l_user_id and group_name = p_group;

    foreach l_role slice 1 in array p_roles loop 
        select role_id into l_role_id from auth.roles where role_name = l_role;
        if role_id is null or role_id = -1 then 
            raise exception 'no matching role for %', l_role;
        end if;

        insert into auth.grants(user_id, role_id, group_name) values (l_user_id, l_role_id, p_group);
    end loop;

end;$$;

-- grant root on any resource
insert into auth.grants(user_id, role_id, group_name) 
select distinct USR.user_id, RCO.role_id, RES.group_name  
from  auth.users USR 
cross join auth.roles RCO 
cross join auth.resources RES 
where USR.user_login = 'root';


-- given a user (per login), get all user's access for resources user could use:
-- role name (root, admin, etc) of the user for that pattern 
-- operator (for resources) the operator to apply to the template
-- template url for url (for instance /user/whoami)
-- roles for this resource as the common roles that user has and resource need
create or replace function auth.get_grants_for_user(p_user text) returns table(operator text, template_url text, roles text[]) language plpgsql as $$
declare 
begin 
    return query
        select distinct VGR.operator, VGR.template_url, VGR.roles
        from auth.v_granted_resources VGR
        where VGR.user_login = p_user ;
end;$$;

select * from auth.get_grants_for_user('root');