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
- [x] Add [redis](https://github.com/go-redis/redis) for local short-link caching
- [x] Add [postgres](https://github.com/lib/pq) to store everything
- [x] IP rate limiting
- [ ] Create makefile for build/local running configs
- [x] Create [docker compose](https://docs.docker.com/compose/) setup for local dev<br />
<strike> Setup kubernetes config and deploy to [GKE](https://cloud.google.com/kubernetes-engine)</strike>
- [ ] Create a frontend for submitting links
- [x] [Normalize](https://github.com/PuerkitoBio/purell) urls so you don't get repeats
- [x] Allow custom short-links for my friends (ex. brill.wtf/minecraft)
- [ ] Auto generate short-links with a hash function (ex. brill.wtf/fhad-4jj)
- [ ] Generate 6-letter readable short-links with a [word bank](https://github.com/dwyl/english-words) (ex. brill.wtf/potato)
- [x] Store secrets in [doppler](https://doppler.com) for easy user adoption
- [x] Create [air](https://github.com/cosmtrek/air) config for easy local dev
- [x] Add [Datadog](https://www.datadoghq.com) for logging/monitoring
- [x] Add [pre-commit](https://pre-commit.com/) hooks
- [x] Migrate to [heroku](https://devcenter.heroku.com/articles/build-docker-images-heroku-yml#run-defining-the-processes-to-run)
- [ ] Ads????
- [ ] Analytics?????
