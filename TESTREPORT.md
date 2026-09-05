VERDICT: BUGS_FOUND

## Fehleranalyse

Der Native-Testlauf (`go test ./...`) schlägt mit **exit 1** fehl. `go build ./...` war grün, es handelt sich also nicht um einen Kompilierfehler, sondern um einen echten Laufzeit-/Routingfehler.

### Bug 1: Route `GET /flags/{key}/evaluate` ist nicht erreichbar und liefert 404

- **Titel**: Evaluate-Route liefert 404 statt des Evaluate-Handlers
- **Symptom**: Der in AC-07 geforderte Endpunkt zur Rollout-Entscheidung ist über den registrierten HTTP-Router nicht erreichbar. Ein Request auf `/flags/some.key/evaluate?user=u1` endet in 404, obwohl der Handler selbst existiert und dessen Direkttests bestanden haben. Damit kann kein Client die Feature-Rollout-Entscheidung abrufen – die Kernfunktion „deterministische Rollout-Entscheidung pro Nutzer“ fehlt zur Laufzeit.
- **Repro**:
  1. `go test ./...`
  2. Oder Server starten und `GET /flags/some.key/evaluate?user=u1` absetzen.
- **Evidence**:
  - Fehlende Testzeile:
    ```
    handlers_crud_test.go:55: route GET /flags/some.key/evaluate?user=u1 returned 404: not registered
    ```
  - Server-Logzeile:
    ```
    2026/09/05 02:28:23 GET /flags/some.key/evaluate 404
    ```
- **Suspected file(s)**: `main.go` (Routerdefinition). Das registrierte Muster `mux.Handle("GET /flags/{key}/evaluate", ...)` matcht offenbar nicht; ggf. auch `go.mod`, falls die deklarierte Go-Version unter 1.22 liegt und `http.ServeMux` die Methoden-/Wildcard-Muster (`{key}`) nicht auswertet. Der Handler `handlers_evaluate.go` selbst ist unschuldig – dessen Direkttests sind grün; der Fehler liegt in der Verdrahtung der Route.
- **Severity**: high