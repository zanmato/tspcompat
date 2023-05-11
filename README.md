# tsp web app

This requires API access to [Svenskt teckenspråkslexikon](https://teckensprakslexikon.su.se)

## Development

* Start the DB with `docker-compose up -d`
* Copy `config.example.toml` and save it as `config.toml`, replace `data_url`
* Start the API service `go run cmd/api/*.go`
* Copy `frontend/.env.example` and save it as `frontend/.env`
* Start the frontend `cd frontend; yarn dev`

### Optional

* Download the [FreeSans SWL font](https://zrajm.github.io/teckentranskription/freesans-swl.woff2) and save it to `frontend/public/static/fonts/freesans-swl.woff2`