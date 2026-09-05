VERDICT: CHANGES_REQUESTED

# Prüfbericht zum Feature-Flag-Service (Go-Backend)

Der vorliegende Stand ist ein reines REST-Backend ohne Endnutzer-UI. Maßgeblich sind daher vor allem die DSGVO (personenbezogene Daten bei `/flags/{key}/evaluate?user=...`, Logging) und der Cyber Resilience Act (Sicherheit, Updates, Dokumentation). Die AI-Act-, Impressums-, Cookie- und Barrierefreiheitspflichten für öffentliche Web-UIs sind nicht anwendbar.

## 1. DSGVO

### Positiv
- Der `user`-Parameter wird ausschließlich in `handlers_evaluate.go` zur Hash-Berechnung verwendet und nicht im `Store` gespeichert. Das entspricht den Anforderungen an Datenminimierung (Art. 5 Abs. 1 lit. c DSGVO).
- Die Logging-Middleware in `middleware.go` protokolliert nur `Methode`, `URL.Path` und `Statuscode`. Query-String und `user` werden nicht geloggt. Die Tests in `middleware_test.go` (`TestLoggingDoesNotLogQueryStringOrUser`) belegen dies.
- Der In-Memory-Store nutzt `sync.RWMutex` und verhindert Datenrenner bei parallelen Zugriffen (`store.go`).

### Befund 1 — hoch — Fehlende Authentifizierung und Autorisierung
`main.go` exponiert sämtliche Verwaltungsendpunkte (`POST /flags`, `PUT /flags/{key}`, `DELETE /flags/{key}`) sowie lesende Endpunkte ohne jegliche Zugriffskontrolle. Jeder, der den Port erreichen kann, kann Flags anlegen, verändern, löschen und auslesen. Das gefährdet die Vertraulichkeit, Integrität und Verfügbarkeit (Art. 32 DSGVO, Sicherheit der Verarbeitung) und widerspricht dem CRA-Grundsatz „security by design/default“.

**Abhilfe:** In `main.go` eine Authentifizierungs-Middleware (z. B. API-Key/Bearer-Token aus der Konfiguration, Vergleich in konstanter Zeit) einführen und alle Routen außer `/healthz` damit schützen. Ein unbekannter oder fehlender Token muss 401 liefern. Die bestehenden Handler-Tests müssen um einen Test-Token ergänzt werden, damit das Produkt unter seiner eigenen Sicherheitsvorgabe weiterhin korrekt läuft. Alternativ kann die Authentifizierung auf einem vorgelagerten Reverse-Proxy erfolgen, muss dann aber in der Betriebsdokumentation verbindlich festgehalten werden.

### Befund 2 — hoch — Keine Transportverschlüsselung
`main.go` startet ausschließlich `http.ListenAndServe` (Klartext-HTTP). Der `user`-Parameter wird bei `GET /flags/{key}/evaluate` im Query-String übertragen und wäre im Netzwerk abhörbar, wenn der Dienst direkt exponiert wird. Das betrifft die Vertraulichkeit personenbezogener Daten (Art. 32 Abs. 1 DSGVO).

**Abhilfe:** In `main.go` eine TLS-fähige Serverkonfiguration ergänzen (`http.Server` mit `ListenAndServeTLS`), wenn Zertifikatspfade gesetzt sind. Falls TLS an einem vorgelagerten Load-Balancer terminiert wird, muss dies in `README.md` eindeutig dokumentiert und der interne Port darf nicht öffentlich erreichbar sein.

### Befund 3 — mittel — Log-Injection möglich
`middleware.go` loggt `r.URL.Path` ungefiltert mit `log.Printf("%s %s %d", ...)`. Ein Angreifer kann über URL-kodierte Steuerzeichen (z. B. `%0A`, `%0D`) im Pfad zusätzliche, gefälschte Logzeilen erzeugen. Das beeinträchtigt die Integrität der Serverprotokolle und ist ein Sicherheitsmanko.

**Abhilfe:** Pfad im Log bereinigen oder mit `%q` ausgeben, z. B. `log.Printf("%s %q %d", r.Method, sanitizeLogToken(r.URL.Path), rec.status)`, wobei `sanitizeLogToken` CR/LF/TAB/andere Steuerzeichen entfernt oder escaped. Die bestehenden Logging-Tests bleiben mit `%q` weiterhin grün.

### Befund 4 — niedrig — Betriebliche Datenschutzhinweise fehlen im sichtbaren Codeumfeld
Der `Flag`-Typ enthält frei befüllbare Felder `Key` und `Description`. Der Betreiber könnte dort personenbezogene Daten hinterlegen, ohne dass der Code dies verhindert. Außerdem müssen Server-Logs (auch ohne `user`-Angabe) einer Aufbewahrungsfrist unterliegen. `README.md` ist vorhanden, aber sein Inhalt ist hier nicht dargestellt; es sollte entsprechende Betriebshinweise enthalten.

**Abhilfe:** In `README.md` einen Datenschutzabschnitt ergänzen: Flag-Key und Description dürfen keine personenbezogenen Daten enthalten, der `user`-Parameter soll ein Pseudonym sein, Logs müssen mit definierter Aufbewahrungsfrist betrieben werden. Falls die Datei diese Hinweise bereits enthält, entfällt dieser Befund.

## 2. Cyber Resilience Act (CRA)

### Positiv
- Es werden keine externen Frameworks verwendet; das Abhängigkeitsrisiko ist minimal.
- Wesentliche Sicherheitsmaßnahmen sind implementiert: Body-Limit von 1 MiB (`body.go`, AC-12), Content-Type-Prüfung (`body.go`, AC-13), restriktives Flag-Key-Format (`handlers_create.go`, AC-14), Mutex gegen Datenrennen (`store.go`) und JSON-Fehlerobjekte.
- Die Codebasis enthält umfangreiche Go-Tests, die das Kernverhalten absichern.

### Befund 5 — mittel — Keine sichtbare CRA-Dokumentation zu Update-/Patch-Prozess, SBOM und Sicherheitskontakt
Die gezeigten Quelltextdateien enthalten keine Informationen zu Sicherheits-Updates, unterstütztem Versionsstand, Schwachstellen-Meldestelle oder einer SBOM. `README.md` ist vorhanden, wurde hier aber nicht inhaltlich dargestellt. Der CRA verlangt für Produkte mit digitalen Elementen dokumentierte Sicherheitseigenschaften und einen Update-/Patch-Weg.

**Abhilfe:** In `README.md` einen Abschnitt „Sicherheit & Compliance“ ergänzen, sofern noch nicht vorhanden: Sicherheitsmerkmale (Body-Limit, Content-Type-Prüfung, Key-Whitelisting, kein PII-Logging), Aktualisierungsprozess, unterstützte Go-Version, SBOM-Hinweis (`go version -m` bzw. `go.mod`) und Kontakt für Sicherheitsmeldungen.

### Befund 6 — mittel — HTTP-Server ohne Timeouts, Header-Limits und Ratenbegrenzung
`main.go` verwendet `http.ListenAndServe` mit einem Standard-Server, bei dem `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout` und `MaxHeaderBytes` nicht gesetzt sind. Das erhöht das Risiko für Slowloris-/Ressourcenerschöpfungsangriffe. Eine Ratenbegrenzung fehlt ebenfalls.

**Abhilfe:** In `main.go` eine eigene `http.Server`-Instanz erzeugen und sinnvolle Werte setzen, z. B. `ReadHeaderTimeout: 5s`, `ReadTimeout: 10s`, `WriteTimeout: 10s`, `IdleTimeout: 60s`, `MaxHeaderBytes: maxBodyBytes`. Optional eine einfache Rate-Limit-Middleware ergänzen.

### Befund 7 — niedrig — Go-Version / Patchstand nicht sichtbar
`go.mod` ist vorhanden, sein Inhalt wurde aber nicht angezeigt. Eine veraltete oder ungepatchte Go-Version wäre ein CRA-relevantes Risiko.

**Abhilfe:** In `go.mod` eine aktuelle, mit Sicherheitsupdates versorgte Go-Version festlegen und den Aktualisierungsprozess in `README.md` dokumentieren.

## 3. EU AI Act
Nicht anwendbar. Es ist keine KI-Funktion oder automatisierte Entscheidungsfindung mit Personenbezug im Sinne des AI Act erkennbar; die deterministische Hash-basierte Rollout-Entscheidung ist eine einfache Regel, keine KI.

## 4. Pflichttexte & UI
Nicht anwendbar. Das Produkt ist ein reines Backend ohne öffentliche Web-Oberfläche, daher bestehen keine Impressums-, Cookie-/Consent- oder Barrierefreiheitspflichten aus dieser Prüfperspektive.

## 5. Barrierefreiheit (WCAG/BITV/EAA)
Nicht anwendbar. Kein öffentliches UI im Lieferumfang.

## Gesamtbewertung
Das Produkt erfüllt die spezifizierten Datenschutz-Anforderungen an Logging und Speicherung des `user`-Parameters. Es bestehen jedoch wesentliche, behebbare Sicherheitslücken: fehlende Authentifizierung, fehlende Transportverschlüsselung und unzureichende Server-Härtung. Diese Punkte gefährden die Vertraulichkeit und Integrität potenziell personenbezogener Anfragen und verhindern derzeit die Marktfreigabe als konformes Backend. Die notwendigen Änderungen sind konkret umsetzbar und brechen keine legitime Produktfunktion, sofern die Tests entsprechend angepasst werden.