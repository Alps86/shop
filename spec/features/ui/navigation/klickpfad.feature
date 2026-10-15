# language: de
@spec-only @integration
Funktionalität: Durchgängiger eigener Gastpfad
  Als Gast erreiche ich Abschluss und Status über öffentliche Links und Formulare.

  Szenariogrundriss: Öffentliche Navigation nutzt die vereinbarten Ports
    Angenommen für jeden Schritt liefert der zuständige Port eine erfolgreiche Ansichtsantwort
      | Source | Action | Target |
      | <quelle> | <aktion> | <ziel> |
    Wenn die dargestellte Aktion über HTTP ausgeführt wird
      | Mode |
      | HTML-form-or-link |
    Dann ist der nächste Schritt erreichbar
      | Target | Preserves | MethodSafety |
      | <ziel> | <erhalt> | GET-read-POST-write |
    Beispiele:
      | quelle | aktion | ziel | erhalt |
      | home | category | category | locale |
      | category | selectBrand | listing | query |
      | listing | addBrand | listing | multiSelection |
      | listing | modelCard | product | chosenSKU |
      | product | variant | product | activeSKU |
      | product | addToCart | cart | guestSession |
      | cart | acceptAccessory | cart | guestSession |
      | cart | declineAccessory | cart | existingItems |
      | cart | removeItem | cart | guestSession |
      | cart | checkout | checkout | cartIdentity |
      | checkout | confirmTotal | checkout | evaluationIdentity |
      | checkout | placeDemoOrder | orderStatus | sameOrderOnRetry |
      | backendOrder | reportPartialShipment | backendOrder | purchaseSnapshot |
      | orderStatus | refresh | orderStatus | guestOwnership |

  Szenario: Filteränderung und Browserzurück erhalten teilbaren Zustand
    Angenommen eine Listingansicht ist geöffnet
      | URL |
      | /de-DE/kategorie/smartphones--c-phone?brand=Alpha&page=2 |
    Wenn die Filterauswahl geändert und anschließend zurück navigiert wird
      | AddedBrand | Navigation |
      | Beta | history-back |
    Dann lauten die beobachteten Zustände
      | Moment | Brands | Page |
      | afterChange | Alpha,Beta | 1 |
      | afterBack | Alpha | 2 |
