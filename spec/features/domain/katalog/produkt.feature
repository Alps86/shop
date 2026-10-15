# language: de
@spec-only @core
Funktionalität: Produktmodell und Varianten
  Als Betreiber publiziere ich gültige Modelle mit eindeutig verkaufbaren Varianten.

  Szenariogrundriss: Veröffentlichung eines Produktmodells
    Angenommen die referenzierten Katalogdaten sind veröffentlicht
      | FamilyID | FamilyVersion | CategoryID | Brand | Locale | Name |
      | f-phone | 1 | c-phone | Demo | de-DE | Testmodell |
    Wenn ein Produkt veröffentlicht wird
      | ID | FamilyID | FamilyVersion | PrimaryCategoryID | Variants | ExistingSKUOwner |
      | <kennung> | <familie> | 1 | <kategorie> | <varianten> | <besitzer> |
    Dann lautet das Veröffentlichungsergebnis
      | Code | ProductCount | VariantCount |
      | <code> | <produkte> | <anzahl> |
    Beispiele:
      | kennung | familie | kategorie | varianten | besitzer | code | produkte | anzahl |
      | p-phone | f-phone | c-phone | PHONE-BLUE-128,PHONE-BLUE-256 | none | OK | 1 | 2 |
      | p-cable | f-phone | c-phone | CABLE | none | OK | 1 | 1 |
      | p-phone | f-phone | c-phone | [] | none | InvalidInput | 0 | 0 |
      | p-phone | f-phone | c-phone | SAME,SAME | none | DuplicateSKU | 0 | 0 |
      | p-phone | f-phone | c-phone | TAKEN | p-other | DuplicateSKU | 0 | 0 |
      | p-phone | f-missing | c-phone | PHONE | none | InvalidReference | 0 | 0 |
      | p-phone | f-phone | c-missing | PHONE | none | InvalidReference | 0 | 0 |

  Szenariogrundriss: Referenzen und Baumgrenzen
    Wenn eine Katalogänderung validiert wird
      | Operation | Input |
      | <operation> | <eingabe> |
    Dann lautet das Validierungsergebnis
      | Code | Published |
      | <code> | <publiziert> |
    Beispiele:
      | operation | eingabe | code | publiziert |
      | categoryParent | c-a:c-b,c-b:none | OK | true |
      | categoryParent | c-a:c-b,c-b:c-a | InvalidReference | false |
      | categoryParent | c-a:c-a | InvalidReference | false |
      | deleteCategory | referenced:c-phone | Conflict | false |
      | accessory | source:p-a,target:p-a | InvalidReference | false |
      | accessory | source:p-a,target:missing | InvalidReference | false |
      | additionalCategory | primary:c-a,additional:c-a | InvalidReference | false |
      | familyVersion | bound:1,published:2,revalidated:false | Conflict | false |
      | categoryAttributes | filters:5,top:3 | InvalidAttribute | false |
      | categoryAttributes | filters:4,top:4 | InvalidAttribute | false |

  Szenariogrundriss: Produktpflichtfelder und gültige Identifikatoren
    Wenn ein ansonsten gültiges Produkt mit einem geänderten Feld publiziert wird
      | Field | Value |
      | <feld> | <wert> |
    Dann lautet das Feldvalidierungsergebnis
      | Code | Published |
      | <code> | <publiziert> |
    Beispiele:
      | feld | wert | code | publiziert |
      | Name.de-DE | empty | InvalidInput | false |
      | Brand | empty | InvalidInput | false |
      | ID | empty | InvalidInput | false |
      | FamilyID | empty | InvalidInput | false |
      | PrimaryCategoryID | empty | InvalidInput | false |
      | Variant.GTIN | 4006381333931 | OK | true |
      | Variant.GTIN | 4006381333932 | InvalidInput | false |
