BEGIN;

CREATE TYPE season AS ENUM (
  'fall',
  'spring',
  'summer',
  'winter');

CREATE TYPE status AS ENUM (
  'Open',
  'Closed',
  'Cancelled'
);

CREATE TYPE period AS ENUM (
  'fall',
  'spring',
  'summer',
  'winter',
  'start_fall',
  'start_spring',
  'start_summer',
  'start_winter',
  'end_fall',
  'end_spring',
  'end_summer',
  'end_winter'
);

CREATE TABLE IF NOT EXISTS public.university
(
  id serial,
  name text NOT NULL,
  abbr text NOT NULL,
  home_page text NOT NULL,
  registration_page text NOT NULL,
  main_color text NOT NULL,
  accent_color text NOT NULL,
  topic_name text,
  topic_id text,
  created_at timestamp without time zone,
  updated_at timestamp without time zone,
  CONSTRAINT university__pk PRIMARY KEY (id),
  CONSTRAINT unique_university_name UNIQUE (name),
  CONSTRAINT unique_university_topic_name UNIQUE (topic_name),
  CONSTRAINT unique_university_topic_id UNIQUE (topic_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.university OWNER TO postgres;

COMMENT ON COLUMN public.university.abbr IS 'Abbreviation of the university name';
COMMENT ON COLUMN public.university.home_page IS 'The homepage page of the university';
COMMENT ON COLUMN public.university.registration_page IS 'The registration page  of the university';
COMMENT ON COLUMN public.university.main_color IS 'ARGB hex of the main color of the university';
COMMENT ON COLUMN public.university.accent_color IS 'ARGB hex of the accent color of the university';
COMMENT ON COLUMN public.university.topic_name IS 'The topic name of this university. Used to build topic url';

CREATE TABLE IF NOT EXISTS public.subject
(
  id serial,
  university_id BIGINT NOT NULL,
  name text NOT NULL,
  number text NOT NULL,
  season season NOT NULL,
  year text NOT NULL,
  topic_name text,
  topic_id text,
  data BYTEA,
  created_at timestamp without time zone,
  updated_at timestamp without time zone,
  CONSTRAINT subject__pk PRIMARY KEY (id),
  CONSTRAINT subject_university__fk FOREIGN KEY (university_id) REFERENCES public.university (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT unique_subject_name_number_year_season UNIQUE (university_id, name, number, year, season),
  CONSTRAINT unique_subject_topic_name UNIQUE (topic_name),
  CONSTRAINT unique_subject_topic_id UNIQUE (topic_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.subject OWNER TO postgres;

COMMENT ON COLUMN public.subject.university_id IS 'The university this subject belongs to';
COMMENT ON COLUMN public.subject.name IS 'The name of the subject';
COMMENT ON COLUMN public.subject.number IS 'The number of the subject.';
COMMENT ON COLUMN public.subject.season IS 'The season for which this subject is offered';
COMMENT ON COLUMN public.subject.year IS 'The year this subject is currently offered. Subjects are not guaranteed to be offered every year.';
COMMENT ON COLUMN public.subject.topic_name IS 'The topic name of this subject. Used to build topic url';
COMMENT ON COLUMN public.subject.updated_at IS 'Time this row was updated';
COMMENT ON TABLE public.subject  IS 'Contains the subject offered from a particular university';
/*
COMMENT ON CONSTRAINT subject_university__fk ON public.subject IS 'Foreign key for university';
*/

CREATE TABLE IF NOT EXISTS public.course
(
  id SERIAL,
  subject_id BIGINT,
  name TEXT NOT NULL,
  number TEXT NOT NULL,
  synopsis TEXT,
  topic_name TEXT NOT NULL,
  topic_id text,
  data BYTEA,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT course__pk PRIMARY KEY (id),
  CONSTRAINT course_subject__fk FOREIGN KEY (subject_id) REFERENCES public.subject (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_course_name_number UNIQUE (subject_id, name, number),
  CONSTRAINT unique_course_topic_name UNIQUE (topic_name),
  CONSTRAINT unique_course_topic_id UNIQUE (topic_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.course OWNER TO postgres;

COMMENT ON COLUMN public.course.subject_id IS 'The subject this course belongs to';
COMMENT ON COLUMN public.course.name IS 'The name of the course';
COMMENT ON COLUMN public.course.number IS 'The number of course';
COMMENT ON COLUMN public.course.synopsis IS 'Courses typically have a description of what why be offered.';
COMMENT ON COLUMN public.course.topic_name IS 'The topic name of this course. Used to build topic url';
COMMENT ON COLUMN public.course.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.course.updated_at IS 'Time this row was updated';

CREATE TABLE IF NOT EXISTS public.section
(
  id SERIAL,
  course_id BIGINT NOT NULL,
  number TEXT NOT NULL,
  call_number TEXT NOT NULL,
  now INTEGER DEFAULT -1 NOT NULL,
  max INTEGER DEFAULT -1 NOT NULL,
  status status NOT NULL,
  credits NUMERIC,
  topic_name TEXT,
  topic_id text,
  data BYTEA,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT section__pk PRIMARY KEY (id),
  CONSTRAINT section_course_id__fk FOREIGN KEY (course_id) REFERENCES public.course (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_number_call_number UNIQUE (course_id, number, call_number),
  CONSTRAINT unique_section_topic_name UNIQUE (topic_name),
  CONSTRAINT unique_section_topic_id UNIQUE (topic_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.section OWNER TO postgres;

COMMENT ON COLUMN public.section.course_id IS 'The course this section belongs to';
COMMENT ON COLUMN public.section.number IS 'Courses may have multiple sections and may choose to have this number to uniquely identify it. If the university doesn’t have a section number, assign one.';
COMMENT ON COLUMN public.section.call_number IS 'Universities use this number for registration. I’ve seen this called, index, class id, class #. ';
COMMENT ON COLUMN public.section.now IS 'Current class enrollment. Some universities don’t provide this. ';
COMMENT ON COLUMN public.section.max IS 'Maximum seats in the class. Some universities don’t provide this. ';
COMMENT ON COLUMN public.section.status IS 'This is the status of the section. While it’s obvious a section can be open or closed, there is a third option, cancelled.';
COMMENT ON COLUMN public.section.credits IS 'This value may live in the course above, but it should be mirrored down to each individual section';
COMMENT ON COLUMN public.section.topic_name IS 'The topic name of this subject. Used to build topic url';
COMMENT ON COLUMN public.section.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.section.updated_at IS 'Time this row was updated';


CREATE TABLE IF NOT EXISTS public.meeting
(
  id SERIAL,
  section_id BIGINT NOT NULL,
  room TEXT,
  day TEXT,
  start_time TIME,
  end_time TIME,
  class_type TEXT,
  index INTEGER,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT meeting__pk PRIMARY KEY (id),
  CONSTRAINT meeting_section_id__fk FOREIGN KEY (section_id) REFERENCES public.section (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_section_index UNIQUE (section_id, index)

)WITH (OIDS = FALSE);

ALTER TABLE public.meeting OWNER TO postgres;

COMMENT ON COLUMN public.meeting.section_id IS 'The section this meeting belongs to';
COMMENT ON COLUMN public.meeting.room IS 'This may include the full name, e.g Central King Building Room 356. To take great care when designing the UI.';
COMMENT ON COLUMN public.meeting.day IS 'The day this meeting is on';
COMMENT ON COLUMN public.meeting.start_time IS 'The start time for this meeting. The scraper should extract values of this format. hh:mm(AM|PM)';
COMMENT ON COLUMN public.meeting.end_time IS 'The end time for this meeting. The scraper should extract values of this format. hh:mm(AM|PM)';
COMMENT ON COLUMN public.meeting.class_type IS 'E,g Lecture, Recitation';
COMMENT ON COLUMN public.meeting.index IS 'The position of this meeting';
COMMENT ON COLUMN public.meeting.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.meeting.updated_at IS 'Time this row was updated';


CREATE TABLE IF NOT EXISTS public.instructor
(
  id SERIAL,
  section_id BIGINT,
  name TEXT,
  index INTEGER,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT instructor__pk PRIMARY KEY (id),
  CONSTRAINT instructor_section_id__fk FOREIGN KEY (section_id) REFERENCES public.section (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_instructor_index UNIQUE (section_id, index)

)WITH (OIDS = FALSE);

ALTER TABLE public.instructor OWNER TO postgres;

COMMENT ON COLUMN public.instructor.section_id IS 'The section this instructor belongs to.';
COMMENT ON COLUMN public.instructor.name IS 'The name of instructor';
COMMENT ON COLUMN public.instructor.index IS 'The index of the instructor';
COMMENT ON COLUMN public.instructor.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.instructor.updated_at IS 'Time this row was updated';

CREATE TABLE IF NOT EXISTS public.book
(
  id SERIAL,
  section_id BIGINT,
  title TEXT,
  url TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT book__pk PRIMARY KEY (id),
  CONSTRAINT book_section_id__fk FOREIGN KEY (section_id) REFERENCES public.section (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_book_title__section_id UNIQUE (title, section_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.book OWNER TO postgres;

COMMENT ON COLUMN public.book.title IS 'The title of book';
COMMENT ON COLUMN public.book.url IS 'The url of the book';
COMMENT ON COLUMN public.book.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.book.updated_at IS 'Time this row was updated';

CREATE TABLE IF NOT EXISTS public.metadata
(
  id SERIAL,
  university_id BIGINT,
  subject_id BIGINT,
  course_id BIGINT,
  section_id BIGINT,
  meeting_id BIGINT,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT metadata__pk PRIMARY KEY (id),
  CONSTRAINT metadata_university_id__fk FOREIGN KEY (university_id) REFERENCES public.university (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT metadata_subject_id__fk FOREIGN KEY (subject_id) REFERENCES public.subject (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT metadata_course_id__fk FOREIGN KEY (course_id) REFERENCES public.course (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT metadata_section_id__fk FOREIGN KEY (section_id) REFERENCES public.section (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT metadata_meeting_id__fk FOREIGN KEY (meeting_id) REFERENCES public.meeting (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_metadata_title__university_id UNIQUE (title, university_id),
  CONSTRAINT unique_metadata_title__subject_id UNIQUE (title, subject_id),
  CONSTRAINT unique_metadata_title__course_id UNIQUE (title, course_id),
  CONSTRAINT unique_metadata_title__section_id UNIQUE (title, section_id),
  CONSTRAINT unique_metadata_title__meeting_id UNIQUE (title, meeting_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.metadata OWNER TO postgres;

COMMENT ON COLUMN public.metadata.content IS 'The title of metadata';
COMMENT ON COLUMN public.metadata.title IS 'The url of the metadata';
COMMENT ON COLUMN public.metadata.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.metadata.updated_at IS 'Time this row was updated';


CREATE TABLE IF NOT EXISTS public.registration
(
  id SERIAL,
  university_id BIGINT,
  period period,
  period_date BIGINT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT registration__pk PRIMARY KEY (id),
  CONSTRAINT registration_university_id__fk FOREIGN KEY (university_id) REFERENCES public.university (id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT unique_registration_period__university_id UNIQUE (period, university_id)

)WITH (OIDS = FALSE);

ALTER TABLE public.registration OWNER TO postgres;

COMMENT ON COLUMN public.registration.period IS 'The period for which a season lasts for a specific university. This has to be manually updated. Comparable with season using ::text';
COMMENT ON COLUMN public.registration.period_date IS 'The url of the registration';
COMMENT ON COLUMN public.registration.created_at IS 'Time this row was inserted';
COMMENT ON COLUMN public.registration.updated_at IS 'Time this row was updated';



CREATE OR REPLACE FUNCTION update_row_time_stamp() RETURNS TRIGGER AS $$
BEGIN
  --
  -- Update the created_at and updated_at columns of each rows
  --
  IF (TG_OP = 'INSERT') THEN
    NEW.created_at = now();
    NEW.updated_at = now();
    RETURN NEW;
  ELSIF (TG_OP = 'UPDATE') THEN
    NEW.updated_at = now();
    RETURN NEW;
  END IF;

  RETURN NEW;
END;
$$
LANGUAGE plpgsql;


CREATE TRIGGER insert_university_time_stamps
BEFORE INSERT ON public.university
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_subject_time_stamps
BEFORE INSERT ON public.subject
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_course_time_stamps
BEFORE INSERT ON public.course
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_section_time_stamps
BEFORE INSERT ON public.section
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_meeting_time_stamps
BEFORE INSERT ON public.meeting
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_instructor_time_stamps
BEFORE INSERT ON public.instructor
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_book_time_stamps
BEFORE INSERT ON public.book
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_metadata_time_stamps
BEFORE INSERT ON public.metadata
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER insert_registration_time_stamps
BEFORE INSERT ON public.registration
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_university_time_stamps
BEFORE UPDATE ON public.university
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_subject_time_stamps
BEFORE UPDATE ON public.subject
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_course_time_stamps
BEFORE UPDATE ON public.course
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_section_time_stamps
BEFORE UPDATE ON public.section
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_meeting_time_stamps
BEFORE UPDATE ON public.meeting
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_instructor_time_stamps
BEFORE UPDATE ON public.instructor
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_book_time_stamps
BEFORE UPDATE ON public.book
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_metadata_time_stamps
BEFORE UPDATE ON public.metadata
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_registration_time_stamps
BEFORE UPDATE ON public.registration
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE OR REPLACE FUNCTION public.notify_status_change()
  RETURNS trigger AS
$BODY$
DECLARE
  _notification json;

  _university record;
  _subject record;
  _course record;
  _section record;
  _temp jsonb;
BEGIN

  SELECT university.id, university.name, abbr, main_color, abbr, home_page, registration_page, university.topic_name, university.topic_id
  INTO _university
  FROM university
    JOIN subject ON university.id = subject.university_id
    JOIN course ON subject.id = course.subject_id
    JOIN section ON course.id = section.course_id
  WHERE section.id = NEW.id;

  SELECT subject.id, subject.university_id, subject.name, subject.number, subject.season, subject.year, subject.topic_name, subject.topic_id
  INTO _subject
  FROM university
    JOIN subject ON university.id = subject.university_id
    JOIN course ON subject.id = course.subject_id
    JOIN section ON course.id = section.course_id
  WHERE section.id = NEW.id;

  SELECT course.id, course.subject_id, course.number, course.name, course.synopsis, course.topic_name,course.topic_id
  INTO _course
  FROM university
    JOIN subject ON university.id = subject.university_id
    JOIN course ON subject.id = course.subject_id
    JOIN section ON course.id = section.course_id
  WHERE section.id = NEW.id;

  SELECT section.id, section.course_id, section.number, section.call_number, section.now, section.max, section.status, section.credits::TEXT, section.topic_name,section.topic_id, subject.created_at, section.updated_at
  INTO _section
  FROM university
    JOIN subject ON university.id = subject.university_id
    JOIN course ON subject.id = course.subject_id
    JOIN section ON course.id = section.course_id
  WHERE section.id = NEW.id;

  _temp = jsonb_set(to_json(_course)::jsonb, '{sections}', json_build_array(to_json(_section))::jsonb);
  _temp = jsonb_set(to_json(_subject)::jsonb, '{courses}', json_build_array(_temp)::jsonb);
  _temp = jsonb_set(to_json(_university)::jsonb, '{subjects}', json_build_array(_temp)::jsonb);

  _notification = json_build_object(
      'topic_name', NEW.topic_name,
      'status', NEW.status,
      'max', NEW.max,
      'now', NEW.now,
      'university', jsonb_strip_nulls(_temp));


  -- Execute pg_notify(channel, notification)
  PERFORM pg_notify('status_events',_notification::text);

  RETURN NULL;
END;
$BODY$
LANGUAGE plpgsql;

CREATE TRIGGER notify_status_change
AFTER UPDATE
ON public.section
FOR EACH ROW
WHEN (OLD.status <> NEW.status)
EXECUTE PROCEDURE public.notify_status_change();

COMMIT;

CREATE extension pg_stat_statements;