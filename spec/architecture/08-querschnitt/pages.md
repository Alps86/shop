# Seiten-, Komponenten- und Referenzvertrag

Beobachtungsstand **15.10.2026**, öffentliche lesende Referenzanalyse. EP ist Hauptreferenz für Grundkonzept/Struktur, MEDIMAX nur Vergleich; keine visuelle Kopierpflicht. Die folgenden Beobachtungen beziehen sich auf zugängliche öffentliche Seiten/Textansichten, keine verifizierte Mobilansicht und keine vollständige Interaktionsanalyse. Eigene responsive Entscheidung: ab schmalem Viewport einspaltiger Inhalt, einklappbare Navigation/Filter, beschriftete Bedienelemente; keine Referenzgleichheit behaupten.

## Quellenregister

| Kürzel | Gelesene Quelle / Beobachtung |
| --- | --- |
| E1 | [EP Start](https://www.ep.de/): Standort, Logo/Navigation/Suche, Kontakt/Cart, Kampagne, Angebote, Marken-/Servicewelt/Footer |
| E2 | [EP Smartphones](https://www.ep.de/c/115/telefon-navi/handy-smartphone): Breadcrumb/Kategorie, Facetten/Sortierung, Grid/Liste/Pagination, Preis-/Liefer-/Energiekarten |
| E3 | [EP Galaxy A57](https://www.ep.de/p/1518897/galaxy-a57-5g-a576b-128gb): Galerie, Marke/Name/Artikelnummer, Preis/UVP/Versand, Energie/Datenblatt, Kauf-/Händlerbereich, Detail-/Techniktabs/Empfehlungen; kein expliziter Farb-/Speicherwähler in geprüfter Ansicht belegt |
| E4 | [EP Waschmaschinen](https://www.ep.de/c/33/haushalt-kueche/waschen-trocknen/waschmaschinen): Kategorie/Haushaltskarten, Energieklasse/Skala/Datenblatt |
| E5 | [EP EcoDecalk](https://www.ep.de/p/1045137/ecodecalk-entkalker-500ml): Füllmenge/Grundpreis/Identifikatoren, Herstellerbereich; konkrete Gefahren-/Kontaktdaten aus Textansicht unvollständig |
| E6 | [EP leerer Cart](https://www.ep.de/cart): leerer Warenkorb öffentlich gelesen; gefüllter Cart/Checkout nicht beobachtet |
| M1 | [MEDIMAX Start](https://www.medimax.de/), [Smartphones](https://www.medimax.de/c/116/telefon-navi/handy-smartphone/smartphones), [Galaxy A37](https://www.medimax.de/p/1518879/galaxy-a37-5g-a376b-128gb), [Waschmaschine](https://www.medimax.de/p/1443886/wge02490), [leerer Cart](https://www.medimax.de/cart): ergänzender Strukturvergleich, keine zweite Marken-/Mandantenpflicht |

Fremde Warenkörbe wurden nicht befüllt, keine Logins/Reservierungen/Zahlungen/Bestellungen ausgeführt. Eigene Cart-/Checkout-/Status-/Backendpfade sind **Projektentscheidungen aus Nutzervertrag**, keine behauptete EP-Beobachtung.

## Vollständige Komponentenmatrix

Jede Zeile nennt Übernahme/Abweichung, Daten und Includegrenze. TTL/Fehlertechnik sind zusätzlich verbindlich unter [Komposition](composition.md), Routen unter [SEO](seo.md). `public5s` bedeutet öffentlicher kontextgebundener ESI-Cache ≤5s, `private` bedeutet no-store.

| Seite / Komponente | Quelle, Übernahme und bewusste Abweichung | Daten / Hugo-Include-HTMX / Cache |
| --- | --- | --- |
| Gemeinsamer Kopf | E1/M1: eigenes Logo, Produktenavigation, Suchfeld, Kontaktlink, Cartlink übernommen; Standort/Marktsuche/Account entfallen, eigener Profilschalter ergänzt | Hugo-Rahmen+Nginx-SSI Header/Footer 1h; CatalogReader.Category Navigation5min; Cartbadge/Profile nur privates HTMX |
| Start: Kampagne/Kategorien | E1: Kampagnenteaser, Kategorieeinstiege und kurze Service-/Markenwelt übernommen; eigene kurze Texte/eigene Identität, kein EP-Branding | Eigener Markdowncontent Hugo 1h; Links auf eigene Kategorie/Markenfilter/Serviceseite |
| Start: Angebote | E1: Karten mit Bild/Preis/Lieferinfo/Energie übernommen; nur verifizierte UVP, keine erfundene Ersparnis/Bewertung | Catalog/Search DTO, ESI public5s, Pflichtdaten vollständig; nicht personenbezogen |
| Kategorie: Navigation/Redaktion | E2/E4: Breadcrumb, Baum, kurze Kategorieeinleitung/Teaser übernommen; wenige vorbereitete Kategorien | CatalogReader.Category + Markdown; statischer Rahmen/SSI, aktuelle Navigationsdaten5min |
| Listing: Filter/Sort/Grid/Liste/Pagination | E2: Bedienelemente übernommen; eigene bessere disjunktive Suche, keine Facettenflut | SearchPort + Kategorieconfig; GET/HTMX ESI public5s nur ohne Personalisierung, sonst private |
| Produktkarte | E2/E4: Bild/Name/Preis/Lieferinfo/Energieklasse+Skala/Datenblatt übernommen; eine Karte pro Product mit passender Variant/Variantenanzahl und maximal3 Topmerkmalen | ProductCard aus Suchprojektion, ESI wie Listing; Link mit passender SKU, keine interne Marge |
| Suche/Autocomplete | E1/E2 Suchzugang; Query/Facetten/Ranking bewusst eigener Vertrag, weitere Marken bleiben auswählbar | SearchPort; q/Filter GET/HTMX, Vorschläge private; bei Suchausfall Kategorie/Detail direkt erreichbar |
| Detail: Galerie/Identität | E3/E5: Galerie, Marke/Name/Artikelkennung übernommen; SKU klar als Variante | CatalogReader aus PostgreSQL; Go-Template-Detail-ESI public5s, Hugo nur Rahmen/Redaktion |
| Detail: Farb-/Speicherwahl | E3 getrennte Farben/fehlender belegter Wähler bewusst verbessert: kombinierte gültige SKUoptionen | VariantDTO; normales GET mit sku, HTMX ersetzt Galerie/Kauf-/Pflichtdaten/Zubehör als Einheit, History aktualisiert |
| Detail: Preis/Kauf/Lieferinfo | E3: Bruttopreis/Versand/Lieferbarkeit/Kaufaktion übernommen; keine Händlerangebote, Abholung/Reservierung, UVP nur verifiziert | Inventory+Compliance; public5s, Kauf POST private an eigenen Cart, Abschluss immer neu maßgeblich prüfen |
| Detail: Pflichtdaten/Technik/Hersteller | E3/E4/E5: Datenblatt/Energie/Hersteller/Details übernommen; einfache zugängliche Abschnitte statt verpflichtendem Tab-JS | ProductDetail vollständiger Compliancekontext; public5s; richtige Variante/Sprache, Warnung nicht nurPDF |
| Detail: Zubehör | E3 Empfehlungen übernommen; konkrete Kompatibilität/Statistik/Rabatt eigener Vertrag | RecommendationPort, persönliche Reihenfolge private HTMX; öffentliche neutrale Liste public5s |
| Leerer Cart | E6 Grundzustand übernommen, eigener Einkauf-/Kategoriepfad | CartReader private, keine fremde Cartimplementierung; dynamische Ganzseite ohne shared ESI |
| Gefüllter Cart | Eigener Nutzervertrag: SKU/Mengen/Entfernen, Bewertung, Zubehör annehmen/ablehnen, Rabattaufschlüsselung/Wiederaufnahme | CartService private HTML/HTMX, aktuelle fachliche Evaluation, kein Preis aus Hugo |
| Gastcheckout | Eigener Nutzervertrag: Adresse, Standard/Express, Rechnung/simulierte Kreditkarte, neue Gesamtbestätigung, Demoabschluss | Cart-/CheckoutService private, Temporalprozess; keine echte Zahlungsmaske/Kartendaten |
| Bestätigung/Status | Eigener Nutzervertrag: Kauf-Snapshots/Adresse/Preis/Versand/Zahlung, getrennte Liefer-/Teilliefermengen | OrderReader private, geschützter Sessionbesitz; Statusrefresh HTTP/HTMX |
| /backend Katalog/Inventory | Eigener lokaler Bereich, keine EP-Beobachtung: Katalog lesend, Preis/Bestand ändern | CatalogReader/InventoryService, privater lokaler Listener, no-store, ExpectedRevision |
| /backend Order/Lieferdemo | Eigener lokaler Bereich: Orderliste/-detail, verifizierte Teilliefermengen/Statusbuttons | AdminAdminOrderReader/FulfillmentService→Temporal, no-store, Simulationslabel |
| Footer/Service/Marken | E1/M1: Kontakt/Service/Marken-/Infostruktur übernommen, eigene kurze Demoseiten | Hugo-Markdown+SSI1h; keine echte Beratung/Servicebuchung, keine erfundenen Betreiber-Rechtstexte |

Bewertungen/Sterne, Karriere, Konto/Login, Händler-/Marktauswahl, lokale Reservierung/Abholung, echte Zahlungsanbieter, Lieferantenportal und Multihändlerpreise sind bewusst **außerhalb des POC**. Nachhaltigkeits-/Herstellerlinks werden als verifizierte Produktinformationen übernommen, kein eigener Bewertungsdienst. EP-UVP/Marketingaktionen werden nicht pauschal als gültige eigene Angebote ausgegeben. So wird aus jedem beobachteten Bereich ausdrücklich Umfang oder Ausschluss, kein stilles Featurewachstum.

## Seitenzustände und nächste Aktion

| Seite | Normal / leer oder nicht vorhanden | Laden/Aktualisieren | Fehler / zulässiger nächster Schritt |
| --- | --- | --- | --- |
| Start/Rahmen | Kampagne+Angebote; ohne Angebote redaktionelle Kategorien sichtbar | Persönlicher Badge mit neutralem Ausgangslink, kein fremder Gastwert | Angebotsfragment nicht verfügbar: Hinweis/Retry, Kategorien weiter nutzbar |
| Kategorie/Listing | Karten+Facetten; 0 Treffer mit Chips und Reset, unbekannte Kategorie404 | Filterpanel/Ergebnisregion aria-busy, Auswahl bleibt sichtbar | SearchUnavailable:503 mit Wiederholen und Kategorie-/Detailnavigation |
| Suche | Modelle+Vorschläge; leerer q zeigt Kategorieeinstiege, keine Treffer bietet Tippfehler-/Filterhilfe | Debounce150ms für Vorschläge, neue Query ersetzt ältere Antwort | Suchfehler behält q/Filter, Retry; keine blockierte Detailseite |
| Produktdetail | Passende SKU+Galerie+Pflichtdaten; unbekannt404/entfernt410 | Variantenwechsel sperrt Kauf bis vollständige neue Einheit da | SKU unavailable: Kauf deaktiviert, verfügbare Varianten aktiv wählbar; Pflichtdaten pending: Prüfungshinweis; Fragmentfehler503/Retry |
| Cart | Aktuelle Positionen+Aufschlüsselung; leer zeigt Einkaufenlink | Mengenänderung regionbusy, vorige Werte als in Aktualisierung markiert | Ungültige Menge422 beim Feld; Preisänderung neue Bewertung+Bestätigen; abgelaufen410→neuer Cart |
| Checkout | Adresse/Versand/Zahlung/Bestätigung; leerer Cart zurück zum Cart | Zahlungsfreigabe/Abschluss als Simulation markiert, Doppelbutton gesperrt | Inputfehler behält Felder; Preis409→neu bestätigen; Fehlbestand→aktive Ersatzwahl; abgelehnte Simulation→Zahlungswahl/Retry |
| Bestätigung/Status | Snapshot und Fortschritt; fremde/unbekannte Order404 | Statusrefresh markiert Aktualisierung, Kauf bleibt stabil | Prozessstart nachgeholt: „Bestellung angenommen, Status wird aktualisiert“; technischer Lesefehler503→Retry |
| Backend | Leselisten/Aktionen; keine Order zeigt explizite leere Liste | Commandbutton bis Quittung gesperrt, neue Revision laden | Konflikt409→neu laden; unzulässige Teilmenge422; temporärer Fehler→gleiche CommandID retry |

Kein Ladezustand zeigt ungeprüfte Pflichtdaten als aktuell/kaufbar. Fehlgeschlagene Änderung bleibt sichtbar ohne stillen Erfolg. Doppelte Abschlussantwort zeigt dieselbe Order, kein zweiter Kauf; Workerwiederanlauf zeigt vorhandenen fachlichen Zustand, keine neue leere Bestellung.

## Verbindliche Klickpfade

1. Start → Smartphonekategorie → Marke Alpha auswählen → Beta ergänzen (mehr Treffer) → Farb-/Speicherfilter → Modellkarte mit passender SKU → Variante aktiv wechseln.
2. Detail → eigenen Cart hinzufügen → Zubehör ablehnen oder annehmen → Menge ändern/Artikel entfernen → aktuelle Aufschlüsselung → Gastcheckout.
3. Checkout → Adresse → Standard/Express → Rechnung oder simulierte Kreditkarte → aktuelle Bewertung ausdrücklich bestätigen → „Demo-Bestellung verbindlich abschließen“ → Bestätigung/Status. Dauerhafter Banner „Simulation – keine echte Bestellung/Zahlung“ auf Cart/Checkout/Order.
4. Preis-/Bestandsänderung im lokalen Backend → Detail/Cart aktualisieren → geänderten Gesamtpreis bestätigen oder ausverkaufte SKU aktiv ersetzen → erneut abschließen.
5. Backend Order → Teilmenge versenden → Kunde sieht teilweise versandt → Rest versenden/liefern → Kunde sieht vollständig geliefert; Kaufwerte unverändert.
6. Vor Ablauf Browser erneut öffnen → eigener Cart wiederaufgenommen; nach168h Hinweis/neuen Cart wählen. Profilwechsel verändert Empfehlungen, nicht Besitz/Rabatt.

Diese Pfade werden über HTTP-/Renderer-/Templateverträge und echte Dienstintegration abgenommen; höchstens ein kurzer Browser-Smoke zur HTMX-/Historyverdrahtung. Keine Screenshot-/Visualmatrix. Fachregeln bleiben in Domain-/Applicationverträgen, UI prüft deren sichtbare Nutzung.
