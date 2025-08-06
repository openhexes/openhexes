-- Create "accounts" table
CREATE TABLE "public"."accounts" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "active" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL, "email" character varying(256) NOT NULL, "display_name" character varying(256) NOT NULL, "picture" character varying(256) NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "accounts_email_key" UNIQUE ("email"));
