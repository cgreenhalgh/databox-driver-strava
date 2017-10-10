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

Now we can (re)build it:
```
docker build -t databox-driver-strava -f Dockerfile.dev .
```

And upload the manifest (databox-manifest.json) to the local app
store, [http://localhost:8181](http://localhost:8181).

Try starting the (empty) driver.
It may not appear in the list of drivers. Perhaps this is because there 
no web server running to return a status?

Try entering the container (you'll need to find its ID using `docker ps`).
```
docker exec -it CONTAINERID /bin/sh
```

To copy files into the container:
```
docker cp . CONTAINERID:/root/
```

Copy in the static JS and CSS files.
Touch main.js.
Copy in app/main.go and rip out all the unneeded stuff to do with plugs
(leave the web server, status).

Build and run...
```
GGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo -ldflags '-d -s -w -extldflags "-static"' -o app src/app.go
./app
```

Hmm, I still don't see it in the driver list.

Try two-phase build and normal deploy...

