## Mealie iCal Proxy

A simple web server that generates an iCal for a Mealie instance's meal plan.

## Usage

To use, simply run with the `MEALIE_API_KEY` and `MEALIE_URL` environment variables set:

```bash
MEALIE_API_KEY=apikeyhere MEALIE_URL=https://mealie.example.com go run .
```

Use the calendar at the URL `http://localhost:3333/mealie.ics`.

## Configuration

Meal start times can be overridden with optional environment variables in 24-hour `HH:MM` format:

| Variable | Default |
| --- | --- |
| `MEALIE_BREAKFAST_TIME` | `09:00` |
| `MEALIE_LUNCH_TIME` | `12:00` |
| `MEALIE_DINNER_TIME` | `19:00` |

For example:

```bash
MEALIE_BREAKFAST_TIME=08:30 MEALIE_DINNER_TIME=18:00 \
  MEALIE_API_KEY=apikeyhere MEALIE_URL=https://mealie.example.com go run .
```
