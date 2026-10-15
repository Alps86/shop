# Pflichtdaten und Anwendbarkeit

Der [Typenkatalog](type-model.md) übernimmt alle Hersteller-, Sicherheits-, Chemie-, Energie-, Dokument- und Mengenfelder des Projektmodells. Manufacturer ist der rechtliche Hersteller, Brand die Marke; ResponsiblePerson ist eine getrennte gegebenenfalls erforderliche EU-Rolle. PostalAddress und ElectronicAddress sind strukturierte echte Lieferantendaten, keine erfundene Kontaktadresse. Das Modell ist keine Rechtskonformitätsgarantie.

## Auflösung und Veröffentlichung

Auflösungskontext ist `(ProductID, SKU, Market CountryCode, Locale, At time.Time)`. `RuleReference` wirkt im halboffenen Zeitraum `[ValidFrom, ValidUntil)`, fehlende Grenze ist unbeschränkt. Überlappende widersprüchliche Regelstände sind Importfehler. Product.Compliance bildet die gemeinsame Basis. Variant.Compliance.Present=true ersetzt das vollständige Pflichtdatenpaket für diese SKU, kein uneindeutiger partieller Merge; der Import verlangt daher vollständige varianteigene Angaben. Herstellerabweichungen benötigen ergänzend `Variant.Manufacturer Optional[EconomicOperator]` und `Variant.ResponsiblePerson Optional[EconomicOperator]`; fehlend erbt die gemeinsame Rolle, keine Löschung einer anwendbaren Pflichtrolle.

Ergänzung `ComplianceContext` mit Market/Locale/ValidFrom/ValidUntil und `VerificationStatus` (`verified`, `pending`) je publizierbarem Pflichtdatenpaket. Nur verifizierte Daten, passende Sprache und gültiger Zeitraum erlauben Kauf. Fehlende Pflichtfelder ergeben ComplianceIncomplete: kein kaufbares Listingangebot, Detail zeigt „Pflichtangaben werden geprüft“, keine Kaufaktion; /backend nennt fehlende Felder. Cartabschluss prüft erneut. Unbekannt ist niemals gleich nicht anwendbar. Nicht anwendbar muss im Import mit Regelbezug und Begründung ausdrücklich freigegeben sein.

Pflichttexte fallen nicht automatisch auf andere Sprache zurück. Für DE muss freigegebener deutscher Text vorliegen; Energie-/Herstellerkennungen sind sprachneutral, Dokument.Locale muss passen. Warnungen, Piktogramme und Signalwort sichtbar an vorgeschriebenen Angebotsstellen, nicht ausschließlich hinter PDF. HTML wird escaped. Label/InformationSheet auf korrekte SKU/Modellnummer prüfen. Klasse samt Skala und zugänglichem Label-/Datenblattlink bei relevanten Karten/Teasern und Detail; HTMX ersetzt stets vollständige SKU-gebundene Pflichtdaten mit.

NetContent/UnitPriceBasis sind positive dimensionierbare Measurements; Grundpreis = aktueller Brutto-Inventorypreis × Bezugsmenge / Füllmenge, kaufmännisch auf Währungs-Minor-Unit gerundet, sichtbar direkt beim Gesamtpreis. Beispiel 500 ml zu 1000 Cent und Basis 1 l ergibt 2000 Cent/l. Keine separat pflegbare Grundpreiswahrheit.

## Geprüfte Anwendbarkeitsmatrix

Prüfstand **15.10.2026**, Absatzmarkt Deutschland/EU. Diese Matrix entscheidet modellierte Prüfgruppen; konkrete Lieferantendaten und Einstufung bleiben je Artikel zu verifizieren. Sie ersetzt keine vollständige juristische Einzelfallprüfung.

| Gruppe / Regel | Zeit/Geltung und Prüfung | Daten-/Darstellungsvertrag |
| --- | --- | --- |
| Alle relevanten Verbraucherprodukte: GPSR (EU)2023/988 Art.19 | Seit 13.12.2024, produktspezifische sektorale Grenzen beachten | Herstellerkontakt, ggf. EU-Verantwortlicher, Produktidentifikation, erforderliche Warn-/Sicherheitsinformationen am Onlineangebot |
| Smartphones/Waschmaschinen: Energie | Smartphone-Regeln (EU)2023/1669 ab 20.06.2025 für erfasste Modelle; Waschmaschinen (EU)2019/2014 seit 01.03.2021 | Klasse+Skala, elektronisches Label/Produktinformationsblatt, Modell-/EPREL-Bezug, keine frei erfundene Klasse |
| Entkalker/Reiniger: CLP (EG)1272/2008 | Nur gemäß tatsächlicher Stoff-/Gemischeinstufung. Aktuelle Werbungspflichten Art.48 prüfen; nicht jeden Reiniger pauschal als Gefahrgut einstufen | ProductIdentifier, Supplier, Piktogramme, Signalwort, H/P/EUH-Statements, UFI soweit erforderlich, Sicherheitsdatenblatt zusätzlich |
| CLP-Änderung (EU)2024/2865 und (EU)2025/2439 | Unterschiedliche Termine, keine pauschale „alles ab Juli 2026“-Regel. Verschobene Werbe-/Fernabsatzvorschriften Art.48/48a ab 01.01.2028; andere Änderungen 01.07.2026/01.01.2027 mit spezifischen Übergängen, Altware teils bis 2028/2029 | Regelcode und Stichtag explizit importieren. Neue Gefahrklassen (EU)2023/707 haben eigene Übergänge, bei Gemischen grundsätzlich 01.05.2026 bzw. für zuvor in Verkehr gebrachte 01.05.2028 prüfen |
| Mengenbezogene Ware: PAngV §§4/5 | Aktuelle Fassung, individuelle Ausnahmen prüfen | Füllmenge und Grundpreis im Angebot, Basis grundsätzlich kg/l; Entkalker 500 ml mit Preis/l |
| Elektrogeräte: ElektroG §§17/18; Batterien: (EU)2023/1542/BattDG | Elektro-/Batterierolle und konkrete Rücknahmepflichten prüfen; aktuelles BattDG statt überholtem BattG-Verweis | Verifizierte Entsorgungs-/Rücknahmehinweise, Lithium-/Entnahmehinweise soweit anwendbar als RequiredNotice; keine erfundene Rücknahmestelle |
| Vernetztes Produkt: Data Act (EU)2023/2854 Art.3(2) | Allgemeine Anwendung seit 12.09.2025; Art.3(1)-Designpflicht für nach 12.09.2026 in Verkehr gebrachte Produkte getrennt betrachten | Vorvertragliche Dateninformationen als sichtbarer Hinweis und dataActInformation-Dokument, bei tatsächlicher Vernetzung/Anwendbarkeit |
| Verbraucherinformation BGB §312j/EGBGB Art.246a | Bei echtem Verbrauchervertrag konkrete rechtliche Prüfung; Demo bleibt ausdrücklich Simulation | Eigenschaften, Gesamtpreis/Versand/Lieferbeschränkung/Zahlungsmittel vor Abschluss; keine echte Zahlung, deutlich beschrifteter Demoabschluss |
| Elektronischer Geschäftsverkehr: BFSG/BFSGV | Seit 28.06.2025, Betreiberrolle/Ausnahmen wie Kleinstunternehmen gesondert prüfen | Semantische Labels, Tastaturbedienbarkeit, Fokus/Fehlermeldungen/Alttexte; keine durch Test behauptete vollständige BFSG-Zertifizierung |

## Amtliche Quellen und Grenzen

- [GPSR](https://eur-lex.europa.eu/eli/reg/2023/988/oj), Art.19/52.
- [Kommission Smartphones](https://energy-efficient-products.ec.europa.eu/product-list/smartphones-and-tablets_en) und [Waschmaschinenverordnung](https://eur-lex.europa.eu/eli/reg_del/2019/2014/oj).
- [ECHA CLP-Grundlage](https://echa.europa.eu/regulations/clp/understanding-clp), [CLP-Änderung konsolidiert](https://eur-lex.europa.eu/eli/reg/2024/2865), [verschobene Termine 2025/2439](https://eur-lex.europa.eu/eli/reg/2025/2439/oj), [ECHA Termintabelle](https://www.echa.europa.eu/documents/10162/17217/application_dates_by_clp_en.pdf/9e72bee2-ef7b-9b5c-c628-b36722082cfe).
- [PAngV §4](https://www.gesetze-im-internet.de/pangv_2022/__4.html), [§5](https://www.gesetze-im-internet.de/pangv_2022/__5.html), [ElektroG §18](https://www.gesetze-im-internet.de/elektrog_2015/__18.html), [BattDG](https://www.gesetze-im-internet.de/battdg/), [Data Act](https://eur-lex.europa.eu/eli/reg/2023/2854/oj).
- [BGB §312j](https://www.gesetze-im-internet.de/bgb/__312j.html), [EGBGB Art.246a](https://www.gesetze-im-internet.de/bgbeg/art_246a__1.html), [Bundesfachstelle BFSG](https://www.bundesfachstelle-barrierefreiheit.de/DE/Barrierefreiheitsstaerkungsgesetz).

EP-Entkalkertext belegte Füllmenge/Grundpreis und Identifikatoren, nicht vollständige Gefahren-/Herstellerdaten. Daraus folgt weder Ungefährlichkeit noch fehlende Kennzeichnung im Originalshop. Demobefüllung muss diese Daten separat verifizieren; unvollständige Produkte bleiben als gesperrter Datensatz sichtbar, ohne Sicherheitsinformationen zu erfinden.
