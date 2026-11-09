# Techstack und Architektur

## Technologieauswahl

- **Backend:** Go
- **Frontend:** HTML, CSS und JavaScript
- **Laufzeit-Datenspeicherung:** thread-safe In-Memory-Store
- **Dauerhafte lokale Speicherung:** JSON-Datei `data.json`, atomar aktualisiert; keine externe Datenbank
- **Versionierung:** lokales Git, bei Bedarf GitHub

Das HTML-Frontend entspricht der im README festgelegten Variante. Flutter/Dart ist eine mögliche allgemeine Alternative laut CLAUDE.md, wird für diesen Prototyp aber nicht eingeplant.

## Architektur

Die Go-Anwendung folgt einer einfachen Trennung der Verantwortlichkeiten:

- `main.go` oder `cmd/server/`: Anwendungseinstieg, HTTP-Server und Routing
- `internal/domain/`: zentrale Datenstrukturen und Methoden für die Geschäftslogik
- `internal/store/`: thread-safe In-Memory-Verwaltung, beispielsweise mit Map und Mutex; lädt und sichert den Zustand in `data.json`
- `internal/handler/`: HTTP-Handler, die Domain-Logik aufrufen
- Frontend-Dateien: HTML/CSS/JavaScript mit klarer Trennung von den Go-API-Aufrufen

Das Frontend greift ausschließlich über die Go-Endpoints auf Anwendungsdaten zu, nicht direkt auf interne Store-Strukturen. Globale Variablen werden vermieden; ein Store-Singleton ist nur dann vorgesehen, wenn er tatsächlich nötig ist.

## Konventionen

- Structs und Methoden verwenden, wenn dadurch Verhalten sinnvoll an Daten gebunden wird.
- Interfaces nur bei konkretem Bedarf einführen, etwa für einen austauschbaren Store.
- Kein Dependency-Injection-Framework und keine neuen Dependencies ohne vorherige Abstimmung.
- Fehler nicht verschlucken; beim Wrappen `fmt.Errorf("%w", err)` verwenden.
- Keine Secrets in den Quellcode aufnehmen.
- Keine generierten Dateien von Hand bearbeiten.

## Entwicklung und Qualitätssicherung

- Backend starten: `go run .` oder `go run ./cmd/server`
- Tests ausführen: `go test ./...`
- Go-Code formatieren: `gofmt -w .` oder `go fmt ./...`
- Optional linten, falls eingerichtet: `golangci-lint run`
- Das HTML-Frontend kann direkt geöffnet oder über einen kleinen Static-Server bereitgestellt werden.

## Erweiterbarkeit

Die Preisberechnung ist eine spätere Erweiterung. Die erste Version soll Ausleihbeginn und -dauer erfassen, aber keine Preislogik voraussetzen oder vorwegnehmen.
