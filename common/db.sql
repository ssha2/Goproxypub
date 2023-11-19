-- Table: public.logging

-- DROP TABLE IF EXISTS public.logging;

CREATE TABLE IF NOT EXISTS public.logging
(
    ltype varchar(50) COLLATE pg_catalog."default" NOT NULL,
    sid varchar(50) COLLATE pg_catalog."default",
    head varchar(500) COLLATE pg_catalog."default",
    body text ,
    t timestamp,
    stamp timestamp with time zone DEFAULT  CURRENT_TIMESTAMP
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.logging
    OWNER to gen_user;
