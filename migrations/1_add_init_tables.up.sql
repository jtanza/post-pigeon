create table post (
    id          integer primary key asc,
    uuid        text unique not null,
    fingerprint text not null,
    key         text not null,
    created_at  datetime,
    updated_at  datetime,
    deleted_at  datetime
);
create index post_uuid_idx on post(uuid);
create index post_fingerprint_idx on post(fingerprint);

create table post_content (
  id         integer primary key asc,
  post_uuid  integer not null,
  title      text not null,
  html       text not null,
  message    text not null,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  foreign key(post_uuid) references post(uuid)
);
create index post_content_uuid_idx on post_content(post_uuid);

