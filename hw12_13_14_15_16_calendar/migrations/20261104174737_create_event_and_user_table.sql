-- +goose Up
create table calendar_user (
    id serial primary key,
    name text
);
comment ON table calendar_user IS 'пользователи';

insert into calendar_user(id, name) values (1, 'main');

create table calendar_event (
    id uuid primary key,
    title varchar(100),
    on_date timestamptz not null,
    duration interval not null,
    description text,
    user_id int not null,
    notificate_before interval,

    constraint event_user_id_fk FOREIGN KEY (user_id) references calendar_user(id) on delete cascade on update cascade
);
create index calendar_event_user_idx on calendar_event(user_id);
create index calendar_event_on_date_idx on calendar_event(on_date);
comment on table calendar_event is 'события';
comment ON column calendar_event.on_date is 'дата и время события';
comment ON column calendar_event.duration is 'продолжительность события';
comment ON column calendar_event.description is 'описание';
comment ON column calendar_event.user_id is 'id владельца события';
comment ON column calendar_event.notificate_before is 'за сколько уведомлять';

comment ON table calendar_event IS 'события';

-- +goose Down
drop table calendar_event;
drop table calendar_user;
