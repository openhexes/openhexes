create extension if not exists "uuid-ossp";

create table accounts
(
    id              uuid default gen_random_uuid() primary key,
    active          bool default false not null,
    created_at      timestamptz not null,
    email           varchar(256) not null unique,
    display_name    varchar(256) not null,
    picture         varchar(256) not null
);

create table roles
(
    id  varchar(256) not null primary key
);

create table role_bindings
(
    account_id uuid references accounts (id) on delete cascade not null,
    role_id varchar(256) references roles (id) on delete cascade not null
);
