# language: de
@http
Funktionalität: Durchgängiger eigener Gastpfad über HTTP
  Die expliziten Applicationportdaten isolieren hier Navigation und Formulare.
  Die reale Transaktion und Prozessintegration bleiben Runtimeverträge.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenario: Veröffentlichte Navigation führt über Kategorie und Modell zur gewählten Variante
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/ |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | href="/de-DE/kategorie/smartphones--c-phone" |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?brand=Alpha |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Body | href="/de-DE/produkt/testmodell--p-phone?sku=RED-256" |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | data-sku="RED-256",href="/de-DE/warenkorb?add=RED-256" |

  Szenario: Normales Formular leitet nach Cartänderung zur lesenden Ansicht
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/warenkorb | cartID=cart-a&commandID=navigation-add&revision=2&sku=RED-256&quantity=1&action=add |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 303 |
      | Header:Location | /de-DE/warenkorb |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | /de-DE/warenkorb | test-session-a |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | RED-256,href="/de-DE/kasse" |

  Szenario: Teilbare Filteradresse wird nach Historywiederherstellung vollständig gerendert
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | HTMX |
      | GET | /de-DE/kategorie/smartphones--c-phone?brand=Alpha&brand=Beta | true |
    Dann der verwendete Applicationport erhält
      | Brands | Page |
      | Alpha,Beta | 1 |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/kategorie/smartphones--c-phone?brand=Alpha&page=2 |
    Dann der verwendete Applicationport erhält
      | Brands | Page |
      | Alpha | 2 |
    Und die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | BodyAll | <!doctype html>,historyRestoreAsHxRequest,hx-push-url="true" |

  Szenariogrundriss: Vollständige Variantenlinkziele aus B bleiben unverändert
    Angenommen HTTPportdaten werden ersetzt
      | JSON |
      | {"Recommendations":{"Cards":[{"ProductID":"p-phone","Name":"Empfohlenes Modell","SKU":"RED-256","URL":"/de-DE/produkt/testmodell--p-phone?sku=RED-256"}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session |
      | GET | <source> | <session> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | href="/de-DE/produkt/testmodell--p-phone?sku=RED-256" |
      | Absent | ?sku=RED-256?sku= |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | data-sku="RED-256" |
    Beispiele:
      | source | session |
      | /de-DE/ | |
      | /de-DE/suche | |
      | /de-DE/kategorie/smartphones--c-phone | |
      | /de-DE/produkt/testmodell--p-phone | |
      | /de-DE/warenkorb | test-session-a |

  Szenario: Aktive Checkoutalternative verwendet das vollständige Linkziel aus dem Kaufresultat
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | Place:Unavailable |
    Und HTTPportdaten werden ersetzt
      | JSON |
      | {"Purchase":{"UnavailableSKU":{"Present":true,"Value":"BLUE-128"},"Alternatives":[{"ProductID":"p-phone","Name":"Verfügbare rote Variante","SKU":"RED-256","URL":"/de-DE/produkt/testmodell--p-phone?sku=RED-256"}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | Session | CSRFSource | Form |
      | POST | /de-DE/warenkorb/aktion | test-session-a | /de-DE/kasse | cartID=cart-a&commandID=alternative-link&revision=2&evaluationID=evaluation-a&action=place&return=checkout |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 409 |
      | Body | href="/de-DE/produkt/testmodell--p-phone?sku=RED-256" |
      | Absent | ?sku=RED-256?sku= |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/produkt/testmodell--p-phone?sku=RED-256 |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | data-sku="RED-256" |
