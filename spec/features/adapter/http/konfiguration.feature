# language: de
@http
Funktionalität: Sichere Adapterkonfiguration und begrenzte Eingaben
  Als Betreiber erkenne ich fehlerhafte Vorlagen und unzureichende Schlüssel vor der Nutzung.

  Szenariogrundriss: Unsichere Konfiguration startet keinen Handler
    Angenommen die HTTPkonfiguration verwendet
      | CSRFKey | Template |
      | <schluessel> | <vorlage> |
    Wenn der HTTPadapter wird initialisiert
      | Listener |
      | <listener> |
    Dann die Adapterinitialisierung meldet
      | Error |
      | <fehler> |
    Beispiele:
      | schluessel | vorlage | listener | fehler |
      | short | {{define "page"}}ok{{end}} | public | mindestens 32 Bytes |
      | short | {{define "page"}}ok{{end}} | backend | mindestens 32 Bytes |
      | synthetic-configuration-key-32-bytes | {{define | public | shoptemplates |

  Szenario: Nicht ausführbare Vorlage liefert persönliche sichere Fehlerseite
    Angenommen die HTTPkonfiguration verwendet
      | Template |
      | {{define "page"}}{{.MissingField}}{{end}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/ |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Cache-Control | private,no-store |
      | Body | Ansicht vorübergehend nicht verfügbar |

  Szenariogrundriss: Grenzen werden am tatsächlich eingehenden Request geprüft
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Append | Count |
      | GET | /de-DE/suche?q= | a | <anzahl> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
    Beispiele:
      | anzahl | status |
      | 101 | 422 |
      | 2050 | 414 |

  Szenario: Kaputte Formularcodierung wird sichtbar abgewiesen
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | RawBody |
      | POST | /de-DE/gast | bad=%xx |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |
      | Body | Formular ungültig |
