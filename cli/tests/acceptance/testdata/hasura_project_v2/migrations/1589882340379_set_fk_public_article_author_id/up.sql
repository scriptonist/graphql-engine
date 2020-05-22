alter table "public"."article"
           add constraint "article_author_id_fkey"
           foreign key ("author_id")
           references "public"."user"
           ("id") on update restrict on delete restrict;
