# Golang Backend Rewrite TDD Parity Plan

This document tracks the backend parity test strategy so future agents can continue implementation without re-reading all PHP files.

## Scope

The current parity suite focuses on the PHP backend classes under:

- `old/api/src/Generate.php`
- `old/api/src/Schedule.php`
- `old/api/src/S3Manager.php`
- `old/inc/timeFunctions.php` (already mirrored via `rewrite/utils/time.go` tests)

## Current Test Files (Go Rewrite)

- `rewrite/api/src/generate_test.go`
- `rewrite/api/src/schedule_test.go`
- `rewrite/connections/s3_manager_test.go`
- `rewrite/utils/time_test.go`

## Harness Pattern

`generate_test.go`, `schedule_test.go`, and `s3_manager_test.go` use a registration-based harness:

- `newGenerateHarness`
- `newScheduleHarness`
- `newS3Harness`

Until real implementations register these harnesses, parity tests fail fast by design. This gives an explicit red TDD baseline until behavior is implemented.

## Required Behavior Contracts

### Generate parity (`old/api/src/Generate.php`)

1. **Overlap logic (`overlapBase`)**
   - Empty times => no conflict.
   - Conflict when starts/ends intersect on the same day.
   - Adjacent boundaries (end == start) are not conflicts.
   - Any matching timeslot pair causing overlap returns true.

2. **Course number normalization (`getCleanCourseNum`)**
   - Strips trailing section designators matching `-?[A-Z]\d{0,2}`.
   - Leaves untouched values that do not match that suffix pattern.

3. **Schedule generation (`generateSchedules`)**
   - Generates full Cartesian product across course option slots.
   - Prunes branches with course/course, course/non-course, and course/no-course conflicts.
   - Emits conflict errors when branches are invalid.

4. **Special course pruning (`pruneSpecialCourses`)**
   - Requires at least one required co-course when a parent course has group requirements.
   - Preserves schedules that satisfy all requirements.

### Schedule parity (`old/api/src/Schedule.php`)

1. **`icalFormatTime`**
   - Format: `HHMMSS`.
   - Minutes from midnight with modulo-24 behavior.

2. **`hashTime`**
   - Key format: `<start>-<end>-<buildingByStyle>-<room>`.

3. **`generateIcal` core structure**
   - Emits calendar header/footer.
   - Emits VEVENT entries for grouped timeslots.
   - Includes RRULE with BYDAY values from numeric day codes.
   - Includes `SUMMARY` and `LOCATION` (except non-course entries where applicable).

### S3 parity (`old/api/src/S3Manager.php`)

1. **`saveImage`**
   - Stores to `<id>.png` with `image/png` content type.

2. **`getImage`**
   - Requests `<id>.png` with `image/png` content type.

3. **`returnImage`**
   - Returns body from `getImage`.

## Next Steps for Implementers

1. Implement concrete Go types in mirror package paths (`rewrite/api/src`, `rewrite/connections`).
2. Register each test harness in `_test.go` helper files near implementations.
3. Turn the suite green by wiring constructors to real services/mocks.
4. Expand with DB-backed repository unit tests for:
   - schedule lookup (`getScheduleFromId` behavior),
   - old-id lookup (`getScheduleFromOldId`),
   - SVG render flow (`renderSvg`) with image conversion abstraction.
5. Add endpoint-layer tests mirroring PHP API scripts (`old/api/*.php`).

## Open Ambiguities to Resolve

- Exact Go package layout for API script equivalents (`entity.php`, `generate.php`, etc.).
- Error collection contract (global in PHP vs explicit return values in Go).
- Whether iCal generation should be deterministic (inject clock/randomness) or best-effort parity.

## Layout guidance: 1:1 mirror vs Go domains

When this document says "domain package", it means grouping files by backend concern (generation, schedules, entities, images, etc.) instead of by technical layer. A practical recommendation:

- Mirror top-level PHP path intent where possible (e.g., `api/src/generate`, `api/src/schedule`, `connections/s3`).
- Inside each mirrored area, use Go-style domain packages and explicit interfaces for DB, clock, random UID, and image rendering dependencies.
- Prefer Go conventions over exact PHP structure when they conflict, especially around:
  - returning `error` values instead of global mutable error state,
  - dependency injection for deterministic tests,
  - package naming and file splitting.
