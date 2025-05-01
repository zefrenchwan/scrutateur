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

create table auth.roles (
    role_id serial primary key, 
    role_name text not null unique,
    role_description text not null,
    created_at timestamp with time zone not null default now(),
    updated_at  timestamp with time zone not null default now()
);

create or replace procedure auth.insert_role(p_role_name text, p_role_description text) language plpgsql as $$
declare 
begin 
    insert into auth.roles(role_name, role_description) values(p_role_name, p_role_description)
    on conflict (role_name) do update set role_description = p_role_description, updated_at = now();
end;$$;

-- link user to role 
create table auth.user_role (
    role_id int not null references auth.roles(role_id),
    user_id bigint not null references auth.users(user_id)
);

-- grant user (by login) to a role (by name)
create or replace procedure grant_user_to_role(p_login text, p_role text) language plpgsql as $$
declare 
    l_user_id bigint;
    l_role_id int;
begin 
    select user_id into l_user_id from auth.users where user_login = p_login and deleted_at is not null;
    select role_id into l_role_id from auth.roles where role_name = p_role and deleted_at is not null;  

    if l_user_id is null then 
        raise exception 'Cannot grant user because user % does not exist', p_login;
    end if;
    if l_role_id is null then 
        raise exception 'Cannot grant user because role % does not exist', p_role;
    end if;

    if not exists (
        select 1 from auth.user_role where user_id = l_user_id and role_id = l_role_id  
    ) then 
        insert into auth.user_role(user_id, role_id) values (l_user_id, l_role_id);
    end if;
end;$$;