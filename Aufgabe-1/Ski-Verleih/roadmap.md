# Roadmap

Diese Roadmap dokumentiert Anforderungen und Umsetzungsstand. Teile a und b des Ski-Verleih-Prototyps sind umgesetzt; Teil c bleibt ein separates Planungsvorhaben. Eine Plugin-Architektur (Teil d) ist nicht erforderlich.

## Teil a – Kern des Ski-Verleihs

### Umsetzung

- Go-HTTP-Anwendung und HTML/CSS/JavaScript-Oberfläche gemäß `techstack.md` aufsetzen.
- Domain-Modelle für Ski, Kunden und Ausleihen erstellen.
- Thread-safe In-Memory-Store mit lokaler JSON-Datei für Neustart-Persistenz einrichten.
- Fortlaufende IDs automatisch vergeben, mindestens für Ski und Kunden.
- Kundenverwaltung mit optionalem Ausweisdokument oder manueller Eingabe von Ausweistyp und Ausweisnummer umsetzen.
- Ausleihe auf ganze Kalendertage begrenzen; Verleihen und Rückgabe abbilden.
- Dashboard für Ski-Bestand und registrierte Kunden erstellen.

### Akzeptanzkriterien

- Beim Erstellen von Ski und Kunden werden jeweils fortlaufende IDs automatisch erzeugt und die Formulare bieten keine manuelle ID-Vergabe.
- Kundenerfassung funktioniert wahlweise mit Ausweisdokument oder mit manuell eingegebenem Ausweistyp und Ausweisnummer.
- Eine Ausleihe für null Tage, negative Dauer oder einen Teil eines Tages wird abgewiesen; positive ganze Tagesdauern werden akzeptiert.
- Erfolgreiche Ausleihe und Rückgabe ändern den Verfügbarkeitsstatus des betroffenen Skis.
- Dashboard-Werte für verfügbare, ausgeliehene und gesamte Ski sowie registrierte Kunden entsprechen den zugrunde liegenden Datensätzen.

## Teil b – Reservierungen und flexible Preisberechnung

**Status: umgesetzt.** Zukünftige Reservierungen können über das Dashboard angelegt werden; der beispielhafte Tagestarif beträgt 20,00 Euro und ist im Dashboard oder beim Serverstart konfigurierbar.

### Umsetzung

- Beginn einer Reservierung flexibel als zukünftiges Datum festlegen können.
- Dauer als positive Anzahl ganzer Kalendertage erfassen.
- Zeitraumverfügbarkeit pro Ski prüfen und Überschneidungen mit bestehenden Reservierungen oder Ausleihen verhindern.
- Tagestarif konfigurierbar machen und Gesamtpreis aus Dauer und Tarif berechnen.

### Akzeptanzkriterien

- Eine gültige Reservierung mit zukünftigem Beginn und positiver ganzer Tagesdauer kann erfasst werden.
- Eine Reservierung desselben Skis wird abgewiesen, wenn ihr Zeitraum mit einer anderen Reservierung oder Ausleihe überlappt; ein nicht überlappender Zeitraum wird akzeptiert.
- Der angezeigte Preis entspricht exakt `Anzahl Verleihtage × Tagestarif`.
- Eine Änderung des Tagestarifs oder der Verleihtage führt zur aktualisierten Preisberechnung.

## Qualitätssicherung für Teile a und b

- Domain- und Store-Tests für ID-Vergabe, Kundenregistrierung, Tagesvalidierung, Statuswechsel, Konfliktprüfung und Preisformel schreiben.
- Backend-Tests mit `go test ./...` ausführen.
- Go-Code gemäß `techstack.md` formatieren; bei vorhandener Einrichtung `golangci-lint run` verwenden.
- Dashboard, Kundenregistrierung, Reservierung, Ausleihe und Rückgabe manuell über die Oberfläche prüfen.

## Teil c – Separates Fahrradverleih-Projekt planen

Der nächste Planungsschritt ist ein eigener, vom Ski-Verleih abgegrenzter Projektplan. Er soll vor einer möglichen Umsetzung mindestens folgende Themen und Kriterien festhalten:

- Fahrräder, Kunden, automatische IDs, Verfügbarkeit, Ausleihe und Rückgabe beschreiben.
- Stundenweise Ausleihe spezifizieren; Dauer muss positiv sein und jede angefangene Stunde zählt voll.
- Preisformel dokumentieren: `Anzahl angefangener Stunden × konfigurierbarer Stundentarif`.
- Dashboard-Anforderungen und Akzeptanzkriterien für Verfügbarkeit, Stundenberechnung, Ausleihe und Rückgabe definieren.
- Vor einer Umsetzung prüfen, ob der technische Rahmen aus `techstack.md` übernommen wird; das Fahrradprojekt bleibt ein separates Projekt.

**Abnahmekriterium für die Planungsaufgabe:** Ein eigenständiger Projektplan liegt vor und enthält die genannten Anforderungen, Preisregel und überprüfbaren Akzeptanzkriterien. Das Fahrradverleih-System selbst wird in dieser Roadmap nicht implementiert.

## Noch zu klären

- Konkreter Fahrrad-Stundentarif sowie dessen Währung und Rundungsregeln
- Technische und fachliche Behandlung hochgeladener Ausweisdokumente, einschließlich erlaubter Dateiformate und Größen
- Ob Kunden neben Kundenverwaltung auch gelöscht werden dürfen
