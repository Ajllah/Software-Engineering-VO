# Arbeitsanweisungen

- Dieses Verzeichnis enthält das eigenständige Projekt „Fahrrad-Verleih“. Keine Dateien, Pakete, Modulpfade oder Laufzeitdaten aus dem Ski-Verleih übernehmen oder dort ändern.
- Vor Änderungen Anforderungen und Akzeptanzkriterien in `mission.md`, `techstack.md` und `roadmap.md` beachten.
- Backend in Go und Frontend mit HTML, CSS sowie minimalem JavaScript umsetzen. Keine externen Abhängigkeiten oder Datenbanken hinzufügen.
- Fachlogik und Validierung im Backend abbilden. Den In-Memory-Store bei gemeinsamem Zugriff mit einem Mutex schützen.
- Preise als ganzzahlige Centbeträge speichern. Jede angefangene Stunde wird voll berechnet; bestehende Buchungen behalten ihren Tarif und Gesamtpreis.
- Für dasselbe Fahrrad sind nur nicht überlappende Zeiträume zulässig. Buchungen verschiedener Fahrräder desselben Kunden dürfen zeitlich gleichzeitig liegen.
- Eingaben und Persistenzfehler explizit behandeln; keine stillen Erfolgs-Fallbacks einbauen.
- Jede fachliche Änderung mit passenden Go-Tests absichern. Vor Abschluss `gofmt` und `go test ./...` ausführen.
- Der Standardport ist 8081, damit Fahrrad- und Ski-Verleih parallel gestartet werden können.
## Weitere Projektregeln

- Go-Backend mit Structs und Methoden umsetzen; Interfaces nur bei echtem Bedarf verwenden.
- Architektur: `main.go` als Einstieg und HTTP-Server, `internal/domain/` für Fachlogik, `internal/store/` für den thread-sicheren In-Memory-Store und `internal/handler/` für HTTP-Handler.
- Keine unnötigen globalen Variablen und kein Dependency-Injection-Framework verwenden.
- Fehler nicht verschlucken; bei Bedarf mit `fmt.Errorf("%w", err)` wrappen.
- Lokales Git zur Versionierung verwenden; Commit-Messages kurz und im Imperativ formulieren.
- Keine Secrets im Code speichern.
- Frontend und Backend-API klar trennen; API-Aufrufe nur über Go-Endpoints.
- Keine generierten Dateien manuell bearbeiten.
- Bei Unklarheiten nachfragen statt raten.
- Prototyp-Fokus: einfache, verständliche und leicht änderbare Lösungen bevorzugen.