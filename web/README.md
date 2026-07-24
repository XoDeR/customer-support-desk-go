# Support Desk web

React frontend for the Customer Support & SLA Desk.

## Run locally

Start the API on `http://localhost:8080`, then:

```bash
cd web
npm install
npm run dev
```

Open `http://localhost:5173`. The API base defaults to
`http://localhost:8080/api/v1`. Override it with `VITE_API_BASE_URL` in a
`web/.env.local` file:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

The API must allow the Vite origin (`http://localhost:5173`) with the
`Authorization` header. The included backend router already has this
development CORS configuration.
# Customer Support & SLA Desk frontend

React/Vite application for the Customer Support & SLA Desk API.

## Run locally

Start the API on `http://localhost:8080`, then:

```sh
cd web
npm install
npm run dev
```

Open the Vite URL shown in the terminal (normally `http://localhost:5173`).

To override the API URL, create `web/.env.local`:

```sh
VITE_API_URL=http://localhost:8080/api/v1
```

The app stores access and refresh tokens in local storage and refreshes an expired access token once after a `401`.
