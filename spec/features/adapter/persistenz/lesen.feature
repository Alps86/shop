# language: de
@runtime @postgres
Funktionalität: SQL-Lesesichten und Gastidentität
  Öffentliche Portantworten werden aus tatsächlichen Datenbankzeilen aufgebaut.

  Hintergrund:
    Angenommen ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist
      | DatasetVersion | ProductID | SKU | Name | Price | StockQuantity | CategoryIDs |
      | demo-v1 | p-phone | PHONE | Synthetisches Testmodell | 10000 | 2 | mobile,home,coffee |
    Wenn der Datensatz veröffentlicht wird

  Szenariogrundriss: Gastfähigkeiten enthalten 256 Bit Entropie und werden ausschließlich gehasht gespeichert
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und die Gastfähigkeit um Tage gealtert wird
      | AgeDays |
      | <alter> |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | resolve | <session> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | OwnUser | RawStored | TokenBytes |
      | <code> | <eigen> | false | 32 |
    Beispiele:
      | alter | session | code | eigen |
      | 0 | own | OK | true |
      | 30 | own | NotFound | false |
      | 0 | unknown | NotFound | false |

  Szenariogrundriss: Gespeicherte Profile liefern fachliche Interessen
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | <profil> | own |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | User |
      | interests | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Count | CategoryID | Weight |
      | OK | <anzahl> | <kategorie> | <gewicht> |
    Beispiele:
      | profil | anzahl | kategorie | gewicht |
      | neutral | 0 | | |
      | mobile | 1 | mobile | 1 |
      | home | 1 | home | 1 |
      | coffee | 1 | coffee | 1 |

  Szenariogrundriss: Produktlesesichten wählen eine echte SKU und einen kanonischen Pfad
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU | Locale |
      | product | <id> | <sku> | <locale> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | ChosenSKU | Cards |
      | <code> | <gewaehlt> | <karten> |
    Beispiele:
      | id | sku | locale | code | gewaehlt | karten |
      | p-phone | | de-DE | OK | PHONE | 1 |
      | p-phone | PHONE | de-DE | OK | PHONE | 1 |
      | p-missing | | de-DE | NotFound | | 0 |
      | p-phone | OTHER | de-DE | InvalidReference | | 0 |
      | p-phone | | invalid | UnsupportedContext | | 0 |

  Szenariogrundriss: Kategoriepfade werden aus Stammdaten aufgebaut
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | Locale |
      | category | <id> | <locale> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Breadcrumb | Children |
      | <code> | <pfad> | 0 |
    Beispiele:
      | id | locale | code | pfad |
      | mobile | de-DE | OK | 1 |
      | missing | de-DE | NotFound | 0 |
      | mobile | invalid | UnsupportedContext | 0 |

  Szenariogrundriss: Ungeprüfte Varianten bleiben sichtbar und sperren keine freigegebene Nachbarvariante
    Angenommen ein Produkt mit gemischten Pflichtdatenvarianten vorbereitet wird
      | MainStatus | AdditionalSKU | AdditionalStatus | AdditionalPrice | AdditionalStock |
      | <hauptstatus> | CASE | <zusatzstatus> | <preis> | 5 |
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU | Locale |
      | product | p-phone | <sku> | de-DE |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | ChosenSKU | Cards | Available | VerificationStatus |
      | OK | <gewaehlt> | 2 | <menge> | <status> |
    Beispiele:
      | hauptstatus | zusatzstatus | preis | sku | gewaehlt | menge | status |
      | pending | verified | 9000 | PHONE | PHONE | 0 | pending |
      | pending | verified | 9000 | | CASE | 5 | verified |
      | verified | pending | 9000 | CASE | CASE | 0 | pending |
      | verified | pending | 9000 | | PHONE | 2 | verified |
      | verified | verified | 9000 | | CASE | 5 | verified |
      | verified | verified | 11000 | | PHONE | 2 | verified |
      | verified | verified | 10000 | | CASE | 5 | verified |

  Szenariogrundriss: SKU-Auflösung verwendet den maßgeblichen Katalog und den angefragten Kontext
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | SKU | Locale |
      | variant | <sku> | <locale> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | ChosenSKU | Cards |
      | <code> | <chosen> | <cards> |
    Beispiele:
      | sku | locale | code | chosen | cards |
      | PHONE | de-DE | OK | PHONE | 1 |
      | MISSING | de-DE | NotFound | | 0 |
      | PHONE | invalid | UnsupportedContext | | 0 |

  Szenario: Kuratierte Familienattribute bleiben mit false-Werten in Modell- und Kategoriesichten erhalten
    Angenommen kuratierte Produkt- und Variantenattribute vorbereitet werden
      | ProductFlag | VariantFlag |
      | false | true |
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | DatasetVersion | demo-v2 |
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU |
      | product | p-phone | PHONE |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | TopAttributes | ProductFlag | VariantFlag |
      | OK | 2 | false | true |
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID |
      | category | mobile |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | FilterDefinitions | TopDefinitions |
      | OK | 1 | 2 |

  Szenariogrundriss: Fehlende Katalogbestandteile erzeugen keine unvollständige Erfolgssicht
    Angenommen folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU |
      | <operation> | <id> | <sku> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | <code> |
    Beispiele:
      | sql | operation | id | sku | code |
      | SELECT 1 | product | p-phone | MISSING | InvalidReference |
      | SELECT 1 | category | missing | | NotFound |
      | DELETE FROM $schema.records WHERE kind='inventory' | product | p-phone | PHONE | NotFound |
      | DELETE FROM $schema.records WHERE kind='category' | product | p-phone | PHONE | NotFound |
