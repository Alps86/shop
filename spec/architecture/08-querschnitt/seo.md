# Lesbare URLs und Indexierung

Eigene Struktur, keine Kopie der EP-Pfade. Stabile IDs bleiben unverändert, Slugs sind lesbare Dekoration. Locale ist expliziter erster Pfadteil, erste Konfiguration `/de-DE/`. `/` leitet302 auf `/de-DE/`, ohne Locale/Lieferland zu verwechseln.

| Ressource | Beispiel / Status / Canonical |
| --- | --- |
| Start | `/de-DE/` 200, self canonical, index/follow |
| Kategorie | `/de-DE/kategorie/smartphones--c-phone` 200, stabile CategoryID nach `--`, self canonical |
| Produkt | `/de-DE/produkt/galaxy-demo--p-phone` 200, stabile ProductID, self canonical |
| Variante | Produktpfad `?sku=PHONE-BLUE-256` 200, aktive SKU vollständig SSR; canonical auf Produktbasis ohne SKU |
| Kategoriepagination | `?page=2` 200, self canonical einschließlich page, erste Seite ohne page; index/follow |
| Suche | `/de-DE/suche?q=phone` 200, noindex/follow, canonical normalisierte Suchroute |
| Facetten/Sort/Grid | Kategorie `?brand=Alpha&brand=Beta&color=blue&sort=priceAsc&view=list` 200, noindex/follow, canonical ungefilterte Kategorie |
| Cart/Checkout | `/de-DE/warenkorb`, `/de-DE/kasse` no-store/noindex/nofollow |
| Orderstatus | `/de-DE/bestellung/{OrderID}` nur Gastfähigkeit im Cookie, no-store/noindex/nofollow |
| Backend | lokaler `/backend` no-store/noindex/nofollow, vom öffentlichen Listener404 |

Kennung zuerst auflösen, danach Slug prüfen. Bekannte ID mit altem/falschem Slug →301 direkt aktueller kanonischer Pfad, erlaubte Query inkl. SKU erhalten, keine Redirectkette. Unbekannte ID404, bewusst endgültig entferntes Product410; ungültige Query/SKU422, SKU eines anderen Product404. Nicht unterstützte valide Locale404, ungültige Locale400. Neue Slugs kleingeschrieben, transliteriert (`ä`→`ae`), ASCII-Wortteile mit `-`, keine Identitätsableitung aus Name. URL-Encoding vor Validierung genau einmal, keine offenen externen Redirectziele.

Queryparameter sortiert, Auswahlwerte dedupliziert und lexikografisch geordnet; Defaultwerte entfernt. Wiederholte Markenwerte sind OR. Maximal vier Attributfilter plus Marke/Preis, höchstens zehn Werte pro Facette, q maximal100 Zeichen, URL maximal2048 Zeichen (414 bei Überschreitung). Unbekannter Filter/Sortoperator422 statt stiller Ersatz. Eigene HTML-Filterform und browser history erhalten Query/Seite; Filteränderung setzt page1 zurück. Canonical reduziert Indexkombinationen, Filterlinks erhalten nofollow, Sitemap enthält ausschließlich veröffentlichte Basisprodukte/ungefilterte Kategorien/Content. Robots.txt erlaubt Ressourcenrendering und sperrt Backend/private Pfade; für Facetten gilt sichtbares noindex, kein robots-Block, der dessen Lesen verhindert.

SSR liefert Titel, Beschreibung, korrekte aktive Variante und sichtbare Pflichtdaten nach Komposition; Canonical zeigt dasselbe Product wie der Inhalt. SEO-Metadaten stammen aus escaped Katalog-/Markdowntext, keine Stock-/Preisgarantie aus strukturierten Daten. Hreflang erst für tatsächlich veröffentlichte weitere Locale, keine erfundenen Übersetzungsseiten. Hugo-Routenindex enthält ID→Buildpfad, dynamischer Router schützt Redirect/Canonical vor gecachtem Rahmen. Keine persönlichen IDs/Token/Adressen in öffentlichen Querys, Logs oder Cachekeys. Kein Rankingversprechen.

Abnahmekern: direkter Produkt-/Variantenlink, alter Slug301 mit SKUerhalt, unbekannte ID404, falsche SKU404, page2 selfcanonical, Facetten noindex+Basiscanonical, geteilte Marken-OR-URL und Browserzurück erhalten Auswahl. HTTP-/Templateverträge prüfen diese Eigenschaften ohne Suchmaschinen-/Browsermatrix.
