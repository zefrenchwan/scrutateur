-- auth.upsert_user_auth upserts an user with that auth
create or replace procedure auth.upsert_user_auth(p_login text, p_password text) language plpgsql as $$
begin 
    insert into auth.users(user_login, user_hash_password) values (p_login,sha256(p_password::bytea)) 
    on conflict (user_login) do update set user_hash_password = sha256(p_password::bytea);
end;$$;

-- auth.delete_user deletes an user per login / username
create or replace procedure auth.delete_user(p_login text) language plpgsql as $$
declare 
    l_user_id int;
begin 
    select user_id into l_user_id from auth.users where user_login = p_login;
    if l_user_id is not null then 
        delete from auth.grants where user_id = l_user_id;
        delete from auth.users where user_id = l_user_id; 
    end if;
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


-- auth.add_resource adds a resource in a feature and expicits roles to access it
create or replace procedure auth.add_resource(p_roles text[], p_operator text, p_template text, p_feature text) language plpgsql as $$
declare 
    l_role text; 
    l_role_id int = -1;
    l_resource_id int;
begin 

    insert into auth.resources(operator,template_url,feature_name) values (p_operator, p_template, p_feature) returning resource_id into l_resource_id;

    foreach l_role in array p_roles loop 
        select role_id into l_role_id from auth.roles where role_name = l_role;
        if l_role_id is null or l_role_id = -1 then 
            raise exception 'no matching role for %', l_role;
        end if;

        insert into auth.authorizations(resource_id, role_id) values (l_resource_id, l_role_id);
    end loop;
end;$$;

-- auth.grant_feature_access sets role for that feature and user. 
-- NOTE THAT: it does not append, it sets. Previous grant values are deleted
create or replace procedure auth.grant_feature_access(p_user text, p_roles text[], p_feature text) language plpgsql as $$
declare 
    l_user_id int = -1;
    l_role_id int = -1;
    l_role text;
begin 

    select user_id into l_user_id  from auth.users where user_login = p_user;
    if l_user_id is null or l_user_id < 0 then 
        raise exception 'no user found with login %', p_user;
    end if;

    delete from auth.grants where user_id = l_user_id and feature_name = p_feature;

    foreach l_role in array p_roles loop 
        select role_id into l_role_id from auth.roles where role_name = l_role;
        if l_role_id is null or l_role_id = -1 then 
            raise exception 'no matching role for %', l_role;
        end if;

        insert into auth.grants(user_id, role_id, feature_name) values (l_user_id, l_role_id, p_feature);
    end loop;

end;$$;

-- auth.remove_feature_access_to_user removes all access to a feature for that user
create or replace procedure auth.remove_feature_access_to_user(p_user text, p_feature text) language plpgsql as $$
declare 
    l_user_id int;
begin 

    select user_id into l_user_id from auth.users where user_login = p_user;

    if l_user_id is not null and l_user_id > 0 then 
        delete from auth.grants where user_id = l_user_id and feature_name = p_feature;
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

-- Given an user (per login), gets all features and grant access for that users
create or replace function auth.get_roles_features_for_user(p_user text) returns table(feature_name text, roles text[]) language plpgsql as $$
begin
    return query 
        select GRA.feature_name, array_agg(distinct ROL.role_name::text) as roles
        from auth.users USR 
        join auth.grants GRA on GRA.user_id = USR.user_id 
        join auth.roles ROL on ROL.role_id = GRA.role_id
        where USR.user_login = p_user
        group by GRA.feature_name;
end;$$;