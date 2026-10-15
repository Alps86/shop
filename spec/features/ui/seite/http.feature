# language: de
@http
Funktionalität: Öffentliche Shopantwort
  Als Gast erreiche ich eigene Seiten mit sichtbaren Applicationdaten.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Öffentliche Seite über gewöhnliche Navigation
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | <url> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <inhalt> |
      | Header:Vary | HX-Request |
      | Absent | Händlerauswahl |
    Beispiele:
      | url | status | inhalt |
      | /de-DE/ | 200 | Technik, die deinen Alltag |
      | /de-DE/marken | 200 | Markenwelt |
      | /de-DE/service | 200 | Kontakt |
      | /de-DE/kategorie/smartphones--c-phone | 200 | Testmodell &lt;b&gt; |
      | /de-DE/suche?q=phone | 200 | Modelle |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | 200 | data-sku="RED-256" |
      | /de-DE/warenkorb | 200 | Dein Warenkorb ist leer |
      | /de-DE/profil | 200 | Eigenen Einkauf starten |

  Szenariogrundriss: Persönliche Ansichten verwenden geschützte Ports
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | <url> | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <inhalt> |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |
    Beispiele:
      | url | inhalt |
      | /de-DE/warenkorb | RED-256 |
      | /de-DE/warenkorb?add=RED-256 | Zum Warenkorb hinzufügen |
      | /de-DE/kasse | Demo-Bestellung verbindlich abschließen |
      | /de-DE/bestellung/order-a | Unveränderlicher Kaufname |
      | /de-DE/profil | Profil bestätigen |
      | /de-DE/warenkorb/badge | Warenkorb |

  Szenariogrundriss: Lokale Verwaltung bleibt ein separater Listener
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Listener |
      | GET | <url> | <listener> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | <status> |
      | Body | <inhalt> |
      | Header:Cache-Control | <cache> |
    Beispiele:
      | url | listener | status | inhalt | cache |
      | /backend | public | 404 | Nicht gefunden | private,no-store |
      | /backend | backend | 200 | Smartphonefamilie | private,no-store |
      | /backend/bestellung/order-a | backend | 200 | Liefer-/Teillieferung | private,no-store |
