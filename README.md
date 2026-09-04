# Feature-Flag-Service (Go)

Ein in Go ausschließlich mit der Standardbibliothek (`net/http`, kein externes
Web-Framework) gebauter Feature-Flag-Service. Er verwaltet Flags über
REST-Endpunkte, trifft deterministische Rollout-Entscheidungen pro Nutzer und
bietet Health-Check, Eingabevalidierung, JSON-Fehlerobjekte und
Zugriffs-Logging.

## Tech Stack

- **language**: Go
- **runtime**: Go-Standardbibliothek (`net/http`)
- **dependencies**: keine externen Frameworks
- **module**: `go.mod` (Go 1.22, nutzt ServeMux-Muster wie `POST /flags` und `{key}`)
- **tests**: Go-Tests mit `net/http/httptest`

## Installation

Voraussetzung: Go 1.22 oder neuer.

```bash
go mod tidy
```

## Ausführen (Dev)

```bash
go run .
```

Der Server startet auf Port **8080**. Über die Umgebungsvariable `PORT` lässt
sich ein anderer Port setzen:

```bash
# Windows (PowerShell)
$env:PORT = "9090"; go run .

# Unix
PORT=9090 go run .
```

## Build (Produktion)

```bash
go build ./...
```

## Endpunkte

Fehlerantworten sind überall ein JSON-Objekt der Form `{"error": "<message>"}`.

| Methode | Pfad                      | Beschreibung                                                        |
|---------|---------------------------|---------------------------------------------------------------------|
| POST    | `/flags`                  | Legt ein Flag an. Body: `{key, enabled, description?, rollout_percent?}` |
| GET     | `/flags`                  | Listet alle Flags (leer = `[]`)                                     |
| GET     | `/flags/{key}`            | Liefert ein einzelnes Flag                                          |
| PUT     | `/flags/{key}`            | Aktualisiert ein Flag (nur übergebene Felder)                       |
| DELETE  | `/flags/{key}`            | Entfernt ein Flag (204)                                             |
| GET     | `/flags/{key}/evaluate?user={id}` | Deterministische Rollout-Entscheidung für einen Nutzer   |
| GET     | `/healthz`                | Health-Check, antwortet mit `200` und dem Text `OK`                 |

### Flag (JSON)

```json
{
  "key": "string",
  "enabled": true,
  "description": "string",
  "rollout_percent": 0
}
```

## Features

- CRUD-Endpunkte für Feature-Flags mit JSON-Fehlerobjekten und passenden Statuscodes
- Deterministische, nutzerstabile Rollout-Entscheidung (FNV-1a-Hash)
- Health-Check (`GET /healthz`)
- Eingabevalidierung inkl. 1-MiB-Body-Limit (413) und Content-Type-Prüfung (415)
- Zugriffs-Logging (Methode, Pfad ohne Query, Statuscode)
- Thread-sicherer In-Memory-Store (`sync.RWMutex`)
