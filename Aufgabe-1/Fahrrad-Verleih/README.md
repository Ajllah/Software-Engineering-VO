# Fahrrad-Verleih

Eigenständige Webanwendung zur Verwaltung eines Fahrradverleihs. Dieses Projekt ist unabhängig vom Ski-Verleih und verwendet einen eigenen Go-Modulpfad, Datenbestand und HTTP-Server.

## Funktionen

- Fahrräder und Kunden mit automatisch vergebenen IDs verwalten.
- Kundendaten einschließlich Ausweistyp und Ausweisnummer erfassen.
- Fahrräder für Zeiträume ab einem frei wählbaren Datum und einer Uhrzeit buchen oder reservieren.
- Rückgaben erfassen und dadurch Fahrräder früher wieder verfügbar machen.
- Überschneidende Buchungen desselben Fahrrads verhindern; ein Kunde darf mehrere Fahrräder gleichzeitig buchen.
- Preis nach angefangenen Stunden mit konfigurierbarem Stundentarif berechnen.
- Dashboard mit Bestands-, Kunden- und Buchungsübersicht.
- Thread-sicherer In-Memory-Store mit optionaler lokaler JSON-Persistenz.

Die Preisformel lautet `aufgerundete Mietdauer in Stunden × Stundentarif`. Eine Mietdauer von 1 Stunde und 10 Minuten wird somit als 2 Stunden berechnet. Der Standardtarif beträgt 5,00 € pro Stunde; Änderungen gelten nur für neue Buchungen.

## Start

Voraussetzung ist Go 1.21 oder neuer. Im Projektordner:

```powershell
go run .
```

Die Anwendung ist unter <http://localhost:8081> erreichbar. Der Server lauscht standardmäßig ausschließlich auf `localhost:8081`. Mit `ADDR` kann die Adresse überschrieben werden.

Optionale Konfiguration:

- `HOURLY_RATE_EUR`: anfänglicher Stundentarif, z. B. `5.00`. Verwendet, wenn noch keine Datendatei existiert.
- `DATA_FILE`: Pfad zur JSON-Datendatei. Standard: `data.json` im Projektordner.
- `ADDR`: HTTP-Adresse. Standard: `localhost:8081`.

Die JSON-Datei enthält personenbezogene Ausweisdaten und sollte vertraulich behandelt werden. Sie wird nicht in Git eingecheckt.

## Tests

```powershell
go test ./...
```

## Projektstruktur

- `main.go`: Konfiguration, Serverstart und Einbindung der Webdateien.
- `internal/domain`: Fachliche Datenmodelle.
- `internal/store`: Thread-sicherer Geschäfts- und Persistenzspeicher.
- `internal/handler`: HTTP-API.
- `web`: eigenständige HTML-, CSS- und JavaScript-Oberfläche.

Dieses Projekt teilt weder Go-Pakete noch Daten oder Konfiguration mit dem Ski-Verleih.
