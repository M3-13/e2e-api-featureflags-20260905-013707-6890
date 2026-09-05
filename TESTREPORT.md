VERDICT: BUGS_FOUND

## Fehler 1

- **Titel:** Evaluate-Endpunkt liefert 404 über den echten Router, obwohl das Flag existiert
- **Symptom:** `GET /flags/{key}/evaluate?user=...` ist über den in `newRouter` bzw. `main.go` registrierten ServeMux nicht nutzbar. Der Handler antwortet mit 404, obwohl der angefragte Flag-Key zuvor im Store angelegt wurde. Damit ist AC-07 (deterministische Rollout-Entscheidung über den HTTP-Endpunkt) im realen Routing-Kontext gebrochen.
- **Repro:** `go test ./...` → Subtest `TestRoutesAreReachable/evaluate_flag`
- **Evidence:**
  - `handlers_crud_test.go:55: route GET /flags/some.key/evaluate?user=u1 returned 404: not registered`
  - Logzeile des Testlaufs: `2026/09/05 02:21:49 GET /flags/some.key/evaluate 404`
- **Suspected file(s):** `handlers_evaluate.go` — der Handler extrahiert den Key manuell aus `r.URL.Path` (`strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")`), statt das vom ServeMux gesetzte `r.PathValue("key")` zu verwenden. Dadurch wird der Key im realen Router-Kontext nicht zuverlässig aufgelöst.
- **Severity:** high