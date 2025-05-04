create extension if not exists pgcrypto;

-- auth for users and roles
create schema auth;


create table auth.users (
    user_id serial primary key,
    user_login text unique not null,
    user_hash_password bytea not null,
    created_at timestamp with time zone not null default now(),
    updated_at  timestamp with time zone not null default now()
);

-- auth.upsert_user_auth upserts an user with that auth
create or replace procedure auth.upsert_user_auth(p_login text, p_password text) language plpgsql as $$
begin 
    insert into auth.users(user_login, user_hash_password) values (p_login,sha256(p_password::bytea)) 
    on conflict (user_login) do update set user_hash_password = sha256(p_password::bytea), updated_at = now();
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

-- auth.roles define role name, a description and a score. 
-- The higher the score, the more operations. 
-- It is useful to define access conditions instead of enumerating all possible roles
create table auth.roles (
    role_id serial primary key, 
    role_name text unique not null, 
    role_description text not null,
    role_score int not null check(role_score >= 0 and role_score <= 100)
);

-- define roles. Changing scores means auditing the full system again for security
insert into auth.roles(role_name, role_description, role_score) values ('super admin','allows any action with grant actions too',100);
insert into auth.roles(role_name, role_description, role_score) values ('admin','allows any action but cannot grant',50);
insert into auth.roles(role_name, role_description, role_score) values ('editor','crud operations are allowed',10);
insert into auth.roles(role_name, role_description, role_score) values ('reader','crud operations are allowed',1);

-- users are linked into roles and then form a group (the admins, etc)
create table auth.groups (
    user_id int not null references auth.users(user_id),
    role_id int not null references auth.roles(role_id)
);

-- grant root the highest scores
with role_content as (
    select role_id from auth.roles where role_score = 100
)
insert into auth.groups(user_id, role_id) 
select user_id, role_id  
from  auth.users USR 
cross join role_content
where USR.user_login = 'root';


-- Given a user login, gets related access groups
create or replace function auth.get_roles_for_user(p_user text) returns table(role_name text, role_score int) language plpgsql as $$
declare 
begin 
    return query 
        select ROL.role_name, ROL.role_score 
        from auth.roles ROL 
        join auth.groups GRO on GRO.role_id = ROL.role_id 
        join auth.users USR on USR.user_id = GRO.user_id 
        where USR.user_login = p_user;
end;$$