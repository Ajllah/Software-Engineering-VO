# Ski-Verleih

Ein einfacher Webapp-Prototyp zur Verwaltung eines Skiverleihs. Das Dashboard und die Arbeitsabläufe sollen schnell verständlich und leicht änderbar sein.

## Umfang

### Teil a – Skiverleih und Kundenverwaltung

- Ski mit Typ und automatisch vergebener, fortlaufender ID verwalten; Ski hinzufügen und bearbeiten.
- Kunden registrieren und verwalten. Kunden erhalten automatisch eine fortlaufende ID.
- Bei der Registrierung Ausweistyp und Ausweisnummer manuell eingeben.
- Ski an Kunden verleihen und Rückgaben erfassen. Ski können ausschließlich für ganze Kalendertage ausgeliehen werden.
- Dashboard mit verfügbaren, ausgeliehenen und gesamten Ski sowie der Anzahl registrierter Kunden.

### Teil b – Reservierungen und Preise

- Ski für einen zukünftigen Zeitraum reservieren.
- Verleihbeginn flexibel auswählen; die Dauer wird in ganzen Kalendertagen festgelegt.
- Den Preis anhand eines konfigurierbaren Tagestarifs berechnen: Anzahl Verleihtage × Tagestarif.
- Reservierungen dürfen sich für dasselbe Ski nicht zeitlich überschneiden.

### Teil c – Separates Fahrradverleih-Projekt planen

Ein eigenständiges Fahrradverleih-Projekt soll geplant werden. Fahrräder sollen stundenweise ausgeliehen werden; jede angefangene Stunde zählt als volle Stunde. Der Projektplan soll Funktionen, Abläufe, Preisberechnung und Akzeptanzkriterien festhalten. Die Umsetzung des Fahrradprojekts ist nicht Teil dieses Ski-Verleih-Prototyps.


## Aktueller Implementierungsstand

Teil a und b sind als lokaler Prototyp umgesetzt. Ski und Kunden erhalten automatisch fortlaufende IDs. Die Kundenerfassung unterstützt derzeit die manuelle Eingabe des Ausweistyps und der Ausweisnummer; ein Dokument-Upload ist nicht enthalten. Ski können tageweise verliehen, zurückgegeben und für zukünftige Zeiträume reserviert werden. Überlappende Zeiträume desselben Skis werden abgewiesen. Der Tagestarif beträgt standardmäßig 20,00 Euro, kann über das Dashboard geändert oder beim ersten Start per `DAILY_RATE_EUR` konfiguriert werden. Der Gesamtpreis wird für jede Buchung gespeichert. Ski, Kunden, Ausleihen, Reservierungen und Tarif werden lokal in `data.json` gespeichert und beim Start geladen.

Teil c wurde als eigenständiges Fahrradverleih-Projekt umgesetzt und befindet sich im separaten Ordner Aufgabe-1/Fahrrad-Verleih. Das Ski-Verleih-Projekt umfasst die Teile a) und b).

## Akzeptanzkriterien

### Teil a

- Beim Anlegen eines Skis werden Nummer/ID und Typ gespeichert. Die ID wird automatisch vergeben, ist fortlaufend und kann bei der Erfassung nicht manuell gesetzt werden.
- Beim Anlegen eines Kunden wird eine automatisch vergebene, fortlaufende ID angezeigt; Kundendaten können anschließend eingesehen und verwaltet werden.
- Eine Kundenregistrierung ist mit hochgeladenem Ausweisdokument oder alternativ mit manuell eingegebenem Ausweistyp und Ausweisnummer möglich.
- Eine Ausleihe mit weniger als einem ganzen Kalendertag oder mit einer Teil-Tagesdauer wird abgewiesen.
- Eine verfügbare Ski kann einem Kunden ausgeliehen und als zurückgegeben erfasst werden; der Verfügbarkeitsstatus aktualisiert sich entsprechend.
- Dashboard-Zahlen für verfügbare, ausgeliehene und gesamte Ski sowie registrierte Kunden stimmen mit den gespeicherten Datensätzen überein.

### Teil b

- Ein Ski kann für einen zukünftigen Beginn und eine positive Anzahl ganzer Verleihtage reserviert werden.
- Eine Reservierung wird abgewiesen, wenn sich ihr Zeitraum mit einer bestehenden Reservierung oder Ausleihe desselben Skis überschneidet; nicht überlappende Zeiträume bleiben reservierbar.
- Für eine gültige Reservierung oder Ausleihe wird der Preis als Anzahl Verleihtage multipliziert mit dem konfigurierten Tagestarif berechnet.
- Bei einer Änderung von Dauer oder Tagestarif wird der angezeigte Gesamtpreis entsprechend neu berechnet.

### Teil c

- Es liegt ein vom Ski-Verleih abgegrenzter Projektplan für einen Fahrradverleih vor.
- Der Plan beschreibt mindestens Fahrrad- und Kundenverwaltung, stundenweise Ausleihe, Verfügbarkeit, Reservierung beziehungsweise Ausleihe und Rückgabe sowie ein Dashboard.
- Der Plan legt fest, dass jede angefangene Stunde voll berechnet wird, und enthält die Formel `Anzahl angefangener Stunden × konfigurierbarer Stundentarif`.
- Der Plan enthält überprüfbare Kriterien für die Stundenberechnung, Verfügbarkeitsprüfung, Ausleihe und Rückgabe.
- Das Fahrradprojekt wird nicht in die Ski-Verleih-Anwendung eingebaut.

## Technischer Rahmen

- Backend: Go
- Frontend: HTML, CSS und JavaScript
- Datenspeicherung: thread-safe In-Memory-Store mit lokaler JSON-Datei als dauerhafter Speicherung; keine externe Datenbank
- Keine neuen Abhängigkeiten ohne vorherige Abstimmung

Der Ski-Tagestarif startet bei 20,00 Euro. Ein im Dashboard geänderter Tarif gilt für neue Buchungen; bereits angelegte Buchungen behalten ihren gespeicherten Tarif und Gesamtpreis. Der Starttarif kann mit `DAILY_RATE_EUR` überschrieben werden. Der Fahrrad-Stundentarif ist noch festzulegen. Ausweisdaten sind personenbezogen und werden nur für die Kundenverwaltung angezeigt. Der thread-safe In-Memory-Store verwaltet den Laufzeitbestand; Änderungen werden zusätzlich in einer lokalen JSON-Datei gesichert. Eine externe Datenbank wird nicht verwendet.

## Entwicklung

- Backend starten: `go run .` oder `go run ./cmd/server`
- Tests: `go test ./...`
- Formatierung: `gofmt -w .` oder `go fmt ./...`
- Optionaler Lint: `golangci-lint run`

### Lokal unter Windows starten

1. Go 1.21 oder neuer installieren.
2. PowerShell im Projektordner öffnen.
3. `go run .` ausführen.
4. Im Browser `http://localhost:8080` öffnen.
5. Den Server mit `Ctrl+C` beenden.

Falls PowerShell `go` nicht erkennt, den Go-Installationsordner zum `PATH` hinzufügen oder `go.exe` über seinen vollständigen Pfad starten.

Optional kann der Starttarif über eine Umgebungsvariable gesetzt werden, zum Beispiel in PowerShell mit `$env:DAILY_RATE_EUR = "20.00"` vor `go run .`. Der Tarif lässt sich auch im Dashboard ändern. Mit `DATA_FILE` kann alternativ ein anderer Speicherort für die JSON-Datei angegeben werden; standardmäßig wird `data.json` im Projektordner verwendet.

Die HTTP-API wird vom Frontend verwendet. Ski, Kunden, Ausleihen und Reservierungen lassen sich über das Dashboard anlegen beziehungsweise verwalten. Die Datendatei `data.json` wird im Projektordner angelegt und sollte nicht gelöscht werden, wenn die Daten erhalten bleiben sollen. Sie ist in Git ignoriert, da sie personenbezogene Kundendaten enthalten kann. Bei einer vorhandenen Datendatei werden gespeicherte Daten einschließlich Tarif geladen; `DAILY_RATE_EUR` wird als Startwert nur verwendet, wenn die Datei noch nicht vorhanden ist.
