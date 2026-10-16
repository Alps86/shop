# language: de
@http @frontend
Funktionalität: Vollständige und geordnete Kartenanzeige
  Als Käufer sehe ich dieselben sicher formatierten Karten in SSR und HX.

  Hintergrund:
    Angenommen explizite HTTPdaten liegen vor
      | Fixture | Fault |
      | spec/fixtures/http/shop.json | |

  Szenariogrundriss: Karten zeigen zwölf vollständige Einträge in stabiler Reihenfolge
    Angenommen explizite Kartenanzeigedaten ersetzen die Suchkarten
      | Suffixes | Card |
      | 01,02,03,04,05,06,07,08,09,10,11,12 | {"ProductID":"p-","URL":"/de-DE/produkt/karte--p-","Name":"Karte <b> ","SKU":"SKU-","Image":{"URL":"/assets/product.svg","AltText":{"de-DE":"Bild \"<&"}},"UnitPrice":{"MinorUnits":12345,"Currency":"EUR"},"AvailableQuantity":7,"VariantCount":3,"TopAttributes":{"z-bool":{"Type":"boolean","Boolean":{"Present":true,"Value":true}},"a-text":{"Type":"text","Text":{"Present":true,"Value":{"de-DE":"Rot <&"}}},"m-number":{"Type":"number","Number":{"Present":true,"Value":"128"}}},"Compliance":{"Context":{"VerificationStatus":"verified"},"Notices":[{"Text":{"de-DE":"Pflicht <&"}}]}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL | HTMX |
      | GET | /de-DE/suche?q=Karte | <hx> |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | alt="Bild &#34;&lt;&amp;" |
      | Body | SKU SKU-01 · 3 Varianten |
      | Body | 123,45 EUR |
      | Body | Lieferbar |
      | Body | Pflicht &lt;&amp; |
      | Absent | Karte <b> |
      | Absent | Grundpreis |
    Und die Kartenansicht folgt den geordneten HTMLwerten
      | Count | HTML |
      | 12 | Karte &lt;b&gt; 01 |
      | | <dt>a-text</dt><dd>Rot &lt;&amp;</dd><dt>m-number</dt><dd>128</dd><dt>z-bool</dt><dd>true</dd> |
      | | Karte &lt;b&gt; 02 |
      | | Karte &lt;b&gt; 03 |
      | | Karte &lt;b&gt; 04 |
      | | Karte &lt;b&gt; 05 |
      | | Karte &lt;b&gt; 06 |
      | | Karte &lt;b&gt; 07 |
      | | Karte &lt;b&gt; 08 |
      | | Karte &lt;b&gt; 09 |
      | | Karte &lt;b&gt; 10 |
      | | Karte &lt;b&gt; 11 |
      | | Karte &lt;b&gt; 12 |
    Beispiele:
      | hx |
      | false |
      | true |

  Szenariogrundriss: Karten erhalten Maße Optionen und die bisherige Grundpreisbedingung
    Angenommen explizite Kartenanzeigedaten ersetzen die Suchkarten
      | Suffixes | Card |
      | | {"ProductID":"p-quantity","URL":"/de-DE/produkt/menge--p-quantity","Name":"Mengenkarte","SKU":"SKU-QUANTITY","Image":{"URL":"/assets/product.svg"},"UnitPrice":{"MinorUnits":799,"Currency":"EUR"},"VariantCount":1,"TopAttributes":{"z-options":{"Type":"select","Options":["rot","blau"]},"a-measure":{"Type":"measurement","Measurement":{"Present":true,"Value":{"Value":"750","Unit":"ml"}}}},"Compliance":<compliance>} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche?q=Menge |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 200 |
      | Body | <expected> |
      | Absent | <absent> |
    Und die Kartenansicht folgt den geordneten HTMLwerten
      | Count | HTML |
      | 1 | <dt>a-measure</dt><dd>750 ml</dd><dt>z-options</dt><dd>rot, blau</dd> |
    Beispiele:
      | compliance | expected | absent |
      | {"NetContent":{"Present":true,"Value":{"Value":"500","Unit":"ml"}},"UnitPriceBasis":{"Present":true,"Value":{"Value":"1","Unit":"l"}}} | Grundpreis 15,98 EUR / 1 l | Nicht vorhandener Inhalt |
      | {"NetContent":{"Present":true,"Value":{"Value":"500","Unit":"ml"}}} | Mengenkarte | Grundpreis 15,98 EUR / 1 l |
      | {"UnitPriceBasis":{"Present":true,"Value":{"Value":"1","Unit":"l"}}} | Mengenkarte | Grundpreis 15,98 EUR / 1 l |

  Szenario: Kartenfehler nach gültigem Eintrag veröffentlichen keine Teilausgabe
    Angenommen explizite Kartenanzeigedaten ersetzen die Suchkarten
      | Suffixes | Card |
      | | {"ProductID":"p-first","URL":"/de-DE/produkt/erste--p-first","Name":"Erste gültige Karte","SKU":"SKU-FIRST","UnitPrice":{"MinorUnits":799,"Currency":"EUR"}} |
      | | {"ProductID":"p-quantity","URL":"/de-DE/produkt/menge--p-quantity","Name":"Spätere fehlerhafte Karte","SKU":"SKU-QUANTITY","Image":{"URL":"/assets/product.svg"},"UnitPrice":{"MinorUnits":799,"Currency":"EUR"},"VariantCount":1,"TopAttributes":{"z-options":{"Type":"select","Options":["rot","blau"]},"a-measure":{"Type":"measurement","Measurement":{"Present":true,"Value":{"Value":"750","Unit":"ml"}}}},"Compliance":{"NetContent":{"Present":true,"Value":{"Value":"500","Unit":"ml"}},"UnitPriceBasis":{"Present":true,"Value":{"Value":"1","Unit":"kg"}}}} |
    Wenn der Shop liefert eine HTTPansicht
      | Method | URL |
      | GET | /de-DE/suche?q=Karte |
    Dann die HTTPantwort enthält die sichtbaren Vertragswerte
      | Field | Value |
      | Status | 503 |
      | Header:Cache-Control | private,no-store |
      | Body | Ansicht vorübergehend nicht verfügbar |
      | Absent | Erste gültige Karte |
      | Absent | <article |
      | Absent | <!doctype |
