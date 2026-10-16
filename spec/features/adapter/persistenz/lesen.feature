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
      | UPDATE $schema.records SET payload=jsonb_set(payload,'{Variants}','[]') WHERE kind='product' | product | p-phone | PHONE | NotFound |
      | SELECT 1 | category | missing | | NotFound |
      | DELETE FROM $schema.records WHERE kind='inventory' | product | p-phone | PHONE | NotFound |
      | DELETE FROM $schema.records WHERE kind='category' | product | p-phone | PHONE | NotFound |

  Szenariogrundriss: Variantenbilder haben Vorrang und direkte Unterkategorien bleiben navigierbar
    Angenommen Modellbilder und eine Unterkategorie vorbereitet werden
      | ProductImage | VariantImage | AltText | ChildID |
      | https://example.invalid/model.svg | <variant> | Neutrale Darstellung | mobile-child |
    Und ein Importfeld vor der Veröffentlichung geändert wird
      | Field | Value |
      | DatasetVersion | demo-v2 |
    Wenn der Datensatz veröffentlicht wird
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID | SKU |
      | product | p-phone | PHONE |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | ImageURL |
      | OK | <image> |
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | ID |
      | category | mobile |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | Children |
      | OK | 1 |
    Beispiele:
      | variant | image |
      | | https://example.invalid/model.svg |
      | https://example.invalid/variant.svg | https://example.invalid/variant.svg |

  Szenario: Die administrative Orderliste akzeptiert keine Seite null
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | Page |
      | adminOrders | 0 |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code |
      | InvalidInput |

  Szenariogrundriss: Der Session-Interessenread prüft echte Fähigkeiten und Payloadfehler
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | mobile | own |
    Und die Gastfähigkeit um Tage gealtert wird
      | AgeDays |
      | <alter> |
    Und folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt
      | SQL |
      | <sql> |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | sessionInterests | <session> |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PayloadCode | Count | CategoryID | Weight |
      | <code> | <payload> | <anzahl> | <kategorie> | <gewicht> |
    Beispiele:
      | alter | session | sql | code | payload | anzahl | kategorie | gewicht |
      | 0 | own | SELECT 1 | OK | OK | 1 | mobile | 1 |
      | 30 | own | SELECT 1 | NotFound | OK | 0 | | |
      | 0 | unknown | SELECT 1 | NotFound | OK | 0 | | |
      | 0 | own | UPDATE $schema.users SET payload='"invalid"'::jsonb | OK | DependencyUnavailable | 0 | | |
      | 0 | own | DROP TABLE $schema.sessions | DependencyUnavailable | OK | 0 | | |

  Szenario: Interessen benötigen kein Leserecht auf Warenkörbe
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | mobile | own |
    Und die Interessenfähigkeit ohne Cartleserecht gelesen wird
      | Code | PayloadCode | Count | CategoryID | Weight | ResolveCode |
      | OK | OK | 1 | mobile | 1 | DependencyUnavailable |

  Szenario: Derselbe Gast liest nach einem Profilwechsel frische Interessen
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | mobile | own |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | sessionInterests | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PayloadCode | Count | CategoryID |
      | OK | OK | 1 | mobile |
    Wenn das persistierte Gastprofil gewählt wird
      | Profile | User |
      | home | own |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | sessionInterests | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PayloadCode | Count | CategoryID |
      | OK | OK | 1 | home |

  Szenario: Ein weiterer Gast übernimmt nicht die gespeicherten Interessen
    Wenn eine neue echte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | mobile | own |
    Und eine zweite getrennte Gastfähigkeit geöffnet wird
    Und das persistierte Gastprofil gewählt wird
      | Profile | User |
      | home | own |
    Und ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | sessionInterests | previous |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PayloadCode | Count | CategoryID |
      | OK | OK | 1 | mobile |
    Wenn ein persistenter Lesevertrag aufgerufen wird
      | Operation | Session |
      | sessionInterests | own |
    Dann lautet das öffentliche SQL-Leseergebnis
      | Code | PayloadCode | Count | CategoryID |
      | OK | OK | 1 | home |
