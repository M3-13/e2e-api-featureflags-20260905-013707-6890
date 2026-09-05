VERDICT: BUGS_FOUND

**Bug 1: Testfall-Reihenfolge in `TestRoutesAreReachable` führt zu fehlgeschlagenem `go test ./...`**

- **Symptom:** Aus Nutzersicht ist der Go-Testlauf rot (`go test ./...` exit 1). Der Subtest `evaluate_flag` meldet für `GET /flags/some.key/evaluate?user=u1` fälschlich einen 404 „not registered“, obwohl die Route registriert und der Endpoint für ein vorhandenes Flag eigentlich korrekt wäre.
- **Repro:** `go test ./...` ausführen. Im Test `TestRoutesAreReachable` werden alle Subtests nacheinander mit demselben Store ausgeführt. Der Subtest `delete flag` löscht zuvor den angelegten Seed-Flag `some.key`; der danach laufende Subtest `evaluate_flag` trifft deshalb auf einen leeren Store und erhält 404.
- **Evidence:**
  - `--- FAIL: TestRoutesAreReachable (0.03s)`
  - `--- FAIL: TestRoutesAreReachable/evaluate_flag (0.00s)`
  - `handlers_crud_test.go:55: route GET /flags/some.key/evaluate?user=u1 returned 404: not registered`
  - Log: `2026/09/05 02:03:04 GET /flags/some.key/evaluate 404`
- **Suspected file(s):** `handlers_crud_test.go` — der Test verwendet einen einzigen, über alle Subtests geteilten Store; die Subtest-Reihenfolge `delete flag` → `evaluate_flag` ist ursächlich. Der Produktcode (`handlers_evaluate.go`, Routing in `main.go`/`newRouter`) ist nicht betroffen.
- **Severity:** high — verletzt AC-11 („Testlauf (`go test ./...`) ist grün“) und blockiert die CI-Pipeline, auch wenn das Verhalten des eigentlichen Feature-Flag-Service unauffällig ist.