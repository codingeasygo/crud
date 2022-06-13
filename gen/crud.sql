--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Debian 13.2-1.pgdg100+1)
-- Dumped by pg_dump version 13.2 (Debian 13.2-1.pgdg100+1)


ALTER TABLE IF EXISTS crud_simple ALTER COLUMN tid DROP DEFAULT;
DROP SEQUENCE IF EXISTS crud_simple_tid_seq;
DROP TABLE IF EXISTS crud_simple;


--
-- Name: crud_simple; Type: TABLE; Schema: public;
--

CREATE TABLE crud_simple (
    tid bigint NOT NULL,
    user_id bigint DEFAULT 0 NOT NULL,
    type character varying(255) DEFAULT ''::character varying NOT NULL,
    title character varying(255),
    image character varying(255),
    description text,
    data jsonb DEFAULT '{}'::jsonb NOT NULL,
    update_time timestamp with time zone NOT NULL,
    create_time timestamp with time zone NOT NULL,
    status integer NOT NULL
);


--
-- Name: COLUMN crud_simple.type; Type: COMMENT; Schema: public;
--

COMMENT ON COLUMN crud_simple.type IS 'simple type in, A=1:test a, B=2:test b, C=3:test c';


--
-- Name: COLUMN crud_simple.status; Type: COMMENT; Schema: public;
--

COMMENT ON COLUMN crud_simple.status IS 'simple status in, Normal=100, Disabled=200, Removed=-1';


--
-- Name: crud_simple_tid_seq; Type: SEQUENCE; Schema: public;
--

CREATE SEQUENCE crud_simple_tid_seq
    START WITH 1000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: crud_simple_tid_seq; Type: SEQUENCE OWNED BY; Schema: public;
--

ALTER SEQUENCE crud_simple_tid_seq OWNED BY crud_simple.tid;


--
-- Name: crud_simple tid; Type: DEFAULT; Schema: public;
--

ALTER TABLE IF EXISTS ONLY crud_simple ALTER COLUMN tid SET DEFAULT nextval('crud_simple_tid_seq'::regclass);


--
-- Name: crud_simple crud_simple_pkey; Type: CONSTRAINT; Schema: public;
--

ALTER TABLE IF EXISTS ONLY crud_simple
    ADD CONSTRAINT crud_simple_pkey PRIMARY KEY (tid);


--
-- PostgreSQL database dump complete
--

