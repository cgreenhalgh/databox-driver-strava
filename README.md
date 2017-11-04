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
docker cp . CONTAINERID:/root/go/src/main/
```

Try entering the container (you'll need to find its ID using `docker ps`).
```
docker exec -it CONTAINERID /bin/sh
```
Note, 2017-10-30 needs 'fixed' version of lib-go-databox, [here](https://github.com/cgreenhalgh/lib-go-databox/tree/store-json-extras).
Check it out in /src/github.com/cgreenhalgh/lib-go-databox

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
required parameters (response_type=code, scope=view_private), while server
also needs the token request URL and client_secret (plus the returned
code) to complete the exchange and obtain the access token for subsequent
calls.

So, service config:
- request_uri - minus client_id and redirect_uri value
- token_uri - minus client_id, client_secret and code

Oauth exchanges:
- driver redirects to strava authorize with client_id, redirect url (which must match strava whitelist) and (optional) returned state
- stava displays authorize page based on client_id; user logs in if not already
- strava redirects (with 302) to redirect url with parameters state and code
- driver posts to strava token with client_id, client_secret and code; gets back access token and athlete record
- (access token provided on each API request to strava)

Revocation with revoke user-specific access token.

Hazzard(s):
- imposter with client_id can trigger authorise request; redirect url must match specified (could be localhost) to get code; if browser is compliant then need to serve (or proxy if http) that redirect to capture code
- if they don't have my client secret they can't get the actual token, but if they have then they have same access as the "legitimate" app and one can't be revoked without the other.
- i'm only reading so it could be worse...

Options:
- each person creates their own app and input their own client id and secret into the driver
- the driver ships with my client id and secret, and consequently other developers can easily get the same access
- my client secret is available through some other restricted channel so that it is less likely to be appropriated (what channel?! and how much less likely??)

Note, [runkeeper](https://runkeeper.com/developer/healthgraph/registration-authorization) is essentially the same (with one extra check, on redirect_url, when converting code to token).

### Oauth intermediary

Dom's [oauth intermediary](https://github.com/me-box/core-oauth-intermediary) 
- page for external API (e.g. healthgraph)
- call initially with param 'start'; redirects to external auth
- handles redirects back with error by redirect to "origin_uri"
- handles redirect with code by calling get token endpoint and returning token as data parameter in redirect to origin_uri
- it has an internal key and encrypts origin uri -> state and vice versa

So...
- still has to end with redirect that will get back to original driver
- return user token rather than code (client id and secret not required)
- currently any client at all could use it (no checks at all, not even client id)
- (current server config is HTTP only)
- the server gets the access token and so could do whatever... but restrains and passes responsibility back to original client
- no guarantee that initiator IS my app...but user will see /something/, then be prompted to give /my app/ access - consistency check?!

Notes
- could use own client secrets to time-limit new token grabs
- could retain user token and act as api proxy, but then need to check each request
- could require unique(ish) client id (could be allocated automatically) for use, and log use, and allow client revoke
- could log use and allow blocking and rate limiting, e.g. by IP
- could limit origin uri (lime redirect uri is limited) to limit exposure from hostile web sites
- could/should refuse iframe?!

Note: look at google info on oauth and not using embedded web views and how to redirect to app (also on iphone!)

## Driver state

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

Q. Why does the databox library for GO just use string for all datastore return values?

### Notes

Databox API driver/app list returns a lot of `docker inspect` information. Status appears to be `State:{status:...}`, which is usually `running`.

Databox JSON store and timeseries store have different APIs for the same logical operations (e.g. since, range) so are not directly replacements. Parameters are handled differently.

Note: GO Json store support incomplete: no since or range

### Debugging

use `docker ps` to find the container ID for your driver and for the driver's store (image type store-json and name including your driver name).

Check the logs with `docker logs CONTAINERID`

Enter the JSON store with `docker exec -it CONTAINERID /bin/sh` and check the database directly:
```
mongo
use db
db.timesseries.find()
db.KV.find()
```

### Fixing the go library

Fork it. 
Check out my version, make branch, checkout branch
```
git clone https://github.com/cgreenhalgh/lib-go-databox.git
cd lib-go-databox
git checkout -b store-json-extras
```

Update `app.go` to import my fork, `github.com/cgreenhalgh/lib-go-databox`.

(repeatedly) copy my version into the dev container:
```
docker exec CONTAINERID mkdir -p /src/github.com/cgreenhalgh/lib-go-databox/
docker cp CONTAINERID . /src/github.com/cgreenhalgh/lib-go-databox/
```

(Note: `go get github.com/cgreenhalgh/lib-go-databox` will only get the default HEAD, not a branch)

## more go notes

GO dependency manager
```
go get -u github.com/golang/dep/cmd/dep
```
initially
```
dep init
```

