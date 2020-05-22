CREATE TABLE "public"."article"("id" serial NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "content" text NOT NULL, "author_id" integer NOT NULL, PRIMARY KEY ("id") );
