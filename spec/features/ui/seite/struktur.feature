# language: de
@http @frontend
Funktionalität: Schlichte Shopseiten nach EP-Grundkonzept
  Als Gast finde ich alle vereinbarten Seitenteile in einer eigenen schlanken Storefront.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Echte HTTPansichten verwenden vollständige öffentliche DTOs
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | Listener |
      | GET | <route> | <session> | <listener> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | <komponenten> |
      | Absent | Händlerauswahl |
      | Absent | accountLogin |
      | Absent | realPayment |
    Beispiele:
      | route | session | listener | komponenten |
      | /de-DE/ | | public | class="logo",Hauptnavigation,name="q",Warenkorb,hero,Ausgewählte Angebote,Markenwelt,Service,footer |
      | /de-DE/kategorie/smartphones--c-phone | | public | Breadcrumb,Unterkategorien,Auswahl eingrenzen,Sortieren,Ansicht,Seiten,class="card",SKU RED-256 |
      | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | | public | gallery,Alpha,Testmodell,SKU RED-256,Farbe und Speicher,color,storageCapacity,120&#44;00 EUR,Bruttopreis,Versandkosten,In den Warenkorb,Details und Technik,Hersteller,Zubehör und Empfehlungen |
      | /de-DE/warenkorb | test-session-a | public | Simulation – keine echte Bestellung/Zahlung,Menge ändern,Entfernen,Aktuelle Bewertung,Zubehör,Zum Gastcheckout |
      | /de-DE/kasse | test-session-a | public | Simulation – keine echte Bestellung/Zahlung,Lieferadresse,Standard,Express,Rechnung,Kreditkarte,Aktuelle Bewertung,ausdrücklich bestätigen,Demo-Bestellung verbindlich abschließen |
      | /de-DE/bestellung/order-a | test-session-a | public | Simulation – keine echte Bestellung/Zahlung,Unveränderlicher Kauf,Lieferfortschritt,gekauft 1,versandt 1,geliefert 0 |
      | /backend | | backend | Katalog,Familien,Kategorien,Aktionen,Preis und Bestand,Bestellungen,Keine Bestellungen |
      | /backend/bestellung/order-a | | backend | Unveränderlicher Kauf,Liefer-/Teillieferung,Lieferkennung,Teilmenge,Lieferfortschritt melden |
      | /de-DE/service | | public | Service und Kontakt,Kontakt,footer |

  Szenario: Leerer eigener Cart bietet einen ausdrücklichen Start
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/warenkorb |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | Dein Warenkorb ist leer,Eigenen Einkauf starten,Weiter einkaufen,Simulation – keine echte Bestellung/Zahlung |
      | Absent | Demo-Bestellung verbindlich abschließen |
