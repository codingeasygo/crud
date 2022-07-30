--
-- PostgreSQL database dump
--



ALTER TABLE IF EXISTS crud_object ALTER COLUMN tid DROP DEFAULT;
DROP SEQUENCE IF EXISTS crud_simple_tid_seq;
DROP TABLE IF EXISTS crud_object;


--
-- Name: crud_object; Type: TABLE; Schema: public;
--

CREATE TABLE crud_object (
    tid bigint NOT NULL,
    user_id bigint DEFAULT 0 NOT NULL,
    type character varying(255) DEFAULT ''::character varying NOT NULL,
    level integer DEFAULT 0 NOT NULL,
    title character varying(255) NOT NULL,
    image character varying(1024),
    description text,
    data jsonb DEFAULT '[]'::jsonb NOT NULL,
    int_value integer DEFAULT 0 NOT NULL,
    int_ptr integer,
    int_array jsonb DEFAULT '[]'::jsonb NOT NULL,
    int64_value bigint DEFAULT 0 NOT NULL,
    int64_ptr bigint,
    int64_array jsonb DEFAULT '[]'::jsonb NOT NULL,
    float64_value double precision DEFAULT 0 NOT NULL,
    float64_ptr double precision,
    float64_array jsonb DEFAULT '[]'::jsonb NOT NULL,
    string_value character varying(255) DEFAULT ''::character varying NOT NULL,
    string_ptr character varying(255),
    string_array jsonb DEFAULT '[]'::jsonb NOT NULL,
    map_value jsonb DEFAULT '{}'::jsonb NOT NULL,
    map_array jsonb DEFAULT '[]'::jsonb NOT NULL,
    time_value timestamp with time zone NOT NULL,
    update_time timestamp with time zone NOT NULL,
    create_time timestamp with time zone NOT NULL,
    status integer NOT NULL
);


--
-- Name: COLUMN crud_object.type; Type: COMMENT; Schema: public;
--

COMMENT ON COLUMN crud_object.type IS 'simple type in, A=1:test a, B=2:test b, C=3:test c';


--
-- Name: COLUMN crud_object.status; Type: COMMENT; Schema: public;
--

COMMENT ON COLUMN crud_object.status IS 'simple status in, Normal=100, Disabled=200, Removed=-1';


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

ALTER SEQUENCE crud_simple_tid_seq OWNED BY crud_object.tid;


--
-- Name: crud_object tid; Type: DEFAULT; Schema: public;
--

ALTER TABLE IF EXISTS ONLY crud_object ALTER COLUMN tid SET DEFAULT nextval('crud_simple_tid_seq'::regclass);


--
-- Name: crud_object crud_simple_pkey; Type: CONSTRAINT; Schema: public;
--

ALTER TABLE IF EXISTS ONLY crud_object
    ADD CONSTRAINT crud_simple_pkey PRIMARY KEY (tid);


--
-- PostgreSQL database dump complete
--

