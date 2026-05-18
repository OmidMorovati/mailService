# Mailing Service API Documentation

This document provides details for both the gRPC and JSON APIs provided by the Mailing Service.

## Server Configuration

- **JSON API Default Port:** `:8080`
- **gRPC API Default Port:** `:8082`
- **Database:** SQLite (`mail.db` by default)

---

## JSON API

The JSON API follows REST-like patterns but often uses JSON bodies for both GET and POST/PUT requests.

### Data Structures

#### EmailEntry
```json
{
  "Id": 1,
  "EmailAddress": "user@example.com",
  "ConfirmedAt": "2026-05-18T12:00:00Z",
  "OptOut": false
}
```
*Note: `ConfirmedAt` can be provided as a Unix timestamp (integer) or an RFC3339 string during input, but is returned as an RFC3339 string.*

### Endpoints

#### Create Email
- **URL:** `/email/create`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
    "EmailAddress": "user@example.com"
  }
  ```
- **Response:** `EmailEntry` object of the created entry.

#### Get Email
- **URL:** `/email/get`
- **Method:** `GET`
- **Request Body:**
  ```json
  {
    "EmailAddress": "user@example.com"
  }
  ```
- **Response:** `EmailEntry` object if found.

#### Get Paginated Emails
- **URL:** `/email/paginated`
- **Method:** `GET`
- **Request Body:**
  ```json
  {
    "Page": 1,
    "Count": 10
  }
  ```
- **Response:** Array of `EmailEntry` objects.

#### Update Email
- **URL:** `/email/update`
- **Method:** `PUT`
- **Request Body:**
  ```json
  {
    "EmailAddress": "user@example.com",
    "ConfirmedAt": 1716033600,
    "OptOut": false
  }
  ```
- **Response:** `EmailEntry` object of the updated entry.

#### Delete Email (Opt-out)
- **URL:** `/email/delete`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
    "EmailAddress": "user@example.com"
  }
  ```
- **Response:** `EmailEntry` object with `OptOut` set to `true`.

---

## gRPC API

The gRPC API is defined in `mail.proto`.

### Service Definition
`service MailingService`

### Methods

#### CreateEmail
- **Request:** `CreateEmailRequest { string email_addr }`
- **Response:** `EmailResponse { EmailEntry email_entry }`

#### GetEmail
- **Request:** `GetEmailRequest { string email_addr }`
- **Response:** `EmailResponse { EmailEntry email_entry }`

#### UpdateEmail
- **Request:** `UpdateEmailRequest { EmailEntry email_entry }`
- **Response:** `EmailResponse { EmailEntry email_entry }`

#### DeleteEmail
- **Request:** `DeleteEmailRequest { string email_addr }`
- **Response:** `EmailResponse { EmailEntry email_entry }`

#### GetEmailPaginated
- **Request:** `GetEmailPaginatedRequest { int32 page, int32 count }`
- **Response:** `GetEmailPaginatedResponse { repeated EmailEntry email_entries }`

### Message Types

#### EmailEntry
```protobuf
message EmailEntry {
  required int64 id = 1;
  required string email = 2;
  optional int64 confirmed_at = 3;
  optional bool opt_out = 4;
}
```

#### Requests & Responses
- `CreateEmailRequest`: Contains `email_addr`.
- `GetEmailRequest`: Contains `email_addr`.
- `UpdateEmailRequest`: Contains `EmailEntry`.
- `DeleteEmailRequest`: Contains `email_addr`.
- `GetEmailPaginatedRequest`: Contains `page` and `count`.
- `EmailResponse`: Contains an optional `EmailEntry`.
- `GetEmailPaginatedResponse`: Contains a list of `EmailEntry`.
