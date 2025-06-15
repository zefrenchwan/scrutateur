create schema orgs;

-- orgs.groups are group of users
create table orgs.groups (
    group_id uuid primary key default gen_random_uuid(), 
    group_name text unique not null,
    created_at timestamp with time zone default now(),
    creator int references auth.users(user_id)
);

-- orgs.memberships contain users within a group
create table orgs.memberships (
    group_id uuid not null references orgs.groups(group_id) on delete cascade,
    user_id int not null references auth.users(user_id)  on delete cascade,
    granter_id int not null references auth.users(user_id),
    created_at timestamp with time zone default now(),
    granting bool default false,
    administrating bool default false, 
    inviting bool default true
);

-- orgs.v_group_and_member contains the users in groups 
create view orgs.v_group_and_member as 
select G.group_id, G.group_name, M.user_id, U.user_login, M.granting, M.administrating, M.inviting
from orgs.groups G 
join orgs.memberships M on M.group_id = G.group_id
join auth.users U on U.user_id = M.user_id;

-- orgs.get_groups_for_user returns the available groups for an user
create or replace function orgs.get_groups_for_user(p_user_login text) returns table(group_name text, granter bool, admin bool, inviter bool) language plpgsql as $$
declare 
begin 
    return query 
        select G.group_name, G.granting as granter, G.administrating as admin, G.inviting as inviter
        from orgs.v_group_and_member G
        where G.user_login = p_user_login;
end;$$;

-- orgs.add_group adds a group created by that user, with initial access rights for that user
create or replace procedure orgs.add_group(p_creator text, p_name text, p_grant bool, p_admin bool, p_invite bool) language plpgsql as $$
declare 
    l_user_id int;
    l_group_id uuid;
begin 
    select user_id into l_user_id from auth.users where user_login = p_creator;
    if l_user_id is null then 
        raise exception 'no user matching %', p_creator;
    end if;

    if exists (select 1 from orgs.groups where group_name ilike p_name) then 
        raise exception 'similar group to % already exists', p_name;
    end if;

    insert into orgs.groups(group_name, creator) values (p_name, l_user_id) returning group_id into l_group_id;
    insert into orgs.memberships(group_id, user_id, granter_id, granting, administrating, inviting) values (l_group_id, l_user_id, l_user_id, p_grant, p_admin, p_invite);

end;$$;

-- orgs.set_user_access_into_group upserts user access rights, granted by creator
create or replace procedure orgs.set_user_access_into_group(p_creator text, p_invited text, p_name text, p_grant bool, p_admin bool, p_invite bool) language plpgsql as $$
declare 
    l_creator_id int;
    l_user_id int;
    l_group_id uuid;
begin 
    select user_id into l_creator_id from auth.users where user_login = p_creator;
    if l_creator_id is null then 
        raise exception 'no creator matching %', p_creator;
    end if;
    select user_id into l_user_id from auth.users where user_login = p_invited;
    if l_user_id is null then 
        raise exception 'no user matching %', p_invited;
    end if;
    select group_id into l_group_id from orgs.groups where group_name = p_name; 
    if l_group_id is null then 
        raise exception 'group % does not exist', p_name;
    end if;

    delete from orgs.memberships where group_id = l_group_id and user_id = l_user_id; 

    insert into orgs.memberships(group_id, user_id, granter_id, granting, administrating, inviting) values (l_group_id, l_user_id, l_creator_id, p_grant, p_admin, p_invite);

end;$$;

-- orgs.delete_group just deletes a group of users
create or replace procedure orgs.delete_group(p_name text) language plpgsql as $$
declare 
    l_group_id uuid;
begin 
    select group_id into l_group_id from orgs.groups where group_name = p_name;
    delete from orgs.memberships where group_id = l_group_id;
    delete from orgs.groups where group_id =  l_group_id;
end;$$
