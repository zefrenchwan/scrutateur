-- auth.upsert_user_auth upserts an user with that auth
create or replace procedure auth.upsert_user_auth(p_login text, p_password text) language plpgsql as $$
begin 
    insert into auth.users(user_login, user_hash_password) values (p_login,sha256(p_password::bytea)) 
    on conflict (user_login) do update set user_hash_password = sha256(p_password::bytea);
end;$$;


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

-- auth.grant_group_access_to_user sets role for that group of resources for that user
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

-- auth.remove_access_from_group_to_user removes all access to a group of resources for that user
create or replace auth.remove_access_from_group_to_user(p_user text, p_group text) language plpgpsl as $$
declare 
    l_user_id int;
begin 

    select user_id from auth.users where user_login = p_user;

    if user_id is not null and user_id > 0 then 
        delete from auth.grants where user_id = l_user_id and group_name = p_group;
    end if;

end;$$;


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
