SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: file; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.file (
    id_file integer NOT NULL,
    id_message integer NOT NULL,
    title character varying(255),
    name character varying(255),
    ext character varying(50),
    link text
);


--
-- Name: file_id_file_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.file_id_file_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: file_id_file_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.file_id_file_seq OWNED BY public.file.id_file;


--
-- Name: kind_worker; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kind_worker (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    config_schema jsonb,
    config jsonb
);


--
-- Name: kind_worker_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.kind_worker_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: kind_worker_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.kind_worker_id_seq OWNED BY public.kind_worker.id;


--
-- Name: kind_worker_system; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kind_worker_system (
    id_system integer NOT NULL,
    id_kind_worker integer NOT NULL
);


--
-- Name: message; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message (
    id integer NOT NULL,
    id_worker integer,
    id_type_worker integer NOT NULL,
    id_system integer NOT NULL,
    id_priority integer NOT NULL,
    uuid uuid NOT NULL,
    value jsonb DEFAULT '{}'::jsonb NOT NULL,
    send_later timestamp without time zone,
    create_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: message_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.message_id_seq OWNED BY public.message.id;


--
-- Name: permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission (
    id integer NOT NULL,
    description text,
    slug character varying(255) NOT NULL
);


--
-- Name: permission_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permission_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permission_id_seq OWNED BY public.permission.id;


--
-- Name: permission_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission_role (
    id_permission integer NOT NULL,
    id_role integer NOT NULL
);


--
-- Name: pipline; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pipline (
    id integer NOT NULL,
    id_pipline_status integer NOT NULL,
    id_message integer NOT NULL,
    step integer NOT NULL,
    name character varying(255) NOT NULL,
    time_start timestamp without time zone,
    time_end timestamp without time zone
);


--
-- Name: pipline_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pipline_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pipline_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pipline_id_seq OWNED BY public.pipline.id;


--
-- Name: pipline_status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pipline_status (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: pipline_status_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pipline_status_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pipline_status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pipline_status_id_seq OWNED BY public.pipline_status.id;


--
-- Name: priority; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.priority (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    weight integer NOT NULL,
    slug character varying(255) NOT NULL,
    CONSTRAINT priority_weight_check CHECK ((weight >= 0))
);


--
-- Name: priority_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.priority_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: priority_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.priority_id_seq OWNED BY public.priority.id;


--
-- Name: role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL
);


--
-- Name: role_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.role_id_seq OWNED BY public.role.id;


--
-- Name: role_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_user (
    id_role integer NOT NULL,
    id_user integer NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(128) NOT NULL
);


--
-- Name: system; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system (
    id integer NOT NULL,
    create_user integer,
    id_priority integer NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: system_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_id_seq OWNED BY public.system.id;


--
-- Name: token; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.token (
    id_system integer NOT NULL,
    id_kind_worker integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    public_token character varying(255) NOT NULL,
    secret_token character varying(255) NOT NULL
);


--
-- Name: type_worker; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.type_worker (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL
);


--
-- Name: type_worker_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.type_worker_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: type_worker_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.type_worker_id_seq OWNED BY public.type_worker.id;


--
-- Name: user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    fio character varying(255) NOT NULL,
    login character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    reset_password_after_login boolean DEFAULT false,
    create_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;


--
-- Name: worker; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.worker (
    id integer NOT NULL,
    is_active boolean DEFAULT false NOT NULL,
    id_type_worker integer NOT NULL,
    id_kind_worker integer NOT NULL
);


--
-- Name: worker_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.worker_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: worker_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.worker_id_seq OWNED BY public.worker.id;


--
-- Name: file id_file; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file ALTER COLUMN id_file SET DEFAULT nextval('public.file_id_file_seq'::regclass);


--
-- Name: kind_worker id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker ALTER COLUMN id SET DEFAULT nextval('public.kind_worker_id_seq'::regclass);


--
-- Name: message id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message ALTER COLUMN id SET DEFAULT nextval('public.message_id_seq'::regclass);


--
-- Name: permission id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission ALTER COLUMN id SET DEFAULT nextval('public.permission_id_seq'::regclass);


--
-- Name: pipline id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline ALTER COLUMN id SET DEFAULT nextval('public.pipline_id_seq'::regclass);


--
-- Name: pipline_status id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline_status ALTER COLUMN id SET DEFAULT nextval('public.pipline_status_id_seq'::regclass);


--
-- Name: priority id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.priority ALTER COLUMN id SET DEFAULT nextval('public.priority_id_seq'::regclass);


--
-- Name: role id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role ALTER COLUMN id SET DEFAULT nextval('public.role_id_seq'::regclass);


--
-- Name: system id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system ALTER COLUMN id SET DEFAULT nextval('public.system_id_seq'::regclass);


--
-- Name: type_worker id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.type_worker ALTER COLUMN id SET DEFAULT nextval('public.type_worker_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);


--
-- Name: worker id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker ALTER COLUMN id SET DEFAULT nextval('public.worker_id_seq'::regclass);


--
-- Name: file file_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_pkey PRIMARY KEY (id_file);


--
-- Name: kind_worker kind_worker_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker
    ADD CONSTRAINT kind_worker_pkey PRIMARY KEY (id);


--
-- Name: kind_worker kind_worker_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker
    ADD CONSTRAINT kind_worker_slug_key UNIQUE (slug);


--
-- Name: kind_worker_system kind_worker_system_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker_system
    ADD CONSTRAINT kind_worker_system_pkey PRIMARY KEY (id_system, id_kind_worker);


--
-- Name: message message_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_pkey PRIMARY KEY (id);


--
-- Name: message message_uuid_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_uuid_key UNIQUE (uuid);


--
-- Name: permission permission_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT permission_pkey PRIMARY KEY (id);


--
-- Name: permission_role permission_role_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_role
    ADD CONSTRAINT permission_role_pkey PRIMARY KEY (id_permission, id_role);


--
-- Name: permission permission_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT permission_slug_key UNIQUE (slug);


--
-- Name: pipline pipline_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline
    ADD CONSTRAINT pipline_pkey PRIMARY KEY (id);


--
-- Name: pipline_status pipline_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline_status
    ADD CONSTRAINT pipline_status_pkey PRIMARY KEY (id);


--
-- Name: pipline pipline_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline
    ADD CONSTRAINT pipline_unique UNIQUE (step, id_message);


--
-- Name: priority priority_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.priority
    ADD CONSTRAINT priority_name_key UNIQUE (name);


--
-- Name: priority priority_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.priority
    ADD CONSTRAINT priority_pkey PRIMARY KEY (id);


--
-- Name: priority priority_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.priority
    ADD CONSTRAINT priority_slug_key UNIQUE (slug);


--
-- Name: priority priority_weight_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.priority
    ADD CONSTRAINT priority_weight_key UNIQUE (weight);


--
-- Name: role role_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_pkey PRIMARY KEY (id);


--
-- Name: role role_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_slug_key UNIQUE (slug);


--
-- Name: role_user role_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_pkey PRIMARY KEY (id_role, id_user);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: system system_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system
    ADD CONSTRAINT system_pkey PRIMARY KEY (id);


--
-- Name: token token_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.token
    ADD CONSTRAINT token_pkey PRIMARY KEY (id_system, id_kind_worker);


--
-- Name: token token_public_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.token
    ADD CONSTRAINT token_public_token_key UNIQUE (public_token);


--
-- Name: type_worker type_worker_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.type_worker
    ADD CONSTRAINT type_worker_pkey PRIMARY KEY (id);


--
-- Name: type_worker type_worker_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.type_worker
    ADD CONSTRAINT type_worker_slug_key UNIQUE (slug);


--
-- Name: user user_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_email_key UNIQUE (email);


--
-- Name: user user_login_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_login_key UNIQUE (login);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: worker worker_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker
    ADD CONSTRAINT worker_pkey PRIMARY KEY (id);


--
-- Name: file file_id_message_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file
    ADD CONSTRAINT file_id_message_fkey FOREIGN KEY (id_message) REFERENCES public.message(id) ON DELETE SET NULL;


--
-- Name: kind_worker_system kind_worker_system_id_kind_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker_system
    ADD CONSTRAINT kind_worker_system_id_kind_worker_fkey FOREIGN KEY (id_kind_worker) REFERENCES public.kind_worker(id) ON DELETE CASCADE;


--
-- Name: kind_worker_system kind_worker_system_id_system_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kind_worker_system
    ADD CONSTRAINT kind_worker_system_id_system_fkey FOREIGN KEY (id_system) REFERENCES public.system(id) ON DELETE CASCADE;


--
-- Name: message message_id_priority_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_id_priority_fkey FOREIGN KEY (id_priority) REFERENCES public.priority(id) ON DELETE CASCADE;


--
-- Name: message message_id_system_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_id_system_fkey FOREIGN KEY (id_system) REFERENCES public.system(id) ON DELETE CASCADE;


--
-- Name: message message_id_type_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_id_type_worker_fkey FOREIGN KEY (id_type_worker) REFERENCES public.type_worker(id) ON DELETE CASCADE;


--
-- Name: message message_id_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_id_worker_fkey FOREIGN KEY (id_worker) REFERENCES public.worker(id) ON DELETE SET NULL;


--
-- Name: permission_role permission_role_id_permission_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_role
    ADD CONSTRAINT permission_role_id_permission_fkey FOREIGN KEY (id_permission) REFERENCES public.permission(id) ON DELETE CASCADE;


--
-- Name: permission_role permission_role_id_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_role
    ADD CONSTRAINT permission_role_id_role_fkey FOREIGN KEY (id_role) REFERENCES public.role(id) ON DELETE CASCADE;


--
-- Name: pipline pipline_id_message_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline
    ADD CONSTRAINT pipline_id_message_fkey FOREIGN KEY (id_message) REFERENCES public.message(id) ON DELETE CASCADE;


--
-- Name: pipline pipline_id_pipline_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pipline
    ADD CONSTRAINT pipline_id_pipline_status_fkey FOREIGN KEY (id_pipline_status) REFERENCES public.pipline_status(id) ON DELETE CASCADE;


--
-- Name: role_user role_user_id_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_id_role_fkey FOREIGN KEY (id_role) REFERENCES public.role(id) ON DELETE CASCADE;


--
-- Name: role_user role_user_id_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_user
    ADD CONSTRAINT role_user_id_user_fkey FOREIGN KEY (id_user) REFERENCES public."user"(id) ON DELETE CASCADE;


--
-- Name: system system_create_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system
    ADD CONSTRAINT system_create_user_fkey FOREIGN KEY (create_user) REFERENCES public."user"(id) ON DELETE SET NULL;


--
-- Name: system system_id_priority_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system
    ADD CONSTRAINT system_id_priority_fkey FOREIGN KEY (id_priority) REFERENCES public.priority(id) ON DELETE SET NULL;


--
-- Name: token token_id_system_id_kind_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.token
    ADD CONSTRAINT token_id_system_id_kind_worker_fkey FOREIGN KEY (id_system, id_kind_worker) REFERENCES public.kind_worker_system(id_system, id_kind_worker) ON DELETE CASCADE;


--
-- Name: worker worker_id_kind_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker
    ADD CONSTRAINT worker_id_kind_worker_fkey FOREIGN KEY (id_kind_worker) REFERENCES public.kind_worker(id) ON DELETE SET NULL;


--
-- Name: worker worker_id_type_worker_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.worker
    ADD CONSTRAINT worker_id_type_worker_fkey FOREIGN KEY (id_type_worker) REFERENCES public.type_worker(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20241124130311'),
    ('20241124130729'),
    ('20241124130842'),
    ('20241124130953'),
    ('20241124131041'),
    ('20241124131213'),
    ('20241124131214'),
    ('20241124131315'),
    ('20241124131400'),
    ('20241124131404'),
    ('20241124131514'),
    ('20241124131553'),
    ('20241124131555'),
    ('20241124132131'),
    ('20241124133111'),
    ('20241215200944');
