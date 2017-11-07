# databox-driver-strava

Databox driver for the Strava API

By Chris Greenhalgh <chris.greenhalgh@nottingham.ac.uk>,
Copyright (C) The University of Nottingham, 2017

Status: working for limited information about activities 

Roadmap:
- support oauth from databox app (requires changes to app)
- ? make it more efficient
- auto-poll at specified time(s)
- sync more information from strava

## Data sources

### Activities

This driver downloads your activities from Strava into a time-series in Databox. 
- datasource ID: `activities`
- store type: `store-json`
- API: time-series
- content type: `application/json`
- Schema: see below

Each activity is a JSON object with fields:
- `id`: activity ID (strava internal ID) (int) 
- `name`: activity name (title) (string)
- `distance`: total distance (metres) (float)
- `moving_time`: time moving (seconds) (int?)
- `elapsed_time`: time elapsed (seconds) (int?)
- `type`: activity type, e.g. "ride", "run" (string)
- `start_date`: start date/time, e.g. "2013-08-24T00:04:12Z" (string)
- `timezone`: local timezone for activity/athlete, e.g. "(GMT-08:00) America/Los_Angeles" (string)`

This is strict subset of the information in the [Strava activity](http://strava.github.io/api/v3/activities/).

The store timestamp is the activity `start_date` (ms since UNIX epoch, UTC).

## Install / use

To install from local build:
```
./databox-install-component cgreenhalgh/databox-driver-strava
```

### Configure

Note: currently authorizing the driver will only work from a browser on the machine running the databox, not from the databox app or from a remote browser.

1. In the Databox UI install the driver, called "Strava". 
1. Open the driver UI and "Link to Strava account"; log into Strava if required and then authorize access.
1. (Probably best for now to re-open the databox UI, or you can switch to the app)
1. In the driver UI "Sync data from strava"; wait a few seconds and reload the page to see the updated status (at some point it will be made to auto-update)

See the [activity summary app](https://github.com/cgreenhalgh/databox-app-activity-summary) for a possible way to view your strava activity data.

### Using a personal strava app

Note that the driver by default authenticates with a test/demo strava 
app (NB not databox app). Each strava app has rate limits that apply 
to total use across all concurrent users. So if other people are using
it a lot then you may not be able to download new activities.

You can link the driver to your own strava app by:

1. Creating a [strava app](https://www.strava.com/settings/api). For the authorization callback domain put 'localhost'.
1. Editing `databox-driver-strava/etc/oauth.json` and replacing the `client_id` and `client_secret` with the [values for your new app](https://www.strava.com/settings/api) ("Client ID", "Client Secret"; make sure you show them first!)
1. Open the driver UI and "Link to Strava account" again.

## Implementation notes

see [implementation notes](docs/implementation-notes.md)
