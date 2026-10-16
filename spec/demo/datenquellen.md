# Quellen und Grenzen des Vorführkatalogs

Stand und Abrufdatum: **16.10.2026**. Datensatz: [`demo-real-v1`](../fixtures/demo/katalog.json). Der begrenzte Vorführkatalog enthält **4 Produktmodelle, 8 Varianten und 5 Kategorien**. Er ist kein Vollimport und kein realer Verkaufsbestand. Die vier Gastprofile stammen aus dem bestehenden Profilvertrag; diese Fixture erzeugt keine zusätzlichen Gäste.

„Verified“ im Pflichtdatenpaket bedeutet hier, dass die aufgenommenen Angaben zu den tatsächlich gelesenen Quellen passen, nicht eine pauschale Rechtskonformitätsgarantie oder Herstellerzustimmung zum Demoangebot. Maßgeblich bleiben der [Pflichtdatenvertrag](../architecture/05-bausteine/compliance.md) und die fachliche Validierung beim Import und Kauf. Ein Quellenlink allein ist kein Ausführungsnachweis. Die Datenquelle für DB, öffentliche Routen und Hugo-Rahmen ist dieselbe JSON-Datei.

## Reale Fakten und ausdrücklich erfundene Demowerte

**Alle Preise, Bestände, internen Produkt-/Familien-/Kategorie-IDs, Shop-SKUs, Einkaufskosten, Rabattregeln und Co-Purchase-Zahlen sind eigene fiktive Testwerte.** Sie stammen weder von den Herstellern noch aus beobachteten Kundenkäufen. Preise sind Bruttowerte in EUR; sie sind keine UVP, aktuellen Marktpreise oder angekündigten Herstellerangebote. Ein vorhandener Bestand bedeutet keine reale Lieferbarkeit. `PurchaseCosts` sind ebenfalls synthetisch, mit ganzzahlig abgerundeten 75 % des jeweiligen Demopreises.

|Produkt-ID / Kategorie|Interne Shop-SKU|Belegte Variante|Fiktiver Preis|Fiktiver Anfangsbestand|
|---|---|---|---|---|
|`p-fairphone5` / `smartphones`|`FP5-BLACK-128`|Matte Black, 128 GB, 6 GB RAM|399,00 EUR|20|
|`p-fairphone5` / `smartphones`|`FP5-GREEN-128`|Moss Green, 128 GB, 6 GB RAM|409,00 EUR|20|
|`p-fairphone5` / `smartphones`|`FP5-BLACK-256`|Matte Black, 256 GB, 8 GB RAM|449,00 EUR|20|
|`p-fairphone5` / `smartphones`|`FP5-BLUE-256`|Sky Blue, 256 GB, 8 GB RAM|459,00 EUR|20|
|`p-fairphone5` / `smartphones`|`FP5-TRANSPARENT-256`|Transparent edition, 256 GB, 8 GB RAM|469,00 EUR|20|
|`p-fairphone5-case` / `accessories`|`FP5-CASE-BLACK`|Fairphone-5-Schutzhülle, Matte Black|19,90 EUR|40|
|`p-constructa-cwf14n30` / `home`|`CWF14N30`|Waschmaschine CWF14N30, eine Variante|649,00 EUR|8|
|`p-drbecher-entkalker` / `coffee`|`DB-ENTKALKER-500`|DR. BECHER@home Schnell Entkalker, 500 ml|7,99 EUR|30|

Die Entkalker-SKU ist intern; nur deren GTIN **4000602009667** und Herstellerartikelnummer **1966000** sind primär belegt. Keine unbekannten Telefon-/Hüllen-/Waschmaschinen-GTIN oder variantenbezogenen Hersteller-MPN erfunden. Die telefonbezogenen Farb-/Speicherkombinationen sind exakt die belegten fünf Kombinationen, kein kartesisches Produkt aller Farben und Speichergrößen.

Die Kategorien `mobile`, `home` und `coffee` entsprechen den bestehenden Interessenprofilen. `smartphones` und `accessories` sind Kinder von `mobile`; Telefon und Hülle führen `mobile` zusätzlich ausdrücklich als Kategorie. Interne Slugs: `mobilgeraete`, `haushalt`, `kaffee-pflege`, `smartphones`, `zubehoer`. Produkt-Slugs: `fairphone-5`, `fairphone-5-schutzhuelle`, `constructa-cwf14n30`, `dr-becher-schnell-entkalker`.

Die kuratierten Familienattribute sind Telefonfarbe `color` mit `matte-black`, `moss-green`, `sky-blue`, `transparent`, Speichermesswert `storageCapacity` in GB und RAM `ram` in GB. Die Hülle besitzt getrennt `caseColor`; die Waschmaschine `loadCapacity` in kg und `spinSpeed` in rpm. Unterschiedliche Attribute werden nicht als dieselbe fachliche Größe ausgegeben. Globales Marke-OR kann die tatsächlichen Marken **Fairphone**, **Constructa** und **Dr. Becher** vergleichen. Das sind Markenbezeichnungen, keine zusätzlichen Händler oder Betreiber.

## FP5: Identität, Varianten, Hersteller und Energie

Quellen:

- [Offizielle Fairphone-Promotionsbedingungen](https://www.fairphone.com/legal/fairphone-promotions-and-offers): nennt genau die fünf oben übernommenen FP5-Farb-/Speicher-/RAM-Kombinationen.
- [Hersteller-FAQ zum FP5](https://support.fairphone.com/hc/en-us/articles/18020671537041-Fairphone-5-FAQ) und [Einführungsmitteilung der 6-GB-Ausführung](https://www.fairphone.com/wp-content/uploads/2024/09/Introducing-the-most-affordable-Fairphone-5-yet-the-6GB-version.pdf): Zuordnung 128 GB/6 GB RAM gegenüber 256 GB/8 GB RAM.
- [Modellbezogene EU-Konformitätserklärung](https://www.fairphone.com/wp-content/uploads/2023/09/FP5_CE-DoC-Declaration-of-Operation-EUUK.pdf): Hersteller FairPhone B.V., Modell FP5.
- [Amtliches deutsches Produktinformationsblatt EPREL 2378858](https://eprel.ec.europa.eu/fiches/smartphonestablets20231669/Fiche_2378858_DE.pdf), [amtliches Energielabel](https://eprel.ec.europa.eu/labels/smartphonestablets20231669/Label_2378858.pdf) und [Herstellerimpressum](https://www.fairphone.com/legal/impressum).

Hersteller: **FairPhone B.V., Van Diemenstraat 200, Floor 3, 1013 CP Amsterdam, NL**, `support@fairphone.com`, +31 85 225 0203. Der EU-Hersteller ist keine erfundene EU-Verantwortlichenvertretung; keine zusätzliche ResponsiblePerson notwendig gemacht. Energieklasse **B**, Skala **A bis G**, Modell **FP5**, EPREL **2378858**. Label und Fiche wurden geladen und tatsächlich gelesen, das Label zusätzlich visuell geprüft. Die gemeinsamen modellbezogenen Energiedokumente werden im Seed für jede der fünf internen SKUs passend gebunden. Keine fremde SKU-Query und kein FP6-Label.

Das aktuelle Label nennt 36 h 0 min, 1300 Zyklen und IP55. Ältere Handbuch-/Suchangaben wie 800 Zyklen oder 37 h 03 min ersetzen diese amtlichen Energieangaben nicht. Die Fixture führt solche widersprüchlichen Nebenwerte nicht als zusätzliches Variantenattribut.

**Lieferbarkeitsgrenze:** Das aktuelle EPREL-Fiche nennt Ende des Inverkehrbringens am **09.02.2026**. Die [heutige Hersteller-Shopseite](https://www.fairphone.com/shop/fairphone-5-1959) ist kein Beleg verfügbarer Neuware. Der Demobestand ist simuliert, keine Behauptung neuer laufender Produktion oder tatsächlicher Herstellerlieferung.

### FP5: deutsche Sicherheits-, Batterie- und Dateninformation

[Mehrsprachiges Hersteller-Sicherheitshandbuch](https://support.fairphone.com/hc/en-us/article_attachments/14779133028626), Modell FP5, **deutscher Fachtext auf Druckseiten 5–7**: passende Ladegeräte, Feuchtigkeit, Akkumissbrauch/Brand-/Explosions-/Auslaufgefahr und Kleinteile. Der URL-Pfad `en-us` bezeichnet den Downloadhost, nicht eine ausschließlich englische PDF: der tatsächlich gelesene deutsche Abschnitt ist der Sprachbeleg. Keine englische Einzeldatei als Deutsch etikettiert.

Akkuentnahme nach [offizieller FP5-Anleitung](https://support.fairphone.com/hc/en-us/articles/18744611991825-Fairphone-5-Replace-a-Spare-Part), Stand 29.01.2026: ausschalten, Rückdeckel von der unteren rechten Kerbe lösen, Akku an der unteren Kerbe herausheben. Der deutsche Notice ist eine eigene sachliche Paraphrase; die englische Schrittanleitung wird nicht als deutsches Pflichtdokument ausgegeben. Modellakku **F5AC1**, wiederaufladbarer Lithium-Ionen-Akku. Getrennte Entsorgung von Telefon/Akku nach deutschem Sicherheitshandbuch; [Hersteller-Recyclingprogramm auf Deutsch](https://www.fairphone.com/de/recycle) und [Ersatzteil-Recyclinginformation](https://support.fairphone.com/hc/en-us/articles/11538143761041-Recycle-Old-Spare-Parts-and-Electronics), Stand 18.08.2026: Altakku lokal sammeln, nicht mit normaler Post versenden. Keine erfundene Rücknahmestelle oder reale Rücknahmeleistung des Demobetreibers.

FP5 ist laut Hersteller ein tatsächlich vernetztes Produkt. Die vollständige **eigene deutsche** [Dateninformation](https://raw.githubusercontent.com/Alps86/shop/53f1a713f898a0d347e48a4d1b6f4524399aed4a/spec/demo/fairphone-dateninformation.de.md) ist unter einem unveränderlichen vollständigen Commit-SHA veröffentlicht und im Seed als `dataActInformation` mit Locale `de-DE` gebunden. Öffentliches HTTPS-GET vor Seedaufnahme erfolgreich; byteidentischer SHA-256 **`dc2c0092678d233e6ebb6be3996fade8cf9c881daeebbac9a671ddc2248ab108`**. Das ist unsere Information, keine offizielle Herstellerübersetzung. Sie deckt Smartphonekategorien, relevante Dienste, Formate, Mengen ohne erfundene Bezugsgröße, Entstehung, Speicherung, Zugriff und Löschung ab. Die dortige [Herstellerquelle April 2026](https://www.fairphone.com/resources/pdf/202604-EN-Data-Disclosure-Statement-1.pdf) trägt trotz Veröffentlichung den Metadatentitel „Draft - Data Disclosure Statement“; diese Unsicherheit bleibt sichtbar. Der zusätzliche sichtbare Notice ersetzt das deutsche Dokument nicht. Die direkt auf Fairphone verfügbaren `/de/legal/`-Data-Act-Seiten enthalten englischen Fachtext beziehungsweise die EN-PDF und sind ausdrücklich kein deutscher Sprachersatz.

## Einfacher Artikel und Zubehör: FP5-Schutzhülle

[Offizielle Fairphone-Produktseite der FP5-Schutzhülle](https://www.fairphone.com/shop/fairphone-5-protective-case-256): ausdrücklich **nur für Fairphone 5**, angebotene Farben unter anderem Matte Black. Im Seed genau eine schwarze Variante. Marken-/Kontaktzuordnung zum Fairphone-eigenen Artikel anhand dieser Herstellerproduktseite und [Fairphone-Impressum](https://www.fairphone.com/legal/impressum); keine davon abweichende Fertigerrolle oder Drittvertretung erfunden. Eine besondere Risikowarnung wird ohne Quelle nicht ergänzt; das ist keine pauschale Erklärung der Gefahrlosigkeit.

Die Zubehörreferenz des FP5 nennt ausschließlich die fünf FP5-SKUs. Sie wird weder als kompatibel mit Waschmaschine, Entkalker oder anderen Fairphone-Generationen geführt noch mit frei erfundenen Schutzleistungen beworben.

**Fiktive Shopregel** `demo-phone-case-v1`, Version 1, Scope **`pair`**: 5,00 EUR je passendem Paar aus `p-fairphone5` und `p-fairphone5-case`; Betrag und Wiederholungslogik folgen der vorhandenen Domain. Beispiel: 399,00 + 19,90 − 5,00 = **413,90 EUR vor Versand**. Ohne Annahme oder nach Entfernung der Hülle kein passendes Paar. Kein Herstellerpreisnachlass.

Die beiden Co-Purchase-Einträge FP5→Hülle und Hülle→FP5 haben je **12** fiktive Warenkörbe, Zeitfenster `[01.09.2026, 01.10.2026)`, DatasetVersion **`synthetic-demo-v1`**. Sie sind technische Vorführstatistik, keine Kundenhistorie. Keine chemische Geräteverträglichkeit aus einer Empfehlung herleiten.

## CWF14N30: Energie, Hersteller und deutsche Dokumente

- [Amtliches deutsches Fiche EPREL 2219955](https://eprel.ec.europa.eu/fiches/washingmachines2019/Fiche_2219955_DE.pdf) und [amtliches Label](https://eprel.ec.europa.eu/labels/washingmachines2019/Label_2219955.pdf), beide tatsächlich gelesen, Label visuell geprüft: Modell **CWF14N30**, Energieklasse **A**, Skala **A bis G**, 8 kg, maximal 1400 U/min. Das Fiche nennt außerdem 46 kWh/100 Eco-40–60-Zyklen, 47 l/Zyklus, 3:33 h, 71 dB; diese Werte werden nicht mit unbekannten anderen Programmen vermischt.
- [Aktuelles Herstellerfiche](https://media3.bsh-group.com/Documents/eudatasheet/de-DE/CWF14N30.pdf), Fußdatum 27.08.2026. Abweichende Tiefenmaße (amtlich 63 cm, Hersteller 59 cm) werden **nicht** zu scheinbar präzisen Maßen kombiniert; Maße fehlen absichtlich im Demoartikel.
- Hersteller aus amtlichem Fiche: **BSH Hausgeräte GmbH, Carl-Wery-Straße 34, 81739 München, DE**, modellbezogener Kontakt **NEFF Constructa Service**, `infoneff@bshg.com`, +49 89 4590 04. Nicht den Vertriebsgesellschaftsnamen als Hersteller ersetzen.
- [Deutsche Gebrauchs- und Montageanleitung 9002037180_A](https://media3.bsh-group.com/Documents/9002037180_A.pdf), 48 Seiten, Modell auf Deckblatt. Sicherheitskapitel und Entsorgung Druckseite 42 gelesen; letztere zusätzlich visuell geprüft. Keine Nutzung der Warnungen eines ähnlichen Modells.
- [Exakte Herstellerproduktseite](https://www.constructa.com/de/de/mkt-product/waschen-und-trocknen/waschmaschinen/CWF14N30): **Home Connect / WiFi: Nein**. Allgemeine Fußnoten zu anderen vernetzten Geräten ändern diesen Modellwert nicht.
- **Dennoch Data Act:** Die Produktseite verlinkt im CWF14N30-Dokumentarray ausdrücklich die [deutsche modellbezogene Rechtssammlung](https://media3.bsh-group.com/Documents/legalcollection_CWF14N30_de-DE.pdf). Zwei Seiten vollständig gelesen, Version **1.0ncp vom 25.08.2025**: Nutzungs-/Zustandsdaten, beispielsweise TXT, geschätzt unter etwa 10 KB ausleitbare Daten pro Tag; lokale Speicherung und regelmäßiges Überschreiben; kein Server und keine kontinuierliche relevante Echtzeitdatengenerierung für den Nutzer. Kundendiensttechniker kann verbinden, auslesen und Daten regelmäßig per E-Mail bereitstellen. Auslese-/Bereitstellungsvorgang kostenlos; weitere Einsatzkosten möglich. Geräteseitige Datenlöschung laut Quelle nicht möglich. Kein N/A aus fehlendem WLAN und keine erfundene Home-Connect-Cloudfunktion.

### CWF14N30: bedingte Entsorgung und Batteriehinweise

Die von Constructa verlinkte [BSH-Entsorgungsinformation vom 08.06.2026](https://media3.bsh-group.com/Documents/29121638_2026_06_08_Website_Entsorgung_und_Recycling.pdf) wurde vollständig gelesen. Getrennte Elektroaltgerätesammlung, kommunale/gesetzliche Handelsrücknahme und **bedingte** Vorabtrennung nicht vom Gerät umschlossener Batterien/Akkus sowie zerstörungsfrei entnehmbarer Lampen. Bei lithiumhaltigen Altbatterien Kontakte isolieren. Zusätzlich [allgemeine Constructa-Batteriehinweise](https://www.constructa.com/de/shop/hinweise_zur_batterie_entsorgung). Diese Angaben werden als bedingte Notices übernommen: **kein bestimmter Akku, keine Chemie eines unbekannten internen Akkus und keine Batterieabwesenheit werden behauptet**. Kein unbelegtes Batterie-N/A. Eine vollständige Fachdemontageanweisung wird daraus nicht abgeleitet.

Die allgemeine BSH-Information nennt Eigenverantwortung für personenbezogene Datenlöschung. Die spezielle CWF14N30-Data-Act-Datei nennt fehlende geräteseitige Löschmöglichkeit. Diese unterschiedliche Reichweite wird erhalten; keine erfundene Rücksetzfunktion zur Löschung sämtlicher Daten versprochen. Der Simulator bietet keinen eigenen realen Rücknahmeservice.

## Schnell Entkalker 500 ml: Identität, CLP und Grundpreis

[Herstellerproduktseite](https://www.drbecherhome.de/produkte/schnell-entkalker), dort verlinktes [deutsches Sicherheitsdatenblatt](https://www.drbecherhome.de/sdb/bah/DE/SDB_4000602009667.pdf), [Inhaltsstoffblatt](https://www.drbecherhome.de/isa/bah/DE/ISA_4000602009667.pdf) und [Herstellerbroschüre](https://www.drbecherhome.de/LinkClick.aspx?fileticket=isnhSd6d2e4%3D&portalid=0). Die Broschüre belegt 500-ml-Dosierflasche, Einzel-GTIN **4000602009667**, Artikel **1966000**. **4000602099668** ist die Verpackungseinheit und wird nicht als Einzel-GTIN verwendet.

Hersteller/Lieferant nach SDB Abschnitt 1.3: **Dr. Becher GmbH, Vor den Specken 3, 30926 Seelze, DE**, `info@drbecher.de`, +49 5137 9901 0. Der im SDB separat genannte Ersteller DLAC ist nicht der Produktlieferant. UFI **FEXK-HXDG-S28D-50NV**.

Das neunseitige SDB wurde vollständig gelesen, Seite 1 visuell geprüft. Ausgabe 20.05.2025, Revision **15.12.2025**, Version **2.0** (ersetzt 1.0). **Dokumentunsicherheit:** Folgefußzeilen nennen 13.12.2024; die PDF-Metadaten nennen Erstellung/Änderung am 13.11.2025, also vor dem Revisionsdatum auf Seite 1. Es ist die aktuell auf der Produktseite verlinkte, eindeutig passende Fassung; daraus keine Garantie einer widerspruchsfreien oder global neuesten Rechtsfreigabe machen. Älteres Inhaltsstoffblatt 04.01.2024 ersetzt nicht die spätere SDB-Kennzeichnung.

Gemischeinstufung Abschnitt 2.1: Hautätzend Kategorie 1/H314 und schwere Augenschädigung Kategorie 1/H318. **Kennzeichnung Abschnitt 2.2** hingegen genau **GHS05**, **Gefahr**, **H314** und die folgenden P-Sätze. Deshalb H318 nicht zusätzlich als Kennzeichnungssatz anhängen; auch H315/H319/H412 einzelner Inhaltsstoffe aus Abschnitt 3 nicht zum Gemischlabel addieren. Keine EUH-Sätze in Abschnitt 2.2 angegeben; keine erfunden.

|Code|Deutscher Kennzeichnungstext aus Abschnitt 2.2|
|---|---|
|H314|Verursacht schwere Verätzungen der Haut und schwere Augenschäden|
|P102|Darf nicht in die Hände von Kindern gelangen|
|P280|Schutzhandschuhe, Schutzkleidung, Augenschutz tragen|
|P303+P361+P353|BEI BERÜHRUNG MIT DER HAUT (oder dem Haar): Alle kontaminierten Kleidungsstücke sofort ausziehen. Haut mit Wasser abwaschen|
|P305+P351+P338|BEI KONTAKT MIT DEN AUGEN: Einige Minuten lang behutsam mit Wasser spülen. Eventuell vorhandene Kontaktlinsen nach Möglichkeit entfernen. Weiter spülen|
|P310|Sofort GIFTINFORMATIONSZENTRUM, Arzt anrufen|
|P501|Inhalt/Behälter gemäß nationalen Vorschriften der Entsorgung zuführen.|

GHS05-Bild: [tatsächliche HTTPS-Datei bei PubChem/NIH](https://pubchem.ncbi.nlm.nih.gov/images/ghs/GHS05.gif), GET **200**, `image/gif`, 720×720, visuell korrektes Ätzwirkungspiktogramm. [Primäre Symbolzuordnung](https://pubchem.ncbi.nlm.nih.gov/ghs/ghs_8.html). Kein erfundenes localhost-HTTPS und keine Lockerung der Domainprüfung. Das regulatorische Symbol wird nur der tatsächlichen SDB-Kennzeichnung zugeordnet, nicht als Produktfoto verwendet.

Füllmenge **500 ml**, Grundpreisbasis **1 l**, beide positive dimensionierbare Werte. Der Grundpreis ist **kein gespeicherter eigener Preiswert**: Bei fiktiv 7,99 EUR/500 ml berechnet die Domain **15,98 EUR/l** aus aktuellem Bruttoinventorypreis. Bei Preisänderung muss er neu entstehen. Keine pauschale Verträglichkeit mit CWF14N30 oder anderen Geräten behaupten. Transportangaben des SDB begründen hier keine ausgeführte Gefahrgutlieferung; sämtliche Bestellungen/Zahlungen bleiben Simulation.

## Regelstand und Verifikation

Die Pflichtdatenpakete gelten für Markt DE und Locale `de-DE`, Verifikations-/Abrufstand 16.10.2026. `ValidFrom` im Kontext ist 16.10.2026, Regelreferenz-Stand 15.10.2026; das sind **Prüfstichtage dieses Seeds**, keine Behauptung, die Gesetze seien erst an diesem Datum in Kraft getreten. Keine unbekannten Herstellerablaufdaten erfunden; bei neuen Quellen oder geändertem Stand ist erneut zu prüfen.

GPSR-/Energie-/Elektro-/Batterie-/Data-Act-/CLP-/PAngV-Gruppen werden artikelbezogen geführt. Die chemischen Übergänge (EU)2024/2865 und (EU)2025/2439 sind im bestehenden Vertrag zu beachten; insbesondere verschobene Fernabsatz-/Werbetermine ab 01.01.2028 nicht als bereits vollständig seit Juli 2026 geltend ausgeben. Das Produktlabel wird aus der tatsächlich vorliegenden Einstufung übernommen, keine neuen Gefahrklassen hinzuerfinden. Amtliche Regelquellen stehen im [Compliancevertrag](../architecture/05-bausteine/compliance.md#amtliche-quellen-und-grenzen).

Die Hülle wird als einfacher nicht vernetzter Schutzartikel ohne eigene Energie-/Chemiekennzeichnung geführt; keine zusätzliche elektrische oder chemische Funktion behauptet. Sicherheitsangaben beziehen sich nur auf belegte Hinweise. Für die Elektrogeräte bleiben anwendbare Entsorgungs-/Batteriehinweise vorhanden; fehlende Belege werden nicht über einen erfundenen pauschalen `NonApplicableReason` weggeschrieben.

## Bilder und Nutzungsgrundlage

Alle Produktkarten verwenden die bereits eigene neutrale Darstellung **`/assets/product.dad5b938174f.svg`**. Sie ist keine Abbildung des tatsächlichen Herstellermodells; Alttexte sagen das ausdrücklich. Es werden keine Herstellerfotos oder Logos kopiert und keine fremden Bildrechte behauptet. Externe Hersteller-/EPREL-/SDB-Dokumente werden verlinkt, nicht als eigene Texte/Assets ausgegeben. Die eigene deutsche Dateninformation benennt Autorschaft, Quelle und Grenzen eigenständig.

Die Fixture ist kein synthetischer 30/90-Lastkatalog. Ein gesondertes technisches Referenzprofil darf diesen kleinen quellengeprüften Vorführkatalog weder ersetzen noch mit ungeprüften real behaupteten Artikeln vermischen.

## Getrenntes synthetisches Benchmarkprofil

[`benchmark.json`](../fixtures/demo/benchmark.json), DatasetVersion **`benchmark-synthetic-v1`**, enthält genau **30 Products, 90 Varianten und 5 Kategorien**, jeweils drei Varianten je Produkt und Anfangsbestand **2000 je SKU**. Produkt-IDs `bench-product-01` bis `bench-product-30`, SKUs `BENCH-01-A` bis `BENCH-30-C`. Auswahlmerkmal `sample` mit `a`, `b`, `c`. Kategorien `mobile`, `home`, `coffee`, `smartphones`, `accessories`; je sechs Produkte direkt zugeordnet, die letzten zwei Kategorien sind Kinder von `mobile`. Vier vorhandene Profile werden über die Anwendung eingerichtet, nicht durch zusätzliche Gastdaten im Seed.

Dies sind **ausschließlich eigene technische Messdaten**, keine 30 realen oder verifizierten Herstellermodelle. Namen, Herstellerkontakt einschließlich fiktiver Anschrift und `example.invalid`-E-Mail, Marken „Synthetisch A/B/C“, Preise, Bestände und Einkaufskosten sind vollständig erfunden. „Verified“ ist hier nur der technische Status eines synthetischen Pflichtdatenpakets für erfolgreiche Messpfade; daraus folgt keine Prüfung realer Ware oder Befreiung tatsächlicher Produkte von Vorschriften. Die neutralen Messartikel behaupten keine elektrische, chemische oder sonstige regulierte Produkteigenschaft. Eigener sichtbarer Notice und Beschreibung kennzeichnen jede Karte/Detaildarstellung als synthetisch. Keine gefälschten EPREL-/SDB-/Herstellerdokumente, N/A-Rechtsbegründungen oder fremden Produktabbildungen. Keine Rabatt- oder Co-Purchase-Historie in diesem Messprofil.

Preise: für Artikelnummer `n` von 1 bis 30 beträgt Ausführung A `1000 + 100 × n` EUR-Cent, B zusätzlich 10 Cent und C zusätzlich 20 Cent; fiktive Einkaufskosten sind abgerundete 75 %. Beispiel `BENCH-01-A`: 11,00 EUR, Bestand 2000. Der Bestand ermöglicht 100 tatsächliche über die Kaufports erzeugte Orders ohne eine Mengenknappheit aus der Fixture. Die Zahl **100 Orders ist keine Seedbehauptung** und muss im Messlauf tatsächlich erreicht und geprüft werden.

Benchmark in einem **getrennten Datenbankschema und Suchindex** verwenden; nicht mit dem quellengeprüften Vorführkatalog vermischen. App-/Proxybuild bleiben identisch. Der technische Lastkatalog hat weder echte Energie-/CLP-Renderlast noch dokumentierte Herstellerdaten und besitzt keine Zubehör-/Rabattregeln; Unterschiede zum realen Vorführpfad sind im Messergebnis offenzulegen. Er ersetzt keine Funktionsabnahme von Energielabels, Chemieinformationen, Zubehör oder Rabatten am RealSeed. Erfolgreiche komplette Anzeigerouten messen; gesperrte oder fehlende Artikel/Fehlerseiten sind kein Leistungsersatz. Tatsächliche Importvalidierung, Datenzahlen, Profile, Orders und Messbedingungen sind gesonderte Ausführungsnachweise.
