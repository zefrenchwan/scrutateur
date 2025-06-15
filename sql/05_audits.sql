create schema evt;

-- evt.actions log any important event from users
create table evt.actions (
    event_date timestamp with time zone default now(),
    event_initiator text not null, 
    event_type text not null,
    event_description text not null,
    event_parameters text[]
);

-- evt.log_action adds an event entry into the database
create or replace procedure evt.log_action(p_login text, p_type text, p_description text, p_params text[]) language plpgsql as $$
declare 
begin
    insert into evt.actions(event_initiator, event_type, event_description, event_parameters)
    values (p_login, p_type, p_description, p_params);
end;$$;