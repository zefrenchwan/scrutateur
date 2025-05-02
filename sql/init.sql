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
