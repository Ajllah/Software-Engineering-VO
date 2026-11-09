# Technischer Stack

## Laufzeit und Abhängigkeiten

- Go 1.21 oder neuer
- Ausschließlich Go-Standardbibliothek; keine externe Datenbank und möglichst keine zusätzlichen Module
- `net/http` für API und Webserver
- Statische Webdateien werden mit `embed` eingebettet

## Architektur

- `main.go`: Konfiguration, Einbettung des Webverzeichnisses und HTTP-Server auf `localhost:8081`
- `internal/domain`: Fahrräder, Kunden, Buchungen und Dashboardmodelle
- `internal/store`: Validierung, Buchungsregeln, Preisberechnung, Synchronisierung und optionale Persistenz
- `internal/handler`: JSON-HTTP-Endpunkte und HTTP-Fehlerabbildung
- `web`: barrierearme, responsive Oberfläche mit HTML, CSS und minimalem JavaScript

## Fachliche Datenregeln

- IDs werden durch den Store erzeugt, nicht durch den Client.
- Preise werden als ganzzahlige Centbeträge gespeichert; die Mietdauer wird für die Preisermittlung auf angefangene Stunden aufgerundet.
- Eine Buchung speichert ihren Tarif und Gesamtpreis zum Buchungszeitpunkt.
- Buchungsintervalle gelten als halboffen `[Beginn, Ende)`. Ein Ende genau zum Beginn einer Folgebuchung ist daher konfliktfrei.
- Eine Rückgabe ersetzt für die Verfügbarkeitsprüfung das geplante Ende durch den tatsächlichen Rückgabezeitpunkt.
- Ein Mutex schützt Lese- und Schreibzugriffe sowie die Konfliktprüfung und Buchungserstellung als eine atomare Operation.

## Konfiguration und Persistenz

- `ADDR`: Serveradresse, Standard `localhost:8081`
- `HOURLY_RATE_EUR`: anfänglicher Stundentarif, Standard `5.00`
- `DATA_FILE`: JSON-Dateipfad, Standard `data.json`
- Die JSON-Datei ist optional wählbar und standardmäßig aktiviert; Schreibvorgänge erfolgen atomar über eine temporäre Datei und Umbenennung.
- Die Datendatei enthält personenbezogene Angaben und ist von Git auszuschließen.

## Qualitätssicherung

- Unit-Tests für Store-Regeln, Zeitüberschneidungen, Preise, Rückgabe und Wiederherstellung.
- HTTP-Tests für Eingabevalidierung, Statuscodes und API-Abläufe.
- Validierung mit `gofmt` und `go test ./...`.
