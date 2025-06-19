create extension if not exists pgcrypto;

-- auth for users and roles
create schema auth;

-------------------------------------------------------
-- useful functions, to include in tables definition --
-------------------------------------------------------
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
-----------------------------
-- end of useful functions --
-----------------------------

-- auth.users contain user information
create table auth.users (
    user_id serial primary key,
    user_login text unique not null,
    user_hash_password bytea not null
);

-- auth.roles define role name and a description
create table auth.roles (
    role_id serial primary key, 
    role_name text unique not null, 
    role_description text not null
);

-- auth.resources are the resources of the application. 
-- They form a group of resources to that one may grant ALL resources in a group at once (simpler, faster)
create table auth.resources (
    resource_id serial primary key, 
    operator text not null default 'EQUALS' check(operator = ANY('{EQUALS,STARTS_WITH,CONTAINS,MATCHES}'::text[])),
    template_url text not null,
    feature_name text not null
);

-- given a resource, get the role it needs to access it
create table auth.authorizations (
    resource_id int not null references auth.resources(resource_id),
    role_id int not null references auth.roles(role_id)
);

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
select RES.operator, RES.template_url, RES.feature_name, RAA.needed_roles
from auth.resources RES
join resources_agg_auth RAA on RAA.resource_id = RES.resource_id;


-- grant a role for a user on a feature
create table auth.grants (
    user_id int not null references auth.users(user_id),
    role_id int not null references auth.roles(role_id),
    feature_name text not null
);

-- auth.v_granted_resources gets login of user, resource operator, template and then roles the user has on this resource 
create view auth.v_granted_resources as
with granted_roles as (
    select USR.user_id, GRA.feature_name, array_agg(distinct ROL.role_name::text) as user_roles
    from auth.users USR 
    join auth.grants GRA on GRA.user_id = USR.user_id 
    join auth.roles ROL on ROL.role_id = GRA.role_id  
    group by USR.user_id, GRA.feature_name
), resources_auths as (
    select AUT.resource_id, RES.feature_name, array_agg(distinct ROL.role_name::text) as expected_roles
    from auth.authorizations AUT 
    join auth.resources RES on RES.resource_id = AUT.resource_id
    join auth.roles ROL on ROL.role_id = AUT.role_id 
    group by AUT.resource_id, RES.feature_name
)
select distinct USR.user_login, RES.operator, RES.template_url, auth.array_intersection(GRO.user_roles, RAU.expected_roles) as roles
from auth.users USR 
join granted_roles GRO on GRO.user_id = USR.user_id 
join resources_auths RAU on RAU.feature_name = GRO.feature_name 
join auth.resources RES on RES.resource_id = RAU.resource_id
where GRO.user_roles && RAU.expected_roles;

---------------------------------------------------------
-- ROLES DEFINITION: needs code refactoring if changed --
---------------------------------------------------------

-- add roles (constants, needs code refactoring for a change)
insert into auth.roles(role_name, role_description) values ('root','allows any action with grant actions too');
insert into auth.roles(role_name, role_description) values ('admin','allows any action but cannot grant');
insert into auth.roles(role_name, role_description) values ('editor','crud operations are allowed');
insert into auth.roles(role_name, role_description) values ('reader','read only operations are allowed');