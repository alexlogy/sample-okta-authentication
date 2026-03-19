# Sample Okta SAML Authentication (Go + Gin)

This is a sample Go application demonstrating SAML 2.0 authentication using Okta as the Identity Provider (IdP) and a sample Service Provider (SP) written in Go.

---

## Features

- SAML 2.0 Login via Okta
- Attribute extraction (email, firstName, lastName, phone)
- Session handling using `crewjam/saml`
- Protected API endpoint
- Local logout + Okta logout redirect
- Structured logging with `slog`

---

## Project Structure

```
.
├── config.json
├── main.go
├── routes/
├── middleware/
├── controller/
├── model/
└── README.md
```

---

## Prerequisites

- Go 1.26+
- Okta Developer Account
- SAML App configured in Okta

---

## Setup

### 1. Install dependencies

```
go mod tidy
```

---

### 2. Generate certificates

Generate SP signing + encryption certificate:

```
openssl req -x509 -newkey rsa:2048 \
  -keyout sample-app.key \
  -out sample-app.crt \
  -days 365 \
  -nodes \
  -subj "/CN=localhost"

openssl x509 -in sample-app.crt -outform PEM -out sample-app.pem
```

---

### 3. Configure `config.json`

Example:

```
{
  "app_url": "http://localhost:8080",
  "saml_metadata_url": "https://<your-okta-domain>/app/.../sso/saml/metadata",
  "saml_sp_key_file": "sample-app.key",
  "saml_sp_cert_file": "sample-app.cert",
  "logout_url": "https://<your-okta-domain>/login/signout?fromURI=http://localhost:8080"
}
```

---

## Okta Configuration

### Single Sign-On URL (ACS)
```
http://localhost:8080/saml/acs
```

### Audience URI (SP Entity ID)
```
http://localhost:8080/saml/metadata
```

⚠️ Must match exactly.

---

### Attribute Statements

| Name       | Value          |
|------------|----------------|
| email      | user.email     |
| firstName  | user.firstName |
| lastName   | user.lastName  |
| phone      | user.mobilePhone |

---

## Run the App

```
go run main.go
```

Server runs at:

```
http://localhost:8080
```

---

## Docker / Podman

Before building the image, ensure the following are completed:

- Generate the SAML certificate and key (see Setup section)
- Update `config.json` with your Okta application configuration

### Build Image

```
podman build -t sample-okta-app .
```

### Run Container

```
podman run -p 8080:8080 sample-okta-app
```

The application will be accessible at:

```
http://localhost:8080
```

---

## Routes

| Route            | Description |
|------------------|------------|
| `/`              | Landing |
| `/user/info`     | Protected endpoint |
| `/signout`       | Logout |
| `/saml/metadata` | SP metadata |
| `/saml/acs`      | Assertion Consumer Service |
| `/saml/sso`      | SAML login |

---

## Authentication Flow

1. Call:
   ```
   GET /user/info
   ```

2. Redirect to Okta

3. Okta → `/saml/acs`

4. Attributes extracted:
   ```
   email := samlsp.AttributeFromContext(ctx, "email")
   ```

---

## Logout Flow

### `/signout`

- Clears local session
- Redirects to Okta logout

---

## Common Issues

### 1. 404 on `/saml/sso`

Fix:
- Ensure correct ACS URL in Okta

---

### 2. Missing attributes

Fix:
- Configure Attribute Statements in Okta
- Match attribute names exactly

---

### 3. Audience restriction error

Fix:
- Ensure Audience URI matches:
  ```
  http://localhost:8080/saml/metadata
  ```

---

### 4. Config not found

Fix:
```
export CONFIG_PATH=./config.json
```

---

## Example Response

```
{
  "email": "xxx@xxx.com",
  "firstName": "Alex",
  "lastName": "Ng",
  "phone": "+6598765432"
}
```
