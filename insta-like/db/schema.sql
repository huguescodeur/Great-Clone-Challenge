-- insta-like database schema
-- PostgreSQL 18
-- Run this file against a fresh database to get started:
--   psql -h localhost -U postgres -d insta -f db/schema.sql

-- Extensions

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;

-- Types

CREATE TYPE public.follow_status AS ENUM (
    'pending',
    'accepted'
);

-- Tables

CREATE TABLE public.users (
    user_id          uuid                        NOT NULL DEFAULT gen_random_uuid(),
    username         public.citext               NOT NULL,
    email            public.citext               NOT NULL,
    password_hash    character varying(255)      NOT NULL,
    full_name        character varying(255)      NOT NULL,
    bio              text,
    profile_pic_url  character varying(500),
    followers_count  bigint                      NOT NULL DEFAULT 0,
    following_count  bigint                      NOT NULL DEFAULT 0,
    posts_count      bigint                      NOT NULL DEFAULT 0,
    created_at       timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT users_pkey            PRIMARY KEY (user_id),
    CONSTRAINT users_email_key       UNIQUE (email),
    CONSTRAINT users_username_key    UNIQUE (username),
    CONSTRAINT users_bio_check       CHECK (length(bio) <= 500),
    CONSTRAINT users_followers_count_check CHECK (followers_count >= 0),
    CONSTRAINT users_following_count_check CHECK (following_count >= 0),
    CONSTRAINT users_posts_count_check     CHECK (posts_count >= 0)
);

CREATE TABLE public.posts (
    post_id         uuid                     NOT NULL DEFAULT gen_random_uuid(),
    user_id         uuid                     NOT NULL,
    content         text                     NOT NULL,
    likes_count     bigint                   NOT NULL DEFAULT 0,
    comments_count  bigint                   NOT NULL DEFAULT 0,
    created_at      timestamp with time zone NOT NULL DEFAULT now(),
    updated_at      timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at      timestamp with time zone,

    CONSTRAINT posts_pkey         PRIMARY KEY (post_id),
    CONSTRAINT posts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT posts_content_check CHECK (length(content) <= 5000)
);

CREATE TABLE public.post_media (
    media_id    uuid                     NOT NULL DEFAULT gen_random_uuid(),
    post_id     uuid                     NOT NULL,
    url         text                     NOT NULL,
    type        text                     NOT NULL,
    created_at  timestamp with time zone NOT NULL DEFAULT now(),

    CONSTRAINT post_media_pkey         PRIMARY KEY (media_id),
    CONSTRAINT post_media_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(post_id) ON DELETE CASCADE,
    CONSTRAINT post_media_type_check   CHECK (type = ANY (ARRAY['image'::text, 'video'::text]))
);

CREATE TABLE public.likes (
    post_id        uuid                     NOT NULL,
    user_id        uuid                     NOT NULL,
    reaction_type  character varying(20)    NOT NULL DEFAULT 'like',
    created_at     timestamp with time zone NOT NULL DEFAULT now(),
    updated_at     timestamp with time zone          DEFAULT now(),

    CONSTRAINT likes_pkey          PRIMARY KEY (post_id, user_id),
    CONSTRAINT likes_post_id_fkey  FOREIGN KEY (post_id) REFERENCES public.posts(post_id) ON DELETE CASCADE,
    CONSTRAINT likes_user_id_fkey  FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE
);

CREATE TABLE public.comments (
    comment_id        uuid                        NOT NULL DEFAULT gen_random_uuid(),
    post_id           uuid                        NOT NULL,
    user_id           uuid                        NOT NULL,
    parent_comment_id uuid,
    content           text                        NOT NULL,
    likes_count       bigint                      NOT NULL DEFAULT 0,
    created_at        timestamp without time zone NOT NULL DEFAULT now(),
    updated_at        timestamp without time zone NOT NULL DEFAULT now(),
    deleted_at        timestamp without time zone,

    CONSTRAINT comments_pkey        PRIMARY KEY (comment_id),
    CONSTRAINT fk_comments_post     FOREIGN KEY (post_id)           REFERENCES public.posts(post_id)    ON DELETE CASCADE,
    CONSTRAINT fk_comments_user     FOREIGN KEY (user_id)           REFERENCES public.users(user_id),
    CONSTRAINT fk_comments_parent   FOREIGN KEY (parent_comment_id) REFERENCES public.comments(comment_id) ON DELETE CASCADE,
    CONSTRAINT comments_content_check CHECK (char_length(content) <= 2000)
);

CREATE TABLE public.comment_likes (
    comment_id     uuid                        NOT NULL,
    user_id        uuid                        NOT NULL,
    reaction_type  character varying(20)       NOT NULL,
    created_at     timestamp without time zone NOT NULL DEFAULT now(),
    updated_at     timestamp without time zone NOT NULL DEFAULT now(),

    CONSTRAINT comment_likes_pkey           PRIMARY KEY (comment_id, user_id),
    CONSTRAINT fk_comment_likes_comment     FOREIGN KEY (comment_id) REFERENCES public.comments(comment_id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_likes_user        FOREIGN KEY (user_id)    REFERENCES public.users(user_id)
);

CREATE TABLE public.followers (
    follower_id  uuid                        NOT NULL,
    followee_id  uuid                        NOT NULL,
    status       public.follow_status        NOT NULL DEFAULT 'accepted',
    created_at   timestamp without time zone NOT NULL DEFAULT now(),

    CONSTRAINT followers_pkey           PRIMARY KEY (follower_id, followee_id),
    CONSTRAINT fk_followers_follower    FOREIGN KEY (follower_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_followers_followee    FOREIGN KEY (followee_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT chk_no_self_follow       CHECK (follower_id <> followee_id)
);

-- Indexes

CREATE INDEX idx_posts_user_created_at  ON public.posts        USING btree (user_id, created_at DESC);
CREATE INDEX idx_posts_user_active      ON public.posts        USING btree (user_id, created_at DESC) WHERE (deleted_at IS NULL);

CREATE INDEX idx_post_media_post_id     ON public.post_media   USING btree (post_id);

CREATE INDEX idx_likes_user_created     ON public.likes        USING btree (user_id, created_at DESC);

CREATE INDEX idx_comments_post_id           ON public.comments USING btree (post_id);
CREATE INDEX idx_comments_parent_comment_id ON public.comments USING btree (parent_comment_id);
CREATE INDEX idx_comments_created_at        ON public.comments USING btree (created_at);
CREATE INDEX idx_comments_deleted_at        ON public.comments USING btree (deleted_at);

CREATE INDEX idx_comment_likes_user_id  ON public.comment_likes USING btree (user_id);

CREATE INDEX idx_followers_followee_id  ON public.followers    USING btree (followee_id);
