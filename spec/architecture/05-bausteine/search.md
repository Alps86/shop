# Suche und kuratierte Merkmale

OpenSearch ist ein abgeleitetes Product-Lesemodell. Ein Dokument pro Product, `variants` als nested Objekte mit SKU, typisierten Attributen, Inventorypreis/-verfügbarkeit, Bild und Pflichtdatenzusammenfassung. Gemeinsame Productattribute/Brand/Kategorien auf Dokumentebene; keine flache Kreuzkombination mehrerer Varianten. ProductCard zeigt eine passende Variant, weitere Varianten bleiben erkennbar. Suchausfall: sichtbarer Hinweis und Retry, Detail/Cart/Checkout lesen PostgreSQL und funktionieren weiter.

## Getrennte Query-Verantwortlichkeiten

Filterplanung validiert Attribute und erzeugt Selektionsprädikate. Facettenplanung bildet pro Facette eigenen Zählkontext. Ranking liefert begrenzte zusätzliche Scores. Sortierung bestimmt eindeutige Ordnung. Vorschläge sind eigener Querypfad. Kleine konkrete Komponenten genügen; Strategy nur bei tatsächlich mehreren Implementierungen, kein monolithischer Universalbuilder.

Textsuche auf deutsch analysiertem Name/Brand/Beschreibung, Name stärker gewichtet als Beschreibung. Die Hauptliste verwendet die eingegebenen Begriffe ohne automatische unscharfe Erweiterung. Eine Tippfehlerkorrektur wird ausschließlich angeboten und erst nach bewusster Auswahl als neue echte Suche ausgeführt; bisher automatisch zusätzlich passende unscharfe Hauptkarten können daher entfallen. Innerhalb einer Facette ausgewählte Werte OR, zwischen verschiedenen Facetten AND. Alle Variantenprädikate einschließlich Preis müssen in **einem** nested-Kontext dieselbe SKU erfüllen. Nichtverfügbare Varianten dürfen sichtbar sein, sind mit unavailable markiert; Defaultkarte bevorzugt verfügbare passende Variant, sonst passende nichtverfügbare. Pflichtdatengesperrte Produkte liefern kein kaufbares Angebot.

Trefferselektion als `post_filter` erhält die Aggregationsbasis. Für jede Facette zählen Aggregationen mit allen aktiven Filtern **außer der eigenen Facette**, einschließlich Text/Kategorie. Bei nested Aggregationen Productanzahl über reverse_nested zählen, nicht SKUanzahl; selbst bei mehreren passenden Varianten zählt ein Product einmal. Marken bleiben mit positiver kontextueller Anzahl sichtbar, ausgewählte Nullwerte bleiben als abwählbare Chips. Bereiche bilden dieselbe disjunktive Logik. Kategorie liefert maximal vier Attributfacetten plus Marke/Preis, höchstens drei Topmerkmale je Karte.

## Exaktes Zählbeispiel

Alle Produkte in derselben Smartphonekategorie:

| Product | Brand | Varianten (SKU: Farbe, Speicher, Preis EUR) |
| --- | --- | --- |
| p-a | Alpha | a-red: red, 128 GB, 100; a-blue: blue, 256 GB, 150 |
| p-b | Beta | b-red: red, 256 GB, 120 |
| p-c | Gamma | c-blue: blue, 128 GB, 90 |

Ohne Filter: 3 Produkte, Marke Alpha=1/Beta=1/Gamma=1. Brand Alpha: Treffer p-a, Markenfacette bleibt Alpha=1/Beta=1/Gamma=1. Zusätzlich Beta: Treffer p-a,p-b (OR erweitert). Farbe red + Speicher256: nur p-b, niemals p-a durch Vermischen von a-red/a-blue. Dazu Brand Alpha: 0 Treffer, Markenfacette Alpha=0/Beta=1 (ausgewähltes Alpha bleibt abwählbar); Farb-/Speicherfacetten ignorieren jeweils nur ihren eigenen Filter. Preis≤130 und Speicher256: p-b, nicht p-a. Ein Modell mit zwei passenden SKUs liefert weiterhin eine Karte/Zählung.

## Sortierung, Boosts und Vorschläge

Passende Kartenvariant: bei `priceAsc` günstigste passende, bei `priceDesc` teuerste passende, sonst verfügbare günstigste passende; Gleichstand SKU lexikografisch. Relevanz = Textscore + 0,10×InterestWeight + 0,05×MarginSignal, beide Signale auf [0,1] begrenzt. MarginSignal intern aus vorbereiteter Einkaufskostengrundlage normalisiert, fehlt → 0. Nicht im öffentlichen DTO/HTML/URL. Profilwirkung nur auf Relevanz, keine unpassenden Treffer. Relevanzgleichstand ProductID aufsteigend.

`priceAsc`/`priceDesc` ordnen strikt nach gewähltem passendem Bruttopreis, dann ProductID; Boost darf explizite Ordnung nicht überstimmen. `nameAsc` normalisierter Name dann ProductID. Alle Preise in angefragter unterstützter Währung. Kein stiller Currencymix. Pagination 12 Modelle, stabile Ordnung innerhalb gleicher Indexversion; Liveänderungen können Seiten verschieben, keine Snapshotgarantie.

Autocomplete ab 2 Zeichen, höchstens 5 Vorschläge aus publizierten Produktnamen/Kategorien, Prefix vor Tippfehler, deterministischer Namens-/ID-Gleichstand. Tippfehler höchstens Levenshtein1 bei 4–7 Zeichen, 2 ab8, erste zwei Zeichen unverändert. Vorschläge kennzeichnen „Meinten Sie …“, niemals automatisch Query ersetzen. Die bisherige Vollnamen-/Kategoriecompletion hat Vorrang. Nur wenn sie leer ist, ergänzt ein öffentlich sichtbarkeitsgebundenes Wort-Completionfeld eine vollständige korrigierte Query. Seine eindeutigen Unicode-Wörter stammen aus Produktname, Brand und Kategorienamen. Pro Eingabewort ab vier Zeichen gilt dieselbe Levenshtein-/Prefixgrenze; exakt bekannte Wörter und unbekannte Wörter bleiben unverändert, andere Kandidaten ordnen Prefix vor Tippfehler und dann lexikografisch. Mindestens ein Wort muss tatsächlich geändert werden; höchstens eine vollständige Wortkorrektur wird angeboten. Originalquery, Satzzeichen und übrige Wörter bleiben erhalten. Nicht veröffentlichte, zurückgezogene oder gelöschte Quellen liefern keine öffentlichen Wortangebote. Leere Suche zeigt Kategorie/Angebote ohne erfundene Treffer; keine Treffer zeigt aktive Chips und Filterzurücksetzen. Indexverzug ist sichtbar über Datenstand, Preise beim Kauf immer maßgeblich neu prüfen.

## Transport für Hauptsuche und Completion

Sobald die bestehende Vorschlagsregel Completion erzeugt, sendet der Adapter genau einen aliasgebundenen `POST /{alias}/_msearch` mit `application/x-ndjson`. Vier JSON-Zeilen mit abschließendem Newline enthalten zweimal Metadaten mit `request_cache: true` und je einen Suchkörper. Ohne Completion bleibt der bisherige `_search`-Pfad bestehen.

Der erste Kontext enthält unverändert Haupttreffer, vollständige gewählte Variantkarten, Sortierung, Mengen und Facetten. Root und gewählter Inner-Hit verwenden `_source: false`. Root liest ausschließlich `stored_fields: [product_id]`; der gewählte Inner-Hit liest `variants.sku` und `variants.card_payload`. Das Binaryfeld enthält die vollständige öffentliche ProductCard als UTF-8-JSON, im Wireformat Base64. ProductID und SKU sind ebenfalls gespeicherte Keywords. Die globale `_source` und die bestehende `variants.card` bleiben erhalten. Der zweite Kontext enthält die unveränderte Vollnamencompletion und gegebenenfalls die Wortcompletiongruppen, `size: 0`, `_source: false` und `match_all`. Dessen Trefferzahl und Treffer werden nicht in das fachliche Ergebnis übernommen. Die Publikationsgrenze bleibt für Hauptdaten über `deleted/published` und für Completion über denselben öffentlichen Visibility-Kontext erhalten.

Beide Teile können bei Liveänderungen oder einem Aliaswechsel unterschiedliche Generationen sehen. Ältere Hauptdaten mit neueren Vorschlägen und die umgekehrte Kombination sind erlaubt; Hauptkarten und Facetten bleiben untereinander ein Kontext. Es gibt keine Snapshotgarantie oder atomare Aliasauflösung über beide Teile. Rebuild mit Catchup und atomarem Aliaswechsel sowie CDC bleiben unverändert.

Äußerer Fehler, unlesbare Antwort, andere Antwortanzahl als zwei, fehlender erfolgreicher Teilstatus, ein Teilfehler, Timeout oder fehlgeschlagene Shards führen insgesamt zu `SearchUnavailable` ohne Teilergebnis. Leere Treffer und fehlende Vorschläge bleiben gültig. Der bestehende 250-ms-Kontext gilt einmal für den gesamten HTTPaufruf einschließlich Bodylesen, ohne Retry oder neues Budget je Teil. Synchrones JSON-Dekodieren wird durch diesen Kontext nicht hart unterbrochen.


Die anonyme öffentliche SSR-Suchseite darf das bereits validierte HTML in Varnish gemeinsam speichern, höchstens fünf Sekunden und kürzer bei Pflichtdaten-/Regelgrenzen. Cookie, Authorization und persönliche Interessen umgehen diesen Cache. Ein gültiger Hit kann bei inzwischen ausgefallenem Suchdienst200 liefern; Miss oder Ablauf erhält den echten Originfehler503/no-store, ohne stale-on-error. Hauptsuche und Completion bleiben auf einem Miss unverändert ein SearchPort-Aufruf mit dem bestehenden gemeinsamen Budget. Der technische globale Suchtag invalidiert auch Leersuchen und Nulltreffer nach bestätigter Projektion; Details und In-flight-Grenze stehen im [Kompositionsvertrag](../08-querschnitt/composition.md#globale-invalidierung-öffentlicher-suchantworten).

Die Payload entsteht erst beim Serialisieren des gesamten Projektionsdokuments aus genau derselben bereits aufgelösten Karte wie `variants.card`. Der Adapter kopiert das Dokument und dessen Variantenslice vor der Payloadbelegung; der Aufruferwert bleibt unverändert. Gespeicherte RootID, SKU und Payload müssen jeweils genau einen Wert enthalten; genau ein gewählter Inner-Hit ist erforderlich. Ungültige IDs, Base64 oder Karten-JSON und abweichende ProductID/SKU führen geschlossen zu `SearchUnavailable` ohne Teilkarten. Ein Sourcefallback existiert nicht.

Das Mapping mit gespeicherten Feldern benötigt eine tatsächlich neue Indexgeneration mit vollständigem Rebuild und Aliaswechsel. `Initialize` auf einem bestehenden Alias, ein Mappingupdate oder erneutes Schreiben in eine Altgeneration ersetzt diesen Rebuild nicht. Der Runtime-Start führt den vorhandenen Rebuild vor Bereitschaft aus. Direkte Adapternutzung muss dieselbe Voraussetzung erfüllen; eine neue Suche gegen einen alten Index scheitert geschlossen. Alte Generationen bleiben für Prüfung und Rollback erhalten. Für einen einzelnen nativen Appprozess besteht keine Zero-Downtime-Zusage.
