create table post (
    id    integer primary key asc,
    uuid  text unique not null,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime
);
create index post_uuid_idx on post(uuid);

create table post_content (
  id    integer primary key asc,
  post_uuid integer not null,
  title text not null,
  html text not null,
  message text not null,
  key text unique not null,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  foreign key(post_uuid) references post(uuid)
);

