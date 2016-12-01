# uct-core [![Build Status](https://ci.tevindev.me/api/badges/tevjef/uct-core/status.svg)](https://ci.tevindev.me/tevjef/uct-core)

Core functions of UCT including servers, database schemas and other microservices.

# An Overview
![full](https://tevinjeffrey.me/content/images/2016/10/stack-8.png)

*This is a simplistic view of each services interactions. In reality they're a bit more connected.*

The Go applications in blue form the core of Course Trakr. Some scrapers, a program to clean, validate and push the scraped data to a database. One to serve data from the database to Android and iOS devices. And lastly, a program to publish notifications to users. 

# Hermes

```language-bash
$ docker exec uct-hermes hermes --help
usage: hermes [<flags>]

An application that listens for events from PostgreSQL and publishes notifications to Firebase Cloud Messaging

Flags:
      --help           Show context-sensitive help (also try --help-long and
                       --help-man).
  -d, --debug          Enable debug mode.
  -c, --config=CONFIG  Configuration file for the application.
```

It's quite cliché, but I named this after the messager of the Greek gods, [Hermes](https://en.wikipedia.org/wiki/Clich%C3%A9). As the usage information says, Hermes is *an application that listens for events from PostgreSQL and publishes notifications to Firebase Cloud Messaging*. It makes use of Postgres's [asynchronous notification](https://www.postgresql.org/docs/current/static/libpq-notify.html) feature to listen for data sent on a channel. In Postgres, I have a trigger that fires when a class opens or closes. It then packages relevant information about the class into a JSON object and sends the notification down a channel.

```language-sql
  -- Build notification
  notification = json_build_object(
      'notification_id', id,
      'status', NEW.status,
      'topic_name', NEW.topic_name,
      'university', _temp);
  -- Execute pg_notify(channel, notification)
  PERFORM pg_notify('status_change',notification::text);
```

Notifications are then sent to through Firebase Cloud Messaging, allowing me unite iOS and Android users under one platform. There really isn't any other messaging platform that competes with FCM, it's completely free for both iOS and Android. There is no limit on the number of notifications you publish and there's no limit on the number of users or subscriptions.

I utilize what FCM calls topic messaging. FCM topic messaging allows you to send a message to multiple devices that have opted in to a particular topic. You compose topic messages as needed, and Firebase handles routing and delivering the message reliably to the right devices.

# Spike

```language-bash
$ docker exec uct-spike spike --help
usage: spike [<flags>]

An application to serve university course information.

Flags:
      --help              Show context-sensitive help (also try --help-long and
                          --help-man).
  -p, --port=9876         Port to start server on.
  -c, --config=CONFIG     Configuration file for the application.
```

As the usage information suggests, Spike is a typical web API for serving data to clients. It serves data in a JSON or Protocol Buffer format depending on client support. Though the JSON route is mainly used for debugging since both Android and iOS clients have support for the Protobuf format. Connections to the database are pooled, queries are prepared, HTTP requests are logged and responses are cached using Redis.

A synthetic benchmark of 5000 requests with 100 concurrent connections will say the server handles ~190 requests/sec with 95% of them completing within 1.1354 secs.

Spike gets its name from Spike Spiegel, the protagonist of the anime Cowboy Bebop.

# Ein

```language-bash
$ docker exec uct-ein ein --help
usage: ein --format=[protobuf, json] [<flags>]

A command-line application for inserting and updated university information

Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --no-diff                  Do not diff against last data.
  -a, --insert-all               Disables optimizations when upserting relations.
  -f, --format=[protobuf, json]  Choose input format.
  -c, --config=CONFIG            Configuration file for the application.
```
Ein, the data daemon, named after [Ein](http://cowboybebop.wikia.com/wiki/Ein), the data dog from Cowboy Bebop. Not to be confused with [Data Dog](https://www.datadoghq.com/product/) the cloud montioring service.

I'm hesitant to call this a microservice since it's quite fat and has many responsibilities.

Scrapers queue university course data to be processed into Redis. Ein uses the `BLPOP` Redis primitive to dequeue the data as a unit of work. It processes and primes the data to be updated in the database. It performs a number of optimizations to reduce the amount of database transactions needed update a university's courses. These optimizations will reduce the number of transactions needed by ~90%. Data is stored in a normalized form for referential integrity and a unnormalized (denormalized) form for better read performance.

PostgreSQL itself was tuned to perform well under the high write volume. Connections to the database are pooled and reused. Queries are meticulously planned. The majority of the engineering effort went into this service. It's very resource hungry and is closely monitored.

# Scrapers

Course Trakr, currently has one stable scraper. The system as a whole is designed to be independent of the programming language used by the scraper. Go was my first choice here because Rutgers University has a [JSON API](http://api.rutgers.edu/) for course information. Go and it's 3rd libraries don't provide the necessary tools for scraping HTML/JS websites and I fully expect to have to use a different language for certain university websites. 

The decision to make it language independent would also allow for greater community contributions if I were to ever decide to make this project open source.

The requirements for Course Trakr scraper are as follows:

* Based on the system date, it must intelligently resolve current, next and last semesters of the university for which it scrapes.
* If nothing changed in the data source the output must be the same every time. Determinism is a must.
* It must reuse connections and rate limit itself. 
* It must produce an output that fits the defined Protobuf schema.
* It must write output to STDOUT

Once a scraper fulfills these requirements, it can be wrapped by a separate Go program (not yet written) to facilitate containerizing and scaling.

## Rutgers University API scraper
```language-bash
$ rutgers --help
usage: rutgers --campus=[CM, NK, NB] --format=[protobuf, json] [<flags>]

A web scraper that retrieves course information from Rutgers University's public api at http://api.rutgers.edu/.

Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
  -u, --campus=[CM, NK, NB]      Choose campus code. NB=New Brunswick,
                                 CM=Camden, NK=Newark
  -f, --format=[protobuf, json]  Choose output format
      --daemon=DAEMON            Run as a daemon with a refresh interval (default: 2m)
      --daemon-dir=DAEMON-DIR    If supplied, the daemon will write files to this directory
  -l, --latest                   Only output the current and next semester
  -v, --verbose                  Verbose log of object representations.
  -c, --config=CONFIG            Configuration file for the application.

```
```language-bash
$ rutgers -u NK --format protobuf
```
This is a command to run this program would typically look like. This will run once then output the scraped data to stdout. This should be typical invocation a CourseTrakr scraper. The wrapper I previously mentioned is currently baked into this program as an experiment. Adding the flag `--daemon=1m` flag will cause the program to scrape at 1 minute intervals then output the data to Redis.

![wide](https://tevinjeffrey.mehttps://tevinjeffrey.me/content/images/2016/10/gopher1-1.png)

This ensures the data in the database is only 1 minute stale. A minute is a long time for a popular class to be open for at the start of the semester when demand is the greatest. Users may be competing with student manually refreshing their browser. Reducing the daemon's interval is one way to increase the database's freshness. 

I found a number of problems with this approach. After the reduced interval, separate goroutines could be scraping simultaneously. One nearing completion and one just starting. This produces more logs means debugging gets a little bit more difficult. Reasoning about the program's internal state becomes difficult because  the complexity has increased. You now potentially have 2x the tcp connections. In a way, reducing the interval on a single instance in response to higher demand will only scale vertically.

Scaling this program is not as trivial as starting a new instance. You may end up in a situation where 2 instances start at nearly the same time. E.g. With an interval of 60 seconds, `rutgers-1` starts at `t=0` and another instance, `rutgers-2` starts at `t=5`. After the first scrape, they start again at `t=60` and `t=65` respectively. There is no benefit to this configuration. 

Each new instance will have to synchronize with other instances to divide the interval and synchronize their scraping. 

![wide](https://tevinjeffrey.mehttps://tevinjeffrey.me/content/images/2016/10/gopher2.png)

It follows that an instance `rutgers-1` would start at `t=0` and a replica of that instance, `rutgers-2`, will discover `rutgers-1`'s start time then resolve itself to begin scraping at `t=30`. This way, the data freshness will be evenly distributed across the 60 second interval. 

![wide](https://tevinjeffrey.mehttps://tevinjeffrey.me/content/images/2016/10/gopher3.png)

Again, it follows that scaling up to 3 instances will shave the data freshness down to 20 seconds. Looking at a 3 minute timespan, `rutgers-1` will scrape at `t=0,60,120,180`, `rutgers-2` at `t=20,80,140` and `rutgers-3` at `t=40,100,160`. The beauty of this approach is that each instance does not need to be on the same machine. They synchronize and discovery each sibling instance through a remote key-value store. They automatically reconfigure themselves when new instances are added or removed. The greater advantage of this approach is that this program could wrap any other scraper from any language. It could be built into a container to be deployed and scaled anywhere.

# Monitoring it all

Log management is not easy and there are tons of services created solely to make this part of application monitoring simple. I often get the feeling that I'm simultaneously logging too much and too little, but I've found that this means I'm not logging the correct information about my application's runtime. My advice is to make judicious use of log levels. Try to draw a line between logs necessary for *debug*ging an application and those necessary for *info*rming you about the internal state.

Docker has a builtin feature where it will collect container logs written to `STDERR` and `STDOUT` and, by default, write them to a file. Extending on top of that, you can specify a logging driver, an intermediary service where Docker will forward logs to. Through careful evaluation I choose the Fluentd driver to collect, filter and parse logs to be sent to **AWS S3** for archival, **AWS CloudWatch** for tailing and [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/) for metrics storage.

I also use [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/) as an agent to collect metrics from all services. This includes metrics from Postgres, Redis, NGINX, Docker and even the host machine itself. All of the metrics are inserted into InfluxDB where they can then be queried and visualized with in [Grafana](http://grafana.org/). 

![wide](https://tevinjeffrey.mehttps://tevinjeffrey.me/content/images/2016/10/Screen-Shot-2016-10-04-at-7.06.44-PM.png)

Being able to visualize your logs and metrics can give you invaluable insight into your service's behavior at runtime. 


