create schema auth;

create table auth.users (
    user_id bigserial primary key,
    user_login text unique not null, 
    user_hash_auth text not null
);
