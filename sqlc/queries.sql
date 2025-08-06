-- name: ListAccounts :many
select * from accounts 
where (active = sqlc.narg('active') or sqlc.narg('active') is null)
order by id;

-- name: GetAccount :one
select * from accounts where email = @email;

-- name: CreateAccount :one
insert into accounts (created_at, email, display_name, picture)
values (now(), @email, @display_name, @picture)
returning *;
