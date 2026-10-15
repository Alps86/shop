# language: de
@spec-only @frontend
Funktionalität: Zusammenhängende Variantenansicht
  Als Käufer wechsle ich die vollständige Variante mit oder ohne HTMX.

  Szenariogrundriss: Variantenwechsel ersetzt eine vollständige sichtbare Einheit
    Angenommen ein explizites ProductDetail enthält Variantenansichten
      | SKU | Color | StorageGB | Gallery | Price | ComplianceMarker | AccessoryMarker |
      | BLUE-128 | blue | 128 | gallery-a | 10000 | verified-a | accessory-a |
      | RED-256 | red | 256 | gallery-b | 12000 | verified-b | accessory-b |
    Wenn die aktive Variante angefordert wird
      | SKU | HX-Request |
      | RED-256 | <htmx> |
    Dann lautet die sichtbare Variantenansicht
      | SKU | Gallery | Price | ComplianceMarker | AccessoryMarker | Vary | HistoryURL | ResponseShape |
      | RED-256 | gallery-b | 12000 | verified-b | accessory-b | HX-Request | ?sku=RED-256 | <form> |
    Beispiele:
      | htmx | form |
      | false | fullPage |
      | true | completeVariantFragment |

  Szenario: Öffentliche Produktkarte nutzt DTO statt interner Rankingdaten
    Angenommen eine ProductCard liegt vor
      | ProductID | ChosenSKU | VariantCount | TopAttributes | ComplianceSummary | UnitPrice |
      | p-phone | RED-256 | 3 | storageCapacity,display,battery | verified-marker | 12000 |
    Wenn die Modellkarte gerendert wird
    Dann lautet der Kartenvertrag
      | CardCount | LinkSKU | VariantCountVisible | TopAttributeCount | ComplianceVisible | MarginVisible | InterestVisible |
      | 1 | RED-256 | true | 3 | true | false | false |
