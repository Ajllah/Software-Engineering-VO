# Mission

## Zweck

Ein einfaches, zuverlässiges Web-Dashboard unterstützt Mitarbeitende bei der Verwaltung eines Fahrradverleihs. Es macht Fahrradbestand, Kunden, Buchungen, Reservierungen und Rückgaben übersichtlich verfügbar.

## Umfang

- Fahrräder und Kunden erfassen, anzeigen und bearbeiten. IDs werden automatisch fortlaufend vergeben.
- Kundendaten umfassen Name, Ausweistyp und Ausweisnummer.
- Eine Buchung enthält ein Fahrrad, einen Kunden, einen Startzeitpunkt mit Datum und Uhrzeit sowie ein frei wählbares Ende.
- Buchungen für zukünftige Zeiträume gelten als Reservierungen; begonnene Zeiträume sind Ausleihen.
- Eine Rückgabe beendet die aktive Nutzung und gibt das Fahrrad ab dem tatsächlichen Rückgabezeitpunkt wieder frei.
- Für ein Fahrrad dürfen sich Buchungszeiträume nicht überschneiden. Zeiträume verschiedener Fahrräder sind unabhängig; ein Kunde darf mehrere Fahrräder gleichzeitig buchen.
- Die Abrechnung erfolgt pro angefangener Stunde: `ceil(Dauer in Minuten / 60) × Stundentarif`. Beispielsweise zählen 70 Minuten als 2 Stunden.
- Der Stundenpreis ist einstellbar. Eine Änderung beeinflusst nur neu angelegte Buchungen.
- Das Dashboard zeigt Fahrradbestand, Verfügbarkeit, Kunden sowie anstehende und aktive Buchungen.

## Technische Leitlinien

- Eigenständiges Go-Modul, getrennt vom Ski-Verleih.
- Weboberfläche mit HTML, CSS und minimalem JavaScript.
- Thread-sicherer In-Memory-Store ohne externe Datenbank.
- Optionale lokale JSON-Datei zur Erhaltung des Zustands bei Neustarts.
- Keine externen Abhängigkeiten, wenn die Go-Standardbibliothek ausreicht.
- HTTP-Standardadresse `localhost:8081`.

## Akzeptanzkriterien

- Fahrrad- und Kunden-IDs werden automatisch vergeben und sind fortlaufend.
- Name, Ausweistyp und Ausweisnummer sind für Kunden erforderlich.
- Eine Buchung mit ungültigem Fahrrad/Kundenbezug oder einem Ende vor beziehungsweise gleich dem Beginn wird abgewiesen.
- Die Dauer von 1 Stunde 10 Minuten ergibt bei 5,00 € pro Stunde genau 10,00 €.
- Zeitlich überlappende Buchungen für dasselbe Fahrrad werden auch bei gleichzeitig eintreffenden Anfragen verhindert.
- Nicht überlappende Buchungen desselben Fahrrads sowie parallele Buchungen unterschiedlicher Fahrräder durch denselben Kunden sind möglich.
- Eine Rückgabe wird für eine aktive Ausleihe erfasst und macht das Fahrrad für nachfolgende Zeiträume verfügbar.
- Dashboard-Zähler stimmen mit dem Bestand zum aktuellen Zeitpunkt überein.
- Der Tarif und Buchungen bleiben bei aktivierter JSON-Persistenz nach einem Neustart erhalten.
- `go test ./...` läuft erfolgreich; der Webserver startet standardmäßig auf Port 8081.

## Nicht-Ziele

- Integration in den Ski-Verleih oder gemeinsame Datenhaltung
- Externe Datenbank, Login-/Rollenverwaltung oder Online-Zahlung
- Hochladen oder Speichern von Ausweisdokumenten
