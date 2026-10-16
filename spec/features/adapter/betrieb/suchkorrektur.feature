# language: de
@demo-http
Funktionalität: Bewusste Suchkorrektur im echten öffentlichen Shop
  Als Gast behalte ich meinen eingegebenen Suchbegriff bis zur ausdrücklichen Auswahl eines sichtbaren Korrekturangebots.

  Hintergrund:
    Angenommen die echte öffentliche Demolaufzeit ist bereit
      | PostgresEnv | TemporalEnv | SearchEnv | Namespace | SeedFile | CacheKey | CSRFKey |
      | SHOP_TEST_POSTGRES_DSN | SHOP_TEST_TEMPORAL_ADDRESS | SHOP_TEST_OPENSEARCH_URL | default | spec/fixtures/demo/katalog.json | demo-cache-key-0123456789abcdef0123456789 | demo-csrf-key-0123456789abcdef0123456789 |
    Und ein eigener HTTPgast wird verwendet
      | Jar |
      | suchkorrektur |

  Szenariogrundriss: Erst der tatsächlich angeklickte Korrekturlink verändert die echte Suche
    Wenn der Gast ruft die öffentliche Ansicht auf
      | Path |
      | <ursprung> |
    Dann die echte Suche erhält Eingabe und Treffer unverändert
      | Query | Models | CanonicalQuery |
      | <eingabe> | <vorher> | <canonical-vorher> |
    Und die öffentliche Antwort enthält
      | Status | Contains | Excludes |
      | 200 | <hilfe> | <nicht-vorher> |
    Wenn der Gast folgt dem tatsächlich angebotenen Korrekturlink
      | LinkText | QueryString |
      | <angebot> | <zielquery> |
    Dann die echte Suche erhält Eingabe und Treffer unverändert
      | Query | Models | CanonicalQuery |
      | <korrektur> | <nachher> | <canonical-nachher> |
    Und die öffentliche Antwort enthält
      | Status | Contains |
      | 200 | DR. BECHER@home Schnell Entkalker 500 ml |
    Beispiele:
      | ursprung | eingabe | vorher | canonical-vorher | hilfe | nicht-vorher | angebot | zielquery | korrektur | nachher | canonical-nachher |
      | /de-DE/suche?q=Fairphone+Entkalkre | Fairphone Entkalkre | 2 | q=Fairphone+Entkalkre | Meintest du Fairphone Entkalker | DR. BECHER@home Schnell Entkalker 500 ml | Meintest du Fairphone Entkalker | q=Fairphone+Entkalker | Fairphone Entkalker | 3 | q=Fairphone+Entkalker |
      | /de-DE/suche?q=Entkalkre&sort=nameAsc&view=list&priceMax=1000000&page=2 | Entkalkre | 0 | page=2&priceMax=1000000&q=Entkalkre&sort=nameAsc&view=list | Keine Treffer. Ändere den Suchbegriff oder setze die Filter zurück. | DR. BECHER@home Schnell Entkalker 500 ml | Meintest du Entkalker | priceMax=1000000&q=Entkalker&sort=nameAsc&view=list | Entkalker | 1 | priceMax=1000000&q=Entkalker&sort=nameAsc&view=list |
