# Code Test 

This Application is intended to test a developers skills against some common problems faced by the Trading Solutions team. It contains a Model, a merger and a Core service. 

The Tasks in this list are intentionally vague to give the developer freedom to pick the solution they think is best.

With each task take into consideration performance, maintainability Documentation and testing.

To make it easier to review each task could you please add a merge request for each task with any documentation you think appropriate.

## Task 1

Add a New value to the system to identify if an Event should be displayed or hidden.

Ensure this new value is able to be returned on the `GetSportEvent` RPC

## Task 2

The current structure sports Sport Match style events but horse / greyhound / harness racing does not really fit into this structure.

Add a new Racing substructure to the system with some fields you think are appropriate.

Add a new `GetRacingEvent` RPC that returns a structure that would be useful to consumers wanting a racing event.

Consider writing a new Transform if there are racing specific transformations you want to make.

## Task 3

Add a ClosedAt time to markets and set this value in the transform the first time a market is closed.

## Task 4

Swap out the Redis database technology for MongoDB. Remember we are using vendoring so you will need to run `go mod tidy` and `go mod vendor` in the root folder after adding any new dependencies


## Task 5

Write a new `SearchEvents` RPC in Core that allows a user to search by date and/or bettingstatus and/or display.

Return a slice of all events that match the criteria.
 
# Apendix

## Prerequisites
* Go 1.25+ installe
* Protoc > 32 installed
* Docker (you could use another container system but might need to make some minor edits to the scripts)
* run `./updategotools.sh` to install some of the linting and formatting tools used

## Layout
### Model

This contains a proto definition of the Event Model used by the code test. You can run `gen-proto.sh` to regenerate the model.

### Merger

This package merges 2 partial events together, when adding new fields you will need to make sure you have updated the code in this package to merge the new fields correctly.

### Core

The main service of the code test. This spins up a GRPC server and exposes an RPC to `Update` and another more user friendly API to retrieve the event `GetSportEvent`

It persists the data in redis and merges a partial update to an event with the existing copy of the event in the database, runs some transformations on the event and saves back to the database.

## Data Flow
This section will just describe how data flows through this codetest service
1. `Update` RPC called the first time
2. Respository Package called to find existing event - does not exist
3. `transforms.sporttransform.TransformEvent` called to mutate the event
4. Respository Package called to save the updated version of the Event
5. `Update` RPC called with a partial update to the existing event
6. Respository Package called to find existing event - the event is retrieved
7. `merger.MergeEvent` called to merge in the new changes with the existing Event
8. `transforms.sporttransform.TransformEvent` called to mutate the event
9. Respository Package called to save the updated version of the Event

## Running the code

you will need to fix up the import paths to match your local repository structure.

`./run_local.sh` will spin up the docker contains in the docker compose file an start running the core service. It will run a rRPC server on port 50051 and http on 8080 (nothing in the test uses the httpserver)

You can then use postman to hit `localhost:50051` with a grpc request, load in the `./core/core.proto` file and then hit the`update` RPC with a payload like `{"Event": {"ID": "testEvent"}}` 

Switch to the GetSportEvent rpc and look up the event you just created `{"EventID": "testEvent"}`

There are some example payloads in the `./exampledata` folder that you can use to send to the `update` RPC


# Sasith Koralagoda

## TASK 01
Added a `Display` field to control whether an event should be displayed or hidden.

### Behaviour
- `Display: true` - Event is visible.
- `Display: false` - Event is hidden.
- When creating an event without `Display`, it is hidden by default (`false`).
- When updating an event without `Display`, the existing visibility remains unchanged.
- When updating with `Display: true` or `Display: false`, the visibility is updated accordingly.

**Note:** If `Display` is not provided, it remains `nil` in the event model, but `GetSportEvent` returns `false` (hidden) by default.

### Example Update Request
```json
{
  "Event": {
    "ID": "testEvent",
    "Display": {
      "Value": true
    }
  }
}
```

### Testing
Implemented unit tests for visibility merging and response conversion.
Run all tests using:
```bash
go test ./... -count=1
```


## TASK 02
Added support for Horse, Greyhound, and Harness racing events using a separate `RacingEvent` model and a new `GetRacingEvent` gRPC endpoint.

### Behaviour
- Racing events support three types: Horse, Greyhound, and Harness.
- Racing-specific details include `RaceName`, `RacingType`, `Distance`, and `Venue`.
- `GetRacingEvent` retrieves racing event details using the event ID.
- Partial updates preserve existing racing information when fields are not provided.
- The `Display` field controls event visibility, as implemented in Task 1.

**Note:**
- `SportTypeID` is a machine-readable identifier, such as `horse_racing`.
- `RacingType` is a human-readable description, such as `Horse`.
- `Distance` is measured in metres.
- When `StartTime` is not provided, the API returns an empty value.

### Example Update Request
```json
{
  "Event": {
    "ID": "brisbaneRace001",
    "Name": {
      "Value": "Brisbane Race 1"
    },
    "EventTypeID": {
      "Value": "horse_racing"
    },
    "BettingStatus": {
      "Value": "BettingOpen"
    },
    "Display": {
      "Value": true
    },
    "RacingData": {
      "RaceName": {
        "Value": "Brisbane Racing Test"
      },
      "RacingType": {
        "Value": "Horse"
      },
      "Distance": {
        "Value": "1200"
      },
      "Venue": {
        "Value": "Brisbane"
      }
    }
  }
}
```

### Example GetRacingEvent Request
```json
{
  "EventID": "brisbaneRace001"
}
```

### Testing
- Manually tested Horse, Greyhound, and Harness racing events using Postman.
- Added unit tests for racing data merging and response conversion.
- Verified partial updates preserve existing racing details.

Run all tests using:
```bash
go test ./... -count=1
```


## TASK 03
Added a `ClosedAt` field to record when a betting market closes for the first time.

### Behaviour
- When a market is `BettingOpen`, `ClosedAt` remains empty.
- When a market changes to `BettingClosed`, the current timestamp is automatically recorded.
- When a closed market receives another normal update, the original `ClosedAt` is preserved.
- Existing market information remains unchanged during timestamp updates.
- `ClosedAt` is stored as a Unix timestamp in nanoseconds.

**Note:** `ClosedAt` is generated by `MarketTransform`. The implementation assumes clients do not manually provide or overwrite this field.

### Example Update Request
Create an open market:
```json
{
  "Event": {
    "ID": "task03Event001",
    "Markets": [
      {
        "ID": "market001",
        "BettingStatus": {
          "Value": "BettingOpen"
        }
      }
    ]
  }
}
```

Close the market:
```json
{
  "Event": {
    "ID": "task03Event001",
    "Markets": [
      {
        "ID": "market001",
        "BettingStatus": {
          "Value": "BettingClosed"
        }
      }
    ]
  }
}
```

### Testing
- Tested open, closed, and updated markets through Postman.
- Verified `ClosedAt` is generated when a market closes.
- Verified subsequent normal updates preserve the original timestamp.
- Added unit tests for open, newly closed, and already closed markets.

Run all tests using:
```bash
go test ./... -v -count=1
```


## TASK 04
Replaced Redis with MongoDB as the event storage database.

### Behaviour
- Events are stored in MongoDB instead of Redis.
- `UpdateEvent` creates or replaces an event in MongoDB.
- `GetEventByID` retrieves an event using its ID.

**Note:**
- MongoDB stores events in the `codetest` database and `events` collection.
- Each event uses its ID as the MongoDB document `_id`.
- Event data is stored as JSON within the MongoDB document.
- The local MongoDB connection uses `mongodb://localhost:27018`.

### Docker Setup
Start MongoDB using:
```bash
docker compose up -d
```

Verify MongoDB is running:
```bash
docker exec codetest-mongodb mongosh --quiet --eval 'db.adminCommand({ping: 1})'
```

### Testing
- Updated repository and service integration tests to use MongoDB.
- Tested event creation and retrieval through Postman.
- Verified event data is stored correctly in MongoDB.
- Confirmed existing functionality from Tasks 1, 2 and 3 still works.
- Updated Go dependencies and vendor files.

Run all tests using:
```bash
go test ./... -count=1
```


## TASK 05
Added a `SearchEvents` gRPC endpoint to search events by date, betting status, and display visibility.

### Behaviour
- Search by `Date`, `BettingStatus`, and/or `Display`.
- All filters are optional and can be combined using AND logic.
- Date uses `StartTime` in UTC (`YYYY-MM-DD`).
- No filters returns all events; no matches returns an empty list.

**Note:**
- Missing `Display` defaults to `false` (hidden).
- Missing `BettingStatus` defaults to `BettingUnknown`.
- Events without `StartTime` are excluded from date searches.

### Testing
- Tested individual filters and all filter combinations using Postman.
- Verified searches with no matching events return an empty list.
- Added automated tests for individual and combined filters.
- Tested `Display: true`, `Display: false`, and invalid date formats.

Run all tests using:

```bash
go test ./... -v -count=1
```


## Submission Contents
- **Source Code:** Go implementation of Tasks 1–5, including unit tests.
- **README:** Implementation details and testing instructions.
- **Postman Collection:** Exported JSON collection containing API requests for testing all five tasks.
