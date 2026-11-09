# Mission

## Zweck

Eine einfache, benutzerfreundliche Webanwendung unterstützt die Verwaltung eines Skiverleihs. Sie gibt Mitarbeitenden einen zuverlässigen Überblick über Ski, Kunden, Ausleihen und Reservierungen.

## Ziel

Der Prototyp deckt zwei fachliche Ausbaustufen für den Ski-Verleih ab:

1. Ski- und Kundenverwaltung mit automatisch vergebenen fortlaufenden IDs, optionalen Ausweisdaten und tageweisen Ausleihen.
2. Vorausreservierungen für flexibel gewählte Starttermine und ganze Verleihtage sowie eine nachvollziehbare Preisberechnung mit konfigurierbarem Tagestarif.

Ein übersichtliches Dashboard zeigt verfügbare, ausgeliehene und gesamte Ski sowie die registrierten Kunden.

Zusätzlich wird ein separates Fahrradverleih-Projekt geplant. Es soll Fahrräder stundenweise verleihen und jede angefangene Stunde voll berechnen. Es wird nicht in die Ski-Verleih-Anwendung integriert.

## Leitlinien

- Prototyp-Fokus: schnell verständlich, leicht änderbar und ohne unnötige Komplexität.
- Backend in Go, Frontend mit HTML, CSS und JavaScript.
- Thread-safe In-Memory-Datenspeicherung mit lokaler JSON-Datei zur Wiederherstellung nach Neustarts; keine externe Datenbank.
- Ausleih- und Preisregeln müssen nachvollziehbar sein; Tagestarif und Stundentarif sind konfigurierbar.
- Kundendokumente sind optional; die manuelle Eingabe von Ausweistyp und Ausweisnummer ist eine Alternative.
- Keine Plugin-Architektur im aktuellen Umfang.

## Akzeptanzkriterien

### Teil a – Skiverleih und Kundenverwaltung

- Ski und Kunden erhalten beim Anlegen automatisch fortlaufende IDs, die nicht manuell vorgegeben werden.
- Ein Kunde kann entweder ein Ausweisdokument hochladen oder Ausweistyp und Ausweisnummer manuell erfassen.
- Ausleihen sind ausschließlich für eine positive Anzahl ganzer Kalendertage möglich; Teil-Tagesdauern werden abgewiesen.
- Verleihen und Zurückgeben aktualisieren den Ski-Status.
- Das Dashboard zeigt korrekte Bestandszahlen für verfügbare, ausgeliehene und alle Ski sowie die Anzahl der Kunden.

### Teil b – Reservierungen und Preisberechnung

- Für ein Ski kann ein zukünftiger Verleihbeginn mit einer positiven Anzahl ganzer Verleihtage reserviert werden.
- Zeitlich überlappende Reservierungen oder Ausleihen desselben Skis werden verhindert.
- Der Preis entspricht `Anzahl Verleihtage × konfigurierbarer Tagestarif`.
- Änderungen an Dauer oder Tarif aktualisieren den Gesamtpreis.

### Teil c – Fahrradverleih-Projektplanung

- Ein eigenständiger Projektplan beschreibt den Fahrradverleih, ohne ihn in die Ski-Verleih-Anwendung zu integrieren.
- Der Plan enthält mindestens Verwaltung, Verfügbarkeit, Ausleih-/Rückgabeablauf, Dashboard und Preisberechnung für Fahrräder.
- Jede angefangene Stunde zählt als volle Stunde; der Preis wird mit `Anzahl angefangener Stunden × konfigurierbarer Stundentarif` berechnet.
- Für Stundenberechnung, Verfügbarkeit, Ausleihe und Rückgabe sind überprüfbare Kriterien angegeben.

## Nicht-Ziele

- Plugin-Architektur (Teil d)
- Dauerhafte Speicherung oder externe Datenbank
- Implementierung des separaten Fahrradverleih-Projekts im Rahmen der Ski-Verleih-Anwendung
- Festlegung konkreter Tarifbeträge, Währung oder Rundungsregeln; diese Werte müssen vor der Preisumsetzung noch bestimmt werden
