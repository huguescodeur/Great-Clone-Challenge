-- Type

CREATE TYPE public.notification_type AS ENUM (
    'like_post',
    'comment_post',
    'like_comment',
    'new_follower'
);

-- Table

CREATE TABLE public.notifications (
    notification_id  uuid                        NOT NULL DEFAULT gen_random_uuid(),
    recipient_id     uuid                        NOT NULL,
    actor_id         uuid                        NOT NULL,
    type             public.notification_type    NOT NULL,
    entity_id        uuid,
    read             boolean                     NOT NULL DEFAULT false,
    created_at       timestamp with time zone    NOT NULL DEFAULT now(),

    CONSTRAINT notifications_pkey         PRIMARY KEY (notification_id),
    CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_notifications_actor     FOREIGN KEY (actor_id)     REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT chk_notif_not_self         CHECK (recipient_id <> actor_id)
);

-- Indexes

CREATE INDEX idx_notifications_recipient_created ON public.notifications USING btree (recipient_id, created_at DESC);
CREATE INDEX idx_notifications_recipient_unread  ON public.notifications USING btree (recipient_id) WHERE (read = false);
