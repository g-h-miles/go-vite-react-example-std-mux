# Go Vite React Example with Standard Mux

A modern web application example using:

- Go (Backend)
- React with Vite (Frontend)
- Standard Go HTTP Mux for routing

## Development

To run the development server:

```bash
make dev
```

## Building

To build the project:

```bash
make build
```

This will:

1. Build the frontend assets
2. Build the Go binary with production settings

## Project Setup

We start with a basic Go application using the standard library HTTP mux where we have a single API endpoint to return some text. The frontend is a Vite React application created using `yarn create vite` in the frontend folder.

The frontend calls the API endpoint to return the text.

We have a [frontend.go](frontend/frontend.go) file which uses the embed package and the standard library to serve the content from the dist folder.

## Run in development mode

Start using `make dev`, will run vite build in watch mode, as well as the golang server using [air](https://github.com/cosmtrek/air)

### Build a binary

Using `make build`, will build frontend assets and then compile the golang binary embeding the frontend/dist folder

### Build a dockerfile

Using `docker build -t go-vite .`, will use a multistage dockerfile to build the vite frontend code using a node image, then the golang using a golang image, the put the single binary into an alpine image.

## Hot Reloading

Instead of serving static assets from Go when running in dev mode, we setup a proxy using `httputil.ReverseProxy` that routes requests to a running vite dev server unless the path is prefixed with `/api`. This allows HMR and live reloading to work as if you were running `vite dev` while still serving API paths from Go.

All the changes required to take the initial project and support hot module reloading can be found in the [pull request](https://github.com/danhawkins/go-vite-react-example/pull/1)

### Step 1

Change the [package.json](frontend/package.json) to run the standard `vite` instead of the `tsc && vite build --watch`

## Step 2

Update the [frontend.go](frontend/frontend.go) so that when we are running in dev mode we proxy requests to the vite dev server

Import a `.env` file using dotenv which just has `ENV=dev` inside

```golang
import(
_ "github.com/joho/godotenv/autoload"
)
```

If we are running in dev mode, setup the dev proxy

```golang
func RegisterHandlers(mux *http.ServeMux) {
  if os.Getenv("ENV") == "dev" {
    log.Println("Running in dev mode")
    setupDevProxy(mux)
    return
  }
  mux.HandleFunc("/", spaHandler(distDirFS))
}

func setupDevProxy(mux *http.ServeMux) {
  target, err := url.Parse("http://localhost:5173")
  if err != nil {
    log.Fatal(err)
  }
  proxy := httputil.NewSingleHostReverseProxy(target)
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/api") {
      http.NotFound(w, r)
      return
    }
    proxy.ServeHTTP(w, r)
  })
}
```

## Step 3

Run `make dev` to start the vite dev server and air for the golang server, changes in the frontend app will now be reflected immediatly.

**IMPORTANT: The go build will fail if frontend/dist/index.html is not available, so even if you are running in dev mode, make sure to run `make build` initially to populate the folder**
