# language: de
@http
Funktionalität: Sichtbare Datenzustände und unveränderte Vorschau
  Als Gast erkenne ich leere gesperrte und nur vorgeschlagene Angebote.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenario: Nulltreffer erhalten die ausgewählten Facetten und den Reset
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Search":{"Cards":[],"TotalProducts":0}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?brand=Alpha |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | 0 Modelle,Keine Treffer,Alpha,Filter zurücksetzen,aria-busy="false",Ergebnisse werden aktualisiert |

  Szenario: Pflichtdatenpending zeigt keinen Kaufbutton
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"Product":{"Manufacturer":{"Name":"","PostalAddress":{"Street":"","PostalCode":"","City":"","Country":""},"ElectronicAddress":""}},"Cards":[{"SKU":"RED-256","UnitPrice":{"MinorUnits":12000,"Currency":"EUR"},"AvailableQuantity":0,"Compliance":{"Context":{"VerificationStatus":"pending"}}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | Pflichtangaben werden geprüft |
      | Absent | Testhersteller |
      | Absent | Rechtskonformität bestätigt |
      | Absent | Sicherheitsgarantie |
      | Absent | In den Warenkorb |

  Szenario: Ausverkaufte gewählte Variante erhält aktive Variantenlinks
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"Cards":[{"SKU":"RED-256","AvailableQuantity":0}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | Derzeit nicht verfügbar,?sku=BLUE-128,?sku=RED-256 |
      | Absent | In den Warenkorb |

  Szenario: Zubehörvorschau zeigt nur die übergebene Bewertung und bleibt GET
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/warenkorb?vorschau=CASE | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | Unverbindliche Zubehör- und Rabattvorschau,135&#44;00 EUR,rule-pair,5&#44;00 EUR,Zubehör übernehmen,Zubehör ablehnen |
      | Header:Cache-Control | private,no-store |
    Und der verwendete Applicationport erhält
      | Operation |
      | |

  Szenariogrundriss: Unverfügbare optionale Vorschau beschädigt keinen Cart
    Angenommen die HTTPkonfiguration verwendet
      | Preview |
      | <port> |
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | <fehler> |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/warenkorb?vorschau=CASE | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Body | <hinweis> |
    Beispiele:
      | port | fehler | hinweis |
      | absent | | Vorschau vorübergehend nicht verfügbar |
      | present | PreviewAccessory | Daten vorübergehend nicht verfügbar |

  Szenario: Erhaltene ungültige Menge bleibt editierbar und derselbe Auftrag wiederholbar
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Change:InvalidInput |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/warenkorb | cartID=cart-a&revision=2&commandID=retry-a&sku=RED-256&quantity=0&action=setQuantity |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 422 |
      | BodyAll | name="quantity" value="0",name="commandID" value="retry-a",Erneut versuchen |

  Szenario: Bereits bestellter Cart führt zur bestehenden Bestellung
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Cart":{"OrderID":{"Present":true,"Value":"order-a"}}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/kasse | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/bestellung/order-a |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |

  Szenario: Leerer Cart führt vor dem Abschluss zurück zur Auswahl
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Cart":{"Cart":{"Items":[]}}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/kasse | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/warenkorb |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |

  Szenario: Ausstehender Gastport liefert einen lesbaren Wiederaufnahmeort
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Open:Pending |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/profil | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 202 |
      | Header:Location | /de-DE/warenkorb |
      | Header:Cache-Control | private,no-store |
      | Header:X-Robots-Tag | noindex,nofollow |
      | Header:Retry-After | 2 |

  Szenario: Abgelaufener Checkout erlaubt nur den ausdrücklichen Neustart
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | GetCart:Expired |
    Und HTTPportdaten werden ersetzt
      | JSON |
      | {"Cart":{"Actions":["startCart"],"Cart":{"State":"expired"}}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/kasse | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 410 |
      | Body | Neuen Einkauf starten |
      | Absent | Adresse speichern |
      | Absent | Demo-Bestellung verbindlich abschließen |
