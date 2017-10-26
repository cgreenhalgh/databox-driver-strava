# databox-driver-strava

Databox driver for the Strava API
Also a walk-through of the databox driver creation process/instructions.

Status: initiating...

Intending to implement it in go lang based on the [tplink plug driver](https://github.com/me-box/driver-tplink-smart-plug).

## Development Diary

Create repo on github (just a README.md to begin with).

Add to databox dev environment:
```
./databox-install-component cgreenhalgh/databox-driver-strava
```

Note, fails initially as there is no Dockerfile to build (or manifest to upload).

Create Dockerfile based on 
[tplink driver Dockerfile](https://github.com/me-box/driver-tplink-smart-plug/blob/master/Dockerfile)
but initially single phase development container, i.e. we'll be
actively developing within it. This strategy is explained in
[docker-dev.md](https://github.com/me-box/documents/blob/master/guides/docker-dev.md).

Create databox-manifest based on
[tplink driver databox-manifest.json](https://github.com/me-box/driver-tplink-smart-plug/blob/master/databox-manifest.json),
being sure to change the name (and preferably the description,
author, tags, homepage and repository.

Copy in the static JS and CSS files.
Touch main.js.
Copy in app/main.go and rip out all the unneeded stuff to do with plugs
(leave the web server, status).

## Dev build

Now we can (re)build it:
```
docker build -t databox-driver-strava -f Dockerfile.dev .
```

And upload the manifest (databox-manifest.json) to the local app
store, [http://localhost:8181](http://localhost:8181).

Try starting the (empty) driver.
It may not appear in the list of drivers. Perhaps this is because there 
no web server running to return a status?

(Note: see the install note below)

To copy files into the container:
```
docker cp . CONTAINERID:/root/
```
Try entering the container (you'll need to find its ID using `docker ps`).
```
docker exec -it CONTAINERID /bin/sh
```

Build and run...
```
GGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-d -s -w -extldflags "-static"' -o app src/app.go
./app
```

Note: need to ensure image is labelled with databox.type = driver.

## Normal build / deploy

Try two-phase build and normal deploy...

Works, but why is it different? (I even briefly saw the container appear 
when switching over).

## Oauth notes

Won't auth in iframe - need to redirect top-level (parent?!) browser,
and then link back into the right place in the app.

Will white-list localhost, but on phone that would have to be the App 
rather than in a browser (no implicit grant flow).

Browser needs the authorization URL and the client_id and a few other
required pameters (response_type=code, scope=view_private), while server
also needs the token request URL and client_secret (plus the returned
code) to complete the exchange and obtain the access token for subsequent
calls.

So, service config:
- request_uri - minus client_id and redirect_uri value
- token_uri - minus client_id, client_secret and code

App-specific service config:
- client_id
- client_secret

User-specific config:
- activity_since
- poll?
- poll pattern (cron-style??)

State:
- authorized?
- authentication token
- athlete id
- cache athlete information, e.g. firstname, lastname
- (last) activity count
- last activity start_date

Note: athletes/ID/stats is recommended polling point; check 
all_ride_totals.count
Then athlete/activities?after=START_DATE

## Install

Note: copy etc/oauth.json.template to etc/oauth.json and fill in 
client_id and client_secret from the Strava app.

## Driver / Databox notes

### Questions / issues

Q. Why does driver wait for store to be ready before starting server? Means you often get no UI on first opening.

A. Probably historical implementation detail. No real reason?!

Q. Why does driver always return 'active' status? what does that mean?

A. [Nominally](https://github.com/me-box/core-arbiter#status) this should return `active` or `standby`, the latter to indicate that it needs configuration.
It is not clear that it is used at all for drivers, although it is used by apps/drivers waiting for stores to indicate they are active.

Note. Driver panics on various problems which just makes databox restart it; probably it should do something more sensible to report unresolvable issues!

Q. How is the go http server threaded? what concurrency control is needed? can it deadlock? can it handle concurrent clients?

A. THreaded using standard go runtime support for go routines which appears to generate threads as required to back go routines. So go routines can be running in different threads. 
Go up front recommends using channels (like erlang) rather than shared state, but does provide mutexes, etc. However this is qualified [here](https://github.com/golang/go/wiki/MutexOrChannel) to suggest using mutexes for shared state.

### Notes

Databox API driver/app list returns a lot of `docker inspect` information. Status appears to be `State:{status:...}`, which is usually `running`.

