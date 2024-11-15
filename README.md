# Auth0 Logs Processor

This repository contains Go scripts to interact with [Auth0 logs](https://auth0.com/docs/deploy-monitor/logs/retrieve-log-events-using-mgmt-api#retrieve-logs-by-checkpoint) and store them in a PostgreSQL database. The main components are:

- **first-log-id/main.go**: Fetches the first log from a provided date.
- **logs-by-checkpoint/main.go**: Fetches logs starting from a specific log ID and inserts them into PostgreSQL.
- **shared/shared.go**: Contains shared functions and types used by the scripts.

## Functionality

### first-log-id/main.go
This script takes a date as an argument, obtains an Auth0 authentication token, and fetches the first log after the provided date. It then prints the log ID found. This ID is used as the input for the logs-by-checkpoint script.

#### Example usage:
```bash
go run first-log-id/main.go "2024-11-14"
```

### logs-by-checkpoint/main.go
This script fetches logs starting from a specific log ID (provided as an argument), retrieves log entries from Auth0, and inserts them into a PostgreSQL database. It continues fetching logs until no more are available.

#### Example usage:
```bash
go run logs-by-checkpoint/main.go <log-id>
```

Replace <log-id> with the starting log ID from which logs should be fetched.

## But why?
You may be wondering, "Why two scripts, each performing part of the work separately, instead of a single script that retrieves the log_id programmatically to use it in subsequent calls to fetch logs from Auth0?"

Answer: The Auth0 API has a mechanism that prevents reading logs from an ID once it has been retrieved, as something internally flags that the ID has already been read. This issue, on the Auth0 side, causes calls to not work as expected if made from the same process. I can go into more detail and demonstrate this behavior for anyone interested.

For a future version, we can try using a single Bash script that leverages both Go scripts under the hood.

## Requirements

- Go 1.18 or higher
- PostgreSQL running locally with the `logs` table set up

## PostgreSQL Setup

To run the program locally, create the `logs` table in PostgreSQL using the following DDL script:

```sql
CREATE TABLE public.logs (
    log_id character varying(100) NOT NULL,
    date timestamp with time zone NOT NULL,
    type character varying(100) NOT NULL,
    size integer,
    PRIMARY KEY (log_id)
);

CREATE UNIQUE INDEX logs_log_id_idx ON public.logs USING btree (log_id);
CREATE INDEX logs_date_idx ON public.logs USING btree (date);
CREATE INDEX logs_type_idx ON public.logs USING btree (type);
```

## Installation

Clone this repository:

```bash
git clone https://github.com/your-username/auth0-logs.git  
cd auth0-logs
```

Install dependencies:
```bash
go mod tidy
```