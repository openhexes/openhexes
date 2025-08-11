-- name: ListAccounts :many
select * from accounts 
where (active = sqlc.narg('active') or sqlc.narg('active') is null)
order by email;

-- name: GetAccountByID :one
select * from accounts where id = @id;

-- name: GetAccountByEmail :one
select * from accounts where email = @email;

-- name: CreateAccount :one
insert into accounts (active, created_at, email, display_name, picture)
values (@active, now(), @email, @display_name, @picture)
returning *;

-- name: CreateRole :exec
insert into roles (id)
values (@id)
on conflict do nothing;

-- name: ListRoles :many
select * from roles order by id;

-- name: GrantRole :exec
insert into role_bindings (role_id, account_id)
values (@role_id, @account_id)
on conflict do nothing;

-- name: RevokeRole :exec
delete from role_bindings
where role_id = @role_id and account_id = @account_id;

-- name: ListGrants :many
select role_id from role_bindings 
where 
    account_id = @account_id
    and (role_id = any(@role_ids::varchar[]) or coalesce(cardinality(@role_ids), 0) = 0)
order by role_id;

-- name: UpdateAccountActivation :exec
update accounts set active = @active where id = any(@ids::uuid[]);
