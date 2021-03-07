# brill.wtf

## Running

### Locally

Install links to everything checked off below.

Run postgres, make sure a db named `postgres` exists.

Run db_setup.sql in the `postgres` console.

Run the following commands:
```
go mod download
doppler run -- air
```

## Features

- [x] Make designs in Figma
- [x] Finalize the schema before getting tons built out
- [ ] Add [redis](https://github.com/go-redis/redis) for local short-link caching
- [x] Add [postgres](https://github.com/lib/pq) to store everything
- [ ] IP rate limiting
- [ ] Create makefile for build/local running configs
- [ ] Create [docker compose](https://docs.docker.com/compose/) setup for local dev
- [ ] Setup kubernetes config and deploy to [GKE](https://cloud.google.com/kubernetes-engine)
- [ ] Create a frontend for submitting links
- [x] Allow custom short-links for my friends (ex. brill.wtf/minecraft)
- [ ] Auto generate short-links with a hash function (ex. brill.wtf/fhad-4jj)
- [x] Store secrets in [doppler](https://doppler.com) for easy user adoption
- [x] Create [air](https://github.com/cosmtrek/air) config for easy local dev
- [ ] Add [Datadog](https://www.datadoghq.com) for logging/monitoring
- [x] Add [pre-commit](https://pre-commit.com/) hooks
- [ ] Ads????
- [ ] Analytics?????
