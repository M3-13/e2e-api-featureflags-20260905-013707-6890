VERDICT: BUGS_FOUND

**Bug 1: Evaluate-Endpoint nicht erreichbar – Route `/flags/{key}/evaluate` liefert 404**
- **Symptom:** Der in der Spec geforderte Endpoint `GET /flags/{key}/evaluate?user={id}` ist zur Laufzeit nicht erreichbar. Obwohl das Flag `some.key` im Store existiert, antwortet der Server mit `404 Not Found` statt einer deterministischen Rollout-Entscheidung. Damit ist AC-07 nicht erfüllt und der Kern-Feature-Flag-Evaluate-Flow gebrochen.
- **Repro:** `go test ./...` ausführen; der Subtest `TestRoutesAreReachable/evaluate_flag` schlägt fehl.
- **Evidence:**
  - `--- FAIL: TestRoutesAreReachable/evaluate_flag (0.00s)`
  - `handlers_crud_test.go:55: route GET /flags/some.key/evaluate?user=u1 returned 404: not registered`
  - `2026/09/05 02:08:39 GET /flags/some.key/evaluate 404`
- **Suspected file(s):** Nicht eindeutig lokalisiert — die eine Route `GET /flags/{key}/evaluate` wird vom `http.ServeMux` nicht gematcht, während die einfacheren key-Routen funktionieren. Plausible Kandidaten sind die Router-Definition in `handlers_crud_test.go` bzw. `main.go` in Zusammenspiel mit `handlers_evaluate.go`.
- **Severity:** high