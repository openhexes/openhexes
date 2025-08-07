-- Create "accounts" table
CREATE TABLE "public"."accounts" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "active" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "email" character varying(256) NOT NULL, "display_name" character varying(256) NOT NULL, "picture" character varying(256) NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "accounts_email_key" UNIQUE ("email"));
-- Create "roles" table
CREATE TABLE "public"."roles" ("id" character varying(256) NOT NULL, PRIMARY KEY ("id"));
-- Create "role_bindings" table
CREATE TABLE "public"."role_bindings" ("account_id" uuid NOT NULL, "role_id" character varying(256) NOT NULL, CONSTRAINT "role_bindings_account_id_fkey" FOREIGN KEY ("account_id") REFERENCES "public"."accounts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "role_bindings_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
