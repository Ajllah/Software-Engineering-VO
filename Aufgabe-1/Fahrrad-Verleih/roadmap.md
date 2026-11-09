# Roadmap

## Phase 1 – Anforderungen und Projektbasis

- Eigenständiges Go-Modul und SDD-Dokumentation erstellen.
- Standardbibliotheksbasierte Paketstruktur und eingebettete Weboberfläche vorbereiten.
- Serverkonfiguration mit Standardport 8081 festlegen.

**Abnahme:** Das Modul ist vom Ski-Verleih getrennt und verwendet einen eigenen Datenpfad.

## Phase 2 – Fachmodelle und Store

- Fahrräder, Kunden, Buchungen und Dashboard-Zusammenfassung modellieren.
- Automatische IDs und Kunden-/Fahrradverwaltung implementieren.
- Thread-sicheren Store mit Zeitüberschneidungsprüfung erstellen.
- Preisberechnung pro angefangener Stunde und Tarifänderung umsetzen.
- JSON-Wiederherstellung und persistente Änderungen implementieren.

**Abnahme:** Tests belegen IDs, Validierung, Mietpreis, parallele Kundenbuchungen, Konfliktverhinderung, Rückgabe und Persistenz.

## Phase 3 – HTTP-API und Oberfläche

- API für Dashboard, Fahrräder, Kunden, Buchungen, Rückgabe und Tarif bereitstellen.
- Responsives Dashboard mit Bestandsanzeige, Verwaltungsformularen, Buchung und Rückgabe erstellen.
- Zukunftsstarts als Reservierungen und laufende Zeiträume als Ausleihen darstellen.

**Abnahme:** Die fachlichen Abläufe sind über Browser und HTTP-API nutzbar; Fehler werden verständlich gemeldet.

## Phase 4 – Verifikation

- `gofmt` und `go test ./...` ausführen.
- API und Benutzeroberfläche mit Standardport 8081 prüfen.
- Parallelbetrieb mit dem Ski-Verleih anhand getrennter Datenpfade bestätigen.

**Abnahme:** Tests erfolgreich, keine externen Abhängigkeiten, keine gemeinsame Projekt- oder Datenbasis.
