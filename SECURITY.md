VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

### Scanner-Hinweis
Für dieses Go-Projekt sind laut Scanner-Output keine anwendbaren Security-Scanner (bandit / pip-audit / npm audit / semgrep) vorgesehen. Das Fehlen von Scanner-Ergebnissen ist kein Nachweis für Abwesenheit von Schwachstellen. Die folgende Bewertung basiert auf manueller Code-Analyse des sichtbaren Produktstands.

### Zusammenfassung
Es wurden keine hochkritischen oder kritischen Schwachstellen gefunden (keine Hardcoded Secrets, keine Injection/RCE, keine bekannten verwundbaren Abhängigkeiten, keine PII-Leaks). Es bestehen jedoch mehrere sicherheitsrelevante Härtungslücken und ein mittleres Risiko durch fehlende Authentifizierung/Autorisierung. Daher ist das Produkt nicht ohne Änderungen freizugeben.

---

### Findings

#### 1. [Mittel] Fehlende Authentifizierung und Autorisierung
**Betroffen:** `main.go` (Routenregistrierung), alle Handler (`handlers_create.go`, `handlers_update.go`, `handlers_delete.go`, `handlers_read.go`, `handlers_evaluate.go`)

**Beschreibung:**  
Der Dienst stellt sämtliche REST-Endpunkte (POST/PUT/DELETE/GET) ohne jegliche Authentifizierung oder Autorisierung bereit. Jeder, der den TCP-Port erreicht, kann Feature-Flags anlegen, verändern, löschen und auslesen. Dies ist insbesondere für die mutierenden Endpunkte ein mittleres Risiko, falls der Dienst nicht ausschließlich in einem vertrauenswürdigen internen Netzwerk betrieben wird.

**Konkrete Lösung:**  
Vor die betroffenen Routen eine Authentifizierungs-Middleware schalten, die z. B. ein Bearer-Token oder mTLS erzwingt. Mindestens sollte die Dokumentation klarstellen, dass der Dienst nur hinter einem authentifizierenden Gateway/Load-Balancer betrieben werden darf. Alternativ das Binding auf Loopback/interne Netze beschränken.

---

#### 2. [Mittel] Fehlende HTTP-Server-Timeouts (Slowloris/DoS)
**Betroffen:** `main.go`, Zeile mit `http.ListenAndServe(":"+port, mux)`

**Beschreibung:**  
Der Server verwendet `http.ListenAndServe` mit den Standard-Timeouts (keine!). Dadurch ist er anfällig für langsame Request-Angriffe (Slowloris), bei denen Verbindungen lange offen gehalten werden und Ressourcen erschöpfen.

**Konkrete Lösung:**  
Expliziten `http.Server` mit sinnvollen Timeouts verwenden:
```go
server := &http.Server{
    Addr:              ":" + port,
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
}
if err := server.ListenAndServe(); err != nil {
    log.Fatal(err)
}
```

---

#### 3. [Niedrig] Log-Injection über unbereinigten Request-Pfad
**Betroffen:** `middleware.go`, Zeile `log.Printf("%s %s %d", r.Method, r.URL.Path, rec.status)`

**Beschreibung:**  
Der Pfad aus `r.URL.Path` wird ungeprüft in den Log geschrieben. Ein Angreifer kann durch URL-kodierte Steuerzeichen wie `%0a` (Newline) den Log-Ausgabe manipulieren, gefälschte Log-Zeilen einschleusen oder echte Einträge verschleiern.

**Konkrete Lösung:**  
Den Pfad in Anführungszeichen oder mit Escaping loggen:
```go
log.Printf("%s %q %d", r.Method, r.URL.Path, rec.status)
```
Alternativ eine eigene Sanitization, die nicht druckbare Zeichen ersetzt.

---

#### 4. [Niedrig] JSON-Decoder akzeptiert zusätzliche Daten nach dem Dokument
**Betroffen:** `body.go`, Funktion `decodeJSONBody`

**Beschreibung:**  
Nach `dec.Decode(dst)` wird nicht geprüft, ob der Body noch weitere JSON-Tokens oder Daten enthält. Dadurch akzeptiert der Server Anfragen wie `{"key":"a"}garbage` oder `{"key":"a"}{"x":1}`. Das kann zu Inkonsistenzen führen und in manchen Proxy-Konstellationen Request-Smuggling begünstigen.

**Konkrete Lösung:**  
Nach erfolgreichem Decode prüfen, dass der Decoder am Ende des Bodys steht:
```go
if err := dec.Decode(dst); err != nil { ... }

if _, err := dec.Token(); err != io.EOF {
    writeError(w, http.StatusBadRequest, "invalid JSON body")
    return errors.New("trailing data")
}
```
`io` muss entsprechend importiert werden.

---

#### 5. [Niedrig] Unbegrenzte Länge des `user`-Query-Parameters
**Betroffen:** `handlers_evaluate.go`, Zeile `user := r.URL.Query().Get("user")`

**Beschreibung:**  
Der Wert des `user`-Query-Parameters wird ohne Längenbegrenzung in die FNV-1a-Hash-Berechnung übernommen (`RolloutHash`). Ein Angreifer kann sehr lange Werte senden, um CPU- und Speicherlast zu erzeugen. Zwar skaliert der Hash linear, aber bei vielen parallelen Requests kann dies zu einer DoS-Wirkung beitragen.

**Konkrete Lösung:**  
Ein sinnvolles Limit definieren und bei Überschreitung mit 400 antworten:
```go
if len(user) > 256 {
    writeError(w, http.StatusBadRequest, "user parameter too long")
    return
}
```
Das Limit kann je nach Produktanforderung angepasst werden.

---

### Weitere Empfehlungen (ohne eigenes Finding)

- **Transportverschlüsselung:** Der Dienst lauscht ohne TLS. In Produktion sollte TLS über einen vorgelagerten Proxy oder direkt im Server implementiert werden.
- **Binding einschränken:** Aktuell bindet der Server auf alle Interfaces (`:port`). In unsicheren Umgebungen sollte auf eine interne IP oder `127.0.0.1` gebunden werden.
- **Unbekannte JSON-Felder:** `encoding/json` ignoriert unbekannte Felder standardmäßig – das ist hier akzeptabel.

---

### Bewertung
Keine kritischen oder hohen Risiken, daher keine Blockade. Die identifizierten mittleren und niedrigen Findings sollten vor einem Produktionsrollout behoben werden. Insbesondere die fehlende Authentifizierung und die Server-Timeouts sind aus Sicherheitssicht dringend empfohlen.