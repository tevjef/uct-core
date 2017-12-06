CREATE TYPE os AS ENUM (
  'android',
  'ios'
);

CREATE TABLE public.subscription
(
  id SERIAL,
  os os,
  os_version TEXT,
  app_version TEXT,
  is_subscribed BOOLEAN NOT NULL,
  topic_name TEXT NOT NULL,
  fcm_token TEXT NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT subscription__pk PRIMARY KEY (id)
);

CREATE TRIGGER insert_subscription_time_stamps
BEFORE INSERT ON public.subscription
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_subscription_time_stamps
BEFORE UPDATE ON public.subscription
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TABLE public.acknowledge
(
  id SERIAL,
  notification_id BIGINT NOT NULL,
  os os,
  os_version TEXT,
  app_version TEXT,
  topic_name TEXT NOT NULL,
  receive_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  CONSTRAINT acknowledge__pk PRIMARY KEY (id),
  CONSTRAINT acknowledge_university__fk FOREIGN KEY (notification_id) REFERENCES public.notification (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TRIGGER insert_message_time_stamps
BEFORE INSERT ON public.acknowledge
FOR EACH ROW
EXECUTE PROCEDURE update_row_time_stamp();

CREATE TRIGGER update_message_time_stamps
BEFORE UPDATE ON public.acknowledge
FOR EACH ROW
WHEN (OLD.* IS DISTINCT FROM NEW.*)
EXECUTE PROCEDURE update_row_time_stamp();
