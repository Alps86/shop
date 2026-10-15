# language: de
@http @frontend
Funktionalität: Zusammenhängende Variantenansicht
  Als Käufer wechsle ich Galerie Preis Verfügbarkeit und Pflichtdaten als eine Einheit.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |
    Und HTTPportdaten werden ersetzt
      | JSON |
      | {"Detail":{"Cards":[{"ProductID":"p-phone","URL":"/de-DE/produkt/testmodell--p-phone","Name":"Testmodell","SKU":"RED-256","Image":{"URL":"/assets/product.svg","AltText":{"de-DE":"gallery-red"}},"UnitPrice":{"MinorUnits":12000,"Currency":"EUR"},"AvailableQuantity":7,"VariantCount":2,"Compliance":{"Context":{"Market":"DE","Locale":"de-DE","VerificationStatus":"verified"},"Notices":[{"Text":{"de-DE":"verified-red"}}]}}]}} |

  Szenariogrundriss: HTML und HTMX zeigen dieselbe vollständige rote Variante
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | HTMX |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 | <htmx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | BodyAll | data-sku="RED-256",gallery-red,120&#44;00 EUR,verified-red,Zubehör und Empfehlungen,hx-push-url="true" |
      | Header:Vary | HX-Request |
      | Absent | verified-blue |
    Beispiele:
      | htmx |
      | false |
      | true |

  Szenario: Eine Karte je Modell zeigt passende SKU und weitere Varianten
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | ?sku=RED-256,2 Varianten,storageCapacity,verified-blue |
      | Absent | MarginSignal |
      | Absent | InterestWeight |
