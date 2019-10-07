FUT
===

This is a prototype Fedora indexing service that will harvest records from a Fedora 3.x instance,
store them in a MySQL database and then provide fast access.
It is designed around the data model used by CurateND, which is based on the very early Samvera Sufia data model.

The name comes from me reusing another project, (the Fedora Utility Tool), to provided a web-based way to navigate and troubleshoot our Fedora models.


# Installing

1. Make sure a Go development environment is installed.
On my mac I can install it using Homebrew: `brew install go`.
Other systems with package managers must surely have similar commands.

1. The application also requires MySQL. Again, I can install it using `brew install mysql`.

1. Check out the code. The easiest way is to run `go get -d github.com/dbrower/fut`.
It will check out the source code to `~/go/src/github.com/dbrower/fut`.
Of course, if you had set `$GOPATH` to something else, your results will be different.
Compile the code by `cd ~/go/src/github.com/dbrower/fut && make`

1. Set up the database:

```
$ mysql -u root
mysql> create database fut;
mysql> create user 'fut'@'localhost';
mysql> grant all on fut.* to 'fut'@'localhost';
```

1. If you have a database dump `dbsource` you can then import the database by running `mysql -u fut fut < dbsource`.
(If you're at Hesburgh Libraries, contact me and I can give you a link to my copy.)

# Configuring and Running

The application is configured using a file (in TOML).
In `~/go/src/github.com/dbrower/fut/` directory make a file with the following,
filling in the correct values for your instance of Fedora:

```
Mysql = "fut@/fut"
Fedora = "https://[user]:[password]@[host]:[port]/fedora/"
TemplatePath = "./web/templates/"
StaticFilePath = "./web/static/"
Port = "8080"
```

Run the application with `fut -config config`.

# Using the Prototype

It is very basic, and only two routes work:

* View an item record. Use the path `http://localhost:8080/obj/und:zp38w953p3c` to view an HTML page giving info of the given item.
* View the config page: Visit `http://localhost:8080/config`.

The harvesting works, but by default only does it on demand. You can trigger a harvest by sending a POST request to the config page, e.g. `curl -X POST 'http://localhost:8080/config'`
To set up a periodic harvest, you need to add the setting to the database and restart the application:
```
$ mysql -u fut fut
mysql> insert into config (c_key, c_value) VALUES ("harvest-interval", "5m");
```

This sets a 5 minute harvest interval. You can set the interval for any length of time, e.g. use `30s` for every thirty seconds.

If you are using my database dump file, the following are interesting items to look at:

* <http://localhost:8080/obj/und:1v53jw84s7d> is a collection with more than 2600 members.
* <http://localhost:8080/obj/und:qv33rv0753n> is my illustrative example of an ETD with 5 attached files.
* <http://localhost:8080/obj/und:zp38w953p3c> is a collection with a lot of titles that use extended Unicode code pages.
* <http://localhost:8080/obj/und:v118rb69n77> is a Person record
* <http://localhost:8080/obj/und:jd472v26530> is a LinkedResource

